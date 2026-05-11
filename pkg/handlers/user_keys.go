package handlers

import (
	"github.com/gofiber/fiber/v3"
)

// ListUserKeys 获取用户的 API Key 列表
func (h *Handler) ListUserKeys(c fiber.Ctx) error {
	_ = c.Locals("user_id").(string)

	// TODO: 从数据库获取用户的 API Keys
	// 返回模拟数据
	return c.JSON(fiber.Map{
		"data": []fiber.Map{
			{
				"id":            "key1",
				"name":          "Production Key",
				"created_at":    "2024-01-15T10:30:00Z",
				"last_used":     "2024-05-10T15:45:00Z",
				"is_active":     true,
				"quota_daily":   100000,
				"quota_monthly": 3000000,
			},
			{
				"id":            "key2",
				"name":          "Development Key",
				"created_at":    "2024-02-20T08:00:00Z",
				"last_used":     "2024-05-11T09:00:00Z",
				"is_active":     true,
				"quota_daily":   10000,
				"quota_monthly": 300000,
			},
		},
		"object": "list",
	})
}

// CreateUserKey 创建新的 API Key
func (h *Handler) CreateUserKey(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Invalid request body",
				"type":    "invalid_request_error",
				"code":    400,
			},
		})
	}

	// TODO: 在数据库中创建 API Key
	// 返回模拟数据
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         "new-key-id",
		"name":       req.Name,
		"api_key":    "sk-new-key-" + userID[:8],
		"created_at": "2024-05-11T10:00:00Z",
	})
}

// DeleteUserKey 删除 API Key
func (h *Handler) DeleteUserKey(c fiber.Ctx) error {
	_ = c.Params("id")
	// TODO: 从数据库删除 API Key
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// DisableUserKey 禁用 API Key
func (h *Handler) DisableUserKey(c fiber.Ctx) error {
	keyID := c.Params("id")
	// TODO: 在数据库中禁用 API Key
	return c.JSON(fiber.Map{
		"id":         keyID,
		"is_active":  false,
		"updated_at": "2024-05-11T10:00:00Z",
	})
}

// EnableUserKey 启用 API Key
func (h *Handler) EnableUserKey(c fiber.Ctx) error {
	keyID := c.Params("id")
	// TODO: 在数据库中启用 API Key
	return c.JSON(fiber.Map{
		"id":         keyID,
		"is_active":  true,
		"updated_at": "2024-05-11T10:00:00Z",
	})
}

// GetUserKeyStats 获取 API Key 使用统计
func (h *Handler) GetUserKeyStats(c fiber.Ctx) error {
	keyID := c.Params("id")
	// TODO: 从数据库获取 API Key 统计
	// 返回模拟数据
	return c.JSON(fiber.Map{
		"key_id":          keyID,
		"total_requests":  15420,
		"total_tokens":    3250000,
		"total_cost":      156.80,
		"daily_requests":  520,
		"daily_tokens":    85000,
		"daily_cost":      5.20,
	})
}
