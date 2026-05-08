# Phase 2: Core Functionality - Pattern Map

**Mapped:** 2026-05-08
**Files analyzed:** 15 new/modified files
**Analogs found:** 12 / 15

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/models/keys.go` | model | CRUD | `internal/models/models.go` | exact |
| `internal/models/membership.go` | model | CRUD | `internal/models/models.go` | exact |
| `internal/models/model_mapping.go` | model | CRUD | `internal/models/models.go` | exact |
| `internal/services/bifrost.go` | service | request-response | `internal/gateway/gateway.go` | role-match |
| `internal/services/key_manager.go` | service | CRUD | `internal/pricing/manager.go` | role-match |
| `internal/services/model_router.go` | service | event-driven | `internal/cost/optimizer.go` | role-match |
| `internal/services/membership.go` | service | CRUD | `internal/pricing/manager.go` | role-match |
| `internal/services/billing.go` | service | batch | `internal/pricing/manager.go` | role-match |
| `pkg/handlers/images.go` | handler | request-response | `pkg/handlers/chat.go` | exact |
| `pkg/handlers/audio.go` | handler | streaming | `pkg/handlers/chat.go` | role-match |
| `pkg/handlers/embeddings.go` | handler | request-response | `pkg/handlers/chat.go` | exact |
| `pkg/handlers/admin.go` | handler | request-response | `pkg/handlers/chat.go` | role-match |
| `internal/middleware/concurrency.go` | middleware | request-response | `internal/middleware/ratelimit.go` | exact |
| `internal/middleware/validation.go` | middleware | request-response | `internal/middleware/auth.go` | role-match |
| `internal/utils/retry.go` | utility | transform | `internal/cost/optimizer.go` | partial |

## Pattern Assignments

### `internal/models/keys.go` (model, CRUD)

**Analog:** `internal/models/models.go` (lines 87-98)

**Model pattern:**
```go
// UserAPIKey 用户 API Key
type UserAPIKey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index"`
	KeyValue    string    `json:"-" gorm:"uniqueIndex"` // 不在 JSON 中显示
	Name        string    `json:"name"`
	IsActive    bool      `json:"is_active" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

**Key patterns:**
- Use `json:"-"` for sensitive fields (passwords, API keys)
- Use `gorm:"index"` for frequently queried fields
- Use `gorm:"uniqueIndex"` for unique constraints
- Use `xid.New().String()` for ID generation

---

### `internal/models/membership.go` (model, CRUD)

**Analog:** `internal/models/models.go` (lines 52-74)

**Model pattern:**
```go
// MembershipTier 会员等级
type MembershipTier struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	DisplayName string    `json:"display_name"`
	IsActive    bool      `json:"is_active" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MembershipDiscount 会员折扣配置
type MembershipDiscount struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	MembershipTierID string    `json:"membership_tier_id" gorm:"index"`
	ModelID          string    `json:"model_id" gorm:"index"`
	DiscountRate     float64   `json:"discount_rate"` // 折扣率 0-1
	IsActive         bool      `json:"is_active" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
```

---

### `internal/models/model_mapping.go` (model, CRUD)

**Analog:** `internal/models/models.go` (lines 76-85)

**Model pattern:**
```go
// ModelMapping 模型映射配置
type ModelMapping struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ExternalModelID string   `json:"external_model_id" gorm:"index"`
	SupplierID     string    `json:"supplier_id" gorm:"index"`
	ActualModelName string   `json:"actual_model_name"`
	Priority       int       `json:"priority"` // 路由优先级
	IsActive       bool      `json:"is_active" gorm:"index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
```

---

### `internal/services/key_manager.go` (service, CRUD)

**Analog:** `internal/pricing/manager.go` (lines 14-31)

**Service pattern:**
```go
// KeyManager API Key 管理器
type KeyManager struct {
	db *database.DB
}

// NewKeyManager 创建 Key 管理器
func NewKeyManager(db *database.DB) (*KeyManager, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	manager := &KeyManager{db: db}
	log.Println(" Key 管理器初始化成功")
	return manager, nil
}
```

**CRUD operation pattern** (from `internal/pricing/manager.go` lines 33-58):
```go
func (m *Manager) RecordUsage(ctx context.Context, usage *models.UsageRecord) error {
	if usage.ID == "" {
		usage.ID = xid.New().String()
	}

	if usage.CreatedAt.IsZero() {
		usage.CreatedAt = time.Now().UTC()
	}

	// 验证数据完整性
	if err := m.validateUsageRecord(usage); err != nil {
		return fmt.Errorf("invalid usage record: %w", err)
	}

	// 保存到数据库
	result := m.db.WithContext(ctx).Create(usage)
	if result.Error != nil {
		return fmt.Errorf("failed to save usage record: %w", result.Error)
	}

	log.Printf(" 使用量记录成功: 用户=%s 模型=%s",
		usage.UserID, usage.ModelID)

	return nil
}
```

---

### `internal/services/model_router.go` (service, event-driven)

**Analog:** `internal/cost/optimizer.go` (lines 66-115)

**Routing pattern:**
```go
func (o *Optimizer) SelectBestSupplier(ctx context.Context, userID, modelID string) (*SupplierSelection, error) {
	// 1. 获取用户所属群体
	userGroupID, err := o.getUserGroupID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user group: %w", err)
	}

	// 2. 获取所有可用供应商
	suppliers, err := o.getAvailableSuppliers(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get suppliers: %w", err)
	}

	if len(suppliers) == 0 {
		return nil, fmt.Errorf("no available suppliers for model: %s", modelID)
	}

	// 3. 评估每个供应商并选择最优
	var bestSupplier *SupplierSelection
	bestScore := math.Inf(-1)

	for _, supplier := range suppliers {
		selection, err := o.evaluateSupplier(ctx, supplier, userGroupID, modelID)
		if err != nil {
			log.Printf("警告: 评估供应商 %s 失败: %v", supplier.Name, err)
			continue
		}

		score := o.calculateScore(selection)
		if score > bestScore {
			bestScore = score
			bestSupplier = selection
		}
	}

	return bestSupplier, nil
}
```

---

### `internal/services/membership.go` (service, CRUD)

**Analog:** `internal/pricing/manager.go`

**Service pattern for membership:**
- Use same structure as `pricing.Manager`
- Methods: `GetUserMembership`, `ApplyDiscount`, `CalculatePrice`
- Use `db.WithContext(ctx)` for all queries

---

### `internal/services/billing.go` (service, batch)

**Analog:** `internal/pricing/manager.go` (lines 93-155)

**Bill generation pattern:**
```go
func (m *Manager) GenerateBill(ctx context.Context, userID string, startDate, endDate time.Time) (*Bill, error) {
	// 查询用户的使用记录
	var records []models.UsageRecord
	result := m.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Order("created_at ASC").
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch usage records: %w", result.Error)
	}

	// 按模型分组统计
	modelStats := make(map[string]*ModelUsage)
	for _, record := range records {
		// Aggregate statistics...
	}

	return bill, nil
}
```

---

### `pkg/handlers/images.go` (handler, request-response)

**Analog:** `pkg/handlers/chat.go` (lines 28-74)

**Handler pattern:**
```go
func (h *Handler) ChatCompletions(c *fiber.Ctx) error {
	// 获取用户 ID
	userID := middleware.GetUserID(c)

	// 解析请求
	var req api.ChatCompletionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Invalid request body",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	// 验证请求
	if err := h.validateChatRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 调用 Gateway
	resp, err := h.gateway.ChatCompletion(ctx, userID, &req)
	if err != nil {
		log.Printf("Chat completion 错误: 用户=%s 模型=%s 错误=%v", userID, req.Model, err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Failed to process request",
				Type:    "api_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	return c.JSON(resp)
}
```

**Validation pattern** (lines 149-196):
```go
func (h *Handler) validateChatRequest(req *api.ChatCompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if len(req.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	// 验证消息格式
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message %d: content is required", i)
		}
	}

	return nil
}
```

---

### `pkg/handlers/audio.go` (handler, streaming)

**Analog:** `pkg/handlers/chat.go` + streaming support

**Streaming pattern:**
```go
func (h *Handler) CreateSpeech(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req api.SpeechRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{...})
	}

	// 设置流式响应
	c.Set("Content-Type", "audio/mpeg")
	c.Set("Transfer-Encoding", "chunked")

	// 流式处理
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := h.gateway.TextToSpeech(ctx, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{...})
	}

	// 写入流
	for chunk := range stream {
		if _, err := c.Write(chunk); err != nil {
			return err
		}
	}

	return nil
}
```

---

### `pkg/handlers/embeddings.go` (handler, request-response)

**Analog:** `pkg/handlers/chat.go` (exact match for structure)

Use same pattern as `ChatCompletions` but with embedding-specific request/response types.

---

### `pkg/handlers/admin.go` (handler, request-response)

**Analog:** `pkg/handlers/chat.go`

**Admin handler pattern:**
```go
func (h *Handler) CreateModel(c *fiber.Ctx) error {
	// 解析请求
	var req api.CreateModelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{...})
	}

	// 调用服务层
	model, err := h.modelService.CreateModel(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{...})
	}

	return c.Status(fiber.StatusCreated).JSON(model)
}
```

---

### `internal/middleware/concurrency.go` (middleware, request-response)

**Analog:** `internal/middleware/ratelimit.go` (lines 12-71)

**Middleware pattern:**
```go
func (rl *RateLimiter) RateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Next()
		}

		// 检查是否超过限制
		allowed, retryAfter := rl.checkRateLimit(userID)
		if !allowed {
			log.Printf("限流触发: 用户=%s IP=%s", userID, c.IP())
			c.Set("Retry-After", retryAfter.String())
			return c.Status(fiber.StatusTooManyRequests).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Rate limit exceeded",
					Type:    "rate_limit_error",
					Code:    fiber.StatusTooManyRequests,
				},
			})
		}

		return c.Next()
	}
}
```

---

### `internal/middleware/validation.go` (middleware, request-response)

**Analog:** `internal/middleware/auth.go` (lines 18-64)

**Validation middleware pattern:**
```go
func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get(HeaderAPIKey)

		if strings.HasPrefix(apiKey, "Bearer ") {
			apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		}

		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Missing API key",
					Type:    "authentication_error",
					Code:    fiber.StatusUnauthorized,
				},
			})
		}

		// 验证逻辑
		userID := extractUserIDFromAPIKey(apiKey)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{...})
		}

		// 存储到上下文
		c.Locals("user_id", userID)
		c.Locals("api_key", apiKey)

		return c.Next()
	}
}
```

---

### `internal/utils/retry.go` (utility, transform)

**Analog:** `internal/cost/optimizer.go` (lines 267-296)

**Retry pattern with exponential backoff:**
```go
func (o *Optimizer) UpdatePerformanceMetric(supplierID string, responseTime float64, success bool) {
	// 指数移动平均 (EMA)
	alpha := 0.2
	perf.AvgResponseTime = alpha*responseTime + (1-alpha)*perf.AvgResponseTime
}
```

**For retry logic:**
```go
func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// 指数退避
		waitTime := time.Duration(1<<uint(i)) * time.Second
		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return lastErr
}
```

---

## Shared Patterns

### Authentication
**Source:** `internal/middleware/auth.go`
**Apply to:** All protected endpoints
```go
// 获取用户 ID
userID := middleware.GetUserID(c)

// 存储到上下文
c.Locals("user_id", userID)
c.Locals("api_key", apiKey)
```

### Error Handling
**Source:** `internal/middleware/error_handler.go` + `pkg/api/types.go`
**Apply to:** All handlers
```go
return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
	Error: api.ErrorDetail{
		Message: "Invalid request body",
		Type:    "invalid_request_error",
		Code:    fiber.StatusBadRequest,
	},
})
```

### ID Generation
**Source:** Multiple files use `xid.New().String()`
**Apply to:** All new models
```go
import "github.com/rs/xid"

if record.ID == "" {
    record.ID = xid.New().String()
}
```

### Database Context
**Source:** `internal/pricing/manager.go`
**Apply to:** All service methods
```go
func (m *Manager) MethodName(ctx context.Context, ...) error {
    result := m.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Find(&records)

    if result.Error != nil {
        return fmt.Errorf("failed to fetch: %w", result.Error)
    }

    return nil
}
```

### Logging Pattern
**Source:** All files
**Apply to:** All services
```go
log.Printf(" 成功: 用户=%s 模型=%s", userID, modelID)
log.Printf("警告: 操作失败: %v", err)
```

### Configuration Loading
**Source:** `internal/config/config.go`
**Apply to:** All new config sections
```go
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

---

## No Analog Found

Files with no close match in the codebase:

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/services/bifrost.go` | service | request-response | Bifrost integration is new, requires SDK study |
| `pkg/handlers/audio.go` | handler | streaming | No streaming handlers exist yet |

**For these files, use RESEARCH.md patterns and Bifrost SDK documentation.**

---

## Project Structure Conventions

```
ai-gateway/
├── cmd/
│   └── ai-gateway/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration
│   ├── database/                # Database connection and migrations
│   ├── gateway/                 # Core gateway logic
│   ├── middleware/              # Fiber middleware
│   ├── models/                  # GORM models
│   ├── pricing/                 # Pricing/billing services
│   ├── cost/                    # Cost optimization
│   ├── supplier/                # Supplier management
│   └── services/                # [NEW] Business logic services
├── pkg/
│   ├── handlers/                # HTTP handlers
│   ├── router/                  # Route setup
│   └── api/                     # API types/contracts
└── .planning/                   # Planning documentation
```

---

## Naming Conventions

1. **Models:** PascalCase, singular (`User`, `UsageRecord`)
2. **Services:** `XxxManager` or `XxxService` (`KeyManager`, `ModelRouter`)
3. **Handlers:** Methods on `Handler` struct (`ChatCompletions`, `ListModels`)
4. **Middleware:** `Xxx()` function returning `fiber.Handler` (`Auth()`, `RateLimit()`)
5. **Constants:** PascalCase or UPPER_SNAKE_CASE (`HeaderUserID`, `HeaderAPIKey`)
6. **Interfaces:** Implicit (no explicit interfaces unless needed for mocking)

---

## Metadata

**Analog search scope:** internal/, pkg/
**Files scanned:** 18 Go files
**Pattern extraction date:** 2026-05-08
**Phase:** 2-核心功能开发
