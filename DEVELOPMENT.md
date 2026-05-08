# 开发环境指南

## 环境要求

### 必需软件

- **Go**: 1.26.2 或更高版本
  - 下载地址: https://golang.org/dl/
  - 验证安装: `go version`

- **PostgreSQL**: 12 或更高版本
  - 下载地址: https://www.postgresql.org/download/
  - 验证安装: `psql --version`

- **Git**: 最新版本
  - 下载地址: https://git-scm.com/downloads
  - 验证安装: `git --version`

### 可选软件

- **Docker**: 20.10+ (用于容器化部署)
- **Docker Compose**: 2.0+ (用于本地开发)
- **VS Code**: 推荐的 IDE (安装 Go 扩展)

## 设置开发环境

### 1. 克隆项目

```bash
git clone https://github.com/yockii/ai-gateway.git
cd ai-gateway
```

### 2. 安装 Go 依赖

```bash
go mod download
```

### 3. 配置数据库

#### 方式 A: 使用本地 PostgreSQL

```bash
# 创建数据库
createdb ai_gateway

# 或使用 psql
psql -U postgres
CREATE DATABASE ai_gateway;
\q
```

#### 方式 B: 使用 Docker Compose

```bash
# 启动 PostgreSQL
docker-compose up -d postgres

# 查看日志
docker-compose logs -f postgres
```

### 4. 配置环境变量

```bash
# 复制示例配置
cp .env.example .env

# 编辑配置
# Windows: notepad .env
# Linux/Mac: nano .env
```

必需配置项：
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ai_gateway
```

### 5. 初始化数据库

```bash
# 运行数据库迁移测试
go test ./internal/database/... -v

# 或运行主程序（自动迁移）
go run ./cmd/ai-gateway/main.go
```

## 开发工作流

### 运行应用

```bash
# 开发模式（自动重载）
# 安装 air: go install github.com/cosmtrek/air@latest
air

# 或直接运行
go run ./cmd/ai-gateway/main.go
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/cost/... -v

# 运行测试并显示覆盖率
go test ./... -cover

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 性能测试

```bash
# 运行基准测试
go test ./benchmarks/... -bench=. -benchmem

# CPU 性能分析
go test ./benchmarks/... -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 内存性能分析
go test ./benchmarks/... -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

### 代码检查

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 使用 golangci-lint (需要安装)
# 安装: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
golangci-lint run
```

## 调试

### VS Code 调试配置

创建 `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch ai-gateway",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/ai-gateway",
      "env": {
        "DB_HOST": "localhost",
        "DB_PORT": "5432",
        "DB_USER": "postgres",
        "DB_PASSWORD": "postgres",
        "DB_NAME": "ai_gateway"
      },
      "args": []
    }
  ]
}
```

### 日志调试

应用会输出详细日志：
- 请求日志：显示每个请求的详细信息
- 错误日志：显示所有错误
- 性能日志：显示关键操作的耗时

## 常见问题

### 1. 数据库连接失败

```
错误: failed to connect to database: connection refused
解决:
- 检查 PostgreSQL 是否运行
- 检查 .env 中的数据库配置
- 检查防火墙设置
```

### 2. 端口被占用

```
错误: bind: address already in use
解决:
- Windows: netstat -ano | findstr :8080
- Linux/Mac: lsof -i :8080
- 修改 .env 中的 SERVER_PORT
```

### 3. 依赖下载失败

```
错误: go: github.com/xxx: module download failed
解决:
- 设置 Go 代理: export GOPROXY=https://goproxy.cn,direct
- 或: go env -w GOPROXY=https://goproxy.cn,direct
```

## 生产部署

### 编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/ai-gateway-linux ./cmd/ai-gateway

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/ai-gateway.exe ./cmd/ai-gateway

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/ai-gateway-macos ./cmd/ai-gateway
```

### Docker 部署

```bash
# 构建镜像
docker build -t ai-gateway:latest .

# 运行容器
docker run -d \
  --name ai-gateway \
  -p 8080:8080 \
  --env-file .env \
  ai-gateway:latest

# 使用 Docker Compose
docker-compose up -d
```

## 性能优化建议

1. **数据库连接池**
   - 默认配置已优化 (MaxIdleConns=10, MaxOpenConns=100)
   - 根据实际负载调整

2. **并发处理**
   - Gofiber 基于 fasthttp，无需额外配置
   - 确保数据库连接池足够大

3. **内存管理**
   - 使用 xid 代替 UUID，减少内存分配
   - 避免在热路径上使用反射

4. **监控**
   - 使用内置的性能日志
   - 集成 Prometheus 监控（待实现）

## 贡献代码

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `go fmt` 格式化代码
- 使用 `go vet` 检查代码
- 为公共函数添加注释
- 为复杂逻辑添加单元测试

## 获取帮助

- 查看文档: `.planning/` 目录
- 提交问题: https://github.com/yockii/ai-gateway/issues
- 查看示例: `cmd/` 目录下的测试程序
