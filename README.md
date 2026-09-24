# 商业潜水减压暴露计划校核

`commercial-diving-decompression-control` 是面向商业潜水培训策划人员和复核主管的离线暴露建模工作台。系统保存最小化训练档案、有序深度/时间/气体假设、确定性组织舱负荷结果和人工复核证据。

> **安全边界：本项目不是医疗器械、认证潜水表或现场生命支持系统。不得连接潜水设备，不生成可直接执行的减压停留深度或时长，不给出“安全可潜”结论，也不能替代潜水主管或医疗专业人员。所有结果仅供训练比较和决策支持，必须人工审批。**

## Docker Compose 一键启动

```bash
cp .env.example .env
docker compose up -d --build
docker compose ps
```

- Web 工作台：<http://127.0.0.1:18526>
- 后端健康检查：<http://127.0.0.1:19526/healthz>
- 后端就绪检查：<http://127.0.0.1:19526/readyz>
- API 前缀：<http://127.0.0.1:19526/api/v1>

测试账号：

| 角色 | 用户名 | 密码 | 权限 |
| --- | --- | --- | --- |
| 计划员 | `planner` | `planner123` | 档案、计划、暴露段、模型运行、提交复核 |
| 主管 | `supervisor` | `supervisor123` | 查看方案、人工批准、审计查询 |
| 管理员 | `admin` | `admin123` | 计划员与主管权限 |

停止并只清理本项目容器、网络和命名卷：

```bash
docker compose down -v --remove-orphans
```

## 主要功能

- `DiverProfile`：最小化训练资料、资格等级、默认气体假设和版本，不保存诊断性医疗记录。
- `DivePlan`：工作地点表面压力、呼吸气体、计划时间、输入版本和完整状态流。
- `ExposureSegment`：同一计划内唯一且连续的序号，严格校验深度、时长、上升速率、气体比例和段间连续性。
- `DecompressionAssessment`：不可覆盖的输入快照、六舱负荷曲线、风险证据、比较指数、算法版本和假设。
- 五个业务页：训练档案、计划编排、暴露剖面、评估复核、审计轨迹；图表只消费真实 API 数据。
- JWT/RBAC、请求 ID、结构化访问日志、panic recovery、本地限流、统一错误码、事务、乐观锁和不可普通删除的审计事件。

项目明确不包含排班预约、工单、库存、采购、订单、保险或财务功能。

## 确定性模型

模型版本默认为 `training-compartment-v1`。固定输入会生成固定结果：

1. 环境压力假设：`ambient_bar = worksite_pressure_bar + depth_m / 10`。
2. 吸入惰性气体分压：`(ambient_bar - 0.0627) × gas_fraction`。
3. 每段按常深近似计算指数趋近平衡：`P(t) = P_inspired + (P_initial - P_inspired) × exp(-ln(2) × t / half_time)`。
4. 配置六个训练比较舱：氮半时 5/10/20/40/80/120 分钟，氦半时 2/4/8/16/32/48 分钟。
5. 输出每段环境压力、N2/He 分压、每舱负荷、相对基线变化、输入快照、假设和风险标记。

比较指数只用于对照同一模型下的方案差异，**不是安全评分**。风险阈值是项目内的复核提示，不代表医学、法规或行业认证结论。

## 状态与枚举位置

`PlanStatus = draft | modeled | pending_supervisor_review | approved_for_training | archived`

- 数据库：`dive_plans.plan_status`、`decompression_assessments.assessment_status` 均使用显式 `CHECK` 约束。
- 后端：`backend/internal/constants/plan.go`，并贯穿 `model/dive_plan.go`、`repository/dive_plan.go`、`repository/decompression_assessment.go`、对应 service/handler/router。
- 前端：`frontend/src/types/plan.ts`、`stores/plan.ts`、`stores/assessment.ts`、`components/common/PlanStatusBadge.tsx`、`pages/PlansPage.tsx`、`pages/AssessmentsPage.tsx`、`pages/AuditPage.tsx`。

`RiskBand = informational | caution | elevated | invalid`

- 数据库：评估表的 `risk_flags_json` 保存完整不可覆盖证据，`highest_risk_band` 使用显式 `CHECK` 约束。
- 后端：`backend/internal/constants/risk.go`、`decompression/risk.go`、`dto/decompression_assessment.go`。
- 前端：`frontend/src/types/risk.ts`、`types/assessment.ts`、`stores/assessment.ts`、`pages/AssessmentsPage.tsx`。

状态流转固定为：

```text
draft -> modeled -> pending_supervisor_review -> approved_for_training -> archived
                 \-> draft（退回）
```

模型输入失败不创建评估，并保持或恢复 `draft`；主管批准在单一事务中使用状态和版本条件更新，同时写审计理由。

## API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/v1/auth/login` | 登录并签发 JWT |
| `GET/POST/PUT` | `/api/v1/divers`、`/divers/:id` | 档案列表、创建、更新 |
| `GET` | `/api/v1/divers/:id/plans` | 档案关联方案 |
| `GET/POST` | `/api/v1/plans` | 方案列表与创建 |
| `GET/POST/PUT` | `/api/v1/plans/:id/segments`、`/segments/:id` | 暴露段列表、创建、更新 |
| `PUT` | `/api/v1/plans/:id/segments/order` | 事务化重排并推进输入版本 |
| `POST` | `/api/v1/plans/:id/assessments/run` | 校验并创建不可覆盖评估 |
| `GET` | `/api/v1/assessments`、`/assessments/:id` | 结果列表与重放数据 |
| `GET` | `/api/v1/assessments/:id/compare?other_id=` | 比较两个不可覆盖结果 |
| `POST` | `/api/v1/assessments/:id/submit` | 计划员提交主管复核 |
| `POST` | `/api/v1/assessments/:id/approve` | 主管人工批准训练用途 |
| `GET` | `/api/v1/audit-events` | 主管/管理员读取不可删除审计轨迹 |

统一响应包含 `data` 或 `error` 及 `request_id`。主要错误码包括 `INVALID_GAS_MIX`、`SEGMENT_SEQUENCE_CONFLICT`、`MODEL_INPUT_INVALID`、`PLAN_VERSION_CONFLICT`、`INVALID_PLAN_TRANSITION`、`AUTH_REQUIRED` 和 `FORBIDDEN`。

## 技术栈与目录

- 前端：React 18、TypeScript、Vite、Material UI、Zustand、ECharts、Lucide。
- 后端：Go 1.22、Gin、GORM、validator/v10、JWT、`slog`。
- 数据：Compose 使用 PostgreSQL 16；runtime smoke 使用独立 SQLite 内存库。
- 部署：Nginx 提供 SPA 并保持 `/api/v1` 路径反向代理。

```text
backend/internal/{config,model,dto,repository,service,handler,router,middleware,constants,decompression,util}
frontend/src/{api,stores,types,components/common,hooks,pages,router,utils}
```

后端坚持 `handler -> service -> repository -> model` 单向依赖；压力换算、气体校验、组织舱计算和风险标记分别位于 `decompression/` 独立文件。

## 环境变量

| 变量 | 默认用途 |
| --- | --- |
| `COMPOSE_PROJECT_NAME` | `commercial-diving-decompression-control` |
| `FRONTEND_PORT/BACKEND_PORT/DB_PORT` | `18526/19526/57526` |
| `POSTGRES_DB/POSTGRES_USER/POSTGRES_PASSWORD` | PostgreSQL 业务库配置 |
| `DB_DRIVER/DB_DSN/DB_AUTO_MIGRATE` | 数据库驱动、连接串与迁移开关 |
| `JWT_SECRET/JWT_TTL_MINUTES` | JWT 签名与有效期；生产必须替换密钥 |
| `CORS_ORIGINS` | 允许的前端源 |
| `MODEL_VERSION` | 固化进每次评估和输入快照 |
| `MAX_SEGMENTS` | 单计划最大暴露段数 |
| `RATE_LIMIT_PER_MINUTE/LOG_LEVEL` | 本地限流与结构化日志级别 |

## 本地开发与验证

```bash
go work sync
go build ./backend/...
go vet ./backend/...
go test ./backend/...
npm --prefix frontend ci
npm --prefix frontend run typecheck
npm --prefix frontend run build


docker compose config --quiet
```

若 Compose 启动失败，先检查 `docker compose logs db backend frontend`。若模型返回 `MODEL_INPUT_INVALID`，按 request ID 检查段序号、气体和为 1、深度方向、时长及上升速率；失败不得在前端显示为通过。

## License

MIT
