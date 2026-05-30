package addon

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Options struct {
	ModuleName   string
	TargetDir    string
	Force        bool
	SkipExisting bool
}

type TemplateData struct {
	ModulePath  string
	ProjectName string
}

func Install(opts Options) error {
	moduleName := strings.TrimSpace(opts.ModuleName)
	if !contains(availableModules(), moduleName) {
		return fmt.Errorf("不支持的模块 %q，可用模块：%s", moduleName, strings.Join(availableModules(), ", "))
	}

	targetDir := strings.TrimSpace(opts.TargetDir)
	if targetDir == "" {
		targetDir = "."
	}

	data, err := loadTemplateData(targetDir)
	if err != nil {
		return err
	}

	if err := ensureBaseFiles(moduleName, targetDir, data); err != nil {
		return err
	}

	for path, raw := range moduleFiles(moduleName) {
		fullPath := filepath.Join(targetDir, path)
		if !opts.Force {
			if _, err := os.Stat(fullPath); err == nil {
				if opts.SkipExisting {
					continue
				}
				return fmt.Errorf("目标文件 %q 已存在，可使用 --force 覆盖", fullPath)
			}
		}

		rendered, err := render(raw, data)
		if err != nil {
			return fmt.Errorf("渲染模块模板 %s 失败：%w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("创建模块目录失败：%w", err)
		}
		if err := os.WriteFile(fullPath, []byte(rendered), 0o644); err != nil {
			return fmt.Errorf("写入模块文件 %s 失败：%w", path, err)
		}
	}

	if err := applyModulePatches(moduleName, targetDir); err != nil {
		return err
	}

	return nil
}

func ensureBaseFiles(moduleName, targetDir string, data TemplateData) error {
	for path, raw := range baseFiles(moduleName) {
		fullPath := filepath.Join(targetDir, path)
		if _, err := os.Stat(fullPath); err == nil {
			continue
		}

		rendered, err := render(raw, data)
		if err != nil {
			return fmt.Errorf("渲染基础模板 %s 失败：%w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("创建基础目录失败：%w", err)
		}
		if err := os.WriteFile(fullPath, []byte(rendered), 0o644); err != nil {
			return fmt.Errorf("写入基础文件 %s 失败：%w", path, err)
		}
	}

	return nil
}

func List() []string {
	return availableModules()
}

func render(raw string, data TemplateData) (string, error) {
	tmpl, err := template.New("addon").Parse(raw)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func availableModules() []string {
	return []string{"docker", "ci", "gitignore", "worker", "webhook", "auth", "postgres", "request-id", "timeout", "metrics", "rate-limit", "error-model", "cors"}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func loadTemplateData(targetDir string) (TemplateData, error) {
	modulePath, err := detectModulePath(targetDir)
	if err != nil {
		return TemplateData{}, err
	}

	projectName := filepath.Base(filepath.Clean(targetDir))
	if projectName == "." || projectName == string(filepath.Separator) {
		projectName = "app"
	}

	return TemplateData{
		ModulePath:  modulePath,
		ProjectName: projectName,
	}, nil
}

func detectModulePath(targetDir string) (string, error) {
	goModPath := filepath.Join(targetDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("目标目录 %q 缺少 go.mod，无法添加需要导入路径的模块", targetDir)
		}
		return "", fmt.Errorf("读取 go.mod 失败：%w", err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if modulePath == "" {
				break
			}
			return modulePath, nil
		}
	}

	return "", fmt.Errorf("未在 %q 中找到有效的 module 声明", goModPath)
}

func applyModulePatches(moduleName, targetDir string) error {
	switch moduleName {
	case "auth":
		return addAuthRoutes(targetDir)
	case "postgres":
		return addPostgresSupport(targetDir)
	case "request-id":
		return addRequestIDSupport(targetDir)
	case "timeout":
		return addTimeoutSupport(targetDir)
	case "metrics":
		return addMetricsSupport(targetDir)
	case "rate-limit":
		return addRateLimitSupport(targetDir)
	case "error-model":
		return addErrorModelSupport(targetDir)
	case "cors":
		return addCORSSupport(targetDir)
	case "worker":
		return addWorkerTarget(targetDir)
	case "webhook":
		return addWebhookRoute(targetDir)
	default:
		return nil
	}
}

func addCORSSupport(targetDir string) error {
	if err := appendEnvValue(targetDir, "CORS_ALLOW_ORIGINS=*"); err != nil {
		return err
	}
	return addMiddlewareLine(targetDir, "handler = corsMiddleware(handler)")
}

func addErrorModelSupport(targetDir string) error {
	return replaceInFile(
		filepath.Join(targetDir, "internal", "http", "middleware.go"),
		"writeJSON(w, http.StatusInternalServerError, map[string]string{\"error\": \"internal server error\"})",
		"writeError(w, r, http.StatusInternalServerError, \"internal_server_error\", \"internal server error\")",
		"统一错误返回",
	)
}

func addMetricsSupport(targetDir string) error {
	if err := addMiddlewareLine(targetDir, "handler = metricsMiddleware(handler)"); err != nil {
		return err
	}
	return addRouteLine(targetDir, "mux.HandleFunc(\"/metrics\", metricsHandler)", "/metrics")
}

func addRateLimitSupport(targetDir string) error {
	if err := appendEnvValue(targetDir, "RATE_LIMIT_PER_SECOND=20"); err != nil {
		return err
	}
	return addMiddlewareLine(targetDir, "handler = rateLimitMiddleware(handler)")
}

func addRequestIDSupport(targetDir string) error {
	return addMiddlewareLine(targetDir, "handler = requestIDMiddleware(handler)")
}

func addTimeoutSupport(targetDir string) error {
	if err := appendEnvValue(targetDir, "REQUEST_TIMEOUT=5s"); err != nil {
		return err
	}
	return addMiddlewareLine(targetDir, "handler = timeoutMiddleware(handler)")
}

func addPostgresSupport(targetDir string) error {
	for _, line := range []string{
		"DATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable",
		"DATABASE_TIMEOUT=3s",
		"LOG_LEVEL=info",
	} {
		if err := appendEnvValue(targetDir, line); err != nil {
			return err
		}
	}

	if err := addPostgresMakefileTarget(targetDir); err != nil {
		return err
	}

	if err := addRouteLine(targetDir, "mux.HandleFunc(\"/db/readyz\", postgresReadyHandler)", "/db/readyz"); err != nil {
		return err
	}

	return patchPostgresLifecycle(targetDir)
}

func addMiddlewareLine(targetDir, line string) error {
	middlewarePath := filepath.Join(targetDir, "internal", "http", "middleware.go")
	content, err := os.ReadFile(middlewarePath)
	if err != nil {
		return fmt.Errorf("读取中间件文件失败：%w", err)
	}

	raw := string(content)
	if strings.Contains(raw, line) {
		return nil
	}

	updated := strings.Replace(raw, "\t// Golider 中间件扩展锚点\n", "\t"+line+"\n\t// Golider 中间件扩展锚点\n", 1)
	if updated == raw {
		return fmt.Errorf("无法自动把中间件 %q 写入 %q", line, middlewarePath)
	}

	if err := os.WriteFile(middlewarePath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入中间件文件失败：%w", err)
	}

	return nil
}

func addPostgresMakefileTarget(targetDir string) error {
	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("读取 Makefile 失败：%w", err)
	}

	raw := string(content)
	targetBlock := "db-check:\n\tgo run ./cmd/dbcheck\n"
	if strings.Contains(raw, targetBlock) {
		return nil
	}

	updated := strings.Replace(raw, "# Golider 扩展命令锚点\n", targetBlock+"\n# Golider 扩展命令锚点\n", 1)
	if updated == raw {
		updated = raw + "\n" + targetBlock
	}

	if err := os.WriteFile(makefilePath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入 Makefile 失败：%w", err)
	}

	return nil
}

func addRouteLine(targetDir, routeLine, routeLabel string) error {
	routerPath := filepath.Join(targetDir, "internal", "http", "router.go")
	content, err := os.ReadFile(routerPath)
	if err != nil {
		return fmt.Errorf("读取路由文件失败：%w", err)
	}

	raw := string(content)
	if strings.Contains(raw, routeLine) {
		return nil
	}

	updated := strings.Replace(raw, "\t// Golider 路由扩展锚点\n", "\t"+routeLine+"\n\t// Golider 路由扩展锚点\n", 1)
	if updated == raw {
		updated = strings.Replace(raw, "\treturn withMiddlewares(mux)\n", "\t"+routeLine+"\n\treturn withMiddlewares(mux)\n", 1)
	}
	if updated == raw {
		return fmt.Errorf("无法自动把路由 %s 写入 %q", routeLabel, routerPath)
	}

	if err := os.WriteFile(routerPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入路由文件失败：%w", err)
	}

	return nil
}

func replaceInFile(path, oldValue, newValue, label string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取%s文件失败：%w", label, err)
	}

	raw := string(content)
	if strings.Contains(raw, newValue) {
		return nil
	}

	updated := strings.Replace(raw, oldValue, newValue, 1)
	if updated == raw {
		return fmt.Errorf("无法自动更新%s，请手动调整 %q", label, path)
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入%s文件失败：%w", label, err)
	}

	return nil
}

func patchPostgresLifecycle(targetDir string) error {
	mainPath := filepath.Join(targetDir, "cmd", "api", "main.go")
	storeImport := "\t\"" + detectOrFallbackModule(targetDir) + "/internal/store\"\n"
	if err := ensureImport(mainPath, "\t\"os\"\n", storeImport); err != nil {
		return err
	}

	block := "\tpostgresManager := store.NewPostgresManager(os.Getenv(\"DATABASE_URL\"))\n\tlifecycle.OnStart(\"postgres\", postgresManager.Start)\n\tlifecycle.OnStop(\"postgres\", postgresManager.Stop)\n"
	return insertAfter(mainPath, "\tlifecycle := app.New()\n", block, "postgres 生命周期")
}

func detectOrFallbackModule(targetDir string) string {
	modulePath, err := detectModulePath(targetDir)
	if err != nil || strings.TrimSpace(modulePath) == "" {
		return "app"
	}
	return modulePath
}

func ensureImport(path, anchor, importLine string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取导入文件失败：%w", err)
	}

	raw := string(content)
	if strings.Contains(raw, strings.TrimSpace(importLine)) {
		return nil
	}

	updated := strings.Replace(raw, anchor, anchor+importLine, 1)
	if updated == raw {
		return fmt.Errorf("无法自动写入导入到 %q", path)
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入导入文件失败：%w", err)
	}

	return nil
}

func insertAfter(path, anchor, block, label string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取%s失败：%w", label, err)
	}

	raw := string(content)
	if strings.Contains(raw, strings.TrimSpace(block)) {
		return nil
	}

	updated := strings.Replace(raw, anchor, anchor+block, 1)
	if updated == raw {
		return fmt.Errorf("无法自动插入%s到 %q", label, path)
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入%s失败：%w", label, err)
	}

	return nil
}

func addAuthRoutes(targetDir string) error {
	if err := appendEnvValue(targetDir, "AUTH_TOKEN=dev-token"); err != nil {
		return err
	}

	routerPath := filepath.Join(targetDir, "internal", "http", "router.go")
	content, err := os.ReadFile(routerPath)
	if err != nil {
		return fmt.Errorf("读取路由文件失败：%w", err)
	}

	raw := string(content)
	loginRoute := "mux.HandleFunc(\"/auth/login\", loginExampleHandler)"
	profileRoute := "mux.HandleFunc(\"/auth/profile\", profileExampleHandler)"
	if strings.Contains(raw, loginRoute) && strings.Contains(raw, profileRoute) {
		return nil
	}

	routeBlock := "\t" + loginRoute + "\n\t" + profileRoute + "\n\t// Golider 路由扩展锚点\n"
	updated := strings.Replace(raw, "\t// Golider 路由扩展锚点\n", routeBlock, 1)
	if updated == raw {
		updated = strings.Replace(raw, "\treturn withMiddlewares(mux)\n", "\t"+loginRoute+"\n\t"+profileRoute+"\n\treturn withMiddlewares(mux)\n", 1)
	}
	if updated == raw {
		return fmt.Errorf("无法自动把 auth 路由写入 %q，请手动注册 /auth/login 和 /auth/profile", routerPath)
	}

	if err := os.WriteFile(routerPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入路由文件失败：%w", err)
	}

	return nil
}

func addWorkerTarget(targetDir string) error {
	if err := appendEnvValue(targetDir, "LOG_LEVEL=info"); err != nil {
		return err
	}

	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("读取 Makefile 失败：%w", err)
	}

	raw := string(content)
	targetBlock := "run-worker:\n\tgo run ./cmd/worker\n"
	if strings.Contains(raw, targetBlock) {
		return nil
	}

	updated := strings.Replace(raw, "# Golider 扩展命令锚点\n", targetBlock+"\n# Golider 扩展命令锚点\n", 1)
	if updated == raw {
		updated = raw + "\n" + targetBlock
	}

	if err := os.WriteFile(makefilePath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("写入 Makefile 失败：%w", err)
	}

	return nil
}

func addWebhookRoute(targetDir string) error {
	if err := appendEnvValue(targetDir, "LOG_LEVEL=info"); err != nil {
		return err
	}

	routerPath := filepath.Join(targetDir, "internal", "http", "router.go")
	content, err := os.ReadFile(routerPath)
	if err != nil {
		return fmt.Errorf("读取路由文件失败：%w", err)
	}

	raw := string(content)
	routeLine := "mux.HandleFunc(\"/webhooks/example\", exampleWebhookHandler)"
	if strings.Contains(raw, routeLine) {
		return nil
	}

	withMarker := strings.Replace(raw, "\t// Golider 路由扩展锚点\n", "\t"+routeLine+"\n\t// Golider 路由扩展锚点\n", 1)
	if withMarker == raw {
		withMarker = strings.Replace(raw, "\treturn withMiddlewares(mux)\n", "\t"+routeLine+"\n\treturn withMiddlewares(mux)\n", 1)
	}
	if withMarker == raw {
		return fmt.Errorf("无法自动把 webhook 路由写入 %q，请手动注册 /webhooks/example", routerPath)
	}

	if err := os.WriteFile(routerPath, []byte(withMarker), 0o644); err != nil {
		return fmt.Errorf("写入路由文件失败：%w", err)
	}

	return nil
}

func appendEnvValue(targetDir, line string) error {
	envPath := filepath.Join(targetDir, ".env.example")
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("读取环境变量模板失败：%w", err)
	}

	raw := string(content)
	if strings.Contains(raw, line) {
		return nil
	}

	if !strings.HasSuffix(raw, "\n") {
		raw += "\n"
	}
	raw += line + "\n"

	if err := os.WriteFile(envPath, []byte(raw), 0o644); err != nil {
		return fmt.Errorf("写入环境变量模板失败：%w", err)
	}

	return nil
}
