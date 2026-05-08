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
	app.Get("/health", func(c fiber.Ctx) error {
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

	// ========== Chat API ==========
	v1.Post("/chat/completions", handler.ChatCompletions)
	v1.Post("/completions", handler.Completions)

	// ========== Images API (per D-03) ==========
	v1.Post("/images/generations", handler.CreateImage)
	v1.Post("/images/edits", handler.CreateImageEdit)
	v1.Post("/images/variations", handler.CreateImageVariation)

	// ========== Audio API (per D-03) ==========
	v1.Post("/audio/speech", handler.CreateSpeech)
	v1.Post("/audio/transcriptions", handler.CreateTranscription)
	v1.Post("/audio/translations", handler.CreateTranslation)

	// ========== Embeddings API ==========
	v1.Post("/embeddings", handler.CreateEmbedding)

	// ========== Models API ==========
	v1.Get("/models", handler.ListModels)

	// ========== User Management ==========
	v1.Get("/usage", handler.GetUsage)
	v1.Get("/bills", handler.GenerateBill)

	// ========== Admin API (per FR-006) ==========
	admin := v1.Group("/admin")
	// TODO: 添加管理员权限中间件
	// admin.Use(middleware.RequireAdmin())

	// Model management
	admin.Post("/models", handler.CreateModel)
	admin.Put("/models/:id", handler.UpdateModel)
	admin.Delete("/models/:id", handler.DeleteModel)
	admin.Get("/models", handler.AdminListModels)

	// Supplier management
	admin.Post("/suppliers", handler.CreateSupplier)
	admin.Put("/suppliers/:id", handler.UpdateSupplier)
	admin.Delete("/suppliers/:id", handler.DeleteSupplier)
	admin.Get("/suppliers", handler.ListSuppliers)
}
