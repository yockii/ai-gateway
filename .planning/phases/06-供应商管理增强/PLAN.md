# Phase 6: 供应商管理增强 - 执行计划

## 基本信息

**Phase ID**: 06
**Phase 名称**: 供应商管理增强
**里程碑**: Milestone 2
**预计时间**: 5 天
**状态**: 📝 计划中 (v2 - 修复验证问题)

---

## Phase 目标

完善供应商管理系统，实现：
1. 供应商 API Key 安全存储和管理
2. 密钥轮换策略
3. 多密钥优先级路由
4. 供应商模型关联管理 (FR-M2-02.2)
5. 供应商管理运维界面

---

## 验收标准

### 功能验收
- [ ] API Key 加密存储 (AES-256-GCM)
- [ ] 支持多密钥管理
- [ ] 密钥轮换自动触发
- [ ] 优先级路由正常工作
- [ ] 供应商模型关联可配置
- [ ] 模型成本价格可管理
- [ ] 运维管理界面完整

### 质量验收
- [ ] 单元测试覆盖率 > 80%
- [ ] 加密功能验证通过
- [ ] 代码审查通过

---

## 执行计划

### Task 6.1: 创建加密服务

**文件**: internal/crypto/encryption.go

**步骤**:
1. 创建 EncryptionService 结构体
2. 实现 Encrypt 方法 (AES-256-GCM)
3. 实现 Decrypt 方法
4. 从环境变量读取加密密钥
5. 添加单元测试

**预计时间**: 3 小时

---

### Task 6.2: 创建 SupplierApiKey 数据模型

**文件**: internal/models/supplier_apikey.go

**数据模型包含字段**:
- ID, SupplierID, Name
- KeyValueEncrypted (加密后的密钥)
- KeyPrefix (显示用，如 sk-****1234)
- Priority, IsPrimary
- MaxRequests, CurrentRequests (轮换用)
- IsActive, LastUsedAt, ExpireAt
- CreatedAt, UpdatedAt

**预计时间**: 2 小时

---

### Task 6.3: 实现 SupplierApiKeyService

**文件**: internal/services/supplier_apikey.go

**方法**:
- CreateApiKey - 创建并加密
- ListApiKeys - 列表查询
- GetBestApiKey - 按优先级获取
- GetApiKeyStats - 使用统计 (新增)
- SetPrimaryApiKey - 设为主密钥
- RotateApiKey - 密钥轮换
- IncrementUsage - 增加计数
- ShouldRotate - 判断是否轮换

**预计时间**: 5 小时

---

### Task 6.4: 创建 API Key 处理器

**文件**: pkg/handlers/supplier_apikey.go

**API 端点**:
- GET /v1/admin/suppliers/:id/api-keys
- POST /v1/admin/suppliers/:id/api-keys
- PUT /v1/admin/suppliers/:id/api-keys/:kid
- DELETE /v1/admin/suppliers/:id/api-keys/:kid
- PATCH /v1/admin/suppliers/:id/api-keys/:id/set-primary
- POST /v1/admin/suppliers/:id/api-keys/:id/rotate
- GET /v1/admin/suppliers/:id/api-keys/:kid/stats (新增)

**预计时间**: 3 小时

---

### Task 6.5: 更新路由配置

**文件**: pkg/router/router.go

**预计时间**: 1 小时

---

### Task 6.6: 创建供应商模型关联服务 (FR-M2-02.2) - 新增

**文件**: internal/services/supplier_model.go

**数据模型**: SupplierModel
- ID, SupplierID, ModelID
- InputCost, OutputCost
- IsActive, EffectiveDate

**方法**:
- AddModelToSupplier
- ListSupplierModels
- UpdateModelCost
- RemoveModelFromSupplier
- GetPriceHistory

**预计时间**: 4 小时

---

### Task 6.7: 创建供应商模型关联 API - 新增

**文件**: pkg/handlers/supplier_model.go

**API 端点**:
- GET /v1/admin/suppliers/:id/models
- POST /v1/admin/suppliers/:id/models
- PUT /v1/admin/suppliers/:id/models/:mid
- DELETE /v1/admin/suppliers/:id/models/:mid
- GET /v1/admin/suppliers/:id/models/:mid/history

**预计时间**: 2 小时

---

### Task 6.8: 更新路由添加模型关联 - 新增

**文件**: pkg/router/router.go

**预计时间**: 1 小时

---

### Task 6.9: 开发运维端供应商管理界面

**文件**: frontend/admin/src/views/SupplierManagement.vue

**功能**:
- 供应商基本信息管理
- API Key 列表展示和操作
- 模型关联管理
- 成本价格设置

**预计时间**: 6 小时

---

### Task 6.10: 集成测试

**文件**: tests/integration/supplier/supplier_test.go

**预计时间**: 3 小时

---

## 依赖关系

```
6.1 (加密服务) -> 6.2 (数据模型) <-> 6.3 (服务) -> 6.4 (处理器) -> 6.5 (路由)
                                                               ↓
                                    6.6 (模型服务) <-> 6.7 (模型API) -> 6.8 (路由)
                                                               ↓
                                                         6.9 (前端) -> 6.10 (测试)
```

---

## 文件变更清单

### 新建
- internal/crypto/encryption.go
- internal/models/supplier_apikey.go
- internal/models/supplier_model.go
- internal/services/supplier_apikey.go
- internal/services/supplier_model.go
- pkg/handlers/supplier_apikey.go
- pkg/handlers/supplier_model.go
- frontend/admin/src/views/SupplierManagement.vue
- tests/integration/supplier/supplier_test.go

### 修改
- internal/database/database.go
- pkg/router/router.go
- pkg/handlers/chat.go
- frontend/admin/src/router/index.ts

---

## 安全检查清单

- [ ] AES-256-GCM 加密存储
- [ ] 环境变量读取密钥
- [ ] API 响应脱敏
- [ ] 日志不记录明文
- [ ] 管理操作需认证
- [ ] 支持密钥过期

---

**计划创建**: 2026-05-11
**更新时间**: 2026-05-11 (修复验证问题)
**预计完成**: 开始后 5 个工作日
