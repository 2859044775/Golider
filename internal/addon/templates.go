package addon

func moduleFiles(name string) map[string]string {
	switch name {
	case "docker":
		return map[string]string{
			"Dockerfile": dockerfileTemplate,
		}
	case "ci":
		return map[string]string{
			".github/workflows/ci.yml": ciTemplate,
		}
	case "gitignore":
		return map[string]string{
			".gitignore": gitignoreTemplate,
		}
	case "worker":
		return map[string]string{
			"cmd/worker/main.go":           workerMainTemplate,
			"internal/worker/worker.go":    workerTemplate,
			"internal/worker/lifecycle.go": workerLifecycleTemplate,
		}
	case "webhook":
		return map[string]string{
			"internal/http/webhook.go": webhookTemplate,
		}
	case "auth":
		return map[string]string{
			"internal/http/auth.go": authTemplate,
		}
	case "postgres":
		return map[string]string{
			"cmd/dbcheck/main.go":                     postgresCheckMainTemplate,
			"internal/http/postgres.go":               postgresHTTPTemplate,
			"internal/store/postgres.go":              postgresStoreTemplate,
			"internal/store/lifecycle.go":             postgresLifecycleTemplate,
			"internal/repository/message_postgres.go": postgresMessageRepositoryTemplate,
			"migrations/001_create_messages.sql":      migrationTemplate,
		}
	case "request-id":
		return map[string]string{
			"internal/http/requestid.go": requestIDTemplate,
		}
	case "timeout":
		return map[string]string{
			"internal/http/timeout.go": timeoutTemplate,
		}
	case "metrics":
		return map[string]string{
			"internal/http/metrics.go": metricsTemplate,
		}
	case "rate-limit":
		return map[string]string{
			"internal/http/ratelimit.go": rateLimitTemplate,
		}
	case "error-model":
		return map[string]string{
			"internal/http/errors.go": errorModelTemplate,
		}
	case "cors":
		return map[string]string{
			"internal/http/cors.go": corsTemplate,
		}
	case "circuit-breaker":
		return map[string]string{
			"internal/http/circuitbreaker.go": circuitBreakerTemplate,
		}
	case "websocket":
		return map[string]string{
			"internal/http/websocket.go": websocketTemplate,
		}
	case "scheduler":
		return map[string]string{
			"internal/scheduler/scheduler.go":  schedulerTemplate,
			"internal/scheduler/lifecycle.go": schedulerLifecycleTemplate,
			"internal/http/scheduler.go":       schedulerHTTPTemplate,
		}
	case "redis":
		return map[string]string{
			"internal/store/redis.go": redisStoreTemplate,
			"internal/http/redis.go":  redisHTTPTemplate,
		}
	case "grpc":
		return map[string]string{
			"proto/service.proto":      grpcProtoTemplate,
			"cmd/grpc/main.go":         grpcMainTemplate,
			"internal/grpc/server.go":  grpcServerTemplate,
			"internal/grpc/service.go": grpcServiceTemplate,
		}
	case "kafka":
		return map[string]string{
			"cmd/kafka/main.go":           kafkaMainTemplate,
			"internal/kafka/consumer.go":  kafkaConsumerTemplate,
			"internal/kafka/producer.go":  kafkaProducerTemplate,
			"internal/kafka/lifecycle.go": kafkaLifecycleTemplate,
		}
	default:
		// 内置模块未命中，检查自定义注册模块
		for _, m := range registeredModules {
			if m.Name == name {
				return m.Files
			}
		}
		return map[string]string{}
	}
}

func baseFiles(name string) map[string]string {
	switch name {
	case "worker", "postgres", "grpc", "kafka":
		return map[string]string{
			"internal/observability/logger.go": addonLoggerTemplate,
			"internal/config/config.go":        addonConfigTemplate,
			"internal/app/app.go":              addonAppTemplate,
		}
	case "webhook", "auth", "metrics", "rate-limit", "cors", "redis", "circuit-breaker", "websocket", "scheduler":
		return map[string]string{
			"internal/observability/logger.go": addonLoggerTemplate,
		}
	default:
		// 内置模块未命中，检查自定义注册模块
		for _, m := range registeredModules {
			if m.Name == name {
				return m.BaseFiles
			}
		}
		return map[string]string{}
	}
}

// CommonBaseFiles 返回自定义模块常用的基础文件（日志模块），供 generated module 使用
func CommonBaseFiles() map[string]string {
	return map[string]string{
		"internal/observability/logger.go": addonLoggerTemplate,
	}
}

const dockerfileTemplate = `FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1001 app

WORKDIR /app

COPY --from=builder /bin/api /app/api

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

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
.env.*
*.local
bin/
dist/
coverage.out
coverage.txt
coverage.html
*.test
*.exe
*.exe~
*.dll
*.so
*.dylib
*.swp
*.swo
*~
.idea/
.vscode/
vendor/
*.log
`

const workerMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"{{ .ModulePath }}/internal/app"
	"{{ .ModulePath }}/internal/config"
	"{{ .ModulePath }}/internal/observability"
	"{{ .ModulePath }}/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败：%v\n", err)
		os.Exit(1)
	}

	observability.SetLevel(cfg.LogLevel)
	logger := observability.New("worker")
	instance := worker.New("{{ .ProjectName }}")
	lifecycle := app.New()
	lifecycle.OnStart("worker", instance.StartHook())
	lifecycle.OnStop("worker", instance.StopHook())

	if err := lifecycle.Start(context.Background()); err != nil {
		logger.Error("worker 启动失败", "error", err.Error())
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := lifecycle.Stop(ctx); err != nil {
		logger.Error("worker 停止失败", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("worker 已停止")
}
`

const workerTemplate = `package worker

import (
	"sync"
	"time"

	"{{ .ModulePath }}/internal/observability"
)

var workerLogger = observability.New("worker")

type Worker struct {
	name string
	stop chan struct{}
	done chan struct{}
	stopOnce sync.Once
}

func New(name string) *Worker {
	return &Worker{
		name: name,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
}

func (w *Worker) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		defer close(w.done)

		workerLogger.Info("worker 已启动", "name", w.name)
		for {
			select {
			case <-ticker.C:
				workerLogger.Info("worker 执行示例任务", "name", w.name)
			case <-w.stop:
				workerLogger.Info("worker 收到停止请求", "name", w.name)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stop)
	})
	<-w.done
	workerLogger.Info("worker 已停止", "name", w.name)
}
`

const workerLifecycleTemplate = `package worker

import "context"

func (w *Worker) StartHook() func(context.Context) error {
	return func(context.Context) error {
		w.Start()
		return nil
	}
}

func (w *Worker) StopHook() func(context.Context) error {
	return func(context.Context) error {
		w.Stop()
		return nil
	}
}
`

const webhookTemplate = `package http

import (
	"net/http"

	"{{ .ModulePath }}/internal/observability"
)

var webhookLogger = observability.New("webhook")

func exampleWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		webhookLogger.Error("webhook 请求方法不允许", "method", r.Method)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	webhookLogger.Info("webhook 已接收", "path", r.URL.Path)
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "webhook accepted",
	})
}
`

const authTemplate = `package http

import (
	"net/http"
	"os"

	"{{ .ModulePath }}/internal/observability"
)

var authLogger = observability.New("auth")

func loginExampleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		authLogger.Error("登录请求方法不允许", "method", r.Method)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": os.Getenv("AUTH_TOKEN"),
	})
}

func profileExampleHandler(w http.ResponseWriter, r *http.Request) {
	expected := "Bearer " + os.Getenv("AUTH_TOKEN")
	if r.Header.Get("Authorization") != expected {
		authLogger.Error("访问受保护接口失败", "reason", "invalid_token")
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	authLogger.Info("访问受保护接口成功")
	writeJSON(w, http.StatusOK, map[string]string{
		"user": "demo-user",
	})
}
`

const postgresCheckMainTemplate = `package main

import (
	"context"
	"os"
	"time"

	"{{ .ModulePath }}/internal/observability"
	"{{ .ModulePath }}/internal/store"
)

func main() {
	observability.SetLevel(os.Getenv("LOG_LEVEL"))
	logger := observability.New("dbcheck")

	timeout := 3 * time.Second
	if raw := os.Getenv("DATABASE_TIMEOUT"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := store.CheckPostgres(ctx, os.Getenv("DATABASE_URL")); err != nil {
		logger.Error("数据库检查失败", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("数据库检查通过")
}
`

const postgresHTTPTemplate = `package http

import (
	"context"
	"net/http"
	"os"
	"time"

	"{{ .ModulePath }}/internal/observability"
	"{{ .ModulePath }}/internal/store"
)

var postgresLogger = observability.New("postgres")

func postgresReadyHandler(w http.ResponseWriter, r *http.Request) {
	timeout := 3 * time.Second
	if raw := os.Getenv("DATABASE_TIMEOUT"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	if err := store.CheckPostgres(ctx, os.Getenv("DATABASE_URL")); err != nil {
		postgresLogger.Error("数据库检查失败", "error", err.Error())
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "error": err.Error()})
		return
	}

	postgresLogger.Info("数据库检查通过")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
`

const postgresStoreTemplate = `package store

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

func CheckPostgres(ctx context.Context, databaseURL string) error {
	if strings.TrimSpace(databaseURL) == "" {
		return fmt.Errorf("DATABASE_URL 不能为空")
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("解析 DATABASE_URL 失败：%w", err)
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("DATABASE_URL 缺少主机名")
	}

	port := parsed.Port()
	if port == "" {
		port = "5432"
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return fmt.Errorf("连接 PostgreSQL 地址失败：%w", err)
	}
	defer conn.Close()

	return nil
}
`

const postgresLifecycleTemplate = `package store

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"{{ .ModulePath }}/internal/observability"
)

var postgresLifecycleLogger = observability.New("postgres-lifecycle")

type PostgresManager struct {
	databaseURL string
}

func NewPostgresManager(databaseURL string) *PostgresManager {
	return &PostgresManager{databaseURL: databaseURL}
}

func (m *PostgresManager) Start(context.Context) error {
	if strings.TrimSpace(m.databaseURL) == "" {
		return fmt.Errorf("DATABASE_URL 不能为空")
	}

	parsed, err := url.Parse(m.databaseURL)
	if err != nil {
		return fmt.Errorf("解析 DATABASE_URL 失败：%w", err)
	}

	postgresLifecycleLogger.Info("数据库生命周期已接入", "host", parsed.Hostname())
	return nil
}

func (m *PostgresManager) Stop(context.Context) error {
	postgresLifecycleLogger.Info("数据库生命周期已停止")
	return nil
}
`

const addonConfigTemplate = `package config

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
	port := getenv("PORT", "8080")
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

const addonAppTemplate = `package app

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

const requestIDTemplate = `package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"{{ .ModulePath }}/internal/observability"
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

	"{{ .ModulePath }}/internal/observability"
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

const metricsTemplate = `package http

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"{{ .ModulePath }}/internal/observability"
)

var metricsLogger = observability.New("metrics")

var totalRequests uint64
var totalRecoveries uint64

func init() {
	recordRecovery = metricsRecoveriesInc
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&totalRequests, 1)
		next.ServeHTTP(w, r)
	})
}

func metricsRecoveriesInc() {
	atomic.AddUint64(&totalRecoveries, 1)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "golider_http_requests_total %d\n", atomic.LoadUint64(&totalRequests))
	_, _ = fmt.Fprintf(w, "golider_http_recoveries_total %d\n", atomic.LoadUint64(&totalRecoveries))
	metricsLogger.Info("metrics 已输出")
}
`

const rateLimitTemplate = `package http

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"{{ .ModulePath }}/internal/observability"
)

var rateLimitLogger = observability.New("rate-limit")

type fixedWindowLimiter struct {
	mu         sync.Mutex
	window     int64
	count      int
	limit      int
}

var limiter = &fixedWindowLimiter{}

func rateLimitMiddleware(next http.Handler) http.Handler {
	limit := 20
	if raw := os.Getenv("RATE_LIMIT_PER_SECOND"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	limiter.limit = limit

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.allow() {
			rateLimitLogger.Error("请求触发限流", "path", r.URL.Path, "limit", limit)
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *fixedWindowLimiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().Unix()
	if l.window != now {
		l.window = now
		l.count = 0
	}

	if l.limit <= 0 {
		l.limit = 20
	}
	if l.count >= l.limit {
		return false
	}

	l.count++
	return true
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

const corsTemplate = `package http

import (
	"net/http"
	"os"
	"strings"

	"{{ .ModulePath }}/internal/observability"
)

var corsLogger = observability.New("cors")

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	if strings.TrimSpace(allowedOrigins) == "" {
		allowedOrigins = "*"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowOrigin(origin, allowedOrigins) {
			if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			corsLogger.Info("跨域预检请求已处理", "path", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowOrigin(origin string, allowedOrigins string) bool {
	if allowedOrigins == "*" {
		return true
	}
	if strings.TrimSpace(origin) == "" {
		return false
	}

	for _, item := range strings.Split(allowedOrigins, ",") {
		if strings.TrimSpace(item) == origin {
			return true
		}
	}

	return false
}
`

const websocketTemplate = `package http

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"{{ .ModulePath }}/internal/observability"
)

var websocketLogger = observability.New("websocket")

const (
	websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	pingInterval  = 30 * time.Second
	writeTimeout  = 10 * time.Second
)

type Message struct {
	Room    string ` + "`json:\"room\"`" + `
	Event   string ` + "`json:\"event\"`" + `
	Payload any    ` + "`json:\"payload\"`" + `
}

type Hub struct {
	mu       sync.RWMutex
	rooms    map[string]map[*Client]bool
	register chan *Client
	unregister chan *Client
	broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.room] == nil {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true
			h.mu.Unlock()
			websocketLogger.Info("客户端加入房间", "room", client.room, "remote", client.conn.RemoteAddr())

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.room]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.rooms, client.room)
				}
			}
			h.mu.Unlock()
			client.conn.Close()
			websocketLogger.Info("客户端离开房间", "room", client.room)

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[msg.Room]
			h.mu.RUnlock()

			data, err := json.Marshal(msg)
			if err != nil {
				websocketLogger.Error("消息序列化失败", "error", err)
				continue
			}

			for client := range clients {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(room, event string, payload any) {
	h.broadcast <- Message{Room: room, Event: event, Payload: payload}
}

type Client struct {
	hub  *Hub
	conn net.Conn
	send chan []byte
	room string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		frame, err := c.readFrame()
		if err != nil {
			websocketLogger.Error("读取帧失败", "error", err)
			return
		}

		if frame.opcode == 0x8 {
			return
		}

		if frame.opcode == 0x9 {
			c.writePong(frame.payload)
			continue
		}

		if frame.opcode == 0x1 {
			var msg Message
			if err := json.Unmarshal(frame.payload, &msg); err == nil {
				if msg.Room != "" && msg.Room != c.room {
					c.hub.unregister <- c
					c.room = msg.Room
					c.hub.register <- c
				}
			}
		}

		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.writeClose()
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.writeFrame(0x1, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.writeFrame(0x9, nil); err != nil {
				return
			}
		}
	}
}

type wsFrame struct {
	opcode  byte
	payload []byte
}

func (c *Client) readFrame() (*wsFrame, error) {
	buf := make([]byte, 2)
	if _, err := c.conn.Read(buf); err != nil {
		return nil, err
	}

	opcode := buf[0] & 0x0F
	masked := buf[1]&0x80 != 0
	length := int(buf[1] & 0x7F)

	if length == 126 {
		ext := make([]byte, 2)
		c.conn.Read(ext)
		length = int(ext[0])<<8 | int(ext[1])
	} else if length == 127 {
		ext := make([]byte, 8)
		c.conn.Read(ext)
		length = 0
		for i := 0; i < 8; i++ {
			length = length<<8 | int(ext[i])
		}
	}

	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		c.conn.Read(maskKey)
	}

	payload := make([]byte, length)
	if _, err := c.conn.Read(payload); err != nil {
		return nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return &wsFrame{opcode: opcode, payload: payload}, nil
}

func (c *Client) writeFrame(opcode byte, payload []byte) error {
	buf := []byte{0x80 | opcode}

	length := len(payload)
	if length < 126 {
		buf = append(buf, byte(length))
	} else if length < 65536 {
		buf = append(buf, 126, byte(length>>8), byte(length))
	} else {
		buf = append(buf, 127)
		for i := 7; i >= 0; i-- {
			buf = append(buf, byte(length>>(i*8)))
		}
	}

	buf = append(buf, payload...)
	_, err := c.conn.Write(buf)
	return err
}

func (c *Client) writePong(payload []byte) error {
	return c.writeFrame(0xA, payload)
}

func (c *Client) writeClose() error {
	return c.writeFrame(0x8, nil)
}

var hub = NewHub()

func init() {
	go hub.Run()
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected websocket upgrade"})
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing Sec-WebSocket-Key"})
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "websocket not supported"})
		return
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hijack connection"})
		return
	}

	acceptKey := computeAcceptKey(key)
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"

	if _, err := conn.Write([]byte(response)); err != nil {
		conn.Close()
		return
	}

	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		room: room,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func BroadcastToRoom(room, event string, payload any) {
	hub.Broadcast(room, event, payload)
}
`

const circuitBreakerTemplate = `package http

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"{{ .ModulePath }}/internal/observability"
)

var circuitBreakerLogger = observability.New("circuit-breaker")

type circuitState int

const (
	circuitClosed   circuitState = 0
	circuitOpen     circuitState = 1
	circuitHalfOpen circuitState = 2
)

type circuitBreaker struct {
	mu               sync.Mutex
	state            circuitState
	failureCount     int
	successCount     int
	threshold        int
	timeout          time.Duration
	successThreshold int
	openedAt         time.Time
}

var breaker = &circuitBreaker{
	threshold:        5,
	timeout:          30 * time.Second,
	successThreshold: 2,
}

func init() {
	if raw := os.Getenv("CIRCUIT_BREAKER_THRESHOLD"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			breaker.threshold = n
		}
	}
	if raw := os.Getenv("CIRCUIT_BREAKER_TIMEOUT"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			breaker.timeout = d
		}
	}
	if raw := os.Getenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			breaker.successThreshold = n
		}
	}
}

type cbResponseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *cbResponseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func circuitBreakerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !breaker.allow() {
			circuitBreakerLogger.Error("熔断器已开启，请求被拒绝", "path", r.URL.Path)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "circuit breaker open"})
			return
		}

		recorder := &cbResponseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		breaker.record(recorder.status)
	})
}

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case circuitClosed:
		return true
	case circuitOpen:
		if time.Since(cb.openedAt) >= cb.timeout {
			cb.state = circuitHalfOpen
			cb.successCount = 0
			circuitBreakerLogger.Info("熔断器进入半开状态")
			return true
		}
		return false
	case circuitHalfOpen:
		return true
	}
	return true
}

func (cb *circuitBreaker) record(status int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	isFailure := status >= 500

	switch cb.state {
	case circuitClosed:
		if isFailure {
			cb.failureCount++
			if cb.failureCount >= cb.threshold {
				cb.state = circuitOpen
				cb.openedAt = time.Now()
				circuitBreakerLogger.Error("熔断器从关闭切换到开启", "failures", cb.failureCount)
			}
		} else {
			cb.failureCount = 0
		}
	case circuitHalfOpen:
		if isFailure {
			cb.state = circuitOpen
			cb.openedAt = time.Now()
			cb.successCount = 0
			circuitBreakerLogger.Error("熔断器从半开切换到开启")
		} else {
			cb.successCount++
			if cb.successCount >= cb.successThreshold {
				cb.state = circuitClosed
				cb.failureCount = 0
				circuitBreakerLogger.Info("熔断器从半开恢复到关闭")
			}
		}
	}
}
`

const addonLoggerTemplate = `package observability

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

const redisStoreTemplate = `package store

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"{{ .ModulePath }}/internal/observability"
)

var redisLifecycleLogger = observability.New("redis-lifecycle")

type RedisManager struct {
	redisURL string
}

func NewRedisManager(redisURL string) *RedisManager {
	return &RedisManager{redisURL: redisURL}
}

func (m *RedisManager) Start(context.Context) error {
	if strings.TrimSpace(m.redisURL) == "" {
		return fmt.Errorf("REDIS_URL 不能为空")
	}

	parsed, err := url.Parse(m.redisURL)
	if err != nil {
		return fmt.Errorf("解析 REDIS_URL 失败：%w", err)
	}

	redisLifecycleLogger.Info("Redis 生命周期已接入", "host", parsed.Hostname())
	return nil
}

func (m *RedisManager) Stop(context.Context) error {
	redisLifecycleLogger.Info("Redis 生命周期已停止")
	return nil
}

func CheckRedis(ctx context.Context, redisURL string) error {
	if strings.TrimSpace(redisURL) == "" {
		return fmt.Errorf("REDIS_URL 不能为空")
	}

	parsed, err := url.Parse(redisURL)
	if err != nil {
		return fmt.Errorf("解析 REDIS_URL 失败：%w", err)
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("REDIS_URL 缺少主机名")
	}

	port := parsed.Port()
	if port == "" {
		port = "6379"
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return fmt.Errorf("连接 Redis 地址失败：%w", err)
	}
	defer conn.Close()

	return nil
}
`

const redisHTTPTemplate = `package http

import (
	"context"
	"net/http"
	"os"
	"time"

	"{{ .ModulePath }}/internal/observability"
	"{{ .ModulePath }}/internal/store"
)

var redisLogger = observability.New("redis")

func redisReadyHandler(w http.ResponseWriter, r *http.Request) {
	timeout := 3 * time.Second
	if raw := os.Getenv("DATABASE_TIMEOUT"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	if err := store.CheckRedis(ctx, os.Getenv("REDIS_URL")); err != nil {
		redisLogger.Error("Redis 检查失败", "error", err.Error())
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "error": err.Error()})
		return
	}

	redisLogger.Info("Redis 检查通过")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
`

const grpcProtoTemplate = `syntax = "proto3";

package {{ .ProjectName }};

option go_package = "{{ .ModulePath }}/proto";

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {}
  rpc HealthCheck (HealthCheckRequest) returns (HealthCheckReply) {}
}

message HelloRequest {
  string name = 1;
}

message HelloReply {
  string message = 1;
}

message HealthCheckRequest {}

message HealthCheckReply {
  string status = 1;
}
`

const grpcMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"{{ .ModulePath }}/internal/app"
	"{{ .ModulePath }}/internal/config"
	"{{ .ModulePath }}/internal/grpc"
	"{{ .ModulePath }}/internal/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败：%v\n", err)
		os.Exit(1)
	}

	observability.SetLevel(cfg.LogLevel)
	logger := observability.New("grpc-server")

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	server := grpc.NewServer("{{ .ProjectName }}", grpcPort)
	lifecycle := app.New()
	lifecycle.OnStart("grpc", server.StartHook())
	lifecycle.OnStop("grpc", server.StopHook())

	if err := lifecycle.Start(context.Background()); err != nil {
		logger.Error("gRPC 服务启动失败", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("gRPC 服务已启动", "port", grpcPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := lifecycle.Stop(ctx); err != nil {
		logger.Error("gRPC 服务停止失败", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("gRPC 服务已停止")
}
`

const grpcServerTemplate = `package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"{{ .ModulePath }}/internal/observability"
)

var grpcLogger = observability.New("grpc")

type Server struct {
	name string
	port string
	srv  *grpc.Server
}

func NewServer(name, port string) *Server {
	srv := grpc.NewServer()

	// 注册示例服务
	RegisterGreeterServer(srv, &greeterService{})

	// 注册 gRPC 反射服务，方便调试
	reflection.Register(srv)

	return &Server{
		name: name,
		port: port,
		srv:  srv,
	}
}

func (s *Server) StartHook() func(context.Context) error {
	return func(context.Context) error {
		lis, err := net.Listen("tcp", ":"+s.port)
		if err != nil {
			return fmt.Errorf("gRPC 监听端口 %s 失败：%w", s.port, err)
		}

		grpcLogger.Info("gRPC 服务开始监听", "port", s.port, "name", s.name)
		go func() {
			if err := s.srv.Serve(lis); err != nil {
				grpcLogger.Error("gRPC 服务异常退出", "error", err.Error())
			}
		}()

		return nil
	}
}

func (s *Server) StopHook() func(context.Context) error {
	return func(context.Context) error {
		grpcLogger.Info("gRPC 服务正在停止", "name", s.name)
		s.srv.GracefulStop()
		return nil
	}
}
`

const grpcServiceTemplate = `package grpc

import (
	"context"

	"{{ .ModulePath }}/internal/observability"
)

var greeterLogger = observability.New("greeter")

type greeterService struct {
	UnimplementedGreeterServer
}

func (s *greeterService) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	greeterLogger.Info("收到 SayHello 请求", "name", req.GetName())
	return &HelloReply{
		Message: "Hello, " + req.GetName() + "!",
	}, nil
}

func (s *greeterService) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckReply, error) {
	greeterLogger.Info("收到 HealthCheck 请求")
	return &HealthCheckReply{
		Status: "SERVING",
	}, nil
}
`

const kafkaMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"{{ .ModulePath }}/internal/app"
	"{{ .ModulePath }}/internal/config"
	"{{ .ModulePath }}/internal/kafka"
	"{{ .ModulePath }}/internal/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败：%v\n", err)
		os.Exit(1)
	}

	observability.SetLevel(cfg.LogLevel)
	logger := observability.New("kafka")

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "{{ .ProjectName }}-events"
	}

	consumer := kafka.NewConsumer(brokers, topic)
	producer := kafka.NewProducer(brokers, topic)

	lifecycle := app.New()
	lifecycle.OnStart("kafka-consumer", consumer.StartHook())
	lifecycle.OnStart("kafka-producer", producer.StartHook())
	lifecycle.OnStop("kafka-producer", producer.StopHook())
	lifecycle.OnStop("kafka-consumer", consumer.StopHook())

	if err := lifecycle.Start(context.Background()); err != nil {
		logger.Error("Kafka 组件启动失败", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Kafka 组件已启动", "brokers", brokers, "topic", topic)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := lifecycle.Stop(ctx); err != nil {
		logger.Error("Kafka 组件停止失败", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Kafka 组件已停止")
}
`

const kafkaConsumerTemplate = `package kafka

import (
	"context"
	"time"

	"{{ .ModulePath }}/internal/observability"
)

var consumerLogger = observability.New("kafka-consumer")

type Consumer struct {
	brokers string
	topic   string
	stop    chan struct{}
}

func NewConsumer(brokers, topic string) *Consumer {
	return &Consumer{
		brokers: brokers,
		topic:   topic,
		stop:    make(chan struct{}),
	}
}

func (c *Consumer) StartHook() func(context.Context) error {
	return func(context.Context) error {
		go c.consume()
		return nil
	}
}

func (c *Consumer) StopHook() func(context.Context) error {
	return func(context.Context) error {
		close(c.stop)
		return nil
	}
}

func (c *Consumer) consume() {
	consumerLogger.Info("消费者已启动", "brokers", c.brokers, "topic", c.topic)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			consumerLogger.Info("消费者轮询消息", "topic", c.topic)
			// TODO: 在此处实现实际的 Kafka 消费逻辑
			// 例如：使用 sarama 库连接 Kafka 集群并消费消息
		case <-c.stop:
			consumerLogger.Info("消费者已停止", "topic", c.topic)
			return
		}
	}
}
`

const kafkaProducerTemplate = `package kafka

import (
	"context"
	"fmt"
	"time"

	"{{ .ModulePath }}/internal/observability"
)

var producerLogger = observability.New("kafka-producer")

type Producer struct {
	brokers string
	topic   string
	stop    chan struct{}
}

func NewProducer(brokers, topic string) *Producer {
	return &Producer{
		brokers: brokers,
		topic:   topic,
		stop:    make(chan struct{}),
	}
}

func (p *Producer) StartHook() func(context.Context) error {
	return func(context.Context) error {
		go p.heartbeat()
		return nil
	}
}

func (p *Producer) StopHook() func(context.Context) error {
	return func(context.Context) error {
		close(p.stop)
		return nil
	}
}

func (p *Producer) heartbeat() {
	producerLogger.Info("生产者已启动", "brokers", p.brokers, "topic", p.topic)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			producerLogger.Info("生产者心跳检测通过", "topic", p.topic)
		case <-p.stop:
			producerLogger.Info("生产者已停止", "topic", p.topic)
			return
		}
	}
}

func (p *Producer) Send(ctx context.Context, key []byte, value []byte) error {
	producerLogger.Info("生产者发送消息", "topic", p.topic, "key", string(key))
	// TODO: 在此处实现实际的 Kafka 生产逻辑
	// 例如：使用 sarama 库连接 Kafka 集群并发送消息
	return fmt.Errorf("未连线 Kafka，消息未能发送到 %s", p.topic)
}
`

const kafkaLifecycleTemplate = `package kafka

// 生命周期钩子定义在 consumer.go 和 producer.go 中。
// 此文件用于导出函数签名与通用常量。
`

const schedulerTemplate = `package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"{{ .ModulePath }}/internal/observability"
)

var slog = observability.New("scheduler")

// Job 表示一个已注册的定时任务
type Job struct {
	Name     string
	Schedule string
	Handler  func(ctx context.Context) error

	interval time.Duration
	stop     chan struct{}
	done     chan struct{}
}

// Scheduler 定时任务调度器，纯标准库实现
type Scheduler struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// New 创建调度器实例
func New() *Scheduler {
	return &Scheduler{
		jobs: make(map[string]*Job),
	}
}

// Register 注册一个定时任务。schedule 支持 @every 格式，例如 @every 5s、@every 1m、@every 2h。
func (s *Scheduler) Register(name, schedule string, handler func(ctx context.Context) error) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("任务名称不能为空")
	}
	if handler == nil {
		return fmt.Errorf("任务 %q 的处理函数不能为空", name)
	}

	interval, err := parseSchedule(schedule)
	if err != nil {
		return fmt.Errorf("解析调度表达式失败：%w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("定时任务 %q 已存在", name)
	}

	s.jobs[name] = &Job{
		Name:     name,
		Schedule: schedule,
		Handler:  handler,
		interval: interval,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	slog.Info("注册定时任务", "name", name, "schedule", schedule, "interval", interval.String())
	return nil
}

// List 返回所有已注册任务的信息（不含处理函数）
func (s *Scheduler) List() []JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]JobInfo, 0, len(s.jobs))
	for _, j := range s.jobs {
		infos = append(infos, JobInfo{
			Name:     j.Name,
			Schedule: j.Schedule,
			Interval: j.interval.String(),
		})
	}
	return infos
}

// JobInfo 任务信息（对外暴露的只读视图）
type JobInfo struct {
	Name     string ` + "`" + `json:"name"` + "`" + `
	Schedule string ` + "`" + `json:"schedule"` + "`" + `
	Interval string ` + "`" + `json:"interval"` + "`" + `
}

// Start 启动所有已注册的定时任务
func (s *Scheduler) Start() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		go s.runJob(job)
	}
	slog.Info("调度器已启动", "job_count", len(s.jobs))
}

// Stop 停止所有定时任务并等待执行中的任务完成
func (s *Scheduler) Stop() {
	s.mu.RLock()
	for _, job := range s.jobs {
		close(job.stop)
	}
	s.mu.RUnlock()

	// 等待所有 job goroutine 退出
	s.mu.RLock()
	for _, job := range s.jobs {
		<-job.done
	}
	s.mu.RUnlock()

	slog.Info("调度器已停止")
}

// Trigger 手动触发指定任务一次（同步执行，不阻塞定时调度）
func (s *Scheduler) Trigger(ctx context.Context, name string) error {
	s.mu.RLock()
	job, exists := s.jobs[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("定时任务 %q 不存在", name)
	}

	slog.Info("手动触发定时任务", "name", name)
	return job.Handler(ctx)
}

func (s *Scheduler) runJob(job *Job) {
	defer close(job.done)

	ticker := time.NewTicker(job.interval)
	defer ticker.Stop()

	slog.Info("定时任务开始运行", "name", job.Name)

	for {
		select {
		case <-ticker.C:
			slog.Info("执行定时任务", "name", job.Name)
			ctx, cancel := context.WithTimeout(context.Background(), job.interval)
			if err := job.Handler(ctx); err != nil {
				slog.Error("定时任务执行失败", "name", job.Name, "error", err.Error())
			}
			cancel()
		case <-job.stop:
			slog.Info("定时任务已停止", "name", job.Name)
			return
		}
	}
}

// parseSchedule 解析调度表达式，支持 @every <duration> 格式
func parseSchedule(schedule string) (time.Duration, error) {
	schedule = strings.TrimSpace(schedule)

	if strings.HasPrefix(schedule, "@every ") {
		d, err := time.ParseDuration(strings.TrimPrefix(schedule, "@every "))
		if err != nil {
			return 0, fmt.Errorf("无效的 @every 时长 %q：%w", schedule, err)
		}
		if d <= 0 {
			return 0, fmt.Errorf("@every 时长必须大于 0")
		}
		return d, nil
	}

	return 0, fmt.Errorf("不支持的调度表达式 %q，请使用 @every <duration> 格式，例如 @every 5s、@every 1m、@every 2h", schedule)
}
`

const schedulerLifecycleTemplate = `package scheduler

import "context"

// StartHook 返回调度器的启动钩子
func (s *Scheduler) StartHook() func(context.Context) error {
	return func(context.Context) error {
		s.Start()
		return nil
	}
}

// StopHook 返回调度器的停止钩子
func (s *Scheduler) StopHook() func(context.Context) error {
	return func(context.Context) error {
		s.Stop()
		return nil
	}
}
`

const schedulerHTTPTemplate = `package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"{{ .ModulePath }}/internal/observability"
	"{{ .ModulePath }}/internal/scheduler"
)

var schedulerHTTPLogger = observability.New("scheduler-http")

// SchedulerInstance 由 main.go 注入的调度器实例
var SchedulerInstance *scheduler.Scheduler

func schedulerListHandler(w http.ResponseWriter, r *http.Request) {
	if SchedulerInstance == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "调度器未初始化"})
		return
	}

	tasks := SchedulerInstance.List()
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func schedulerTriggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	if SchedulerInstance == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "调度器未初始化"})
		return
	}

	// 从路径中提取任务名称：/scheduler/trigger/<name>
	name := strings.TrimPrefix(r.URL.Path, "/scheduler/trigger/")
	if name == "" {
		var body struct {
			Name string ` + "`" + `json:"name"` + "`" + `
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供任务名称（路径参数或 JSON body.name）"})
			return
		}
		name = body.Name
	}

	if err := SchedulerInstance.Trigger(r.Context(), name); err != nil {
		schedulerHTTPLogger.Error("手动触发任务失败", "name", name, "error", err.Error())
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	schedulerHTTPLogger.Info("手动触发任务成功", "name", name)
	writeJSON(w, http.StatusOK, map[string]string{"status": "triggered", "name": name})
}
`
