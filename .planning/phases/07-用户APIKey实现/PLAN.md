# Phase 7: 用户 API Key 实现 - 执行计划

## 基本信息

**Phase ID**: 07
**Phase 名称**: 用户 API Key 实现
**里程碑**: Milestone 2
**预计时间**: 3 天
**状态**: 📝 计划中

---

## Phase 目标

实现完整的用户 API Key 管理功能：
1. 重构认证中间件使用 KeyManager 验证
2. 实现后端 CRUD API 处理器
3. 添加使用统计和额度检查功能
4. 完善限流逻辑

---

## 验收标准

### 功能验收
- [ ] API Key 认证中间件使用 KeyManager 验证
- [ ] 支持创建、查询、更新、删除 API Key
- [ ] 支持禁用/启用 API Key
- [ ] 额度检查正常工作
- [ ] 使用统计 API 返回真实数据

### 质量验收
- [ ] Go 代码编译通过
- [ ] 集成测试覆盖核心场景
- [ ] 前后端联调通过

---

## 执行计划

### Task 7.1: 扩展 KeyManager 功能

**文件**: internal/services/key_manager.go

**新增方法**:
- `UpdateKey` - 更新 Key 配置（名称、额度）
- `GetKeyStats` - 获取使用统计
- `CheckQuota` - 检查额度限制
- `RecordUsage` - 记录使用量

**预计时间**: 3 小时

---

### Task 7.2: 扩展 UsageRecord 模型

**文件**: internal/models/models.go

**修改**:
- 在 UsageRecord 中添加 `key_id` 字段
- 添加索引支持按 Key 统计

**预计时间**: 1 小时

---

### Task 7.3: 重构认证中间件

**文件**: internal/middleware/auth.go

**修改内容**:
1. 接受 KeyManager 作为依赖
2. 使用 KeyManager.ValidateKey 验证
3. 集成额度检查
4. 返回清晰的错误信息

**预计时间**: 2 小时

---

### Task 7.4: 实现 API Key 处理器

**文件**: pkg/handlers/user_keys.go

**实现方法**:
- `ListUserKeys` - 获取列表（使用 KeyManager）
- `CreateUserKey` - 创建（使用 KeyManager，返回完整密钥）
- `GetUserKey` - 获取详情
- `UpdateUserKey` - 更新配置
- `DeleteUserKey` - 删除
- `DisableUserKey` - 禁用
- `EnableUserKey` - 启用
- `GetUserKeyStats` - 使用统计

**预计时间**: 3 小时

---

### Task 7.5: 更新路由配置

**文件**: pkg/router/router.go

**修改内容**:
1. 注入 KeyManager 到 Handler
2. 注入 KeyManager 到 Auth 中间件
3. 确保所有路由正确配置

**预计时间**: 1 小时

---

### Task 7.6: 集成测试

**文件**: tests/integration/user_apikey_test.go

**测试覆盖**:
- 创建 API Key 测试
- 验证 API Key 测试
- 额度检查测试
- 禁用/启用测试
- 使用统计测试

**预计时间**: 2 小时

---

## 依赖关系

```
7.1 (扩展 KeyManager) -> 7.3 (重构中间件)
                       -> 7.4 (实现处理器)
7.2 (扩展模型) -> 7.1 (使用新字段)
7.4 (处理器) -> 7.5 (路由配置)
7.3 (中间件) -> 7.5 (路由配置)
-> 7.6 (测试)
```

---

## Wave 分组

**Wave 1** (可并行):
- Task 7.1: 扩展 KeyManager 功能
- Task 7.2: 扩展 UsageRecord 模型

**Wave 2** (依赖 Wave 1):
- Task 7.3: 重构认证中间件
- Task 7.4: 实现 API Key 处理器

**Wave 3** (依赖 Wave 2):
- Task 7.5: 更新路由配置
- Task 7.6: 集成测试

---

## 文件变更清单

### 修改
- `internal/services/key_manager.go` - 添加新方法
- `internal/models/models.go` - 扩展 UsageRecord
- `internal/middleware/auth.go` - 重构验证逻辑
- `pkg/handlers/user_keys.go` - 实现真实处理
- `pkg/router/router.go` - 注入依赖

### 新建
- `tests/integration/user_apikey_test.go` - 集成测试

---

## 安全检查清单

- [ ] API Key 仅在创建时完整显示
- [ ] 列表和详情返回脱敏数据
- [ ] 额度超限返回 429 状态码
- [ ] 禁用 Key 被正确拒绝
- [ ] 过期 Key 被正确拒绝

---

**计划创建**: 2026-05-11
**预计完成**: 开始后 3 个工作日
