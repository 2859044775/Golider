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
		{Name: "依赖装配", Path: filepath.Join("internal", "app", "dependencies.go")},
		{Name: "配置模块", Path: filepath.Join("internal", "config", "config.go")},
		{Name: "HTTP 中间件", Path: filepath.Join("internal", "http", "middleware.go")},
		{Name: "HTTP 路由", Path: filepath.Join("internal", "http", "router.go")},
		{Name: "消息仓储", Path: filepath.Join("internal", "repository", "message.go")},
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
	serviceFile := readMaybe(filepath.Join(projectDir, "internal", "service", "message.go"))
	depsFile := readMaybe(filepath.Join(projectDir, "internal", "app", "dependencies.go"))
	repositoryFile := readMaybe(filepath.Join(projectDir, "internal", "repository", "message.go"))
	postgresRepoFile := readMaybe(filepath.Join(projectDir, "internal", "repository", "message_postgres.go"))
	redisStoreFile := readMaybe(filepath.Join(projectDir, "internal", "store", "redis.go"))

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
			Name:    "安全响应头",
			Exists:  strings.Contains(middleware, "securityHeadersMiddleware") && strings.Contains(middleware, "X-Content-Type-Options"),
			Detail:  "默认添加 nosniff、frame、XSS 等安全响应头",
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
			Name:    "输入校验",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "binding.go")) && strings.Contains(router, `mux.HandleFunc("/echo"`) && strings.Contains(router, "decodeJSON(") && strings.Contains(router, "validateCreateMessageRequest"),
			Detail:  "默认包含 JSON 解码、请求体校验和写接口参数校验",
			Related: "internal/http/binding.go",
		},
		{
			Name:    "查询解析",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "query.go")) && strings.Contains(router, `mux.HandleFunc("/messages"`) && strings.Contains(router, "parseListQuery("),
			Detail:  "默认包含列表查询参数解析与分页响应",
			Related: "internal/http/query.go",
		},
		{
			Name:    "服务分层",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "service", "message.go")) && strings.Contains(router, "deps.MessageService.List(") && strings.Contains(router, "deps.MessageService.Create(") && strings.Contains(router, "deps.MessageService.Update(") && strings.Contains(router, "deps.MessageService.Delete("),
			Detail:  "默认包含查询、创建、更新、删除服务层示例",
			Related: "internal/service/message.go",
		},
		{
			Name:    "仓储抽象",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "repository", "message.go")) && strings.Contains(serviceFile, "type MessageRepository interface") && strings.Contains(depsFile, "repository.NewInMemoryMessageRepository()") && strings.Contains(repositoryFile, "type InMemoryMessageRepository struct"),
			Detail:  "默认通过仓储接口隔离数据访问",
			Related: "internal/repository/message.go",
		},
		{
			Name:    "幂等写入",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "service", "message.go")) && strings.Contains(router, "Idempotency-Key") && strings.Contains(router, "idempotency_key_conflict") && strings.Contains(serviceFile, "IdempotencyReplay"),
			Detail:  "默认创建接口支持幂等键回放与冲突识别",
			Related: "internal/service/message.go",
		},
		{
			Name:    "资源详情",
			Exists:  strings.Contains(router, `mux.HandleFunc("/messages/"`) && strings.Contains(router, "func getMessageHandler") && strings.Contains(serviceFile, "func (s *MessageService) GetByID"),
			Detail:  "默认包含单资源查询接口",
			Related: "internal/http/router.go",
		},
		{
			Name:    "状态流转校验",
			Exists:  strings.Contains(router, "func updateMessageHandler") && strings.Contains(router, "message_status_transition_invalid") && strings.Contains(serviceFile, "func canTransitionMessageStatus"),
			Detail:  "默认更新接口包含状态流转约束",
			Related: "internal/service/message.go",
		},
		{
			Name:    "软删除",
			Exists:  strings.Contains(router, "func deleteMessageHandler") && strings.Contains(serviceFile, "func (s *MessageService) Delete") && strings.Contains(serviceFile, "DeletedAt"),
			Detail:  "默认删除接口采用软删除并保留审计字段",
			Related: "internal/service/message.go",
		},
		{
			Name:    "依赖装配",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "app", "dependencies.go")) && strings.Contains(apiMain, "deps := app.NewDependencies(cfg)") && strings.Contains(apiMain, "httptransport.NewRouter(deps)"),
			Detail:  "默认通过应用层统一装配依赖",
			Related: "internal/app/dependencies.go",
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
			Name:    "服务超时护栏",
			Exists:  strings.Contains(apiMain, "ReadHeaderTimeout: cfg.ReadHeaderTimeout") && strings.Contains(apiMain, "MaxHeaderBytes:    cfg.MaxHeaderBytes") && strings.Contains(apiMain, "WriteTimeout:      cfg.WriteTimeout") && strings.Contains(envFile, "HTTP_READ_HEADER_TIMEOUT=") && strings.Contains(envFile, "HTTP_READ_TIMEOUT=") && strings.Contains(envFile, "HTTP_WRITE_TIMEOUT=") && strings.Contains(envFile, "HTTP_IDLE_TIMEOUT=") && strings.Contains(envFile, "MAX_HEADER_BYTES="),
			Detail:  "HTTP 服务默认启用读写超时与请求头大小保护",
			Related: "cmd/api/main.go",
		},
		{
			Name:    "指标采集",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "metrics.go")) && strings.Contains(router, `mux.HandleFunc("/metrics"`) && strings.Contains(middleware, "metricsMiddleware"),
			Detail:  "Prometheus 标准格式 /metrics 端点，含请求计数与延迟直方图",
			Related: "internal/http/metrics.go",
		},
		{
			Name:    "结构化 JSON 日志",
			Exists:  strings.Contains(envFile, "LOG_FORMAT=") && strings.Contains(configFile, "LogFormat"),
			Detail:  "支持 LOG_FORMAT=json 切换为 JSON 结构化日志",
			Related: "internal/observability/logger.go",
		},
		{
			Name:    "TLS/HTTPS 支持",
			Exists:  strings.Contains(envFile, "TLS_CERT=") && strings.Contains(envFile, "TLS_KEY=") && strings.Contains(configFile, "TLSCert") && strings.Contains(apiMain, "ListenAndServeTLS"),
			Detail:  "支持 TLS_CERT/TLS_KEY 配置 HTTPS",
			Related: "cmd/api/main.go",
		},
		{
			Name:    "深度健康检查",
			Exists:  strings.Contains(readinessFile, "RegisterHealthCheck") && strings.Contains(readinessFile, "runHealthChecks"),
			Detail:  "/healthz 支持注册依赖检查函数",
			Related: "internal/http/readiness.go",
		},
		{
			Name:    "分布式追踪",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "tracing.go")) && strings.Contains(middleware, "tracingMiddleware"),
			Detail:  "W3C Trace Context 分布式追踪上下文传播",
			Related: "internal/http/tracing.go",
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
			Name:    "熔断器保护",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "circuitbreaker.go")) && strings.Contains(middleware, "circuitBreakerMiddleware") && strings.Contains(envFile, "CIRCUIT_BREAKER_THRESHOLD="),
			Detail:  "支持 CIRCUIT_BREAKER_THRESHOLD/TIMEOUT/SUCCESS_THRESHOLD 配置",
			Related: "internal/http/circuitbreaker.go",
		},
		{
			Name:    "WebSocket 实时推送",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "http", "websocket.go")) && strings.Contains(router, `mux.HandleFunc("/ws"`) && strings.Contains(router, "websocketHandler"),
			Detail:  "支持 /ws 端点，纯标准库实现，房间订阅，心跳保活",
			Related: "internal/http/websocket.go",
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
		{
			Name:    "乐观锁版本字段",
			Exists:  strings.Contains(serviceFile, "Version") && strings.Contains(serviceFile, "MessageVersionConflictError") && strings.Contains(repositoryFile, "SaveVersioned") && strings.Contains(router, "message_version_conflict"),
			Detail:  "资源更新与删除使用版本号实现乐观锁冲突检测",
			Related: "internal/service/message.go",
		},
		{
			Name:    "数据库仓储占位",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "repository", "message_postgres.go")) && strings.Contains(postgresRepoFile, "NewDatabaseMessageService") && fileExists(filepath.Join(projectDir, "migrations", "001_create_messages.sql")),
			Detail:  "包含 PostgreSQL 仓储实现和数据库迁移模板",
			Related: "internal/repository/message_postgres.go",
		},
		{
			Name:    "Redis 支持",
			Exists:  fileExists(filepath.Join(projectDir, "internal", "store", "redis.go")) && strings.Contains(redisStoreFile, "CheckRedis") && strings.Contains(router, `mux.HandleFunc("/redis/readyz"`) && strings.Contains(envFile, "REDIS_URL="),
			Detail:  "包含 Redis 连接检查和就绪端点",
			Related: "internal/store/redis.go",
		},
		{
			Name:    "gRPC 支持",
			Exists:  fileExists(filepath.Join(projectDir, "cmd", "grpc", "main.go")) && strings.Contains(makefile, "run-grpc:") && strings.Contains(envFile, "GRPC_PORT="),
			Detail:  "包含 gRPC 服务入口、proto 定义和 Greeter 示例",
			Related: "cmd/grpc/main.go",
		},
		{
			Name:    "Kafka 支持",
			Exists:  fileExists(filepath.Join(projectDir, "cmd", "kafka", "main.go")) && strings.Contains(makefile, "run-kafka:") && strings.Contains(envFile, "KAFKA_BROKERS="),
			Detail:  "包含 Kafka 消费者、生产者和生命周期管理",
			Related: "cmd/kafka/main.go",
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
		buildConfigRequirement(envValues, "LOG_FORMAT", "日志格式", func(v string) bool { return v == "text" || v == "json" }),
		buildConfigRequirement(envValues, "HTTP_READ_HEADER_TIMEOUT", "请求头读取超时", validateDurationValue),
		buildConfigRequirement(envValues, "HTTP_READ_TIMEOUT", "请求读取超时", validateDurationValue),
		buildConfigRequirement(envValues, "HTTP_WRITE_TIMEOUT", "响应写入超时", validateDurationValue),
		buildConfigRequirement(envValues, "HTTP_IDLE_TIMEOUT", "连接空闲超时", validateDurationValue),
		buildConfigRequirement(envValues, "DEFAULT_PAGE_SIZE", "默认分页大小", validatePositiveIntValue),
		buildConfigRequirement(envValues, "MAX_PAGE_SIZE", "最大分页大小", validatePositiveIntValue),
		buildConfigRequirement(envValues, "MAX_HEADER_BYTES", "最大请求头字节数", validatePositiveIntValue),
		buildConfigRequirement(envValues, "BODY_LIMIT_BYTES", "请求体大小限制", validatePositiveIntValue),
	)

	middleware := readMaybe(filepath.Join(projectDir, "internal", "http", "middleware.go"))
	router := readMaybe(filepath.Join(projectDir, "internal", "http", "router.go"))
	makefile := readMaybe(filepath.Join(projectDir, "Makefile"))

	if fileExists(filepath.Join(projectDir, "internal", "http", "timeout.go")) || strings.Contains(middleware, "timeoutMiddleware") {
		items = append(items, buildConfigRequirement(envValues, "REQUEST_TIMEOUT", "请求超时", validateDurationValue))
	}
	if fileExists(filepath.Join(projectDir, "internal", "http", "ratelimit.go")) || strings.Contains(middleware, "rateLimitMiddleware") {
		items = append(items, buildConfigRequirement(envValues, "RATE_LIMIT_PER_SECOND", "每秒限流阈值", validatePositiveIntValue))
	}
	if fileExists(filepath.Join(projectDir, "internal", "http", "cors.go")) || strings.Contains(middleware, "corsMiddleware") {
		items = append(items, buildConfigRequirement(envValues, "CORS_ALLOW_ORIGINS", "允许跨域来源", validateNonEmptyValue))
	}
	if fileExists(filepath.Join(projectDir, "internal", "http", "circuitbreaker.go")) || strings.Contains(middleware, "circuitBreakerMiddleware") {
		items = append(items,
			buildConfigRequirement(envValues, "CIRCUIT_BREAKER_THRESHOLD", "熔断器失败阈值", validatePositiveIntValue),
			buildConfigRequirement(envValues, "CIRCUIT_BREAKER_TIMEOUT", "熔断器开启持续时间", validateDurationValue),
			buildConfigRequirement(envValues, "CIRCUIT_BREAKER_SUCCESS_THRESHOLD", "半开状态成功阈值", validatePositiveIntValue),
		)
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
	if fileExists(filepath.Join(projectDir, "internal", "store", "redis.go")) || strings.Contains(router, `mux.HandleFunc("/redis/readyz"`) {
		items = append(items,
			buildConfigRequirement(envValues, "REDIS_URL", "Redis 地址", validateNonEmptyValue),
		)
	}
	if fileExists(filepath.Join(projectDir, "cmd", "grpc", "main.go")) || strings.Contains(makefile, "run-grpc:") {
		items = append(items,
			buildConfigRequirement(envValues, "GRPC_PORT", "gRPC 服务端口", validatePortValue),
		)
	}
	if fileExists(filepath.Join(projectDir, "cmd", "kafka", "main.go")) || strings.Contains(makefile, "run-kafka:") {
		items = append(items,
			buildConfigRequirement(envValues, "KAFKA_BROKERS", "Kafka 集群地址", validateNonEmptyValue),
			buildConfigRequirement(envValues, "KAFKA_TOPIC", "Kafka 主题名称", validateNonEmptyValue),
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
