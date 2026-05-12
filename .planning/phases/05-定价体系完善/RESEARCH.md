# Phase 5: 定价体系完善 - 研究文档

## 研究日期
2026-05-11

## 研究目标

研究 AI Gateway 定价体系的完善方案，包括：
1. 大客户独立定价实现
2. 套餐价格应用逻辑
3. 利润保护机制
4. 定价管理界面

---

## 当前实现分析

### 已有组件

#### 1. MembershipService (internal/services/membership.go)
- ✅ GetUserMembership: 获取用户会员信息
- ✅ GetModelDiscount: 获取模型折扣率
- ✅ ApplyDiscount: 应用折扣
- ✅ CalculatePrice: 计算会员价格
- ✅ UpdateUserMembership: 更新用户会员
- ✅ SetModelDiscount: 设置模型折扣

**问题**: CalculatePrice 方法已存在但未在路由决策中集成

#### 2. Pricing Manager (internal/pricing/manager.go)
- ✅ RecordUsage: 记录使用量
- ✅ GenerateBill: 生成账单
- ✅ GetUserUsageSummary: 获取使用摘要
- ✅ GetModelUsageStats: 获取模型统计

**问题**: 缺少大客户定价支持

#### 3. 数据模型 (internal/models/)
- ✅ MembershipTier: 会员等级
- ✅ MembershipDiscount: 会员折扣
- ✅ UserMembership: 用户会员关联
- ✅ UserGroupPricing: 用户群体定价
- ❌ EnterprisePricing: **缺失**

---

## 技术方案设计

### 方案 1: 大客户独立定价

#### 数据模型
```go
// EnterprisePricing 大客户独立定价
type EnterprisePricing struct {
    ID               string    `gorm:"primaryKey"`
    CustomerID       string    `gorm:"index"`
    CustomerName     string
    ModelID          string    `gorm:"index"`
    
    // 定价配置
    InputPrice       float64   // 输入价格 per 1M tokens
    OutputPrice      float64   // 输出价格 per 1M tokens
    
    // 利润保护
    MinProfitMargin  float64   // 最低利润率 (0.1 = 10%)
    MaxCostPrice     float64   // 最高可接受成本价
    
    // 生效时间
    EffectiveDate    time.Time
    ExpiryDate       *time.Time // NULL = 无限期
    
    // 状态
    IsActive         bool      `gorm:"index"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
    CreatedBy        string    // 创建人
    UpdatedBy        string    // 更新人
}
```

#### 服务方法
```go
type EnterprisePricingService struct {
    db *database.DB
}

// GetEnterprisePrice 获取大客户价格
func (s *EnterprisePricingService) GetEnterprisePrice(
    ctx context.Context, 
    customerID, modelID string,
) (*EnterprisePricing, error)

// SetEnterprisePrice 设置大客户价格
func (s *EnterprisePricingService) SetEnterprisePrice(
    ctx context.Context,
    pricing *EnterprisePricing,
) error

// ValidateProfitMargin 验证利润率
func (s *EnterprisePricingService) ValidateProfitMargin(
    sellingPrice, costPrice, minMargin float64,
) error
```

---

### 方案 2: 统一价格计算服务

#### 价格优先级
```
1. 大客户独立定价 (EnterprisePricing)
   ↓ 如果不存在
2. 会员折扣价格 (MembershipDiscount)
   ↓ 如果不存在
3. 用户群体定价 (UserGroupPricing)
   ↓ 默认
4. 基础价格 (UserGroupPricing with basic group)
```

#### 服务接口
```go
type PricingService struct {
    db                *database.DB
    membershipService *MembershipService
    enterpriseService *EnterprisePricingService
}

// GetUserPrice 获取用户适用的价格
// 自动应用价格优先级：大客户 > 会员 > 用户组 > 基础
func (s *PricingService) GetUserPrice(
    ctx context.Context,
    userID, modelID string,
) (*UserPricingResult, error)

type UserPricingResult struct {
    UserID           string
    ModelID          string
    InputPrice       float64
    OutputPrice      float64
    AppliedRule      string  // "enterprise", "membership", "group", "basic"
    DiscountRate     float64 // 应用的折扣率
    EffectiveDate    time.Time
}
```

---

### 方案 3: 利润保护机制

#### 验证流程
```go
func (s *PricingService) ValidatePriceWithProfitProtection(
    ctx context.Context,
    userID, modelID, supplierID string,
    sellingPrice float64,
) (*PriceValidationResult, error) {
    
    // 1. 获取供应商成本价
    costPrice, err := s.getSupplierCostPrice(ctx, supplierID, modelID)
    if err != nil {
        return nil, err
    }
    
    // 2. 计算利润
    profit := sellingPrice - costPrice
    profitMargin := profit / sellingPrice
    
    // 3. 获取最低利润率要求
    minMargin, err := s.getMinProfitMargin(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 4. 验证
    if profitMargin < minMargin {
        return &PriceValidationResult{
            Valid: false,
            Reason: "profit_margin_too_low",
            ProfitMargin: profitMargin,
            RequiredMargin: minMargin,
            CostPrice: costPrice,
            SellingPrice: sellingPrice,
        }, nil
    }
    
    return &PriceValidationResult{
        Valid: true,
        ProfitMargin: profitMargin,
        CostPrice: costPrice,
        SellingPrice: sellingPrice,
    }, nil
}
```

---

## API 设计

### 大客户定价管理 API (运维端)

```go
// EnterprisePricingHandler 大客户定价处理器
type EnterprisePricingHandler struct {
    service *EnterprisePricingService
}

// ListEnterprisePricing 获取大客户定价列表
// GET /api/v1/admin/enterprise-pricing
func (h *EnterprisePricingHandler) ListEnterprisePricing(c fiber.Ctx) error

// CreateEnterprisePricing 创建大客户定价
// POST /api/v1/admin/enterprise-pricing
func (h *EnterprisePricingHandler) CreateEnterprisePricing(c fiber.Ctx) error

// GetEnterprisePricing 获取大客户定价详情
// GET /api/v1/admin/enterprise-pricing/:id
func (h *EnterprisePricingHandler) GetEnterprisePricing(c fiber.Ctx) error

// UpdateEnterprisePricing 更新大客户定价
// PUT /api/v1/admin/enterprise-pricing/:id
func (h *EnterprisePricingHandler) UpdateEnterprisePricing(c fiber.Ctx) error

// DeleteEnterprisePricing 删除大客户定价
// DELETE /api/v1/admin/enterprise-pricing/:id
func (h *EnterprisePricingHandler) DeleteEnterprisePricing(c fiber.Ctx) error
```

### 用户价格查询 API (用户端)

```go
// GetUserPricing 获取用户适用的价格
// GET /api/v1/user/pricing?model_id=gpt-4
func (h *UserPricingHandler) GetUserPricing(c fiber.Ctx) error

// 响应示例
{
    "model_id": "gpt-4",
    "input_price": 15.0,      // per 1M tokens
    "output_price": 30.0,     // per 1M tokens
    "applied_rule": "enterprise",
    "discount_rate": 0.0,
    "currency": "CNY"
}
```

---

## 前端界面设计

### 运维端 - 大客户定价管理

**页面路径**: `/admin/pricing/enterprise`

**功能**:
1. 大客户列表
2. 定价配置表单
3. 利润率验证
4. 定价历史查看

**组件结构**:
```
EnterprisePricing.vue
├── CustomerSelector.vue     // 客户选择器
├── PricingForm.vue          // 定价表单
├── ProfitMarginDisplay.vue  // 利润率展示
└── PricingHistory.vue       // 定价历史
```

### 运维端 - 用户组定价管理

**页面路径**: `/admin/pricing/groups`

**功能**:
1. 用户组管理
2. 定价配置
3. 折扣设置

---

## 实现计划

### 任务分解

1. **数据模型扩展** (1 天)
   - 创建 EnterprisePricing 模型
   - 数据库迁移
   - GORM 自动迁移配置

2. **后端服务实现** (2 天)
   - EnterprisePricingService
   - 统一 PricingService
   - 利润保护验证
   - API 处理器

3. **前端界面开发** (1 天)
   - 大客户定价管理页面
   - 用户组定价管理页面
   - API 集成

4. **集成测试** (0.5 天)
   - 价格计算测试
   - 利润保护测试
   - API 测试

5. **文档更新** (0.5 天)
   - API 文档
   - 数据模型文档
   - 部署指南

---

## 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 价格计算错误 | 高 | 单元测试覆盖 |
| 利润率验证遗漏 | 高 | 集成测试验证 |
| 性能问题 | 中 | 缓存价格计算结果 |
| 数据迁移失败 | 中 | 备份和回滚计划 |

---

## 参考资料

- 现有 MembershipService 实现
- Bifrost 价格体系文档
- Phase 2 数据模型设计

---

**研究者**: GSD Phase 5 Planner
**研究完成时间**: 2026-05-11
