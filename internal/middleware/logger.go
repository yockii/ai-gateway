package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/logging"
	"go.uber.org/zap"
)

// Logger 请求日志中间件
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// 处理请求
		err := c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 记录结构化日志
		logging.Info("HTTP Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
			zap.String("user_id", c.Query("user_id", "")),
		)

		return err
	}
}
