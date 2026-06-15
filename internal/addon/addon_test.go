package addon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallWorkerModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\nSHUTDOWN_TIMEOUT=10s\n")
	writeFile(t, filepath.Join(projectDir, "Makefile"), "run:\n\tgo run ./cmd/api\n\n# Golider 扩展命令锚点\n")

	err := Install(Options{
		ModuleName: "worker",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 worker 模块失败: %v", err)
	}

	workerMain := readFile(t, filepath.Join(projectDir, "cmd", "worker", "main.go"))
	if !strings.Contains(workerMain, "\"github.com/acme/demo/internal/worker\"") {
		t.Fatalf("worker 入口未写入正确导入路径: %s", workerMain)
	}
	for _, fragment := range []string{
		"\"github.com/acme/demo/internal/app\"",
		"\"github.com/acme/demo/internal/config\"",
		"lifecycle := app.New()",
		"lifecycle.OnStart(\"worker\", instance.StartHook())",
		"lifecycle.OnStop(\"worker\", instance.StopHook())",
	} {
		if !strings.Contains(workerMain, fragment) {
			t.Fatalf("worker 入口缺少生命周期片段 %q: %s", fragment, workerMain)
		}
	}

	makefile := readFile(t, filepath.Join(projectDir, "Makefile"))
	if !strings.Contains(makefile, "run-worker:\n\tgo run ./cmd/worker\n") {
		t.Fatalf("Makefile 未追加 worker 命令: %s", makefile)
	}

	lifecycleFile := readFile(t, filepath.Join(projectDir, "internal", "worker", "lifecycle.go"))
	if !strings.Contains(lifecycleFile, "func (w *Worker) StartHook()") || !strings.Contains(lifecycleFile, "func (w *Worker) StopHook()") {
		t.Fatalf("worker 生命周期文件未生成: %s", lifecycleFile)
	}

	appFile := readFile(t, filepath.Join(projectDir, "internal", "app", "app.go"))
	if !strings.Contains(appFile, "func (a *App) OnStart") || !strings.Contains(appFile, "func (a *App) OnStop") {
		t.Fatalf("worker 模块未补齐应用生命周期基础文件: %s", appFile)
	}

	configFile := readFile(t, filepath.Join(projectDir, "internal", "config", "config.go"))
	if !strings.Contains(configFile, "func Load() (Config, error)") {
		t.Fatalf("worker 模块未补齐配置基础文件: %s", configFile)
	}
}

func TestInstallWebhookModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
`)

	err := Install(Options{
		ModuleName: "webhook",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 webhook 模块失败: %v", err)
	}

	router := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	if !strings.Contains(router, "mux.HandleFunc(\"/webhooks/example\", exampleWebhookHandler)") {
		t.Fatalf("router.go 未注入 webhook 路由: %s", router)
	}

	webhookHandler := readFile(t, filepath.Join(projectDir, "internal", "http", "webhook.go"))
	if !strings.Contains(webhookHandler, "func exampleWebhookHandler") {
		t.Fatalf("webhook 处理器未生成: %s", webhookHandler)
	}

	loggerFile := readFile(t, filepath.Join(projectDir, "internal", "observability", "logger.go"))
	if !strings.Contains(loggerFile, "type Logger struct") {
		t.Fatalf("日志基础文件未自动补齐: %s", loggerFile)
	}
}

func TestInstallAuthModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
`)

	err := Install(Options{
		ModuleName: "auth",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 auth 模块失败: %v", err)
	}

	router := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	if !strings.Contains(router, "mux.HandleFunc(\"/auth/login\", loginExampleHandler)") {
		t.Fatalf("router.go 未注入 auth 登录路由: %s", router)
	}
	if !strings.Contains(router, "mux.HandleFunc(\"/auth/profile\", profileExampleHandler)") {
		t.Fatalf("router.go 未注入 auth 资料路由: %s", router)
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	if !strings.Contains(envFile, "AUTH_TOKEN=dev-token") {
		t.Fatalf(".env.example 未追加 AUTH_TOKEN: %s", envFile)
	}

	authFile := readFile(t, filepath.Join(projectDir, "internal", "http", "auth.go"))
	if !strings.Contains(authFile, "func loginExampleHandler") {
		t.Fatalf("auth 处理器未生成: %s", authFile)
	}
}

func TestInstallPostgresModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\nSHUTDOWN_TIMEOUT=10s\nLOG_LEVEL=info\n")
	writeFile(t, filepath.Join(projectDir, "Makefile"), "run:\n\tgo run ./cmd/api\n\n# Golider 扩展命令锚点\n")
	writeFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"), `package main

import (
	"context"
	"os"

	"github.com/acme/demo/internal/app"
)

func main() {
	lifecycle := app.New()
	_ = os.Getenv("PORT")
	_ = context.Background()
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "app", "dependencies.go"), `package app

import (
	"github.com/acme/demo/internal/config"
	"github.com/acme/demo/internal/service"
)

type Dependencies struct {
	MessageService   *service.MessageService
	DefaultPageSize int
	MaxPageSize     int
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
`)

	err := Install(Options{
		ModuleName: "postgres",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 postgres 模块失败: %v", err)
	}

	router := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	if !strings.Contains(router, "mux.HandleFunc(\"/db/readyz\", postgresReadyHandler)") {
		t.Fatalf("router.go 未注入 postgres 路由: %s", router)
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	if !strings.Contains(envFile, "DATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable") {
		t.Fatalf(".env.example 未追加 DATABASE_URL: %s", envFile)
	}
	if !strings.Contains(envFile, "DATABASE_TIMEOUT=3s") {
		t.Fatalf(".env.example 未追加 DATABASE_TIMEOUT: %s", envFile)
	}

	makefile := readFile(t, filepath.Join(projectDir, "Makefile"))
	if !strings.Contains(makefile, "db-check:\n\tgo run ./cmd/dbcheck\n") {
		t.Fatalf("Makefile 未追加 db-check 命令: %s", makefile)
	}

	storeFile := readFile(t, filepath.Join(projectDir, "internal", "store", "postgres.go"))
	if !strings.Contains(storeFile, "func CheckPostgres") {
		t.Fatalf("postgres 存储文件未生成: %s", storeFile)
	}

	storeLifecycleFile := readFile(t, filepath.Join(projectDir, "internal", "store", "lifecycle.go"))
	if !strings.Contains(storeLifecycleFile, "type PostgresManager struct") || !strings.Contains(storeLifecycleFile, "func (m *PostgresManager) Start") {
		t.Fatalf("postgres 生命周期文件未生成: %s", storeLifecycleFile)
	}

	apiMain := readFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"))
	for _, fragment := range []string{
		"\"github.com/acme/demo/internal/store\"",
		"postgresManager := store.NewPostgresManager(os.Getenv(\"DATABASE_URL\"))",
		"lifecycle.OnStart(\"postgres\", postgresManager.Start)",
		"lifecycle.OnStop(\"postgres\", postgresManager.Stop)",
	} {
		if !strings.Contains(apiMain, fragment) {
			t.Fatalf("api 入口缺少 postgres 生命周期片段 %q: %s", fragment, apiMain)
		}
	}

	depsFile := readFile(t, filepath.Join(projectDir, "internal", "app", "dependencies.go"))
	if !strings.Contains(depsFile, "Golider 数据库切换位点") {
		t.Fatalf("dependencies.go 未注入数据库切换位点: %s", depsFile)
	}

	repoFile := readFile(t, filepath.Join(projectDir, "internal", "repository", "message_postgres.go"))
	for _, fragment := range []string{
		"PostgresMessageRepository",
		"NewDatabaseMessageService",
		"SaveVersioned",
		"scanMessage",
	} {
		if !strings.Contains(repoFile, fragment) {
			t.Fatalf("message_postgres.go 缺少片段 %q: %s", fragment, repoFile)
		}
	}

	migrationFile := readFile(t, filepath.Join(projectDir, "migrations", "001_create_messages.sql"))
	if !strings.Contains(migrationFile, "CREATE TABLE IF NOT EXISTS messages") {
		t.Fatalf("迁移文件未正确生成: %s", migrationFile)
	}
}

func TestInstallRequestIDModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func requestLogMiddleware(next http.Handler) http.Handler { return next }
func recoverMiddleware(next http.Handler) http.Handler { return next }
`)

	err := Install(Options{
		ModuleName: "request-id",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 request-id 模块失败: %v", err)
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	if !strings.Contains(middlewareFile, "handler = requestIDMiddleware(handler)") {
		t.Fatalf("middleware.go 未注入 request-id 中间件: %s", middlewareFile)
	}

	requestIDFile := readFile(t, filepath.Join(projectDir, "internal", "http", "requestid.go"))
	if !strings.Contains(requestIDFile, "func requestIDMiddleware") {
		t.Fatalf("request-id 文件未生成: %s", requestIDFile)
	}
}

func TestInstallTimeoutModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func requestLogMiddleware(next http.Handler) http.Handler { return next }
func recoverMiddleware(next http.Handler) http.Handler { return next }
`)

	err := Install(Options{
		ModuleName: "timeout",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 timeout 模块失败: %v", err)
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	if !strings.Contains(middlewareFile, "handler = timeoutMiddleware(handler)") {
		t.Fatalf("middleware.go 未注入 timeout 中间件: %s", middlewareFile)
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	if !strings.Contains(envFile, "REQUEST_TIMEOUT=5s") {
		t.Fatalf(".env.example 未追加 REQUEST_TIMEOUT: %s", envFile)
	}
}

func TestInstallMetricsModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func requestLogMiddleware(next http.Handler) http.Handler { return next }
func recoverMiddleware(next http.Handler) http.Handler { return next }
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
`)

	err := Install(Options{
		ModuleName: "metrics",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 metrics 模块失败: %v", err)
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	if !strings.Contains(middlewareFile, "handler = metricsMiddleware(handler)") {
		t.Fatalf("middleware.go 未注入 metrics 中间件: %s", middlewareFile)
	}

	router := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	if !strings.Contains(router, "mux.HandleFunc(\"/metrics\", metricsHandler)") {
		t.Fatalf("router.go 未注入 metrics 路由: %s", router)
	}
}

func TestInstallRateLimitModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func requestLogMiddleware(next http.Handler) http.Handler { return next }
func recoverMiddleware(next http.Handler) http.Handler { return next }
`)

	err := Install(Options{
		ModuleName: "rate-limit",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 rate-limit 模块失败: %v", err)
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	if !strings.Contains(middlewareFile, "handler = rateLimitMiddleware(handler)") {
		t.Fatalf("middleware.go 未注入 rate-limit 中间件: %s", middlewareFile)
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	if !strings.Contains(envFile, "RATE_LIMIT_PER_SECOND=20") {
		t.Fatalf(".env.example 未追加 RATE_LIMIT_PER_SECOND: %s", envFile)
	}
}

func TestInstallErrorModelModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	})
}
`)

	err := Install(Options{
		ModuleName: "error-model",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 error-model 模块失败: %v", err)
	}

	errorFile := readFile(t, filepath.Join(projectDir, "internal", "http", "errors.go"))
	if !strings.Contains(errorFile, "func writeError") {
		t.Fatalf("errors.go 未生成统一错误输出: %s", errorFile)
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	if !strings.Contains(middlewareFile, "writeError(w, r, http.StatusInternalServerError, \"internal_server_error\", \"internal server error\")") {
		t.Fatalf("middleware.go 未切换为 writeError: %s", middlewareFile)
	}
}

func TestInstallCORSModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func requestLogMiddleware(next http.Handler) http.Handler { return next }
func recoverMiddleware(next http.Handler) http.Handler { return next }
`)

	err := Install(Options{
		ModuleName: "cors",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 cors 模块失败: %v", err)
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	if !strings.Contains(middlewareFile, "handler = corsMiddleware(handler)") {
		t.Fatalf("middleware.go 未注入 cors 中间件: %s", middlewareFile)
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	if !strings.Contains(envFile, "CORS_ALLOW_ORIGINS=*") {
		t.Fatalf(".env.example 未追加 CORS_ALLOW_ORIGINS: %s", envFile)
	}

	corsFile := readFile(t, filepath.Join(projectDir, "internal", "http", "cors.go"))
	if !strings.Contains(corsFile, "func corsMiddleware") {
		t.Fatalf("cors.go 未生成: %s", corsFile)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
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
