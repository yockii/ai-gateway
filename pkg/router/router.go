package router

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yockii/ai-gateway/internal/auth"
	"github.com/yockii/ai-gateway/internal/gateway"
	"github.com/yockii/ai-gateway/internal/metrics"
	"github.com/yockii/ai-gateway/internal/middleware"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/pkg/handlers"
)

// Setup 设置路由
func Setup(app *fiber.App, gw *gateway.Gateway) {
	// 创建 API 处理器
	handler := handlers.New(gw)

	// 初始化管理员服务（使用 Redis token 管理）
	adminService := services.NewAdminService(gw.GetDB(), gw.GetRedis())
	handler.SetAdminService(adminService)

	// 初始化用户服务（使用 Redis token 管理）
	userService := services.NewUserService(gw.GetDB(), gw.GetRedis())
	handler.SetUserService(userService)

	// 创建 token manager 用于中间件
	tokenManager := auth.NewTokenManager(gw.GetRedis())

	// 全局中间件
	app.Use(middleware.Recovery())
	app.Use(middleware.Logger())
	app.Use(middleware.ErrorHandler())

	// Prometheus metrics 中间件
	app.Use(metrics.PrometheusMiddleware())

	// 健康检查端点（不需要认证）
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "ai-gateway",
		})
	})

	// Metrics 端点（不需要认证）
	app.Get("/metrics", ServeHTTPAdapter(promhttp.Handler()))

	// ========== 公开 API ==========
	public := app.Group("/v1/public")

	// 管理员登录（不需要认证）
	public.Post("/admin/login", handler.AdminLogin)

	// 用户登录（不需要认证）
	public.Post("/user/login", handler.UserLogin)

	// 用户注册（不需要认证）
	public.Post("/user/register", handler.UserRegister)

	// ========== Admin API（使用 Token 认证）==========
	admin := app.Group("/v1/admin")
	admin.Use(middleware.RequireAdmin(tokenManager))

	// 用户管理
	admin.Get("/users", handler.ListUsers)

	// 管理员信息
	admin.Get("/me", handler.GetMe)

	// 管理员管理（仅限已登录管理员）
	admin.Post("/admins", handler.CreateAdmin)

	// Model management
	admin.Post("/models", handler.CreateModel)
	admin.Put("/models/:id", handler.UpdateModel)
	admin.Delete("/models/:id", handler.DeleteModel)
	admin.Get("/models", middleware.CacheMiddleware(gw.GetRedis(), 15*time.Minute), handler.AdminListModels)

	// Supplier management
	admin.Post("/suppliers", handler.CreateSupplier)
	admin.Put("/suppliers/:id", handler.UpdateSupplier)
	admin.Delete("/suppliers/:id", handler.DeleteSupplier)
	admin.Get("/suppliers", middleware.CacheMiddleware(gw.GetRedis(), 15*time.Minute), handler.ListSuppliers)

	// Plan management
	admin.Post("/plans", handler.CreatePlan)
	admin.Put("/plans/:id", handler.UpdatePlan)
	admin.Delete("/plans/:id", handler.DeletePlan)
	admin.Get("/plans", middleware.CacheMiddleware(gw.GetRedis(), 15*time.Minute), handler.ListPlans)

	// ========== Monitoring API (per 03-04) ==========
	monitoringHandler := handlers.NewMonitoringHandler("http://prometheus:9090")

	admin.Get("/monitoring/metrics", monitoringHandler.GetSystemMetrics)
	admin.Get("/monitoring/overview", monitoringHandler.GetMetricsOverview)
	admin.Get("/monitoring/model-metrics", monitoringHandler.GetModelMetrics)
	admin.Get("/monitoring/alerts", monitoringHandler.GetAlerts)
	admin.Get("/monitoring/logs", monitoringHandler.GetLogs)

	// ========== 用户 API（使用 Token 认证）==========
	userJWT := app.Group("/v1/user")
	userJWT.Use(middleware.RequireUser(tokenManager))

	// 用户信息
	userJWT.Get("/me", handler.GetUserMe)

	// 使用统计
	userJWT.Get("/usage/stats", handler.GetUsageStats)
	userJWT.Get("/usage", handler.GetUserUsage)

	// API Key 管理
	userJWT.Get("/keys", handler.ListUserKeys)
	userJWT.Post("/keys", handler.CreateUserKey)
	userJWT.Delete("/keys/:id", handler.DeleteUserKey)
	userJWT.Patch("/keys/:id/disable", handler.DisableUserKey)
	userJWT.Patch("/keys/:id/enable", handler.EnableUserKey)
	userJWT.Get("/keys/:id/stats", handler.GetUserKeyStats)

	// 账单管理
	userJWT.Get("/bills", handler.ListUserBills)
	userJWT.Get("/bills/:id", handler.GetUserBill)
	userJWT.Post("/bills/:id/export", handler.ExportUserBill)

	// 套餐管理
	userJWT.Get("/plans/current", handler.GetUserCurrentPlan)
	userJWT.Post("/plans/:id/subscribe", handler.SubscribePlan)
	userJWT.Post("/plans/cancel", handler.CancelPlan)

	// ========== 公共 API（不需要认证）==========
	public.Get("/plans", handler.ListPublicPlans)

	// ========== 用户 API（使用 API Key 认证 - 兼容旧版）==========
	v1 := app.Group("/v1")

	// 认证中间件（API Key）
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
	// 缓存模型列表 (per UAT-003: 30分钟缓存)
	v1.Get("/models", middleware.CacheMiddleware(gw.GetRedis(), 30*time.Minute), handler.ListModels)

	// ========== User Management ==========
	v1.Get("/usage", handler.GetUsage)
	v1.Get("/bills", handler.GenerateBill)
}
