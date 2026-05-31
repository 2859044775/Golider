package check

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCapabilities(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"), "package main\n")
	writeFile(t, filepath.Join(projectDir, "internal", "app", "app.go"), `package app

import "context"

type App struct{}

func (a *App) OnStart(name string, fn func(context.Context) error) {}
func (a *App) OnStop(name string, fn func(context.Context) error) {}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "app", "dependencies.go"), `package app

type Dependencies struct{}

func newDependencies() {
	_ = repository.NewInMemoryMessageRepository()
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "config", "config.go"), `package config

type Config struct{}

func Load() (Config, error) { return Config{}, nil }
func validate(cfg Config) error { return nil }
`)
	writeFile(t, filepath.Join(projectDir, "internal", "observability", "logger.go"), "package observability\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "readiness.go"), "package http\nfunc MarkNotReady(string) {}\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "binding.go"), "package http\nfunc decodeJSON(any, any) error { return nil }\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "query.go"), "package http\nfunc parseListQuery(any, int, int) (any, error) { return nil, nil }\n")
	writeFile(t, filepath.Join(projectDir, "internal", "service", "message.go"), `package service

type CreateMessageOutput struct {
	IdempotencyReplay bool
}

type MessageRepository interface{}

type Message struct {
	DeletedAt any
}

func (s *MessageService) GetByID() {}
func (s *MessageService) Update() {}
func (s *MessageService) Delete() {}
func canTransitionMessageStatus() {}

func NewMessageService() any { return nil }
`)
	writeFile(t, filepath.Join(projectDir, "internal", "repository", "message.go"), `package repository

type InMemoryMessageRepository struct{}

func NewInMemoryMessageRepository() *InMemoryMessageRepository { return nil }
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = corsMiddleware(handler)
	handler = requestIDMiddleware(handler)
	handler = timeoutMiddleware(handler)
	handler = metricsMiddleware(handler)
	handler = rateLimitMiddleware(handler)
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusInternalServerError, "internal_server_error", "internal server error")
	})
}

func requestLogMiddleware(next http.Handler) http.Handler { return next }
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import "net/http"

func NewRouter(deps app.Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", nil)
	mux.HandleFunc("/readyz", nil)
	mux.HandleFunc("/messages", listMessagesHandler)
	mux.HandleFunc("/messages/", getMessageHandler)
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/auth/login", loginExampleHandler)
	mux.HandleFunc("/webhooks/example", exampleWebhookHandler)
	mux.HandleFunc("/db/readyz", postgresReadyHandler)
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func listMessagesHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = parseListQuery(r, 10, 100)
	_ = deps.MessageService.List(
}

func createMessageHandler(w http.ResponseWriter, r *http.Request) {
	_ = decodeJSON(r, nil)
	_ = validateCreateMessageRequest(nil, r.Header.Get("Idempotency-Key"))
	_ = deps.MessageService.Create(
	_ = "idempotency_key_conflict"
}

func getMessageHandler(w http.ResponseWriter, r *http.Request) {
	_ = deps.MessageService.GetByID(
}

func updateMessageHandler(w http.ResponseWriter, r *http.Request) {
	_ = deps.MessageService.Update(
	_ = "message_status_transition_invalid"
}

func deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	_ = deps.MessageService.Delete(
	_ = `+"`"+`DeletedAt`+"`"+`
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	_ = decodeJSON(r, nil)
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "errors.go"), "package http\nfunc writeError() {}\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "requestid.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "timeout.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "metrics.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "ratelimit.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "cors.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "auth.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "webhook.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "store", "postgres.go"), "package store\n")
	writeFile(t, filepath.Join(projectDir, "cmd", "worker", "main.go"), "package main\n")
	writeFile(t, filepath.Join(projectDir, "cmd", "dbcheck", "main.go"), "package main\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\nSHUTDOWN_TIMEOUT=10s\nLOG_LEVEL=info\nREQUEST_TIMEOUT=5s\nHTTP_READ_HEADER_TIMEOUT=2s\nHTTP_READ_TIMEOUT=10s\nHTTP_WRITE_TIMEOUT=15s\nHTTP_IDLE_TIMEOUT=60s\nDEFAULT_PAGE_SIZE=10\nMAX_PAGE_SIZE=100\nRATE_LIMIT_PER_SECOND=20\nCORS_ALLOW_ORIGINS=*\nAUTH_TOKEN=dev-token\nDATABASE_URL=postgres://demo\n")
	writeFile(t, filepath.Join(projectDir, ".gitignore"), "bin/\n")
	writeFile(t, filepath.Join(projectDir, "Dockerfile"), "FROM golang:1.20\n")
	writeFile(t, filepath.Join(projectDir, ".github", "workflows", "ci.yml"), "name: ci\n")
	writeFile(t, filepath.Join(projectDir, "Makefile"), "run-worker:\n\tgo run ./cmd/worker\n\ndb-check:\n\tgo run ./cmd/dbcheck\n")
	writeFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"), "package main\n\nfunc main() {\n\tlifecycle.OnStop(\"http-server\", nil)\n\tMarkNotReady(\"shutting_down\")\n\t_ = `ReadHeaderTimeout: cfg.ReadHeaderTimeout`\n\t_ = `WriteTimeout:      cfg.WriteTimeout`\n\t_ = `deps := app.NewDependencies(cfg)`\n\t_ = `httptransport.NewRouter(deps)`\n}\n")

	capabilities := Capabilities(projectDir)
	if len(capabilities) == 0 {
		t.Fatal("能力列表不应为空")
	}

	for _, capability := range capabilities {
		if !capability.Exists {
			t.Fatalf("能力 %s 应该被识别为存在，详情：%s", capability.Name, capability.Detail)
		}
	}
}

func TestMissingOrInvalidConfig(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=70000\nSHUTDOWN_TIMEOUT=0s\nLOG_LEVEL=verbose\nREQUEST_TIMEOUT=1s\nHTTP_READ_HEADER_TIMEOUT=2s\nHTTP_READ_TIMEOUT=10s\nHTTP_WRITE_TIMEOUT=15s\nHTTP_IDLE_TIMEOUT=60s\nDEFAULT_PAGE_SIZE=10\nMAX_PAGE_SIZE=100\nRATE_LIMIT_PER_SECOND=20\nCORS_ALLOW_ORIGINS=*\nAUTH_TOKEN=dev-token\nDATABASE_URL=postgres://demo\nDATABASE_TIMEOUT=3s\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	handler = timeoutMiddleware(handler)
	handler = rateLimitMiddleware(handler)
	handler = corsMiddleware(handler)
	return handler
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", nil)
	mux.HandleFunc("/db/readyz", nil)
	return mux
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "timeout.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "ratelimit.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "cors.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "auth.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "store", "postgres.go"), "package store\n")

	items := MissingOrInvalidConfig(projectDir)
	if len(items) != 3 {
		t.Fatalf("应识别出 3 项非法配置，实际为 %d", len(items))
	}

	got := map[string]bool{}
	for _, item := range items {
		got[item.Name] = true
	}

	for _, name := range []string{"PORT", "SHUTDOWN_TIMEOUT", "LOG_LEVEL"} {
		if !got[name] {
			t.Fatalf("缺少配置项 %s 的非法识别", name)
		}
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
