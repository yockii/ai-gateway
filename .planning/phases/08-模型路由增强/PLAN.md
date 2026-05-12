# Phase 8: 模型路由增强 - 执行计划

## 基本信息

**Phase ID**: 08
**Phase 名称**: 模型路由增强
**里程碑**: Milestone 2
**预计时间**: 3 天
**状态**: 📝 计划中

---

## Phase 目标

实现供应商健康检查和故障转移机制：
1. 实现真实的供应商健康检查
2. 实现自动故障转移
3. 记录故障转移事件
4. 提供健康状态查询 API

---

## 验收标准

### 功能验收
- [ ] 健康检查定期执行
- [ ] 不健康供应商被自动排除
- [ ] 健康恢复后自动加回
- [ ] 提供健康状态查询 API
- [ ] 故障检测时间 < 5 秒
- [ ] 切换时间 < 1 秒
- [ ] 切换事件可追溯
- [ ] 支持手动切换

### 质量验收
- [ ] Go 代码编译通过
- [ ] 集成测试覆盖核心场景
- [ ] 性能测试通过

---

## 执行计划

### Task 8.1: 创建失败事件数据模型

**文件**: internal/models/supplier_event.go

**新增模型**:
- `SupplierFailureEvent` - 供应商失败事件
- `FailoverEvent` - 故障转移事件

**字段**:
- ID, SupplierID, ModelID, ErrorType, ErrorMsg, Timestamp
- FromSupplierID, ToSupplierID, Reason

**预计时间**: 1 小时

---

### Task 8.2: 实现真实健康检查

**文件**: internal/supplier/health_checker.go

**新增方法**:
- `PerformRealHealthCheck()` - 真实 API 探测
- `checkOpenAI()` - OpenAI 健康检查
- `checkAnthropic()` - Anthropic 健康检查
- `checkGeneric()` - 通用健康检查

**改进**:
- 使用真实 HTTP 请求
- 记录实际响应时间
- 返回详细错误信息

**预计时间**: 3 小时

---

### Task 8.3: 增强路由决策逻辑

**文件**: internal/services/model_router.go

**修改内容**:
1. `SelectBestSupplier` 添加健康状态过滤
2. 实现故障转移逻辑
3. 添加重试机制
4. 记录故障转移事件

**预计时间**: 3 小时

---

### Task 8.4: 健康状态查询 API

**文件**: pkg/handlers/health.go

**新增端点**:
- GET /v1/admin/suppliers/health - 获取所有供应商健康状态
- GET /v1/admin/suppliers/:id/health - 获取单个供应商健康状态
- POST /v1/admin/suppliers/:id/health/check - 手动触发健康检查
- GET /v1/admin/suppliers/events/failures - 获取失败事件
- GET /v1/admin/suppliers/events/failovers - 获取故障转移事件

**预计时间**: 2 小时

---

### Task 8.5: 更新路由配置

**文件**: pkg/router/router.go

**修改内容**:
1. 注入 SupplierManager
2. 添加健康状态路由

**预计时间**: 1 小时

---

### Task 8.6: 集成测试

**文件**: tests/integration/supplier/health_test.go

**测试覆盖**:
- 健康检查测试
- 故障转移测试
- 事件记录测试
- 路由决策测试

**预计时间**: 2 小时

---

## 依赖关系

```
8.1 (事件模型) -> 8.3 (路由决策)
               -> 8.4 (查询 API)
8.2 (健康检查) -> 8.3 (路由决策)
8.3 (路由决策) -> 8.6 (测试)
8.4 (查询 API) -> 8.5 (路由配置)
-> 8.6 (测试)
```

---

## Wave 分组

**Wave 1** (可并行):
- Task 8.1: 创建失败事件数据模型
- Task 8.2: 实现真实健康检查

**Wave 2** (依赖 Wave 1):
- Task 8.3: 增强路由决策逻辑
- Task 8.4: 健康状态查询 API

**Wave 3** (依赖 Wave 2):
- Task 8.5: 更新路由配置
- Task 8.6: 集成测试

---

## 文件变更清单

### 新建
- `internal/models/supplier_event.go` - 事件模型
- `internal/supplier/health_checker.go` - 健康检查器
- `pkg/handlers/health.go` - 健康状态 API
- `tests/integration/supplier/health_test.go` - 集成测试

### 修改
- `internal/supplier/manager.go` - 集成真实健康检查
- `internal/services/model_router.go` - 增强路由决策
- `pkg/router/router.go` - 添加健康路由
- `internal/database/database.go` - 添加事件模型迁移

---

## 安全检查清单

- [ ] 健康检查不暴露敏感信息
- [ ] 故障转移事件记录完整
- [ ] API 需要管理员权限
- [ ] 健康检查不影响主请求性能

---

**计划创建**: 2026-05-11
**预计完成**: 开始后 3 个工作日
