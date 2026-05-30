package check

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Item struct {
	Name   string
	Path   string
	Exists bool
}

type Capability struct {
	Name    string
	Exists  bool
	Detail  string
	Related string
}

type ConfigRequirement struct {
	Name   string
	Exists bool
	Valid  bool
	Value  string
	Detail string
}

func RequiredItems(projectDir string) []Item {
	items := []Item{
		{Name: "Go module", Path: "go.mod"},
		{Name: "服务入口", Path: filepath.Join("cmd", "api", "main.go")},
		{Name: "应用装配", Path: filepath.Join("internal", "app", "app.go")},
		{Name: "配置模块", Path: filepath.Join("internal", "config", "config.go")},
		{Name: "HTTP 中间件", Path: filepath.Join("internal", "http", "middleware.go")},
		{Name: "HTTP 路由", Path: filepath.Join("internal", "http", "router.go")},
		{Name: "日志模块", Path: filepath.Join("internal", "observability", "logger.go")},
		{Name: "环境变量模板", Path: ".env.example"},
		{Name: "Git 忽略文件", Path: ".gitignore"},
		{Name: "Dockerfile", Path: "Dockerfile"},
		{Name: "GitHub Actions", Path: filepath.Join(".github", "workflows", "ci.yml")},
		{Name: "Makefile", Path: "Makefile"},
	}

	for idx := range items {
		_, err := os.Stat(filepath.Join(projectDir, items[idx].Path))
		items[idx].Exists = err == nil
	}

	return items
}

func MissingItems(projectDir string) []Item {
	var missing []Item
	for _, item := range RequiredItems(projectDir) {
		if !item.Exists {
			missing = append(missing, item)
		}
	}
	return missing
}

func Capabilities(projectDir string) []Capability {
	router := readMaybe(filepath.Join(projectDir, "internal", "http", "router.go"))
	middleware := readMaybe(filepath.Join(projectDir, "internal", "http", "middleware.go"))
	envFile := readMaybe(filepath.Join(projectDir, ".env.example"))
	makefile := readMaybe(filepath.Join(projectDir, "Makefile"))
	configFile := readMaybe(filepath.Join(projectDir, "internal", "config", "config.go"))
	apiMain := readMaybe(filepath.Join(projectDir, "cmd", "api", "main.go"))
	appFile := readMaybe(filepath.Join(projectDir, "internal", "app", "app.go"))
	readinessFile := readMaybe(filepath.Join(projectDir, "internal", "http", "readiness.go"))

	return []Capability{
		{
			Name:    "健康检查",
			Exists:  strings.Contains(router, `mux.HandleFunc("/healthz"`) && strings.Contains(router, `mux.HandleFunc("/readyz"`),
			Detail:  "包含 /healthz 与 /readyz",
			Related: "internal/http/router.go",
		},
		{
			Name:    "路由扩展锚点",
			Exists:  strings.Contains(router, "Golider 路由扩展锚点"),
			Detail:  "支持后续自动注入路由",
			Related: "internal/http/router.go",
		},
		{
			Name:    "中间件扩展锚点",
			Exists:  strings.Contains(middleware, "Golider 中间件扩展锚点"),
			Detail:  "支持后续自动注入中间件",
			Related: "internal/http/middleware.go",
		},
		{
			Name:    "请求日志",
			Exists:  strings.Contains(middleware, "requestLogMiddleware") && fileExists(filepath.Join(projectDir, "internal", "observability", "logger.go")),
			Detail:  "请求完成日志已接入",
			Related: "internal/http/middleware.go",
		},
		{
			Name:    "配置校验",
			Exists:  strings.Contains(configFile, "func Load() (Config, error)") && strings.Contains(configFile, "func validate(cfg Config) error"),
			Detail:  "配置加载包含显式校验",
			Related: "internal/config/config.go",
		},
		{
			Name:    "生命周期钩子",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "app", "app.go")) && strings.Contains(appFile, "func (a *App) OnStart") && strings.Contains(appFile, "func (a *App) OnStop") && strings.Contains(apiMain, "lifecycle.OnStop"),
			Detail:  "启动与停机使用生命周期钩子装配",
			Related: "internal/app/app.go",
		},
		{
			Name:    "优雅摘流",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "readiness.go")) && strings.Contains(readinessFile, "func MarkNotReady") && strings.Contains(apiMain, `MarkNotReady("shutting_down")`),
			Detail:  "停机前切换为未就绪状态",
			Related: "internal/http/readiness.go",
		},
		{
			Name:    "统一错误模型",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "errors.go")) && strings.Contains(middleware, "writeError("),
			Detail:  "统一错误返回结构已接入 recover",
			Related: "internal/http/errors.go",
		},
		{
			Name:    "请求标识",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "requestid.go")) && strings.Contains(middleware, "requestIDMiddleware"),
			Detail:  "请求标识中间件已接入",
			Related: "internal/http/requestid.go",
		},
		{
			Name:    "请求超时",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "timeout.go")) && strings.Contains(envFile, "REQUEST_TIMEOUT="),
			Detail:  "支持 REQUEST_TIMEOUT 配置",
			Related: "internal/http/timeout.go",
		},
		{
			Name:    "指标采集",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "metrics.go")) && strings.Contains(router, `mux.HandleFunc("/metrics"`) && strings.Contains(middleware, "metricsMiddleware"),
			Detail:  "包含 /metrics 与请求计数",
			Related: "internal/http/metrics.go",
		},
		{
			Name:    "限流保护",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "ratelimit.go")) && strings.Contains(envFile, "RATE_LIMIT_PER_SECOND="),
			Detail:  "支持 RATE_LIMIT_PER_SECOND 配置",
			Related: "internal/http/ratelimit.go",
		},
		{
			Name:    "跨域支持",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "cors.go")) && strings.Contains(envFile, "CORS_ALLOW_ORIGINS=") && strings.Contains(middleware, "corsMiddleware"),
			Detail:  "支持 CORS_ALLOW_ORIGINS 配置",
			Related: "internal/http/cors.go",
		},
		{
			Name:    "鉴权示例",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "auth.go")) && strings.Contains(router, `mux.HandleFunc("/auth/login"`) && strings.Contains(envFile, "AUTH_TOKEN="),
			Detail:  "包含鉴权示例路由与令牌配置",
			Related: "internal/http/auth.go",
		},
		{
			Name:    "Webhook 示例",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "webhook.go")) && strings.Contains(router, `mux.HandleFunc("/webhooks/example"`),
			Detail:  "包含 webhook 示例接收接口",
			Related: "internal/http/webhook.go",
		},
		{
			Name:    "PostgreSQL 检查",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "store", "postgres.go")) && strings.Contains(router, `mux.HandleFunc("/db/readyz"`) && strings.Contains(envFile, "DATABASE_URL="),
			Detail:  "包含数据库检查接口和环境变量",
			Related: "internal/store/postgres.go",
		},
		{
			Name:    "Worker 入口",
			Exists:  fileExists(filepath.Join(projectDir, "cmd", "worker", "main.go")) && strings.Contains(makefile, "run-worker:"),
			Detail:  "包含独立 worker 入口与命令",
			Related: "cmd/worker/main.go",
		},
		{
			Name:    "数据库检查命令",
			Exists:  fileExists(filepath.Join(projectDir, "cmd", "dbcheck", "main.go")) && strings.Contains(makefile, "db-check:"),
			Detail:  "包含独立数据库检查命令",
			Related: "cmd/dbcheck/main.go",
		},
	}
}

func MissingCapabilities(projectDir string) []Capability {
	var missing []Capability
	for _, capability := range Capabilities(projectDir) {
		if !capability.Exists {
			missing = append(missing, capability)
		}
	}
	return missing
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readMaybe(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func ConfigRequirements(projectDir string) []ConfigRequirement {
	envValues := parseEnvTemplate(readMaybe(filepath.Join(projectDir, ".env.example")))
	var items []ConfigRequirement

	items = append(items,
		buildConfigRequirement(envValues, "PORT", "基础服务端口", validatePortValue),
		buildConfigRequirement(envValues, "SHUTDOWN_TIMEOUT", "优雅停机超时", validateDurationValue),
		buildConfigRequirement(envValues, "LOG_LEVEL", "日志级别", validateLogLevelValue),
	)

	middleware := readMaybe(filepath.Join(projectDir, "internal", "http", "middleware.go"))
	router := readMaybe(filepath.Join(projectDir, "internal", "http", "router.go"))

	if fileExists(filepath.Join(projectDir, "internal", "http", "timeout.go")) || strings.Contains(middleware, "timeoutMiddleware") {
		items = append(items, buildConfigRequirement(envValues, "REQUEST_TIMEOUT", "请求超时", validateDurationValue))
	}
	if fileExists(filepath.Join(projectDir, "internal", "http", "ratelimit.go")) || strings.Contains(middleware, "rateLimitMiddleware") {
		items = append(items, buildConfigRequirement(envValues, "RATE_LIMIT_PER_SECOND", "每秒限流阈值", validatePositiveIntValue))
	}
	if fileExists(filepath.Join(projectDir, "internal", "http", "cors.go")) || strings.Contains(middleware, "corsMiddleware") {
		items = append(items, buildConfigRequirement(envValues, "CORS_ALLOW_ORIGINS", "允许跨域来源", validateNonEmptyValue))
	}
	if fileExists(filepath.Join(projectDir, "internal", "http", "auth.go")) || strings.Contains(router, `mux.HandleFunc("/auth/login"`) {
		items = append(items, buildConfigRequirement(envValues, "AUTH_TOKEN", "鉴权令牌", validateNonEmptyValue))
	}
	if fileExists(filepath.Join(projectDir, "internal", "store", "postgres.go")) || strings.Contains(router, `mux.HandleFunc("/db/readyz"`) {
		items = append(items,
			buildConfigRequirement(envValues, "DATABASE_URL", "数据库地址", validateNonEmptyValue),
			buildConfigRequirement(envValues, "DATABASE_TIMEOUT", "数据库检查超时", validateDurationValue),
		)
	}

	return items
}

func MissingOrInvalidConfig(projectDir string) []ConfigRequirement {
	var items []ConfigRequirement
	for _, item := range ConfigRequirements(projectDir) {
		if !item.Exists || !item.Valid {
			items = append(items, item)
		}
	}
	return items
}

func parseEnvTemplate(content string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return values
}

func buildConfigRequirement(values map[string]string, key, detail string, validator func(string) bool) ConfigRequirement {
	value, ok := values[key]
	if !ok {
		return ConfigRequirement{Name: key, Exists: false, Valid: false, Detail: detail}
	}
	return ConfigRequirement{Name: key, Exists: true, Valid: validator(value), Value: value, Detail: detail}
}

func validatePortValue(value string) bool {
	value = strings.TrimSpace(value)
	if !validatePositiveIntValue(value) {
		return false
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	return port >= 1 && port <= 65535
}

func validateDurationValue(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return false
	}
	return duration > 0
}

func validateLogLevelValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func validatePositiveIntValue(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	return number > 0
}

func validateNonEmptyValue(value string) bool {
	return strings.TrimSpace(value) != ""
}
