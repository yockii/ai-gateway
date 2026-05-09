# AI Gateway 项目路线图

## 项目概述

**项目名称**: AI Gateway Platform
**技术方案**: 基于 Bifrost 深度二次开发
**开发周期**: 12-14 周
**团队规模**: 3-4 人
**目标**: 构建企业级 AI 大模型网关平台

---

## 总体时间规划

```
Phase 1:  准备和基础搭建 (3-4 周)
Phase 2:  核心功能开发 (5-6 周)
Phase 3:  高级功能和优化 (3-4 周)
Phase 4:  测试和部署 (1-2 周)
```

---

## Phase 1: 准备和基础搭建

**时间**: Week 1-4 (3-4 周)
**目标**: 完成 Bifrost 集成和基础架构

### 1.1 Bifrost 项目集成和环境搭建 (Week 1)

**任务**:
- [ ] Fork Bifrost 项目到本地仓库
- [ ] 搭建开发环境 (Go 1.26.2, Node.js 18+, PostgreSQL)
- [ ] 运行 Bifrost 原版项目
- [ ] 理解 Bifrost 核心架构
- [ ] 研究 Bifrost 价格体系实现
- [ ] 阅读关键文档 (AGENTS.md, README.md)

**产出**:
- 本地开发环境可正常运行
- 核心架构理解文档
- Bifrost 价格体系分析文档
- 开发工具配置完成

**验收标准**:
- Bifrost 原版项目本地运行成功
- 能调用至少一个模型 API
- 团队成员理解核心架构
- 理解价格体系实现

---

### 1.2 服务架构设计和数据库架构 (Week 2)

**任务**:
- [ ] 设计服务分离架构
  - API 网关服务
  - 用户端应用
  - 运维管理端应用
  - 业务微服务
- [ ] 设计数据库表结构
  - 用户表 (users)
  - 管理员表 (admins)
  - 会员套餐表 (membership_plans)
  - 用户套餐关联表 (user_memberships)
  - 用户 API Key 表 (user_api_keys)
  - 对外模型配置表 (external_models)
  - 扩展 Bifrost 价格表
- [ ] 配置 GORM 自动迁移
- [ ] 设计服务间通信协议

**服务架构设计**:
```
┌──────────────────────────────────────┐
│  用户端 (Vue + shadcn-vue)            │
│  运维端 (Vue + shadcn-vue)            │
└──────────────────────────────────────┘
              ↓
┌──────────────────────────────────────┐
│  API 网关服务 (独立，水平扩展)        │
└──────────────────────────────────────┘
              ↓
┌──────────────────────────────────────┐
│  业务微服务                           │
│  - 用户服务                           │
│  - 模型服务                           │
│  - 计费服务                           │
│  - 管理服务                           │
└──────────────────────────────────────┘
              ↓
┌──────────────────────────────────────┐
│  Bifrost 核心引擎                     │
└──────────────────────────────────────┘
```

**数据模型**:
```go
// 用户表
type User struct {
    ID           string
    Email        string
    Phone        string
    Password     string
    Name         string
    Avatar       string
    Status       string
    Role         string
    MembershipID string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// 管理员表
type Admin struct {
    ID        string
    Username  string
    Password  string
    Name      string
    Role      string // superadmin/admin/operator
    Status    string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 对外模型配置表
type ExternalModel struct {
    ID           string
    Name         string // 对外名称
    Description  string
    Type         string // chat/completion/image/video/tts/stt/embedding/rerank
    Provider     string
    ModelName    string
    Capabilities []string
    Pricing      ModelPricing // 参考 Bifrost
    Status       string
    CreatedAt    time.Time
}
```

**产出**:
- 服务架构设计文档
- 完整的数据库迁移脚本
- 数据模型文档
- 测试数据脚本

**验收标准**:
- GORM 自动迁移成功
- 所有表结构符合需求
- 服务架构清晰合理

---

### 1.3 前端项目搭建 (Week 3)

**任务**:
- [ ] 搭建用户端 Vue 3 项目
  - 配置 TypeScript
  - 集成 shadcn-vue
  - 配置路由和状态管理
  - 配置 API 请求库
- [ ] 搭建运维端 Vue 3 项目
  - 配置 TypeScript
  - 集成 shadcn-vue
  - 配置路由和状态管理
  - 配置 API 请求库
- [ ] 设计统一的 UI 组件库
- [ ] 实现基础布局组件

**前端技术栈**:
```json
{
  "framework": "Vue 3",
  "language": "TypeScript",
  "ui": "shadcn-vue",
  "router": "Vue Router",
  "state": "Pinia",
  "http": "axios",
  "validation": "zod"
}
```

**项目结构**:
```
frontend/
├── user-portal/          # 用户端
│   ├── src/
│   │   ├── components/   # 组件
│   │   ├── views/        # 页面
│   │   ├── router/       # 路由
│   │   └── stores/       # 状态
│   └── package.json
├── admin-portal/         # 运维端
│   ├── src/
│   │   ├── components/   # 组件
│   │   ├── views/        # 页面
│   │   ├── router/       # 路由
│   │   └── stores/       # 状态
│   └── package.json
└── shared/               # 共享组件
    └── components/
```

**产出**:
- 用户端 Vue 项目基础架构
- 运维端 Vue 项目基础架构
- 基础 UI 组件库
- 开发规范文档

**验收标准**:
- 两个前端项目可以正常运行
- shadcn-vue 组件正常工作
- 路由和状态管理配置正确

---

### 1.4 用户认证和管理员认证系统 (Week 4)

**任务**:
- [ ] 实现用户注册 API
- [ ] 实现用户登录 API (JWT)
- [ ] 实现密码加密和验证
- [ ] 实现第三方登录 (Google/GitHub)
- [ ] 实现权限中间件
- [ ] 实现管理员认证系统 (独立)
- [ ] 编写认证测试

**API 设计**:
```
# 用户认证
POST   /api/v1/user/register        # 用户注册
POST   /api/v1/user/login           # 用户登录
POST   /api/v1/user/logout          # 用户登出
GET    /api/v1/user/profile         # 获取用户信息
PUT    /api/v1/user/profile         # 更新用户信息

# 管理员认证
POST   /api/v1/admin/login          # 管理员登录
POST   /api/v1/admin/logout         # 管理员登出
GET    /api/v1/admin/profile        # 获取管理员信息
```

**产出**:
- 用户认证 API
- 管理员认证 API
- JWT Token 管理
- 权限控制中间件

**验收标准**:
- 注册登录流程正常
- JWT Token 验证正确
- 权限控制有效
- 用户和管理员认证完全分离

---

## Phase 2: 核心功能开发

**时间**: Week 5-10 (5-6 周)
**目标**: 完成核心业务功能

### 2.1 API 网关服务开发 (Week 5)

**任务**:
- [ ] 实现 API 网关核心功能
- [ ] 实现 OpenAI 兼容协议转换
- [ ] 实现负载均衡和故障转移
- [ ] 实现请求路由和分发
- [ ] 实现限流和熔断
- [ ] 实现监控和日志
- [ ] 性能优化和测试

**API 网关功能**:
```go
type APIGateway struct {
    router      *Router
    lb          *LoadBalancer
    limiter     *RateLimiter
    circuit     *CircuitBreaker
    monitor     *Monitor
}

// 支持的 OpenAI 兼容接口
GET    /v1/models
POST   /v1/chat/completions
POST   /v1/completions
POST   /v1/images/generations
POST   /v1/videos/generations
POST   /v1/audio/speech
POST   /v1/audio/transcriptions
POST   /v1/embeddings
POST   /v1/rerank
```

**产出**:
- API 网关服务
- 协议转换层
- 负载均衡器
- 性能测试报告

**验收标准**:
- 支持所有 OpenAI 兼容接口
- 响应时间 < 20ms
- 支持 10000+ QPS
- 故障转移时间 < 1s

---

### 2.2 用户 API Key 管理 (Week 6)

**任务**:
- [ ] 实现 API Key 生成逻辑
- [ ] 实现 API Key CRUD API
- [ ] 实现 API Key 验证中间件
- [ ] 实现 API Key 使用统计
- [ ] 开发 API Key 管理界面（用户端）

**API 设计**:
```
GET    /api/v1/user/keys            # 获取 API Key 列表
POST   /api/v1/user/keys            # 创建 API Key
GET    /api/v1/user/keys/:id        # 获取 API Key 详情
PUT    /api/v1/user/keys/:id        # 更新 API Key
DELETE /api/v1/user/keys/:id        # 删除 API Key
PATCH  /api/v1/user/keys/:id/disable # 禁用 API Key
GET    /api/v1/user/keys/:id/stats  # 获取使用统计
```

**产出**:
- API Key 管理功能
- API Key 管理界面
- 使用统计功能

**验收标准**:
- 用户可创建和管理 API Key
- API Key 验证正确
- 使用统计准确
- 界面操作流畅

---

### 2.3 对外模型管理 (Week 7)

**任务**:
- [ ] 扩展 Bifrost 模型配置
- [ ] 实现对外模型 CRUD API
- [ ] 实现模型元数据管理
- [ ] 实现模型路由配置
- [ ] 开发模型管理界面（运维端）
- [ ] 实现模型类型分类（文本/图片/视频/TTS/Embedding/Reranking）

**API 设计**:
```
GET    /api/v1/admin/models          # 获取对外模型列表
POST   /api/v1/admin/models          # 创建对外模型
GET    /api/v1/admin/models/:id      # 获取模型详情
PUT    /api/v1/admin/models/:id      # 更新模型配置
DELETE /api/v1/admin/models/:id      # 删除模型
PATCH  /api/v1/admin/models/:id/status # 更新模型状态
GET    /api/v1/models                # 公开接口：获取可用模型列表
```

**产出**:
- 对外模型管理功能
- 模型配置界面
- 模型路由配置

**验收标准**:
- 支持所有模型类型
- 模型路由正确工作
- 配置实时生效
- 界面操作便捷

---

### 2.4 完善的价格体系 (Week 8)

**任务**:
- [ ] 集成 Bifrost 价格体系
- [ ] 实现价格配置管理
- [ ] 实现实时计费逻辑
- [ ] 实现价格验证和转换
- [ ] 开发价格管理界面（运维端）
- [ ] 支持所有模型类型定价

**价格体系功能**:
```go
// 文本模型定价
type TextPricing struct {
    InputCostPerToken          float64
    OutputCostPerToken         float64
    InputCostPerTokenAbove128k *float64
    CacheCreationInputTokenCost *float64
    CacheReadInputTokenCost     *float64
}

// 图片模型定价
type ImagePricing struct {
    OutputCostPerImage            *float64
    OutputCostPerImageHighQuality *float64
    OutputCostPerImageAbove1024x1024 *float64
}

// 视频/音频定价
type VideoAudioPricing struct {
    InputCostPerVideoPerSecond  *float64
    OutputCostPerVideoPerSecond *float64
    InputCostPerAudioPerSecond  *float64
}
```

**产出**:
- 价格配置系统
- 实时计费功能
- 价格管理界面

**验收标准**:
- 支持所有 Bifrost 定价模式
- 实时计费准确
- 价格配置灵活
- 界面功能完善

---

### 2.5 会员套餐体系 (Week 9)

**任务**:
- [ ] 实现定价配置管理
- [ ] 实现会员套餐管理
- [ ] 实现用户套餐关联
- [ ] 实现费率计算逻辑
- [ ] 开发套餐管理界面（运维端）
- [ ] 开发套餐购买流程（用户端）

**API 设计**:
```
# 套餐管理（运维端）
GET    /api/v1/admin/plans           # 获取套餐列表
POST   /api/v1/admin/plans           # 创建套餐
GET    /api/v1/admin/plans/:id       # 获取套餐详情
PUT    /api/v1/admin/plans/:id       # 更新套餐

# 用户套餐（用户端）
GET    /api/v1/user/membership       # 获取当前套餐
POST   /api/v1/user/membership/purchase # 购买套餐
PUT    /api/v1/user/membership/upgrade   # 升级套餐
```

**产出**:
- 定价配置系统
- 会员套餐功能
- 套餐管理界面
- 套餐购买流程

**验收标准**:
- 套餐配置正确应用
- 费率计算准确
- 套餐购买流程完整
- 界面功能完善

---

### 2.6 账单和对账系统 (Week 10)

**任务**:
- [ ] 扩展 Bifrost 日志记录
- [ ] 实现实时计费逻辑
- [ ] 实现账单生成任务
- [ ] 实现账单明细查询
- [ ] 实现对账数据导出
- [ ] 开发账单查询界面（用户端）
- [ ] 开发账单管理界面（运维端）

**API 设计**:
```
# 用户端
GET    /api/v1/user/bills            # 获取账单列表
GET    /api/v1/user/bills/:id        # 获取账单详情
GET    /api/v1/user/bills/:id/details # 获取账单明细
GET    /api/v1/user/usage            # 获取使用统计
POST   /api/v1/user/bills/:id/export # 导出对账数据

# 运维端
GET    /api/v1/admin/bills           # 获取所有账单
GET    /api/v1/admin/bills/:id       # 获取账单详情
POST   /api/v1/admin/bills/generate  # 生成账单
```

**产出**:
- 账单生成系统
- 使用统计功能
- 对账导出功能
- 账单查询界面

**验收标准**:
- 实时计费准确
- 账单自动生成
- 导出数据完整
- 界面功能完善

---

### Phase 2 执行计划

**Status:** ✅ COMPLETE (2026-05-08)

**Plans:** 4/4 plans executed

**Wave Structure:**
- **Wave 1:** 02-01 (Data Models), 02-02 (Bifrost Integration & Gateway Services) - ✅ Complete
- **Wave 2:** 02-03 (OpenAI API Handlers), 02-04 (Membership, Billing & Routing) - ✅ Complete

**Plan Details:**
- [x] `02-01-PLAN.md` — Data Model Foundation (UserAPIKey, Membership, ModelMapping, Pricing)
- [x] `02-02-PLAN.md` — Bifrost Integration & API Gateway Services
- [x] `02-03-PLAN.md` — OpenAI-Compatible API Handlers (Images, Audio, Embeddings, Admin)
- [x] `02-04-PLAN.md` — Membership, Billing & Routing Services

**Summary Documents:**
- `02-01-SUMMARY.md` — 5 tasks, 6 files created/modified
- `02-02-SUMMARY.md` — 5 tasks, 5 files created/modified
- `02-03-SUMMARY.md` — 5 tasks, 7 files created/modified
- `02-04-SUMMARY.md` — 4 tasks, 6 files created/modified

---

## Phase 3: 高级功能和优化

**时间**: Week 11-14 (3-4 周)
**目标**: 完善功能和优化性能

### 3.1 用户端界面完善 (Week 11)

**任务**:
- [ ] 实现用户注册登录页面
- [ ] 实现控制台概览页面
- [ ] 实现 API Key 管理页面
- [ ] 实现使用统计页面
- [ ] 实现账单查询页面
- [ ] 实现个人设置页面
- [ ] 实现套餐购买页面

**页面列表**:
- 登录/注册页面
- 控制台概览
- API Key 管理
- 使用统计和图表
- 账单查询和导出
- 个人设置
- 套餐购买和管理

**产出**:
- 完整的用户界面
- 响应式设计
- 良好的用户体验
- 数据可视化

**验收标准**:
- 界面美观易用
- 支持主流浏览器
- 移动端适配良好
- 交互流畅

---

### 3.2 运维管理端界面完善 (Week 12)

**任务**:
- [ ] 实现管理员登录页面
- [ ] 实现用户管理页面
- [ ] 实现模型管理页面
- [ ] 实现套餐管理页面
- [ ] 实现供应商管理页面
- [ ] 实现账单管理页面
- [ ] 实现系统监控页面
- [ ] 实现运维日志页面

**管理功能**:
- 用户列表和详情管理
- 用户状态管理
- 批量操作支持
- 数据统计和分析
- 实时监控展示

**产出**:
- 完整的管理后台
- 权限控制完善
- 操作便捷高效
- 数据可视化

**验收标准**:
- 所有管理功能可用
- 权限控制正确
- 数据统计准确
- 实时监控正常

---

### 3.3 性能优化和安全加固 (Week 13)

**任务**:
- [ ] API 网关性能优化
- [ ] 数据库查询优化
- [ ] 缓存策略优化
- [ ] 连接池优化
- [ ] 安全漏洞扫描
- [ ] 权限控制审计
- [ ] 数据加密验证
- [ ] 压力测试和调优

**优化目标**:
- API 响应时间 < 20ms
- API 网关支持 10000+ QPS
- 99.9% 可用性
- 无高危安全漏洞
- 支持水平扩展

**性能测试**:
```bash
# 压力测试
ab -n 100000 -c 1000 http://gateway/v1/chat/completions

# 水平扩展测试
# 启动多个 API 网关实例，验证负载均衡
```

**产出**:
- 性能测试报告
- 安全审计报告
- 优化文档
- 部署指南

**验收标准**:
- 性能指标达标
- 安全测试通过
- 压力测试通过
- 水平扩展验证通过

---

### 3.4 监控和日志系统 (Week 14)

**任务**:
- [ ] 实现服务监控
- [ ] 实现性能监控
- [ ] 实现告警系统
- [ ] 实现日志收集和分析
- [ ] 实现统计报表
- [ ] 配置监控仪表盘

**监控指标**:
```go
type Metrics struct {
    // 服务指标
    QPS              int64
    ResponseTime     time.Duration
    ErrorRate        float64

    // 业务指标
    ActiveUsers      int64
    APIKeyCount      int64
    RequestCount     int64
    Cost             float64

    // 系统指标
    CPU              float64
    Memory           float64
    DiskIO           float64
    NetworkIO        float64
}
```

**产出**:
- 监控系统
- 告警系统
- 日志分析系统
- 监控仪表盘

**验收标准**:
- 监控数据准确
- 告警及时有效
- 日志分析完善
- 仪表盘直观

---

### Phase 3 执行计划

**Status:** ⚠️ COMPLETE (2026-05-09) - 90% 完成，待修复小问题

**Plans:** 4 plans created

**Wave Structure:**
- **Wave 1:** 03-01 (User Portal UI), 03-02 (Admin Portal UI) - 并行执行
- **Wave 2:** 03-03 (Performance & Security), 03-04 (Monitoring & Logging) - 依赖 Wave 1

**Plan Details:**
- [x] `03-01-PLAN.md` — 用户端界面完善 (Vue 3 + shadcn-vue, 11 个任务)
- [x] `03-02-PLAN.md` — 运维管理端界面完善 (Vue 3 + shadcn-vue, 11 个任务)
- [x] `03-03-PLAN.md` — 性能优化和安全加固 (7 个任务)
- [x] `03-04-PLAN.md` — 监控和日志系统 (6 个任务)

**Plan Summaries:**
- **03-01:** 用户端前端项目初始化、API 客户端封装、路由配置、状态管理、布局组件、所有页面（登录/注册、Dashboard、API Keys、Usage、Bills、Settings）
- **03-02:** 运维端前端项目初始化、API 客户端封装、管理页面（用户、模型、供应商、套餐、监控、运维大屏）、表格组件
- **03-03:** Redis 缓存层、缓存中间件、数据库查询优化、API 网关性能优化、安全扫描集成（GoSec + SonarQube）、性能测试脚本
- **03-04:** Prometheus metrics 暴露、Grafana 仪表盘、Loki 日志聚合、告警系统、Docker Compose 部署配置

---
---

## Phase 4: 测试和部署

**时间**: Week 15-16 (1-2 周)
**目标**: 完成测试和生产部署

### 4.1 集成测试 (Week 15.1)

**任务**:
- [ ] 编写集成测试用例
- [ ] 端到端测试
- [ ] 用户验收测试
- [ ] 性能测试
- [ ] 安全测试

**测试覆盖**:
- 用户注册登录流程
- API Key 创建和使用
- 所有模型接口调用
- 套餐购买和生效
- 账单生成和查询
- 运维管理功能

**产出**:
- 测试报告
- 问题清单
- 修复方案

**验收标准**:
- 核心流程测试通过
- 无 P0/P1 级 Bug
- 性能指标达标
- 安全测试通过

---

### 4.2 部署和上线 (Week 15.2-16)

**任务**:
- [ ] 准备部署环境
- [ ] 配置 Docker 镜像
- [ ] 配置数据库迁移
- [ ] 配置监控告警
- [ ] 执行部署上线
- [ ] 生产环境验证
- [ ] 用户培训

**部署架构**:
```yaml
services:
  # API 网关（集群）
  api-gateway:
    image: ai-gateway:latest
    replicas: 3
    ports:
      - "8080:8080"

  # 用户服务
  user-service:
    image: user-service:latest
    replicas: 2

  # 模型服务
  model-service:
    image: model-service:latest
    replicas: 2

  # 用户端前端
  user-portal:
    image: user-portal:latest
    replicas: 2

  # 运维端前端
  admin-portal:
    image: admin-portal:latest
    replicas: 1
```

**部署清单**:
- [ ] API 网关集群部署
- [ ] 业务服务部署
- [ ] 前端应用部署
- [ ] 数据库迁移
- [ ] 负载均衡配置
- [ ] 监控配置
- [ ] 备份策略
- [ ] 回滚方案

**产出**:
- 生产环境运行
- 监控告警正常
- 部署文档
- 运维手册

**验收标准**:
- 生产环境稳定运行
- 监控指标正常
- 用户可正常使用
- 性能指标达标

---

### Phase 4 执行计划

**Status:** ✅ COMPLETE (2026-05-09)

**Plans:** 4/4 plans executed

**Wave Structure:**
- **Wave 1:** 04-01 (Go Integration Testing Infrastructure) - ✅ Complete
- **Wave 2:** 04-02 (Frontend Testing Infrastructure) - ✅ Complete
- **Wave 3:** 04-03 (CI/CD Pipeline & Performance Testing) - ✅ Complete
- **Wave 4:** 04-04 (Production Deployment Configuration) - ✅ Complete

**Plan Details:**
- [x] `04-01-PLAN.md` — Go Integration Testing Infrastructure (6 个任务) - 30 个集成测试
- [x] `04-02-PLAN.md` — Frontend Testing Infrastructure (7 个任务) - 79 个测试
- [x] `04-03-PLAN.md` — CI/CD Pipeline & Performance Testing (6 个任务) - GitHub Actions 工作流
- [x] `04-04-PLAN.md` — Production Deployment Configuration (8 个任务) - 蓝绿部署 + 1658 行文档

**Plan Summaries:**
- **04-01:** Testcontainers 集成、API 集成测试、数据库迁移测试、Makefile 测试目标
- **04-02:** Vitest 单元测试配置、Playwright E2E 测试、MSW API Mock（用户端 + 运维端） ✅
- **04-03:** K6 负载测试、GitHub Actions 工作流（test/build/deploy）、性能基准测试
- **04-04:** Docker Compose 多环境配置、nginx 反向代理、蓝绿部署脚本、部署文档

---

## 关键里程碑

| 里程碑 | 时间 | 标志性成果 |
|--------|------|------------|
| M1: 基础搭建完成 | Week 4 | Bifrost 集成，双端认证可用，前端项目就绪 |
| M2: 核心功能完成 | Week 10 | API 网关、模型管理、计费系统可用 |
| M3: 界面完成 | Week 12 | 用户端和运维端界面完成 |
| M4: 项目上线 | Week 16 | 生产环境部署完成 |

---

## 风险和挑战

### 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Bifrost 学习曲线 | 中 | 预留学习时间，利用详细文档 |
| API 网关性能要求高 | 高 | 早期性能测试，充分优化 |
| 服务分离架构复杂 | 中 | 充分的架构设计评审 |
| Vue 生态熟悉度 | 低 | 团队培训，技术预研 |

### 进度风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 功能延期 | 中 | 分阶段交付，优先核心功能 |
| 资源不足 | 高 | 及时调整计划，申请资源 |
| 需求变更 | 中 | 变更控制流程，影响评估 |

---

## 资源需求

### 人力资源

- **后端开发**: 2 人
- **前端开发**: 1 人 (Vue + shadcn-vue)
- **测试**: 0.5 人
- **运维**: 0.5 人

### 技术资源

- **开发环境**: Go 1.26.2, Node.js 18+, PostgreSQL 14+
- **服务器**: 开发服务器 3 台，测试服务器 2 台
- **第三方服务**: Google/GitHub OAuth, 监控服务

---

## 质量保证

### 代码质量

- 代码审查制度
- 单元测试覆盖率 > 80%
- 代码规范检查
- 性能测试

### 文档要求

- API 文档完整
- 数据模型文档
- 部署文档
- 用户手册
- 运维手册

---

## 下一步行动

### 立即开始

1. ✅ 技术方案确认
2. ⏳ Fork Bifrost 项目
3. ⏳ 搭建开发环境 (Go 1.26.2, Vue 3)
4. ⏳ 组建开发团队

### 本周目标

1. 完成环境搭建
2. 运行 Bifrost 原版
3. 理解核心架构和价格体系
4. 开始服务架构设计

---

**路线图维护**: 根据项目进展每周更新
**最后更新**: 2026-05-09
**更新者**: 项目团队
