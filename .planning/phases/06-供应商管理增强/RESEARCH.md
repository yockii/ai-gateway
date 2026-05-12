# Phase 6: 供应商管理增强 - 研究文档

## 基本信息

**Phase ID**: 06
**Phase 名称**: 供应商管理增强
**研究日期**: 2026-05-11

---

## 现状分析

### 已实现功能

1. **Supplier 数据模型** (`internal/models/models.go`)
   ```go
   type Supplier struct {
       ID          string
       Name        string
       DisplayName string
       Provider    string  // openai, anthropic, etc.
       IsActive    bool
       CreatedAt   time.Time
       UpdatedAt   time.Time
   }
   ```

2. **SupplierCostPricing 数据模型**
   ```go
   type SupplierCostPricing struct {
       ID            string
       SupplierID    string
       ModelID       string
       InputCost     float64
       OutputCost    float64
       EffectiveDate time.Time
       IsActive      bool
   }
   ```

3. **Supplier Manager** (`internal/supplier/manager.go`)
   - 基础 CRUD 操作
   - 健康检查机制
   - 供应商状态管理

### 缺失功能

| 功能 | 状态 | 优先级 |
|------|------|--------|
| API Key 加密存储 | 缺失 | P0 |
| 密钥轮换策略 | 缺失 | P0 |
| 多密钥优先级管理 | 缺失 | P0 |
| 供应商模型关联 API | 缺失 | P1 |
| 使用统计追踪 | 缺失 | P1 |

---

## 技术方案设计

### 1. API Key 加密存储

**加密方案**: AES-256-GCM

```go
type SupplierApiKey struct {
    ID                string    `gorm:"primaryKey"`
    SupplierID        string    `gorm:"index"`
    Name              string    // 密钥名称，如 "生产环境主密钥"
    KeyValueEncrypted string    // AES-256 加密后的密钥
    KeyPrefix         string    // 用于显示，如 "sk-****1234"
    
    // 优先级管理
    Priority          int       // 数字越小优先级越高
    IsPrimary         bool      // 是否为主密钥
    
    // 使用限制
    MaxRequests       int64     // 最大请求数（用于密钥轮换）
    CurrentRequests   int64     // 当前请求数
    
    // 状态管理
    IsActive          bool      `gorm:"index"`
    LastUsedAt        *time.Time
    ExpireAt          *time.Time
    
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

**加密工具**:
```go
// internal/crypto/encryption.go
type EncryptionService struct {
    key []byte // 32字节密钥
}

func (e *EncryptionService) Encrypt(plaintext string) (string, error)
func (e *EncryptionService) Decrypt(ciphertext string) (string, error)
```

### 2. 密钥轮换策略

**自动轮换触发条件**:
1. 请求数达到阈值 (`MaxRequests`)
2. 密钥过期 (`ExpireAt`)
3. 手动触发

**轮换逻辑**:
```
1. 标记旧密钥为非主密钥
2. 激活备用密钥
3. 通知运维人员
```

### 3. 供应商模型关联 API

**API 端点设计**:
```
GET    /v1/admin/suppliers/:id/api-keys       # 获取供应商的 API Keys
POST   /v1/admin/suppliers/:id/api-keys       # 添加 API Key
PUT    /v1/admin/suppliers/:id/api-keys/:kid  # 更新 API Key
DELETE /v1/admin/suppliers/:id/api-keys/:kid  # 删除 API Key
PATCH  /v1/admin/suppliers/:id/api-keys/:id/set-primary  # 设为主密钥
```

### 4. 路由时密钥选择逻辑

```go
func (s *SupplierService) GetBestApiKey(supplierID string) (*SupplierApiKey, error) {
    // 1. 获取供应商所有活跃密钥
    // 2. 按优先级排序
    // 3. 选择优先级最高的主密钥
    // 4. 如果主密钥不可用，选择优先级最高的备用密钥
}
```

---

## 安全考虑

### 密钥安全
1. **加密存储**: 所有 API Key 使用 AES-256-GCM 加密
2. **密钥管理**: 加密密钥从环境变量读取
3. **日志脱敏**: 日志中只记录密钥前缀
4. **传输安全**: API 传输使用 HTTPS

### 访问控制
1. 只有管理员可以管理供应商 API Key
2. 所有操作需要审计日志
3. 支持密钥过期自动失效

---

## 依赖关系

### 内部依赖
- `internal/models` - 数据模型
- `internal/database` - 数据库访问
- `internal/crypto` - 加密服务（新建）

### 外部依赖
- `crypto/aes` - Go 标准库加密
- `crypto/rand` - 随机数生成

---

## 实现风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 加密密钥泄露 | 高 | 使用环境变量，定期轮换 |
| 密钥丢失 | 高 | 备份机制，密钥恢复流程 |
| 性能影响 | 中 | 加密结果缓存 |
| 迁移风险 | 低 | 逐步迁移，保持兼容 |

---

## 参考资料

- [NIST 加密标准](https://csrc.nist.gov/publications/detail/fips/197/final)
- [Go crypto/aes 文档](https://pkg.go.dev/crypto/aes)
- Phase 5 实现参考 (定价体系)

---

**研究完成时间**: 2026-05-11
**研究者**: GSD Phase 6 Researcher
