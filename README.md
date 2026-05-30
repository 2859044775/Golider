# Golider

为 AI 时代生成生产可用的 Go 后端工程。

Golider 不是另一个只会生成 CRUD 目录的模板，而是一个面向真实服务骨架的 Go 工程脚手架：默认带日志、配置校验、生命周期装配、就绪摘流、基础测试和可持续扩展的模块体系。

## 为什么是 Golider

- 默认值偏生产，而不是只生成一个能跑起来的 `main.go`
- 对 AI 协作友好，工程结构清晰、规则明确、扩展锚点稳定
- 支持在已有项目上按需追加能力，而不是每次都从头换模板
- 提供 `doctor`、`doctor fix`、`verify-config` 这类工程治理命令

## 安装

```bash
go install github.com/golider/golider@latest
```

如果你正在本地开发 Golider，也可以直接运行：

```bash
go run . help
```

## 快速开始

```bash
golider new demo --module github.com/acme/demo
cd demo
cp .env.example .env
make run
```

默认生成的服务会提供：

- `GET /`
- `GET /healthz`
- `GET /readyz`

## 默认工程能力

- 统一日志、请求标识、请求超时、请求日志和 panic recover
- 统一错误模型，错误响应自动附带 `request_id`
- 配置加载与显式校验，默认覆盖 `PORT`、`SHUTDOWN_TIMEOUT`、`LOG_LEVEL`、`REQUEST_TIMEOUT`
- 生命周期装配层，统一管理启动钩子和停机钩子
- 真实 `readyz` 状态管理，停机前先摘流再执行优雅停机
- 基础 HTTP 回归测试，默认覆盖默认路由、错误响应和超时行为
- Dockerfile、GitHub Actions、Makefile、`.env.example`

## 核心命令

- `golider new`：生成最小可运行且带生产默认值的 Go API 工程
- `golider add`：为现有工程追加模块能力
- `golider verify`：校验目标工程是否具备 Golider 最小结构
- `golider verify-config`：校验 `.env.example` 是否完整且值是否合法
- `golider doctor`：检查目标工程缺少哪些基础文件、能力和配置
- `golider doctor fix`：自动补齐常见通用能力
- `golider version`：输出当前 CLI 版本信息

## add 模块

当前支持：

- `docker`
- `ci`
- `gitignore`
- `worker`
- `webhook`
- `auth`
- `postgres`
- `request-id`
- `timeout`
- `metrics`
- `rate-limit`
- `error-model`
- `cors`

其中比较核心的模块方向：

- `worker`：生成独立 worker 入口，并接入统一生命周期装配
- `postgres`：补数据库检查命令、`/db/readyz` 路由和生命周期占位管理器
- `request-id`、`timeout`、`metrics`、`rate-limit`、`cors`：补中间件链路与对应配置
- `error-model`：统一错误返回结构，并和 recover 流程打通

## 常用示例

```bash
golider version
golider new demo --module github.com/acme/demo
golider add postgres ./demo
golider add worker ./demo
golider verify ./demo
golider verify-config ./demo
golider doctor ./demo
golider doctor fix ./demo
```

## 发布信息

- 当前版本：`0.1.0`
- Go 最低版本：`1.20`
- 开源协议：`MIT`

## 路线方向

- 继续强化默认工程护栏，让生成结果更接近真实线上服务骨架
- 持续扩展 `add` 模块，覆盖更多常见后端能力
- 打磨发布体验，包括版本发布、安装方式和示例项目展示
