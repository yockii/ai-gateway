package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// ListPlans 获取套餐列表
func (h *Handler) ListPlans(c fiber.Ctx) error {
	// TODO: 从数据库查询
	plans := []fiber.Map{
		{
			"id":          "plan-free",
			"name":        "free",
			"display_name": "免费版",
			"level":       0,
			"price":       0,
			"currency":    "CNY",
			"is_active":   true,
			"features": []string{
				"每月 10,000 Token",
				"基础模型访问",
				"社区支持",
			},
			"created_at": time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			"id":          "plan-pro",
			"name":        "pro",
			"display_name": "专业版",
			"level":       1,
			"price":       99,
			"currency":    "CNY",
			"is_active":   true,
			"features": []string{
				"每月 100,000 Token",
				"所有模型访问",
				"优先处理",
				"邮件支持",
			},
			"created_at": time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			"id":          "plan-enterprise",
			"name":        "enterprise",
			"display_name": "企业版",
			"level":       2,
			"price":       999,
			"currency":    "CNY",
			"is_active":   true,
			"features": []string{
				"无限 Token",
				"所有模型访问",
				"专属服务",
				"SLA 保证",
				"技术支持",
			},
			"created_at": time.Now().Add(-30 * 24 * time.Hour),
		},
	}

	return c.JSON(fiber.Map{
		"object": "list",
		"data":   plans,
	})
}

// CreatePlan 创建套餐
func (h *Handler) CreatePlan(c fiber.Ctx) error {
	var req struct {
		Name        string   `json:"name"`
		DisplayName string   `json:"display_name"`
		Level       int      `json:"level"`
		Price       float64  `json:"price"`
		Currency    string   `json:"currency"`
		Features    []string `json:"features"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// TODO: 保存到数据库

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":          "plan-new",
		"name":        req.Name,
		"display_name": req.DisplayName,
		"level":       req.Level,
		"price":       req.Price,
		"currency":    req.Currency,
		"features":    req.Features,
		"is_active":   true,
		"created_at":  time.Now(),
	})
}

// UpdatePlan 更新套餐
func (h *Handler) UpdatePlan(c fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		DisplayName string   `json:"display_name"`
		Price       float64  `json:"price"`
		Features    []string `json:"features"`
		IsActive    *bool    `json:"is_active"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// TODO: 更新数据库

	return c.JSON(fiber.Map{
		"id":          id,
		"display_name": req.DisplayName,
		"price":       req.Price,
		"features":    req.Features,
		"updated_at":  time.Now(),
	})
}

// DeletePlan 删除套餐
func (h *Handler) DeletePlan(c fiber.Ctx) error {
	id := c.Params("id")
	_ = id // TODO: 使用 id 删除数据库中的记录

	return c.Status(fiber.StatusNoContent).Send(nil)
}
