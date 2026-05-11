package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// GetUserUsage 获取用户使用记录列表
func (h *Handler) GetUserUsage(c fiber.Ctx) error {
	_ = c.Locals("user_id").(string)

	// TODO: 从数据库获取用户使用记录
	// 返回模拟数据 - 字段名与前端 UsageRecord 类型匹配
	now := time.Now()
	records := make([]fiber.Map, 50)

	models := []string{"gpt-4", "gpt-3.5-turbo", "claude-3-opus"}
	for i := 0; i < 50; i++ {
		model := models[i%len(models)]
		inputTokens := 500 + (i * 100)
		outputTokens := 300 + (i * 80)
		totalTokens := inputTokens + outputTokens
		costPrice := float64(totalTokens) * 0.00001
		sellingPrice := float64(totalTokens) * 0.000015

		records[i] = fiber.Map{
			"id":            "usage-" + string(rune(i)),
			"request_id":    "req-" + string(rune(i)),
			"user_id":       "user-123",
			"model_id":      model,
			"supplier_id":   "openai",
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
			"total_tokens":  totalTokens,
			"cost_price":    costPrice,
			"selling_price": sellingPrice,
			"profit":        sellingPrice - costPrice,
			"created_at":    now.Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		}
	}

	return c.JSON(fiber.Map{
		"data":  records,
		"total": 50,
		"object": "list",
	})
}
