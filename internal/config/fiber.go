package config

import (
	"github.com/gofiber/fiber/v3"
)

// FiberConfig 返回优化的 Fiber 配置
func (c *Config) FiberConfig() fiber.Config {
	return fiber.Config{
		// 服务器信息
		ServerHeader: "AI-Gateway",

		// 性能优化配置
		BodyLimit: 4 * 1024 * 1024, // 4MB 最大请求体

		// 路由配置
		StrictRouting: false,
		CaseSensitive: false,

		// 其他优化
		Immutable: false,
	}
}
