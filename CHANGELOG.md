# Changelog

## 0.6.0 - 2026-06-29

### 新增

- `Golider add websocket` 模块：纯标准库实现的 WebSocket 实时推送
  - `/ws` 端点，支持房间订阅（`?room=xxx` 查询参数）
  - Hub-Client 架构，支持多房间管理和消息广播
  - 心跳保活（30秒 ping/pong）
  - 运行时房间切换（通过发送 JSON 消息）
  - 导出 `BroadcastToRoom(room, event, payload)` 函数供业务调用

### 改进

- `doctor` 新增 WebSocket 实时推送能力检测

## 0.5.1 - 2026-06-29

### 新增

- `Golider create-module <名称>` 命令：生成自定义 addon 模块骨架到 `internal/addon/modules/`，开发者填写模板代码和 patch 逻辑后重新编译即可使用
- addon 模块注册系统：`addon.RegisterModule()` / `addon.ModuleDefinition`，支持通过 `init()` 自注册模块
- 导出 addon 辅助函数供自定义模块使用：`addon.AddMiddlewareLine()`、`addon.AddRouteLine()`、`addon.AppendEnvValue()`、`addon.EnsureImport()`、`addon.InsertAfter()`、`addon.DetectModulePath()`、`addon.CommonBaseFiles()`
- 内置模块优先级保证：自定义模块不会覆盖同名内置模块

## 0.5.0 - 2026-06-18

### 新增

- 分布式追踪：默认接入 W3C Trace Context 上下文传播，自动解析/生成 `traceparent` 头，日志关联 `trace_id`，纯标准库实现
- `Golider add circuit-breaker`：熔断器中间件模块，支持关闭/开启/半开三态切换，通过 `CIRCUIT_BREAKER_THRESHOLD`/`CIRCUIT_BREAKER_TIMEOUT`/`CIRCUIT_BREAKER_SUCCESS_THRESHOLD` 配置

### 改进

- `doctor` 新增分布式追踪能力检测
- `doctor` 新增熔断器保护能力检测
- `verify-config` 新增熔断器配置校验（`CIRCUIT_BREAKER_THRESHOLD`/`CIRCUIT_BREAKER_TIMEOUT`/`CIRCUIT_BREAKER_SUCCESS_THRESHOLD`）
- 请求日志和异常恢复日志新增 `trace_id` 字段

## 0.4.1 - 2026-06-18

### 新增

- 结构化 JSON 日志：通过 `LOG_FORMAT=json` 切换为 JSON 格式输出，便于日志聚合系统解析
- Prometheus 标准指标：默认 `/metrics` 端点输出 Prometheus text 格式，含请求计数（按 method/status 分类）、延迟直方图、panic 恢复计数
- TLS/HTTPS 支持：通过 `TLS_CERT`/`TLS_KEY` 配置证书，启用 `ListenAndServeTLS`
- 深度健康检查：`/healthz` 支持通过 `RegisterHealthCheck()` 注册依赖检查函数，检查失败返回 `503`

### 改进

- `doctor` 新增结构化 JSON 日志、Prometheus 指标、TLS/HTTPS、深度健康检查能力检测
- `verify-config` 新增 `LOG_FORMAT` 配置校验

## 0.3.1 - 2026-06-16

### 新增

- 安全响应头中间件：默认添加 `X-Content-Type-Options`、`X-Frame-Options`、`X-XSS-Protection`、`Referrer-Policy`、`X-Permitted-Cross-Domain-Policies`
- HTTP Server 请求头大小限制：通过 `MAX_HEADER_BYTES` 环境变量配置
- 请求体大小限制可配置化：通过 `BODY_LIMIT_BYTES` 环境变量配置（原硬编码 1MB）
- pprof 性能诊断端点：通过 `ENABLE_PPROF=true` 开启 `/debug/pprof/`
- Homebrew 安装支持：`brew install 2859044775/Golider/golider`
- README 添加 GitCode 仓库地址

### 改进

- `doctor` 新增安全响应头能力检测
- `verify-config` 新增 `MAX_HEADER_BYTES`、`BODY_LIMIT_BYTES` 配置校验

## 0.3.0 - 2026-06-15

### 新增

- `Golider add redis`：Redis 连接检查、`/redis/readyz` 端点、生命周期管理
- `Golider add grpc`：gRPC 服务入口、proto 模板、Greeter 示例服务、反射注册
- `Golider add kafka`：Kafka 消费者/生产者模板、生命周期管理

### 改进

- `doctor` 新增 Redis、gRPC、Kafka 能力检测
- `verify-config` 新增 `REDIS_URL`、`GRPC_PORT`、`KAFKA_BROKERS`、`KAFKA_TOPIC` 配置校验
- README 添加宣传海报图片

## 0.2.0 - 2026-06-15

### 新增

- 新增 PostgreSQL 仓储实现，支持数据库可切换
- 乐观锁版本字段，`PATCH`/`DELETE` 支持版本号驱动的并发冲突检测
- `doctor` 命令表格化输出，异常项突出显示
- `new`/`add` 命令增强安装进度感知与分彩日志

## 0.1.0 - 2026-05-30

### 新增

- 首个可公开版本的 CLI 命令集：`new`、`add`、`verify`、`verify-config`、`doctor`、`doctor fix`
- 默认生成统一日志、请求标识、请求超时、统一错误模型、配置校验、生命周期钩子与优雅摘流
- 默认生成基础 HTTP 回归测试，覆盖默认路由、错误响应与请求超时
- 模块化扩展能力：`worker`、`webhook`、`auth`、`postgres`、`request-id`、`timeout`、`metrics`、`rate-limit`、`error-model`、`cors`

### 改进

- `doctor` 现在会同时检查基础文件、工程能力和配置模板状态
- `doctor fix` 可自动补齐常见通用工程能力
- `verify-config` 可独立校验 `.env.example` 的完整性和有效性

