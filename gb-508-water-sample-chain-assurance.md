请生成 `water-sample-chain-assurance`「水质检测样本链路审核」Go 全栈项目，面向环境检测实验室管理采样批次、样本交接、检测方法和结果复核。它是实验室质量保证系统，不是大众记录工具，也不做账务、预约或工单。

## 项目主要需求

复杂度下限：核心实体不少于 3 个、核心页面不少于 4 个、横切关注点不少于 2 个、共享前端组件不少于 3 个、自定义 hooks/utils 不少于 2 个、后端中间件不少于 2 个。

### 核心实体

`SamplingBatch`（采样批次与地点）、`LabSample`（样本和保存条件）、`AssayMethod`（方法版本与适用范围）、`ResultReview`（结果复核与签发）贯穿数据库、Go 分层和前端。

### 核心页面

`/sampling-batches` 采样批次；`/samples` 样本接收；`/methods` 方法版本；`/reviews` 结果签发；`/audit` 审计。`ChainBadge` 在批次和样本页共用，`MethodSelector` 在方法和复核页共用。

### 横切关注点

RBAC + 双人复核联动角色表、Go middleware、前端守卫和按钮；审计日志保存样本链路、方法版本、操作者和 request ID；实现错误处理与限流。

### 共享枚举/组件

同步 `SampleState`（received/accepted/testing/hold/disposed）与 `ReviewState`（draft/peer_review/signed/rejected）。共享 `StatusBadge`、`ChainTimeline`、`EmptyState`，hooks 为 `useAuth`、`usePagination`。

### 技术与规模要求

前端 Angular 17 + TypeScript + Vite；后端 Go 1.22 + Gin + GORM；MySQL + Redis。目标 2700–3900 行、26–38 个 `.go` 文件，严格保持多文件分层。

### 文件结构强制清单

前端使用 `api/stores/types/components/common/hooks/pages/router/utils`；后端使用 `model/dto/repository/service/handler/router/middleware/constants/util`，README 列出枚举位置。

### 结构红线

严禁合并职责到单一文件；样本链路和结果复核必须保持跨层拆分。

### 部署与交付

根目录必须提供 `docker-compose.yml`（顶层 `name: water-sample-chain-assurance`，且不写 `version:`）、`.env` 和 `.env.example`（均含 `COMPOSE_PROJECT_NAME=water-sample-chain-assurance`）、`README.md`、`frontend/Dockerfile`、`backend/Dockerfile` 和 `frontend/nginx.conf`。前端端口 `18508`、后端端口 `19508`；Nginx 反代 `/api`、数据库健康检查、命名卷和 `condition: service_healthy` 全部提供，并提供真实 `/healthz`、Git 初始化。
