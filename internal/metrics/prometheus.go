package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

// PrometheusMiddleware 记录 HTTP 请求指标
func PrometheusMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// 增加并发连接数
		ConcurrentConnections.Inc()
		defer ConcurrentConnections.Dec()

		// 处理请求
		err := c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		HTTPRequestsTotal.WithLabelValues(
			c.Method(),
			c.Path(),
			status,
		).Inc()

		HTTPRequestDuration.WithLabelValues(
			c.Method(),
			c.Path(),
		).Observe(duration)

		// 记录错误
		if err != nil || c.Response().StatusCode() >= 400 {
			ErrorsTotal.WithLabelValues(
				"http",
				c.Path(),
			).Inc()
		}

		return err
	}
}

// RecordAPIRequest 记录 API 请求
func RecordAPIRequest(model, provider string) {
	APIRequestsTotal.WithLabelValues(model, provider).Inc()
}

// RecordTokenUsage 记录 Token 使用量
func RecordTokenUsage(model string, inputTokens, outputTokens int32) {
	APITokensTotal.WithLabelValues(model, "input").Add(float64(inputTokens))
	APITokensTotal.WithLabelValues(model, "output").Add(float64(outputTokens))
}

// RecordCost 记录费用
func RecordCost(model, currency string, cost float64) {
	APICostTotal.WithLabelValues(model, currency).Add(cost)
}

// IncrementActiveUsers 增加活跃用户数
func IncrementActiveUsers() {
	ActiveUsers.Inc()
}

// DecrementActiveUsers 减少活跃用户数
func DecrementActiveUsers() {
	ActiveUsers.Dec()
}

// RecordError 记录错误
func RecordError(errorType, location string) {
	ErrorsTotal.WithLabelValues(errorType, location).Inc()
}
