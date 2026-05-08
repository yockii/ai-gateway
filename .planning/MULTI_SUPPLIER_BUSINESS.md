# 多供应商路由和成本优化业务需求确认

## 需求理解确认

您提出的业务需求非常关键，我完全理解了核心逻辑。以下是详细的业务模型说明：

---

## 核心业务模式

### 1. 第三方供应商模式

**不是直接对接原厂商**，而是对接第三方供应商，原因：
- ✅ 可能拿到更优价格
- ✅ 更灵活的付款条件
- ✅ 更好的服务保障
- ✅ 降低对单一供应商的依赖

### 2. 多供应商配置示例

```
对外模型: "GPT-4-Turbo"

供应商配置:
├── 供应商A (主要)
│   ├── 提供: OpenAI GPT-4-Turbo 接口
│   ├── 成本: 10元/M tokens
│   ├── 权重: 70% (主要流量)
│   └── 状态: 正常
│
├── 供应商B (辅助)
│   ├── 提供: OpenAI GPT-4-Turbo 接口
│   ├── 成本: 15元/M tokens
│   ├── 权重: 30% (辅助流量)
│   └── 状态: 正常
│
└── 供应商C (备用)
    ├── 提供: OpenAI GPT-4-Turbo 接口
    ├── 成本: 12元/M tokens
    ├── 权重: 0% (仅故障转移)
    └── 状态: 备用
```

### 3. 差异化定价策略

```
对外模型: "GPT-4-Turbo" (输入 tokens)

用户群体定价:
├── 普通用户
│   ├── 售价: 18元/M tokens
│   ├── 优先供应商A: 成本10元, 利润8元 (44% 利润率)
│   └── 备选供应商B: 成本15元, 利润3元 (17% 利润率)
│
├── VIP会员
│   ├── 售价: 16元/M tokens
│   ├── 优先供应商A: 成本10元, 利润6元 (38% 利润率)
│   └── 备选供应商B: 成本15元, 利润1元 (6% 利润率)
│
└── 大客户X
    ├── 独立定价: 14元/M tokens
    ├── 优先供应商A: 成本10元, 利润4元 (29% 利润率)
    └── 备选供应商B: 成本15元, 利润-1元 (亏损!) ❌
```

---

## 智能路由策略

### 1. 成本优先路由

**路由决策逻辑**:
```
用户请求: VIP会员调用 "GPT-4-Turbo"
↓
确定售价: 16元/M tokens
↓
计算各供应商利润:
  ├── 供应商A: 16-10=6元 (利润率 38%) ✅ 优先选择
  ├── 供应商B: 16-15=1元 (利润率 6%)  ⚠️ 可选
  └── 供应商C: 16-12=4元 (利润率 25%) ✅ 备选
↓
检查供应商状态和负载:
  ├── 供应商A: 状态正常, 负载40% ✅ 选择
  ├── 供应商B: 状态正常, 负载20%
  └── 供应商C: 状态正常, 负载10%
↓
最终决策: 使用供应商A
```

### 2. 利润保护机制

**大客户场景的特殊处理**:
```
大客户X: 售价14元/M tokens
↓
计算各供应商利润:
  ├── 供应商A: 14-10=4元 (29% 利润率) ✅ 可选
  ├── 供应商B: 14-15=-1元 (亏损!) ❌ 禁用
  └── 供应商C: 14-12=2元 (14% 利润率) ✅ 可选
↓
路由决策: 仅使用供应商A和C，B被排除
```

### 3. 故障转移逻辑

**供应商故障时的转移**:
```
正常状态: 使用供应商A (成本最优)
    ↓ 故障发生
检查供应商A: 连接超时 ❌
    ↓ 立即切换
检查供应商C: 备用供应商, 成本次优 ✅
    ↓ 继续服务
使用供应商C处理请求
    ↓ 后台监控
供应商A恢复健康
    ↓ 自动加回
下次请求重新使用供应商A
```

---

## 价格体系架构

### 三层价格结构

```go
// 1. 成本价层 (Supplier Cost Price)
type SupplierCostPrice struct {
    SupplierID     string
    ModelID        string
    InputCostPerToken  float64 // 10元/M tokens
    OutputCostPerToken float64 // 30元/M tokens
}

// 2. 售价层 (User Selling Price)
type UserSellingPrice struct {
    UserGroupID    string
    ModelID        string
    InputPricePerToken  float64 // 18元/M tokens (普通)
    OutputPricePerToken float64 // 60元/M tokens (普通)
}

// 3. 利润管理层 (Profit Management)
type ProfitManagement struct {
    MinProfitMargin float64 // 最低利润率 10%
    MaxCostRatio   float64 // 最高成本占比 90%
}
```

### 实时利润计算

```go
// 路由决策时的利润验证
func ValidateProfitForRoute(sellingPrice float64, supplierCosts []float64, minMargin float64) error {
    for _, cost := range supplierCosts {
        profit := sellingPrice - cost
        margin := profit / sellingPrice

        // 检查是否满足最低利润率
        if margin < minMargin {
            return fmt.Errorf("supplier cost %.2f exceeds profit margin", cost)
        }
    }
    return nil
}
```

---

## 技术实现要点

### 1. 数据库设计

**供应商成本价表**:
```sql
CREATE TABLE supplier_cost_pricing (
    id SERIAL PRIMARY KEY,
    supplier_id VARCHAR(255),
    model_id VARCHAR(255),
    input_cost_per_token DECIMAL(10,4),
    output_cost_per_token DECIMAL(10,4),
    effective_date TIMESTAMP,
    expiry_date TIMESTAMP,
    status VARCHAR(50)
);
```

**用户群体售价表**:
```sql
CREATE TABLE user_group_pricing (
    id SERIAL PRIMARY KEY,
    user_group_id VARCHAR(255),
    model_id VARCHAR(255),
    input_price_per_token DECIMAL(10,4),
    output_price_per_token DECIMAL(10,4),
    min_profit_margin DECIMAL(5,2),
    effective_date TIMESTAMP,
    status VARCHAR(50)
);
```

### 2. 路由算法

```go
type SupplierSelector struct {
    // 获取可用供应商
    GetAvailableSuppliers(modelID string) []Supplier

    // 按利润排序
    SortByProfitMargin(suppliers []Supplier, sellingPrice float64) []Supplier

    // 检查供应商健康状态
    CheckHealth(supplier Supplier) bool

    // 检查供应商负载
    CheckLoad(supplier Supplier) bool
}
```

### 3. 实时监控

**监控指标**:
```go
type SupplierMetrics struct {
    SupplierID     string

    // 性能指标
    AverageLatency time.Duration
    SuccessRate    float64
    CurrentQPS     int64

    // 成本指标
    TotalCost      float64
    AverageCost    float64

    // 利润指标
    TotalProfit    float64
    ProfitMargin   float64
}
```

---

## 业务优势

### 1. 成本优化
- ✅ 优先选择低成本供应商
- ✅ 动态调整供应商选择
- ✅ 实时监控成本变化

### 2. 利润最大化
- ✅ 确保每个请求都有合理利润
- ✅ 不同用户群体差异化定价
- ✅ 利润空间实时计算和监控

### 3. 可靠性保障
- ✅ 多供应商冗余配置
- ✅ 自动故障转移
- ✅ 供应商健康监控

### 4. 灵活定价
- ✅ 支持不同用户群体定价
- ✅ 支持大客户独立定价
- ✅ 支持促销和优惠策略

---

## 验收标准

### 功能验收
- [ ] 支持多供应商配置
- [ ] 成本优先路由正确工作
- [ ] 利润保护机制有效
- [ ] 故障转移时间 < 1s
- [ ] 实时利润计算准确

### 业务验收
- [ ] 普通用户利润率 > 30%
- [ ] VIP会员利润率 > 20%
- [ ] 大客户利润率 > 10%
- [ ] 无亏损路由发生

### 性能验收
- [ ] 路由决策时间 < 10ms
- [ ] 支持同时配置 10+ 供应商
- [ ] 供应商故障检测时间 < 30s

---

## 总结

这个多供应商路由和成本优化机制是项目的核心竞争优势：

1. **成本优势**: 通过多供应商比价，获得最优成本
2. **利润保障**: 实时利润计算和保护机制
3. **高可用性**: 多供应商冗余和自动故障转移
4. **灵活定价**: 支持复杂的定价策略和用户分层

这确实是一个非常重要且具有商业价值的业务需求！

---

**需求确认时间**: 2026-05-08
**需求状态**: ✅ 已完全理解并更新到需求文档
**下一步**: 开始技术实现设计
