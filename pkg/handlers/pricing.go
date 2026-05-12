package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/models"
)

type enterprisePricingRequest struct {
	CustomerID      string   `json:"customer_id"`
	CustomerName    string   `json:"customer_name"`
	ModelID         string   `json:"model_id"`
	InputPrice      float64  `json:"input_price"`
	OutputPrice     float64  `json:"output_price"`
	MinProfitMargin float64  `json:"min_profit_margin"`
	MaxCostPrice    float64  `json:"max_cost_price"`
	EffectiveDate   string   `json:"effective_date"`
	ExpiryDate      *string  `json:"expiry_date"`
}

func (h *Handler) ListEnterprisePricing(c fiber.Ctx) error {
	filter := &models.EnterprisePricingFilter{Limit: 100}
	pricings, err := h.enterpriseService.ListEnterprisePricing(c.Context(), filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": err.Error(), "type": "internal_error"}})
	}
	return c.JSON(fiber.Map{"data": pricings, "object": "list"})
}

func (h *Handler) CreateEnterprisePricing(c fiber.Ctx) error {
	// 修复 CR-11: 使用安全的类型断言
	adminID, ok := c.Locals("admin_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": fiber.Map{"message": "Unauthorized", "type": "authentication_error"}})
	}

	var req enterprisePricingRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "Invalid request", "type": "invalid_request"}})
	}
	effectiveDate, _ := time.Parse(time.RFC3339, req.EffectiveDate)
	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		ed, _ := time.Parse(time.RFC3339, *req.ExpiryDate)
		expiryDate = &ed
	}
	pricing := &models.EnterprisePricing{
		CustomerID: req.CustomerID, CustomerName: req.CustomerName, ModelID: req.ModelID,
		InputPrice: req.InputPrice, OutputPrice: req.OutputPrice, MinProfitMargin: req.MinProfitMargin,
		MaxCostPrice: req.MaxCostPrice, EffectiveDate: effectiveDate, ExpiryDate: expiryDate, IsActive: true,
	}
	if err := h.enterpriseService.SetEnterprisePrice(c.Context(), pricing, adminID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": err.Error(), "type": "internal_error"}})
	}

	// 审计日志记录
	h.logAuditAsync(c, "pricing", pricing.ID, "create", models.ChangeLog{
		After: map[string]interface{}{
			"customer_id":      pricing.CustomerID,
			"customer_name":    pricing.CustomerName,
			"model_id":         pricing.ModelID,
			"input_price":      pricing.InputPrice,
			"output_price":     pricing.OutputPrice,
			"min_profit_margin": pricing.MinProfitMargin,
			"effective_date":   pricing.EffectiveDate,
		},
	})

	return c.Status(201).JSON(pricing)
}

func (h *Handler) GetEnterprisePricing(c fiber.Ctx) error {
	id := c.Params("id")
	filter := &models.EnterprisePricingFilter{}
	pricings, _ := h.enterpriseService.ListEnterprisePricing(c.Context(), filter)
	for _, p := range pricings {
		if p.ID == id {
			return c.JSON(p)
		}
	}
	return c.Status(404).JSON(fiber.Map{"error": fiber.Map{"message": "Not found", "type": "not_found"}})
}

func (h *Handler) UpdateEnterprisePricing(c fiber.Ctx) error {
	// 修复 CR-11: 使用安全的类型断言
	adminID, ok := c.Locals("admin_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": fiber.Map{"message": "Unauthorized", "type": "authentication_error"}})
	}

	id := c.Params("id")
	var req struct {
		InputPrice      *float64 `json:"input_price"`
		OutputPrice     *float64 `json:"output_price"`
		MinProfitMargin *float64 `json:"min_profit_margin"`
		MaxCostPrice    *float64 `json:"max_cost_price"`
		IsActive        *bool    `json:"is_active"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "Invalid request"}})
	}
	filter := &models.EnterprisePricingFilter{}
	pricings, _ := h.enterpriseService.ListEnterprisePricing(c.Context(), filter)
	var existing *models.EnterprisePricing
	for _, p := range pricings {
		if p.ID == id {
			existing = p
			break
		}
	}
	if existing == nil {
		return c.Status(404).JSON(fiber.Map{"error": fiber.Map{"message": "Not found"}})
	}
	updated := *existing
	beforeData := map[string]interface{}{
		"input_price":       existing.InputPrice,
		"output_price":      existing.OutputPrice,
		"min_profit_margin": existing.MinProfitMargin,
		"max_cost_price":    existing.MaxCostPrice,
		"is_active":         existing.IsActive,
	}
	if req.InputPrice != nil {
		updated.InputPrice = *req.InputPrice
	}
	if req.OutputPrice != nil {
		updated.OutputPrice = *req.OutputPrice
	}
	if req.MinProfitMargin != nil {
		updated.MinProfitMargin = *req.MinProfitMargin
	}
	if req.MaxCostPrice != nil {
		updated.MaxCostPrice = *req.MaxCostPrice
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}
	if err := h.enterpriseService.UpdateEnterprisePrice(c.Context(), id, &updated, adminID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}

	// 审计日志记录
	afterData := map[string]interface{}{
		"input_price":       updated.InputPrice,
		"output_price":      updated.OutputPrice,
		"min_profit_margin": updated.MinProfitMargin,
		"max_cost_price":    updated.MaxCostPrice,
		"is_active":         updated.IsActive,
	}
	h.logAuditAsync(c, "pricing", id, "update", models.ChangeLog{
		Before: beforeData,
		After:  afterData,
	})

	return c.JSON(&updated)
}

func (h *Handler) DeleteEnterprisePricing(c fiber.Ctx) error {
	id := c.Params("id")

	// 获取定价信息用于审计日志
	filter := &models.EnterprisePricingFilter{}
	pricings, _ := h.enterpriseService.ListEnterprisePricing(c.Context(), filter)
	var existing *models.EnterprisePricing
	for _, p := range pricings {
		if p.ID == id {
			existing = p
			break
		}
	}

	if err := h.enterpriseService.DeleteEnterprisePrice(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}

	// 审计日志记录
	if existing != nil {
		h.logAuditAsync(c, "pricing", id, "delete", models.ChangeLog{
			Before: map[string]interface{}{
				"customer_id":   existing.CustomerID,
				"customer_name": existing.CustomerName,
				"model_id":      existing.ModelID,
				"input_price":   existing.InputPrice,
				"output_price":  existing.OutputPrice,
			},
		})
	}

	return c.SendStatus(204)
}

func (h *Handler) GetUserPricing(c fiber.Ctx) error {
	// 修复 CR-11: 使用安全的类型断言
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": fiber.Map{"message": "Unauthorized", "type": "authentication_error"}})
	}

	modelID := c.Query("model_id", "")
	if modelID == "" {
		return c.Status(400).JSON(fiber.Map{"error": fiber.Map{"message": "model_id required"}})
	}
	pricing, err := h.pricingService.GetUserPrice(c.Context(), userID, modelID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": err.Error()}})
	}
	return c.JSON(pricing)
}
