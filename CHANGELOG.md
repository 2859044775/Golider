# Changelog

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
