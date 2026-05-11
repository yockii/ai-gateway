package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/pkg/api"
)

const (
	// HeaderUserID 用户 ID 请求头
	HeaderUserID = "X-User-ID"
	// HeaderAPIKey API Key 请求头
	HeaderAPIKey = "Authorization"
)

// Auth 认证中间件
func Auth() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 跳过管理员和用户 API 路径（这些路径有自己的认证中间件）
		path := c.Path()
		if strings.HasPrefix(path, "/v1/admin") || strings.HasPrefix(path, "/v1/user") {
			return c.Next()
		}

		// 获取 API Key
		apiKey := c.Get(HeaderAPIKey)

		// 检查 Bearer token 格式
		if strings.HasPrefix(apiKey, "Bearer ") {
			apiKey = strings.TrimPrefix(apiKey, "Bearer ")
		}

		// 如果没有 API Key，返回 401
		if apiKey == "" {
			log.Printf("认证失败: 缺少 API Key - IP: %s", c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Missing API key",
					Type:    "authentication_error",
					Code:    fiber.StatusUnauthorized,
				},
			})
		}

		// TODO: 验证 API Key
		// 当前简化实现：从 API Key 提取用户 ID
		// 实际应该查询数据库验证
		userID := extractUserIDFromAPIKey(apiKey)
		if userID == "" {
			log.Printf("认证失败: 无效的 API Key - IP: %s", c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Invalid API key",
					Type:    "authentication_error",
					Code:    fiber.StatusUnauthorized,
				},
			})
		}

		// 将用户 ID 存储到上下文
		c.Locals("user_id", userID)
		c.Locals("api_key", apiKey)

		log.Printf("认证成功: 用户=%s IP=%s", userID, c.IP())

		return c.Next()
	}
}

// extractUserIDFromAPIKey 从 API Key 提取用户 ID（临时实现）
// TODO: 实现 API Key 验证逻辑
func extractUserIDFromAPIKey(apiKey string) string {
	// 临时实现：假设 API Key 格式为 "sk-<userID>-<random>"
	// 实际应该查询数据库验证
	if strings.HasPrefix(apiKey, "sk-") {
		parts := strings.Split(apiKey, "-")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// 用于测试的默认用户 ID
	if apiKey == "test-api-key" {
		return "test-user-001"
	}

	return ""
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c fiber.Ctx) string {
	userID, _ := c.Locals("user_id").(string)
	return userID
}
