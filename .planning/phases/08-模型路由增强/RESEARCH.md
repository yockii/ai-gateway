# Phase 8: 模型路由增强 - 技术研究

**日期**: 2026-05-11
**Phase**: 08 - 模型路由增强
**状态**: 研究完成

---

## 1. 现有实现分析

### 1.1 供应商管理器 ✅ 基础框架已实现

**文件**: `internal/supplier/manager.go`

**已有组件**:
- `HealthStatus` 结构体 - 健康状态跟踪
- `HealthCheckResult` 结构体 - 检查结果
- `PerformHealthCheck()` - 健康检查方法（TODO: 简化实现）
- `IsHealthy()` - 检查是否健康
- `GetHealthySuppliers()` - 获取健康供应商列表
- `backgroundHealthCheck()` - 后台定期检查
- 失败阈值机制

**问题**:
- `PerformHealthCheck()` 是模拟实现（10ms sleep）
- 没有实际的 API 探测
- 没有错误率统计
- 没有响应时间历史

### 1.2 模型路由器 ✅ 基础路由已实现

**文件**: `internal/services/model_router.go`

**已有组件**:
- `SelectBestSupplier()` - 基于利润选择最优供应商
- `getAvailableRoutes()` - 获取可用路由
- `evaluateSupplier()` - 评估供应商
- `CheckSupplierHealth()` - 健康检查（TODO）
- `IsOverloaded()` - 过载检查（TODO）

**问题**:
- 路由选择没有考虑健康状态
- 没有故障转移逻辑
- 没有失败重试机制

### 1.3 数据模型 ✅ 已有基础

**已有模型**:
- `Supplier` - 供应商信息
- `ModelMapping` - 模型映射
- `SupplierCostPricing` - 供应商成本定价

---

## 2. 需求分析 (FR-M2-04)

### FR-M2-04.1: 健康检查实现

**需求状态**: ⚠️ 框架存在，需要完善

| 功能 | 现状 | 需要动作 |
|------|------|----------|
| 主动探测供应商健康 | ❌ 模拟实现 | 实现真实 API 调用 |
| 检查 API 可用性 | ❌ 未实现 | 发送测试请求 |
| 测量响应延迟 | ⚠️ 模拟值 | 记录真实延迟 |
| 错误率统计 | ❌ 未实现 | 添加错误计数 |
| 健康状态查询 API | ❌ 未实现 | 添加 Handler |

### FR-M2-04.2: 故障转移

**需求状态**: ❌ 未实现

| 功能 | 现状 | 需要动作 |
|------|------|----------|
| 检测请求失败 | ❌ 未实现 | 添加失败检测 |
| 自动切换备用供应商 | ❌ 未实现 | 修改路由逻辑 |
| 记录切换事件 | ❌ 未实现 | 添加事件日志 |
| 恢复后自动加回 | ⚠️ 部分实现 | 后台检查已有 |
| 手动切换 | ❌ 未实现 | 添加管理 API |

---

## 3. 技术方案

### 3.1 健康检查实现方案

**供应商健康检查器**:
```go
type HealthChecker struct {
    httpClient *http.Client
    manager    *Manager
}

func (hc *HealthChecker) Check(ctx context.Context, supplier *Supplier) (*HealthCheckResult, error) {
    // 1. 获取供应商主 API Key
    apiKey := hc.getPrimaryApiKey(supplier.ID)
    
    // 2. 发送轻量级探测请求
    // OpenAI: GET /models
    // Anthropic: POST /messages (minimal)
    
    // 3. 测量响应时间
    // 4. 检查返回状态
    // 5. 返回结果
}
```

### 3.2 故障转移实现方案

**路由决策增强**:
```go
func (mr *ModelRouter) SelectBestSupplierWithFailover(ctx, userID, modelID) (*SupplierSelection, error) {
    routes := mr.getAvailableRoutes(ctx, modelID)
    
    // 按优先级排序，过滤掉不健康的
    healthyRoutes := filterHealthy(routes)
    
    for _, route := range healthyRoutes {
        selection := mr.evaluateSupplier(route)
        if selection != nil {
            return selection, nil
        }
    }
    
    return nil, ErrNoHealthySupplier
}
```

**失败记录表**:
```go
type SupplierFailureEvent struct {
    ID          string
    SupplierID  string
    ModelID     string
    ErrorType   string
    ErrorMsg    string
    Timestamp   time.Time
}
```

### 3.3 事件追踪

**故障转移事件表**:
```go
type FailoverEvent struct {
    ID              string
    FromSupplierID  string
    ToSupplierID    string
    ModelID         string
    Reason          string
    Timestamp       time.Time
}
```

---

## 4. 待完成任务

### Task 8.1: 实现真实健康检查
- 创建 HealthChecker 服务
- 实现各供应商的探测逻辑
- 记录响应时间和错误率

### Task 8.2: 添加健康状态 API
- 获取供应商健康状态
- 获取健康检查历史
- 手动触发健康检查

### Task 8.3: 实现故障转移路由
- 修改 SelectBestSupplier 考虑健康状态
- 添加自动切换逻辑
- 实现重试机制

### Task 8.4: 事件记录和查询
- 创建失败事件模型
- 创建故障转移事件模型
- 实现 API 查询接口

### Task 8.5: 集成和测试
- 集成健康检查到路由
- 故障转移测试
- 性能验证

---

## 5. 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 健康检查增加延迟 | 中 | 使用异步检查，缓存结果 |
| 频繁探测消耗配额 | 低 | 使用轻量级端点，控制频率 |
| 故障转移误判 | 中 | 设置合理阈值，人工确认 |

---

**研究完成时间**: 2026-05-11
