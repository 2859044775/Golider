package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateProjectIncludesLifecycleAndValidation(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "demo")

	err := CreateProject(Options{
		AppName:     "demo",
		Module:      "github.com/acme/demo",
		TargetDir:   projectDir,
		DefaultPort: "8080",
	})
	if err != nil {
		t.Fatalf("创建项目失败: %v", err)
	}

	for _, path := range []string{
		filepath.Join(projectDir, "internal", "app", "app.go"),
		filepath.Join(projectDir, "internal", "config", "config.go"),
		filepath.Join(projectDir, "cmd", "api", "main.go"),
		filepath.Join(projectDir, "internal", "http", "errors.go"),
		filepath.Join(projectDir, "internal", "http", "requestid.go"),
		filepath.Join(projectDir, "internal", "http", "readiness.go"),
		filepath.Join(projectDir, "internal", "http", "timeout.go"),
		filepath.Join(projectDir, "internal", "http", "router_test.go"),
		filepath.Join(projectDir, "internal", "http", "middleware_test.go"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("缺少生成文件 %s: %v", path, err)
		}
	}

	configFile := readFile(t, filepath.Join(projectDir, "internal", "config", "config.go"))
	if !strings.Contains(configFile, "func Load() (Config, error)") {
		t.Fatalf("config.go 未生成新的 Load 签名: %s", configFile)
	}
	if !strings.Contains(configFile, "func validate(cfg Config) error") {
		t.Fatalf("config.go 未生成配置校验函数: %s", configFile)
	}
	if !strings.Contains(configFile, "LOG_LEVEL 必须是 debug、info、warn、error 之一") {
		t.Fatalf("config.go 未生成日志级别校验: %s", configFile)
	}

	appFile := readFile(t, filepath.Join(projectDir, "internal", "app", "app.go"))
	if !strings.Contains(appFile, "func (a *App) OnStart") {
		t.Fatalf("app.go 未生成 OnStart: %s", appFile)
	}
	if !strings.Contains(appFile, "func (a *App) OnStop") {
		t.Fatalf("app.go 未生成 OnStop: %s", appFile)
	}

	mainFile := readFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"))
	if !strings.Contains(mainFile, "lifecycle := app.New()") {
		t.Fatalf("main.go 未接入生命周期装配: %s", mainFile)
	}
	if !strings.Contains(mainFile, "cfg, err := config.Load()") {
		t.Fatalf("main.go 未接入配置校验: %s", mainFile)
	}
	if !strings.Contains(mainFile, "httptransport.MarkNotReady(\"shutting_down\")") {
		t.Fatalf("main.go 未接入停机摘流: %s", mainFile)
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	for _, fragment := range []string{
		"handler = requestIDMiddleware(handler)",
		"handler = timeoutMiddleware(handler)",
		"writeError(w, r, http.StatusInternalServerError, \"internal_server_error\", \"internal server error\")",
	} {
		if !strings.Contains(middlewareFile, fragment) {
			t.Fatalf("middleware.go 缺少默认生产能力片段 %q: %s", fragment, middlewareFile)
		}
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	if !strings.Contains(envFile, "REQUEST_TIMEOUT=5s") {
		t.Fatalf(".env.example 未生成 REQUEST_TIMEOUT: %s", envFile)
	}

	routerTestFile := readFile(t, filepath.Join(projectDir, "internal", "http", "router_test.go"))
	if !strings.Contains(routerTestFile, "TestNewRouterDefaultEndpoints") {
		t.Fatalf("router_test.go 未生成默认接口测试: %s", routerTestFile)
	}
	if !strings.Contains(routerTestFile, "TestReadyHandlerDrainingState") {
		t.Fatalf("router_test.go 未生成就绪摘流测试: %s", routerTestFile)
	}

	middlewareTestFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware_test.go"))
	for _, fragment := range []string{
		"TestRecoverMiddlewareWritesErrorResponse",
		"TestTimeoutMiddleware",
	} {
		if !strings.Contains(middlewareTestFile, fragment) {
			t.Fatalf("middleware_test.go 缺少测试片段 %q: %s", fragment, middlewareTestFile)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	return string(content)
}
