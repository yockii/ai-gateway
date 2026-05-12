# Milestone 2 - 集成审计报告

**审计日期**: 2026-05-12
**审计范围**: Phase 5-9 跨阶段集成

---

## 执行摘要

| 状态 | 数量 |
|------|------|
| ✅ 正常连接 | 7 |
| ⚠️ 部分连接 | 3 |
| ❌ 断开连接 | 3 |

---

## 关键问题 (BLOCKER)

### 1. 定价服务与网关未集成

**问题**: `PricingService.GetUserPrice()` 已创建，但 `Gateway.ChatCompletion()` 不使用它
- **位置**: `internal/gateway/gateway.go:182-198`
- **影响**: FR-M2-01.2 (套餐价格应用逻辑) 无法工作
- **修复**: 在 Gateway 中集成 PricingService

### 2. 健康检查未使用真实 API Key

**问题**: `HealthChecker` 接受 `getAPIKey` 回调，但 `Manager.PerformHealthCheck()` 不调用 `GetBestApiKey()`
- **位置**: `internal/supplier/manager.go:209-250`
- **影响**: FR-M2-04.1 (健康检查实现) 无法进行真实 API 检查
- **修复**: 在 Manager 中集成 SupplierApiKeyService

### 3. 网关未使用 ModelRouter

**问题**: `ModelRouter.SelectBestSupplier()` 完整实现，但 `Gateway.SelectBestRoute()` 返回硬编码数据
- **位置**: `internal/gateway/gateway.go`
- **影响**: FR-M2-04.2 (故障转移) 无法工作
- **修复**: 在 Gateway 中集成 ModelRouter

---

## 警告问题 (WARNING)

### 4. 使用量记录未实现

**问题**: `KeyManager.RecordUsage()` 是存根实现
- **影响**: API Key 统计功能不完整

### 5. 前端使用轮询而非 SSE

**问题**: `HealthIndicator.vue` 使用 setInterval 而非 EventSource
- **影响**: 实时性较差

---

## 集成映射

| 需求 | 集成路径 | 状态 |
|------|----------|------|
| FR-M2-01.1 | 定价管理界面 → 后端 API | ✅ |
| FR-M2-01.2 | PricingService → Gateway | ❌ |
| FR-M2-02.1 | 供应商密钥界面 → 后端 API | ✅ |
| FR-M2-02.2 | 供应商模型界面 → 后端 API | ✅ |
| FR-M2-03.1 | 用户 API Key API | ✅ |
| FR-M2-03.2 | 认证中间件 → KeyManager | ✅ |
| FR-M2-04.1 | HealthChecker → API Key 服务 | ❌ |
| FR-M2-04.2 | Gateway → ModelRouter | ❌ |
| FR-M2-05.1 | 供应商管理界面 | ✅ |
| FR-M2-05.2 | 定价管理界面 | ✅ |
| 审计日志 | 后端处理器 → AuditService | ✅ |
| SSE 推送 | health.go → 前端 | ⚠️ |

---

## 修复建议

### 优先级 1 (必须修复)

1. **集成 ModelRouter 到 Gateway**
   - 在 router.go 中初始化 ModelRouter
   - 注入到 Gateway
   - 更新 SelectBestRoute()

2. **集成 SupplierApiKeyService 到 HealthChecker**
   - 添加服务字段到 Manager
   - 更新 PerformHealthCheck()

3. **集成 PricingService 到 Gateway**
   - 添加价格计算逻辑
   - 应用用户特定价格

### 优先级 2 (应该修复)

4. 实现使用量记录
5. 前端使用 EventSource

---

**审计人**: gsd-integration-checker
