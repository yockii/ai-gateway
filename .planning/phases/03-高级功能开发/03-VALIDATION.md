# Phase 3: 高级功能开发 - Validation Criteria

**Created:** 2026-05-08
**Phase:** 03-高级功能开发
**Status:** Ready for execution

## Overview

本文档定义 Phase 03 前端界面和后端性能优化的可测试验收标准。每个标准包含：
- Given/When/Then 测试场景
- 性能基准
- UI/UX 验证要求
- 边缘情况覆盖

## Feature 1: 用户端项目初始化 (D-01, D-02, D-03, D-04)

### AC-1.1: Vite 项目创建
**Given:** Node.js 和 npm 已安装
**When:** 在 `frontend/user/` 目录运行 `npm create vite@latest . -- --template vue-ts`
**Then:**
- 项目创建成功
- `package.json` 包含 Vite 依赖
- `tsconfig.json` 配置正确
- `vite.config.ts` 存在
- 运行 `npm run dev` 启动开发服务器

**Test Command:**
```bash
cd frontend/user && npm run dev
curl -s http://localhost:5173 | grep -q "Vite"
```

### AC-1.2: shadcn-vue 初始化
**Given:** Vite 项目已创建
**When:** 运行 `npx shadcn-vue@latest init`
**Then:**
- `components.json` 配置文件创建
- `tailwind.config.js` 更新
- `src/components/ui/` 目录创建
- `src/lib/utils.ts` 工具函数创建
- 样式正确应用

**Test Command:**
```bash
cd frontend/user
test -f components.json
test -d src/components/ui
test -f src/lib/utils.ts
```

### AC-1.3: API 客户端封装
**Given:** axios 已安装
**When:** 创建 `src/api/client.ts`
**Then:**
- axios 实例正确配置
- baseURL 从环境变量读取
- 请求拦截器添加 JWT Token
- 响应拦截器处理 401 错误
- 错误统一处理

**Test Command:**
```bash
cd frontend/user
npm run test:unit -- src/api/client.test.ts
```

### AC-1.4: Pinia 状态管理
**Given:** Pinia 已安装
**When:** 创建 `src/stores/auth.ts`
**Then:**
- 认证状态正确管理
- Token 存储在 localStorage
- 登出时清除状态
- isAuthenticated computed 属性正确

**Test Command:**
```bash
cd frontend/user
npm run test:unit -- src/stores/auth.test.ts
```

## Feature 2: 用户端页面实现 (D-05, D-06, D-07)

### AC-2.1: 登录页面
**Given:** 用户访问 `/login`
**When:** 输入有效的邮箱和密码
**Then:**
- 表单验证正确（邮箱格式、密码长度）
- 登录请求发送到 `/v1/user/login`
- 成功后 Token 存储到 localStorage
- 跳转到 Dashboard
- 错误时显示错误消息

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- login.spec.ts
```

### AC-2.2: Dashboard 页面
**Given:** 已登录用户访问 `/dashboard`
**When:** 页面加载完成
**Then:**
- 显示用户基本信息
- 显示 API Key 数量
- 显示本月使用统计
- 显示本月费用
- 显示快捷操作按钮

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- dashboard.spec.ts
```

### AC-2.3: API Key 管理页面
**Given:** 已登录用户访问 `/api-keys`
**When:** 页面加载完成
**Then:**
- 显示所有 API Key 列表
- Key 值部分隐藏（显示前 8 位）
- 显示每个 Key 的状态、额度、使用量
- 支持创建新 Key
- 支持删除 Key
- 支持禁用 Key
- 创建时显示完整 Key（仅一次）

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- api-keys.spec.ts
```

### AC-2.4: 使用统计页面
**Given:** 已登录用户访问 `/usage`
**When:** 页面加载完成
**Then:**
- ECharts 图表正确渲染
- 显示 Token 使用趋势（折线图）
- 显示费用趋势（柱状图）
- 支持日期范围筛选
- 支持按模型筛选
- 图表响应式缩放

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- usage.spec.ts
```

**Performance:** 图表首次渲染 < 500ms

### AC-2.5: 账单查询页面
**Given:** 已登录用户访问 `/bills`
**When:** 页面加载完成
**Then:**
- 显示所有账单列表
- 显示账单周期、总费用、状态
- 支持查看账单详情
- 支持导出 PDF
- 支持导出 CSV
- 下载链接列表显示（异步生成的文件）

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- bills.spec.ts
```

### AC-2.6: 个人设置页面
**Given:** 已登录用户访问 `/settings`
**When:** 页面加载完成
**Then:**
- 显示用户基本信息表单
- 支持修改姓名、头像
- 支持修改密码
- 表单验证正确
- 保存成功显示提示

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- settings.spec.ts
```

### AC-2.7: 套餐购买页面
**Given:** 已登录用户访问 `/plans`
**When:** 页面加载完成
**Then:**
- 显示所有可用套餐
- 显示套餐名称、价格、功能列表
- 支持购买套餐
- 显示当前套餐状态
- 支持升级套餐

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- plans.spec.ts
```

## Feature 3: 布局组件 (D-05)

### AC-3.1: 侧边栏组件
**Given:** 已登录用户
**When:** 页面加载完成
**Then:**
- 侧边栏显示所有导航项
- 当前页面高亮显示
- 支持折叠/展开
- 折叠后显示图标
- 移动端默认隐藏

**Test Command:**
```bash
cd frontend/user
npm run test:unit -- src/components/layout/Sidebar.test.ts
```

### AC-3.2: 顶部栏组件
**Given:** 已登录用户
**When:** 页面加载完成
**Then:**
- 显示 Logo
- 显示用户菜单
- 支持登出操作
- 响应式布局

**Test Command:**
```bash
cd frontend/user
npm run test:unit -- src/components/layout/TopBar.test.ts
```

### AC-3.3: 路由守卫
**Given:** 未登录用户
**When:** 访问需要认证的页面
**Then:**
- 自动跳转到登录页
- 记录原始访问路径
- 登录后跳转回原页面

**Test Command:**
```bash
cd frontend/user
npm run test:unit -- src/router/index.test.ts
```

## Feature 4: 运维端项目初始化

### AC-4.1: 运维端 Vite 项目
**Given:** 用户端项目已完成
**When:** 在 `frontend/admin/` 创建项目
**Then:**
- 独立的 Vite 项目
- 独立的 API 客户端
- 独立的 Pinia stores
- shadcn-vue 初始化

**Test Command:**
```bash
cd frontend/admin
npm run dev
curl -s http://localhost:5174 | grep -q "Vite"
```

### AC-4.2: 管理员认证
**Given:** 运维端项目
**When:** 管理员登录
**Then:**
- 登录端点为 `/v1/admin/login`
- Token 正确存储
- 权限验证正确
- 登出清除状态

**Test Command:**
```bash
cd frontend/admin
npm run test:e2e -- admin-login.spec.ts
```

## Feature 5: 运维端页面实现 (D-09, D-10, D-11)

### AC-5.1: 用户管理页面
**Given:** 已登录管理员
**When:** 访问 `/admin/users`
**Then:**
- DataTable 显示用户列表
- 支持搜索（邮箱、姓名）
- 支持分页
- 每行显示操作按钮（查看、编辑、禁用、删除）
- 操作确认对话框
- 操作后显示成功/失败提示

**Test Command:**
```bash
cd frontend/admin
npm run test:e2e -- users.spec.ts
```

### AC-5.2: 模型管理页面
**Given:** 已登录管理员
**When:** 访问 `/admin/models`
**Then:**
- 显示所有对外模型
- 支持添加新模型
- 支持编辑模型配置
- 支持启用/禁用模型
- 显示模型类型标签

**Test Command:**
```bash
cd frontend/admin
npm run test:e2e -- models.spec.ts
```

### AC-5.3: 供应商管理页面
**Given:** 已登录管理员
**When:** 访问 `/admin/suppliers`
**Then:**
- 显示所有供应商
- 支持配置供应商 API Key
- 支持测试连接
- 显示供应商状态

**Test Command:**
```bash
cd frontend/admin
npm run test:e2e -- suppliers.spec.ts
```

### AC-5.4: 运维大屏
**Given:** 已登录管理员
**When:** 访问 `/admin/operations`
**Then:**
- 全屏布局（无侧边栏）
- 显示系统关键指标卡片（QPS、延迟、错误率）
- 实时更新（30 秒轮询）
- 页面不可见时暂停轮询
- 恢复可见时继续轮询

**Test Command:**
```bash
cd frontend/admin
npm run test:unit -- src/views/Operations.test.ts
```

**Performance:** 轮询请求 < 100ms，页面渲染 < 200ms

### AC-5.5: 系统监控页面
**Given:** 已登录管理员
**When:** 访问 `/admin/monitoring`
**Then:**
- 显示服务状态
- 显示性能指标
- 显示告警列表
- 支持查看 Grafana 仪表盘（链接）

**Test Command:**
```bash
cd frontend/admin
npm run test:e2e -- monitoring.spec.ts
```

## Feature 6: 共享组件

### AC-6.1: DataTable 组件
**Given:** 需要显示数据列表
**When:** 使用 DataTable 组件
**Then:**
- 支持列配置
- 支持排序
- 支持分页
- 支持行操作
- 支持空状态显示
- 支持加载状态

**Test Command:**
```bash
cd frontend/shared
npm run test:unit -- components/DataTable.test.ts
```

### AC-6.2: ChartContainer 组件
**Given:** 需要显示图表
**When:** 使用 ChartContainer 组件
**Then:**
- ECharts 实例正确创建
- 响应式缩放
- 组件卸载时清理实例
- 支持配置更新
- 支持加载状态

**Test Command:**
```bash
cd frontend/shared
npm run test:unit -- components/ChartContainer.test.ts
```

### AC-6.3: ConfirmDialog 组件
**Given:** 需要确认危险操作
**When:** 使用 ConfirmDialog 组件
**Then:**
- 显示确认消息
- 显示操作类型（删除/禁用等）
- 支持取消
- 支持确认
- 样式区分危险操作

**Test Command:**
```bash
cd frontend/shared
npm run test:unit -- components/ConfirmDialog.test.ts
```

## Feature 7: 性能优化 (D-14)

### AC-7.1: Redis 缓存集成
**Given:** Go 服务启动
**When:** Redis 连接建立
**Then:**
- Redis 客户端正确初始化
- 连接池配置正确
- 支持健康检查
- Redis 降级到内存缓存

**Test Command:**
```bash
go test -v -run TestRedisInit ./internal/cache/...
```

### AC-7.2: 模型列表缓存
**Given:** 模型列表已缓存
**When:** 请求 `/v1/models`
**Then:**
- 缓存命中时返回缓存数据
- 缓存未命中时查询数据库
- TTL 设置为 5 分钟
- 模型变更时清除缓存

**Test Command:**
```bash
go test -v -run TestModelCache ./internal/cache/...
```

**Performance:** 缓存命中响应 < 5ms

### AC-7.3: 用户配额缓存
**Given:** 用户配额已缓存
**When:** 检查 API Key 配额
**Then:**
- Redis INCR 操作计数
- 配额数据缓存 1 分钟
- 请求完成后递减

**Test Command:**
```bash
go test -v -run TestQuotaCache ./internal/cache/...
```

**Performance:** 配额检查 < 5ms

### AC-7.4: 数据库查询优化
**Given:** 用户列表查询
**When:** 执行查询
**Then:**
- 使用索引字段
- 避免 N+1 查询
- 使用预加载 (Preload)
- 分页查询使用 LIMIT/OFFSET

**Test Command:**
```bash
go test -v -run TestQueryOptimization ./internal/services/...
```

**Performance:** 查询时间 < 50ms

### AC-7.5: API 网关性能
**Given:** 网关服务运行
**When:** 发送 10000 并发请求
**Then:**
- 网关处理时间 < 20ms
- 支持 10000+ QPS
- 无内存泄漏
- CPU 使用率 < 80%

**Test Command:**
```bash
wrk -t12 -c400 -d30s http://localhost:8080/v1/chat/completions
```

## Feature 8: 监控系统 (D-13, D-15)

### AC-8.1: Prometheus Metrics 暴露
**Given:** Go 服务运行
**When:** 访问 `/metrics` 端点
**Then:**
- 返回 Prometheus 格式指标
- 包含 QPS 指标
- 包含延迟指标
- 包含错误率指标
- 包含业务指标（用户数、请求数等）

**Test Command:**
```bash
curl -s http://localhost:8080/metrics | grep -q "go_"
```

### AC-8.2: Grafana 仪表盘
**Given:** Grafana 运行
**When:** 访问仪表盘
**Then:**
- API Gateway 仪表盘显示 QPS
- 系统指标仪表盘显示 CPU/内存
- 数据源连接正常
- 告警规则配置

**Test Command:**
```bash
curl -s http://localhost:3000/api/health
```

### AC-8.3: Loki 日志聚合
**Given:** Loki 服务运行
**When:** 应用发送日志
**Then:**
- 结构化日志格式 (JSON)
- 日志包含请求 ID
- 日志包含用户 ID
- 日志包含时间戳
- Loki 正确接收

**Test Command:**
```bash
curl -s http://localhost:3100/ready
```

### AC-8.4: 告警系统
**Given:** Prometheus 告警规则配置
**When:** 触发告警条件
**Then:**
- 告警触发
- 发送通知（邮件/Webhook）
- 告警恢复通知
- 告警历史记录

**Test Command:**
```bash
curl -s http://localhost:9093/api/v1/alerts
```

## Feature 9: 安全加固 (D-16)

### AC-9.1: GoSec 安全扫描
**Given:** Go 代码
**When:** 运行 GoSec
**Then:**
- 无高危漏洞
- 无中危漏洞（或已审核）
- 扫描报告可查看

**Test Command:**
```bash
gosec -fmt=json -out=gosec-report.json ./...
```

### AC-9.2: npm 依赖审计
**Given:** 前端项目
**When:** 运行 npm audit
**Then:**
- 无高危漏洞
- 依赖版本正确

**Test Command:**
```bash
cd frontend/user && npm audit --audit-level=high
cd frontend/admin && npm audit --audit-level=high
```

### AC-9.3: CI/CD 安全集成
**Given:** GitHub Actions 配置
**When:** Push 代码
**Then:**
- 自动运行 GoSec
- 自动运行 npm audit
- 失败时阻止合并
- 报告上传到 GitHub

**Test Command:**
```bash
# 手动触发 workflow
gh workflow run security.yml
```

## UI/UX 验证

### UU-1: 响应式设计
**Test:**
- 桌面 (> 1024px): 侧边栏展开
- 平板 (768px - 1024px): 侧边栏默认折叠
- 移动 (< 768px): 侧边栏隐藏，汉堡菜单

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- responsive.spec.ts
```

### UU-2: 无障碍访问
**Test:**
- 所有交互元素可键盘访问
- 焦点指示器可见
- ARIA 标签正确
- 颜色对比度符合 WCAG AA

**Test Command:**
```bash
cd frontend/user
npx pa11y ci http://localhost:5173
```

### UU-3: 加载状态
**Test:**
- 按钮提交时显示加载状态
- 表格加载时显示骨架屏
- 页面加载时显示 Spinner
- 图表加载时显示占位符

**Test Command:**
```bash
cd frontend/user
npm run test:unit -- components/Loading.test.ts
```

### UU-4: 错误处理
**Test:**
- 网络错误显示友好提示
- 表单验证错误内联显示
- 401 自动跳转登录
- 500 显示错误页面
- Toast 通知自动消失

**Test Command:**
```bash
cd frontend/user
npm run test:unit -- components/ErrorState.test.ts
```

## 性能基准

### PB-1: 前端构建时间
**Target:** 生产构建 < 60 秒
**Test:**
```bash
cd frontend/user
time npm run build
```

### PB-2: 前端首次加载
**Target:** FCP < 1.5s, LCP < 2.5s
**Test:**
```bash
cd frontend/user
npm run build && npm run preview
lighthouse http://localhost:4173 --view
```

### PB-3: 前端包大小
**Target:** gzip 后 < 500KB
**Test:**
```bash
cd frontend/user
npm run build
du -sh dist/assets/*.js
```

### PB-4: API 响应时间
**Target:** p95 < 100ms (缓存命中)
**Test:**
```bash
wrk -t4 -c100 -d30s http://localhost:8080/v1/models
```

## 边缘情况

### EC-1: 网络断开
**Given:** 用户正在使用应用
**When:** 网络断开
**Then:** 显示离线提示
**When:** 网络恢复
**Then:** 自动重试请求

### EC-2: Token 过期
**Given:** 用户已登录
**When:** Token 过期
**Then:** 自动跳转登录页
**显示提示："登录已过期，请重新登录"**

### EC-3: 表单提交失败
**Given:** 用户填写表单
**When:** 提交失败
**Then:** 表单数据保留
**显示错误提示**
**支持重新提交**

### EC-4: 图表无数据
**Given:** 用户访问使用统计
**When:** 没有使用数据
**Then:** 显示空状态占位符
**提示："暂无使用数据"**

### EC-5: DataTable 无数据
**Given:** 用户访问 API Key 管理
**When:** 没有创建 API Key
**Then:** 显示空状态
**显示 "创建 API Key" 按钮**

### EC-6: 轮询页面切换
**Given:** 用户在运维大屏
**When:** 切换到其他标签页
**Then:** 轮询暂停
**When:** 切换回标签页
**Then:** 轮询恢复

### EC-7: Redis 不可用
**Given:** Redis 服务停止
**When:** 应用请求缓存
**Then:** 降级到内存缓存
**记录错误日志**
**功能正常可用**

## 集成测试

### IT-1: 用户注册登录流程
**Scenario:** 注册 → 登录 → Dashboard → 登出
**Steps:**
1. 访问注册页面
2. 填写注册信息
3. 验证邮箱格式
4. 提交注册
5. 跳转到登录页
6. 输入凭据登录
7. 验证跳转到 Dashboard
8. 验证用户信息显示
9. 点击登出
10. 验证跳转到登录页

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- auth-flow.spec.ts
```

### IT-2: API Key 创建使用流程
**Scenario:** 创建 Key → 复制 Key → 使用 Key → 查看统计
**Steps:**
1. 访问 API Key 管理
2. 点击 "创建 API Key"
3. 填写名称和额度
4. 提交创建
5. 复制完整 Key
6. 使用 Key 调用 API
7. 返回查看使用统计

**Test Command:**
```bash
cd frontend/user
npm run test:e2e -- apikey-flow.spec.ts
```

### IT-3: 管理员用户管理流程
**Scenario:** 查看用户 → 禁用用户 → 启用用户
**Steps:**
1. 管理员登录
2. 访问用户管理
3. 搜索用户
4. 查看用户详情
5. 禁用用户
6. 确认禁用
7. 验证用户状态更新
8. 启用用户

**Test Command:**
```bash
cd frontend/admin
npm run test:e2e -- user-management.spec.ts
```

## 测试执行顺序

### Wave 1: 用户端基础 (03-01)
- AC-1.1 到 AC-1.4: 项目初始化
- AC-2.1 到 AC-2.3: 登录、Dashboard、API Key
- AC-3.1 到 AC-3.3: 布局组件

### Wave 2: 用户端完整 (03-01)
- AC-2.4 到 AC-2.7: 使用统计、账单、设置、套餐
- AC-6.1 到 AC-6.3: 共享组件

### Wave 3: 运维端 (03-02)
- AC-4.1 到 AC-4.2: 运维端初始化
- AC-5.1 到 AC-5.5: 管理页面和运维大屏

### Wave 4: 性能和监控 (03-03, 03-04)
- AC-7.1 到 AC-7.5: 性能优化
- AC-8.1 到 AC-8.4: 监控系统
- AC-9.1 到 AC-9.3: 安全加固

## 验收清单

在 `/gsd-verify-work` 之前：

- [ ] 所有 AC-1.x 测试通过（用户端初始化）
- [ ] 所有 AC-2.x 测试通过（用户端页面）
- [ ] 所有 AC-3.x 测试通过（布局组件）
- [ ] 所有 AC-4.x 测试通过（运维端初始化）
- [ ] 所有 AC-5.x 测试通过（运维端页面）
- [ ] 所有 AC-6.x 测试通过（共享组件）
- [ ] 所有 AC-7.x 测试通过（性能优化）
- [ ] 所有 AC-8.x 测试通过（监控系统）
- [ ] 所有 AC-9.x 测试通过（安全加固）
- [ ] 所有性能基准达标（PB-1 到 PB-4）
- [ ] 所有 UI/UX 验证通过（UU-1 到 UU-4）
- [ ] 所有边缘情况处理（EC-1 到 EC-7）
- [ ] 所有集成测试通过（IT-1 到 IT-3）
- [ ] 前端代码覆盖率 > 70%
- [ ] 后端代码覆盖率 > 80%
- [ ] 无高危安全漏洞
- [ ] 文档已更新

---

**Phase:** 03-高级功能开发
**Validation criteria created:** 2026-05-08
**Ready for execution:** Yes
