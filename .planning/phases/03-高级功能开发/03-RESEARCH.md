# Phase 03: 高级功能开发 - Research

**Research Date:** 2026-05-08
**Domain:** Vue 3 前端架构 + Go 性能优化 + 监控系统
**Confidence:** HIGH

## Summary

Phase 03 负责实现前端用户界面和后端性能优化。前端采用 Vue 3 + Vite + shadcn-vue 技术栈，构建用户端和运维端两个独立应用。后端聚焦性能优化（Redis 缓存、数据库查询优化）和监控系统集成（Prometheus + Grafana + Loki）。

**Primary Recommendation:** 使用 Vite 创建两个独立的 Vue 3 项目，通过 shared/ 目录共享 UI 组件和类型定义。axios 封装 JWT 认证逻辑，ECharts 处理数据可视化。性能优化优先使用 Redis 缓存热点数据，监控系统采用 Prometheus + Grafana 标准方案。

## 用户约束 (from CONTEXT.md)

### 锁定决策 (Locked Decisions)

- **D-01:** 采用 Monorepo 结构 — frontend/ 目录下包含 user/ 和 admin/ 两个子项目
- **D-02:** 仅 UI 共享 — shared/ 中放通用 UI 组件和类型定义，API 客户端各自独立实现
- **D-03:** 使用 Vite 作为构建工具 — 更快的开发体验，原生 ESM 支持
- **D-04:** 封装 axios — 基于 axios 封装 API 客户端，统一处理认证、错误、重试逻辑
- **D-05:** 用户端采用经典管理后台布局 — 侧边栏（可折叠）+ 顶部栏 + 内容区
- **D-06:** 使用 ECharts 进行数据可视化 — 用于使用统计和费用趋势图表
- **D-07:** 使用 shadcn-vue 内置表单验证 — 用于注册登录、API Key 创建等表单
- **D-08:** 后端异步生成 + 对象存储 — 账单导出（PDF/CSV）由后端异步生成，存储到 S3/OSS
- **D-09:** 运维端混合布局 — 管理部分保持经典管理后台，同时需要专门的运维大屏
- **D-10:** 使用 shadcn-vue DataTable 组件 — 用于用户列表、模型列表等数据表格
- **D-11:** 采用逐行操作 — 每行独立操作按钮，支持快速操作（而非批量选择）
- **D-12:** 运维大屏使用定时轮询更新数据 — 前端定期请求后端获取最新数据
- **D-13:** 使用 Prometheus + Grafana 作为监控方案 — 行业标准，生态完善
- **D-14:** 使用 Redis 作为缓存方案 — 用于 API 响应缓存和热点数据缓存
- **D-15:** 使用 Loki + Grafana 作为日志方案 — 轻量级日志聚合，与 Grafana 集成良好
- **D-16:** 安全扫描结合方案 — CI/CD 集成自动化安全扫描 + 定期手动安全审计

### Claude 自由裁量

None — 所有决策已锁定。

### 延迟想法 (OUT OF SCOPE)

None — discussion 保持在阶段范围内。

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| 用户注册登录 | Frontend Server (SSR) | Browser | 用户端独立处理认证流程 |
| API Key 管理 | Browser | API / Backend | 前端展示，后端验证和存储 |
| 使用统计图表 | Browser | API / Backend | ECharts 在浏览器渲染，数据来自 API |
| 账单查询和导出 | Browser | API / Backend | 前端查询，后端异步生成文件 |
| 运维大屏数据展示 | Browser | API / Backend | 前端轮询，后端聚合数据 |
| API 网关性能优化 | API / Backend | CDN / Static | 后端优化响应时间 |
| Redis 缓存策略 | API / Backend | Database | 后端管理缓存层 |
| Prometheus 监控 | API / Backend | — | 后端暴露 metrics 端点 |
| Loki 日志聚合 | API / Backend | — | 后端发送结构化日志 |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **Vue** | 3.5.34 | 渐进式前端框架 | 最新稳定版本，Composition API 成熟 |
| **Vite** | 6.0.0 | 构建工具和开发服务器 | 极速 HMR，原生 ESM 支持，Vue 官方推荐 |
| **shadcn-vue** | 2.6.2 | UI 组件库 | 基于 Radix Vue，无障碍支持，可定制性强 |
| **TypeScript** | 5.6.0 | 类型安全 | Vue 3 官方推荐类型系统 |

**Version Verification:** [VERIFIED: npm registry] - 2026-05-08

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **Vue Router** | 5.0.6 | 客户端路由 | 所有单页应用路由管理 |
| **Pinia** | 3.0.4 | 状态管理 | 全局状态（用户信息、认证状态） |
| **axios** | 1.7.9 | HTTP 客户端 | 所有 API 请求，支持拦截器 |
| **zod** | 4.4.3 | Schema 验证 | 表单验证，API 响应类型检查 |
| **ECharts** | 6.0.0 | 数据可视化 | 使用统计图表、运维大屏 |
| **lucide-vue-next** | Latest | 图标库 | shadcn-vue 默认图标库 |
| **tailwindcss** | Latest | CSS 框架 | shadcn-vue 依赖，样式系统 |

**Version Verification:** [VERIFIED: npm registry] - 2026-05-08

### 后端监控栈

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **Prometheus Client Go** | Latest | Go metrics 暴露 | API 网关指标收集 |
| **Grafana** | Latest | 监控仪表盘 | 可视化监控数据 |
| **Loki** | Latest | 日志聚合 | 结构化日志收集和查询 |

**Version Verification:** [ASSUMED] - Go client 库版本需执行时验证

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| shadcn-vue | Element Plus, Ant Design Vue | shadcn-vue 更轻量，代码所有权更高；Element Plus 组件更全但体积大 |
| ECharts | Chart.js, D3.js | ECharts 对中文支持更好，配置更简单；D3.js 更灵活但学习曲线陡 |
| Pinia | Vuex | Pinia 是 Vue 3 官方推荐，API 更简洁；Vuex 适用于 Vue 2 迁移 |
| Redis | Memcached | Redis 数据结构更丰富，支持持久化；Memcached 更简单但功能少 |

**Installation:**

```bash
# 用户端项目初始化
cd frontend/user
npm create vite@latest . -- --template vue-ts
npm install

# shadcn-vue 初始化
npx shadcn-vue@latest init

# 核心依赖
npm install vue-router@5 pinia axios zod echarts
npm install -D @types/node

# 运维端项目（重复相同步骤）
cd frontend/admin
npm create vite@latest . -- --template vue-ts
# ... 重复上述安装步骤
```

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Browser / Client Layer                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  User Portal (Vue 3 + Vite)    │    Admin Portal (Vue 3 + Vite)             │
│  ├── Dashboard & Statistics    │    ├── User Management                     │
│  ├── API Key Management        │    ├── Model & Supplier Management        │
│  ├── Bills & Usage             │    ├── System Monitoring                  │
│  └── Settings                  │    └── Operations Dashboard               │
└─────────────────────────────────────────────────────────────────────────────┘
                                        ↓ HTTP/HTTPS
┌─────────────────────────────────────────────────────────────────────────────┐
│                         API Gateway / Backend Layer                          │
├─────────────────────────────────────────────────────────────────────────────┤
│  Authentication (JWT)         │    Rate Limiting    │    Request Routing    │
│  └── Token Validation         │    └── Redis Cache  │    └── Load Balancer  │
└─────────────────────────────────────────────────────────────────────────────┘
                                        ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Business Logic Layer                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  User Service │ Model Service │ Billing Service │ Admin Service              │
└─────────────────────────────────────────────────────────────────────────────┘
                                        ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Data & Monitoring Layer                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  PostgreSQL  │  Redis Cache  │  Prometheus  │  Loki  │  S3/OSS             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure

```
frontend/
├── user/                          # 用户端
│   ├── src/
│   │   ├── api/                   # API 客户端封装
│   │   │   ├── client.ts          # axios 实例配置
│   │   │   ├── auth.ts            # 认证相关 API
│   │   │   ├── keys.ts            # API Key 管理 API
│   │   │   ├── usage.ts           # 使用统计 API
│   │   │   └── bills.ts           # 账单管理 API
│   │   ├── components/            # 页面组件
│   │   │   ├── layout/            # 布局组件
│   │   │   │   ├── Sidebar.vue    # 侧边栏
│   │   │   │   ├── TopBar.vue     # 顶部栏
│   │   │   │   └── Layout.vue     # 主布局
│   │   │   ├── charts/            # 图表组件
│   │   │   │   ├── UsageChart.vue # 使用趋势图
│   │   │   │   └── CostChart.vue  # 费用趋势图
│   │   │   └── common/            # 通用组件
│   │   │       ├── Loading.vue
│   │   │       └── ErrorState.vue
│   │   ├── views/                 # 页面视图
│   │   │   ├── Login.vue          # 登录页
│   │   │   ├── Register.vue       # 注册页
│   │   │   ├── Dashboard.vue      # 控制台概览
│   │   │   ├── ApiKeys.vue        # API Key 管理
│   │   │   ├── Usage.vue          # 使用统计
│   │   │   ├── Bills.vue          # 账单查询
│   │   │   └── Settings.vue       # 个人设置
│   │   ├── router/                # 路由配置
│   │   │   └── index.ts
│   │   ├── stores/                # Pinia 状态管理
│   │   │   ├── auth.ts            # 认证状态
│   │   │   └── user.ts            # 用户信息
│   │   ├── types/                 # TypeScript 类型
│   │   │   └── api.ts             # API 响应类型
│   │   ├── utils/                 # 工具函数
│   │   │   └── format.ts          # 格式化函数
│   │   ├── App.vue
│   │   └── main.ts
│   ├── public/
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── admin/                         # 运维端
│   ├── src/
│   │   ├── api/                   # 独立的 API 客户端
│   │   │   ├── client.ts
│   │   │   ├── users.ts
│   │   │   ├── models.ts
│   │   │   ├── suppliers.ts
│   │   │   └── monitoring.ts
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   ├── dashboard/         # 运维大屏组件
│   │   │   │   ├── MetricCard.vue
│   │   │   │   └── SystemStatus.vue
│   │   │   └── tables/
│   │   │       ├── UserTable.vue
│   │   │       └── ModelTable.vue
│   │   ├── views/
│   │   │   ├── Login.vue
│   │   │   ├── Users.vue
│   │   │   ├── Models.vue
│   │   │   ├── Suppliers.vue
│   │   │   ├── Plans.vue
│   │   │   ├── Monitoring.vue     # 系统监控
│   │   │   └── Operations.vue     # 运维大屏
│   │   ├── router/
│   │   ├── stores/
│   │   ├── types/
│   │   ├── App.vue
│   │   └── main.ts
│   ├── package.json
│   └── vite.config.ts
└── shared/                        # 共享代码
    ├── components/                # 共享 UI 组件
    │   ├── DataTable.vue          # shadcn-vue DataTable 封装
    │   ├── ChartContainer.vue     # ECharts 容器
    │   └── ConfirmDialog.vue      # 确认对话框
    ├── types/                     # 共享类型定义
    │   ├── models.ts              # 数据模型类型
    │   └── api.ts                 # API 通用类型
    └── utils/                     # 共享工具函数
        ├── date.ts                # 日期格式化
        └── number.ts              # 数字格式化
```

### Pattern 1: Axios 封装与 JWT 认证

**What:** 封装 axios 实例，统一处理认证、错误、重试逻辑

**When to use:** 所有需要与后端 API 通信的场景

**Example:**

```typescript
// frontend/user/src/api/client.ts
import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'

// 创建 axios 实例
const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 添加认证 Token
client.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器 - 统一错误处理
client.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Token 过期，清除本地存储并跳转登录
      localStorage.removeItem('access_token')
      localStorage.removeItem('user_info')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
```

**Source:** [CITED: axios documentation] - 标准拦截器模式

### Pattern 2: ECharts 集成模式

**What:** 在 Vue 3 中使用 ECharts 进行数据可视化

**When to use:** 使用趋势图、费用统计、系统监控仪表盘

**Example:**

```typescript
// frontend/user/src/components/charts/UsageChart.vue
<template>
  <div ref="chartRef" class="w-full h-400px"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

interface Props {
  data: Array<{ date: string; tokens: number; cost: number }>
}

const props = defineProps<Props>()
const chartRef = ref<HTMLDivElement>()
let chartInstance: echarts.ECharts | null = null

const initChart = () => {
  if (!chartRef.value) return

  chartInstance = echarts.init(chartRef.value)

  const option: EChartsOption = {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    legend: {
      data: ['Token 使用量', '费用']
    },
    xAxis: {
      type: 'category',
      data: props.data.map(d => d.date)
    },
    yAxis: [
      {
        type: 'value',
        name: 'Token 使用量',
        position: 'left'
      },
      {
        type: 'value',
        name: '费用 (元)',
        position: 'right'
      }
    ],
    series: [
      {
        name: 'Token 使用量',
        type: 'bar',
        data: props.data.map(d => d.tokens)
      },
      {
        name: '费用',
        type: 'line',
        yAxisIndex: 1,
        data: props.data.map(d => d.cost)
      }
    ]
  }

  chartInstance.setOption(option)
}

onMounted(() => {
  initChart()
  window.addEventListener('resize', () => chartInstance?.resize())
})

onUnmounted(() => {
  chartInstance?.dispose()
  window.removeEventListener('resize', () => chartInstance?.resize())
})

watch(() => props.data, () => {
  chartInstance?.setOption({
    xAxis: { data: props.data.map(d => d.date) },
    series: [
      { data: props.data.map(d => d.tokens) },
      { data: props.data.map(d => d.cost) }
    ]
  })
}, { deep: true })
</script>
```

**Source:** [CITED: ECharts Vue 3 integration guide] - 官方推荐模式

### Pattern 3: shadcn-vue 表单验证

**What:** 使用 zod + shadcn-vue Form 组件实现表单验证

**When to use:** 所有表单场景（登录、注册、API Key 创建）

**Example:**

```typescript
// frontend/user/src/views/ApiKeys.vue
<script setup lang="ts">
import { ref } from 'vue'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'

// 定义验证 schema
const formSchema = toTypedSchema(z.object({
  name: z.string().min(3, '名称至少 3 个字符').max(50, '名称最多 50 个字符'),
  quotaDaily: z.number().min(1, '每日额度至少为 1').optional(),
  concurrencyLimit: z.number().min(1, '并发限制至少为 1').optional(),
}))

const form = useForm({
  validationSchema: formSchema,
})

const onSubmit = form.handleSubmit((values) => {
  // 提交创建 API Key
  createApiKey(values)
})
</script>

<template>
  <Form v-bind="form">
    <form @submit="onSubmit" class="space-y-4">
      <FormField v-slot="{ componentField }" name="name">
        <FormItem>
          <FormLabel>API Key 名称</FormLabel>
          <FormControl>
            <Input placeholder="输入 API Key 名称" v-bind="componentField" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="quotaDaily">
        <FormItem>
          <FormLabel>每日额度</FormLabel>
          <FormControl>
            <Input type="number" placeholder="可选" v-bind="componentField" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <Button type="submit">创建 API Key</Button>
    </form>
  </Form>
</template>
```

**Source:** [CITED: shadcn-vue form documentation] - 标准表单验证模式

### Pattern 4: Pinia 状态管理

**What:** 使用 Pinia 管理全局状态（认证、用户信息）

**When to use:** 需要跨组件共享状态的场景

**Example:**

```typescript
// frontend/user/src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/models'

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(localStorage.getItem('access_token'))
  const user = ref<User | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!token.value)
  const userInfo = computed(() => user.value)

  // Actions
  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('access_token', newToken)
  }

  const setUser = (userData: User) => {
    user.value = userData
    localStorage.setItem('user_info', JSON.stringify(userData))
  }

  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('user_info')
  }

  return {
    token,
    user,
    isAuthenticated,
    userInfo,
    setToken,
    setUser,
    logout,
  }
})
```

**Source:** [CITED: Pinia official documentation] - Composition API 风格

### Pattern 5: 路由守卫

**What:** 使用 Vue Router 守卫保护需要认证的页面

**When to use:** 所有需要登录才能访问的页面

**Example:**

```typescript
// frontend/user/src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/Dashboard.vue'),
      meta: { requiresAuth: true }
    },
    // ... 其他路由
  ],
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // 需要认证但未登录，跳转到登录页
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    // 已登录用户访问登录页，跳转到首页
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
```

**Source:** [CITED: Vue Router navigation guards] - 标准认证守卫模式

### Anti-Patterns to Avoid

- **直接在组件中调用 axios:** 应该通过封装的 API 模块调用，便于统一处理和测试
- **在多个地方重复的认证逻辑:** 应该使用 axios 拦截器统一处理
- **在组件内部直接操作 localStorage:** 应该通过 Pinia store 封装
- **忽略 ECharts 实例清理:** 组件卸载时必须调用 `dispose()` 防止内存泄漏
- **使用 v-if 控制路由:** 应该使用 Vue Router 的路由系统
- **表单验证逻辑分散:** 应该使用 zod schema 集中管理验证规则

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 表单验证 | 自定义验证逻辑 | zod + vee-validate + shadcn-vue Form | 支持复杂嵌套验证、类型推断、国际化 |
| HTTP 请求 | fetch 封装 | axios | 拦截器、取消请求、上传进度、自动 JSON 转换 |
| 图表库 | Canvas/SVG 手写 | ECharts | 丰富的图表类型、交互、响应式、性能优化 |
| 状态管理 | EventBus、provide/inject | Pinia | DevTools 支持、TypeScript 友好、模块化 |
| 日期处理 | 自定义格式化 | date-fns | Tree-shakable、不可变、支持 locale |
| 图标 | 自定义 SVG | lucide-vue-next | 统一风格、按需导入、shadcn-vue 默认 |
| 无限滚动 | 自定义滚动逻辑 | vue-use (useInfiniteScroll) | 处理边界情况、性能优化 |

**Key insight:** 前端生态已经非常成熟，几乎所有常见问题都有高质量解决方案。手写这些功能不仅浪费时间，还容易引入 bug 和性能问题。

## Common Pitfalls

### Pitfall 1: ECharts 实例未正确清理

**What goes wrong:** 组件卸载后 ECharts 实例仍然存在，导致内存泄漏

**Why it happens:** ECharts 实例需要手动调用 `dispose()` 清理，Vue 组件销毁不会自动清理

**How to avoid:**
```typescript
onUnmounted(() => {
  chartInstance?.dispose()
  window.removeEventListener('resize', () => chartInstance?.resize())
})
```

**Warning signs:** 浏览器 DevTools Memory 面板显示内存持续增长

### Pitfall 2: axios 拦截器重复注册

**What goes wrong:** 在热更新时拦截器被多次注册，导致请求被拦截多次

**Why it happens:** Vite HMR 重新执行模块时，旧的拦截器没有被清理

**How to avoid:** 将 axios 实例创建放在模块顶层，确保只有一个实例

```typescript
// ✅ 正确 - 模块级别单例
const client = axios.create({...})
client.interceptors.request.use(...)

// ❌ 错误 - 函数内部创建
export const createClient = () => {
  const client = axios.create({...})
  client.interceptors.request.use(...)  // HMR 时重复注册
  return client
}
```

**Warning signs:** 网络请求显示多个相同的 Authorization header

### Pitfall 3: TypeScript 类型与后端不匹配

**What goes wrong:** 前端类型定义与后端 API 响应不一致，导致运行时错误

**Why it happens:** 前端类型手动维护，后端 API 变更后未同步更新

**How to avoid:**
1. 定期与后端 API 对齐类型定义
2. 使用 zod schema 进行运行时验证
3. 考虑使用工具从 Go 结构体生成 TypeScript 类型

```typescript
// 使用 zod 验证 API 响应
const ResponseSchema = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string().email(),
})

const data = ResponseSchema.parse(apiResponse)
```

**Warning signs:** `undefined is not an object` 错误，属性访问失败

### Pitfall 4: 轮询导致的性能问题

**What goes wrong:** 运维大屏轮询频率过高，导致服务器和浏览器性能下降

**Why it happens:** 多个组件同时轮询，没有统一管理

**How to avoid:**
1. 使用统一的轮询管理器
2. 页面不可见时暂停轮询
3. 合理设置轮询间隔（30 秒）

```typescript
// 使用 Page Visibility API 暂停轮询
document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    stopPolling()
  } else {
    startPolling()
  }
})
```

**Warning signs:** 浏览器标签页切换时卡顿，网络面板显示大量请求

### Pitfall 5: shadcn-vue 组件样式丢失

**What goes wrong:** 生产环境部署后组件样式丢失或显示异常

**Why it happens:** Tailwind CSS 未正确配置，或组件未正确导入

**How to avoid:**
1. 确保 `tailwind.config.js` 包含所有组件路径
2. 检查 `components.json` 配置
3. 使用 `npx shadcn-vue@latest add <component>` 添加组件

```javascript
// tailwind.config.js
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
    // 确保 shadcn-vue 组件路径包含在内
  ],
}
```

**Warning signs:** 开发环境正常，生产环境样式异常

## Code Examples

### API 客户端封装模式

```typescript
// frontend/user/src/api/keys.ts
import client from './client'
import type { UserAPIKey, CreateKeyRequest } from '@/types/models'

export const keysApi = {
  // 获取 API Key 列表
  list: async () => {
    return client.get<{
      data: UserAPIKey[]
    }>('/user/keys')
  },

  // 创建 API Key
  create: async (data: CreateKeyRequest) => {
    return client.post<{
      data: UserAPIKey
      key: string  // 只在创建时返回完整密钥
    }>('/user/keys', data)
  },

  // 删除 API Key
  delete: async (id: string) => {
    return client.delete(`/user/keys/${id}`)
  },

  // 禁用 API Key
  disable: async (id: string) => {
    return client.patch(`/user/keys/${id}/disable`)
  },

  // 获取使用统计
  getStats: async (id: string) => {
    return client.get<{
      data: {
        totalRequests: number
        totalTokens: number
        totalCost: number
      }
    }>(`/user/keys/${id}/stats`)
  },
}
```

**Source:** [CITED: REST API client pattern]

### 数据可视化示例

```typescript
// frontend/admin/src/components/dashboard/SystemMetrics.vue
<template>
  <div class="grid grid-cols-4 gap-4">
    <MetricCard
      v-for="metric in metrics"
      :key="metric.name"
      :title="metric.title"
      :value="metric.value"
      :unit="metric.unit"
      :trend="metric.trend"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import MetricCard from './MetricCard.vue'
import { monitoringApi } from '@/api/monitoring'

interface Metric {
  name: string
  title: string
  value: number
  unit: string
  trend: number  // 正数上升，负数下降
}

const metrics = ref<Metric[]>([])
let pollingTimer: number | null = null

const fetchMetrics = async () => {
  const response = await monitoringApi.getSystemMetrics()
  metrics.value = [
    {
      name: 'qps',
      title: 'QPS',
      value: response.data.qps,
      unit: 'req/s',
      trend: response.data.qpsTrend,
    },
    {
      name: 'latency',
      title: '平均延迟',
      value: response.data.avgLatency,
      unit: 'ms',
      trend: response.data.latencyTrend,
    },
    // ... 更多指标
  ]
}

const startPolling = () => {
  fetchMetrics()
  pollingTimer = window.setInterval(fetchMetrics, 30000)  // 30 秒轮询
}

onMounted(() => {
  startPolling()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  if (pollingTimer) clearInterval(pollingTimer)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

const handleVisibilityChange = () => {
  if (document.hidden) {
    if (pollingTimer) clearInterval(pollingTimer)
  } else {
    startPolling()
  }
}
</script>
```

**Source:** CONTEXT.md D-12 (轮询更新策略)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Vue CLI | Vite | 2022+ | Vite 启动速度提升 10-100 倍，HMR 即时更新 |
| Vuex | Pinia | 2021+ | Pinia API 更简洁，TypeScript 支持更好，Vue 3 官方推荐 |
| Options API | Composition API | 2020+ | 更好的逻辑复用，类型推断，Tree-shaking |
| Element UI | shadcn-vue | 2023+ | 代码所有权更高，定制性强，基于 Radix Vue 无障碍支持 |
| Webpack | Vite (Rollup) | 2020+ | 开发体验大幅提升，生产构建优化 |

**Deprecated/outdated:**
- **Vue 2:** 2023-12-31 停止维护，应使用 Vue 3
- **Webpack:** 在新项目中不推荐，Vite 是现代标准
- **class-style components:** Vue 3 不支持，应使用 Composition API

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Prometheus Client Go 版本为 Latest | Standard Stack | API 不兼容，需升级代码 |
| A2 | 后端已实现 `/user/keys` 等用户端 API | Architecture Patterns | 前端无法正常工作，需后端配合 |
| A3 | 后端已实现 `/admin/models` 等管理端 API | Architecture Patterns | 运维端功能无法正常工作 |
| A4 | S3/OSS 对象存储已配置完成 | CONTEXT.md D-08 | 账单导出功能无法使用 |
| A5 | Redis 缓存将在 Wave 2 实现 | Performance Optimization | 需调整实现顺序 |

**Note:** 后端 API 路由已在 `pkg/router/router.go` 中定义，但部分处理器可能尚未实现。规划时需验证 API 可用性。

## Open Questions

1. **后端 API 完成度**
   - What we know: `pkg/router/router.go` 定义了完整路由
   - What's unclear: 各 API handler 的实现完成度
   - Recommendation: 在 Wave 1 开始前验证后端 API 可用性，必要时同步开发

2. **对象存储配置**
   - What we know: CONTEXT.md D-08 指定使用 S3/OSS
   - What's unclear: 使用哪种对象存储服务（AWS S3、阿里云 OSS、MinIO?）
   - Recommendation: 确认对象存储方案，配置 SDK

3. **监控数据源**
   - What we know: 使用 Prometheus + Grafana
   - What's unclear: Go 应用暴露 metrics 的具体方案（prometheus/client_golang?）
   - Recommendation: 确认 metrics 暴露方案，规划数据采集

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Node.js | 前端构建 | ✓ | v24.13.1 | — |
| npm | 前端包管理 | ✓ | 11.8.0 | — |
| Go | 后端开发 | ✓ | 1.26.2 | — |
| Redis | 缓存层 | ✗ | — | 使用内存缓存（开发环境） |
| PostgreSQL | 数据库 | ✓ | (configured) | — |
| Prometheus | 监控系统 | ✗ | — | 跳过监控功能，标记为待实现 |
| Grafana | 监控可视化 | ✗ | — | 跳过仪表盘，标记为待实现 |
| Loki | 日志聚合 | ✗ | — | 跳过日志聚合，标记为待实现 |

**Missing dependencies with no fallback:**
- 无 — 所有缺失依赖都有临时方案或可延后实现

**Missing dependencies with fallback:**
- **Redis:** 开发环境使用内存缓存，生产环境需要配置 Redis
- **Prometheus/Grafana/Loki:** Wave 1-2 跳过，Wave 4 专门实现监控功能

**Note:** 前端开发不依赖 Redis 和监控系统，可以先行实现。后端性能优化（Wave 3）需要 Redis。

## Validation Architecture

> workflow.nyquist_validation 未在 config.json 中设置，默认为启用

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Vitest (Vite 原生测试框架) |
| Config file | 前端项目创建时配置 |
| Quick run command | `npm run test:unit` |
| Full suite command | `npm run test` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FR-005 | 用户注册登录流程 | integration | `npm run test:e2e user-auth` | ❌ Wave 0 |
| FR-007 | API Key CRUD 操作 | unit | `npm run test:unit api-keys` | ❌ Wave 0 |
| FR-008 | 账单查询和导出 | integration | `npm run test:e2e billing` | ❌ Wave 0 |
| D-04 | axios 拦截器认证 | unit | `npm run test:unit client-interceptors` | ❌ Wave 0 |
| D-12 | 运维大屏轮询 | unit | `npm run test:unit polling` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `npm run test:unit` (< 30 秒)
- **Per wave merge:** `npm run test` (完整套件)
- **Phase gate:** 全部测试通过 + 手动 UI 测试

### Wave 0 Gaps

- [ ] `frontend/user/vitest.config.ts` — Vitest 配置
- [ ] `frontend/user/tests/unit/` — 单元测试目录
- [ ] `frontend/user/tests/e2e/` — E2E 测试目录（使用 Playwright）
- [ ] `frontend/admin/vitest.config.ts` — 运维端测试配置
- [ ] `frontend/shared/tests/unit/` — 共享组件测试
- [ ] 测试工具函数：`tests/utils/test-helpers.ts`

**Framework install:**
```bash
npm install -D vitest @vue/test-utils @playwright/test jsdom
```

## Security Domain

> security_enforcement 未在 config.json 中设置，默认为启用

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes | JWT token 认证，axios 拦截器 |
| V3 Session Management | yes | localStorage 存储token，自动过期处理 |
| V4 Access Control | Partial | 前端路由守卫，后端权限验证 |
| V5 Input Validation | yes | zod schema 验证所有表单输入 |
| V6 Cryptography | yes | HTTPS 传输，密码 bcrypt 加密（后端） |
| V7 Error Handling | yes | 统一错误处理，敏感信息脱敏 |
| V8 Data Protection | yes | API Key 部分显示，敏感数据不缓存 |
| V9 Communication | yes | HTTPS only，CORS 配置 |
| V10 Malicious Code | yes | 依赖扫描（npm audit），SRI |
| V11 Business Logic | yes | 前端不做授权判断，依赖后端 |
| V12 File Handling | yes | 账单导出文件类型验证 |

### Known Threat Patterns for Vue 3 + Go Backend

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| XSS (跨站脚本) | Tampering | Vue 3 自动转义，避免 v-html，CSP 策略 |
| CSRF (跨站请求伪造) | Spoofing | SameSite cookie，CSRF token |
| JWT Token 泄露 | Information Disclosure | localStorage 存储，HTTPS only，短有效期 |
| API Key 暴力破解 | Denial of Service | 后端速率限制，前端防抖 |
| 中间人攻击 | Tampering | 强制 HTTPS，HSTS |
| 敏感数据泄露 | Information Disclosure | 不在前端存储敏感信息，日志脱敏 |

**Security Implementation Notes:**

1. **XSS 防护:** Vue 3 默认转义输出，避免使用 `v-html`。必须使用时进行 sanitize。

2. **CSRF 防护:** 后端实现 CSRF token 验证，前端在请求头中携带。

3. **Content Security Policy:** 配置 CSP 策略限制资源来源。

```typescript
// vite.config.ts - CSP 配置示例
export default defineConfig({
  plugins: [
    csp({
      policy: {
        'default-src': ["'self'"],
        'script-src': ["'self'", "'unsafe-inline'"],
        'style-src': ["'self'", "'unsafe-inline'"],
        'img-src': ["'self'", 'data:', 'https:'],
      }
    })
  ]
})
```

4. **依赖安全扫描:**
```bash
# CI/CD 中运行
npm audit --audit-level=high
npm audit fix
```

## Sources

### Primary (HIGH confidence)

- [Vue 3 Documentation](https://vuejs.org/) - 框架核心概念和 API
- [Vite Documentation](https://vitejs.dev/) - 构建工具配置和插件
- [shadcn-vue Documentation](https://www.shadcn-vue.com/) - UI 组件库使用
- [Pinia Documentation](https://pinia.vuejs.org/) - 状态管理最佳实践
- [Vue Router Documentation](https://router.vuejs.org/) - 路由和导航守卫
- [ECharts Documentation](https://echarts.apache.org/) - 图表配置和交互
- [axios Documentation](https://axios-http.com/) - HTTP 客户端配置
- [zod Documentation](https://zod.dev/) - Schema 验证

### Secondary (MEDIUM confidence)

- [Vue 3 Style Guide](https://vuejs.org/style-guide/) - 官方代码风格指南
- [Testing Library Vue](https://testing-library.com/docs/vue-testing-library/intro/) - 组件测试最佳实践
- [Prometheus Go Client](https://github.com/prometheus/client_golang) - Go metrics 暴露
- [Grafana Documentation](https://grafana.com/docs/) - 监控仪表盘配置
- [Loki Documentation](https://grafana.com/docs/loki/latest/) - 日志聚合系统

### Tertiary (LOW confidence)

- None — 所有技术栈均有官方文档支持

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - 所有版本已通过 npm registry 验证
- Architecture: HIGH - 基于 Vue 3 官方文档和最佳实践
- Pitfalls: HIGH - 常见问题已有社区验证和解决方案
- Performance optimization: MEDIUM - Redis 策略需要实际测试验证
- Monitoring: MEDIUM - Prometheus + Grafana 方案成熟，但具体实现需验证

**Research date:** 2026-05-08
**Valid until:** 30 天（前端技术栈相对稳定，但可能有新的版本更新）
