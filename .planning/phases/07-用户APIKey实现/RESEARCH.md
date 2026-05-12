# Phase 7: 用户 API Key 实现 - 技术研究

**日期**: 2026-05-11
**Phase**: 07 - 用户 API Key 实现
**状态**: 研究完成

---

## 1. 现有实现分析

### 1.1 数据模型 ✅ 已完成

**文件**: `internal/models/models.go`

```go
type UserAPIKey struct {
    ID                string         `json:"id" gorm:"primaryKey"`
    UserID            string         `json:"user_id" gorm:"index"`
    KeyValue          string         `json:"-" gorm:"uniqueIndex"` // 不在 JSON 中显示
    Name              string         `json:"name"`
    QuotaDaily        int64          `json:"quota_daily"`
    QuotaMonthly      int64          `json:"quota_monthly"`
    ConcurrencyLimit  int64          `json:"concurrency_limit"`
    ModelConcurrency  map[string]int64 `json:"model_concurrency" gorm:"serializer:json"`
    LastCalculatedAt  *time.Time     `json:"last_calculated_at"`
    ExpiresAt         time.Time      `json:"expires_at"`
    IsActive          bool           `json:"is_active" gorm:"index"`
    CreatedAt         time.Time      `json:"created_at"`
    UpdatedAt         time.Time      `json:"updated_at"`
}
```

**状态**: 模型完整，支持所有需求字段。

### 1.2 KeyManager 服务 ✅ 已完成

**文件**: `internal/services/key_manager.go`

**已实现方法**:
- `CreateKey` - 创建 API Key (UUID v4 with sk- prefix)
- `ValidateKey` - 验证 API Key 并返回用户信息
- `GetUserKeys` - 获取用户所有 Key (脱敏)
- `UpdateKeyStatus` - 更新 Key 状态
- `DeleteKey` - 删除 Key

**状态**: 核心功能完整。

### 1.3 认证中间件 ⚠️ 待完善

**文件**: `internal/middleware/auth.go`

**问题**:
- `extractUserIDFromAPIKey` 是临时实现，未使用 KeyManager
- 没有真正的数据库验证
- 没有额度检查

**需要实现**:
1. 使用 KeyManager.ValidateKey 进行验证
2. 添加额度检查逻辑
3. 添加过期检查

### 1.4 API 处理器 ⚠️ 返回模拟数据

**文件**: `pkg/handlers/user_keys.go`

**问题**:
- 所有方法返回 TODO 模拟数据
- 没有连接到 KeyManager

**需要实现**:
1. 注入 KeyManager
2. 实现真实的 CRUD 操作
3. 添加使用统计功能

### 1.5 前端 ✅ 已完成

**文件**: `frontend/user/src/views/ApiKeys.vue`

**状态**: 完整实现，包括:
- 列表展示
- 创建对话框（密钥仅显示一次）
- 禁用/启用切换
- 删除确认

---

## 2. 需求分析 (FR-M2-03)

### FR-M2-03.1: 后端 CRUD 实现

**需求状态**: ⚠️ 部分完成

| 功能 | 现状 | 需要动作 |
|------|------|----------|
| 创建 API Key | KeyManager 已实现 | 连接到 Handler |
| 列表查询 | KeyManager 已实现 | 连接到 Handler |
| 更新配置 | ❌ 未实现 | 添加到 KeyManager |
| 删除 | KeyManager 已实现 | 连接到 Handler |
| 禁用/启用 | KeyManager 已实现 | 连接到 Handler |
| 使用统计 | ❌ 未实现 | 新增功能 |

### FR-M2-03.2: API Key 认证中间件

**需求状态**: ⚠️ 部分完成

| 功能 | 现状 | 需要动作 |
|------|------|----------|
| Bearer Token 解析 | ✅ 已实现 | 无 |
| 验证有效性 | ⚠️ 临时实现 | 使用 KeyManager |
| 检查禁用状态 | KeyManager 支持 | 集成到中间件 |
| 检查过期时间 | KeyManager 支持 | 集成到中间件 |
| 检查额度限制 | ❌ 未实现 | 新增功能 |

---

## 3. 技术方案

### 3.1 认证中间件重构

需要修改 `internal/middleware/auth.go`，使用 KeyManager 进行真正的验证。

### 3.2 使用统计实现

需要在 `UsageRecord` 表中添加 `key_id` 字段关联。

### 3.3 额度检查实现

需要在 KeyManager 中添加 CheckQuota 方法。

---

## 4. 待完成任务

### Task 7.1: 重构认证中间件
### Task 7.2: 实现 API Key 处理器
### Task 7.3: 添加使用统计功能
### Task 7.4: 添加更新配置功能
### Task 7.5: 路由和服务初始化
### Task 7.6: 测试和验证

---

**研究完成时间**: 2026-05-11
