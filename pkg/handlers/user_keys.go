package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/middleware"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
)

// SetKeyManager 设置 KeyManager
func (h *Handler) SetKeyManager(km *services.KeyManager) {
	h.keyManager = km
}

// ListUserKeys 获取用户的 API Key 列表
func (h *Handler) ListUserKeys(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	keys, err := h.keyManager.GetUserKeys(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Failed to fetch keys",
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"data":   keys,
		"object": "list",
	})
}

// CreateUserKey 创建新的 API Key
func (h *Handler) CreateUserKey(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req struct {
		Name             string           `json:"name"`
		QuotaDaily       *int64           `json:"quota_daily"`
		QuotaMonthly     *int64           `json:"quota_monthly"`
		ConcurrencyLimit *int64           `json:"concurrency_limit"`
		ModelConcurrency map[string]int64 `json:"model_concurrency"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Invalid request body",
				"type":    "invalid_request_error",
				"code":    400,
			},
		})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Name is required",
				"type":    "invalid_request_error",
				"code":    400,
			},
		})
	}

	// 设置默认值
	quotaDaily := int64(0)
	quotaMonthly := int64(0)
	concurrencyLimit := int64(10)

	if req.QuotaDaily != nil {
		quotaDaily = *req.QuotaDaily
	}
	if req.QuotaMonthly != nil {
		quotaMonthly = *req.QuotaMonthly
	}
	if req.ConcurrencyLimit != nil {
		concurrencyLimit = *req.ConcurrencyLimit
	}

	key, err := h.keyManager.CreateKey(c.Context(), userID, req.Name, services.CreateKeyOptions{
		QuotaDaily:       quotaDaily,
		QuotaMonthly:     quotaMonthly,
		ConcurrencyLimit: concurrencyLimit,
		ModelConcurrency: req.ModelConcurrency,
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	// 返回完整密钥（仅在创建时显示一次）
	return c.Status(201).JSON(fiber.Map{
		"data": fiber.Map{
			"id":         key.ID,
			"name":       key.Name,
			"api_key":    key.KeyValue,
			"created_at": key.CreatedAt,
		},
	})
}

// GetUserKey 获取 API Key 详情
func (h *Handler) GetUserKey(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	keyID := c.Params("id")

	keys, err := h.keyManager.GetUserKeys(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Failed to fetch key",
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	for _, key := range keys {
		if key.ID == keyID {
			return c.JSON(fiber.Map{
				"data": key,
			})
		}
	}

	return c.Status(404).JSON(fiber.Map{
		"error": fiber.Map{
			"message": "Key not found",
			"type":    "not_found_error",
			"code":    404,
		},
	})
}

// UpdateUserKey 更新 API Key 配置
func (h *Handler) UpdateUserKey(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	keyID := c.Params("id")

	var req struct {
		Name             *string           `json:"name"`
		QuotaDaily       *int64            `json:"quota_daily"`
		QuotaMonthly     *int64            `json:"quota_monthly"`
		ConcurrencyLimit *int64            `json:"concurrency_limit"`
		ModelConcurrency map[string]int64 `json:"model_concurrency"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Invalid request body",
				"type":    "invalid_request_error",
				"code":    400,
			},
		})
	}

	// 获取当前配置
	keys, err := h.keyManager.GetUserKeys(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Failed to fetch key",
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	var currentKey *models.UserAPIKey
	for i := range keys {
		if keys[i].ID == keyID {
			currentKey = &keys[i]
			break
		}
	}

	if currentKey == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Key not found",
				"type":    "not_found_error",
				"code":    404,
			},
		})
	}

	// 使用当前值作为默认值
	name := currentKey.Name
	quotaDaily := currentKey.QuotaDaily
	quotaMonthly := currentKey.QuotaMonthly
	concurrencyLimit := currentKey.ConcurrencyLimit
	modelConcurrency := currentKey.ModelConcurrency

	if req.Name != nil {
		name = *req.Name
	}
	if req.QuotaDaily != nil {
		quotaDaily = *req.QuotaDaily
	}
	if req.QuotaMonthly != nil {
		quotaMonthly = *req.QuotaMonthly
	}
	if req.ConcurrencyLimit != nil {
		concurrencyLimit = *req.ConcurrencyLimit
	}
	if req.ModelConcurrency != nil {
		modelConcurrency = req.ModelConcurrency
	}

	if err := h.keyManager.UpdateKey(c.Context(), keyID, userID, services.UpdateKeyOptions{
		Name:             name,
		QuotaDaily:       quotaDaily,
		QuotaMonthly:     quotaMonthly,
		ConcurrencyLimit: concurrencyLimit,
		ModelConcurrency: modelConcurrency,
	}); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Key updated successfully",
	})
}

// DeleteUserKey 删除 API Key
func (h *Handler) DeleteUserKey(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	keyID := c.Params("id")

	if err := h.keyManager.DeleteKey(c.Context(), keyID, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.Status(204).Send(nil)
}

// DisableUserKey 禁用 API Key
func (h *Handler) DisableUserKey(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	keyID := c.Params("id")

	if err := h.keyManager.UpdateKeyStatus(c.Context(), keyID, userID, false); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"id":        keyID,
		"is_active": false,
	})
}

// EnableUserKey 启用 API Key
func (h *Handler) EnableUserKey(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	keyID := c.Params("id")

	if err := h.keyManager.UpdateKeyStatus(c.Context(), keyID, userID, true); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"id":        keyID,
		"is_active": true,
	})
}

// GetUserKeyStats 获取 API Key 使用统计
func (h *Handler) GetUserKeyStats(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	keyID := c.Params("id")

	// 修复 CR-09: 验证用户拥有该密钥
	keys, err := h.keyManager.GetUserKeys(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Failed to fetch keys",
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	// 查找目标密钥并验证所有权
	var found bool
	for _, key := range keys {
		if key.ID == keyID {
			found = true
			break
		}
	}
	if !found {
		return c.Status(404).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Key not found",
				"type":    "not_found_error",
				"code":    404,
			},
		})
	}

	stats, err := h.keyManager.GetKeyStats(c.Context(), keyID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": err.Error(),
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": stats,
	})
}
