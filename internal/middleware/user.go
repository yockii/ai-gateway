package middleware

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/auth"
	"github.com/yockii/ai-gateway/pkg/api"
)

// RequireUser 要求用户权限的中间件
func RequireUser(tokenManager *auth.TokenManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 获取 Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Missing authorization header",
					Type:    "authentication_error",
					Code:    fiber.StatusUnauthorized,
				},
			})
		}

		// 检查 Bearer 格式
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Invalid authorization format",
					Type:    "authentication_error",
					Code:    fiber.StatusUnauthorized,
				},
			})
		}

		token := authHeader[7:]

		// 验证 token（从 Redis）
		tokenInfo, err := tokenManager.ValidateToken(c.Context(), token)
		if err != nil {
			log.Printf("Token 验证失败: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Invalid or expired token",
					Type:    "authentication_error",
					Code:    fiber.StatusUnauthorized,
				},
			})
		}

		// 检查是否为用户
		if tokenInfo.Type != "user" {
			return c.Status(fiber.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "User access required",
					Type:    "permission_error",
					Code:    fiber.StatusForbidden,
				},
			})
		}

		// 将用户信息存储到上下文
		c.Locals("user_id", tokenInfo.UserID)
		c.Locals("user_email", tokenInfo.Email)

		return c.Next()
	}
}
