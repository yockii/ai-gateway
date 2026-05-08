# AI Gateway

统一的 AI 模型网关，支持 OpenAI、Claude、Qwen 等多个 AI 提供商，提供智能路由、成本优化和 OpenAI 兼容 API。

## 特性

- ✅ **多供应商支持** - 集成多个 AI 提供商（OpenAI、Claude、Qwen 等）
- ✅ **智能路由** - 自动选择最优供应商，平衡成本和性能
- ✅ **成本优化** - 实时成本控制和利润率验证
- ✅ **OpenAI 兼容** - 完全兼容 OpenAI API 格式
- ✅ **高性能** - 基于 Gofiber v3，支持 10000+ QPS
- ✅ **灵活定价** - 支持按用户群体的差异化定价

## 技术栈

- **后端**: Go 1.26.2
- **Web 框架**: Gofiber v3 (基于 fasthttp)
- **数据库**: PostgreSQL + GORM
- **AI 引擎**: Bifrost v1.5.8
- **ID 生成**: xid (性能优于 UUID)

## 快速开始

### 前置要求

- Go 1.26.2+
- PostgreSQL 12+
- Git

### 安装

```bash
# 克隆项目
git clone https://github.com/yockii/ai-gateway.git
cd ai-gateway

# 安装依赖
go mod download

# 复制配置文件
cp .env.example .env

# 编辑配置文件
# 修改数据库连接信息等配置
```

### 配置

编辑 `.env` 文件：

```bash
# 服务器配置
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ai_gateway

# 成本优化配置
DEFAULT_PROFIT_MARGIN=0.1
MIN_PROFIT_MARGIN=0.05
```

### 运行

```bash
# 开发模式运行
go run ./cmd/ai-gateway/main.go

# 编译后运行
go build -o bin/ai-gateway.exe ./cmd/ai-gateway
./bin/ai-gateway.exe
```

### 验证

```bash
# 健康检查
curl http://localhost:8080/health

# 测试 API
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer test-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## API 文档

### 聊天完成接口

**请求**:
```http
POST /v1/chat/completions
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "model": "gpt-3.5-turbo",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 100
}
```

**响应**:
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-3.5-turbo",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 9,
    "total_tokens": 19
  }
}
```

### 其他端点

- `GET /v1/models` - 列出可用模型
- `GET /v1/usage` - 获取使用记录
- `GET /v1/bills` - 生成账单
- `POST /v1/completions` - 文本完成接口

## 开发

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test ./... -cover

# 运行性能基准测试
go test ./benchmarks/... -bench=.
```

### 项目结构

```
ai_gateway/
├── cmd/                    # 应用入口
│   ├── ai-gateway/         # 主程序
│   ├── bifrost-test/       # Bifrost 集成测试
│   └── db-test/           # 数据库测试
├── internal/              # 私有代码
│   ├── config/            # 配置
│   ├── database/          # 数据库
│   ├── gateway/           # 网关核心
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   ├── cost/              # 成本优化
│   ├── pricing/           # 定价管理
│   └── supplier/          # 供应商管理
├── pkg/                   # 公共代码
│   ├── api/               # API 类型
│   ├── handlers/          # API 处理器
│   └── router/            # 路由
├── benchmarks/            # 性能测试
├── .planning/             # 项目规划文档
└── README.md
```

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t ai-gateway:latest .

# 运行容器
docker run -d \
  --name ai-gateway \
  -p 8080:8080 \
  -e DB_HOST=your_db_host \
  -e DB_PASSWORD=your_db_password \
  ai-gateway:latest
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | 服务端口 | `8080` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_USER` | 数据库用户 | `postgres` |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_NAME` | 数据库名称 | `ai_gateway` |
| `DEFAULT_PROFIT_MARGIN` | 默认利润率 | `0.1` |
| `MIN_PROFIT_MARGIN` | 最小利润率 | `0.05` |

## 性能

- **响应时间**: < 20ms (不含 AI 提供商)
- **吞吐量**: 10000+ QPS
- **ID 生成**: 34ns/op，0 分配
- **利润率计算**: 0.47ns/op，0 分配

## 贡献

欢迎提交 Pull Request！

## 许可证

MIT License

## 联系方式

- 项目地址: https://github.com/yockii/ai-gateway
- 问题反馈: https://github.com/yockii/ai-gateway/issues
