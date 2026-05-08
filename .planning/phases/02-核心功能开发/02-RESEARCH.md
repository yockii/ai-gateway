# Phase 2: 核心功能开发 - Research

**Researched:** 2026-05-08
**Domain:** AI Gateway Core Implementation
**Confidence:** HIGH

## Summary

Phase 2 implements the core AI Gateway functionality including Bifrost integration, OpenAI-compatible API endpoints, pricing/billing systems, API key management, model routing, and membership tiers. This research documents the technical implementation patterns, library choices, and architectural decisions needed for successful execution.

**Primary recommendation:** Use Bifrost as a library (in-process integration) via `bifrost.Init()` with custom Account interface for business logic injection. Leverage existing patterns from Phase 1 (GORM, Fiber, xid IDs) and extend pricing model based on Bifrost's `TableModelPricing` structure.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| API Request Handling | API Gateway (Fiber) | Business Logic | Gateway handles HTTP parsing, authentication, and OpenAI format conversion |
| Model Routing & Failover | API Gateway (Bifrost) | Business Logic | Bifrost manages provider selection and automatic failover within candidates |
| Pricing Calculation | Business Logic | API Gateway | Business layer owns profit calculation and user-specific pricing |
| Usage Recording | API Gateway | Business Logic | Dual-path recording (async DB + Redis Stream) for reliability |
| API Key Validation | API Gateway | Business Logic | Gateway validates keys synchronously for request authorization |
| Membership Discount | Business Logic | API Gateway | Pricing lookup happens in business layer, gateway applies final price |
| Bill Generation | Business Logic | — | Async batch operation, no gateway involvement |
| PDF/CSV Export | Business Logic | — | One-time file generation, not request-path critical |

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** 采用 **库集成模式** — 将 Bifrost 作为 Go 库直接集成到同一进程
- **D-02:** **混合管理模式** — Bifrost 管理底层供应商路由，业务层管理对外展示、定价和利润
- **D-03:** **全接口兼容** — 实现所有 OpenAI 接口（Chat/Completions/Images/Audio/Embeddings 等）
- **D-04:** **双模式支持** — 同时支持流式（SSE）和非流式响应
- **D-05:** **协同故障转移** — 业务层智能路由 + Bifrost 原生故障转移
- **D-06:** **全面监控** — 基础指标（QPS、延迟、错误率）+ 业务指标（利润率、供应商成本）
- **D-07:** **参考 Bifrost 结构 + 业务层利润管理** — 供应商成本价参考 Bifrost，用户售价由业务层管理
- **D-08:** **双重记录保障** — 同一请求数据通过两条独立路径记录（协程 A: Redis Stream, 协程 B: 异步 DB）
- **D-09:** **数据模型扩展** — SupplierCostPricing 参考 Bifrost 结构支持多层级定价
- **D-10:** **一对多映射** — 一个对外模型映射到多个供应商的实际模型
- **D-11:** **UUID 标准 API Key** — 使用 UUID v4 生成，前缀 `sk-` 开头
- **D-12:** **精细权限控制** — 额度限制、并发限制、时间限制
- **D-13:** **并发控制实现** — 使用 Redis INCR/DECR 记录每个 Key 的当前并发数
- **D-14:** **等级制会员** — 普通/高级/VIP 等级体系
- **D-15:** **差异化定价折扣** — 不同模型不同折扣率
- **D-16:** **即时生效** — 会员升级/降级立即应用
- **D-17:** **自动生成账单** — 每月 1 号自动生成上月账单
- **D-18:** **双格式导出** — 总结性账单 PDF + 明细表 CSV

### Claude's Discretion
None — all decisions locked from context session.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **github.com/maximhq/bifrost/core** | v1.5.8 | Unified AI provider interface | OpenAI-compatible routing, failover, pricing catalog [VERIFIED: go.mod] |
| **github.com/gofiber/fiber/v3** | v3.2.0 | High-performance web framework | Already in use, fasthttp-based for <20ms overhead target [VERIFIED: go.mod] |
| **gorm.io/gorm** | v1.31.1 | ORM for PostgreSQL | Existing database layer, auto-migration pattern established [VERIFIED: go.mod] |
| **github.com/rs/xid** | v1.6.0 | Unique ID generation | Existing pattern for all entity IDs [VERIFIED: go.mod] |
| **github.com/google/uuid** | v1.6.0 | UUID v4 for API keys | Standard UUID library, already available as indirect dependency [VERIFIED: go.mod] |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **github.com/jung-kurt/gofpdf** | latest | PDF generation for billing summaries | When generating formal bills for users |
| **encoding/csv** (stdlib) | — | CSV export for usage details | When exporting billing data for analysis |
| **github.com/redis/go-redis/v9** | v9.0.0 | Redis client for streams & concurrency | For dual-record billing and API key concurrency tracking |
| **github.com/bytedance/sonic** | v1.15.0 | Fast JSON (from Bifrost) | Already pulled in by Bifrost, use for performance |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| gofpdf | signintech/gopdf | gofpdf is more mature, better documentation |
| go-redis | radix.v3 | go-redis has better cluster support and stream handling |
| google/uuid | rs/xid for keys | UUID v4 is industry standard for API keys (OpenAI format) |

**Installation:**
```bash
# Required new dependencies for Phase 2
go get github.com/jung-kurt/gofpdf
go get github.com/redis/go-redis/v9

# Bifrost already included
# google/uuid already available
```

**Version verification:** As of 2026-05-08, all core versions verified from `go.mod`.

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │ User Portal  │  │ Admin Portal │  │  API Clients (OpenAI)   │  │
│  │   (Vue 3)    │  │   (Vue 3)    │  │  SDK/curl/Python/etc.  │  │
│  └──────────────┘  └──────────────┘  └────────────────────────┘  │
└─────────────────────────────┬───────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────┐
│                      API Gateway Layer                          │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Fiber HTTP Server (fasthttp)               │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │  │ Auth Middleware│ │Rate Limit    │  │Concurrency   │   │    │
│  │  │ (API Key)     │  │Middleware    │  │Control       │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘   │    │
│  │  ┌──────────────────────────────────────────────────┐  │    │
│  │  │          Request Handlers (OpenAI-compatible)    │  │    │
│  │  │  /v1/chat/completions  /v1/images/generations     │  │    │
│  │  │  /v1/audio/speech       /v1/embeddings            │  │    │
│  │  └──────────────────────────────────────────────────┘  │    │
│  └──────────────────────────────┬───────────────────────────┘    │
│                                 │                                 │
│  ┌─────────────────────────────▼───────────────────────────┐    │
│  │              Business Logic Layer                        │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │    │
│  │  │ Key Manager  │  │Model Router  │  │Membership    │  │    │
│  │  │              │  │(Profit-based)│  │Manager       │  │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  │    │
│  │  ┌──────────────────────────────────────────────────┐  │    │
│  │  │           Pricing & Billing Manager               │  │    │
│  │  │  • Cost price (from Bifrost)                      │  │    │
│  │  │  • Selling price (user group + membership)       │  │    │
│  │  │  • Profit calculation & protection               │  │    │
│  │  │  • Dual-record billing (DB + Redis Stream)       │  │    │
│  │  └──────────────────────────────────────────────────┘  │    │
│  └──────────────────────────────┬───────────────────────────┘    │
└─────────────────────────────────┼─────────────────────────────────┘
                                  │
┌─────────────────────────────────▼─────────────────────────────────┐
│                      Bifrost Core Layer                            │
│  ┌────────────────────────────────────────────────────────────┐   │
│  │                 Bifrost Instance (In-Process)               │   │
│  │  ┌────────────────────────────────────────────────────┐   │   │
│  │  │          Provider Queue Management                  │   │   │
│  │  │  • Request queuing per provider                     │   │   │
│  │  │  • Concurrent worker pools                          │   │   │
│  │  │  • Failover within candidate list                   │   │   │
│  │  └────────────────────────────────────────────────────┘   │   │
│  │  ┌────────────────────────────────────────────────────┐   │   │
│  │  │              Model Catalog & Pricing                │   │   │
│  │  │  • Provider cost pricing (TableModelPricing)        │   │   │
│  │  │  • Model alias resolution                           │   │   │
│  │  │  • Tiered pricing calculation                       │   │   │
│  │  └────────────────────────────────────────────────────┘   │   │
│  └──────────────────────────────┬───────────────────────────────┘
└─────────────────────────────────┼─────────────────────────────────┘
                                  │
┌─────────────────────────────────▼─────────────────────────────────┐
│                    AI Provider Layer                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐       │
│  │ OpenAI   │  │ Anthropic│  │ Qwen     │  │ Other...     │       │
│  │ Provider │  │ Provider │  │ Provider │  │              │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────────────┘

                              ┌─────────────────┐
                              │  Data Storage   │
                              ├─────────────────┤
                              │  PostgreSQL     │
                              │  • Users        │
                              │  • API Keys     │
                              │  • Models       │
                              │  • Pricing      │
                              │  • Usage Records│
                              │  • Bills        │
                              ├─────────────────┤
                              │  Redis          │
                              │  • Streams      │
                              │  • Concurrency  │
                              └─────────────────┘
```

### Recommended Project Structure

```
internal/
├── services/
│   ├── bifrost.go          # Bifrost integration wrapper
│   ├── key_manager.go      # API Key management service
│   ├── model_router.go     # Smart routing based on profit
│   ├── membership.go       # Membership tier management
│   └── billing.go          # Bill generation & export
├── models/
│   ├── keys.go             # UserAPIKey, KeyPermission models
│   ├── membership.go       # MembershipTier, MembershipDiscount models
│   └── model_mapping.go    # ModelMapping, SupplierRoute models
├── middleware/
│   ├── concurrency.go      # Concurrent request limit middleware
│   └── validation.go       # Request validation middleware
├── utils/
│   ├── pdf.go              # PDF generation for bills
│   ├── csv.go              # CSV export for usage details
│   └── retry.go            # Retry logic with exponential backoff
└── streaming/
    └── sse.go              # Server-Sent Events streaming utilities

pkg/
├── handlers/
│   ├── images.go           # Image generation handler
│   ├── audio.go            # TTS/STT handlers
│   ├── embeddings.go       # Embeddings handler
│   └── admin.go            # Admin API handlers
└── api/
    └── types.go            # Extend with new request/response types
```

### Pattern 1: Bifrost Library Integration

**What:** Integrate Bifrost as an in-process library using `bifrost.Init()` with custom Account interface

**When to use:** All AI model requests need routing through Bifrost

**Example:**
```go
// Source: D:/projects/github.com/maximhq/bifrost/core/bifrost.go:211-299
package bifrostservice

import (
    "context"
    bifrost "github.com/maximhq/bifrost/core"
    "github.com/maximhq/bifrost/core/schemas"
    "github.com/yockii/ai-gateway/internal/config"
)

// GatewayAccount implements Bifrost Account interface for business logic injection
type GatewayAccount struct {
    // Business layer fields
    pricingManager *PricingManager
    keyManager     *KeyManager
}

func NewBifrostService(cfg *config.Config, pricingMgr *PricingManager) (*BifrostService, error) {
    // Initialize Bifrost with custom configuration
    bifrostConfig := schemas.BifrostConfig{
        Account: &GatewayAccount{pricingManager: pricingMgr},
        Logger:  NewDefaultLogger(schemas.LogLevelInfo),
        DropExcessRequests: true,  // Don't queue when full
    }

    // Initialize Bifrost instance
    b, err := bifrost.Init(context.Background(), bifrostConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Bifrost: %w", err)
    }

    // Add providers dynamically
    // Providers managed by business layer based on Supplier model

    return &BifrostService{bifrost: b}, nil
}

// ChatCompletion forwards requests to Bifrost
func (s *BifrostService) ChatCompletion(ctx context.Context, req *schemas.BifrostChatRequest) (*schemas.BifrostChatResponse, *schemas.BifrostError) {
    return s.bifrost.ChatCompletionRequest(ctx, req)
}

// ChatCompletionStream for streaming responses
func (s *BifrostService) ChatCompletionStream(ctx context.Context, req *schemas.BifrostChatRequest) (<-chan *schemas.BifrostStreamChunk, *schemas.BifrostError) {
    return s.bifrost.ChatCompletionStreamRequest(ctx, req)
}
```

### Pattern 2: Profit-Based Model Routing

**What:** Select suppliers based on profit margin (selling price - cost price)

**When to use:** Every model request needs to select optimal supplier

**Example:**
```go
// Source: D:/projects/github.com/yockii/ai_gateway/internal/cost/optimizer.go (conceptual)
type ModelRouter struct {
    db              *database.DB
    pricingManager  *PricingManager
    bifrostService  *BifrostService
}

type SupplierSelection struct {
    SupplierID      string
    ModelName       string
    CostPrice       float64
    SellingPrice    float64
    Profit          float64
    ProfitMargin    float64
}

func (r *ModelRouter) SelectSupplier(ctx context.Context, userID, externalModelID string) (*SupplierSelection, error) {
    // 1. Get user's selling price
    userGroupID, err := r.getUserGroupID(ctx, userID)
    sellingPrice := r.pricingManager.GetSellingPrice(userGroupID, externalModelID)

    // 2. Get all configured suppliers for this external model
    routes, err := r.getModelRoutes(ctx, externalModelID)
    if err != nil {
        return nil, err
    }

    // 3. Calculate profit for each supplier
    var candidates []*SupplierSelection
    for _, route := range routes {
        costPrice := r.pricingManager.GetCostPrice(route.SupplierID, route.ActualModelName)

        // Filter: selling price must be > cost price (profit protection)
        if sellingPrice.InputPrice <= costPrice.InputCost {
            continue
        }

        profit := sellingPrice.InputPrice - costPrice.InputCost
        profitMargin := profit / sellingPrice.InputPrice

        candidates = append(candidates, &SupplierSelection{
            SupplierID:   route.SupplierID,
            ModelName:    route.ActualModelName,
            CostPrice:    costPrice.InputCost,
            SellingPrice: sellingPrice.InputPrice,
            Profit:       profit,
            ProfitMargin: profitMargin,
        })
    }

    // 4. Sort by profit (highest first), then by priority
    sort.Slice(candidates, func(i, j int) bool {
        if candidates[i].Profit != candidates[j].Profit {
            return candidates[i].Profit > candidates[j].Profit
        }
        // TODO: add priority comparison
        return true
    })

    // 5. Return best candidate
    // Bifrost will handle failover within this candidate list
    if len(candidates) == 0 {
        return nil, fmt.Errorf("no available supplier with positive margin")
    }

    return candidates[0], nil
}
```

### Pattern 3: Dual-Record Billing with Redis Streams

**What:** Record usage data via both async DB write and Redis Stream for reliability

**When to use:** Every successful model response needs usage recording

**Example:**
```go
// Source: Conceptual pattern for CONTEXT.md D-08 implementation
type BillingRecorder struct {
    db         *database.DB
    redis      *redis.Client
    streamName string
}

func (r *BillingRecorder) RecordUsage(ctx context.Context, usage *UsageRecord) error {
    // RequestID for idempotency (stored in UsageRecord.RequestID as unique index)
    requestID := usage.RequestID
    if requestID == "" {
        requestID = xid.New().String()
        usage.RequestID = requestID
    }

    // Goroutine A: Push to Redis Stream (sync, returns immediately)
    streamData := map[string]interface{}{
        "request_id":     requestID,
        "user_id":        usage.UserID,
        "model_id":       usage.ModelID,
        "supplier_id":    usage.SupplierID,
        "input_tokens":   usage.InputTokens,
        "output_tokens":  usage.OutputTokens,
        "cost_price":     usage.CostPrice,
        "selling_price":  usage.SellingPrice,
        "profit":         usage.Profit,
        "created_at":     time.Now().UTC().Format(time.RFC3339),
    }

    if err := r.redis.XAdd(ctx, &redis.XAddArgs{
        Stream: r.streamName,
        ID:     "*",  // Auto-generated ID
        Values: streamData,
    }).Err(); err != nil {
        log.Printf("WARNING: Redis Stream write failed: %v", err)
        // Continue anyway - async DB write is backup
    }

    // Goroutine B: Async direct DB write
    go func() {
        dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := r.db.WithContext(dbCtx).Create(usage).Error; err != nil {
            log.Printf("ERROR: Direct DB write failed: %v", err)
            // Redis Stream consumer will handle retry
        }
    }()

    return nil
}

// Consumer (runs in separate process/goroutine)
func (r *BillingRecorder) StreamConsumer(ctx context.Context) {
    for {
        // Read from stream
        results, err := r.redis.XRead(ctx, &redis.XReadArgs{
            Streams: []string{r.streamName, "$"},
            Count:   100,
            Block:   5 * time.Second,
        }).Result()

        if err != nil {
            log.Printf("Stream read error: %v", err)
            continue
        }

        for _, stream := range results {
            for _, message := range stream.Messages {
                r.processMessage(ctx, message)
                r.redis.XAck(ctx, r.streamName, "consumer-group", message.ID)
            }
        }
    }
}

func (r *BillingRecorder) processMessage(ctx context.Context, message redis.XMessage) {
    // Parse and write to DB with idempotency check
    var usage UsageRecord
    // ... parse message.Values into usage ...

    // Idempotent insert (ignore duplicate key error)
    r.db.WithContext(ctx).Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "request_id"}},
        DoNothing: true,
    }).Create(&usage)
}
```

### Pattern 4: API Key Generation with UUID v4

**What:** Generate secure API keys using UUID v4 with `sk-` prefix (OpenAI standard)

**When to use:** Creating new API keys for users

**Example:**
```go
// Source: Based on CONTEXT.md D-11 requirement
package keymanager

import (
    "fmt"
    "github.com/google/uuid"
)

type KeyManager struct {
    db *database.DB
}

func (km *KeyManager) GenerateAPIKey(ctx context.Context, userID, name string) (*UserAPIKey, string, error) {
    // Generate UUID v4
    keyUUID, err := uuid.NewRandom()
    if err != nil {
        return nil, "", fmt.Errorf("failed to generate UUID: %w", err)
    }

    // Format: sk-<uuid>
    apiKeyValue := fmt.Sprintf("sk-%s", keyUUID.String())

    // Create key record
    apiKey := &UserAPIKey{
        ID:           xid.New().String(),
        UserID:       userID,
        KeyValue:     apiKeyValue,  // Store hashed in production
        Name:         name,
        IsActive:     true,
        CreatedAt:    time.Now().UTC(),
    }

    if err := km.db.WithContext(ctx).Create(apiKey).Error; err != nil {
        return nil, "", err
    }

    // Return raw key (only time it's visible)
    return apiKey, apiKeyValue, nil
}
```

### Pattern 5: Concurrency Control with Redis INCR/DECR

**What:** Track concurrent requests per API key using Redis counters

**When to use:** Enforce concurrent request limits before processing

**Example:**
```go
// Source: Based on CONTEXT.md D-13 requirement
package middleware

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

type ConcurrencyLimiter struct {
    redis *redis.Client
}

func (cl *ConcurrencyLimiter) AcquireSlot(ctx context.Context, apiKey string, limit int64) error {
    key := fmt.Sprintf("concurrency:%s", apiKey)

    // Increment counter
    current, err := cl.redis.Incr(ctx, key).Result()
    if err != nil {
        return fmt.Errorf("failed to check concurrency: %w", err)
    }

    // Check if exceeded limit
    if current > limit {
        // Rollback increment
        cl.redis.Decr(ctx, key)
        return fmt.Errorf("concurrency limit exceeded: %d/%d", current, limit)
    }

    // Set expiry (in case request crashes without ReleaseSlot)
    cl.redis.Expire(ctx, key, 5*time.Minute)

    return nil
}

func (cl *ConcurrencyLimiter) ReleaseSlot(ctx context.Context, apiKey string) {
    key := fmt.Sprintf("concurrency:%s", apiKey)
    cl.redis.Decr(ctx, key)
}

// Middleware usage
func Concurrency(cl *ConcurrencyLimiter) fiber.Handler {
    return func(c *fiber.Ctx) error {
        apiKey := middleware.GetAPIKey(c)
        keyRecord := getKeyRecord(c)  // Get from context

        limit := keyRecord.ConcurrencyLimit
        if limit == 0 {
            limit = 10  // Default
        }

        // Check model-specific limit
        if modelLimit := keyRecord.ModelConcurrency[c.Query("model")]; modelLimit > 0 {
            limit = modelLimit
        }

        if err := cl.AcquireSlot(c.Context(), apiKey, limit); err != nil {
            return c.Status(429).JSON(api.ErrorResponse{
                Error: api.ErrorDetail{
                    Message: err.Error(),
                    Type:    "concurrency_limit_exceeded",
                    Code:    429,
                },
            })
        }

        defer cl.ReleaseSlot(c.Context(), apiKey)

        return c.Next()
    }
}
```

### Anti-Patterns to Avoid

- **Direct HTTP calls to providers**: Always use Bifrost for provider communication — it handles failover, queuing, and format conversion
- **Synchronous billing writes**: Never block response for billing — use dual-path async recording
- **Storing API keys in plain text**: Always hash keys before storing (use bcrypt/scrypt)
- **Hardcoded model names**: Use ExternalModel mapping for flexibility
- **Ignoring profit margins**: Never route to suppliers where cost > selling price
- **Blocking Redis operations**: All Redis calls should have timeouts and fallback behavior

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| AI provider routing | Custom HTTP clients for OpenAI/Anthropic/etc | Bifrost library | Handles failover, queuing, format conversion, 30+ providers |
| PDF generation | Custom PDF rendering from scratch | gofpdf library | Complex format specification, edge cases with fonts/encoding |
| UUID generation | Custom random string generators | google/uuid | Crypto-secure, standard format, collision-resistant |
| Token counting | Custom tokenization logic | Bifrost's built-in token counting | Provider-specific tokenization rules |
| Rate limiting | In-memory only with mutexes | Redis-backed with go-redis/v9 | Distributed systems need shared state |

**Key insight:** Bifrost already solves the hardest problem (multi-provider routing with failover). Leverage it rather than rebuilding.

## Runtime State Inventory

> Not a rename/refactor phase — this section omitted per research protocol.

## Common Pitfalls

### Pitfall 1: Ignoring Bifrost's Provider Queues
**What goes wrong:** Sending requests faster than providers can process causes memory bloat and timeouts
**Why it happens:** Bifrost queues requests per provider, but unbounded queues grow without limit
**How to avoid:** Set `DropExcessRequests: true` in BifrostConfig and monitor queue depths
**Warning signs:** Memory usage growing linearly with request rate, increasing latency

### Pitfall 2: Streaming Response Connection Leaks
**What goes wrong:** Streaming goroutines never exit when clients disconnect
**Why it happens:** Not checking context cancellation in streaming loops
**How to avoid:** Always select on `<-ctx.Done()` in streaming loops
```go
for chunk := range streamChan {
    select {
    case <-ctx.Done():
        return  // Client disconnected
    default:
        // Process chunk
    }
}
```
**Warning signs:** Increasing goroutine count, connection limit errors

### Pitfall 3: Race Conditions in Concurrency Control
**What goes wrong:** Concurrent requests exceed limits due to INCR/DECR timing gaps
**Why it happens:** Check-and-increment is not atomic without proper locking
**How to avoid:** Use INCR + check limit + DECR on overflow pattern (shown in Pattern 5)
**Warning signs:** Limits not enforced, sporadic 429s when they shouldn't occur

### Pitfall 4: Incorrect Profit Margin Calculation
**What goes wrong:** Selling at loss when cost price changes but selling price doesn't
**Why it happens:** Not validating `sellingPrice > costPrice` on every request
**How to avoid:** Always fetch current prices and validate margin > 0 before routing
**Warning signs:** Negative profit in usage records, billing discrepancies

### Pitfall 5: PDF Generation Encoding Issues
**What goes wrong:** Chinese characters or emojis garbled in PDF bills
**Why it happens:** gofpdf default fonts don't support UTF-8 beyond ASCII
**How to avoid:** Bundle and use proper UTF-8 fonts (e.g., Noto Sans CJK)
**Warning signs:** Question marks or boxes in generated PDFs

## Code Examples

### Bifrost Initialization with Custom Providers

```go
// Source: D:/projects/github.com/maximhq/bifrost/core/bifrost.go:211-299
func InitBifrostWithProviders(ctx context.Context, suppliers []Supplier) (*bifrost.Bifrost, error) {
    bifrostConfig := schemas.BifrostConfig{
        Account:              &GatewayAccount{},
        Logger:               NewDefaultLogger(schemas.LogLevelInfo),
        DropExcessRequests:   true,
        LLMPlugins:           []schemas.LLMPlugin{},
        MCPPlugins:           []schemas.MCPPlugin{},
    }

    b, err := bifrost.Init(ctx, bifrostConfig)
    if err != nil {
        return nil, err
    }

    // Add providers from database
    for _, supplier := range suppliers {
        providerConfig := schemas.Provider{
            ProviderID:        supplier.ID,
            ProviderType:      supplier.Provider,  // "openai", "anthropic", etc.
            APIKeys:           getAPIKeysForSupplier(supplier.ID),
            BaseURL:           supplier.BaseURL,
            MaxRetries:        3,
            Timeout:           30 * time.Second,
            Priority:          supplier.Priority,
            MaxQueueSize:      supplier.MaxQPS,
        }

        if err := b.AddProvider(ctx, providerConfig); err != nil {
            log.Printf("WARNING: Failed to add provider %s: %v", supplier.ID, err)
        }
    }

    return b, nil
}
```

### SSE Streaming Response

```go
// Source: Based on Bifrost streaming pattern + Fiber SSE
func (h *Handler) ChatCompletionsStream(c *fiber.Ctx) error {
    userID := middleware.GetUserID(c)
    var req api.ChatCompletionRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(api.ErrorResponse{Error: api.ErrorDetail{
            Message: "Invalid request body",
            Type:    "invalid_request_error",
            Code:    400,
        }})
    }

    // Set SSE headers
    c.Set("Content-Type", "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection", "keep-alive")

    // Convert to Bifrost request
    bifrostReq := h.convertToBifrostRequest(&req, userID)

    // Get streaming channel from Bifrost
    streamChan, err := h.bifrost.ChatCompletionStream(c.Context(), bifrostReq)
    if err != nil {
        return c.Status(500).JSON(api.ErrorResponse{Error: api.ErrorDetail{
            Message: "Failed to start stream",
            Type:    "api_error",
            Code:    500,
        }})
    }

    // Stream response
    for chunk := range streamChan {
        // Check if client disconnected
        if c.Context().Err() != nil {
            return nil
        }

        // Convert to OpenAI SSE format
        sseData := h.formatSSE(chunk)

        // Write to response
        if _, err := fmt.Fprintf(c, "data: %s\n\n", sseData); err != nil {
            return err
        }
    }

    // Send final [DONE] message
    fmt.Fprint(c, "data: [DONE]\n\n")

    return nil
}
```

### PDF Bill Generation

```go
// Source: Conceptual implementation for CONTEXT.md D-18
package utils

import (
    "github.com/jung-kurt/gofpdf"
    "time"
)

type BillGenerator struct {
    pdf *gofpdf.Fpdf
}

func (bg *BillGenerator) GeneratePDF(bill *Bill) ([]byte, error) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Arial", "", 12)

    // Header
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(0, 10, "AI Gateway - Monthly Bill")
    pdf.Ln(12)

    // Bill info
    pdf.SetFont("Arial", "", 12)
    pdf.Printf("Bill ID: %s\n", bill.ID)
    pdf.Printf("Period: %s to %s\n",
        bill.StartDate.Format("2006-01-02"),
        bill.EndDate.Format("2006-01-02"))
    pdf.Printf("Generated: %s\n", time.Now().Format("2006-01-02"))
    pdf.Ln(10)

    // Summary table
    pdf.SetFont("Arial", "B", 12)
    pdf.Cell(60, 7, "Model")
    pdf.Cell(30, 7, "Requests")
    pdf.Cell(30, 7, "Tokens")
    pdf.Cell(30, 7, "Cost")
    pdf.Ln(7)

    pdf.SetFont("Arial", "", 10)
    for _, item := range bill.Items {
        pdf.Cell(60, 6, item.ModelID)
        pdf.Cell(30, 6, fmt.Sprintf("%d", item.TotalRequests))
        pdf.Cell(30, 6, fmt.Sprintf("%d", item.TotalTokens))
        pdf.Cell(30, 6, fmt.Sprintf("$%.4f", item.TotalCost))
        pdf.Ln(6)
    }

    // Total
    pdf.Ln(5)
    pdf.SetFont("Arial", "B", 12)
    pdf.Printf("Total Cost: $%.4f\n", bill.TotalCost)
    pdf.Printf("Total Revenue: $%.4f\n", bill.TotalRevenue)
    pdf.Printf("Total Profit: $%.4f\n", bill.TotalProfit)

    // Write to bytes
    var buf bytes.Buffer
    err := pdf.Output(&buf)
    return buf.Bytes(), err
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Direct provider HTTP calls | Unified provider abstraction (Bifrost) | 2023-2024 | Eliminates 1000s of lines of provider-specific code |
| Fixed pricing per model | Tiered pricing (128k/200k/272k) + cache pricing | 2024 | Accurate billing for modern context sizes |
| Single supplier per model | Multi-supplier with intelligent routing | 2023-2024 | Cost optimization + automatic failover |
| In-memory rate limiting | Redis-backed distributed limiting | 2022-2023 | Horizontal scaling support |
| Manual failover logic | Built-in failover with queue management | 2023 | Improved reliability, reduced downtime |

**Deprecated/outdated:**
- Direct OpenAI SDK usage (migrate to Bifrost)
- Custom token counting (use Bifrost's catalog)
- Synchronous billing writes (use dual-path async)
- Per-process rate limiting (use Redis for distributed)

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Redis is available for streams and concurrency | Standard Stack | Fallback to in-memory for concurrency, accept billing delay risk |
| A2 | gofpdf supports required UTF-8 characters for Chinese bills | Standard Stack | PDF garbling — need to bundle proper fonts |
| A3 | Bifrost v1.5.8 API is stable for library mode | Standard Stack | API breaks would require adaptation layer |
| A4 | PostgreSQL performance adequate for 10000+ QPS billing writes | Architecture | May need to offload to time-series DB at scale |
| A5 | 20ms gateway overhead achievable with Fiber + Bifrost | Performance | May need optimization or compiled middleware |

## Open Questions

1. **Redis High Availability**
   - What we know: Redis used for concurrency and billing streams
   - What's unclear: Is Redis Sentinel/Cluster configured for HA?
   - Recommendation: Assume single instance for Phase 2, document HA requirements for Phase 3

2. **PDF Font Bundling**
   - What we know: gofpdf needs UTF-8 font support for Chinese characters
   - What's unclear: License/size constraints for bundling Noto Sans CJK
   - Recommendation: Use system fonts or download-on-demand for Phase 2

3. **Bifrost Provider Credential Management**
   - What we know: Suppliers have API keys stored in database
   - What's unclear: Encryption/rotation strategy for supplier credentials
   - Recommendation: Store encrypted in Supplier model, decrypt at AddProvider time

4. **Stream Consumer Deployment**
   - What we know: Redis Stream consumer needed for billing backup
   - What's unclear: Run as separate service or goroutine within gateway?
   - Recommendation: Start with goroutine, move to separate service if performance issues

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| PostgreSQL (llm_gateway) | Data layer | ✓ | 15.x | — |
| Bifrost (library) | AI routing | ✓ | v1.5.8 | — |
| Redis | Concurrency, billing | ? | — | In-memory (limited) |
| gofpdf | Bill export | ✗ | — | Skip PDF, CSV only |
| go-redis/v9 | Redis client | ✗ | — | Standard library only |

**Missing dependencies with no fallback:**
- None — core functionality doesn't require Redis (degraded mode acceptable)

**Missing dependencies with fallback:**
- Redis: Use in-memory concurrency limiting (not distributed), accept billing latency risk
- gofpdf: Generate CSV-only bills, PDF can be added later

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify (stdlib) |
| Config file | None — using table-driven tests |
| Quick run command | `go test -v -run TestSpecific ./internal/services/...` |
| Full suite command | `go test -v ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FR-001 | OpenAI-compatible API responses | integration | `go test -v -run TestChatCompletion ./pkg/handlers/...` | ✅ Wave 0 (chat.go exists) |
| FR-002 | Multi-supplier model mapping | unit | `go test -v -run TestModelRouter ./internal/services/...` | ❌ Wave 0 |
| FR-003 | Tiered pricing calculation | unit | `go test -v -run TestPricing ./internal/pricing/...` | ✅ Wave 0 (manager_test.go exists) |
| FR-007 | API Key CRUD operations | unit | `go test -v -run TestKeyManager ./internal/services/...` | ❌ Wave 0 |
| FR-008 | Dual-record billing | integration | `go test -v -run TestBilling ./internal/services/...` | ✅ Wave 0 (manager_test.go exists) |
| FR-009 | Profit-based routing | unit | `go test -v -run TestProfitRouting ./internal/services/...` | ❌ Wave 0 |
| FR-010 | Membership discount application | unit | `go test -v -run TestMembership ./internal/services/...` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test -v -run <specific_test> <package>`
- **Per wave merge:** `go test -v ./...`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- `internal/services/keymanager/keymanager_test.go` — API Key CRUD tests
- `internal/services/modelrouter/modelrouter_test.go` — Routing logic tests
- `internal/services/membership/membership_test.go` — Membership tests
- `internal/middleware/concurrency_test.go` — Concurrency limit tests
- `pkg/handlers/images_test.go` — Image generation handler tests
- `pkg/handlers/audio_test.go` — TTS/STT handler tests
- Framework install: Already available (stdlib + testify)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes | Bcrypt-hashed API keys, secure UUID v4 generation |
| V3 Session Management | no | Stateless API gateway, no sessions |
| V4 Access Control | yes | API key scope validation, user group permissions |
| V5 Input Validation | yes | Request validation against OpenAI schemas, GORM input sanitization |
| V6 Cryptography | yes | UUID v4 (crypto-secure), bcrypt for key hashing |

### Known Threat Patterns for AI Gateway

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| API Key leakage | Tampering/Disclosure | Hash keys in DB, never log raw keys, HTTPS only |
| Cost manipulation | Tampering | Server-side pricing calculation, never trust client input |
| Resource exhaustion | Denial of Service | Rate limiting, concurrency limits, request size limits |
| Invoice fraud | Spoofing | Idempotent billing via RequestID unique index |
| Privilege escalation | Elevation | Validate user group membership on each request |
| Supplier credential exposure | Disclosure | Encrypt supplier API keys at rest, rotate regularly |

## Sources

### Primary (HIGH confidence)
- [D:/projects/github.com/maximhq/bifrost/core/bifrost.go](file:///D:/projects/github.com/maximhq/bifrost/core/bifrost.go) - Bifrost initialization and API
- [D:/projects/github.com/maximhq/bifrost/framework/modelcatalog/pricing.go](file:///D:/projects/github.com/maximhq/bifrost/framework/modelcatalog/pricing.go) - Pricing structure and calculation
- [D:/projects/github.com/maximhq/bifrost/framework/configstore/tables/modelpricing.go](file:///D:/projects/github.com/maximhq/bifrost/framework/configstore/tables/modelpricing.go) - TableModelPricing schema
- [go.mod](file:///D:/projects/github.com/yockii/ai_gateway/go.mod) - Verified dependency versions

### Secondary (MEDIUM confidence)
- [internal/pricing/manager.go](file:///D:/projects/github.com/yockii/ai_gateway/internal/pricing/manager.go) - Existing pricing patterns
- [pkg/handlers/chat.go](file:///D:/projects/github.com/yockii/ai_gateway/pkg/handlers/chat.go) - Handler pattern reference
- [internal/middleware/auth.go](file:///D:/projects/github.com/yockii/ai_gateway/internal/middleware/auth.go) - Authentication pattern
- [internal/middleware/ratelimit.go](file:///D:/projects/github.com/yockii/ai_gateway/internal/middleware/ratelimit.go) - Rate limiting pattern

### Tertiary (LOW confidence)
- Web search (rate-limited) — General Go library recommendations, not verified

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All versions verified from go.mod
- Architecture: HIGH - Bifrost API verified from source, patterns match existing codebase
- Pitfalls: HIGH - Based on verified Bifrost behavior and common Go concurrency issues
- PDF generation: MEDIUM - gofpdf choice unverified (web search rate-limited), but well-established library

**Research date:** 2026-05-08
**Valid until:** 2026-06-08 (30 days for stable dependencies)

---

**Phase:** 2-核心功能开发
**Research completed:** 2026-05-08
**Next action:** Create VALIDATION.md with testable acceptance criteria
