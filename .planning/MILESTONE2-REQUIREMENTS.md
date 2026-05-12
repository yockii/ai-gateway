# AI Gateway - Milestone 2 需求文档

## 文档信息

**文档版本**: 1.0
**创建日期**: 2026-05-11
**里程碑**: Milestone 2 - 核心功能补全
**状态**: 待批准

---

## 1. 背景与目标

### 1.1 背景

Milestone 1 完成了基础架构搭建，但经过需求与实现对比分析，发现以下核心功能存在缺失或不完整：

| 功能域 | 问题 | 影响 |
|--------|------|------|
| 定价体系 | 大客户独立定价缺失、套餐价格未应用 | 无法实现差异化定价 |
| 供应商管理 | API Key 管理缺失 | 无法管理供应商密钥 |
| 用户 API Key | 后端仅返回模拟数据 | 用户无法真正管理 API Key |
| 模型路由 | 健康检查和故障转移不完整 | 生产稳定性风险 |

### 1.2 目标

完善核心业务功能，使系统达到可生产使用状态。

---

## 2. 功能需求

### FR-M2-01: 完善定价体系

#### 2.1.1 大客户独立定价

**需求描述**: 支持为大客户设置独立的价格体系。

**功能详情**:
- 创建大客户定价记录
- 按模型设置独立售价
- 利润保护验证
- 定价生效时间管理

**数据模型**:
```go
type EnterprisePricing struct {
    ID               string    `gorm:"primaryKey"`
    CustomerID       string    `gorm:"index"`
    CustomerName     string
    ModelID          string    `gorm:"index"`
    
    // 独立定价
    InputPrice       float64
    OutputPrice      float64
    
    // 利润保护
    MinProfitMargin  float64
    MaxCostPrice     float64
    
    // 生效时间
    EffectiveDate    time.Time
    ExpiryDate       *time.Time
    
    IsActive         bool      `gorm:"index"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

**API 设计**:
```
POST   /api/v1/admin/enterprise-pricing     # 创建大客户定价
GET    /api/v1/admin/enterprise-pricing     # 获取大客户定价列表
PUT    /api/v1/admin/enterprise-pricing/:id # 更新定价
DELETE /api/v1/admin/enterprise-pricing/:id # 删除定价
```

**验收标准**:
- [ ] 支持为大客户设置独立价格
- [ ] 售价生效时验证利润率
- [ ] 支持定价生效时间管理
- [ ] 提供运维管理界面

**优先级**: P0

---

#### 2.1.2 套餐价格应用逻辑

**需求描述**: 实现会员折扣在实际计费中的应用。

**功能详情**:
- 获取用户会员等级
- 应用对应折扣率
- 计算最终售价
- 折扣记录和审计

**实现逻辑**:
```go
func (s *MembershipService) CalculatePrice(
    ctx context.Context, 
    userID string, 
    modelID string, 
    basePrice float64,
) (float64, error) {
    // 1. 获取用户会员等级
    membership, err := s.GetUserMembership(ctx, userID)
    if err != nil {
        return basePrice, nil // 无会员，使用原价
    }
    
    // 2. 获取折扣率
    discount, err := s.GetDiscount(ctx, membership.TierID, modelID)
    if err != nil || !discount.IsActive {
        return basePrice, nil
    }
    
    // 3. 应用折扣
    finalPrice := basePrice * (1 - discount.DiscountRate)
    
    // 4. 记录折扣应用
    s.logDiscountApplication(userID, modelID, discount.DiscountRate)
    
    return finalPrice, nil
}
```

**验收标准**:
- [ ] 会员价格正确计算
- [ ] 折扣记录可追溯
- [ ] 支持折扣实时生效

**优先级**: P0

---

### FR-M2-02: 供应商 API Key 管理

#### 2.2.1 供应商密钥存储

**需求描述**: 安全存储和管理供应商的 API Key。

**功能详情**:
- API Key 加密存储
- 密钥轮换策略
- 优先级管理
- 使用统计

**数据模型**:
```go
type SupplierApiKey struct {
    ID           string    `gorm:"primaryKey"`
    SupplierID   string    `gorm:"index"`
    Name         string
    KeyValueEncrypted string // 加密存储
    KeyPrefix    string    // 用于显示，如 sk-****...
    
    // 轮换配置
    Priority     int       // 优先级，数字越小优先级越高
    IsPrimary    bool      // 是否为主密钥
    
    // 使用限制
    MaxRequests  int64     // 最大请求数（用于密钥轮换）
    CurrentRequests int64  // 当前请求数
    
    // 状态管理
    IsActive     bool      `gorm:"index"`
    LastUsedAt   *time.Time
    ExpireAt     *time.Time
    
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**验收标准**:
- [ ] 密钥使用 AES-256 加密存储
- [ ] 支持密钥轮换
- [ ] 优先级路由正常工作
- [ ] 提供运维管理界面

**优先级**: P0

---

#### 2.2.2 供应商模型关联管理

**需求描述**: 管理供应商与模型之间的关联关系。

**功能详情**:
- 配置供应商支持的模型
- 设置模型成本价格
- 批量导入价格
- 价格变更历史

**API 设计**:
```
POST   /api/v1/admin/suppliers/:id/models        # 添加供应商模型
GET    /api/v1/admin/suppliers/:id/models        # 获取供应商模型列表
PUT    /api/v1/admin/suppliers/:id/models/:mid   # 更新模型价格
DELETE /api/v1/admin/suppliers/:id/models/:mid   # 删除模型关联
```

**验收标准**:
- [ ] 支持配置供应商模型
- [ ] 价格变更可追溯
- [ ] 提供管理界面

**优先级**: P1

---

### FR-M2-03: 用户 API Key 管理

#### 2.3.1 后端 CRUD 实现

**需求描述**: 实现用户 API Key 的完整后端功能。

**功能详情**:
- 创建 API Key（带安全生成）
- 列表查询（带分页）
- 更新配置
- 删除/禁用/启用
- 使用统计

**API 设计**:
```
GET    /api/v1/user/keys                    # 获取列表
POST   /api/v1/user/keys                    # 创建新 Key
GET    /api/v1/user/keys/:id                # 获取详情
PUT    /api/v1/user/keys/:id                # 更新配置
DELETE /api/v1/user/keys/:id                # 删除
PATCH  /api/v1/user/keys/:id/disable        # 禁用
PATCH  /api/v1/user/keys/:id/enable         # 启用
GET    /api/v1/user/keys/:id/stats          # 使用统计
```

**验收标准**:
- [ ] 所有 API 正常工作
- [ ] API Key 安全生成（sk-前缀 + 随机字符串）
- [ ] 密钥仅在创建时显示一次
- [ ] 支持限额配置（日额度、月额度、并发限制）

**优先级**: P0

---

#### 2.3.2 API Key 认证中间件

**需求描述**: 实现基于 API Key 的认证中间件。

**功能详情**:
- 解析 Bearer Token
- 验证 API Key 有效性
- 检查禁用状态
- 检查过期时间
- 检查额度限制
- 提取用户信息到上下文

**验收标准**:
- [ ] 认证正确有效
- [ ] 禁用 Key 被拒绝
- [ ] 过期 Key 被拒绝
- [ ] 超额 Key 被限流

**优先级**: P0

---

### FR-M2-04: 模型路由增强

#### 2.4.1 健康检查实现

**需求描述**: 实现供应商的实际健康检查。

**功能详情**:
- 主动探测供应商健康状态
- 检查 API 可用性
- 测量响应延迟
- 错误率统计

**实现逻辑**:
```go
func (m *Manager) PerformHealthCheck(ctx context.Context, supplierID string) (*HealthCheckResult, error) {
    // 1. 获取供应商主 API Key
    apiKey, err := m.getPrimaryApiKey(ctx, supplierID)
    if err != nil {
        return &HealthCheckResult{IsHealthy: false, Error: err.Error()}, nil
    }
    
    // 2. 发送探测请求
    startTime := time.Now()
    resp, err := m.sendProbeRequest(ctx, supplierID, apiKey)
    latency := time.Since(startTime)
    
    // 3. 评估结果
    result := &HealthCheckResult{
        SupplierID: supplierID,
        IsHealthy:  err == nil && resp.StatusCode < 500,
        Latency:    latency,
        Error:      getErrorString(err),
    }
    
    // 4. 更新状态
    m.updateHealthStatus(result)
    
    return result, nil
}
```

**验收标准**:
- [ ] 健康检查定期执行
- [ ] 不健康供应商被自动排除
- [ ] 健康恢复后自动加回
- [ ] 提供健康状态查询 API

**优先级**: P0

---

#### 2.4.2 故障转移

**需求描述**: 实现供应商故障时的自动切换。

**功能详情**:
- 检测请求失败
- 自动切换到备用供应商
- 记录切换事件
- 供应商恢复后自动加回

**验收标准**:
- [ ] 故障检测时间 < 5 秒
- [ ] 切换时间 < 1 秒
- [ ] 切换事件可追溯
- [ ] 支持手动切换

**优先级**: P0

---

### FR-M2-05: 运维管理界面补全

#### 2.5.1 供应商管理界面

**需求描述**: 完善供应商管理的前端界面。

**功能详情**:
- 供应商列表
- API Key 管理
- 模型关联管理
- 健康状态展示

**验收标准**:
- [ ] 界面美观易用
- [ ] 支持所有管理操作
- [ ] 实时显示健康状态

**优先级**: P1

---

#### 2.5.2 定价管理界面

**需求描述**: 提供定价配置的管理界面。

**功能详情**:
- 用户组定价管理
- 大客户定价管理
- 会员折扣配置
- 价格历史查看

**验收标准**:
- [ ] 支持所有定价配置
- [ ] 价格变更可审计
- [ ] 界面操作便捷

**优先级**: P1

---

## 3. 非功能性需求

### NFR-M2-01: 安全性

- API Key 必须加密存储
- 密钥仅在创建时完整显示
- 支持密钥轮换
- 敏感操作记录审计日志

### NFR-M2-02: 性能

- API Key 验证 < 5ms
- 价格计算 < 1ms
- 路由决策 < 10ms
- 健康检查不影响主流程

### NFR-M2-03: 可靠性

- 故障自动转移
- 密钥轮换不中断服务
- 定价变更实时生效

---

## 4. 优先级矩阵

| 需求 | 优先级 | 预计工作量 |
|------|--------|-----------|
| FR-M2-01.1 大客户定价 | P0 | 3 天 |
| FR-M2-01.2 套餐价格应用 | P0 | 2 天 |
| FR-M2-02.1 供应商密钥存储 | P0 | 3 天 |
| FR-M2-02.2 供应商模型关联 | P1 | 2 天 |
| FR-M2-03.1 用户 API Key CRUD | P0 | 3 天 |
| FR-M2-03.2 API Key 认证中间件 | P0 | 2 天 |
| FR-M2-04.1 健康检查实现 | P0 | 2 天 |
| FR-M2-04.2 故障转移 | P0 | 2 天 |
| FR-M2-05.1 供应商管理界面 | P1 | 2 天 |
| FR-M2-05.2 定价管理界面 | P1 | 2 天 |

**总计**: 约 23 工作日

---

## 5. 里程碑验收标准

### 功能验收

- [ ] 所有 P0 功能完整实现
- [ ] 所有 P0 功能测试通过
- [ ] 前端管理界面可用

### 质量验收

- [ ] 代码审查通过
- [ ] 单元测试覆盖率 > 80%
- [ ] 安全扫描无高危漏洞

### 文档验收

- [ ] API 文档更新
- [ ] 数据库变更文档
- [ ] 部署指南更新

---

**文档维护**: 本文档随 Milestone 2 进展持续更新
**最后更新**: 2026-05-11
