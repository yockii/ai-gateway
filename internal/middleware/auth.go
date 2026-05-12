package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/pkg/api"
)

const (
	// HeaderUserID 用户 ID 请求头
	HeaderUserID = "X-User-ID"
	// HeaderAPIKey API Key 请求头
	HeaderAPIKey = "Authorization"
)

// Auth 认证中间件（使用 KeyManager）
func Auth(keyManager *services.KeyManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 跳过管理员和用户 API 路径（这些路径有自己的认证中间件）
		path := c.Path()
		if strings.HasPrefix(path, "/v1/admin") || strings.HasPrefix(path, "/v1/user") {
			return c.Next()
		}

		// 获取 API Key
		apiKey := extractBearerToken(c.Get(HeaderAPIKey))

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

		// 使用 KeyManager 验证 API Key
		keyInfo, err := keyManager.ValidateKey(c.Context(), apiKey)
		if err != nil {
			log.Printf("认证失败: 无效的 API Key - IP: %s 错误: %v", c.IP(), err)
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Invalid API key",
					Type:    "authentication_error",
					Code:    fiber.StatusUnauthorized,
				},
			})
		}

		// 检查额度限制
		if err := keyManager.CheckQuota(c.Context(), keyInfo.KeyID); err != nil {
			log.Printf("认证失败: 额度超限 - KeyID: %s 错误: %v", keyInfo.KeyID, err)
			return c.Status(fiber.StatusTooManyRequests).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Quota exceeded: " + err.Error(),
					Type:    "quota_exceeded_error",
					Code:    fiber.StatusTooManyRequests,
				},
			})
		}

		// 将用户信息存储到上下文
		c.Locals("user_id", keyInfo.UserID)
		c.Locals("key_id", keyInfo.KeyID)
		c.Locals("api_key", apiKey)

		log.Printf("认证成功: 用户=%s KeyID=%s IP=%s", keyInfo.UserID, keyInfo.KeyID, c.IP())

		return c.Next()
	}
}

// extractBearerToken 从 Authorization 头提取 Bearer Token
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	// 检查 Bearer prefix
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 直接返回（兼容旧格式）
	return authHeader
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c fiber.Ctx) string {
	userID, _ := c.Locals("user_id").(string)
	return userID
}

// GetKeyID 从上下文获取 Key ID
func GetKeyID(c fiber.Ctx) string {
	keyID, _ := c.Locals("key_id").(string)
	return keyID
}
