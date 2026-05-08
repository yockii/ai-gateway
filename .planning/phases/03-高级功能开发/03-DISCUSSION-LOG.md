# Phase 03: 高级功能开发 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-08
**Phase:** 03-高级功能开发
**Areas discussed:** 前端项目架构, 用户端 UI 设计, 运维端 UI 设计, 性能与监控

---

## 1. 前端项目架构 - 项目结构

| Option | Description | Selected |
|--------|-------------|----------|
| 完全分离 | user-portal/ 和 admin-portal/ 分别独立的 Git 仓库，完全独立部署 | |
| Monorepo | 在同一仓库中，frontend/ 下有 user/ 和 admin/ 两个子项目，共享部分代码 | ✓ |
| 单项目分离 | 同一个 Vue 项目，通过路由和权限控制区分用户端和运维端 | |

**User's choice:** Monorepo
**Notes:** 用户选择了 Monorepo 结构，保持代码在一个仓库中管理

---

## 2. 前端项目架构 - 共享代码

| Option | Description | Selected |
|--------|-------------|----------|
| 完整共享层 | shared/ 中放通用组件、工具函数、API 客户端、类型定义 | |
| 仅 UI 共享 | 仅共享 UI 组件和类型定义，API 客户端各自独立 | ✓ |
| 最小共享 | 不设置共享目录，需要时通过 npm link 或文件引用 | |

**User's choice:** 仅 UI 共享
**Notes:** API 客户端各自独立，仅共享 UI 组件和类型定义

---

## 3. 前端项目架构 - 构建工具

| Option | Description | Selected |
|--------|-------------|----------|
| Vite | Vite - 更快的开发体验，原生 ESM 支持 | ✓ |
| Vue CLI | Vue CLI - 成熟稳定，生态完善 | |
| Nuxt 3 | Nuxt 3 - 全栈 Vue 框架，支持 SSR | |

**User's choice:** Vite
**Notes:** 选择 Vite 以获得更快的开发体验

---

## 4. 前端项目架构 - API 客户端

| Option | Description | Selected |
|--------|-------------|----------|
| 原生 axios | 直接使用 axios，手动管理请求/响应/错误 | |
| 封装 axios | 基于 axios 封装，统一处理认证、错误、重试 | ✓ |
| Vue 请求库 | 使用专用的 Vue 请求库（如 vue-request 或 useRequest） | |

**User's choice:** 封装 axios
**Notes:** 统一处理认证、错误、重试逻辑

---

## 5. 用户端 UI 设计 - 布局风格

| Option | Description | Selected |
|--------|-------------|----------|
| 侧边栏布局 | 左侧固定导航 + 顶部内容区，类似 Vercel/GitHub | |
| 顶部导航 | 顶部横向导航 + 内容区，类似 Google/Linear | |
| 经典管理后台 | 侧边栏（可折叠）+ 顶部栏 + 内容区，经典管理后台风格 | ✓ |

**User's choice:** 经典管理后台
**Notes:** 选择经典的三栏布局

---

## 6. 用户端 UI 设计 - 数据可视化

| Option | Description | Selected |
|--------|-------------|----------|
| ECharts | Apache ECharts - 功能强大，中文文档完善 | ✓ |
| Chart.js | Chart.js - 轻量级，API 简洁 | |
| Vue-ECharts | Vue-ECharts - ECharts 的 Vue 3 封装组件 | |

**User's choice:** ECharts
**Notes:** 选择 ECharts 进行数据可视化

---

## 7. 用户端 UI 设计 - 表单验证

| Option | Description | Selected |
|--------|-------------|----------|
| Zod | Zod - TypeScript-first，与 Vue-Tailwind 完美配合 | |
| VeeValidate | VeeValidate - Vue 专用，功能完善 | |
| shadcn-vue 内置 | shadcn-vue 内置的表单验证 | ✓ |

**User's choice:** shadcn-vue 内置
**Notes:** 使用 shadcn-vue 内置的表单验证

---

## 8. 用户端 UI 设计 - 账单导出

| Option | Description | Selected |
|--------|-------------|----------|
| 前端生成 | 前端生成，使用 jsPDF + SheetJS | |
| 后端生成 | 后端生成 PDF/CSV，前端直接下载 | |
| 混合方式 | 前端 CSV（简单），后端 PDF（复杂） | |

**User's choice:** 后端异步生成，存在s3/oss中，在界面上给出生成时间、下载连接的列表供用户下载
**Notes:** 用户选择了自定义方案 - 后端异步生成 + 对象存储 + 下载链接列表

---

## 9. 运维端 UI 设计 - 管理界面布局

| Option | Description | Selected |
|--------|-------------|----------|
| 与用户端一致 | 与用户端相同的经典布局，保持一致性 | |
| 紧凑高效 | 更紧凑的布局，强调数据密度和操作效率 | |
| 运维大屏 | 专门的运维大屏布局，支持自定义仪表盘 | |

**User's choice:** 管理部分与用户端一致，但也需要一个运维大屏来实时展示各类数据指标
**Notes:** 用户选择了混合方案 - 管理界面保持一致，同时需要运维大屏

---

## 10. 运维端 UI 设计 - 数据表格

| Option | Description | Selected |
|--------|-------------|----------|
| Tanstack Table | Tanstack Table - 无头组件，完全自定义 | |
| shadcn-vue DataTable | shadcn-vue 的 DataTable 组件 | ✓ |
| VxeTable | VxeTable - 功能强大的 Vue 表格组件 | |

**User's choice:** shadcn-vue DataTable
**Notes:** 使用 shadcn-vue 的 DataTable 组件

---

## 11. 运维端 UI 设计 - 批量操作

| Option | Description | Selected |
|--------|-------------|----------|
| 选择后批量操作 | 表格行多选 + 顶部批量操作栏 | |
| 逐行操作 | 每行独立操作按钮，支持快速操作 | ✓ |
| 两者都支持 | 两种模式都支持，根据场景选择 | |

**User's choice:** 逐行操作
**Notes:** 选择逐行操作而非批量选择

---

## 12. 运维端 UI 设计 - 实时监控

| Option | Description | Selected |
|--------|-------------|----------|
| WebSocket | WebSocket 实时推送，服务端主动推送数据 | |
| 定时轮询 | 定时轮询（polling），前端定期请求数据 | ✓ |
| SSE | Server-Sent Events (SSE)，单向推送 | |

**User's choice:** 定时轮询
**Notes:** 运维大屏使用定时轮询更新数据

---

## 13. 性能与监控 - 监控工具

| Option | Description | Selected |
|--------|-------------|----------|
| Prometheus + Grafana | Prometheus + Grafana - 行业标准，生态完善 | ✓ |
| 自研监控 | 自研轻量级监控 - 基于 Go metrics + 简单仪表盘 | |
| 云服务监控 | 云服务监控 - 如阿里云 CloudMonitor / AWS CloudWatch | |

**User's choice:** Prometheus + Grafana
**Notes:** 选择行业标准方案

---

## 14. 性能与监控 - 缓存策略

| Option | Description | Selected |
|--------|-------------|----------|
| Redis | Redis - 完整的缓存功能，支持多种数据结构 | ✓ |
| 内存缓存 | 内存缓存 - 进程内缓存，简单快速 | |
| 两级缓存 | Redis + 内存缓存两级 - 热数据内存，温数据 Redis | |

**User's choice:** Redis
**Notes:** 选择 Redis 作为缓存方案

---

## 15. 性能与监控 - 日志系统

| Option | Description | Selected |
|--------|-------------|----------|
| ELK Stack | ELK Stack (Elasticsearch + Logstash + Kibana) | |
| Loki + Grafana | Loki + Grafana - 轻量级日志聚合 | ✓ |
| 结构化日志 | 结构化日志 + 文件存储 - 简单实用 | |

**User's choice:** Loki + Grafana
**Notes:** 选择 Loki + Grafana，与监控系统统一使用 Grafana

---

## 16. 性能与监控 - 安全加固

| Option | Description | Selected |
|--------|-------------|----------|
| 自动化扫描 | 自动化安全扫描 - GoSec + SonarQube + OWASP ZAP | |
| 手动审计 | 手动安全审计 - 代码审查 + 渗透测试 | |
| 两者结合 | 两者结合 - CI/CD 集成扫描 + 定期手动审计 | ✓ |

**User's choice:** 你推荐着来 → 推荐两者结合
**Notes:** 推荐 CI/CD 集成自动化扫描 + 定期手动审计

---

## Claude's Discretion

**安全加固方案** - 用户询问推荐，基于最佳实践推荐"两者结合"方案（CI/CD 集成自动化扫描 + 定期手动安全审计），这是最全面的安全保障方案。

## Deferred Ideas

None — discussion stayed within phase scope.
