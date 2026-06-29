# Golider

<p align="center">
  <img src="WechatIMG730.jpg" alt="Golider" width="600">
</p>

<p align="center">
  <strong>为 AI 时代生成生产可用的 Go 后端工程</strong>
</p>

<p align="center">
  <a href="https://github.com/2859044775/Golider"><img src="https://img.shields.io/badge/version-0.5.0-blue" alt="version"></a>
  <a href="https://github.com/2859044775/Golider/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-green" alt="license"></a>
  <a href="https://github.com/2859044775/Golider"><img src="https://img.shields.io/badge/go-%3E%3D1.20-00ADD8?logo=go" alt="go version"></a>
</p>

<p align="center">
  <a href="https://github.com/2859044775/Golider">GitHub</a> ·
  <a href="https://gitee.com/eason4798_admin/Golider">Gitee</a> ·
  <a href="https://gitcode.com/gcw_a5oyjfMg/Golider">GitCode</a>
</p>

---

Golider 不是另一个只会生成 CRUD 目录的模板。它是一个面向**真实服务骨架**的 Go 工程脚手架——默认具备结构化日志（支持 JSON 格式）、Prometheus 指标、安全响应头、输入校验、查询解析、分页、排序、过滤、仓储抽象、幂等写入、冲突校验、状态流转、软删除、审计字段、配置校验、生命周期、就绪摘流、深度健康检查、超时护栏、TLS 支持、分布式追踪和基础测试。生成即具备生产讨论的基础。

## 目录

- [为什么是 Golider](#为什么是-golider)
- [快速开始](#快速开始)
- [生成项目结构](#生成项目结构)
- [默认接口](#默认接口)
- [默认工程能力](#默认工程能力)
- [核心命令](#核心命令)
- [add 模块](#add-模块)
- [常见示例](#常见示例)
- [发布信息](#发布信息)
- [路线方向](#路线方向)

---

## 为什么是 Golider

| 出发点 | 说明 |
|--------|------|
| **默认偏生产** | 不只是能跑起来的 `main.go`，默认就带日志、超时、校验、就绪摘流和错误模型 |
| **AI 协作友好** | 工程结构清晰、规则明确、扩展锚点稳定，方便你和 AI 一起在上面迭代 |
| **模块化追加** | 支持在已有项目上 `add` 能力，而不是每次都从头换模板 |
| **工程治理** | 提供 `doctor`、`doctor fix`、`verify`、`verify-config` 这类诊断与修复命令 |

---

## 安装

```bash
# 方式一：go install（推荐）
go install github.com/2859044775/Golider@latest

# 方式二：Homebrew
brew install 2859044775/Golider/golider
```

本地开发时也可以直接运行：

```bash
go run . help
```

---

## 快速开始

```bash
# 1. 生成项目
Golider new demo --module github.com/acme/demo

# 2. 进入目录
cd demo

# 3. 复制环境变量模板
cp .env.example .env

# 4. 启动服务
make run
```

终端预览：

```
== 项目生成完成 ==
[完成] 已生成项目 demo
下一步
  cd demo
  go run ./cmd/api
```

---

## 生成项目结构

```
demo/
├── .env.example               # 环境变量模板
├── .gitignore                  # Git 忽略规则
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions CI
├── Dockerfile                  # 容器构建文件
├── Makefile                    # 常用命令入口
├── README.md                   # 项目说明
├── go.mod
├── cmd/
│   └── api/
│       └── main.go             # 服务入口
└── internal/
    ├── app/
    │   ├── app.go              # 生命周期装配
    │   └── dependencies.go     # 依赖装配
    ├── config/
    │   └── config.go           # 配置加载与校验
    ├── http/
    │   ├── binding.go          # JSON 输入校验
    │   ├── binding_test.go
    │   ├── errors.go           # 统一错误模型
    │   ├── metrics.go          # Prometheus 指标
    │   ├── middleware.go       # 中间件装配
    │   ├── middleware_test.go
    │   ├── query.go            # 查询参数解析
    │   ├── query_test.go
    │   ├── readiness.go        # 就绪摘流
    │   ├── requestid.go        # 请求标识
    │   ├── router.go           # 路由定义
    │   ├── router_test.go
    │   ├── timeout.go          # 请求超时
    │   └── tracing.go          # 分布式追踪（W3C Trace Context）
    ├── observability/
    │   └── logger.go           # 统一日志
    ├── repository/
    │   └── message.go          # 内存仓储
    └── service/
        └── message.go          # 业务服务层
```

---

## 默认接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/` | 欢迎页面 |
| `GET` | `/healthz` | 健康检查（含依赖检查） |
| `GET` | `/readyz` | 就绪检查 |
| `GET` | `/messages` | 消息列表（支持分页、搜索、排序、状态过滤、时间范围） |
| `POST` | `/messages` | 创建消息（支持 `Idempotency-Key` 幂等写入） |
| `GET` | `/messages/{id}` | 消息详情 |
| `PATCH` | `/messages/{id}` | 局部更新消息（支持状态流转校验） |
| `DELETE` | `/messages/{id}` | 软删除消息 |
| `POST` | `/echo` | 请求回显（输入校验示例） |
| `GET` | `/metrics` | Prometheus 标准指标 |

**分页查询示例：**

```
GET /messages?page=1&page_size=20&search=hello&status=active&sort_by=created_at&sort_order=desc&created_from=2024-01-01T00:00:00Z
```

**创建消息示例：**

```bash
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: a1b2c3d4" \
  -d '{"title": "Hello World", "content": "Golider is awesome"}'
```

---

## 默认工程能力

### 基础层
| 能力 | 说明 |
|------|------|
| 安全响应头 | 默认添加 `X-Content-Type-Options`、`X-Frame-Options`、`X-XSS-Protection`、`Referrer-Policy` 等安全头 |
| 统一日志 | 结构化日志，按级别输出，支持 `LOG_FORMAT=json` 切换 JSON 格式 |
| 请求标识 | 每个请求自动注入 `X-Request-ID` |
| 请求超时 | 通过 `REQUEST_TIMEOUT` 配置，默认 5 秒 |
| Panic Recovery | 自动捕获 panic 并返回统一错误 |
| 统一错误模型 | `{ "code": "...", "message": "...", "request_id": "..." }` |
| Prometheus 指标 | 默认 `/metrics` 端点，含请求计数、延迟直方图、状态码分类 |
| 分布式追踪 | W3C Trace Context 上下文传播，自动解析/生成 `traceparent`，日志关联 `trace_id` |

### 传输层
| 能力 | 说明 |
|------|------|
| JSON 输入校验 | 必填检查、类型校验、未知字段拒绝、单 JSON 对象约束 |
| 查询参数解析 | 分页、搜索、状态过滤、排序、时间范围 |
| 统一分页响应 | `{ "items": [...], "page": 1, "page_size": 10, "total": 100 }` |
| 幂等写入 | `Idempotency-Key` 同键回放返回已创建资源，不同负载返回 `409` |
| 冲突校验 | 标题重复 → `409 message_title_conflict` |
| 乐观锁 | 版本号驱动的并发冲突检测，`PATCH`/`DELETE` 可选传入 `version` |
| 状态流转 | 消息从 `active` 可归档到 `archived`，不可回退 |

### 服务层
| 能力 | 说明 |
|------|------|
| 仓储抽象 | 通过 `MessageRepository` 接口隔离数据访问 |
| 内存仓储 | 默认提供内存实现，可直接切换数据库 |
| 软删除 | `DELETE` 仅标记 `deleted_at`，不物理删除 |
| 审计字段 | `updated_at`、`archived_at`、`deleted_at` |

### 基础设施层
| 能力 | 说明 |
|------|------|
| 配置校验 | 端口范围、日志级别、超时合法性、分页默认值合理性、请求体大小与请求头大小限制 |
| 请求体限制 | 通过 `BODY_LIMIT_BYTES` 配置，默认 1MB |
| pprof 诊断 | 通过 `ENABLE_PPROF=true` 开启 `/debug/pprof/` 性能分析端点 |
| TLS/HTTPS | 通过 `TLS_CERT`/`TLS_KEY` 配置 HTTPS，支持 `ListenAndServeTLS` |
| 深度健康检查 | `/healthz` 支持注册依赖检查函数，检查失败返回 `503` |
| 生命周期装配 | 启动钩子和停止钩子统一管理 |
| 就绪摘流 | 停机前先切换到未就绪状态 |
| HTTP 超时护栏 | `ReadHeaderTimeout`、`ReadTimeout`、`WriteTimeout`、`IdleTimeout` |
| 基础测试 | 路由测试、中间件测试、输入校验测试、查询解析测试、服务层测试 |
| 容器化 | Dockerfile + `.dockerignore` |
| CI | GitHub Actions 自动化构建与测试 |

---

## 核心命令

| 命令 | 功能 |
|------|------|
| `Golider new` | 生成带生产默认值的 Go API 工程 |
| `Golider add <模块> [目录]` | 为现有工程追加模块能力 |
| `Golider verify [目录]` | 校验目标工程是否具备最小结构 |
| `Golider verify-config [目录]` | 校验 `.env.example` 完整性与值的合法性 |
| `Golider doctor [目录]` | 检查缺少哪些基础文件、能力和配置 |
| `Golider doctor fix [目录]` | 自动补齐常见通用能力 |
| `Golider version` | 输出版本信息 |

---

## add 模块

### 基础设施

| 模块 | 说明 |
|------|------|
| `docker` | 生成 Dockerfile |
| `ci` | 生成 GitHub Actions CI 配置 |
| `gitignore` | 生成 `.gitignore` |

### 中间件与传输

| 模块 | 说明 |
|------|------|
| `request-id` | 注入请求标识中间件 |
| `timeout` | 注入请求超时中间件 |
| `metrics` | 注入 `/metrics` 路由与计数中间件 |
| `rate-limit` | 注入限流中间件 |
| `cors` | 注入跨域中间件 |
| `error-model` | 统一错误返回结构，接入 recover |
| `circuit-breaker` | 注入熔断器中间件，保护下游依赖 |

### 业务扩展

| 模块 | 说明 |
|------|------|
| `auth` | 鉴权示例：`/auth/login` 路由与 `AUTH_TOKEN` 配置 |
| `webhook` | Webhook 示例：`/webhooks/example` 接收接口 |
| `postgres` | 数据库检查命令、`/db/readyz` 路由、生命周期管理器、PostgreSQL 仓储实现、数据库迁移模板 |
| `redis` | Redis 连接检查、`/redis/readyz` 路由、生命周期管理器 |
| `grpc` | gRPC 服务入口、proto 模板文件、Greeter 示例服务、反射注册 |
| `kafka` | Kafka 消费者/生产者模板、生命周期管理 |
| `worker` | 独立 worker 入口，接入生命周期装配 |

---

## 常见示例

```bash
# 查看版本
Golider version

# 生成新项目
Golider new demo --module github.com/acme/demo

# 追加数据库能力
Golider add postgres ./demo

# 追加 worker 能力
Golider add worker ./demo

# 工程结构校验
Golider verify ./demo

# 配置校验
Golider verify-config ./demo

# 诊断缺失能力
Golider doctor ./demo

# 自动修复
Golider doctor fix ./demo
```

---

## 发布信息

| 项目 | 内容 |
|------|------|
| 当前版本 | `0.5.0` |
| Go 最低版本 | `1.20` |
| 开源协议 | `MIT` |
| 代码仓库 | [GitHub](https://github.com/2859044775/Golider) · [Gitee](https://gitee.com/eason4798_admin/Golider) · [GitCode](https://gitcode.com/gcw_a5oyjfMg/Golider) |

---

## 路线方向

- [x] 继续强化默认工程护栏（安全响应头、请求头大小限制、可配置请求体限制、pprof 诊断端点）
- [x] 持续扩展 `add` 模块，覆盖更多常见后端能力（Redis、gRPC、Kafka 等）
- [x] 增加数据库可切换仓储占位与乐观锁/版本字段
- [x] 打磨发布体验（Homebrew 安装、版本发布）
- [x] 生成项目运行日志做分彩级别输出
- [x] `doctor` 做成更强的表格感输出，成功项折叠、突出异常项
- [x] 结构化 JSON 日志（`LOG_FORMAT=json`）
- [x] Prometheus 标准指标（延迟直方图、状态码分类）
- [x] TLS/HTTPS 支持（`TLS_CERT`/`TLS_KEY`）
- [x] 深度健康检查（`/healthz` 依赖检查）
- [x] 分布式追踪（W3C Trace Context 上下文传播）
- [x] 熔断器模块（`Golider add circuit-breaker`）
- [ ] WebSocket 模块（`Golider add websocket`）
- [ ] 定时任务模块（`Golider add scheduler`）
- [ ] GraphQL 模块（`Golider add graphql`）
- [ ] 多租户支持（`Golider add multi-tenant`）
