# Phase 5: 定价体系完善 - 执行计划

## 基本信息

**Phase ID**: 05
**Phase 名称**: 定价体系完善
**里程碑**: Milestone 2
**预计时间**: 5 天
**状态**: 📝 计划中

---

## Phase 目标

完善 AI Gateway 的定价体系，实现：
1. 大客户独立定价功能
2. 套餐价格应用逻辑集成
3. 利润保护机制增强
4. 定价管理运维界面

---

## 验收标准

### 功能验收
- [ ] 大客户可设置独立价格
- [ ] 会员价格正确计算和应用
- [ ] 利润保护机制有效工作
- [ ] 运维管理界面完整可用

### 质量验收
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试通过
- [ ] 代码审查通过

---

## 执行计划

### Task 5.1: 创建 EnterprisePricing 数据模型

**文件**: `internal/models/enterprise_pricing.go`

**描述**: 创建大客户独立定价的数据模型

**步骤**:
1. 创建 `EnterprisePricing` 结构体
2. 添加 GORM 标签和验证
3. 添加数据库索引
4. 更新 `database.go` 自动迁移

**验收**:
- [ ] 模型定义完成
- [ ] GORM 自动迁移成功
- [ ] 表结构正确创建

**预计时间**: 2 小时

---

### Task 5.2: 实现 EnterprisePricingService

**文件**: `internal/services/enterprise_pricing.go`

**描述**: 实现大客户定价的业务逻辑

**方法列表**:
```go
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

// ListEnterprisePricing 列出大客户定价
func (s *EnterprisePricingService) ListEnterprisePricing(
    ctx context.Context,
    filter *EnterprisePricingFilter,
) ([]*EnterprisePricing, error)

// UpdateEnterprisePrice 更新大客户价格
func (s *EnterprisePricingService) UpdateEnterprisePrice(
    ctx context.Context,
    pricing *EnterprisePricing,
) error

// DeleteEnterprisePrice 删除大客户价格
func (s *EnterprisePricingService) DeleteEnterprisePrice(
    ctx context.Context,
    id string,
) error

// ValidateProfitMargin 验证利润率
func (s *EnterprisePricingService) ValidateProfitMargin(
    sellingPrice, costPrice, minMargin float64,
) error
```

**验收**:
- [ ] 所有方法实现完成
- [ ] 单元测试通过
- [ ] 错误处理完善

**预计时间**: 4 小时

---

### Task 5.3: 实现统一 PricingService

**文件**: `internal/services/pricing_service.go`

**描述**: 创建统一的价格计算服务，集成大客户定价、会员折扣、用户组定价

**核心方法**:
```go
// GetUserPrice 获取用户适用的价格
// 价格优先级: 大客户 > 会员 > 用户组 > 基础
func (s *PricingService) GetUserPrice(
    ctx context.Context,
    userID, modelID string,
) (*UserPricingResult, error)

// CalculatePriceWithProfit 计算带利润保护的价格
func (s *PricingService) CalculatePriceWithProfit(
    ctx context.Context,
    userID, modelID, supplierID string,
) (*PriceCalculationResult, error)

// RecordPriceApplication 记录价格应用
func (s *PricingService) RecordPriceApplication(
    ctx context.Context,
    record *PriceApplicationRecord,
) error
```

**数据结构**:
```go
type UserPricingResult struct {
    UserID           string
    ModelID          string
    InputPrice       float64
    OutputPrice      float64
    AppliedRule      string  // "enterprise", "membership", "group", "basic"
    DiscountRate     float64
    EffectiveDate    time.Time
}

type PriceCalculationResult struct {
    CostPrice        float64
    SellingPrice     float64
    Profit           float64
    ProfitMargin     float64
    IsValid          bool
    ValidationErrors []string
}
```

**验收**:
- [ ] 价格优先级正确实现
- [ ] 利润保护验证有效
- [ ] 单元测试覆盖所有场景

**预计时间**: 6 小时

---

### Task 5.4: 创建 API 处理器

**文件**: `pkg/handlers/pricing.go`

**描述**: 实现定价管理的 API 端点

**API 端点**:
```
# 运维端 - 大客户定价管理
GET    /api/v1/admin/enterprise-pricing          # 列表
POST   /api/v1/admin/enterprise-pricing          # 创建
GET    /api/v1/admin/enterprise-pricing/:id      # 详情
PUT    /api/v1/admin/enterprise-pricing/:id      # 更新
DELETE /api/v1/admin/enterprise-pricing/:id      # 删除

# 用户端 - 价格查询
GET    /api/v1/user/pricing                      # 获取价格
```

**验收**:
- [ ] 所有 API 端点实现
- [ ] 请求验证正确
- [ ] 错误响应符合规范
- [ ] 集成测试通过

**预计时间**: 4 小时

---

### Task 5.5: 更新路由配置

**文件**: `pkg/router/router.go`

**描述**: 添加新的定价管理路由

**步骤**:
1. 注册大客户定价路由
2. 注册用户价格查询路由
3. 添加认证中间件
4. 添加管理员权限验证

**验收**:
- [ ] 路由正确注册
- [ ] 权限控制有效
- [ ] 路由测试通过

**预计时间**: 1 小时

---

### Task 5.6: 开发运维端定价管理界面

**文件**: `frontend/admin/src/views/Pricing.vue`

**描述**: 创建大客户定价管理页面

**组件结构**:
```
Pricing.vue
├── PricingTabs.vue           // 定价类型标签页
├── EnterprisePricingTab.vue  // 大客户定价
├── GroupPricingTab.vue       // 用户组定价
└── PricingForm.vue           // 定价表单
```

**功能**:
- 大客户定价列表
- 创建/编辑定价表单
- 利润率实时计算
- 定价历史查看

**验收**:
- [ ] 界面美观易用
- [ ] 所有功能可用
- [ ] API 集成正确

**预计时间**: 6 小时

---

### Task 5.7: 集成测试

**文件**: `tests/integration/pricing/pricing_test.go`

**描述**: 编写定价体系的集成测试

**测试场景**:
1. 大客户定价创建和查询
2. 价格优先级验证
3. 利润保护验证
4. API 端到端测试

**验收**:
- [ ] 所有测试通过
- [ ] 测试覆盖率 > 80%

**预计时间**: 3 小时

---

## 依赖关系

```
Task 5.1 (数据模型)
    ↓
Task 5.2 (EnterprisePricingService) ←→ Task 5.3 (PricingService)
    ↓                              ↓
Task 5.4 (API 处理器) ←──────────────┘
    ↓
Task 5.5 (路由配置)
    ↓
Task 5.6 (前端界面)
    ↓
Task 5.7 (集成测试)
```

---

## Wave 并行计划

可与 **Phase 6 (供应商管理增强)** 并行执行：

- **Wave 1-并行**: Phase 5 Task 5.1-5.3 || Phase 6 Task 6.1-6.2
- **Wave 2-并行**: Phase 5 Task 5.4-5.5 || Phase 6 Task 6.3-6.4
- **Wave 3-串行**: Phase 5 Task 5.6-5.7 → 前端集成测试

---

## 文件变更清单

### 新建文件
- `internal/models/enterprise_pricing.go`
- `internal/services/enterprise_pricing.go`
- `internal/services/pricing_service.go`
- `pkg/handlers/pricing.go`
- `frontend/admin/src/views/Pricing.vue`
- `frontend/admin/src/components/pricing/*.vue`
- `tests/integration/pricing/pricing_test.go`

### 修改文件
- `internal/database/database.go` (添加自动迁移)
- `pkg/router/router.go` (添加路由)
- `frontend/admin/src/router/index.ts` (添加路由)

---

## 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 价格计算错误 | 高 | 低 | 单元测试 + 人工验证 |
| 数据库迁移失败 | 中 | 低 | 备份 + 回滚脚本 |
| 性能问题 | 中 | 中 | 价格结果缓存 |
| 前后端接口不一致 | 中 | 中 | API 文档 + 类型共享 |

---

## 完成标志

- [ ] 所有 Task 完成
- [ ] 验收标准全部满足
- [ ] 代码审查通过
- [ ] 测试报告生成
- [ ] SUMMARY.md 编写

---

**计划创建时间**: 2026-05-11
**计划创建者**: GSD Phase 5 Planner
**预计开始时间**: 待确认
**预计完成时间**: 开始后 5 个工作日
