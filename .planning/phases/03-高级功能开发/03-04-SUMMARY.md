---
phase: 03-高级功能开发
plan: 04
subsystem: 监控和日志系统
tags: [monitoring, logging, metrics, alerts, grafana, prometheus, loki]
completed_date: "2026-05-08T16:00:44Z"
duration_seconds: 1032
wave: 2
dependency_graph:
  requires:
    - 03-01 (数据库集成)
    - 03-02 (成本优化系统)
  provides:
    - id: D-13
      description: "Prometheus metrics 暴露和收集"
    - id: D-15
      description: "结构化日志输出"
    - id: NFR-004
      description: "系统监控能力"
    - id: NFR-005
      description: "告警通知能力"
    - id: NFR-006
      description: "日志聚合和查询"
  affects:
    - 03-05 (缓存系统) - 可通过监控评估缓存效果
    - 04-01 (系统测试) - 监控数据用于性能测试验证
tech_stack:
  added:
    - github.com/prometheus/client_golang (Prometheus metrics client)
    - go.uber.org/zap (结构化日志库)
  patterns:
    - Prometheus metrics 中间件模式
    - 结构化 JSON 日志输出
    - 多渠道告警通知模式
    - Docker Compose 监控栈部署
key_files:
  created:
    - internal/metrics/prometheus.go
    - internal/metrics/metrics.go
    - internal/metrics/prometheus_test.go
    - internal/logging/logger.go
    - internal/logging/logger_test.go
    - internal/alerting/alerts.go
    - internal/alerting/notifiers.go
    - internal/alerting/alerts_test.go
    - pkg/handlers/monitoring.go
    - pkg/handlers/monitoring_test.go
    - pkg/router/fasthttp_adapter.go
    - deployments/monitoring/prometheus.yml
    - deployments/monitoring/alerts.yml
    - deployments/monitoring/loki.yml
    - deployments/monitoring/promtail.yml
    - deployments/monitoring/grafana/provisioning/datasources/prometheus.yml
    - deployments/monitoring/grafana/provisioning/datasources/loki.yml
    - deployments/monitoring/grafana/provisioning/dashboards/dashboards.yml
    - deployments/monitoring/grafana/dashboards/api-gateway.json
    - deployments/monitoring/grafana/dashboards/system-metrics.json
    - deployments/docker-compose.monitoring.yml
  modified:
    - pkg/router/router.go
    - internal/middleware/logger.go
    - cmd/ai-gateway/main.go
    - go.mod
    - go.sum
decisions:
  - "使用 Prometheus 作为主要监控系统"
    rationale: "Prometheus 是云原生监控标准，与 Go 集成良好，支持强大的查询语言"
  - "使用 Loki 进行日志聚合"
    rationale: "Loki 是 Grafana Labs 的轻量级日志聚合系统，与 Grafana 无缝集成，成本低于 ELK"
  - "使用 zap 进行结构化日志"
    rationale: "zap 是高性能的 Go 结构化日志库，支持零内存分配和 JSON 输出"
  - "监控数据通过 Docker Compose 独立部署"
    rationale: "监控系统独立于应用部署，便于升级和维护，不影响主应用性能"
metrics:
  duration_seconds: 1032
  tasks_completed: 6
  tests_created: 6
  commits: 6
---

# Phase 03 Plan 04: 监控和日志系统 Summary

## One-liner
实现了完整的监控和日志系统，包括 Prometheus 指标暴露、Grafana 仪表盘、Loki 日志聚合和告警系统，提供全面的系统可观测性。

## 完成状态

✅ **所有任务已完成**

- Task 1: Prometheus Metrics 暴露
- Task 2: Prometheus 配置和部署
- Task 3: Grafana 仪表盘配置
- Task 4: Loki 日志聚合
- Task 5: 告警系统实现
- Task 6: 运维大屏数据 API

## 关键交付物

### 1. Prometheus 指标系统

**文件**: `internal/metrics/`

实现了完整的 Prometheus metrics 暴露层：

- **HTTP 请求指标**: 请求总数、请求持续时间 (P95/P99)
- **API 业务指标**: API 请求数、Token 使用量、费用统计
- **系统运行时指标**: Goroutine 数量、内存分配、并发连接数、活跃用户数
- **错误指标**: 按类型和位置统计的错误数量

**关键函数**:
- `PrometheusMiddleware()` - 自动记录 HTTP 请求指标
- `RecordAPIRequest()` - 记录 API 请求
- `RecordTokenUsage()` - 记录 Token 使用量
- `RecordCost()` - 记录费用

### 2. Prometheus 配置和部署

**文件**: `deployments/monitoring/prometheus.yml`, `deployments/docker-compose.monitoring.yml`

- 配置了 Prometheus 抓取目标 (API Gateway /metrics 端点)
- 设置了 15 秒的抓取间隔
- 配置了告警规则文件引用
- Docker Compose 一键启动完整监控栈

### 3. Grafana 仪表盘

**文件**: `deployments/monitoring/grafana/dashboards/`

创建了两个预配置仪表盘：

1. **API Gateway Dashboard**:
   - 请求速率 (按方法和端点)
   - P95 延迟仪表盘
   - 按模型分布的请求饼图

2. **System Metrics Dashboard**:
   - Goroutine 数量趋势
   - 内存分配趋势
   - 并发连接数趋势

### 4. Loki 日志聚合

**文件**: `internal/logging/logger.go`, `deployments/monitoring/loki.yml`

- 实现了基于 zap 的结构化日志
- 支持请求 ID 追踪
- JSON 格式输出，便于 Loki 解析
- Promtail 配置自动收集应用日志

### 5. 告警系统

**文件**: `internal/alerting/`

实现了完整的告警系统：

- **告警级别**: Info, Warning, Critical
- **通知渠道**: Email, Webhook, Slack
- **告警规则**:
  - 高错误率 (>5%)
  - 高延迟 (P95 > 1s)
  - 服务不可用
  - 内存使用过高 (>512MB)
  - Goroutine 泄漏 (>1000)

### 6. 运维大屏 API

**文件**: `pkg/handlers/monitoring.go`

提供了 REST API 获取监控数据：

- `GET /admin/monitoring/metrics` - 系统指标 (QPS, 延迟, 错误率, 活跃用户)
- `GET /admin/monitoring/overview` - 指标概览 (总请求数, Token, 费用, 连接数)
- `GET /admin/monitoring/model-metrics` - 模型级别指标
- `GET /admin/monitoring/alerts` - 告警列表
- `GET /admin/monitoring/logs` - 日志查询

## 技术实现亮点

### 1. Fiber/FastHTTP 适配器

创建了 `fasthttp_adapter.go` 来解决 Prometheus 的 net/http handler 与 Fiber 的 fasthttp 基础设施不兼容的问题：

```go
func ServeHTTPAdapter(handler http.Handler) fiber.Handler
```

这个适配器允许在 Fiber 应用中使用标准的 net/http 中间件。

### 2. 自定义指标命名

为了避免与 Go 标准库的运行时指标冲突，我们使用了自定义前缀：

```go
ai_gateway_go_goroutines
ai_gateway_go_memstats_alloc_bytes
```

### 3. 请求 ID 追踪

实现了完整的请求 ID 追踪链路：

```go
ctx = logging.WithRequestID(ctx)
reqID := logging.GetRequestID(ctx)
```

请求 ID 会自动添加到所有日志中，便于问题追踪。

## Deviations from Plan

**无偏离** - 计划完全按照设计执行，所有功能都已实现。

## Known Stubs

**无存根** - 所有功能都已完整实现，没有占位符代码。

## Threat Flags

根据威胁模型分析，本计划实现了以下安全缓解措施：

| Flag | File | Description |
|------|------|-------------|
| T-03-23 | /metrics | Prometheus 部署在内网，/metrics 端点不包含敏感信息 |
| T-03-24 | internal/metrics | Prometheus 指标只读，不接受外部输入 |
| T-03-25 | internal/logging | 日志不包含敏感信息 (API Key、密码) |
| T-03-26 | deployments/monitoring/loki.yml | Loki 配置了日志保留期限 (168h) |
| T-03-27 | internal/alerting/notifiers.go | Webhook 支持 HTTPS，邮件使用 SMTP TLS |
| T-03-28 | deployments/monitoring/grafana/provisioning | Grafana 配置了认证 (admin/admin) |

## 验证结果

### 自动化验证

- ✅ 所有单元测试通过
- ✅ 项目构建成功
- ✅ 代码覆盖率符合要求

### 功能验证

以下功能需要在实际部署环境中验证：

1. Prometheus 成功抓取 /metrics 端点
2. Grafana 成功连接 Prometheus 数据源
3. 仪表盘正确显示指标数据
4. Loki 成功接收和存储日志
5. 告警规则正确加载
6. 监控数据持久化到 Docker volumes

### 部署验证命令

```bash
# 启动监控栈
docker-compose -f deployments/docker-compose.monitoring.yml up -d

# 验证 Prometheus
curl http://localhost:9090/-/healthy

# 验证 Grafana
curl http://localhost:3000/api/health

# 验证 Loki
curl http://localhost:3100/ready

# 查看指标
curl http://localhost:8080/metrics
```

## User Setup Required

根据计划定义，用户需要完成以下配置：

### Grafana 配置

1. **访问 Grafana**: http://localhost:3000
   - 默认用户: admin
   - 默认密码: admin

2. **配置 Prometheus 数据源** (通常已自动配置):
   - Grafana UI → Configuration → Data Sources → Prometheus
   - URL: http://prometheus:9090

3. **导入仪表盘** (通常已自动加载):
   - API Gateway Dashboard
   - System Metrics Dashboard

### Prometheus 环境变量

- `PROMETHEUS_RETENTION`: 数据保留时间 (默认: 15天)
  - 可在 `prometheus.yml` 中配置 `--storage.tsdb.retention.time`

## 下一步建议

1. **生产环境配置**:
   - 修改 Grafana 默认密码
   - 配置邮件服务器用于告警通知
   - 设置 Prometheus 数据保留策略

2. **增强监控**:
   - 添加更多业务指标 (按用户组、按模型)
   - 实现自定义告警规则
   - 集成 AlertManager 进行告警路由

3. **性能优化**:
   - 调整 Prometheus 抓取间隔
   - 配置 Loki 日志采样
   - 优化指标基数

## Commit History

1. `9039d8a` - feat(03-04): implement Prometheus metrics exposure
2. `5fde5f6` - feat(03-04): add Prometheus configuration and deployment
3. `af5291c` - feat(03-04): add Grafana dashboard configuration
4. `3410f79` - feat(03-04): implement Loki log aggregation and structured logging
5. `bb7c3ed` - feat(03-04): implement alerting system
6. `adf7386` - feat(03-04): implement monitoring dashboard APIs
