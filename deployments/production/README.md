# AI Gateway 生产环境部署指南

## 目录结构

```
production/
├── docker-compose.prod.yml    # 生产环境 Docker Compose 配置
├── .env.example                # 环境变量模板
├── deploy.sh                   # 部署脚本
├── rollback.sh                 # 回滚脚本
├── nginx/                      # Nginx 配置
│   ├── nginx.conf              # 主配置文件
│   ├── ssl/                    # SSL 证书目录
│   └── logs/                   # 日志目录
├── monitoring/                 # 监控配置
│   ├── prometheus.yml          # Prometheus 配置
│   ├── grafana/                # Grafana 配置
│   └── rules/                  # 告警规则
└── backups/                    # 备份目录
```

## 部署前准备

### 1. 生成安全密钥

```bash
# 生成加密密钥 (32字符)
openssl rand -hex 32

# 生成 JWT 密钥 (32字符)
openssl rand -hex 32

# 生成管理员 JWT 密钥 (32字符)
openssl rand -hex 32

# 生成数据库密码
openssl rand -base64 24

# 生成 Redis 密码
openssl rand -base64 24
```

### 2. 配置环境变量

```bash
cd deployments/production
cp .env.example .env
nano .env  # 或使用其他编辑器
```

填写以下必要配置：
- `ENCRYPTION_KEY`: 加密密钥
- `JWT_SECRET`: JWT 密钥
- `ADMIN_JWT_SECRET`: 管理员 JWT 密钥
- `POSTGRES_PASSWORD`: 数据库密码
- `REDIS_PASSWORD`: Redis 密码
- `GRAFANA_ADMIN_PASSWORD`: Grafana 管理员密码

### 3. 配置 SSL 证书 (可选)

将 SSL 证书放在 `nginx/ssl/` 目录：
- `cert.pem`: 证书文件
- `key.pem`: 私钥文件

## 部署流程

### 方式一：使用部署脚本 (推荐)

```bash
cd deployments/production
chmod +x deploy.sh rollback.sh
./deploy.sh
```

按照提示选择部署类型：
1. 完整部署 (构建 + 启动)
2. 仅重启 (使用已有镜像)
3. 仅更新配置

### 方式二：手动部署

```bash
cd deployments/production

# 1. 构建镜像
docker-compose -f docker-compose.prod.yml build

# 2. 启动服务
docker-compose -f docker-compose.prod.yml up -d

# 3. 查看状态
docker-compose -f docker-compose.prod.yml ps

# 4. 查看日志
docker-compose -f docker-compose.prod.yml logs -f
```

## 服务访问

部署成功后，可以通过以下地址访问服务：

| 服务 | 地址 | 说明 |
|------|------|------|
| API | http://localhost:8080 | 后端 API |
| 管理端 | http://localhost:5174 | 运维管理界面 |
| 用户端 | http://localhost:5173 | 用户应用界面 |
| Prometheus | http://localhost:9091 | 监控指标 |
| Grafana | http://localhost:3000 | 监控面板 |

## 健康检查

```bash
# 检查后端健康状态
curl http://localhost:8080/health

# 检查容器状态
docker-compose -f docker-compose.prod.yml ps

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f ai-gateway
```

## 数据备份

### 自动备份脚本

```bash
# 创建备份目录
mkdir -p backups

# 备份数据库
docker exec ai-gateway-postgres-prod pg_dump -U postgres ai_gateway > backups/db-$(date +%Y%m%d-%H%M%S).sql

# 备份配置文件
tar czf backups/config-$(date +%Y%m%d-%H%M%S).tar.gz .env nginx/ monitoring/
```

### 定时备份 (Cron)

```bash
# 编辑 crontab
crontab -e

# 添加每日凌晨 2 点备份
0 2 * * * cd /path/to/deployments/production && ./backup.sh
```

## 回滚

如果新版本出现问题，可以快速回滚：

```bash
cd deployments/production
./rollback.sh
```

按照提示选择要回滚到的备份版本。

## 监控告警

### Prometheus 告警规则

告警规则位于 `monitoring/rules/` 目录，包括：
- 服务可用性告警
- 响应时间告警
- 错误率告警
- 资源使用告警

### Grafana 面板

导入预配置的仪表板：
- 系统概览
- API 性能
- 数据库性能
- 错误追踪

## 故障排查

### 容器无法启动

```bash
# 查看容器日志
docker-compose -f docker-compose.prod.yml logs [service_name]

# 检查容器资源使用
docker stats
```

### 数据库连接失败

```bash
# 检查数据库状态
docker exec ai-gateway-postgres-prod pg_isready -U postgres

# 查看数据库日志
docker logs ai-gateway-postgres-prod
```

### 性能问题

```bash
# 查看资源使用情况
docker stats

# 进入容器调试
docker exec -it ai-gateway-app-prod sh

# 查看进程状态
docker top ai-gateway-app-prod
```

## 安全建议

1. **定期更新**: 及时更新 Docker 镜像和依赖
2. **密钥轮换**: 定期更换加密密钥和 JWT 密钥
3. **访问控制**: 限制容器网络访问，使用防火墙
4. **日志审计**: 定期检查访问日志和异常行为
5. **备份验证**: 定期验证备份的完整性和可恢复性

## 联系支持

如遇到问题，请查看：
- 项目文档: `../../README.md`
- 问题追踪: GitHub Issues
- 技术支持: support@example.com
