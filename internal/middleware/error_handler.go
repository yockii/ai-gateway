package middleware

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/pkg/api"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 处理请求
		err := c.Next()

		// 如果没有错误，直接返回
		if err == nil {
			return nil
		}

		// 检查是否是 fiber.Error
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: e.Message,
					Type:    "api_error",
					Code:    e.Code,
				},
			})
		}

		// 其他错误，返回 500
		log.Printf("未处理的错误: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Internal server error",
				Type:    "internal_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}
}

// Recovery 恢复中间件，捕获 panic
func Recovery() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", r)
				}
				log.Printf("Panic 恢复: %v", err)

				c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
					Error: api.ErrorDetail{
						Message: "Internal server error",
						Type:    "panic",
						Code:    fiber.StatusInternalServerError,
					},
				})
			}
		}()

		return c.Next()
	}
}
