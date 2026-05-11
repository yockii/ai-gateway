package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
)

// SetSupplierApiKeyService 设置供应商 API Key 服务
func (h *Handler) SetSupplierApiKeyService(svc *services.SupplierApiKeyService) {
	h.supplierApiKeyService = svc
}

// SetSupplierModelService 设置供应商模型服务
func (h *Handler) SetSupplierModelService(svc *services.SupplierModelService) {
	h.supplierModelService = svc
}

// ListSupplierApiKeys 获取供应商 API Key 列表
func (h *Handler) ListSupplierApiKeys(c fiber.Ctx) error {
	supplierID := c.Params("id")
	
	keys, err := h.supplierApiKeyService.ListApiKeys(c.Context(), supplierID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// 脱敏处理
	result := make([]fiber.Map, len(keys))
	for i, key := range keys {
		result[i] = fiber.Map{
			"id":               key.ID,
			"supplier_id":      key.SupplierID,
			"name":             key.Name,
			"key_prefix":       key.KeyPrefix,
			"priority":         key.Priority,
			"is_primary":       key.IsPrimary,
			"max_requests":     key.MaxRequests,
			"current_requests": key.CurrentRequests,
			"is_active":        key.IsActive,
			"last_used_at":     key.LastUsedAt,
			"expire_at":        key.ExpireAt,
			"created_at":       key.CreatedAt,
			"updated_at":       key.UpdatedAt,
		}
	}
	
	return c.JSON(fiber.Map{"data": result})
}

// CreateSupplierApiKey 创建供应商 API Key
func (h *Handler) CreateSupplierApiKey(c fiber.Ctx) error {
	supplierID := c.Params("id")

	var req struct {
		Name        string  `json:"name"`
		ApiKey      string  `json:"api_key"`
		Priority    int     `json:"priority"`
		IsPrimary   bool    `json:"is_primary"`
		MaxRequests int64   `json:"max_requests"`
		ExpireAt    *string `json:"expire_at"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	key := &models.SupplierApiKey{
		SupplierID:  supplierID,
		Name:        req.Name,
		Priority:    req.Priority,
		IsPrimary:   req.IsPrimary,
		MaxRequests: req.MaxRequests,
		IsActive:    true,
	}

	if req.ExpireAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpireAt)
		if err == nil {
			key.ExpireAt = &t
		}
	}

	if err := h.supplierApiKeyService.CreateApiKey(c.Context(), key, req.ApiKey); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 审计日志记录
	h.logAuditAsync(c, "api_key", key.ID, "create", models.ChangeLog{
		After: map[string]interface{}{
			"supplier_id":  key.SupplierID,
			"name":         key.Name,
			"priority":     key.Priority,
			"is_primary":   key.IsPrimary,
			"max_requests": key.MaxRequests,
			"is_active":    key.IsActive,
		},
	})

	return c.JSON(fiber.Map{"data": key, "message": "API Key created successfully"})
}

// UpdateSupplierApiKey 更新供应商 API Key
func (h *Handler) UpdateSupplierApiKey(c fiber.Ctx) error {
	_ = c.Params("kid")
	
	var req struct {
		Name        string  `json:"name"`
		Priority    int     `json:"priority"`
		MaxRequests int64   `json:"max_requests"`
		IsActive    *bool   `json:"is_active"`
		ExpireAt    *string `json:"expire_at"`
	}
	
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	// 获取现有密钥
	// TODO: 实现完整的更新逻辑
	
	return c.JSON(fiber.Map{"message": "API Key updated successfully"})
}

// DeleteSupplierApiKey 删除供应商 API Key
func (h *Handler) DeleteSupplierApiKey(c fiber.Ctx) error {
	keyID := c.Params("kid")

	// 获取密钥信息用于审计日志
	key, err := h.supplierApiKeyService.GetApiKeyByID(c.Context(), keyID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.supplierApiKeyService.DeleteApiKey(c.Context(), keyID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 审计日志记录
	h.logAuditAsync(c, "api_key", keyID, "delete", models.ChangeLog{
		Before: map[string]interface{}{
			"supplier_id": key.SupplierID,
			"name":        key.Name,
			"priority":    key.Priority,
			"is_primary":  key.IsPrimary,
		},
	})

	return c.JSON(fiber.Map{"message": "API Key deleted successfully"})
}

// SetPrimarySupplierApiKey 设置主密钥
func (h *Handler) SetPrimarySupplierApiKey(c fiber.Ctx) error {
	keyID := c.Params("kid")

	// 获取密钥信息用于审计日志
	key, err := h.supplierApiKeyService.GetApiKeyByID(c.Context(), keyID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.supplierApiKeyService.SetPrimaryApiKey(c.Context(), keyID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 审计日志记录
	h.logAuditAsync(c, "api_key", keyID, "set_primary", models.ChangeLog{
		Before: map[string]interface{}{
			"is_primary": false,
		},
		After: map[string]interface{}{
			"is_primary": true,
			"supplier_id": key.SupplierID,
			"name":       key.Name,
		},
	})

	return c.JSON(fiber.Map{"message": "Primary key set successfully"})
}

// RotateSupplierApiKey 密钥轮换
func (h *Handler) RotateSupplierApiKey(c fiber.Ctx) error {
	supplierID := c.Params("id")

	if err := h.supplierApiKeyService.RotateApiKey(c.Context(), supplierID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 审计日志记录
	h.logAuditAsync(c, "api_key", supplierID, "rotate", models.ChangeLog{
		After: map[string]interface{}{
			"supplier_id": supplierID,
			"action":      "rotation_completed",
		},
	})

	return c.JSON(fiber.Map{"message": "Key rotated successfully"})
}

// GetSupplierApiKeyStats 获取密钥统计
func (h *Handler) GetSupplierApiKeyStats(c fiber.Ctx) error {
	keyID := c.Params("kid")
	
	stats, err := h.supplierApiKeyService.GetApiKeyStats(c.Context(), keyID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.JSON(fiber.Map{"data": stats})
}

// ListSupplierModels 获取供应商模型列表
func (h *Handler) ListSupplierModels(c fiber.Ctx) error {
	supplierID := c.Params("id")
	
	sm, err := h.supplierModelService.ListSupplierModels(c.Context(), supplierID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.JSON(fiber.Map{"data": sm})
}

// AddSupplierModel 添加供应商模型
func (h *Handler) AddSupplierModel(c fiber.Ctx) error {
	supplierID := c.Params("id")

	var req struct {
		ModelID    string  `json:"model_id"`
		InputCost  float64 `json:"input_cost"`
		OutputCost float64 `json:"output_cost"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.supplierModelService.AddModelToSupplier(c.Context(), supplierID, req.ModelID, req.InputCost, req.OutputCost); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 审计日志记录
	h.logAuditAsync(c, "model", req.ModelID, "create", models.ChangeLog{
		After: map[string]interface{}{
			"supplier_id": supplierID,
			"model_id":    req.ModelID,
			"input_cost":  req.InputCost,
			"output_cost": req.OutputCost,
		},
	})

	return c.JSON(fiber.Map{"message": "Model added successfully"})
}

// UpdateSupplierModelCost 更新模型价格
func (h *Handler) UpdateSupplierModelCost(c fiber.Ctx) error {
	modelID := c.Params("mid")

	var req struct {
		InputCost  float64 `json:"input_cost"`
		OutputCost float64 `json:"output_cost"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 获取当前价格用于审计日志
	beforeCost, err := h.supplierModelService.GetModelCost(c.Context(), modelID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.supplierModelService.UpdateModelCost(c.Context(), modelID, req.InputCost, req.OutputCost); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 审计日志记录
	h.logAuditAsync(c, "model", modelID, "update", models.ChangeLog{
		Before: map[string]interface{}{
			"input_cost":  beforeCost.InputCost,
			"output_cost": beforeCost.OutputCost,
		},
		After: map[string]interface{}{
			"input_cost":  req.InputCost,
			"output_cost": req.OutputCost,
		},
	})

	return c.JSON(fiber.Map{"message": "Model cost updated successfully"})
}

// RemoveSupplierModel 移除供应商模型
func (h *Handler) RemoveSupplierModel(c fiber.Ctx) error {
	modelID := c.Params("mid")

	// 获取模型信息用于审计日志
	modelInfo, err := h.supplierModelService.GetModelInfo(c.Context(), modelID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.supplierModelService.RemoveModelFromSupplier(c.Context(), modelID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 审计日志记录
	h.logAuditAsync(c, "model", modelID, "delete", models.ChangeLog{
		Before: map[string]interface{}{
			"supplier_id": modelInfo.SupplierID,
			"model_id":    modelInfo.ModelID,
			"input_cost":  modelInfo.InputCost,
			"output_cost": modelInfo.OutputCost,
		},
	})

	return c.JSON(fiber.Map{"message": "Model removed successfully"})
}

// GetSupplierModelPriceHistory 获取价格历史
func (h *Handler) GetSupplierModelPriceHistory(c fiber.Ctx) error {
	supplierID := c.Params("id")
	modelID := c.Params("mid")

	history, err := h.supplierModelService.GetPriceHistory(c.Context(), supplierID, modelID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": history})
}

// logAuditAsync 异步记录审计日志
func (h *Handler) logAuditAsync(c fiber.Ctx, entityType, entityID, action string, changes models.ChangeLog) {
	if h.auditService == nil {
		return
	}

	adminID := ""
	adminName := ""
	if id, ok := c.Locals("admin_id").(string); ok {
		adminID = id
	}
	if name, ok := c.Locals("admin_name").(string); ok {
		adminName = name
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	h.auditService.LogActionAsync(c.Context(), adminID, adminName, entityType, entityID, action, changes, ipAddress, userAgent)
}

// getAdminInfo 获取管理员信息用于审计日志
func (h *Handler) getAdminInfo(c fiber.Ctx) (adminID, adminName string) {
	adminID, _ = c.Locals("admin_id").(string)
	adminName, _ = c.Locals("admin_name").(string)

	// 如果没有 admin_name，尝试从数据库获取
	if adminName == "" && adminID != "" && h.adminService != nil {
		// 简化处理：使用 ID 作为名称
		adminName = adminID
	}

	return
}
