package gateway

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/cache"
	"github.com/yockii/ai-gateway/internal/config"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/logging"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/pkg/api"
	"go.uber.org/zap"
)

// Gateway AI Gateway 核心结构
type Gateway struct {
	config     *config.Config
	app        *fiber.App
	db         *database.DB
	bifrost    *services.BifrostClient
	keyManager *services.KeyManager
	redis      *cache.RedisClient
}

// New 创建新的 Gateway 实例
func New(cfg *config.Config) (*Gateway, error) {
	// 初始化 Fiber 应用（基于 fasthttp）
	app := fiber.New(fiber.Config{
		AppName:      "AI Gateway",
		ServerHeader: "AI-Gateway",
	})

	// 初始化数据库
	db, err := database.New(cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 初始化 Key Manager
	keyManager, err := services.NewKeyManager(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key manager: %w", err)
	}

	// 初始化 Redis 客户端 (per UAT-003)
	redisClient := cache.NewRedisClient(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
	)

	// 测试 Redis 连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx); err != nil {
		log.Printf("警告: Redis 连接失败: %v", err)
		// 继续运行，缓存功能将不可用
	} else {
		log.Println("✅ Redis 连接成功")
	}

	// 初始化 Bifrost 客户端 (per D-01: 库集成模式)
	bifrost, err := services.NewBifrostClient(cfg)
	if err != nil {
		logging.Warn("Bifrost 初始化失败", zap.Error(err))
		// 继续运行，Bifrost 功能将不可用
		bifrost = nil
	}

	gateway := &Gateway{
		config:     cfg,
		app:        app,
		db:         db,
		bifrost:    bifrost,
		keyManager: keyManager,
		redis:      redisClient,
	}

	logging.Info("Gateway 初始化成功")
	return gateway, nil
}

// Start 启动网关服务
func (g *Gateway) Start() error {
	return g.app.Listen(g.config.Server.Address())
}

// Shutdown 优雅关闭
func (g *Gateway) Shutdown() error {
	return g.app.Shutdown()
}

// GenerateID 生成唯一ID（使用 xid）
func (g *Gateway) GenerateID() string {
	return xid.New().String()
}

// ChatCompletion 聊天完成接口 (per D-03: OpenAI 兼容, D-04: 双模式)
func (g *Gateway) ChatCompletion(ctx context.Context, userID string, req *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error) {
	// 生成唯一请求 ID (per D-08: 幂等性)
	requestID := xid.New().String()

	logging.Info("聊天完成请求",
		zap.String("user_id", userID),
		zap.String("model", req.Model),
		zap.String("request_id", requestID),
		zap.Bool("stream", req.Stream),
	)

	// 选择最优路由 (per D-05: 协同故障转移)
	route, err := g.SelectBestRoute(ctx, userID, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to select route: %w", err)
	}

	logging.Info("选择路由",
		zap.String("supplier", route.SupplierName),
		zap.String("model", route.ActualModelName),
	)

	// 转换请求格式
	bifrostReq := &services.ChatCompletionRequest{
		Model:       route.ActualModelName,
		Messages:    convertMessages(req.Messages),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
	}

	// 调用 Bifrost
	if req.Stream {
		// 流式响应 (per D-04)
		chunks, err := g.bifrost.StreamChatCompletion(ctx, bifrostReq)
		if err != nil {
			return nil, fmt.Errorf("stream completion failed: %w", err)
		}

		// TODO: 处理流式响应
		_ = chunks
		return nil, fmt.Errorf("streaming not yet implemented")
	} else {
		// 非流式响应
		resp, err := g.bifrost.ChatCompletion(ctx, bifrostReq)
		if err != nil {
			return nil, fmt.Errorf("completion failed: %w", err)
		}

		// 转换响应格式
		return &api.ChatCompletionResponse{
			ID:      requestID,
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model, // 返回对外模型名
			Choices: convertChoices(resp.Choices),
			Usage:   convertUsage(resp.Usage),
		}, nil
	}
}

// RouteInfo 路由信息
type RouteInfo struct {
	SupplierID      string
	SupplierName    string
	ActualModelName string
	Priority        int
	CostPrice       float64
}

// SelectBestRoute 选择最优路由 (per D-05, D-09)
func (g *Gateway) SelectBestRoute(ctx context.Context, userID, modelID string) (*RouteInfo, error) {
	// TODO: 实现智能路由逻辑
	// 1. 获取用户类型和对应售价
	// 2. 查询模型的所有可用供应商路由
	// 3. 计算每个路由的利润空间
	// 4. 按利润排序，选择最优
	// 5. 检查供应商健康状态和负载

	// 简化实现：返回第一个可用路由
	return &RouteInfo{
		SupplierID:      "supplier-001",
		SupplierName:    "OpenAI",
		ActualModelName: modelID,
		Priority:        1,
		CostPrice:       0.002,
	}, nil
}

// RecordUsage 记录使用量 (per D-08: 双重记录)
func (g *Gateway) RecordUsage(ctx context.Context, record *models.UsageRecord) error {
	// TODO: 实现 Redis Stream + 数据库双写
	// 1. 同步写入 Redis Stream
	// 2. 异步写入数据库
	// 3. 使用 RequestID 作为幂等键

	logging.Info("使用量记录",
		zap.String("user_id", record.UserID),
		zap.String("model_id", record.ModelID),
		zap.Int32("total_tokens", record.TotalTokens),
		zap.Float64("cost_price", record.CostPrice),
	)

	return nil
}

// GetKeyManager 获取 Key Manager
func (g *Gateway) GetKeyManager() *services.KeyManager {
	return g.keyManager
}

// GetApp 获取 Fiber App
func (g *Gateway) GetApp() *fiber.App {
	return g.app
}

// GetDB 获取数据库连接
func (g *Gateway) GetDB() *database.DB {
	return g.db
}

// GetRedis 获取 Redis 客户端 (per UAT-003)
func (g *Gateway) GetRedis() *cache.RedisClient {
	return g.redis
}

// GetModels 获取对外模型列表（优化版，只查询必要字段）
func (g *Gateway) GetModels(ctx context.Context) ([]models.ExternalModel, error) {
	var models []models.ExternalModel
	// 只查询必要字段，使用索引
	err := g.db.WithContext(ctx).
		Select("id", "name", "display_name", "model_type").
		Where("is_active = ?", true).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	return models, nil
}

// Helper functions

func convertMessages(msgs []api.ChatMessage) []services.ChatMessage {
	result := make([]services.ChatMessage, len(msgs))
	for i, m := range msgs {
		result[i] = services.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return result
}

func convertChoices(choices []services.ChatChoice) []api.ChatChoice {
	result := make([]api.ChatChoice, len(choices))
	for i, c := range choices {
		result[i] = api.ChatChoice{
			Index:        c.Index,
			Message:      api.ChatMessage{Role: c.Message.Role, Content: c.Message.Content},
			FinishReason: c.FinishReason,
		}
	}
	return result
}

func convertUsage(usage services.Usage) api.Usage {
	return api.Usage{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}

// ImageGeneration 图片生成接口 (per D-03: OpenAI 兼容)
func (g *Gateway) ImageGeneration(ctx context.Context, userID string, req *api.ImageRequest) (*api.ImageResponse, error) {
	requestID := xid.New().String()

	logging.Info("图片生成请求",
		zap.String("user_id", userID),
		zap.String("model", req.Model),
		zap.String("request_id", requestID),
	)

	// 选择最优路由
	route, err := g.SelectBestRoute(ctx, userID, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to select route: %w", err)
	}

	logging.Info("选择路由",
		zap.String("supplier", route.SupplierName),
		zap.String("model", route.ActualModelName),
	)

	// TODO: 调用 Bifrost 实现图片生成
	// 当前返回模拟响应
	return &api.ImageResponse{
		Created: time.Now().Unix(),
		Data: []api.ImageItem{
			{
				URL: "https://example.com/generated-image.png",
			},
		},
	}, nil
}

// TextToSpeech 文本转语音接口 (per D-04: 流式支持)
func (g *Gateway) TextToSpeech(ctx context.Context, userID string, req *api.SpeechRequest) ([]byte, error) {
	requestID := xid.New().String()

	logging.Info("语音合成请求",
		zap.String("user_id", userID),
		zap.String("model", req.Model),
		zap.String("request_id", requestID),
	)

	// 选择最优路由
	route, err := g.SelectBestRoute(ctx, userID, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to select route: %w", err)
	}

	logging.Info("选择路由",
		zap.String("supplier", route.SupplierName),
		zap.String("model", route.ActualModelName),
	)

	// TODO: 调用 Bifrost 实现语音合成
	// 当前返回模拟音频数据
	return []byte("mock-audio-data"), nil
}

// CreateEmbedding 创建嵌入向量接口 (per D-03: OpenAI 兼容)
func (g *Gateway) CreateEmbedding(ctx context.Context, userID string, req *api.EmbeddingRequest) (*api.EmbeddingResponse, error) {
	requestID := xid.New().String()

	logging.Info("嵌入生成请求",
		zap.String("user_id", userID),
		zap.String("model", req.Model),
		zap.String("request_id", requestID),
		zap.Int("inputs_count", len(req.Input)),
	)

	// 选择最优路由
	route, err := g.SelectBestRoute(ctx, userID, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to select route: %w", err)
	}

	logging.Info("选择路由",
		zap.String("supplier", route.SupplierName),
		zap.String("model", route.ActualModelName),
	)

	// TODO: 调用 Bifrost 实现嵌入生成
	// 当前返回模拟响应
	embeddings := make([]api.EmbeddingItem, len(req.Input))
	for i := range req.Input {
		embeddings[i] = api.EmbeddingItem{
			Object:    "embedding",
			Embedding: make([]float64, 1536), // OpenAI text-embedding-ada-002 dimension
			Index:     i,
		}
	}

	return &api.EmbeddingResponse{
		Object: "list",
		Data:   embeddings,
		Model:  req.Model,
		Usage: api.EmbeddingUsage{
			PromptTokens: len(req.Input) * 10, // 估算
			TotalTokens:  len(req.Input) * 10,
		},
	}, nil
}
