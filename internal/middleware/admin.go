package middleware

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/auth"
	"github.com/yockii/ai-gateway/pkg/api"
)

const (
	// HeaderAdminToken 管理员 Token 请求头
	HeaderAdminToken = "Authorization"
)

// RequireAdmin 要求管理员权限的中间件
func RequireAdmin(tokenManager *auth.TokenManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		fmt.Println("=== RequireAdmin middleware called for path:", c.Path(), "===")

		// 获取 Authorization header
		authHeader := c.Get(HeaderAdminToken)
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

		// 检查是否为管理员
		if tokenInfo.Type != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Admin access required",
					Type:    "permission_error",
					Code:    fiber.StatusForbidden,
				},
			})
		}

		// 将管理员信息存储到上下文
		c.Locals("admin_id", tokenInfo.UserID)
		c.Locals("admin_email", tokenInfo.Email)

		return c.Next()
	}
}
