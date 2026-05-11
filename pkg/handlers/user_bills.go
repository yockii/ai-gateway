package handlers

import (
	"github.com/gofiber/fiber/v3"
)

// ListUserBills 获取用户账单列表
func (h *Handler) ListUserBills(c fiber.Ctx) error {
	_ = c.Locals("user_id").(string)

	// TODO: 从数据库获取用户账单
	// 返回模拟数据 - 字段名与前端 Bill 类型匹配
	return c.JSON(fiber.Map{
		"data": []fiber.Map{
			{
				"id":             "bill-001",
				"user_id":        "user-123",
				"period":         "2024-04",
				"start_date":     "2024-04-01T00:00:00Z",
				"end_date":       "2024-04-30T23:59:59Z",
				"total_requests": 12500,
				"total_cost":     120.00,
				"total_revenue":  158.50,
				"total_profit":   38.50,
				"status":         "paid",
				"created_at":     "2024-05-01T10:00:00Z",
				"updated_at":     "2024-05-01T10:00:00Z",
				"items": []fiber.Map{
					{
						"model_id":       "gpt-4",
						"request_count":  5200,
						"input_tokens":   800000,
						"output_tokens":  700000,
						"total_tokens":   1500000,
						"total_cost":     90.00,
						"total_revenue":  120.00,
						"total_profit":   30.00,
					},
					{
						"model_id":       "gpt-3.5-turbo",
						"request_count":  7300,
						"input_tokens":   600000,
						"output_tokens":  400000,
						"total_tokens":   1000000,
						"total_cost":     30.00,
						"total_revenue":  38.50,
						"total_profit":   8.50,
					},
				},
			},
			{
				"id":             "bill-002",
				"user_id":        "user-123",
				"period":         "2024-05",
				"start_date":     "2024-05-01T00:00:00Z",
				"end_date":       "2024-05-31T23:59:59Z",
				"total_requests": 5200,
				"total_cost":     45.00,
				"total_revenue":  58.50,
				"total_profit":   13.50,
				"status":         "pending",
				"created_at":     "2024-06-01T10:00:00Z",
				"updated_at":     "2024-06-01T10:00:00Z",
				"items": []fiber.Map{
					{
						"model_id":       "gpt-4",
						"request_count":  2200,
						"input_tokens":   350000,
						"output_tokens":  300000,
						"total_tokens":   650000,
						"total_cost":     38.00,
						"total_revenue":  48.00,
						"total_profit":   10.00,
					},
					{
						"model_id":       "gpt-3.5-turbo",
						"request_count":  3000,
						"input_tokens":   250000,
						"output_tokens":  200000,
						"total_tokens":   450000,
						"total_cost":     7.00,
						"total_revenue":  10.50,
						"total_profit":   3.50,
					},
				},
			},
		},
		"object": "list",
	})
}

// GetUserBill 获取账单详情
func (h *Handler) GetUserBill(c fiber.Ctx) error {
	billID := c.Params("id")
	// TODO: 从数据库获取账单详情
	// 返回模拟数据 - 字段名与前端 Bill 类型匹配
	return c.JSON(fiber.Map{
		"id":             billID,
		"user_id":        "user-123",
		"period":         "2024-04",
		"start_date":     "2024-04-01T00:00:00Z",
		"end_date":       "2024-04-30T23:59:59Z",
		"total_requests": 12500,
		"total_cost":     120.00,
		"total_revenue":  158.50,
		"total_profit":   38.50,
		"status":         "paid",
		"created_at":     "2024-05-01T10:00:00Z",
		"updated_at":     "2024-05-01T10:00:00Z",
		"items": []fiber.Map{
			{
				"model_id":       "gpt-4",
				"request_count":  5200,
				"input_tokens":   800000,
				"output_tokens":  700000,
				"total_tokens":   1500000,
				"total_cost":     90.00,
				"total_revenue":  120.00,
				"total_profit":   30.00,
			},
			{
				"model_id":       "gpt-3.5-turbo",
				"request_count":  7300,
				"input_tokens":   600000,
				"output_tokens":  400000,
				"total_tokens":   1000000,
				"total_cost":     30.00,
				"total_revenue":  38.50,
				"total_profit":   8.50,
			},
		},
	})
}

// ExportUserBill 导出账单
func (h *Handler) ExportUserBill(c fiber.Ctx) error {
	billID := c.Params("id")

	var req struct {
		Format string `json:"format"` // pdf or csv
	}
	if err := c.Bind().Body(&req); err != nil {
		req.Format = "pdf" // 默认 PDF
	}

	// TODO: 生成账单文件并返回下载链接
	// 返回模拟数据
	return c.JSON(fiber.Map{
		"download_url": "/api/v1/user/bills/" + billID + "/download." + req.Format,
		"expires_at":   "2024-05-12T10:00:00Z",
	})
}
