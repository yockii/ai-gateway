package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/gateway"
	"github.com/yockii/ai-gateway/internal/middleware"
	"github.com/yockii/ai-gateway/pkg/handlers"
)

// Setup 设置路由
func Setup(app *fiber.App, gw *gateway.Gateway) {
	// 创建 API 处理器
	handler := handlers.New(gw)

	// 全局中间件
	app.Use(middleware.Recovery())
	app.Use(middleware.Logger())
	app.Use(middleware.ErrorHandler())

	// 健康检查端点（不需要认证）
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"service": "ai-gateway",
		})
	})

	// API v1 路由组
	v1 := app.Group("/v1")

	// 认证中间件
	v1.Use(middleware.Auth())

	// 限流中间件
	v1.Use(middleware.RateLimit())

	// 聊天完成接口
	v1.Post("/chat/completions", handler.ChatCompletions)

	// 文本完成接口
	v1.Post("/completions", handler.Completions)

	// 模型列表接口
	v1.Get("/models", handler.ListModels)

	// 使用记录接口
	v1.Get("/usage", handler.GetUsage)

	// 账单接口
	v1.Get("/bills", handler.GenerateBill)
}
