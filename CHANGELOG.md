# Changelog

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
