package admin

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/pkg/handlers"
)

// TestSupplierApiKeyCRUD 测试供应商 API Key CRUD 操作
func TestSupplierApiKeyCRUD(t *testing.T) {
	db, err := database.NewTestDB()
	if err != nil {
		t.Skip("Skipping test: database not available", err)
		return
	}

	// 创建测试应用
	app := fiber.New()
	handler := &handlers.Handler{}

	// 初始化服务
	supplierApiKeyService, err := services.NewSupplierApiKeyService(db)
	require.NoError(t, err)
	supplierModelService, err := services.NewSupplierModelService(db)
	require.NoError(t, err)
	auditService, err := services.NewAuditService(db)
	require.NoError(t, err)

	handler.SetSupplierApiKeyService(supplierApiKeyService)
	handler.SetSupplierModelService(supplierModelService)
	handler.SetAuditService(auditService)

	// 添加测试中间件设置上下文
	app.Use(func(c fiber.Ctx) error {
		c.Locals("admin_id", "test-admin-1")
		c.Locals("admin_name", "test-admin-1")
		return c.Next()
	})

	// 设置路由
	api := app.Group("/api/v1/admin")
	api.Get("/suppliers/:id/api-keys", handler.ListSupplierApiKeys)
	api.Post("/suppliers/:id/api-keys", handler.CreateSupplierApiKey)
	api.Delete("/suppliers/:id/api-keys/:kid", handler.DeleteSupplierApiKey)
	api.Patch("/suppliers/:id/api-keys/:kid/set-primary", handler.SetPrimarySupplierApiKey)
	api.Get("/suppliers/:id/models", handler.ListSupplierModels)
	api.Post("/suppliers/:id/models", handler.AddSupplierModel)
	api.Put("/suppliers/:id/models/:mid", handler.UpdateSupplierModelCost)
	api.Delete("/suppliers/:id/models/:mid", handler.RemoveSupplierModel)
	api.Get("/audit/logs", handler.ListAuditLogs)

	// 创建测试供应商
	supplierModel := &models.Supplier{
		ID:          "test-supplier-1",
		Name:        "openai",
		DisplayName: "OpenAI",
		Provider:    "openai",
		IsActive:    true,
	}
	err = db.Create(supplierModel).Error
	require.NoError(t, err)

	t.Run("Create API Key", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":         "Test Key",
			"api_key":      "sk-test123456789",
			"priority":     1,
			"is_primary":   true,
			"max_requests": 1000,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/admin/suppliers/test-supplier-1/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("List API Keys", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/suppliers/test-supplier-1/api-keys", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "data")
	})

	// 清理
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM price_history")
	db.Exec("DELETE FROM supplier_models")
	db.Exec("DELETE FROM supplier_api_keys")
	db.Exec("DELETE FROM suppliers")
}

// TestSupplierModelManagement 测试供应商模型管理
func TestSupplierModelManagement(t *testing.T) {
	db, err := database.NewTestDB()
	if err != nil {
		t.Skip("Skipping test: database not available", err)
		return
	}

	app := fiber.New()
	handler := &handlers.Handler{}

	supplierModelService, err := services.NewSupplierModelService(db)
	require.NoError(t, err)
	auditService, err := services.NewAuditService(db)
	require.NoError(t, err)

	handler.SetSupplierModelService(supplierModelService)
	handler.SetAuditService(auditService)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("admin_id", "test-admin-2")
		c.Locals("admin_name", "test-admin-2")
		return c.Next()
	})

	api := app.Group("/api/v1/admin")
	api.Post("/suppliers/:id/models", handler.AddSupplierModel)
	api.Get("/suppliers/:id/models", handler.ListSupplierModels)
	api.Put("/suppliers/:id/models/:mid", handler.UpdateSupplierModelCost)
	api.Delete("/suppliers/:id/models/:mid", handler.RemoveSupplierModel)

	supplier := &models.Supplier{
		ID:          "test-supplier-2",
		Name:        "anthropic",
		DisplayName: "Anthropic",
		Provider:    "anthropic",
		IsActive:    true,
	}
	err = db.Create(supplier).Error
	require.NoError(t, err)

	t.Run("Add Model", func(t *testing.T) {
		payload := map[string]interface{}{
			"model_id":    "claude-3-opus",
			"input_cost":  0.015,
			"output_cost": 0.075,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/admin/suppliers/test-supplier-2/models", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	// 清理
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM price_history")
	db.Exec("DELETE FROM supplier_models")
	db.Exec("DELETE FROM suppliers")
}

// TestAuditLogRecording 测试审计日志记录
func TestAuditLogRecording(t *testing.T) {
	db, err := database.NewTestDB()
	if err != nil {
		t.Skip("Skipping test: database not available", err)
		return
	}

	app := fiber.New()
	handler := &handlers.Handler{}

	supplierModelService, err := services.NewSupplierModelService(db)
	require.NoError(t, err)
	auditService, err := services.NewAuditService(db)
	require.NoError(t, err)

	handler.SetSupplierModelService(supplierModelService)
	handler.SetAuditService(auditService)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("admin_id", "test-admin-audit")
		c.Locals("admin_name", "test-admin-audit")
		return c.Next()
	})

	api := app.Group("/api/v1/admin")
	api.Post("/suppliers/:id/models", handler.AddSupplierModel)
	api.Get("/audit/logs", handler.ListAuditLogs)

	supplier := &models.Supplier{
		ID:          "test-supplier-3",
		Name:        "openai",
		DisplayName: "OpenAI",
		Provider:    "openai",
		IsActive:    true,
	}
	err = db.Create(supplier).Error
	require.NoError(t, err)

	t.Run("Audit Log Created on Model Addition", func(t *testing.T) {
		payload := map[string]interface{}{
			"model_id":    "gpt-4",
			"input_cost":  0.03,
			"output_cost": 0.06,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/admin/suppliers/test-supplier-3/models", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		req.Header.Set("User-Agent", "test-agent")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// 等待审计日志异步写入
		time.Sleep(100 * time.Millisecond)

		// 检查审计日志
		var auditLogs []models.AuditLog
		err = db.Where("entity_type = ?", "model").Find(&auditLogs).Error
		require.NoError(t, err)
		assert.Greater(t, len(auditLogs), 0, "Audit log should be created")
	})

	// 清理
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM supplier_models")
	db.Exec("DELETE FROM suppliers")
}
