package handlers

import (
	"github.com/gofiber/fiber/v3"
)

// ListPublicPlans 获取公共套餐列表（不需要认证）
func (h *Handler) ListPublicPlans(c fiber.Ctx) error {
	// TODO: 从数据库获取可用套餐
	// 返回模拟数据 - 字段名与前端 MembershipPlan 类型匹配
	return c.JSON(fiber.Map{
		"data": []fiber.Map{
			{
				"id":            "plan-free",
				"name":          "免费版",
				"description":   "适合个人开发者试用",
				"price_monthly": 0,
				"price_yearly":  0,
				"features":      []string{"10,000 Tokens/月", "基础模型访问", "社区支持"},
				"is_active":     true,
				"is_popular":    false,
			},
			{
				"id":            "plan-pro",
				"name":          "专业版",
				"description":   "适合中小型项目",
				"price_monthly": 99,
				"price_yearly":  950,
				"features":      []string{"100,000 Tokens/月", "所有模型访问", "优先处理", "邮件支持"},
				"is_active":     true,
				"is_popular":    true,
			},
			{
				"id":            "plan-enterprise",
				"name":          "企业版",
				"description":   "适合大型企业",
				"price_monthly": 999,
				"price_yearly":  9990,
				"features":      []string{"无限 Tokens", "专属模型", "SLA 保证", "7x24 技术支持"},
				"is_active":     true,
				"is_popular":    false,
			},
		},
		"object": "list",
	})
}

// GetUserCurrentPlan 获取用户当前套餐
func (h *Handler) GetUserCurrentPlan(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	_ = userID // TODO: 使用 userID 获取套餐

	// TODO: 从数据库获取用户当前套餐
	// 返回模拟数据 - 字段名与前端 UserMembership 类型匹配
	return c.JSON(fiber.Map{
		"id":                     "plan-pro",
		"user_id":                userID,
		"plan_id":                "plan-pro",
		"status":                 "active",
		"current_period_start":   "2024-05-01T00:00:00Z",
		"current_period_end":     "2024-06-01T00:00:00Z",
		"cancel_at_period_end":   false,
		"auto_renew":             true,
		"tokens_used":            2500000,
		"tokens_limit":           100000,
		"requests":               12500,
	})
}

// SubscribePlan 订阅套餐
func (h *Handler) SubscribePlan(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	planID := c.Params("id")
	_ = userID // TODO: 使用 userID 创建订阅
	_ = planID // TODO: 使用 planID 订阅

	var req struct {
		Interval string `json:"interval"` // monthly or yearly
	}
	if err := c.Bind().Body(&req); err != nil || req.Interval == "" {
		req.Interval = "monthly"
	}

	// TODO: 创建订阅
	// 返回模拟数据
	return c.JSON(fiber.Map{
		"client_secret": "pi_test_secret",
		"plan_id":       planID,
		"interval":      req.Interval,
	})
}

// CancelPlan 取消套餐
func (h *Handler) CancelPlan(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	_ = userID // TODO: 使用 userID 取消订阅

	// TODO: 取消用户订阅
	return c.JSON(fiber.Map{
		"cancelled":      true,
		"effective_date": "2024-06-01T00:00:00Z",
	})
}
