# 验收记录

- 日期：2026-08-22
- 静态检查：Go 1.22.12 下 `go test ./...`、`go build ./...` 通过；Angular 类型检查和 Vite 生产构建通过；`docker compose config --quiet` 通过。
- 容器启动：MySQL、Redis、backend、frontend 均达到 healthy，`GET /healthz` 返回 200。
- API 流程：管理员登录、概览、4 个实体列表、创建采样批次、状态迁移、会话、脱敏运行配置、审计列表及审计汇总均通过。
- RBAC 与双人复核：验证 `draft` 不可跳过 `peer_review`；提交复核者不可自签；operator 不可签发；独立 reviewer 可签发并记录提交人、复核人和签发人。
- 内置 Browser：验证采样批次、实验室样本、检测方法、结果复核、审计 5 个页面；搜索、重置、新增、状态确认弹窗、`ChainBadge`、`ChainTimeline`、`MethodSelector` 和审计回显均正常；控制台 0 error / 0 warning。
- 规模：3045 行 Go 功能代码，38 个非测试 `.go` 文件。
- 清理：已执行 `docker compose down -v --remove-orphans`，无项目容器和数据卷残留。
