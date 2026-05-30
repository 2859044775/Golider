package scaffold

func files() map[string]string {
	return map[string]string{
		".env.example":                     envExampleTemplate,
		".gitignore":                       gitignoreTemplate,
		".github/workflows/ci.yml":         ciTemplate,
		"Dockerfile":                       dockerfileTemplate,
		"Makefile":                         makefileTemplate,
		"README.md":                        projectReadmeTemplate,
		"go.mod":                           goModTemplate,
		"cmd/api/main.go":                  apiMainTemplate,
		"internal/app/app.go":              appTemplate,
		"internal/config/config.go":        configTemplate,
		"internal/http/errors.go":          errorModelTemplate,
		"internal/http/middleware.go":      middlewareTemplate,
		"internal/http/middleware_test.go": middlewareTestTemplate,
		"internal/http/readiness.go":       readinessTemplate,
		"internal/http/router.go":          routerTemplate,
		"internal/http/router_test.go":     routerTestTemplate,
		"internal/http/requestid.go":       requestIDTemplate,
		"internal/http/timeout.go":         timeoutTemplate,
		"internal/observability/logger.go": loggerTemplate,
	}
}

const goModTemplate = `module {{ .Module }}

go 1.20
`

const envExampleTemplate = `PORT={{ .DefaultPort }}
SHUTDOWN_TIMEOUT=10s
LOG_LEVEL=info
REQUEST_TIMEOUT=5s
`

const makefileTemplate = `run:
	go run ./cmd/api

build:
	go build ./...

test:
	go test ./...

verify:
	go test ./...
	go build ./...

# Golider 扩展命令锚点
`

const projectReadmeTemplate = `# {{ .ProjectTitle }}

由 Golider 生成的最小可运行 Go API 工程，默认包含健康检查、请求标识、请求超时、统一错误模型、配置校验、生命周期钩子、就绪摘流、基础测试、Dockerfile 和 GitHub Actions。

## 启动

1. 复制环境变量模板
2. 执行服务

` + "```bash" + `
cp .env.example .env
make run
` + "```" + `

## 默认接口

- ` + "`GET /healthz`" + `
- ` + "`GET /readyz`" + `
- ` + "`GET /`" + `

## 默认验证

` + "```bash" + `
make test
make verify
` + "```" + `
`

const configTemplate = `package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var allowedLogLevels = map[string]struct{}{
	"debug": {},
	"info":  {},
	"warn":  {},
	"error": {},
}

type Config struct {
	Port            string
	ShutdownTimeout time.Duration
	LogLevel        string
}

func Load() (Config, error) {
	port := getenv("PORT", "{{ .DefaultPort }}")
	timeout := getenv("SHUTDOWN_TIMEOUT", "10s")
	logLevel := strings.ToLower(getenv("LOG_LEVEL", "info"))

	d, err := time.ParseDuration(timeout)
	if err != nil {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT 解析失败：%w", err)
	}

	cfg := Config{
		Port:            port,
		ShutdownTimeout: d,
		LogLevel:        logLevel,
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if strings.TrimSpace(cfg.Port) == "" {
		return fmt.Errorf("PORT 不能为空")
	}

	portNumber, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return fmt.Errorf("PORT 必须是数字：%w", err)
	}
	if portNumber <= 0 || portNumber > 65535 {
		return fmt.Errorf("PORT 必须在 1 到 65535 之间")
	}

	if cfg.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT 必须大于 0")
	}

	if _, ok := allowedLogLevels[cfg.LogLevel]; !ok {
		return fmt.Errorf("LOG_LEVEL 必须是 debug、info、warn、error 之一")
	}

	return nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
`

const appTemplate = `package app

import (
	"context"
	"fmt"
	"strings"
)

type HookFunc func(context.Context) error

type hook struct {
	name string
	run  HookFunc
}

type App struct {
	startHooks []hook
	stopHooks  []hook
}

func New() *App {
	return &App{}
}

func (a *App) OnStart(name string, fn HookFunc) {
	if fn == nil {
		return
	}
	a.startHooks = append(a.startHooks, hook{name: name, run: fn})
}

func (a *App) OnStop(name string, fn HookFunc) {
	if fn == nil {
		return
	}
	a.stopHooks = append(a.stopHooks, hook{name: name, run: fn})
}

func (a *App) Start(ctx context.Context) error {
	for _, item := range a.startHooks {
		if err := item.run(ctx); err != nil {
			return fmt.Errorf("启动钩子 %s 执行失败：%w", item.name, err)
		}
	}
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	var messages []string
	for idx := len(a.stopHooks) - 1; idx >= 0; idx-- {
		item := a.stopHooks[idx]
		if err := item.run(ctx); err != nil {
			messages = append(messages, item.name+": "+err.Error())
		}
	}

	if len(messages) > 0 {
		return fmt.Errorf("停止钩子执行失败：%s", strings.Join(messages, "; "))
	}

	return nil
}
`

const middlewareTemplate = `package http

import (
	"net/http"
	"time"

	"{{ .Module }}/internal/observability"
)

var httpLogger = observability.New("http")
var recordRecovery = func() {}

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = requestIDMiddleware(handler)
	handler = timeoutMiddleware(handler)
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func requestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		httpLogger.Info("请求完成", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String(), "request_id", requestIDFromRequest(r))
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				recordRecovery()
				httpLogger.Error("请求异常恢复", "panic", rec, "request_id", requestIDFromRequest(r))
				writeError(w, r, http.StatusInternalServerError, "internal_server_error", "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
`

const routerTemplate = `package http

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"service": "{{ .AppName }}", "status": "running"})
	})
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readyHandler)
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
`

const routerTestTemplate = `package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRouterDefaultEndpoints(t *testing.T) {
	router := NewRouter()

	for _, item := range []struct {
		name string
		path string
	}{
		{name: "根路径", path: "/"},
		{name: "健康检查", path: "/healthz"},
		{name: "就绪检查", path: "/readyz"},
	} {
		req := httptest.NewRequest(http.MethodGet, item.path, nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s 状态码错误，期望 %d，实际 %d", item.name, http.StatusOK, recorder.Code)
		}
		if recorder.Header().Get("X-Request-Id") == "" {
			t.Fatalf("%s 未写入请求标识头", item.name)
		}
	}
}

func TestReadyHandlerDrainingState(t *testing.T) {
	markReady()
	t.Cleanup(func() {
		markReady()
	})

	router := NewRouter()
	markNotReady("shutting_down")

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if recorder.Header().Get("X-Request-Id") == "" {
		t.Fatal("就绪摘流响应未写入请求标识头")
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"status\":\"not_ready\"", "\"reason\":\"shutting_down\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("就绪摘流响应缺少片段 %q: %s", fragment, body)
		}
	}
}
`

const middlewareTestTemplate = `package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRecoverMiddlewareWritesErrorResponse(t *testing.T) {
	handler := requestIDMiddleware(recoverMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusInternalServerError, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"internal_server_error\"", "\"message\":\"internal server error\"", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("错误响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	previous := os.Getenv("REQUEST_TIMEOUT")
	t.Cleanup(func() {
		if previous == "" {
			_ = os.Unsetenv("REQUEST_TIMEOUT")
			return
		}
		_ = os.Setenv("REQUEST_TIMEOUT", previous)
	})
	if err := os.Setenv("REQUEST_TIMEOUT", "1ms"); err != nil {
		t.Fatalf("设置 REQUEST_TIMEOUT 失败: %v", err)
	}

	handler := timeoutMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "request timeout") {
		t.Fatalf("超时响应内容错误: %s", recorder.Body.String())
	}
}
`

const readinessTemplate = `package http

import (
	"net/http"
	"strings"
	"sync/atomic"
)

type readinessState struct {
	ready  atomic.Bool
	reason atomic.Value
}

var readiness = newReadinessState()

func newReadinessState() *readinessState {
	state := &readinessState{}
	state.markReady()
	return state
}

func markReady() {
	readiness.markReady()
}

func markNotReady(reason string) {
	readiness.markNotReady(reason)
}

func MarkReady() {
	markReady()
}

func MarkNotReady(reason string) {
	markNotReady(reason)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	ready, reason := readiness.status()
	if !ready {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": reason})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *readinessState) markReady() {
	s.ready.Store(true)
	s.reason.Store("ready")
}

func (s *readinessState) markNotReady(reason string) {
	if strings.TrimSpace(reason) == "" {
		reason = "not_ready"
	}
	s.ready.Store(false)
	s.reason.Store(reason)
}

func (s *readinessState) status() (bool, string) {
	reason, _ := s.reason.Load().(string)
	return s.ready.Load(), reason
}
`

const requestIDTemplate = `package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"{{ .Module }}/internal/observability"
)

type requestIDKey struct{}

var requestIDLogger = observability.New("request-id")

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = newRequestID()
			r.Header.Set("X-Request-Id", requestID)
		}

		w.Header().Set("X-Request-Id", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		requestIDLogger.Info("请求标识已注入", "request_id", requestID, "path", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "fallback-request-id"
	}
	return hex.EncodeToString(buf)
}

func requestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey{}).(string)
	return value
}
`

const timeoutTemplate = `package http

import (
	"net/http"
	"os"
	"time"

	"{{ .Module }}/internal/observability"
)

var timeoutLogger = observability.New("timeout")

func timeoutMiddleware(next http.Handler) http.Handler {
	timeout := 5 * time.Second
	if raw := os.Getenv("REQUEST_TIMEOUT"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			timeout = parsed
		}
	}

	timeoutLogger.Info("启用请求超时中间件", "timeout", timeout.String())
	return http.TimeoutHandler(next, timeout, "{\"error\":\"request timeout\"}\n")
}
`

const errorModelTemplate = `package http

import "net/http"

type errorResponse struct {
	Code      string ` + "`json:\"code\"`" + `
	Message   string ` + "`json:\"message\"`" + `
	RequestID string ` + "`json:\"request_id,omitempty\"`" + `
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	writeJSON(w, statusCode, errorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestIDFromRequest(r),
	})
}

func requestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	return r.Header.Get("X-Request-Id")
}
`

const apiMainTemplate = `package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"{{ .Module }}/internal/app"
	"{{ .Module }}/internal/config"
	httptransport "{{ .Module }}/internal/http"
	"{{ .Module }}/internal/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败：%v\n", err)
		os.Exit(1)
	}

	observability.SetLevel(cfg.LogLevel)
	logger := observability.New("api")
	httptransport.MarkReady()
	handler := httptransport.NewRouter()

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	lifecycle := app.New()
	errCh := make(chan error, 1)
	lifecycle.OnStart("http-server", func(context.Context) error {
		httptransport.MarkReady()
		logger.Info("服务启动中", "port", cfg.Port, "log_level", cfg.LogLevel)
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()
		return nil
	})
	lifecycle.OnStop("http-server", func(ctx context.Context) error {
		httptransport.MarkNotReady("shutting_down")
		logger.Info("服务开始摘流")
		logger.Info("开始优雅停机")
		return server.Shutdown(ctx)
	})

	if err := lifecycle.Start(context.Background()); err != nil {
		logger.Error("服务启动失败", "error", err.Error())
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Error("服务异常退出", "error", err.Error())
		os.Exit(1)
	case <-sigCh:
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := lifecycle.Stop(ctx); err != nil {
		logger.Error("服务停机失败", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("服务已停止")
}
`

const loggerTemplate = `package observability

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type Logger struct {
	component string
}

var std = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
var (
	levelMu      sync.RWMutex
	currentLevel = "info"
)

func New(component string) *Logger {
	return &Logger{component: component}
}

func SetLevel(level string) {
	levelMu.Lock()
	defer levelMu.Unlock()

	if level == "" {
		currentLevel = "info"
		return
	}

	currentLevel = strings.ToLower(level)
}

func (l *Logger) Info(message string, fields ...any) {
	l.log("info", message, fields...)
}

func (l *Logger) Error(message string, fields ...any) {
	l.log("error", message, fields...)
}

func (l *Logger) log(level string, message string, fields ...any) {
	if !enabled(level) {
		return
	}
	std.Printf("level=%s component=%s message=%q %s", level, l.component, message, formatFields(fields...))
}

func enabled(level string) bool {
	levelMu.RLock()
	defer levelMu.RUnlock()

	weights := map[string]int{
		"debug": 10,
		"info":  20,
		"warn":  30,
		"error": 40,
	}

	currentWeight, ok := weights[currentLevel]
	if !ok {
		currentWeight = weights["info"]
	}

	levelWeight, ok := weights[level]
	if !ok {
		levelWeight = weights["info"]
	}

	return levelWeight >= currentWeight
}

func formatFields(fields ...any) string {
	if len(fields) == 0 {
		return ""
	}

	parts := make([]string, 0, len(fields)/2+1)
	for idx := 0; idx < len(fields); idx += 2 {
		key := fmt.Sprintf("field_%d", idx)
		value := "<missing>"
		if idx < len(fields) {
			key = fmt.Sprintf("%v", fields[idx])
		}
		if idx+1 < len(fields) {
			value = fmt.Sprintf("%v", fields[idx+1])
		}
		parts = append(parts, key+"="+value)
	}

	return strings.Join(parts, " ")
}
`

const dockerfileTemplate = `FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /bin/api /app/api

EXPOSE {{ .DefaultPort }}

CMD ["/app/api"]
`

const ciTemplate = `name: ci

on:
  push:
    branches:
      - main
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version:
          - "1.20"
          - "1.22"
    steps:
      - name: 拉取代码
        uses: actions/checkout@v4

      - name: 安装 Go
        uses: actions/setup-go@v5
        with:
          go-version: {{ "${{ matrix.go-version }}" }}

      - name: 安装依赖
        run: go mod download

      - name: 执行测试
        run: go test ./...

      - name: 执行构建
        run: go build ./...
`

const gitignoreTemplate = `.DS_Store
.env
bin/
dist/
coverage.out
`
