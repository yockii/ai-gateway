# Phase 2: 核心功能开发 - Context

**Gathered:** 2026-05-08
**Status:** Ready for planning

## Phase Boundary

阶段 2 负责实现核心业务功能，包括：
- **API 网关服务**：Bifrost 集成、OpenAI 兼容接口、协议转换、负载均衡
- **用户 API Key 管理**：Key 生成验证、权限控制、使用统计
- **对外模型管理**：模型配置、一对多供应商映射、智能路由
- **价格与计费系统**：Bifrost 价格体系集成、双重记录保障、实时计费
- **会员套餐体系**：等级制会员、差异化定价折扣
- **账单与对账**：自动账单生成、PDF 总结 + CSV 明细导出

## Implementation Decisions

### Bifrost 集成策略
- **D-01:** 采用 **库集成模式** — 将 Bifrost 作为 Go 库直接集成到同一进程，实现最优性能
- **D-02:** **混合管理模式** — Bifrost 管理底层供应商路由和调用，业务层管理对外展示、定价和利润计算
  - 业务层：ExternalModel（对外名称）、UserGroupPricing（用户售价）、SupplierCostPricing（供应商成本）
  - Bifrost 层：Provider（实际供应商）、Model（实际模型调用）
  - 适配层：将两者映射连接
- **D-03:** **全接口兼容** — 实现所有 OpenAI 接口（Chat/Completions/Images/Audio/Embeddings 等），满足 FR-001 需求
- **D-04:** **双模式支持** — 同时支持流式（SSE）和非流式响应，通过 `stream` 参数控制
- **D-05:** **协同故障转移** — 业务层智能路由（按利润排序候选供应商）+ Bifrost 原生故障转移（在候选列表内自动切换）
- **D-06:** **全面监控** — 基础指标（QPS、延迟、错误率）+ 业务指标（利润率、供应商成本）完整监控

### 价格与计费系统
- **D-07:** **参考 Bifrost 结构 + 业务层利润管理** — 供应商成本价参考 Bifrost 定价结构扩展，用户售价由业务层独立管理
- **D-08:** **双重记录保障** — 同一请求数据通过两条独立路径记录：
  1. 协程 B：异步直接写入数据库
  2. 协程 A：同步推送 Redis Stream
  3. 协程 C：消费者从 Redis Stream 读取并写入数据库
  4. 幂等性：RequestID 作为唯一索引，自动去重
- **D-09:** **数据模型扩展** — SupplierCostPricing 参考 Bifrost 结构支持多层级定价、缓存定价等完整模式

### 模型与 Key 管理
- **D-10:** **一对多映射** — 一个对外模型映射到多个供应商的实际模型，支持成本优先路由
- **D-11:** **UUID 标准 API Key** — 使用 UUID v4 生成，前缀 `sk-` 开头，符合 OpenAI 标准
- **D-12:** **精细权限控制**：
  - 额度限制：每日/每月额度
  - 并发限制：默认并发 + 模型特定并发（如 `{"gpt-4": 5, "gpt-3.5-turbo": 10}`）
  - 时间限制：过期时间
- **D-13:** **并发控制实现** — 使用 Redis 记录每个 Key 的当前并发数，请求前 INCR，响应后 DECR

### 会员套餐体系
- **D-14:** **等级制会员** — 普通/高级/VIP 等级体系，每个等级固定权益
- **D-15:** **差异化定价折扣** — 不同模型不同折扣率，精细化管理
- **D-16:** **即时生效** — 会员升级/降级立即应用

### 账单与对账系统
- **D-17:** **自动生成账单** — 每月 1 号自动生成上月账单
- **D-18:** **双格式导出** — 总结性账单 PDF（正式、可打印）+ 明细表 CSV（便于数据分析）

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 需求文档
- `.planning/REQUIREMENTS.md` — 完整的功能需求规格，包括所有 10 个核心功能模块
- `.planning/ROADMAP.md` — 开发路线图，阶段 2 的详细任务分解

### 现有代码模型
- `internal/models/models.go` — 数据模型定义（User, Admin, Supplier, ExternalModel, UsageRecord 等）
- `internal/database/database.go` — GORM 配置和自动迁移逻辑
- `internal/pricing/manager.go` — 定价管理器基础实现
- `pkg/handlers/chat.go` — Chat Completions 处理器实现
- `pkg/router/router.go` — 基础路由配置

### Bifrost 参考
- `D:/projects/github.com/maximhq/bifrost/` — Bifrost 项目路径，需要研究其价格体系和模型路由实现

## Existing Code Insights

### Reusable Assets
- **数据库层**：`internal/database/database.go` 已配置 GORM + PostgreSQL，支持自动迁移和连接池
- **定价管理器**：`internal/pricing/manager.go` 已实现 RecordUsage、GenerateBill、GetUserUsageSummary 基础功能
- **API 处理器**：`pkg/handlers/chat.go` 已实现 Chat Completions 处理逻辑和请求验证
- **路由系统**：`pkg/router/router.go` 已配置基础 /v1 路由和中间件

### Established Patterns
- **xid.New().String()** — 用于生成唯一 ID
- **GORM 自动迁移** — 禁用外键约束，应用层维护关系
- **Fiber 框架** — 基于 fasthttp 的高性能 Web 框架
- **中间件模式** — Recovery、Logger、Auth、RateLimit 中间件链

### Integration Points
- **Gateway 初始化**：`internal/gateway/gateway.go` 中的 Bifrost 集成（TODO - Wave 2）
- **API 路由扩展**：`pkg/router/router.go` 需要添加新的模型接口端点
- **数据库模型扩展**：需要在 `internal/models/models.go` 中添加会员套餐相关模型

## Specific Ideas

### Bifrost 价格体系参考
需要研究 Bifrost 的以下定价结构：
- TextModelPricing：多层级定价、缓存定价
- ImageModelPricing：按图片/像素/分辨率/质量定价
- VideoAudioPricing：按秒/按 Token 定价

### API Key 格式
```
sk-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

### RequestID 幂等性设计
```go
type UsageRecord struct {
    RequestID string `json:"request_id" gorm:"uniqueIndex"` // 幂等键
    // ...
}
```

### 并发限制配置示例
```go
type UserAPIKey struct {
    ConcurrencyLimit int64            `json:"concurrency_limit"`    // 默认并发
    ModelConcurrency map[string]int64 `json:"model_concurrency"`    // 模型特定并发
    // 例如: {"gpt-4": 5, "gpt-3.5-turbo": 10}
}
```

## Deferred Ideas

None — discussion stayed within phase scope.

---

*Phase: 2-核心功能开发*
*Context gathered: 2026-05-08*
