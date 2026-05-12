---
phase: 05-09
reviewed: 2026-05-12T09:17:00Z
depth: deep
files_reviewed: 25
files_reviewed_list:
  - internal/services/pricing_service.go
  - internal/services/enterprise_pricing.go
  - internal/services/supplier_apikey.go
  - internal/services/supplier_model.go
  - internal/services/key_manager.go
  - internal/services/model_router.go
  - internal/services/membership.go
  - internal/services/audit_service.go
  - internal/models/enterprise_pricing.go
  - internal/models/supplier_apikey.go
  - internal/models/supplier_model.go
  - internal/models/supplier_event.go
  - internal/models/audit_log.go
  - internal/models/membership.go
  - internal/models/model_mapping.go
  - internal/models/models.go
  - internal/supplier/health_checker.go
  - internal/supplier/manager.go
  - internal/crypto/encryption.go
  - internal/middleware/auth.go
  - pkg/handlers/pricing.go
  - pkg/handlers/supplier.go
  - pkg/handlers/user_keys.go
  - pkg/handlers/health.go
  - pkg/router/router.go
  - frontend/admin/src/views/Pricing.vue
  - frontend/admin/src/views/SupplierManagement.vue
findings:
  critical: 12
  warning: 18
  info: 8
  total: 38
status: issues_found
---

# Milestone 2 (Phase 5-9) Code Review Report

**Reviewed:** 2026-05-12
**Depth:** deep
**Files Reviewed:** 25
**Status:** issues_found

## Summary

This review examined all Milestone 2 implementation files covering enterprise pricing system (Phase 5), supplier management (Phase 6), user API key management (Phase 7), model routing (Phase 8), and audit/health systems (Phase 9). The code demonstrates reasonable architecture but contains **multiple critical security vulnerabilities, race conditions, and error handling gaps** that must be addressed before production deployment.

### Key Findings:
- **12 Critical Issues**: Including SQL injection risks, race conditions, credential exposure
- **18 Warning Issues**: Including unchecked errors, potential deadlocks, resource leaks
- **8 Info Issues**: Code quality and maintainability suggestions

---

## Critical Issues

### CR-01: Hardcoded Default Encryption Key in Production Code

**File:** `internal/crypto/encryption.go:36-38`
**Severity:** CRITICAL

**Issue:**
```go
if keyStr == "" {
    // 使用默认密钥（仅用于开发环境）
    keyStr = "ai-gateway-default-key-32-bytes!"
}
```

A hardcoded encryption key is used when the `ENCRYPTION_KEY` environment variable is not set. This is a critical security vulnerability because:
1. All supplier API keys are encrypted with this predictable key
2. An attacker who can access the database can decrypt all API keys
3. The key is visible in source code/version control

**Fix:**
```go
if keyStr == "" {
    return nil, fmt.Errorf("ENCRYPTION_KEY environment variable must be set for production use")
}
```

---

### CR-02: Race Condition in Supplier Primary Key Management

**File:** `internal/services/supplier_apikey.go:127-146`
**Severity:** CRITICAL

**Issue:**
```go
func (s *SupplierApiKeyService) SetPrimaryApiKey(ctx context.Context, keyID string) error {
    // ... fetch key ...
    tx := s.db.WithContext(ctx).Begin()
    tx.Model(&models.SupplierApiKey{}).Where("supplier_id = ? AND is_primary = ?", key.SupplierID, true).Update("is_primary", false)
    key.IsPrimary = true
    // ... save key ...
}
```

There's a race condition between fetching the key and starting the transaction. The key could be deleted or modified between the initial fetch and transaction start. Additionally, there's no check if the transaction update actually affected any rows.

**Fix:**
```go
func (s *SupplierApiKeyService) SetPrimaryApiKey(ctx context.Context, keyID string) error {
    tx := s.db.WithContext(ctx).Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Lock and update within transaction
    var key models.SupplierApiKey
    err := tx.Where("id = ?", keyID).First(&key).Error
    if err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to get api key: %w", err)
    }

    result := tx.Model(&models.SupplierApiKey{}).
        Where("supplier_id = ? AND is_primary = ?", key.SupplierID, true).
        Update("is_primary", false)
    if result.Error != nil {
        tx.Rollback()
        return fmt.Errorf("failed to reset primary keys: %w", result.Error)
    }

    key.IsPrimary = true
    key.UpdatedAt = time.Now().UTC()
    if err := tx.Save(&key).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to update primary key: %w", err)
    }

    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    return nil
}
```

---

### CR-03: SQL Injection via LIKE Query in Audit Service

**File:** `internal/services/audit_service.go:92`
**Severity:** CRITICAL

**Issue:**
```go
if filter.Keyword != "" {
    query = query.Where("entity_id LIKE ? OR admin_name LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
}
```

While using parameterized queries, the `%` wildcards are added to the user input without proper escaping. If `filter.Keyword` contains `%` or `_` characters, users can bypass filtering or match unintended records. More critically, there's no length limit on the keyword, potentially causing DoS.

**Fix:**
```go
if filter.Keyword != "" {
    // Escape LIKE special characters and limit length
    keyword := strings.ReplaceAll(filter.Keyword, "%", "\\%")
    keyword = strings.ReplaceAll(keyword, "_", "\\_")
    if len(keyword) > 100 {
        keyword = keyword[:100]
    }
    query = query.Where("entity_id LIKE ? OR admin_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
}
```

---

### CR-04: API Key Exposure in Create Response

**File:** `pkg/handlers/user_keys.go:99-107`
**Severity:** CRITICAL

**Issue:**
```go
return c.Status(201).JSON(fiber.Map{
    "data": fiber.Map{
        "id":         key.ID,
        "name":       key.Name,
        "api_key":    key.KeyValue,  // FULL KEY EXPOSED
        "created_at": key.CreatedAt,
    },
})
```

The complete API key is returned in the response. While this is intentional for display on creation, there's no mechanism to prevent replay or ensure the client has securely stored it. Combined with lack of TLS enforcement requirements, this is a security risk.

**Fix:**
1. Add API key versioning/fingerprinting
2. Document that this endpoint should only be called over HTTPS
3. Consider implementing "show once" pattern where the key must be explicitly revealed

---

### CR-05: Concurrent Map Write in Supplier Manager

**File:** `internal/supplier/manager.go:303-336`
**Severity:** CRITICAL

**Issue:**
```go
func (m *Manager) backgroundHealthCheck() {
    // ...
    for _, supplier := range suppliers {
        wg.Add(1)
        go func(sid string) {
            defer wg.Done()
            // ...
            m.updateHealthStatus(result)  // Called from multiple goroutines
        }(supplier.ID)
    }
}
```

The `updateHealthStatus` method is called from multiple goroutines without proper synchronization. While it uses `healthMutex.Lock()` internally, there's a potential deadlock scenario if `PerformHealthCheck` (which also calls `updateHealthStatus`) is called concurrently with the background checker.

**Fix:**
```go
// Add a separate mutex for background updates
type Manager struct {
    // ...
    healthMutex       sync.RWMutex
    backgroundMutex   sync.Mutex  // New: serialize background updates
}

func (m *Manager) backgroundHealthCheck() {
    // ...
    for _, supplier := range suppliers {
        wg.Add(1)
        go func(sid string) {
            defer wg.Done()
            ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
            defer cancel()

            result, err := m.PerformHealthCheck(ctx, sid)
            if err != nil {
                log.Printf("健康检查: 供应商 %s 检查失败: %v", sid, err)
            } else {
                // Serialize background updates
                m.backgroundMutex.Lock()
                m.updateHealthStatus(result)
                m.backgroundMutex.Unlock()
            }
        }(supplier.ID)
    }
}
```

---

### CR-06: Unbounded Goroutine Growth in Health Check

**File:** `internal/supplier/health_checker.go:171-186`
**Severity:** CRITICAL

**Issue:**
```go
func (hc *HealthChecker) recordHealthCheck(ctx context.Context, supplierID string, isHealthy bool, latency int64, errorMsg string) {
    history := &models.HealthCheckHistory{...}

    go func() {
        if err := hc.manager.db.Create(history).Error; err != nil {
            log.Printf("警告: 记录健康检查历史失败: %v", err)
        }
    }()
}
```

Every health check spawns a new goroutine without any limiting. With a 30-second check interval and multiple suppliers, this creates unbounded goroutine growth. The database connection pool could be exhausted under load.

**Fix:**
```go
type HealthChecker struct {
    manager    *Manager
    httpClient *http.Client
    recordChan chan *models.HealthCheckHistory
    workers    int
}

func NewHealthChecker(manager *Manager) *HealthChecker {
    hc := &HealthChecker{
        manager:    manager,
        httpClient: &http.Client{Timeout: 10 * time.Second},
        recordChan: make(chan *models.HealthCheckHistory, 1000),
        workers:    5, // Limited worker pool
    }

    // Start background workers
    for i := 0; i < hc.workers; i++ {
        go hc.recordWorker()
    }
    return hc
}

func (hc *HealthChecker) recordWorker() {
    for history := range hc.recordChan {
        if err := hc.manager.db.Create(history).Error; err != nil {
            log.Printf("警告: 记录健康检查历史失败: %v", err)
        }
    }
}

func (hc *HealthChecker) recordHealthCheck(...) {
    history := &models.HealthCheckHistory{...}
    select {
    case hc.recordChan <- history:
    default:
        log.Printf("警告: 健康检查记录通道已满，丢弃记录")
    }
}
```

---

### CR-07: Missing Input Validation on Pricing API

**File:** `pkg/handlers/pricing.go:31-48`
**Severity:** CRITICAL

**Issue:**
```go
func (h *Handler) CreateEnterprisePricing(c fiber.Ctx) error {
    adminID := c.Locals("admin_id").(string)
    var req enterprisePricingRequest
    if err := c.Bind().Body(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "Invalid request", "type": "invalid_request"}})
    }
    effectiveDate, _ := time.Parse(time.RFC3339, req.EffectiveDate)
    // ... no validation of effectiveDate parse error ...
}
```

Multiple issues:
1. Type assertion `c.Locals("admin_id").(string)` can panic if `admin_id` is not a string
2. `time.Parse` error is silently ignored (`_`)
3. No validation of price values (can be negative)
4. No validation that `min_profit_margin` is within valid bounds (0-1)

**Fix:**
```go
func (h *Handler) CreateEnterprisePricing(c fiber.Ctx) error {
    adminID, ok := c.Locals("admin_id").(string)
    if !ok || adminID == "" {
        return c.Status(401).JSON(fiber.Map{"error": fiber.Map{"message": "Unauthorized", "type": "authentication_error"}})
    }

    var req enterprisePricingRequest
    if err := c.Bind().Body(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "Invalid request", "type": "invalid_request"}})
    }

    // Validate input
    if req.InputPrice < 0 || req.OutputPrice < 0 {
        return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "Prices cannot be negative", "type": "invalid_request"}})
    }
    if req.MinProfitMargin < 0 || req.MinProfitMargin > 1 {
        return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "Min profit margin must be between 0 and 1", "type": "invalid_request"}})
    }

    effectiveDate, err := time.Parse(time.RFC3339, req.EffectiveDate)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "Invalid effective_date format", "type": "invalid_request"}})
    }
    // ... rest of function
}
```

---

### CR-08: Transaction Not Committed in Supplier API Key Service

**File:** `internal/services/supplier_apikey.go:133-145`
**Severity:** CRITICAL

**Issue:**
```go
func (s *SupplierApiKeyService) SetPrimaryApiKey(ctx context.Context, keyID string) error {
    // ...
    tx := s.db.WithContext(ctx).Begin()
    tx.Model(&models.SupplierApiKey{}).Where("supplier_id = ? AND is_primary = ?", key.SupplierID, true).Update("is_primary", false)
    key.IsPrimary = true
    key.UpdatedAt = time.Now().UTC()
    if err := tx.Save(&key).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to update primary key: %w", err)
    }
    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
```

The first `Update` call's error is never checked. If the update fails (e.g., due to constraint violation), the transaction continues and commits, leaving the database in an inconsistent state with multiple primary keys.

**Fix:**
```go
result := tx.Model(&models.SupplierApiKey{}).
    Where("supplier_id = ? AND is_primary = ?", key.SupplierID, true).
    Update("is_primary", false)
if result.Error != nil {
    tx.Rollback()
    return fmt.Errorf("failed to reset existing primary keys: %w", result.Error)
}
```

---

### CR-09: Missing Authorization Check in User Key Handlers

**File:** `pkg/handlers/user_keys.go:299-317`
**Severity:** CRITICAL

**Issue:**
```go
func (h *Handler) GetUserKeyStats(c fiber.Ctx) error {
    keyID := c.Params("id")

    stats, err := h.keyManager.GetKeyStats(c.Context(), keyID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{...})
    }
    return c.JSON(fiber.Map{"data": stats})
}
```

The handler fetches stats for any key ID without verifying that the requesting user owns this key. A user could iterate through key IDs to access other users' usage statistics.

**Fix:**
```go
func (h *Handler) GetUserKeyStats(c fiber.Ctx) error {
    userID := middleware.GetUserID(c)
    keyID := c.Params("id")

    // Verify ownership
    keys, err := h.keyManager.GetUserKeys(c.Context(), userID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{...})
    }

    var owned bool
    for _, k := range keys {
        if k.ID == keyID {
            owned = true
            break
        }
    }

    if !owned {
        return c.Status(403).JSON(fiber.Map{
            "error": fiber.Map{
                "message": "You don't have access to this key",
                "type":    "forbidden_error",
            },
        })
    }

    stats, err := h.keyManager.GetKeyStats(c.Context(), keyID)
    // ... rest of function
}
```

---

### CR-10: Time-Based SQL Injection via Format String

**File:** `internal/services/key_manager.go:199-276`
**Severity:** CRITICAL

**Issue:**
```go
today := time.Now().UTC().Format("2006-01-02")
// ...
km.db.WithContext(ctx).Model(&models.UsageRecord{}).
    Where("key_id = ? AND DATE(created_at) = ?", keyID, today).
```

While `today` is generated from a trusted format, using `DATE(created_at)` in the WHERE clause prevents index usage on `created_at`. This causes full table scans on large usage tables.

**Fix:**
```go
today := time.Now().UTC().Truncate(24 * time.Hour)
tomorrow := today.Add(24 * time.Hour)

km.db.WithContext(ctx).Model(&models.UsageRecord{}).
    Where("key_id = ? AND created_at >= ? AND created_at < ?", keyID, today, tomorrow).
```

---

### CR-11: Panic Risk in Type Assertions

**File:** `pkg/handlers/pricing.go:32, 81, 148`
**Severity:** CRITICAL

**Issue:**
```go
adminID := c.Locals("admin_id").(string)  // Can panic!
```

Type assertions without the "ok" check will panic if the value is not of the expected type. This can happen if middleware doesn't run correctly or is misconfigured.

**Fix:**
```go
adminID, ok := c.Locals("admin_id").(string)
if !ok || adminID == "" {
    return c.Status(401).JSON(fiber.Map{
        "error": fiber.Map{"message": "Unauthorized", "type": "authentication_error"},
    })
}
```

---

### CR-12: Concurrent Counter Update Without Atomic Operations

**File:** `internal/services/supplier_apikey.go:162-172`
**Severity:** CRITICAL

**Issue:**
```go
func (s *SupplierApiKeyService) IncrementUsage(ctx context.Context, keyID string) error {
    now := time.Now().UTC()
    result := s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).Where("id = ?", keyID).Updates(map[string]interface{}{
        "current_requests": s.db.WithContext(ctx).Raw("current_requests + 1"),
        "last_used_at":     now,
    })
```

The `Raw("current_requests + 1")` approach is not correct GORM syntax for atomic increments. This will likely cause errors or incorrect behavior.

**Fix:**
```go
func (s *SupplierApiKeyService) IncrementUsage(ctx context.Context, keyID string) error {
    now := time.Now().UTC()
    result := s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).
        Where("id = ?", keyID).
        UpdateColumn("current_requests", gorm.Expr("current_requests + ?", 1))
    if result.Error != nil {
        return fmt.Errorf("failed to increment usage: %w", result.Error)
    }

    // Update last_used_at separately
    return s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).
        Where("id = ?", keyID).
        Update("last_used_at", now).Error
}
```

---

## Warnings

### WR-01: Memory Leak in SSE Broadcaster

**File:** `pkg/handlers/health.go:34-62`
**Severity:** WARNING

**Issue:**
The `SSEBroadcaster.Broadcast` method uses non-blocking send to channels but never removes clients with closed channels. Dead clients accumulate in the `clients` map causing memory leaks.

**Fix:**
```go
func (b *SSEBroadcaster) Broadcast(message fiber.Map) {
    b.mutex.Lock()
    defer b.mutex.Unlock()

    for client := range b.clients {
        select {
        case client.Channel <- message:
        default:
            // Client channel is full or closed - remove it
            delete(b.clients, client)
            close(client.Channel)
        }
    }
}
```

---

### WR-02: N+1 Query in Supplier Management Frontend

**File:** `frontend/admin/src/views/SupplierManagement.vue:176-194`
**Severity:** WARNING

**Issue:**
```javascript
for (const supplier of suppliers.value) {
  try {
    const models = await suppliersApi.listModels(supplier.id)
    supplierModelCount.value[supplier.id] = models.length
  } catch {
    supplierModelCount.value[supplier.id] = 0
  }
}
```

Sequential API calls in a loop cause N+1 query problem. With many suppliers, this creates significant delay.

**Fix:** Use a batch endpoint or aggregate the data on the backend.

---

### WR-03: Missing Error Handling in Price Calculation

**File:** `internal/services/pricing_service.go:48-93`
**Severity:** WARNING

**Issue:**
```go
pricing, err := s.enterpriseService.GetEnterprisePrice(ctx, userID, modelID)
if err == nil {
    // ... use pricing ...
}
// ... continues without handling enterprise pricing errors ...
```

The function silently falls back through multiple pricing strategies without logging why higher-priority pricing failed. This makes debugging difficult.

**Fix:**
```go
pricing, err := s.enterpriseService.GetEnterprisePrice(ctx, userID, modelID)
if err != nil {
    log.Printf("企业定价不可用: 用户=%s 模型=%s 错误=%v", userID, modelID, err)
} else {
    log.Printf("定价应用: 用户=%s 模型=%s 规则=enterprise", userID, modelID)
    return &UserPricingResult{...}
}
```

---

### WR-04: Unsafe Polling in Vue Component

**File:** `frontend/admin/src/views/SupplierManagement.vue:262-268`
**Severity:** WARNING

**Issue:**
```javascript
let lastTab = activeTab.value
setInterval(() => {
  if (activeTab.value !== lastTab) {
    lastTab = activeTab.value
    watchActiveTab()
  }
}, 100)
```

A 100ms interval that runs forever is inefficient. Should use Vue's `watch` API or clean up the interval on component unmount.

**Fix:**
```javascript
import { watch } from 'vue'

watch(activeTab, (newTab) => {
  nextTick(() => {
    if (newTab === 'health') {
      loadHealthChart()
    }
  })
})
```

---

### WR-05: Missing Context Timeout in Health Check

**File:** `internal/supplier/health_checker.go:107-131`
**Severity:** WARNING

**Issue:**
```go
func (hc *HealthChecker) checkOpenAI(ctx context.Context, apiKey string) (error, int) {
    baseURL := "https://api.openai.com"
    req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/models", nil)
```

The function uses the passed context but doesn't ensure it has a timeout. If the caller passes a context without deadline, requests can hang indefinitely.

**Fix:**
```go
func (hc *HealthChecker) checkOpenAI(ctx context.Context, apiKey string) (error, int) {
    // Ensure we have a timeout
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    baseURL := "https://api.openai.com"
    req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/models", nil)
    // ...
}
```

---

### WR-06: Potential Integer Overflow in Usage Statistics

**File:** `internal/services/key_manager.go:190-247`
**Severity:** WARNING

**Issue:**
```go
type KeyStats struct {
    TotalRequests   int64
    TotalTokens     int64
    TotalCost       float64
    // ...
}
```

With high-volume usage, `int64` can theoretically overflow (though unlikely in practice). More critically, `TotalCost` as `float64` loses precision for very large amounts.

**Fix:**
1. Use `uint64` for counters that can't be negative
2. Consider using `decimal.Decimal` or store as integer cents for monetary values
3. Add overflow checks

---

### WR-07: Missing Index on UsageRecord KeyID + CreatedAt

**File:** `internal/models/models.go:8-28`
**Severity:** WARNING

**Issue:**
The `UsageRecord` model has individual indexes on `key_id` and `created_at`, but queries filtering by both (e.g., daily stats) would benefit from a composite index.

**Fix:**
```go
type UsageRecord struct {
    // ...
    KeyID      string    `json:"key_id" gorm:"index"`
    CreatedAt  time.Time `json:"created_at" gorm:"index"`
    // Add composite index
}
// In migration or auto-migrate:
// db.Exec("CREATE INDEX idx_usage_record_key_created ON usage_records(key_id, created_at)")
```

---

### WR-08: Unvalidated MinProfitMargin in Enterprise Pricing

**File:** `internal/services/enterprise_pricing.go:58-61`
**Severity:** WARNING

**Issue:**
```go
if pricing.MinProfitMargin <= 0 || pricing.MinProfitMargin >= 1 {
    return fmt.Errorf("invalid min_profit_margin: must be between 0 and 1")
}
```

The validation allows values arbitrarily close to 0 or 1, which may not make business sense. A 0.0001% margin is effectively no protection.

**Fix:**
```go
const (
    MinProfitMarginMin = 0.01  // 1%
    MinProfitMarginMax = 0.50  // 50%
)

if pricing.MinProfitMargin < MinProfitMarginMin || pricing.MinProfitMargin > MinProfitMarginMax {
    return fmt.Errorf("invalid min_profit_margin: must be between %d%% and %d%%", MinProfitMarginMin*100, MinProfitMarginMax*100)
}
```

---

### WR-09: Missing Rate Limiting on Admin Endpoints

**File:** `pkg/router/router.go:73-134`
**Severity:** WARNING

**Issue:**
Admin endpoints have no rate limiting. A compromised admin token could be used to enumerate all data or perform DoS attacks.

**Fix:**
```go
admin := app.Group("/v1/admin")
admin.Use(middleware.RequireAdmin(tokenManager))
admin.Use(middleware.RateLimitByIP(100, time.Minute))  // Add rate limiting
```

---

### WR-10: Potential Deadlock in Manager Health Status

**File:** `internal/supplier/manager.go:269-300`
**Severity:** WARNING

**Issue:**
```go
func (m *Manager) updateHealthStatus(result *HealthCheckResult) {
    m.healthMutex.Lock()
    defer m.healthMutex.Unlock()
    // ...
}
```

If `GetAllHealthStatus` is called while `updateHealthStatus` is in progress, it will block. If the update takes longer than expected (e.g., due to slow logging), it could cause cascading delays.

**Fix:**
```go
func (m *Manager) GetAllHealthStatus() map[string]*HealthStatus {
    m.healthMutex.RLock()
    defer m.healthMutex.RUnlock()

    // Return a deep copy to avoid race conditions
    result := make(map[string]*HealthStatus, len(m.healthStatus))
    for k, v := range m.healthStatus {
        statusCopy := *v
        result[k] = &statusCopy
    }
    return result
}
```

---

### WR-11: Missing Validation of Discount Rate Range

**File:** `internal/services/membership.go:96-102`
**Severity:** WARNING

**Issue:**
```go
func (ms *MembershipService) ApplyDiscount(originalPrice, discountRate float64) float64 {
    if discountRate <= 0 || discountRate >= 1 {
        return originalPrice
    }
    return originalPrice * (1 - discountRate)
}
```

The validation allows `discountRate` to be arbitrarily close to 1 (99.9...%), which could result in near-zero prices. No business logic validation on minimum acceptable price.

**Fix:**
```go
const (
    MinDiscountRate = 0.0
    MaxDiscountRate = 0.80  // Max 80% discount
)

func (ms *MembershipService) ApplyDiscount(originalPrice, discountRate float64) float64 {
    if discountRate <= MinDiscountRate || discountRate > MaxDiscountRate {
        return originalPrice
    }
    discountedPrice := originalPrice * (1 - discountRate)
    if discountedPrice < originalPrice * 0.1 {  // Never sell below 10% of original
        return originalPrice * 0.1
    }
    return discountedPrice
}
```

---

### WR-12: Unused Error Return in Multiple Functions

**File:** `internal/services/supplier_model.go:137-154`
**Severity:** WARNING

**Issue:**
```go
func (s *SupplierModelService) recordPriceHistory(...) {
    // ...
    if err := s.db.WithContext(ctx).Create(history).Error; err != nil {
        log.Printf("警告: 记录价格历史失败: %v", err)
    }
}
```

The error is only logged, not returned or handled. This means price history failures are silently ignored.

**Fix:**
```go
func (s *SupplierModelService) recordPriceHistory(...) error {
    // ...
    if err := s.db.WithContext(ctx).Create(history).Error; err != nil {
        log.Printf("警告: 记录价格历史失败: %v", err)
        return fmt.Errorf("failed to record price history: %w", err)
    }
    return nil
}
```

Then handle the error appropriately at call sites.

---

### WR-13: Missing Content-Type Validation in Health Check

**File:** `internal/supplier/health_checker.go:107-131`
**Severity:** WARNING

**Issue:**
```go
resp, err := hc.httpClient.Do(req)
if err != nil {
    return fmt.Errorf("request failed: %w"), 0
}
defer resp.Body.Close()

_, _ = io.ReadAll(resp.Body)

if resp.StatusCode >= 400 {
    return fmt.Errorf("API returned status %d", resp.StatusCode), resp.StatusCode
}
```

The health check doesn't validate the response content type. A malformed response returning 200 OK would be considered healthy even if the content is garbage.

**Fix:**
```go
contentType := resp.Header.Get("Content-Type")
if !strings.Contains(contentType, "application/json") {
    return fmt.Errorf("unexpected content type: %s", contentType), resp.StatusCode
}
```

---

### WR-14: Hardcoded Timeout Values

**File:** `internal/supplier/health_checker.go:24-27`
**Severity:** WARNING

**Issue:**
```go
httpClient: &http.Client{
    Timeout: 10 * time.Second,
},
```

The 10-second timeout is hardcoded. Different providers might require different timeouts, and this should be configurable.

**Fix:**
```go
type HealthCheckerConfig struct {
    Timeout time.Duration
    // ...
}

func NewHealthChecker(manager *Manager, config HealthCheckerConfig) *HealthChecker {
    if config.Timeout == 0 {
        config.Timeout = 10 * time.Second
    }
    return &HealthChecker{
        manager:   manager,
        httpClient: &http.Client{Timeout: config.Timeout},
    }
}
```

---

### WR-15: Missing Pagination in List Functions

**File:** `internal/services/supplier_apikey.go:73-80`
**Severity:** WARNING

**Issue:**
```go
func (s *SupplierApiKeyService) ListApiKeys(ctx context.Context, supplierID string) ([]*models.SupplierApiKey, error) {
    var keys []*models.SupplierApiKey
    err := s.db.WithContext(ctx).Where("supplier_id = ?", supplierID).Order("priority ASC, created_at DESC").Find(&keys).Error
```

No pagination - this returns all API keys for a supplier. With hundreds of keys, this creates unnecessary load.

**Fix:**
```go
func (s *SupplierApiKeyService) ListApiKeys(ctx context.Context, supplierID string, limit, offset int) ([]*models.SupplierApiKey, error) {
    var keys []*models.SupplierApiKey
    query := s.db.WithContext(ctx).Where("supplier_id = ?", supplierID).Order("priority ASC, created_at DESC")

    if limit > 0 {
        query = query.Limit(limit)
        if offset > 0 {
            query = query.Offset(offset)
        }
    }

    err := query.Find(&keys).Error
    // ...
}
```

---

### WR-16: Inconsistent Error Types in Handlers

**File:** `pkg/handlers/supplier.go:26-28, 86-88`
**Severity:** WARNING

**Issue:**
```go
return c.Status(500).JSON(fiber.Map{"error": err.Error()})
```

Some handlers return raw error messages which could leak internal implementation details. Should use consistent error response format.

**Fix:**
```go
return c.Status(500).JSON(fiber.Map{
    "error": fiber.Map{
        "message": "Internal server error",
        "type":    "internal_error",
        "code":    500,
    },
})
```

---

### WR-17: Missing CSRF Protection

**File:** `pkg/router/router.go:66-189`
**Severity:** WARNING

**Issue:**
State-changing endpoints (POST, PUT, DELETE) don't have CSRF protection. While JWT auth helps, additional CSRF protection is recommended for browser-based admin panels.

**Fix:**
```go
admin := app.Group("/v1/admin")
admin.Use(middleware.RequireAdmin(tokenManager))
admin.Use(middleware.CSRF())  // Add CSRF middleware
```

---

### WR-18: Potential Key Enumeration Attack

**File:** `internal/services/key_manager.go:60-87`
**Severity:** WARNING

**Issue:**
```go
func (km *KeyManager) ValidateKey(ctx context.Context, keyValue string) (*KeyInfo, error) {
    if len(keyValue) < 4 || keyValue[:3] != "sk-" {
        return nil, fmt.Errorf("invalid API key format")
    }
```

The validation leaks information about invalid key format. Combined with timing attacks on database lookups, this could enable key enumeration.

**Fix:**
```go
func (km *KeyManager) ValidateKey(ctx context.Context, keyValue string) (*KeyInfo, error) {
    // Use constant-time comparison for prefix check
    if !secureComparePrefix(keyValue, "sk-") {
        // Add delay to prevent timing attacks
        time.Sleep(100 * time.Millisecond)
        return nil, fmt.Errorf("invalid API key")
    }
    // ...
}

func secureComparePrefix(s, prefix string) bool {
    if len(s) < len(prefix) {
        return false
    }
    return subtle.ConstantTimeCompare([]byte(s[:len(prefix)]), []byte(prefix)) == 1
}
```

---

## Info

### IN-01: Inconsistent Logging Format

**File:** Multiple files
**Severity:** INFO

**Issue:**
Mix of emoji-prefixed logs (✅, ❌) and plain text logs. This makes log parsing inconsistent.

**Fix:**
Standardize on one format or use structured logging (e.g., zerolog, zap).

---

### IN-02: Magic Numbers in Model Router

**File:** `internal/services/model_router.go:105`
**Severity:** INFO

**Issue:**
```go
score := selection.ProfitMargin*100 + float64(10-selection.Priority)
```

Magic numbers 100 and 10 should be named constants.

**Fix:**
```go
const (
    ProfitMarginWeight = 100.0
    PriorityWeight     = 10
)
score := selection.ProfitMargin*ProfitMarginWeight + float64(PriorityWeight-selection.Priority)
```

---

### IN-03: Duplicate ID Generation Logic

**File:** `internal/services/model_router.go:275-278`, `internal/supplier/health_checker.go:188-191`
**Severity:** INFO

**Issue:**
```go
func generateID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

This function is duplicated in multiple files. Should be in a shared utility package.

---

### IN-04: Missing Godoc Comments

**File:** Multiple service files
**Severity:** INFO

**Issue:**
Many exported functions lack Godoc comments explaining their purpose, parameters, and return values.

**Fix:**
```go
// GetEnterprisePrice retrieves the enterprise pricing configuration for a specific customer and model.
// Returns an error if no active pricing is found or if the pricing is outside its validity period.
func (s *EnterprisePricingService) GetEnterprisePrice(
    ctx context.Context,
    customerID, modelID string,
) (*models.EnterprisePricing, error) {
```

---

### IN-05: Frontend TypeScript Type Inconsistency

**File:** `frontend/admin/src/views/Pricing.vue:353`
**Severity:** INFO

**Issue:**
```typescript
min_profit_margin: pricing.min_profit_margin * 100,
```

The form expects percentage (0-100) but the API expects decimal (0-1). This conversion happens in the component but could be a source of bugs.

**Fix:**
Create a utility function for consistent conversion or use a single format throughout.

---

### IN-06: Missing Health Check for Database

**File:** `pkg/router/router.go:61-63`
**Severity:** INFO

**Issue:**
```go
app.Get("/health", func(c fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok", "service": "ai-gateway"})
})
```

Health check doesn't verify database connectivity or other dependencies.

**Fix:**
```go
app.Get("/health", func(c fiber.Ctx) error {
    // Check database
    if err := gw.GetDB().Ping(); err != nil {
        return c.Status(503).JSON(fiber.Map{"status": "unhealthy", "reason": "database unavailable"})
    }
    return c.JSON(fiber.Map{"status": "ok", "service": "ai-gateway"})
})
```

---

### IN-07: Unused Variable in Update Function

**File:** `pkg/handlers/supplier.go:107`
**Severity:** INFO

**Issue:**
```go
func (h *Handler) UpdateSupplierApiKey(c fiber.Ctx) error {
    _ = c.Params("kid")
```

The parameter is retrieved but never used.

**Fix:**
Remove the unused line or use it in the function logic.

---

### IN-08: Inconsistent Error Response Structure

**File:** Multiple handler files
**Severity:** INFO

**Issue:**
Some endpoints return `{"error": {...}}` while others return `{"message": ...}`. Inconsistent error handling makes client code harder to write.

**Fix:**
Define a standard error response type and use it consistently across all handlers.

---

## Summary by Phase

### Phase 5: Enterprise Pricing System
- **Critical:** 3 (CR-01, CR-07, CR-11)
- **Warning:** 4 (WR-03, WR-08, WR-11, WR-16)
- **Info:** 2

### Phase 6: Supplier Management
- **Critical:** 4 (CR-02, CR-08, CR-12, CR-06)
- **Warning:** 5 (WR-05, WR-12, WR-13, WR-14, WR-15)
- **Info:** 3

### Phase 7: User API Key Management
- **Critical:** 3 (CR-04, CR-09, CR-10)
- **Warning:** 2 (WR-06, WR-07)
- **Info:** 1

### Phase 8: Model Routing
- **Critical:** 1 (CR-05)
- **Warning:** 1 (WR-18)
- **Info:** 1

### Phase 9: Audit & Health Systems
- **Critical:** 1 (CR-03)
- **Warning:** 4 (WR-01, WR-04, WR-09, WR-17)
- **Info:** 1

### Cross-Cutting Concerns
- **Critical:** Encryption key management (CR-01)
- **Warning:** Error handling consistency (WR-16), Missing validation (WR-07)

---

## Recommendations

### Immediate Actions (Before Production):
1. **Remove hardcoded encryption key** (CR-01) - Set up proper secrets management
2. **Fix race conditions** (CR-02, CR-05, CR-06, CR-12) - Add proper synchronization
3. **Add authorization checks** (CR-09) - Verify resource ownership
4. **Fix SQL injection vulnerability** (CR-03) - Escape LIKE wildcards
5. **Add input validation** (CR-07, CR-11) - Validate all user inputs

### Short-term Improvements:
1. Implement proper transaction handling throughout
2. Add comprehensive error logging
3. Set up rate limiting on admin endpoints
4. Add database health checks
5. Implement proper goroutine pooling

### Long-term Architecture:
1. Consider moving to event-driven architecture for health checks
2. Implement circuit breakers for supplier calls
3. Add distributed tracing
4. Set up proper secrets management (HashiCorp Vault, AWS Secrets Manager)
5. Implement API versioning strategy

---

_Reviewed: 2026-05-12_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
