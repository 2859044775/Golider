package scaffold

func files() map[string]string {
	return map[string]string{
		".env.example":                        envExampleTemplate,
		".gitignore":                          gitignoreTemplate,
		".github/workflows/ci.yml":            ciTemplate,
		"Dockerfile":                          dockerfileTemplate,
		"Makefile":                            makefileTemplate,
		"README.md":                           projectReadmeTemplate,
		"go.mod":                              goModTemplate,
		"cmd/api/main.go":                     apiMainTemplate,
		"internal/app/app.go":                 appTemplate,
		"internal/app/dependencies.go":        dependenciesTemplate,
		"internal/config/config.go":           configTemplate,
		"internal/http/binding.go":            bindingTemplate,
		"internal/http/errors.go":             errorModelTemplate,
		"internal/http/middleware.go":         middlewareTemplate,
		"internal/http/binding_test.go":       bindingTestTemplate,
		"internal/http/middleware_test.go":    middlewareTestTemplate,
		"internal/http/query.go":              queryTemplate,
		"internal/http/query_test.go":         queryTestTemplate,
		"internal/http/readiness.go":          readinessTemplate,
		"internal/http/router.go":             routerTemplate,
		"internal/http/router_test.go":        routerTestTemplate,
		"internal/http/requestid.go":          requestIDTemplate,
		"internal/http/timeout.go":            timeoutTemplate,
		"internal/observability/logger.go":    loggerTemplate,
		"internal/repository/message.go":      messageRepositoryTemplate,
		"internal/repository/message_test.go": messageRepositoryTestTemplate,
		"internal/service/message.go":         messageServiceTemplate,
		"internal/service/message_test.go":    messageServiceTestTemplate,
	}
}

const goModTemplate = `module {{ .Module }}

go 1.20
`

const envExampleTemplate = `PORT={{ .DefaultPort }}
SHUTDOWN_TIMEOUT=10s
LOG_LEVEL=info
REQUEST_TIMEOUT=5s
HTTP_READ_HEADER_TIMEOUT=2s
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
DEFAULT_PAGE_SIZE=10
MAX_PAGE_SIZE=100
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

由 Golider 生成的 Go API 工程。

## 快速启动

` + "```bash" + `
cp .env.example .env
make run
` + "```" + `

## 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| ` + "`GET`" + ` | ` + "`/`" + ` | 欢迎 |
| ` + "`GET`" + ` | ` + "`/healthz`" + ` | 健康检查 |
| ` + "`GET`" + ` | ` + "`/readyz`" + ` | 就绪检查 |
| ` + "`GET`" + ` | ` + "`/messages`" + ` | 消息列表（分页、搜索、排序、过滤） |
| ` + "`POST`" + ` | ` + "`/messages`" + ` | 创建消息（Idempotency-Key 幂等） |
| ` + "`GET`" + ` | ` + "`/messages/{id}`" + ` | 消息详情 |
| ` + "`PATCH`" + ` | ` + "`/messages/{id}`" + ` | 局部更新（状态流转） |
| ` + "`DELETE`" + ` | ` + "`/messages/{id}`" + ` | 软删除 |
| ` + "`POST`" + ` | ` + "`/echo`" + ` | 请求回显 |

` + "```bash" + `
# 创建消息
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{"title":"hello","content":"world"}'

# 查询列表
curl http://localhost:8080/messages?page=1&page_size=10
` + "```" + `

## 工程能力

日志 · 请求标识 · 请求超时 · Panic Recovery · 统一错误模型 · JSON 输入校验 · 查询解析 · 分页 · 幂等写入 · 冲突校验 · 状态流转 · 软删除 · 审计字段 · 仓储抽象 · 配置校验 · 生命周期 · 就绪摘流 · HTTP 超时护栏 · Dockerfile · CI

## 验证

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
	Port              string
	ShutdownTimeout   time.Duration
	LogLevel          string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	DefaultPageSize   int
	MaxPageSize       int
}

func Load() (Config, error) {
	port := getenv("PORT", "{{ .DefaultPort }}")
	timeout := getenv("SHUTDOWN_TIMEOUT", "10s")
	logLevel := strings.ToLower(getenv("LOG_LEVEL", "info"))
	readHeaderTimeout := getenv("HTTP_READ_HEADER_TIMEOUT", "2s")
	readTimeout := getenv("HTTP_READ_TIMEOUT", "10s")
	writeTimeout := getenv("HTTP_WRITE_TIMEOUT", "15s")
	idleTimeout := getenv("HTTP_IDLE_TIMEOUT", "60s")
	defaultPageSize := getenv("DEFAULT_PAGE_SIZE", "10")
	maxPageSize := getenv("MAX_PAGE_SIZE", "100")

	d, err := time.ParseDuration(timeout)
	if err != nil {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT 解析失败：%w", err)
	}
	readHeaderDuration, err := time.ParseDuration(readHeaderTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_READ_HEADER_TIMEOUT 解析失败：%w", err)
	}
	readDuration, err := time.ParseDuration(readTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_READ_TIMEOUT 解析失败：%w", err)
	}
	writeDuration, err := time.ParseDuration(writeTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_WRITE_TIMEOUT 解析失败：%w", err)
	}
	idleDuration, err := time.ParseDuration(idleTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_IDLE_TIMEOUT 解析失败：%w", err)
	}
	defaultPageValue, err := strconv.Atoi(defaultPageSize)
	if err != nil {
		return Config{}, fmt.Errorf("DEFAULT_PAGE_SIZE 解析失败：%w", err)
	}
	maxPageValue, err := strconv.Atoi(maxPageSize)
	if err != nil {
		return Config{}, fmt.Errorf("MAX_PAGE_SIZE 解析失败：%w", err)
	}

	cfg := Config{
		Port:              port,
		ShutdownTimeout:   d,
		LogLevel:          logLevel,
		ReadHeaderTimeout: readHeaderDuration,
		ReadTimeout:       readDuration,
		WriteTimeout:      writeDuration,
		IdleTimeout:       idleDuration,
		DefaultPageSize:   defaultPageValue,
		MaxPageSize:       maxPageValue,
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
	if cfg.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("HTTP_READ_HEADER_TIMEOUT 必须大于 0")
	}
	if cfg.ReadTimeout <= 0 {
		return fmt.Errorf("HTTP_READ_TIMEOUT 必须大于 0")
	}
	if cfg.WriteTimeout <= 0 {
		return fmt.Errorf("HTTP_WRITE_TIMEOUT 必须大于 0")
	}
	if cfg.IdleTimeout <= 0 {
		return fmt.Errorf("HTTP_IDLE_TIMEOUT 必须大于 0")
	}
	if cfg.DefaultPageSize <= 0 {
		return fmt.Errorf("DEFAULT_PAGE_SIZE 必须大于 0")
	}
	if cfg.MaxPageSize <= 0 {
		return fmt.Errorf("MAX_PAGE_SIZE 必须大于 0")
	}
	if cfg.MaxPageSize < cfg.DefaultPageSize {
		return fmt.Errorf("MAX_PAGE_SIZE 不能小于 DEFAULT_PAGE_SIZE")
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

const dependenciesTemplate = `package app

import (
	"{{ .Module }}/internal/config"
	"{{ .Module }}/internal/repository"
	"{{ .Module }}/internal/service"
)

type Dependencies struct {
	MessageService   *service.MessageService
	DefaultPageSize int
	MaxPageSize     int
}

func NewDependencies(cfg config.Config) Dependencies {
	repo := repository.NewInMemoryMessageRepository()
	return Dependencies{
		MessageService:   service.NewMessageService(repo),
		DefaultPageSize: cfg.DefaultPageSize,
		MaxPageSize:     cfg.MaxPageSize,
	}
}

func (d Dependencies) WithDefaults() Dependencies {
	if d.MessageService == nil {
		d.MessageService = service.NewMessageService(repository.NewInMemoryMessageRepository())
	}
	if d.DefaultPageSize <= 0 {
		d.DefaultPageSize = 10
	}
	if d.MaxPageSize <= 0 {
		d.MaxPageSize = 100
	}
	if d.MaxPageSize < d.DefaultPageSize {
		d.MaxPageSize = d.DefaultPageSize
	}
	return d
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

const bindingTemplate = `package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type requestValidationError struct {
	Code    string
	Message string
}

func (e *requestValidationError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func decodeJSON(r *http.Request, dst any) error {
	if r == nil {
		return &requestValidationError{Code: "invalid_request", Message: "request is required"}
	}
	if r.Body == nil {
		return &requestValidationError{Code: "invalid_request", Message: "request body is required"}
	}
	if r.Header.Get("Content-Type") != "" && !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		return &requestValidationError{Code: "invalid_content_type", Message: "content type must be application/json"}
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		switch {
		case errors.As(err, &syntaxError):
			return &requestValidationError{Code: "invalid_json", Message: fmt.Sprintf("invalid json at position %d", syntaxError.Offset)}
		case errors.Is(err, io.EOF):
			return &requestValidationError{Code: "invalid_request", Message: "request body is required"}
		default:
			return &requestValidationError{Code: "invalid_request", Message: "request body is invalid"}
		}
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return &requestValidationError{Code: "invalid_request", Message: "request body must contain a single json object"}
	}

	return nil
}

func writeBindingError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *requestValidationError
	if errors.As(err, &validationErr) {
		writeError(w, r, http.StatusBadRequest, validationErr.Code, validationErr.Message)
		return
	}
	writeError(w, r, http.StatusBadRequest, "invalid_request", "request body is invalid")
}
`

const queryTemplate = `package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type listQuery struct {
	Page        int
	PageSize    int
	Search      string
	Status      string
	SortBy      string
	SortOrder   string
	CreatedFrom time.Time
	CreatedTo   time.Time
}

func parseListQuery(r *http.Request, defaultPageSize int, maxPageSize int) (listQuery, error) {
	if r == nil {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "request is required"}
	}
	if defaultPageSize <= 0 {
		defaultPageSize = 10
	}
	if maxPageSize <= 0 {
		maxPageSize = 100
	}
	if maxPageSize < defaultPageSize {
		maxPageSize = defaultPageSize
	}

	query := r.URL.Query()
	page, err := parsePositiveInt(query.Get("page"), 1)
	if err != nil {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "page must be a positive integer"}
	}
	pageSize, err := parsePositiveInt(query.Get("page_size"), defaultPageSize)
	if err != nil {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "page_size must be a positive integer"}
	}
	if pageSize > maxPageSize {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "page_size must be less than or equal to max page size"}
	}
	status := strings.TrimSpace(query.Get("status"))
	if status != "" && status != "active" && status != "archived" {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "status must be active or archived"}
	}
	sortBy := strings.TrimSpace(query.Get("sort_by"))
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortBy != "created_at" && sortBy != "title" {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "sort_by must be created_at or title"}
	}
	sortOrder := strings.TrimSpace(query.Get("sort_order"))
	if sortOrder == "" {
		sortOrder = "desc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "sort_order must be asc or desc"}
	}
	createdFrom, err := parseTime(query.Get("created_from"))
	if err != nil {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "created_from must be RFC3339 format"}
	}
	createdTo, err := parseTime(query.Get("created_to"))
	if err != nil {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "created_to must be RFC3339 format"}
	}
	if !createdFrom.IsZero() && !createdTo.IsZero() && createdFrom.After(createdTo) {
		return listQuery{}, &requestValidationError{Code: "invalid_query", Message: "created_from must be earlier than or equal to created_to"}
	}

	return listQuery{
		Page:        page,
		PageSize:    pageSize,
		Search:      strings.TrimSpace(query.Get("q")),
		Status:      status,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
	}, nil
}

func parsePositiveInt(raw string, fallback int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func writePaginatedJSON(w http.ResponseWriter, statusCode int, items any, page int, pageSize int, total int) {
	writeJSON(w, statusCode, map[string]any{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

func parseTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, raw)
}
`

const routerTemplate = `package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"{{ .Module }}/internal/app"
	"{{ .Module }}/internal/service"
)

func NewRouter(deps app.Dependencies) http.Handler {
	deps = deps.WithDefaults()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"service": "{{ .AppName }}", "status": "running"})
	})
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readyHandler)
	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listMessagesHandler(w, r, deps)
		case http.MethodPost:
			createMessageHandler(w, r, deps)
		default:
			writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	})
	mux.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		messageID, ok := messageIDFromPath(r.URL.Path)
		if !ok {
			writeError(w, r, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		switch r.Method {
		case http.MethodGet:
			getMessageHandler(w, r, deps, messageID)
		case http.MethodPatch:
			updateMessageHandler(w, r, deps, messageID)
		case http.MethodDelete:
			deleteMessageHandler(w, r, deps, messageID)
		default:
			writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	})
	mux.HandleFunc("/echo", echoHandler)
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

type createMessageRequest struct {
	Title   string ` + "`json:\"title\"`" + `
	Content string ` + "`json:\"content\"`" + `
	Status  string ` + "`json:\"status\"`" + `
}

type updateMessageRequest struct {
	Title   *string ` + "`json:\"title\"`" + `
	Content *string ` + "`json:\"content\"`" + `
	Status  *string ` + "`json:\"status\"`" + `
}

type echoRequest struct {
	Name    string ` + "`json:\"name\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

func listMessagesHandler(w http.ResponseWriter, r *http.Request, deps app.Dependencies) {
	query, err := parseListQuery(r, deps.DefaultPageSize, deps.MaxPageSize)
	if err != nil {
		writeBindingError(w, r, err)
		return
	}

	result := deps.MessageService.List(context.Background(), service.ListMessagesInput{
		Page:        query.Page,
		PageSize:    query.PageSize,
		Search:      query.Search,
		Status:      query.Status,
		SortBy:      query.SortBy,
		SortOrder:   query.SortOrder,
		CreatedFrom: query.CreatedFrom,
		CreatedTo:   query.CreatedTo,
	})
	writePaginatedJSON(w, http.StatusOK, result.Items, result.Page, result.PageSize, result.Total)
}

func getMessageHandler(w http.ResponseWriter, r *http.Request, deps app.Dependencies, messageID string) {
	message, err := deps.MessageService.GetByID(context.Background(), messageID)
	if err != nil {
		writeMessageServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message":    message,
		"request_id": requestIDFromRequest(r),
	})
}

func createMessageHandler(w http.ResponseWriter, r *http.Request, deps app.Dependencies) {
	var input createMessageRequest
	if err := decodeJSON(r, &input); err != nil {
		writeBindingError(w, r, err)
		return
	}
	if err := validateCreateMessageRequest(input, r.Header.Get("Idempotency-Key")); err != nil {
		writeBindingError(w, r, err)
		return
	}

	result, err := deps.MessageService.Create(context.Background(), service.CreateMessageInput{
		Title:          input.Title,
		Content:        input.Content,
		Status:         input.Status,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		writeMessageServiceError(w, r, err)
		return
	}

	statusCode := http.StatusCreated
	if result.IdempotencyReplay {
		statusCode = http.StatusOK
	}
	writeJSON(w, statusCode, map[string]any{
		"message":            result.Message,
		"idempotency_replay": result.IdempotencyReplay,
		"request_id":         requestIDFromRequest(r),
	})
}

func updateMessageHandler(w http.ResponseWriter, r *http.Request, deps app.Dependencies, messageID string) {
	var input updateMessageRequest
	if err := decodeJSON(r, &input); err != nil {
		writeBindingError(w, r, err)
		return
	}
	if err := validateUpdateMessageRequest(input); err != nil {
		writeBindingError(w, r, err)
		return
	}

	result, err := deps.MessageService.Update(context.Background(), service.UpdateMessageInput{
		ID:      messageID,
		Title:   input.Title,
		Content: input.Content,
		Status:  input.Status,
	})
	if err != nil {
		writeMessageServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":    result.Message,
		"request_id": requestIDFromRequest(r),
	})
}

func deleteMessageHandler(w http.ResponseWriter, r *http.Request, deps app.Dependencies, messageID string) {
	result, err := deps.MessageService.Delete(context.Background(), messageID)
	if err != nil {
		writeMessageServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":    result.Message,
		"deleted":    true,
		"request_id": requestIDFromRequest(r),
	})
}

func validateCreateMessageRequest(input createMessageRequest, idempotencyKey string) error {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	status := strings.TrimSpace(input.Status)
	idempotencyKey = strings.TrimSpace(idempotencyKey)

	if title == "" {
		return &requestValidationError{Code: "invalid_request", Message: "title is required"}
	}
	if len([]rune(title)) > 120 {
		return &requestValidationError{Code: "invalid_request", Message: "title must be within 120 characters"}
	}
	if content == "" {
		return &requestValidationError{Code: "invalid_request", Message: "content is required"}
	}
	if len([]rune(content)) > 2000 {
		return &requestValidationError{Code: "invalid_request", Message: "content must be within 2000 characters"}
	}
	if status != "" && status != "active" && status != "archived" {
		return &requestValidationError{Code: "invalid_request", Message: "status must be active or archived"}
	}
	if len([]rune(idempotencyKey)) > 128 {
		return &requestValidationError{Code: "invalid_request", Message: "idempotency key must be within 128 characters"}
	}
	return nil
}

func validateUpdateMessageRequest(input updateMessageRequest) error {
	if input.Title == nil && input.Content == nil && input.Status == nil {
		return &requestValidationError{Code: "invalid_request", Message: "at least one field must be provided"}
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return &requestValidationError{Code: "invalid_request", Message: "title cannot be empty"}
		}
		if len([]rune(title)) > 120 {
			return &requestValidationError{Code: "invalid_request", Message: "title must be within 120 characters"}
		}
	}
	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return &requestValidationError{Code: "invalid_request", Message: "content cannot be empty"}
		}
		if len([]rune(content)) > 2000 {
			return &requestValidationError{Code: "invalid_request", Message: "content must be within 2000 characters"}
		}
	}
	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)
		if status != "active" && status != "archived" {
			return &requestValidationError{Code: "invalid_request", Message: "status must be active or archived"}
		}
	}
	return nil
}

func writeMessageServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var messageConflict *service.MessageConflictError
	if errors.As(err, &messageConflict) {
		writeError(w, r, http.StatusConflict, "message_title_conflict", "message title already exists")
		return
	}
	var idempotencyConflict *service.IdempotencyConflictError
	if errors.As(err, &idempotencyConflict) {
		writeError(w, r, http.StatusConflict, "idempotency_key_conflict", "idempotency key has been used with a different payload")
		return
	}
	var notFound *service.MessageNotFoundError
	if errors.As(err, &notFound) {
		writeError(w, r, http.StatusNotFound, "message_not_found", "message not found")
		return
	}
	var transitionConflict *service.MessageStatusTransitionError
	if errors.As(err, &transitionConflict) {
		writeError(w, r, http.StatusConflict, "message_status_transition_invalid", "message status transition is not allowed")
		return
	}
	writeError(w, r, http.StatusInternalServerError, "internal_server_error", "internal server error")
}

func messageIDFromPath(path string) (string, bool) {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/messages/") {
		return "", false
	}
	messageID := strings.Trim(strings.TrimPrefix(path, "/messages/"), "/")
	if messageID == "" || strings.Contains(messageID, "/") {
		return "", false
	}
	return messageID, true
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	var input echoRequest
	if err := decodeJSON(r, &input); err != nil {
		writeBindingError(w, r, err)
		return
	}

	input.Message = strings.TrimSpace(input.Message)
	input.Name = strings.TrimSpace(input.Name)
	if input.Message == "" {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "message is required")
		return
	}
	if len([]rune(input.Message)) > 200 {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "message must be within 200 characters")
		return
	}
	if input.Name == "" {
		input.Name = "guest"
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"name":       input.Name,
		"message":    input.Message,
		"request_id": requestIDFromRequest(r),
	})
}
`

const routerTestTemplate = `package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"{{ .Module }}/internal/app"
)

func TestNewRouterDefaultEndpoints(t *testing.T) {
	router := NewRouter(app.Dependencies{})

	for _, item := range []struct {
		name string
		path string
	}{
		{name: "根路径", path: "/"},
		{name: "健康检查", path: "/healthz"},
		{name: "就绪检查", path: "/readyz"},
		{name: "消息列表", path: "/messages"},
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

func TestMessagesHandlerSuccess(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/messages?page=1&page_size=2&q=go&status=active&sort_by=title&sort_order=asc&created_from=2024-01-01T00:00:00Z&created_to=2024-12-31T23:59:59Z", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"items\"", "\"page\":1", "\"page_size\":2", "\"total\":2", "\"title\":\"Go service template\"", "\"status\":\"active\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("messages 响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestMessagesHandlerInvalidQuery(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/messages?sort_order=sideways", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusBadRequest, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"invalid_query\"", "\"message\":\"sort_order must be asc or desc\"", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("messages 错误响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestCreateMessageHandlerSuccess(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(` + "`" + `{"title":"New message","content":"Create endpoint example","status":"active"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "msg-create-1")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusCreated, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"message\"", "\"title\":\"New message\"", "\"content\":\"Create endpoint example\"", "\"idempotency_replay\":false", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("create 响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestCreateMessageHandlerIdempotencyReplay(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	body := ` + "`" + `{"title":"Replay message","content":"same payload","status":"active"}` + "`" + `

	firstReq := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(body))
	firstReq.Header.Set("Content-Type", "application/json")
	firstReq.Header.Set("Idempotency-Key", "msg-replay-1")
	firstRecorder := httptest.NewRecorder()
	router.ServeHTTP(firstRecorder, firstReq)
	if firstRecorder.Code != http.StatusCreated {
		t.Fatalf("首次创建状态码错误，期望 %d，实际 %d", http.StatusCreated, firstRecorder.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(body))
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.Header.Set("Idempotency-Key", "msg-replay-1")
	secondRecorder := httptest.NewRecorder()
	router.ServeHTTP(secondRecorder, secondReq)

	if secondRecorder.Code != http.StatusOK {
		t.Fatalf("幂等回放状态码错误，期望 %d，实际 %d", http.StatusOK, secondRecorder.Code)
	}
	secondBody := secondRecorder.Body.String()
	for _, fragment := range []string{"\"idempotency_replay\":true", "\"title\":\"Replay message\""} {
		if !strings.Contains(secondBody, fragment) {
			t.Fatalf("幂等回放响应缺少片段 %q: %s", fragment, secondBody)
		}
	}
}

func TestCreateMessageHandlerConflict(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(` + "`" + `{"title":"Go service template","content":"duplicate title"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusConflict, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"message_title_conflict\"", "\"message\":\"message title already exists\"", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("冲突响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestCreateMessageHandlerValidation(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(` + "`" + `{"title":"","content":"hello"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusBadRequest, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"invalid_request\"", "\"message\":\"title is required\"", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("创建校验响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestGetMessageHandlerSuccess(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/messages/msg_1", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"message\"", "\"id\":\"msg_1\"", "\"title\":\"Go service template\"", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("详情响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestGetMessageHandlerNotFound(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/messages/msg_404", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusNotFound, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"message_not_found\"", "\"message\":\"message not found\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("详情错误响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestUpdateMessageHandlerSuccess(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPatch, "/messages/msg_1", strings.NewReader(` + "`" + `{"title":"Go service template updated","content":"updated content","status":"archived"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"title\":\"Go service template updated\"", "\"content\":\"updated content\"", "\"status\":\"archived\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("更新响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestUpdateMessageHandlerTransitionConflict(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPatch, "/messages/msg_3", strings.NewReader(` + "`" + `{"status":"active"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusConflict, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"message_status_transition_invalid\"", "\"message\":\"message status transition is not allowed\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("状态流转冲突响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestUpdateMessageHandlerConflict(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPatch, "/messages/msg_2", strings.NewReader(` + "`" + `{"title":"Go service template"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusConflict, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"message_title_conflict\"", "\"message\":\"message title already exists\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("更新冲突响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestUpdateMessageHandlerValidation(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPatch, "/messages/msg_1", strings.NewReader(` + "`" + `{}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusBadRequest, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"invalid_request\"", "\"message\":\"at least one field must be provided\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("更新校验响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestDeleteMessageHandlerSuccess(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	deleteReq := httptest.NewRequest(http.MethodDelete, "/messages/msg_1", nil)
	deleteRecorder := httptest.NewRecorder()

	router.ServeHTTP(deleteRecorder, deleteReq)

	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusOK, deleteRecorder.Code)
	}
	deleteBody := deleteRecorder.Body.String()
	for _, fragment := range []string{"\"deleted\":true", "\"id\":\"msg_1\"", "\"deleted_at\"", "\"request_id\""} {
		if !strings.Contains(deleteBody, fragment) {
			t.Fatalf("删除响应缺少片段 %q: %s", fragment, deleteBody)
		}
	}

	getReq := httptest.NewRequest(http.MethodGet, "/messages/msg_1", nil)
	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, getReq)
	if getRecorder.Code != http.StatusNotFound {
		t.Fatalf("软删除后查询状态码错误，期望 %d，实际 %d", http.StatusNotFound, getRecorder.Code)
	}
}

func TestEchoHandlerSuccess(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(` + "`" + `{"name":"golider","message":"hello"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusOK, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"name\":\"golider\"", "\"message\":\"hello\"", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("echo 响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestEchoHandlerValidation(t *testing.T) {
	router := NewRouter(app.Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(` + "`" + `{"message":""}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("状态码错误，期望 %d，实际 %d", http.StatusBadRequest, recorder.Code)
	}
	body := recorder.Body.String()
	for _, fragment := range []string{"\"code\":\"invalid_request\"", "\"message\":\"message is required\"", "\"request_id\""} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("echo 校验响应缺少片段 %q: %s", fragment, body)
		}
	}
}

func TestReadyHandlerDrainingState(t *testing.T) {
	markReady()
	t.Cleanup(func() {
		markReady()
	})

	router := NewRouter(app.Dependencies{})
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

const bindingTestTemplate = `package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(` + "`" + `{"message":"hello","extra":true}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")

	var body echoRequest
	err := decodeJSON(req, &body)
	if err == nil {
		t.Fatal("decodeJSON 本应拒绝未知字段")
	}

	validationErr, ok := err.(*requestValidationError)
	if !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
	if validationErr.Code != "invalid_request" {
		t.Fatalf("错误码不正确，期望 %q，实际 %q", "invalid_request", validationErr.Code)
	}
}

func TestDecodeJSONRejectsNonJSONContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(` + "`" + `{"message":"hello"}` + "`" + `))
	req.Header.Set("Content-Type", "text/plain")

	var body echoRequest
	err := decodeJSON(req, &body)
	if err == nil {
		t.Fatal("decodeJSON 本应拒绝非 JSON 内容类型")
	}

	validationErr, ok := err.(*requestValidationError)
	if !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
	if validationErr.Code != "invalid_content_type" {
		t.Fatalf("错误码不正确，期望 %q，实际 %q", "invalid_content_type", validationErr.Code)
	}
}

func TestDecodeJSONRejectsMultipleObjects(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(` + "`" + `{"message":"hello"}{"message":"again"}` + "`" + `))
	req.Header.Set("Content-Type", "application/json")

	var body echoRequest
	err := decodeJSON(req, &body)
	if err == nil {
		t.Fatal("decodeJSON 本应拒绝多个 JSON 对象")
	}

	validationErr, ok := err.(*requestValidationError)
	if !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
	if validationErr.Code != "invalid_request" {
		t.Fatalf("错误码不正确，期望 %q，实际 %q", "invalid_request", validationErr.Code)
	}
}
`

const queryTestTemplate = `package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseListQueryDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)

	query, err := parseListQuery(req, 20, 50)
	if err != nil {
		t.Fatalf("parseListQuery 默认值解析失败: %v", err)
	}
	if query.Page != 1 || query.PageSize != 20 {
		t.Fatalf("默认分页参数错误: %+v", query)
	}
	if query.SortBy != "created_at" || query.SortOrder != "desc" {
		t.Fatalf("默认排序参数错误: %+v", query)
	}
}

func TestParseListQueryInvalidPageSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/messages?page_size=101", nil)

	_, err := parseListQuery(req, 10, 100)
	if err == nil {
		t.Fatal("parseListQuery 本应拒绝过大的 page_size")
	}

	validationErr, ok := err.(*requestValidationError)
	if !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
	if validationErr.Code != "invalid_query" {
		t.Fatalf("错误码不正确，期望 %q，实际 %q", "invalid_query", validationErr.Code)
	}
}

func TestParseListQueryInvalidTimeRange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/messages?created_from=2024-02-01T00:00:00Z&created_to=2024-01-01T00:00:00Z", nil)

	_, err := parseListQuery(req, 10, 100)
	if err == nil {
		t.Fatal("parseListQuery 本应拒绝反向时间范围")
	}

	validationErr, ok := err.(*requestValidationError)
	if !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
	if validationErr.Code != "invalid_query" {
		t.Fatalf("错误码不正确，期望 %q，实际 %q", "invalid_query", validationErr.Code)
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
	deps := app.NewDependencies(cfg)
	handler := httptransport.NewRouter(deps)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	lifecycle := app.New()
	errCh := make(chan error, 1)
	lifecycle.OnStart("http-server", func(context.Context) error {
		httptransport.MarkReady()
		logger.Info("服务启动中", "port", cfg.Port, "log_level", cfg.LogLevel, "read_header_timeout", cfg.ReadHeaderTimeout.String(), "read_timeout", cfg.ReadTimeout.String(), "write_timeout", cfg.WriteTimeout.String(), "idle_timeout", cfg.IdleTimeout.String(), "default_page_size", cfg.DefaultPageSize, "max_page_size", cfg.MaxPageSize)
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

const messageRepositoryTemplate = `package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"{{ .Module }}/internal/service"
)

type InMemoryMessageRepository struct {
	mu     sync.RWMutex
	items  []service.Message
	nextID int
}

func NewInMemoryMessageRepository() *InMemoryMessageRepository {
	return &InMemoryMessageRepository{
		items: []service.Message{
			{ID: "msg_1", Title: "Go service template", Content: "A minimal service layer example", Status: "active", CreatedAt: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC)},
			{ID: "msg_2", Title: "Golider release note", Content: "Production defaults for Go backends", Status: "active", CreatedAt: time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC)},
			{ID: "msg_3", Title: "Webhook event", Content: "An incoming event payload", Status: "archived", CreatedAt: time.Date(2023, 11, 20, 8, 15, 0, 0, time.UTC), ArchivedAt: timePointer(time.Date(2023, 12, 1, 8, 15, 0, 0, time.UTC))},
		},
		nextID: 4,
	}
}

func (r *InMemoryMessageRepository) List(ctx context.Context) ([]service.Message, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]service.Message(nil), r.items...), nil
}

func (r *InMemoryMessageRepository) FindByID(ctx context.Context, id string) (service.Message, bool, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.items {
		if item.ID == id {
			return item, true, nil
		}
	}
	return service.Message{}, false, nil
}

func (r *InMemoryMessageRepository) Save(ctx context.Context, message service.Message) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, item := range r.items {
		if item.ID == message.ID {
			r.items[idx] = message
			return nil
		}
	}

	r.items = append([]service.Message{message}, r.items...)
	return nil
}

func (r *InMemoryMessageRepository) NextID(ctx context.Context) (string, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("msg_%d", r.nextID)
	r.nextID++
	return id, nil
}

func timePointer(value time.Time) *time.Time {
	return &value
}
`

const messageRepositoryTestTemplate = `package repository

import (
	"context"
	"testing"
	"time"

	"{{ .Module }}/internal/service"
)

func TestInMemoryMessageRepositorySaveAndFind(t *testing.T) {
	repo := NewInMemoryMessageRepository()

	id, err := repo.NextID(context.Background())
	if err != nil {
		t.Fatalf("NextID 返回错误: %v", err)
	}
	message := service.Message{
		ID:        id,
		Title:     "Repository message",
		Content:   "stored in memory",
		Status:    "active",
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.Save(context.Background(), message); err != nil {
		t.Fatalf("Save 返回错误: %v", err)
	}

	found, ok, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID 返回错误: %v", err)
	}
	if !ok {
		t.Fatal("FindByID 应该找到刚保存的消息")
	}
	if found.Title != message.Title {
		t.Fatalf("消息标题错误，期望 %q，实际 %q", message.Title, found.Title)
	}
}
`

const messageServiceTemplate = `package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

type Message struct {
	ID         string     ` + "`json:\"id\"`" + `
	Title      string     ` + "`json:\"title\"`" + `
	Content    string     ` + "`json:\"content\"`" + `
	Status     string     ` + "`json:\"status\"`" + `
	CreatedAt  time.Time  ` + "`json:\"created_at\"`" + `
	UpdatedAt  *time.Time ` + "`json:\"updated_at,omitempty\"`" + `
	ArchivedAt *time.Time ` + "`json:\"archived_at,omitempty\"`" + `
	DeletedAt  *time.Time ` + "`json:\"deleted_at,omitempty\"`" + `
}

type MessageRepository interface {
	List(context.Context) ([]Message, error)
	FindByID(context.Context, string) (Message, bool, error)
	Save(context.Context, Message) error
	NextID(context.Context) (string, error)
}

type CreateMessageInput struct {
	Title          string
	Content        string
	Status         string
	IdempotencyKey string
}

type CreateMessageOutput struct {
	Message           Message
	IdempotencyReplay bool
}

type UpdateMessageInput struct {
	ID      string
	Title   *string
	Content *string
	Status  *string
}

type UpdateMessageOutput struct {
	Message Message
}

type DeleteMessageOutput struct {
	Message Message
}

type ListMessagesInput struct {
	Page        int
	PageSize    int
	Search      string
	Status      string
	SortBy      string
	SortOrder   string
	CreatedFrom time.Time
	CreatedTo   time.Time
}

type ListMessagesOutput struct {
	Items    []Message
	Page     int
	PageSize int
	Total    int
}

type MessageConflictError struct {
	Title string
}

func (e *MessageConflictError) Error() string {
	return "message title conflict"
}

type IdempotencyConflictError struct {
	Key string
}

func (e *IdempotencyConflictError) Error() string {
	return "idempotency key conflict"
}

type MessageNotFoundError struct {
	ID string
}

func (e *MessageNotFoundError) Error() string {
	return "message not found"
}

type MessageStatusTransitionError struct {
	From string
	To   string
}

func (e *MessageStatusTransitionError) Error() string {
	return "message status transition invalid"
}

type idempotencyRecord struct {
	fingerprint string
	message     Message
}

type MessageService struct {
	mu                sync.Mutex
	repo              MessageRepository
	idempotencyRecord map[string]idempotencyRecord
}

func NewMessageService(repo MessageRepository) *MessageService {
	return &MessageService{
		repo:              repo,
		idempotencyRecord: map[string]idempotencyRecord{},
	}
}

func (s *MessageService) Create(ctx context.Context, input CreateMessageInput) (CreateMessageOutput, error) {
	title := normalizeMessageTitle(input.Title)
	content := strings.TrimSpace(input.Content)
	status := normalizeMessageStatus(input.Status)
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	fingerprint := buildCreateFingerprint(title, content, status)

	s.mu.Lock()
	defer s.mu.Unlock()

	if idempotencyKey != "" {
		if record, ok := s.idempotencyRecord[idempotencyKey]; ok {
			if record.fingerprint != fingerprint {
				return CreateMessageOutput{}, &IdempotencyConflictError{Key: idempotencyKey}
			}
			return CreateMessageOutput{
				Message:           record.message,
				IdempotencyReplay: true,
			}, nil
		}
	}

	items, err := s.repo.List(ctx)
	if err != nil {
		return CreateMessageOutput{}, err
	}
	for _, item := range items {
		if isDeletedMessage(item) {
			continue
		}
		if strings.EqualFold(normalizeMessageTitle(item.Title), title) {
			return CreateMessageOutput{}, &MessageConflictError{Title: title}
		}
	}

	id, err := s.repo.NextID(ctx)
	if err != nil {
		return CreateMessageOutput{}, err
	}
	message := Message{
		ID:        id,
		Title:     title,
		Content:   content,
		Status:    status,
		CreatedAt: time.Now().UTC(),
	}
	if status == "archived" {
		archivedAt := message.CreatedAt
		message.ArchivedAt = &archivedAt
	}
	if err := s.repo.Save(ctx, message); err != nil {
		return CreateMessageOutput{}, err
	}

	if idempotencyKey != "" {
		s.idempotencyRecord[idempotencyKey] = idempotencyRecord{
			fingerprint: fingerprint,
			message:     message,
		}
	}

	return CreateMessageOutput{Message: message}, nil
}

func (s *MessageService) GetByID(ctx context.Context, id string) (Message, error) {
	item, ok, err := s.repo.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return Message{}, err
	}
	if !ok || isDeletedMessage(item) {
		return Message{}, &MessageNotFoundError{ID: strings.TrimSpace(id)}
	}
	return item, nil
}

func (s *MessageService) Update(ctx context.Context, input UpdateMessageInput) (UpdateMessageOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, err := s.GetByID(ctx, input.ID)
	if err != nil {
		return UpdateMessageOutput{}, err
	}

	next := current
	if input.Title != nil {
		next.Title = normalizeMessageTitle(*input.Title)
	}
	if input.Content != nil {
		next.Content = strings.TrimSpace(*input.Content)
	}
	if input.Status != nil {
		targetStatus := normalizeMessageStatus(*input.Status)
		if !canTransitionMessageStatus(current.Status, targetStatus) {
			return UpdateMessageOutput{}, &MessageStatusTransitionError{
				From: current.Status,
				To:   targetStatus,
			}
		}
		next.Status = targetStatus
		if targetStatus == "archived" && next.ArchivedAt == nil {
			archivedAt := time.Now().UTC()
			next.ArchivedAt = &archivedAt
		}
	}

	items, err := s.repo.List(ctx)
	if err != nil {
		return UpdateMessageOutput{}, err
	}
	for _, item := range items {
		if item.ID == next.ID || isDeletedMessage(item) {
			continue
		}
		if strings.EqualFold(normalizeMessageTitle(item.Title), normalizeMessageTitle(next.Title)) {
			return UpdateMessageOutput{}, &MessageConflictError{Title: next.Title}
		}
	}

	updatedAt := time.Now().UTC()
	next.UpdatedAt = &updatedAt
	if err := s.repo.Save(ctx, next); err != nil {
		return UpdateMessageOutput{}, err
	}

	return UpdateMessageOutput{Message: next}, nil
}

func (s *MessageService) Delete(ctx context.Context, id string) (DeleteMessageOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, err := s.GetByID(ctx, id)
	if err != nil {
		return DeleteMessageOutput{}, err
	}

	now := time.Now().UTC()
	item.UpdatedAt = &now
	item.DeletedAt = &now
	if err := s.repo.Save(ctx, item); err != nil {
		return DeleteMessageOutput{}, err
	}

	return DeleteMessageOutput{Message: item}, nil
}

func (s *MessageService) List(ctx context.Context, input ListMessagesInput) ListMessagesOutput {
	page := input.Page
	if page <= 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	items, err := s.repo.List(ctx)
	if err != nil {
		return ListMessagesOutput{Page: page, PageSize: pageSize}
	}

	filtered := make([]Message, 0, len(items))
	search := strings.ToLower(strings.TrimSpace(input.Search))
	status := strings.TrimSpace(input.Status)
	for _, item := range items {
		if isDeletedMessage(item) {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(item.Title), search) && !strings.Contains(strings.ToLower(item.Content), search) {
			continue
		}
		if !input.CreatedFrom.IsZero() && item.CreatedAt.Before(input.CreatedFrom) {
			continue
		}
		if !input.CreatedTo.IsZero() && item.CreatedAt.After(input.CreatedTo) {
			continue
		}
		filtered = append(filtered, item)
	}

	sortBy := strings.TrimSpace(input.SortBy)
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := strings.TrimSpace(input.SortOrder)
	if sortOrder == "" {
		sortOrder = "desc"
	}
	sort.Slice(filtered, func(i int, j int) bool {
		switch sortBy {
		case "title":
			left := strings.ToLower(filtered[i].Title)
			right := strings.ToLower(filtered[j].Title)
			if sortOrder == "asc" {
				return left < right
			}
			return left > right
		default:
			if sortOrder == "asc" {
				return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
			}
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		}
	})

	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return ListMessagesOutput{
		Items:    filtered[start:end],
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
}

func normalizeMessageTitle(title string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
}

func normalizeMessageStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active"
	}
	return status
}

func buildCreateFingerprint(title string, content string, status string) string {
	return strings.ToLower(title) + "\n" + content + "\n" + status
}

func canTransitionMessageStatus(from string, to string) bool {
	from = normalizeMessageStatus(from)
	to = normalizeMessageStatus(to)
	if from == to {
		return true
	}
	if from == "active" && to == "archived" {
		return true
	}
	return false
}

func isDeletedMessage(message Message) bool {
	return message.DeletedAt != nil
}
`

const messageServiceTestTemplate = `package service

import (
	"context"
	"strconv"
	"testing"
	"time"
)

type stubMessageRepository struct {
	items  []Message
	nextID int
}

func newStubMessageRepository() *stubMessageRepository {
	return &stubMessageRepository{
		items: []Message{
			{ID: "msg_1", Title: "Go service template", Content: "A minimal service layer example", Status: "active", CreatedAt: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC)},
			{ID: "msg_2", Title: "Golider release note", Content: "Production defaults for Go backends", Status: "active", CreatedAt: time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC)},
			{ID: "msg_3", Title: "Webhook event", Content: "An incoming event payload", Status: "archived", CreatedAt: time.Date(2023, 11, 20, 8, 15, 0, 0, time.UTC), ArchivedAt: timePointerForTest(time.Date(2023, 12, 1, 8, 15, 0, 0, time.UTC))},
		},
		nextID: 4,
	}
}

func (r *stubMessageRepository) List(ctx context.Context) ([]Message, error) {
	_ = ctx
	return append([]Message(nil), r.items...), nil
}

func (r *stubMessageRepository) FindByID(ctx context.Context, id string) (Message, bool, error) {
	_ = ctx
	for _, item := range r.items {
		if item.ID == id {
			return item, true, nil
		}
	}
	return Message{}, false, nil
}

func (r *stubMessageRepository) Save(ctx context.Context, message Message) error {
	_ = ctx
	for idx, item := range r.items {
		if item.ID == message.ID {
			r.items[idx] = message
			return nil
		}
	}
	r.items = append([]Message{message}, r.items...)
	return nil
}

func (r *stubMessageRepository) NextID(ctx context.Context) (string, error) {
	_ = ctx
	id := "msg_" + strconv.Itoa(r.nextID)
	r.nextID++
	return id, nil
}

func newTestMessageService() *MessageService {
	return NewMessageService(newStubMessageRepository())
}

func timePointerForTest(value time.Time) *time.Time {
	return &value
}

func TestMessageServiceCreateSuccess(t *testing.T) {
	svc := newTestMessageService()

	result, err := svc.Create(context.Background(), CreateMessageInput{
		Title:          "New message",
		Content:        "Create endpoint example",
		Status:         "active",
		IdempotencyKey: "create-1",
	})
	if err != nil {
		t.Fatalf("Create 返回错误: %v", err)
	}
	if result.IdempotencyReplay {
		t.Fatal("首次创建不应被识别为幂等回放")
	}
	if result.Message.ID == "" {
		t.Fatal("创建结果缺少消息 ID")
	}
	if result.Message.Status != "active" {
		t.Fatalf("消息状态错误，期望 %q，实际 %q", "active", result.Message.Status)
	}
	if result.Message.UpdatedAt != nil {
		t.Fatal("新建消息默认不应带 updated_at")
	}
}

func TestMessageServiceCreateConflict(t *testing.T) {
	svc := newTestMessageService()

	_, err := svc.Create(context.Background(), CreateMessageInput{
		Title:   "Go service template",
		Content: "duplicate title",
	})
	if err == nil {
		t.Fatal("Create 本应拒绝重复标题")
	}
	if _, ok := err.(*MessageConflictError); !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
}

func TestMessageServiceCreateIdempotencyReplay(t *testing.T) {
	svc := newTestMessageService()
	input := CreateMessageInput{
		Title:          "Replay message",
		Content:        "same payload",
		Status:         "active",
		IdempotencyKey: "replay-1",
	}

	first, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("首次 Create 返回错误: %v", err)
	}
	second, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("幂等回放 Create 返回错误: %v", err)
	}
	if !second.IdempotencyReplay {
		t.Fatal("第二次请求应被识别为幂等回放")
	}
	if first.Message.ID != second.Message.ID {
		t.Fatalf("幂等回放应返回相同消息，首次 %q，第二次 %q", first.Message.ID, second.Message.ID)
	}
}

func TestMessageServiceCreateIdempotencyConflict(t *testing.T) {
	svc := newTestMessageService()

	_, err := svc.Create(context.Background(), CreateMessageInput{
		Title:          "Replay message",
		Content:        "first payload",
		Status:         "active",
		IdempotencyKey: "replay-2",
	})
	if err != nil {
		t.Fatalf("首次 Create 返回错误: %v", err)
	}

	_, err = svc.Create(context.Background(), CreateMessageInput{
		Title:          "Replay message",
		Content:        "second payload",
		Status:         "active",
		IdempotencyKey: "replay-2",
	})
	if err == nil {
		t.Fatal("Create 本应拒绝不同负载复用同一个幂等键")
	}
	if _, ok := err.(*IdempotencyConflictError); !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
}

func TestMessageServiceGetByIDSuccess(t *testing.T) {
	svc := newTestMessageService()

	message, err := svc.GetByID(context.Background(), "msg_1")
	if err != nil {
		t.Fatalf("GetByID 返回错误: %v", err)
	}
	if message.Title != "Go service template" {
		t.Fatalf("消息标题错误，期望 %q，实际 %q", "Go service template", message.Title)
	}
}

func TestMessageServiceUpdateSuccess(t *testing.T) {
	svc := newTestMessageService()
	title := "Go service template updated"
	content := "updated content"
	status := "archived"

	result, err := svc.Update(context.Background(), UpdateMessageInput{
		ID:      "msg_1",
		Title:   &title,
		Content: &content,
		Status:  &status,
	})
	if err != nil {
		t.Fatalf("Update 返回错误: %v", err)
	}
	if result.Message.Title != title {
		t.Fatalf("更新后的标题错误，期望 %q，实际 %q", title, result.Message.Title)
	}
	if result.Message.Status != status {
		t.Fatalf("更新后的状态错误，期望 %q，实际 %q", status, result.Message.Status)
	}
	if result.Message.UpdatedAt == nil {
		t.Fatal("更新后应写入 updated_at")
	}
	if result.Message.ArchivedAt == nil {
		t.Fatal("归档后应写入 archived_at")
	}
}

func TestMessageServiceUpdateNotFound(t *testing.T) {
	svc := newTestMessageService()
	status := "archived"

	_, err := svc.Update(context.Background(), UpdateMessageInput{
		ID:     "msg_404",
		Status: &status,
	})
	if err == nil {
		t.Fatal("Update 本应在资源不存在时返回错误")
	}
	if _, ok := err.(*MessageNotFoundError); !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
}

func TestMessageServiceUpdateTransitionConflict(t *testing.T) {
	svc := newTestMessageService()
	status := "active"

	_, err := svc.Update(context.Background(), UpdateMessageInput{
		ID:     "msg_3",
		Status: &status,
	})
	if err == nil {
		t.Fatal("Update 本应拒绝不允许的状态回退")
	}
	if _, ok := err.(*MessageStatusTransitionError); !ok {
		t.Fatalf("错误类型不正确: %T", err)
	}
}

func TestMessageServiceDeleteSuccess(t *testing.T) {
	svc := newTestMessageService()

	result, err := svc.Delete(context.Background(), "msg_1")
	if err != nil {
		t.Fatalf("Delete 返回错误: %v", err)
	}
	if result.Message.DeletedAt == nil {
		t.Fatal("软删除后应写入 deleted_at")
	}
	if _, err := svc.GetByID(context.Background(), "msg_1"); err == nil {
		t.Fatal("软删除后不应还能查询到消息")
	}
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
