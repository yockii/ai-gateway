# AI Gateway 项目需求文档

## 项目概述

**项目名称**: AI Gateway Platform
**项目版本**: v1.0
**文档版本**: 2.0
**创建日期**: 2026-05-08
**最后更新**: 2026-05-08
**技术方案**: 基于 Bifrost 深度二次开发

---

## 业务背景

AI Gateway 项目旨在构建一个统一的大模型调用网关平台，为开发者和企业提供统一、高效、经济的 AI 模型访问服务。平台通过统一接口屏蔽不同 AI 提供商的复杂性，提供灵活的定价策略和完善的管理体系。

---

## 系统架构

### 服务架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端层                              │
├─────────────────────────────────────────────────────────────┤
│  用户端 (Vue)  │  运维管理端 (Vue)  │  API 客户端            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      API 网关层 (独立)                       │
├─────────────────────────────────────────────────────────────┤
│  统一 API 入口  │  协议转换  │  负载均衡  │  水平扩展        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                     业务逻辑层 (微服务)                      │
├─────────────────────────────────────────────────────────────┤
│  用户服务  │  模型服务  │  计费服务  │  管理服务             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   Bifrost 核心引擎层                         │
├─────────────────────────────────────────────────────────────┤
│  模型路由  │  供应商管理  │  故障转移  │  性能监控           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      AI 提供商层                             │
├─────────────────────────────────────────────────────────────┤
│  OpenAI  │  Claude  │  Qwen  │  其他模型                    │
└─────────────────────────────────────────────────────────────┘
```

### 服务分离原则

1. **API 网关服务**: 独立部署，专门处理高并发 API 调用
2. **用户端应用**: 独立前端应用，面向普通用户
3. **运维管理端**: 独立前端应用，面向运维管理员
4. **业务服务**: 按功能拆分的微服务

---

## 核心功能需求

### 1. 统一模型接口 (FR-001)

**需求描述**: 提供 OpenAI 兼容的统一 API 接口，支持多种模型类型

#### 1.1 文本模型接口

**功能详情**:
- **Chat Completions**: `/v1/chat/completions`
- **Completions**: `/v1/completions`
- 支持流式和非流式响应
- 支持多轮对话
- 支持 Function Calling

**OpenAI 兼容性**:
```json
{
  "model": "gpt-4",
  "messages": [{"role": "user", "content": "Hello"}],
  "stream": true,
  "temperature": 0.7
}
```

#### 1.2 图片生成接口

**功能详情**:
- **Images Generation**: `/v1/images/generations`
- 支持多种分辨率
- 支持不同质量级别
- 支持图片编辑和变体

**OpenAI 兼容性**:
```json
{
  "model": "dall-e-3",
  "prompt": "A beautiful sunset",
  "n": 1,
  "size": "1024x1024",
  "quality": "hd"
}
```

#### 1.3 视频生成接口

**功能详情**:
- **Video Generation**: `/v1/videos/generations`
- 支持按秒计费
- 支持不同分辨率和时长
- 支持异步生成

**接口格式**:
```json
{
  "model": "sora",
  "prompt": "A cat playing piano",
  "duration": 10,
  "resolution": "1080p"
}
```

#### 1.4 语音合成接口 (TTS)

**功能详情**:
- **Audio Speech**: `/v1/audio/speech`
- 支持多种声音
- 支持不同语速和音调
- 支持流式输出

**OpenAI 兼容性**:
```json
{
  "model": "tts-1",
  "input": "Hello world",
  "voice": "alloy"
}
```

#### 1.5 语音识别接口 (STT)

**功能详情**:
- **Audio Transcription**: `/v1/audio/transcriptions`
- **Audio Translation**: `/v1/audio/translations`
- 支持多语言识别
- 支持时间戳输出

#### 1.6 Embedding 接口

**功能详情**:
- **Embeddings**: `/v1/embeddings`
- 支持不同维度的向量
- 支持批量嵌入
- 支持多种嵌入模型

**OpenAI 兼容性**:
```json
{
  "model": "text-embedding-3-small",
  "input": "Hello world",
  "dimensions": 1536
}
```

#### 1.7 Reranking 接口

**功能详情**:
- **Reranking**: `/v1/rerank`
- 支持文档重排序
- 返回相关性分数
- 支持批量处理

**接口格式**:
```json
{
  "model": "rerank-v2",
  "query": "What is AI?",
  "documents": ["doc1", "doc2", "doc3"],
  "top_n": 3
}
```

#### 1.8 其他接口

**功能详情**:
- **Files**: 文件上传和管理
- **Fine-tuning**: 模型微调
- **Batches**: 批量处理
- **Moderation**: 内容审核

**验收标准**:
- [ ] 所有接口兼容 OpenAI 格式
- [ ] 支持 10+ 主流模型提供商
- [ ] 统一的错误处理和响应格式
- [ ] 完整的 API 文档

**优先级**: P0 (核心功能)

---

### 2. 对外模型管理 (FR-002)

**需求描述**: 支持自定义对外模型名称和配置，隐藏底层供应商复杂性，支持多供应商配置和智能路由

**核心业务模型**:
```
对外模型: "GPT-4-Turbo"
├── 供应商A: 第三方OpenAI接口, 成本价 10元/M tokens, 权重 70%
├── 供应商B: 第三方OpenAI接口, 成本价 15元/M tokens, 权重 30%
└── 供应商C: 第三方OpenAI接口, 成本价 12元/M tokens, 备用

对外定价:
├── 普通用户: 18元/M tokens (利润空间: 8元/M tokens)
├── VIP会员:  16元/M tokens (利润空间: 6元/M tokens)
└── 大客户:   独立定价, 如 14元/M tokens (利润空间: 4元/M tokens)
```

**功能详情**:
- **多供应商配置**: 同一对外模型可配置多个供应商
- **成本价管理**: 每个供应商独立的成本价格
- **售价管理**: 支持不同用户群体的独立售价
- **智能路由**: 基于成本优先、可用性、权重的智能路由
- **利润保护**: 确保售价始终高于最高成本价
- **动态切换**: 供应商故障时自动切换到备用供应商

**数据模型**:
```go
// 对外模型配置
type ExternalModel struct {
    ID           string
    Name         string // 对外名称，如 "GPT-4-Turbo"
    Description  string // 模型介绍
    Type         string // 模型类型: chat/completion/image/video/tts/stt/embedding/rerank
    Capabilities []string // 能力标签
    Status       string // active/inactive

    // 供应商路由配置
    SupplierRoutes []SupplierRouteConfig // 供应商路由配置

    // 定价配置
    UserPricing     ModelPricing // 普通用户售价
    VipPricing      ModelPricing // VIP会员售价
    CustomPricings  []CustomPricing // 大客户独立定价
}

// 供应商路由配置
type SupplierRouteConfig struct {
    ID              string
    SupplierName    string // 供应商名称
    SupplierType    string // 第三方供应商类型
    ModelName       string // 实际模型名称
    CostPrice       ModelPricing // 成本价格

    // 路由策略
    Weight          float64 // 权重 (0-100)
    Priority        int     // 优先级 (数字越小优先级越高)
    IsBackup        bool    // 是否为备用供应商
    MaxQPS          int     // QPS 限制

    // 可用性配置
    HealthCheckURL  string // 健康检查URL
    Timeout         int    // 超时时间(秒)
    RetryCount      int    // 重试次数

    Status          string // active/inactive/error
}

// 大客户独立定价
type CustomPricing struct {
    CustomerID      string
    CustomerName    string
    Pricing         ModelPricing // 独立售价
    MinCostPrice    float64 // 最低成本价(用于利润保护)
    Status          string
}
```

**验收标准**:
- [ ] 支持创建自定义对外模型
- [ ] 模型配置支持实时生效
- [ ] 提供模型管理 API 和界面
- [ ] 支持模型版本管理

**优先级**: P0 (核心功能)

---

### 3. 完善的价格体系 (FR-003)

**需求描述**: 支持多供应商成本价管理、差异化售价、利润保护的价格体系

#### 3.1 价格体系架构

**三层价格结构**:
```
┌─────────────────────────────────────────────────────────┐
│  对外售价层 (Selling Price)                             │
│  ├─ 普通用户价格: 18元/M tokens                         │
│  ├─ VIP会员价格: 16元/M tokens                          │
│  └─ 大客户价格: 14元/M tokens (独立定价)                │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  供应商成本层 (Cost Price)                               │
│  ├─ 供应商A: 10元/M tokens (优先选择)                    │
│  ├─ 供应商B: 15元/M tokens (备选)                        │
│  └─ 供应商C: 12元/M tokens (备用)                        │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  利润管理层 (Profit Management)                          │
│  ├─ 利润保护: 确保售价 > 成本价                          │
│  ├─ 利润优化: 优先选择高利润供应商                       │
│  └─ 利润监控: 实时监控利润空间                           │
└─────────────────────────────────────────────────────────┘
```

#### 3.2 成本价管理

**功能详情**:
- **多供应商成本价**: 每个供应商独立的成本价格
- **成本价验证**: 确保成本价合理性和竞争力
- **成本价历史**: 记录成本价变化历史
- **成本价分析**: 成本趋势分析和预警

**成本价配置**:
```go
type SupplierCostPricing struct {
    SupplierID      string
    SupplierName    string

    // 文本模型成本价
    InputCostPerToken          float64
    OutputCostPerToken         float64

    // 图片模型成本价
    OutputCostPerImage         *float64
    OutputCostPerImageHD       *float64

    // 视频/音频成本价
    InputCostPerVideoPerSecond *float64
    OutputCostPerVideoPerSecond *float64
    InputCostPerAudioPerSecond *float64

    // 有效期
    EffectiveDate    time.Time
    ExpiryDate       time.Time

    // 成本价状态
    Status           string // active/inactive
}
```

#### 3.3 售价管理

**功能详情**:
- **用户群体定价**: 不同用户群体独立售价
- **会员定价**: VIP会员专属价格
- **大客户定价**: 独立 negotiated 价格
- **利润保护**: 实时验证售价高于成本价
- **价格策略**: 支持多种定价策略

**售价配置**:
```go
type UserGroupPricing struct {
    UserGroupID     string
    UserGroupName   string // 普通/VIP/企业

    // 文本模型售价
    InputPricePerToken          float64
    OutputPricePerToken         float64

    // 图片模型售价
    OutputPricePerImage         *float64
    OutputPricePerImageHD       *float64

    // 视频/音频售价
    InputPricePerVideoPerSecond *float64
    OutputPricePerVideoPerSecond *float64
    InputPricePerAudioPerSecond *float64

    // 利润保护
    MinProfitMargin  float64 // 最低利润率 (如 10%)

    // 有效期
    EffectiveDate    time.Time
    ExpiryDate       time.Time

    Status           string
}
```

#### 3.4 利润管理

**实时利润计算**:
```go
type ProfitCalculator struct {
    // 获取用户售价
    GetUserSellingPrice(userID string, modelID string) (*ModelPricing, error)

    // 获取供应商成本价
    GetSupplierCostPrice(supplierID string, modelID string) (*ModelPricing, error)

    // 计算利润空间
    CalculateProfit(sellingPrice *ModelPricing, costPrice *ModelPricing) (*ProfitMetrics, error)

    // 验证利润保护
    ValidateProfitMargin(profit *ProfitMetrics, minMargin float64) error
}

type ProfitMetrics struct {
    // 单次请求利润
    RequestProfit    float64
    RequestCost      float64
    RequestRevenue   float64

    // 利润率
    ProfitMargin     float64 // 利润 / 售价

    // 统计数据
    TotalRequests    int64
    TotalProfit      float64
    TotalRevenue     float64
    AverageProfit    float64
}
```

#### 3.1 文本模型定价

**功能详情**:
- **基础定价**: 输入/输出 token 价格
- **层级定价**: 128k/200k/272k token 层级价格
- **模式定价**: 标准/Batches/Priority/Flex 模式
- **缓存定价**: 缓存创建/读取价格
- **时长定价**: 缓存存储时长价格

**定价结构** (参考 Bifrost):
```go
type TextModelPricing struct {
    // 基础定价
    InputCostPerToken  float64
    OutputCostPerToken float64

    // 层级定价
    InputCostPerTokenAbove128kTokens  *float64
    OutputCostPerTokenAbove128kTokens *float64
    InputCostPerTokenAbove200kTokens  *float64
    OutputCostPerTokenAbove200kTokens *float64

    // 缓存定价
    CacheCreationInputTokenCost *float64
    CacheReadInputTokenCost     *float64
    CacheCreationInputTokenCostAbove1hr *float64
}
```

#### 3.2 图片模型定价

**功能详情**:
- **按图片计费**: 每张图片价格
- **按像素计费**: 每像素价格
- **分辨率定价**: 512x512/1024x1024/2048x2048/4096x4096
- **质量定价**: Low/Medium/High/Auto 质量
- **Premium 定价**: 高级图片价格

**定价结构**:
```go
type ImageModelPricing struct {
    InputCostPerImage             *float64
    OutputCostPerImage            *float64
    OutputCostPerImageLowQuality  *float64
    OutputCostPerImageHighQuality *float64
    OutputCostPerImageAbove1024x1024Pixels *float64
    OutputCostPerImageAbove2048x2048Pixels *float64
}
```

#### 3.3 视频/音频定价

**功能详情**:
- **按秒计费**: 视频/音频每秒价格
- **按 Token 计费**: 音频 token 价格
- **输入/输出定价**: 输入和输出分别定价
- **层级定价**: 128k token 以上层级价格

**定价结构**:
```go
type VideoAudioPricing struct {
    InputCostPerVideoPerSecond  *float64
    OutputCostPerVideoPerSecond *float64
    InputCostPerAudioPerSecond  *float64
    OutputCostPerAudioPerSecond *float64
    InputCostPerAudioToken      *float64
    OutputCostPerAudioToken     *float64
}
```

#### 3.4 其他功能定价

**功能详情**:
- **Embedding**: 按 token 计费
- **Reranking**: 按搜索次数计费
- **OCR**: 按页计费
- **标注**: 按页计费
- **代码解释器**: 按会话计费
- **搜索上下文**: 按查询次数计费

**验收标准**:
- [ ] 支持所有 Bifrost 定价模式
- [ ] 价格配置支持热更新
- [ ] 提供价格历史记录
- [ ] 支持价格批量导入导出
- [ ] 实时计费准确无误

**优先级**: P0 (核心功能)

---

### 4. 服务分离架构 (FR-004)

**需求描述**: 实现服务分离，用户端和运维端独立管理，API 网关独立部署

#### 4.1 API 网关服务 (独立)

**功能详情**:
- **高性能**: 专门优化 API 调用性能
- **水平扩展**: 支持多实例部署
- **负载均衡**: 请求分发和负载均衡
- **协议转换**: OpenAI 格式到各供应商格式转换
- **统一入口**: 所有 API 调用的统一入口

**技术要求**:
- 无状态设计，支持水平扩展
- 连接池优化
- 请求/响应缓存
- 限流和熔断

**部署架构**:
```
Internet → Load Balancer → API Gateway (集群) → Business Services
```

#### 4.2 用户端应用 (独立)

**功能详情**:
- **用户注册登录**: 邮箱/手机/第三方登录
- **控制台概览**: 使用统计、费用概览
- **API Key 管理**: 创建、管理、统计 API Key
- **账单查询**: 账单列表、明细、导出
- **使用分析**: 模型使用、费用趋势
- **个人设置**: 个人信息、安全设置

**技术栈**: Vue 3 + TypeScript + shadcn-vue

#### 4.3 运维管理端 (独立)

**功能详情**:
- **用户管理**: 用户列表、详情、状态管理
- **模型管理**: 对外模型配置、路由管理
- **供应商管理**: 供应商配置、密钥管理
- **套餐管理**: 套餐配置、用户套餐
- **系统监控**: 服务状态、性能指标
- **运维日志**: 操作日志、审计日志

**技术栈**: Vue 3 + TypeScript + shadcn-vue
**访问控制**: 独立的运维管理员认证体系

#### 4.4 业务服务 (微服务)

**功能详情**:
- **用户服务**: 用户管理、认证授权
- **模型服务**: 模型配置、路由管理
- **计费服务**: 实时计费、账单生成
- **管理服务**: 运维管理功能

**验收标准**:
- [ ] API 网关独立部署运行
- [ ] 用户端和运维端完全分离
- [ ] 支持水平扩展
- [ ] 服务间通信高效可靠
- [ ] 完整的服务监控

**优先级**: P0 (核心功能)

---

### 5. 用户管理系统 (FR-005)

**需求描述**: 提供完整的用户注册、登录和权限管理（用户端）

**功能详情**:
- **用户注册**: 邮箱/手机号注册
- **用户登录**: 邮箱密码、第三方登录
- **用户信息**: 基本信息管理
- **权限管理**: 基于角色的权限控制
- **密码管理**: 修改密码、找回密码
- **账户安全**: 两步验证、登录日志

**用户模型**:
```go
type User struct {
    ID           string
    Email        string
    Phone        string
    Password     string
    Name         string
    Avatar       string
    Status       string // active/suspended/deleted
    Role         string // user/admin
    MembershipID string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**验收标准**:
- [ ] 支持邮箱和手机号注册
- [ ] 密码加密存储
- [ ] 支持第三方登录 (Google、GitHub)
- [ ] 提供用户管理功能
- [ ] 记录用户操作日志

**优先级**: P0 (核心功能)

---

### 6. 运维管理员系统 (FR-006)

**需求描述**: 提供完整的运维管理员功能（运维端）

**功能详情**:
- **管理员认证**: 独立的认证体系
- **用户管理**: 用户列表、详情、状态管理
- **模型管理**: 对外模型配置、路由管理
- **供应商管理**: 供应商配置、API Key 管理
- **套餐管理**: 套餐配置、用户套餐管理
- **系统监控**: 服务状态、性能指标、告警
- **数据统计**: 用户统计、使用统计、费用统计
- **日志管理**: 操作日志、审计日志、系统日志

**管理员模型**:
```go
type Admin struct {
    ID        string
    Username  string
    Password  string
    Name      string
    Role      string // superadmin/admin/operator
    Status    string // active/suspended
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**验收标准**:
- [ ] 独立的管理员认证体系
- [ ] 完整的用户管理功能
- [ ] 完善的权限控制
- [ ] 实时系统监控
- [ ] 详细的操作日志

**优先级**: P0 (核心功能)

---

### 7. API Key 管理 (FR-007)

**需求描述**: 用户可以创建和管理自己的 API Key

**功能详情**:
- **Key 创建**: 用户自主创建 API Key
- **Key 限制**: 默认 10 个，可申请增加
- **Key 配置**: 权限、额度、过期时间
- **Key 管理**: 查看、编辑、删除、禁用
- **Key 统计**: 使用量、费用统计
- **Key 安全**: 显示部分、重新生成

**API Key 模型**:
```go
type UserAPIKey struct {
    ID          string
    UserID      string
    Name        string
    Key         string // 密钥
    Permissions []string
    Quota       int64 // 额度限制
    ExpireAt    time.Time
    Status      string // active/disabled/deleted
    CreatedAt   time.Time
}
```

**验收标准**:
- [ ] 用户可创建至少 10 个 API Key
- [ ] 支持权限和额度配置
- [ ] 提供 Key 使用统计
- [ ] Key 安全存储和显示
- [ ] 支持 Key 审批流程

**优先级**: P0 (核心功能)

---

### 8. 账单和对账系统 (FR-008)

**需求描述**: 提供完整的费用统计和对账功能

**功能详情**:
- **实时计费**: 实时统计调用费用（参考 Bifrost）
- **账单生成**: 按月生成详细账单
- **账单明细**: 每笔调用的详细信息
- **对账导出**: 异步导出对账数据
- **费用分析**: 模型使用、费用趋势分析
- **预警提醒**: 额度不足、超额提醒

**账单模型**:
```go
type Bill struct {
    ID         string
    UserID     string
    Period     string // 账期: 2024-05
    TotalCost  float64 // 总费用
    TotalUsage int64   // 总使用量
    Details    []BillDetail
    Status     string // pending/paid/overdue
    CreatedAt  time.Time
}

type UsageRecord struct {
    ID         string
    UserID     string
    APIKeyID   string
    Model      string
    ModelType  string // chat/image/video/tts/embedding/rerank
    Tokens     int64
    Images     int64
    Seconds    int64
    Cost       float64
    Timestamp  time.Time
}
```

**验收标准**:
- [ ] 实时统计用户费用
- [ ] 每月自动生成账单
- [ ] 支持对账数据导出 (CSV/Excel)
- [ ] 提供费用分析图表
- [ ] 支持费用预警设置

**优先级**: P1 (重要功能)

---

### 9. 智能路由和成本优化 (FR-009)

**需求描述**: 基于成本优先、多供应商配置的智能路由系统

#### 9.1 多供应商路由策略

**核心逻辑**:
```
请求: 用户调用 "GPT-4-Turbo" 模型
↓
识别用户类型: VIP会员
↓
确定售价: 16元/M tokens
↓
选择最优供应商:
  1. 供应商A: 成本10元, 利润6元, 优先级最高 ✅
  2. 供应商C: 成本12元, 利润4元, 备用
  3. 供应商B: 成本15元, 利润1元, 最后选择
↓
实时监控: 供应商A状态正常, 负载不高
↓
路由决策: 使用供应商A
↓
故障处理: 如果A失败, 自动切换到C, 再到B
```

**路由策略详解**:

##### 1. 成本优先路由
- **利润计算**: `利润 = 售价 - 供应商成本价`
- **选择逻辑**: 优先选择利润空间最大的供应商
- **约束条件**: 供应商状态正常、QPS未超限

##### 2. 可用性保障
- **健康检查**: 定期检查供应商可用性
- **故障转移**: 供应商故障时自动切换
- **恢复机制**: 供应商恢复后自动加回路由

##### 3. 负载均衡
- **权重分配**: 按配置的权重分配请求
- **QPS限制**: 尊重每个供应商的QPS限制
- **动态调整**: 根据实时性能动态调整

**功能详情**:
- **多供应商配置**: 同一模型可配置多个供应商
- **成本价管理**: 每个供应商独立的成本价格
- **售价管理**: 不同用户群体独立售价
- **智能路由**: 基于成本、可用性、权重的智能决策
- **利润保护**: 实时计算利润空间，确保盈利
- **故障转移**: 自动切换到备用供应商
- **性能监控**: 实时监控供应商性能指标

**路由算法**:
```go
type RouterStrategy struct {
    // 路由决策
    SelectSupplier(userID string, modelID string) (*SupplierRouteConfig, error)

    // 成本计算
    CalculateProfit(userID string, supplier *SupplierRouteConfig) (float64, error)

    // 可用性检查
    CheckSupplierHealth(supplier *SupplierRouteConfig) bool

    // 性能监控
    MonitorSupplierPerformance(supplier *SupplierRouteConfig) *PerformanceMetrics
}

// 路由决策逻辑
func (r *RouterStrategy) SelectSupplier(userID string, modelID string) (*SupplierRouteConfig, error) {
    // 1. 获取用户类型和对应售价
    userType := r.getUserType(userID)
    sellingPrice := r.getSellingPrice(modelID, userType)

    // 2. 获取所有可用供应商
    suppliers := r.getAvailableSuppliers(modelID)

    // 3. 过滤掉成本高于售价的供应商
    validSuppliers := r.filterByProfitMargin(suppliers, sellingPrice)

    // 4. 按利润空间排序 (利润高的优先)
    sortedSuppliers := r.sortByProfit(validSuppliers, sellingPrice)

    // 5. 检查可用性和负载
    for _, supplier := range sortedSuppliers {
        if r.CheckSupplierHealth(supplier) && !r.isOverloaded(supplier) {
            return supplier, nil
        }
    }

    // 6. 如果没有合适供应商，返回错误
    return nil, errors.New("no available supplier")
}
```

#### 9.2 成本优化策略

**利润管理**:
```go
type ProfitMetrics struct {
    // 实时利润计算
    CurrentProfit      float64 // 当前利润空间
    ProfitMargin       float64 // 利润率 (利润/售价)
    CostVsSellingRatio float64 // 成本售价比

    // 统计数据
    TotalRequests      int64
    TotalCost          float64
    TotalRevenue       float64
    TotalProfit        float64
}

// 利润保护机制
func (r *RouterStrategy) ValidateProfitMargin(sellingPrice float64, costPrice float64) error {
    minProfitMargin := 0.1 // 最低利润率 10%

    profitMargin := (sellingPrice - costPrice) / sellingPrice
    if profitMargin < minProfitMargin {
        return fmt.Errorf("profit margin too low: %.2f%%", profitMargin*100)
    }

    return nil
}
```

#### 9.3 实时监控和告警

**监控指标**:
```go
type SupplierMetrics struct {
    SupplierID        string
    SupplierName      string

    // 性能指标
    AverageLatency    time.Duration
    SuccessRate       float64
    ErrorRate         float64
    CurrentQPS        int64

    // 成本指标
    TotalCost         float64
    AverageCostPerRequest float64

    // 利润指标
    TotalProfit       float64
    AverageProfitPerRequest float64
    ProfitMargin      float64

    // 状态指标
    Status            string // healthy/degraded/error
    LastHealthCheck   time.Time
}
```

**告警机制**:
- **成本告警**: 供应商成本异常上升
- **可用性告警**: 供应商故障率上升
- **利润告警**: 利润空间低于阈值
- **性能告警**: 响应时间异常

**验收标准**:
- [ ] 支持多供应商配置
- [ ] 成本优先路由正确工作
- [ ] 故障自动转移时间 < 1s
- [ ] 利润保护机制有效
- [ ] 路由配置实时生效
- [ ] 提供路由监控和统计
- [ ] 支持手动切换路由
- [ ] 实时利润计算准确

**优先级**: P0 (核心功能)

---

### 10. 会员套餐体系 (FR-010)

**需求描述**: 建立完整的会员和套餐管理体系

**功能详情**:
- **会员等级**: 普通、高级、VIP 等
- **套餐配置**: 包含免费额度、费率折扣、优先级等
- **套餐购买**: 在线购买和激活
- **套餐管理**: 套餐列表、详情、修改
- **到期管理**: 套餐到期提醒和自动续费

**套餐模型**:
```go
type MembershipPlan struct {
    ID          string
    Name        string
    Level       string // basic/premium/vip
    FreeQuota   int64  // 免费额度
    Discount    float64 // 费率折扣
    Priority    int    // 优先级
    Price       float64 // 套餐价格
    Duration    int    // 有效期(天)
    Features    []string // 特性列表
}
```

**验收标准**:
- [ ] 支持至少 3 种会员等级
- [ ] 套餐配置实时生效
- [ ] 提供套餐管理界面
- [ ] 支持套餐统计和分析

**优先级**: P1 (重要功能)

---

## 非功能性需求

### 性能要求 (NFR-001)

- **响应时间**: API 调用增加延迟 < 20ms
- **并发能力**: API 网关支持 10000+ QPS
- **可用性**: 99.9% 可用性
- **数据一致性**: 强一致性保证
- **水平扩展**: API 网关支持水平扩展

### 安全要求 (NFR-002)

- **数据加密**: 传输加密 (TLS)、存储加密
- **访问控制**: 基于角色的权限控制
- **审计日志**: 完整的操作审计
- **安全认证**: 支持多种认证方式
- **API 安全**: API Key 验证、限流保护

### 可扩展性 (NFR-003)

- **水平扩展**: API 网关支持分布式部署
- **模型扩展**: 易于添加新的模型供应商
- **功能扩展**: 插件化架构
- **服务扩展**: 微服务架构

### 可维护性 (NFR-004)

- **代码质量**: 清晰的代码结构和文档
- **测试覆盖**: 单元测试覆盖率 > 80%
- **监控告警**: 完善的监控和告警系统
- **日志管理**: 结构化日志和日志分析

---

## 技术约束

### 技术栈要求

- **后端**: Go 1.26.2
- **数据库**: PostgreSQL (已配置)
- **ORM**: GORM (自动迁移，禁用外键)
- **前端**: Vue 3 + TypeScript + shadcn-vue
- **核心**: 基于 Bifrost (Go)
- **部署**: Docker 容器化

### 数据库约束

- **数据库**: llm_gateway
- **用户**: llm_gateway
- **密码**: llm_gateway
- **端口**: 5432
- **外键**: 禁用 (由应用层维护)

---

## 开发优先级

### P0 - 核心功能 (必须完成)

1. 统一模型接口 (FR-001) - 所有模型类型
2. 对外模型管理 (FR-002)
3. 完善的价格体系 (FR-003) - 参考 Bifrost
4. 服务分离架构 (FR-004) - API 网关独立
5. 用户管理系统 (FR-005)
6. 运维管理员系统 (FR-006)
7. API Key 管理 (FR-007)
8. 路由和负载均衡 (FR-009)

### P1 - 重要功能 (尽快完成)

1. 账单和对账系统 (FR-008)
2. 会员套餐体系 (FR-010)

---

## 验收标准

### 功能验收

- [ ] 所有 P0 功能完整实现
- [ ] 所有模型接口正常工作
- [ ] 核心流程测试通过
- [ ] UI/UX 验收通过

### 性能验收

- [ ] API 响应时间满足要求
- [ ] 并发能力测试通过
- [ ] 稳定性测试通过
- [ ] 水平扩展验证通过

### 安全验收

- [ ] 安全扫描无高危漏洞
- [ ] 权限控制测试通过
- [ ] 数据加密验证通过
- [ ] API 安全测试通过

---

## 附录

### 术语表

- **对外模型**: 对用户展示的模型名称，可自定义
- **供应商**: 实际提供 AI 模型的服务商
- **API 网关**: 独立部署的 API 调用网关服务
- **用户端**: 面向普通用户的前端应用
- **运维端**: 面向运维管理员的前端应用

### 参考资料

- Bifrost 项目文档
- OpenAI API 文档
- PostgreSQL 数据库设计
- GORM 使用指南
- Vue 3 文档
- shadcn-vue 文档

---

**文档维护**: 本文档随项目进展持续更新
**最后更新**: 2026-05-08
**更新者**: 项目团队
