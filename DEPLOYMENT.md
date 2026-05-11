# 部署说明

## 快速启动

### 完整开发环境（推荐）
包含所有服务：后端 API、前端、数据库、Redis、监控服务

```bash
docker-compose -f deployments/docker/docker-compose.full.yml up -d --build
```

访问地址：
- 管理端前端：http://localhost:5174
- 用户端前端：http://localhost:5173
- 后端 API：http://localhost:8080
- Grafana 监控：http://localhost:3000 (admin/admin)
- Prometheus：http://localhost:9091

### 核心服务（仅后端）
只启动数据库、Redis 和后端 API

```bash
docker-compose up -d --build
```

## 服务配置

### 端口映射

| 服务 | 容器端口 | 主机端口 |
|------|----------|----------|
| ai-gateway | 8080 | 8080 |
| postgres | 5432 | 5432 |
| redis | 6379 | 6379 |
| admin-frontend | 80 | 5174 |
| user-frontend | 80 | 5173 |
| prometheus | 9090 | 9091 |
| grafana | 3000 | 3000 |

### 默认账号

**管理员账号**
- 邮箱：admin@example.com
- 密码：admin123456

**Grafana**
- 用户名：admin
- 密码：admin

## 目录结构

```
ai_gateway/
├── docker-compose.yml              # 核心服务（根目录，快速启动）
├── deployments/
│   ├── docker/
│   │   ├── docker-compose.base.yml    # 基础服务配置
│   │   ├── docker-compose.dev.yml     # 开发环境
│   │   ├── docker-compose.staging.yml # 测试环境
│   │   ├── docker-compose.prod.yml    # 生产环境
│   │   └── docker-compose.full.yml    # 完整开发环境（含监控）
│   ├── docker-compose.monitoring.yml  # 监控服务配置
│   └── monitoring/
│       ├── prometheus.yml
│       └── grafana/
```

## 环境变量

后端服务主要环境变量：

```bash
# 数据库
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ai_gateway

# Redis
REDIS_ADDR=redis:6379

# 服务器
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 日志
LOG_LEVEL=debug
ENV=development
```

## 常用命令

```bash
# 启动所有服务
docker-compose -f deployments/docker/docker-compose.full.yml up -d

# 查看服务状态
docker-compose -f deployments/docker/docker-compose.full.yml ps

# 查看日志
docker-compose -f deployments/docker/docker-compose.full.yml logs -f ai-gateway

# 停止所有服务
docker-compose -f deployments/docker/docker-compose.full.yml down

# 重新构建并启动
docker-compose -f deployments/docker/docker-compose.full.yml up -d --build
```

## 故障排查

### 容器无法启动
```bash
# 查看详细日志
docker logs <container-name>

# 检查端口占用
netstat -ano | findstr :8080
```

### 数据库连接失败
确保 postgres 容器已启动并健康：
```bash
docker ps | grep postgres
docker exec ai-gateway-postgres pg_isready -U postgres
```

### Redis 连接失败
```bash
docker exec ai-gateway-redis redis-cli ping
```

### 前端无法访问后端
检查 nginx 配置中的 proxy_pass 地址是否正确。
