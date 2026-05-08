package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Logger 请求日志中间件
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// 处理请求
		err := c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 记录日志
		log.Printf("[%s] %s %s - %d - %v - %s",
			c.IP(),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			duration,
			c.Query("user_id", ""), // 记录用户 ID（如果有）
		)

		return err
	}
}
