# Changelog

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

- `golider add redis`：Redis 连接检查、`/redis/readyz` 端点、生命周期管理
- `golider add grpc`：gRPC 服务入口、proto 模板、Greeter 示例服务、反射注册
- `golider add kafka`：Kafka 消费者/生产者模板、生命周期管理

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
