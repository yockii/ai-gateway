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
	"github.com/yockii/ai-gateway/internal/supplier"
	"github.com/yockii/ai-gateway/pkg/handlers"
)

func Setup(app *fiber.App, gw *gateway.Gateway) {
	handler := handlers.New(gw)
	adminService := services.NewAdminService(gw.GetDB(), gw.GetRedis())
	handler.SetAdminService(adminService)
	userService := services.NewUserService(gw.GetDB(), gw.GetRedis())
	handler.SetUserService(userService)
	tokenManager := auth.NewTokenManager(gw.GetRedis())
	
	app.Use(middleware.Recovery())
	app.Use(middleware.Logger())
	app.Use(middleware.ErrorHandler())
	app.Use(metrics.PrometheusMiddleware())
	
	// Phase 5: 初始化定价服务
	membershipService, _ := services.NewMembershipService(gw.GetDB())
	enterpriseService, _ := services.NewEnterprisePricingService(gw.GetDB())
	pricingService, _ := services.NewPricingService(gw.GetDB(), membershipService, enterpriseService)
	handler.SetEnterpriseService(enterpriseService)
	handler.SetPricingService(pricingService)
	
	// Phase 6: 初始化供应商服务
	supplierApiKeyService, _ := services.NewSupplierApiKeyService(gw.GetDB())
	supplierModelService, _ := services.NewSupplierModelService(gw.GetDB())
	handler.SetSupplierApiKeyService(supplierApiKeyService)
	handler.SetSupplierModelService(supplierModelService)
	
	// Phase 7: 初始化 KeyManager
	keyManager, _ := services.NewKeyManager(gw.GetDB())
	handler.SetKeyManager(keyManager)
	
	// Phase 8: 初始化供应商管理器和健康检查器
	supplierManager, _ := supplier.NewManager(gw.GetDB())
	handler.SetSupplierManager(supplierManager)
	// Phase 9: 初始化审计日志服务
	auditService, _ := services.NewAuditService(gw.GetDB())
	handler.SetAuditService(auditService)
	
	
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "ai-gateway"})
	})
	app.Get("/metrics", ServeHTTPAdapter(promhttp.Handler()))
	
	// 公开 API
	public := app.Group("/v1/public")
	public.Post("/admin/login", handler.AdminLogin)
	public.Post("/user/login", handler.UserLogin)
	public.Post("/user/register", handler.UserRegister)
	public.Get("/plans", handler.ListPublicPlans)
	
	// Admin API
	admin := app.Group("/v1/admin")
	admin.Use(middleware.RequireAdmin(tokenManager))
	
	admin.Get("/users", handler.ListUsers)
	admin.Get("/me", handler.GetMe)
	admin.Post("/admins", handler.CreateAdmin)
	
	admin.Post("/models", handler.CreateModel)
	admin.Put("/models/:id", handler.UpdateModel)
	admin.Delete("/models/:id", handler.DeleteModel)
	admin.Get("/models", middleware.CacheMiddleware(gw.GetRedis(), 15*time.Minute), handler.AdminListModels)
	
	admin.Post("/suppliers", handler.CreateSupplier)
	admin.Put("/suppliers/:id", handler.UpdateSupplier)
	admin.Delete("/suppliers/:id", handler.DeleteSupplier)
	admin.Get("/suppliers", middleware.CacheMiddleware(gw.GetRedis(), 15*time.Minute), handler.ListSuppliers)
	
	admin.Post("/plans", handler.CreatePlan)
	admin.Put("/plans/:id", handler.UpdatePlan)
	admin.Delete("/plans/:id", handler.DeletePlan)
	admin.Get("/plans", middleware.CacheMiddleware(gw.GetRedis(), 15*time.Minute), handler.ListPlans)
	
	// Phase 5: 大客户定价管理
	admin.Get("/enterprise-pricing", handler.ListEnterprisePricing)
	admin.Post("/enterprise-pricing", handler.CreateEnterprisePricing)
	admin.Get("/enterprise-pricing/:id", handler.GetEnterprisePricing)
	admin.Put("/enterprise-pricing/:id", handler.UpdateEnterprisePricing)
	admin.Delete("/enterprise-pricing/:id", handler.DeleteEnterprisePricing)
	
	// Phase 6: 供应商 API Key 管理
	admin.Get("/suppliers/:id/api-keys", handler.ListSupplierApiKeys)
	admin.Post("/suppliers/:id/api-keys", handler.CreateSupplierApiKey)
	admin.Put("/suppliers/:id/api-keys/:kid", handler.UpdateSupplierApiKey)
	admin.Delete("/suppliers/:id/api-keys/:kid", handler.DeleteSupplierApiKey)
	admin.Patch("/suppliers/:id/api-keys/:id/set-primary", handler.SetPrimarySupplierApiKey)
	admin.Post("/suppliers/:id/api-keys/:id/rotate", handler.RotateSupplierApiKey)
	admin.Get("/suppliers/:id/api-keys/:kid/stats", handler.GetSupplierApiKeyStats)
	
	// Phase 6: 供应商模型关联管理
	admin.Get("/suppliers/:id/models", handler.ListSupplierModels)
	admin.Post("/suppliers/:id/models", handler.AddSupplierModel)
	admin.Put("/suppliers/:id/models/:mid", handler.UpdateSupplierModelCost)
	admin.Delete("/suppliers/:id/models/:mid", handler.RemoveSupplierModel)
	admin.Get("/suppliers/:id/models/:mid/history", handler.GetSupplierModelPriceHistory)
	
	// Phase 8: 健康状态和故障转移管理
	admin.Get("/suppliers/health", handler.GetAllSuppliersHealth)
	admin.Get("/suppliers/:id/health", handler.GetSupplierHealth)
	admin.Post("/suppliers/:id/health/check", handler.TriggerHealthCheck)
	admin.Get("/suppliers/:id/health/history", handler.GetHealthCheckHistory)
	admin.Get("/events/failures", handler.GetSupplierFailureEvents)
	admin.Get("/events/failovers", handler.GetFailoverEvents)
		// Phase 9: 审计日志
		admin.Get("/audit/logs", handler.ListAuditLogs)
		admin.Get("/audit/logs/:id", handler.GetAuditLogDetail)
		admin.Get("/audit/history", handler.GetEntityAuditHistory)
	
	monitoringHandler := handlers.NewMonitoringHandler("http://prometheus:9090")
	admin.Get("/monitoring/metrics", monitoringHandler.GetSystemMetrics)
	admin.Get("/monitoring/overview", monitoringHandler.GetMetricsOverview)
	admin.Get("/monitoring/model-metrics", monitoringHandler.GetModelMetrics)
	admin.Get("/monitoring/alerts", monitoringHandler.GetAlerts)
	admin.Get("/monitoring/logs", monitoringHandler.GetLogs)
	
	// User JWT API
	userJWT := app.Group("/v1/user")
	userJWT.Use(middleware.RequireUser(tokenManager))
	
	userJWT.Get("/me", handler.GetUserMe)
	userJWT.Get("/usage/stats", handler.GetUsageStats)
	userJWT.Get("/usage", handler.GetUserUsage)
	
	// Phase 7: 用户 API Key 管理
	userJWT.Get("/keys", handler.ListUserKeys)
	userJWT.Post("/keys", handler.CreateUserKey)
	userJWT.Get("/keys/:id", handler.GetUserKey)
	userJWT.Put("/keys/:id", handler.UpdateUserKey)
	userJWT.Delete("/keys/:id", handler.DeleteUserKey)
	userJWT.Patch("/keys/:id/disable", handler.DisableUserKey)
	userJWT.Patch("/keys/:id/enable", handler.EnableUserKey)
	userJWT.Get("/keys/:id/stats", handler.GetUserKeyStats)
	
	userJWT.Get("/bills", handler.ListUserBills)
	userJWT.Get("/bills/:id", handler.GetUserBill)
	userJWT.Post("/bills/:id/export", handler.ExportUserBill)
	
	userJWT.Get("/plans/current", handler.GetUserCurrentPlan)
	userJWT.Post("/plans/:id/subscribe", handler.SubscribePlan)
	userJWT.Post("/plans/cancel", handler.CancelPlan)
	
	// Phase 5: 用户价格查询
	userJWT.Get("/pricing", handler.GetUserPricing)
	
	// OpenAI 兼容 API - 使用 KeyManager 认证
	v1 := app.Group("/v1")
	v1.Use(middleware.Auth(keyManager))
	v1.Use(middleware.RateLimit())
	
	v1.Post("/chat/completions", handler.ChatCompletions)
	v1.Post("/completions", handler.Completions)
	v1.Post("/images/generations", handler.CreateImage)
	v1.Post("/images/edits", handler.CreateImageEdit)
	v1.Post("/images/variations", handler.CreateImageVariation)
	v1.Post("/audio/speech", handler.CreateSpeech)
	v1.Post("/audio/transcriptions", handler.CreateTranscription)
	v1.Post("/audio/translations", handler.CreateTranslation)
	v1.Post("/embeddings", handler.CreateEmbedding)
	v1.Get("/models", middleware.CacheMiddleware(gw.GetRedis(), 30*time.Minute), handler.ListModels)
	v1.Get("/usage", handler.GetUsage)
	v1.Get("/bills", handler.GenerateBill)
}
