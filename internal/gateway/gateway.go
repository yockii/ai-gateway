package gateway

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/config"
	"github.com/yockii/ai-gateway/internal/database"
)

// Gateway AI Gateway 核心结构
type Gateway struct {
	config  *config.Config
	app     *fiber.App
	bifrost *bifrost.Bifrost
	db      *database.DB
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

	// TODO: Wave 2 实现 Bifrost 初始化
	// Bifrost API 需要进一步研究
	_ = bifrost.Bifrost{}

	gateway := &Gateway{
		config: cfg,
		app:    app,
		db:     db,
	}

	// 设置路由
	// TODO: Wave 3 实现路由设置
	_ = gateway

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
