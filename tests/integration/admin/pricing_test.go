package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// TestEnterprisePricingCRUD 测试企业定价 CRUD 操作
func TestEnterprisePricingCRUD(t *testing.T) {
	db, err := database.NewTestDB()
	if err != nil {
		t.Skip("Skipping test: database not available", err)
		return
	}

	app := fiber.New()
	handler := &handlers.Handler{}

	enterpriseService, err := services.NewEnterprisePricingService(db)
	require.NoError(t, err)
	auditService, err := services.NewAuditService(db)
	require.NoError(t, err)

	handler.SetEnterpriseService(enterpriseService)
	handler.SetAuditService(auditService)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("admin_id", "test-admin-1")
		c.Locals("admin_name", "test-admin-1")
		return c.Next()
	})

	api := app.Group("/api/v1/admin")
	api.Post("/enterprise-pricing", handler.CreateEnterprisePricing)
	api.Get("/enterprise-pricing", handler.ListEnterprisePricing)
	api.Get("/enterprise-pricing/:id", handler.GetEnterprisePricing)
	api.Put("/enterprise-pricing/:id", handler.UpdateEnterprisePricing)
	api.Delete("/enterprise-pricing/:id", handler.DeleteEnterprisePricing)

	t.Run("Create Enterprise Pricing", func(t *testing.T) {
		payload := map[string]interface{}{
			"customer_id":       "enterprise-1",
			"customer_name":     "Acme Corp",
			"model_id":          "gpt-4",
			"input_price":       0.05,
			"output_price":      0.10,
			"min_profit_margin": 10.0,
			"max_cost_price":    0.04,
			"effective_date":    time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/admin/enterprise-pricing", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var result models.EnterprisePricing
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "enterprise-1", result.CustomerID)
		assert.Equal(t, "Acme Corp", result.CustomerName)
		assert.Equal(t, "gpt-4", result.ModelID)
	})

	t.Run("List Enterprise Pricing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/enterprise-pricing", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "data")
	})

	t.Run("Update Enterprise Pricing", func(t *testing.T) {
		pricing := &models.EnterprisePricing{
			CustomerID:       "enterprise-2",
			CustomerName:     "Globex Inc",
			ModelID:          "gpt-4",
			InputPrice:       0.04,
			OutputPrice:      0.08,
			MinProfitMargin:  15.0,
			MaxCostPrice:     0.035,
			EffectiveDate:    time.Now(),
			IsActive:         true,
		}
		err := db.Create(pricing).Error
		require.NoError(t, err)

		payload := map[string]interface{}{
			"input_price":  0.045,
			"output_price": 0.09,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/admin/enterprise-pricing/%s", pricing.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result models.EnterprisePricing
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, 0.045, result.InputPrice)
		assert.Equal(t, 0.09, result.OutputPrice)
	})

	t.Run("Delete Enterprise Pricing", func(t *testing.T) {
		pricing := &models.EnterprisePricing{
			CustomerID:       "enterprise-3",
			CustomerName:     "ToDelete Corp",
			ModelID:          "gpt-4",
			InputPrice:       0.03,
			OutputPrice:      0.06,
			MinProfitMargin:  10.0,
			MaxCostPrice:     0.025,
			EffectiveDate:    time.Now(),
			IsActive:         true,
		}
		err := db.Create(pricing).Error
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/admin/enterprise-pricing/%s", pricing.ID), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	// 清理
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM enterprise_pricing")
}

// TestPricingAuditLog 测试定价操作的审计日志
func TestPricingAuditLog(t *testing.T) {
	db, err := database.NewTestDB()
	if err != nil {
		t.Skip("Skipping test: database not available", err)
		return
	}

	app := fiber.New()
	handler := &handlers.Handler{}

	enterpriseService, err := services.NewEnterprisePricingService(db)
	require.NoError(t, err)
	auditService, err := services.NewAuditService(db)
	require.NoError(t, err)

	handler.SetEnterpriseService(enterpriseService)
	handler.SetAuditService(auditService)

	app.Use(func(c fiber.Ctx) error {
		c.Locals("admin_id", "test-admin-audit")
		c.Locals("admin_name", "test-admin-audit")
		return c.Next()
	})

	api := app.Group("/api/v1/admin")
	api.Post("/enterprise-pricing", handler.CreateEnterprisePricing)
	api.Get("/audit/logs", handler.ListAuditLogs)

	t.Run("Audit Log Records Pricing Creation", func(t *testing.T) {
		payload := map[string]interface{}{
			"customer_id":       "audit-test-customer",
			"customer_name":     "Audit Test Corp",
			"model_id":          "claude-3-opus",
			"input_price":       0.015,
			"output_price":      0.075,
			"min_profit_margin": 15.0,
			"max_cost_price":    0.06,
			"effective_date":    time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/admin/enterprise-pricing", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		req.Header.Set("User-Agent", "audit-test")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		// 等待审计日志异步写入
		time.Sleep(100 * time.Millisecond)

		// 检查审计日志
		var auditLogs []models.AuditLog
		err = db.Where("entity_type = ? AND action = ?", "pricing", "create").Find(&auditLogs).Error
		require.NoError(t, err)

		// 验证至少有一条审计日志
		found := false
		for _, log := range auditLogs {
			if log.EntityType == "pricing" && log.Action == "create" {
				found = true
				assert.NotEmpty(t, log.AdminID)
				assert.NotEmpty(t, log.Changes.After)
				break
			}
		}
		assert.True(t, found, "Pricing creation audit log should be recorded")
	})

	// 清理
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM enterprise_pricing")
}
