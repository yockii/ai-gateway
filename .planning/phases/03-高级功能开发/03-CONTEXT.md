# Phase 03: 高级功能开发 - Context

**Gathered:** 2026-05-08
**Status:** Ready for planning

## Phase Boundary

阶段 3 负责实现前端用户界面和后端性能优化，包括：
- **用户端界面**：Vue 3 + shadcn-vue 用户门户，包含控制台概览、API Key 管理、使用统计、账单查询等页面
- **运维端界面**：Vue 3 + shadcn-vue 管理后台，包含用户管理、模型管理、供应商管理、运维大屏等页面
- **性能优化**：API 网关性能优化、缓存策略、数据库查询优化、安全加固
- **监控系统**：Prometheus + Grafana 监控、Loki + Grafana 日志、告警系统

## Implementation Decisions

### 前端项目架构
- **D-01:** 采用 **Monorepo 结构** — frontend/ 目录下包含 user/ 和 admin/ 两个子项目，共享部分代码
- **D-02:** **仅 UI 共享** — shared/ 中放通用 UI 组件和类型定义，API 客户端各自独立实现
- **D-03:** 使用 **Vite** 作为构建工具 — 更快的开发体验，原生 ESM 支持
- **D-04:** **封装 axios** — 基于 axios 封装 API 客户端，统一处理认证、错误、重试逻辑

### 用户端 UI 设计
- **D-05:** 采用 **经典管理后台布局** — 侧边栏（可折叠）+ 顶部栏 + 内容区
- **D-06:** 使用 **ECharts** 进行数据可视化 — 用于使用统计和费用趋势图表
- **D-07:** 使用 **shadcn-vue 内置表单验证** — 用于注册登录、API Key 创建等表单
- **D-08:** **后端异步生成 + 对象存储** — 账单导出（PDF/CSV）由后端异步生成，存储到 S3/OSS，前端提供下载链接列表

### 运维端 UI 设计
- **D-09:** **混合布局** — 管理部分与用户端保持一致（经典管理后台），同时需要专门的运维大屏来实时展示各类数据指标
- **D-10:** 使用 **shadcn-vue DataTable** 组件 — 用于用户列表、模型列表等数据表格
- **D-11:** 采用 **逐行操作** — 每行独立操作按钮，支持快速操作（而非批量选择）
- **D-12:** 运维大屏使用 **定时轮询** 更新数据 — 前端定期请求后端获取最新数据

### 性能与监控
- **D-13:** 使用 **Prometheus + Grafana** 作为监控方案 — 行业标准，生态完善
- **D-14:** 使用 **Redis** 作为缓存方案 — 用于 API 响应缓存和热点数据缓存
- **D-15:** 使用 **Loki + Grafana** 作为日志方案 — 轻量级日志聚合，与 Grafana 集成良好
- **D-16:** **两者结合** — CI/CD 集成自动化安全扫描（GoSec + SonarQube + OWASP ZAP）+ 定期手动安全审计

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 需求文档
- `.planning/REQUIREMENTS.md` — 完整的功能需求规格，包括所有 10 个核心功能模块
- `.planning/ROADMAP.md` — 开发路线图，阶段 3 的详细任务分解（3.1-3.4 节）

### 后端 API 参考
- `pkg/router/router.go` — 完整的 API 路由定义，包括用户端和管理端端点
- `internal/models/models.go` — 核心数据模型（User, Supplier, ExternalModel, UsageRecord 等）
- `internal/models/membership.go` — 会员体系数据模型（MembershipTier, MembershipDiscount, UserMembership）
- `internal/models/model_mapping.go` — 模型映射配置（ModelMapping）

### 前端技术栈文档
- [Vue 3 官方文档](https://vuejs.org/) — Vue 3 框架参考
- [shadcn-vue 文档](https://www.shadcn-vue.com/) — UI 组件库参考
- [Vite 文档](https://vitejs.dev/) — 构建工具参考
- [ECharts 文档](https://echarts.apache.org/) — 图表库参考

### 监控与日志
- [Prometheus 文档](https://prometheus.io/docs/) — 监控系统参考
- [Grafana 文档](https://grafana.com/docs/) — 可视化仪表盘参考
- [Loki 文档](https://grafana.com/docs/loki/latest/) — 日志聚合系统参考

## Existing Code Insights

### Reusable Assets
- **后端 API 层**：`pkg/router/router.go` 已定义完整的 RESTful API，包括 `/v1/` 路由组和 `/admin/` 管理路由
- **数据模型**：`internal/models/` 已实现所有核心数据模型，前端可直接对应 TypeScript 接口
- **认证中间件**：`internal/middleware/auth.go` 已实现 JWT 认证，前端需要配合 Token 管理

### Established Patterns
- **Fiber 框架**：后端使用 Fiber v3 框架，API 响应格式统一
- **GORM 数据模型**：所有模型使用 GORM 标签，前端可据此生成 TypeScript 类型
- **xid.New().String()** — 用于生成唯一 ID 的模式

### Integration Points
- **API 基础路径**：`http://localhost:8080/v1/`（开发环境）
- **认证端点**：用户登录 `/v1/user/login`，管理员登录 `/v1/admin/login`
- **WebSocket/SSE**：当前未实现，运维大屏使用轮询方式
- **文件存储**：账单导出需要配置 S3/OSS 对象存储

## Specific Ideas

### 前端项目结构
```
frontend/
├── user/                 # 用户端 (Vite + Vue 3 + TypeScript)
│   ├── src/
│   │   ├── components/   # 页面组件
│   │   ├── views/        # 页面视图
│   │   ├── router/       # Vue Router 配置
│   │   ├── stores/       # Pinia 状态管理
│   │   ├── api/          # API 客户端（封装 axios）
│   │   ├── types/        # TypeScript 类型定义
│   │   └── main.ts
│   └── package.json
├── admin/                # 运维端 (Vite + Vue 3 + TypeScript)
│   ├── src/
│   │   ├── components/
│   │   ├── views/
│   │   ├── router/
│   │   ├── stores/
│   │   ├── api/
│   │   ├── types/
│   │   └── main.ts
│   └── package.json
└── shared/               # 共享代码
    ├── components/       # 通用 UI 组件
    ├── types/            # 共享类型定义
    └── utils/            # 通用工具函数
```

### 用户端页面清单
- 登录/注册页面
- 控制台概览（Dashboard）
- API Key 管理（创建、编辑、删除、禁用、统计）
- 使用统计和图表（ECharts 可视化）
- 账单查询和导出（下载链接列表）
- 个人设置（基本信息、安全设置）
- 套餐购买和管理

### 运维端页面清单
- 管理员登录页面
- 用户管理（列表、详情、状态管理）
- 模型管理（对外模型配置、路由管理）
- 供应商管理（供应商配置、API Key 管理）
- 套餐管理（套餐配置、用户套餐）
- 账单管理（查看所有账单）
- 系统监控（服务状态、性能指标）
- 运维大屏（实时数据展示）
- 运维日志（操作日志、审计日志）

### API 客户端封装模式
```typescript
// api/client.ts
import axios from 'axios'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/v1',
  timeout: 30000,
})

// 请求拦截器 - 添加认证 Token
client.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器 - 统一错误处理
client.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      // Token 过期，跳转登录
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
```

## Deferred Ideas

None — discussion stayed within phase scope.

---

*Phase: 03-高级功能开发*
*Context gathered: 2026-05-08*
