package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/cache"
)

// CacheMiddlewareConfig 缓存中间件配置
type CacheMiddlewareConfig struct {
	TTL                 time.Duration // 缓存过期时间
	SkipPaths           []string      // 跳过缓存的路径
	CacheableMethods    []string      // 可缓存的方法
	CacheableStatusCodes []int        // 可缓存的状态码
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheMiddlewareConfig {
	return &CacheMiddlewareConfig{
		TTL:                 30 * time.Minute,
		SkipPaths:           []string{"/health", "/metrics"},
		CacheableMethods:    []string{"GET", "HEAD"},
		CacheableStatusCodes: []int{200, 301, 302},
	}
}

// CacheMiddleware 创建基于请求路径的缓存中间件
func CacheMiddleware(redis *cache.RedisClient, ttl time.Duration) fiber.Handler {
	return CacheMiddlewareWithConfig(redis, &CacheMiddlewareConfig{
		TTL: ttl,
	})
}

// CacheMiddlewareWithConfig 使用配置创建缓存中间件
func CacheMiddlewareWithConfig(redis *cache.RedisClient, config *CacheMiddlewareConfig) fiber.Handler {
	if config == nil {
		config = DefaultCacheConfig()
	}

	return func(c fiber.Ctx) error {
		// 检查方法是否可缓存
		if !isCacheableMethod(c.Method(), config.CacheableMethods) {
			return c.Next()
		}

		// 检查路径是否应该跳过
		if shouldSkipPath(c.Path(), config.SkipPaths) {
			return c.Next()
		}

		// 生成缓存键
		cacheKey := generateCacheKey(c)

		// 尝试从缓存获取
		if cached, err := redis.Get(c.Context(), cacheKey); err == nil {
			c.Set("Content-Type", "application/json")
			c.Set("X-Cache", "HIT")
			return c.SendString(cached)
		}

		// 继续处理请求
		if err := c.Next(); err != nil {
			return err
		}

		// 只缓存成功的响应
		if isCacheableStatusCode(c.Response().StatusCode(), config.CacheableStatusCodes) {
			responseBody := c.Response().Body()
			if len(responseBody) > 0 {
				_ = redis.Set(c.Context(), cacheKey, responseBody, config.TTL)
				c.Set("X-Cache", "MISS")
			}
		}

		return nil
	}
}

// InvalidateCacheMiddleware 缓存失效中间件
// 用于在数据更新时清除相关缓存
func InvalidateCacheMiddleware(redis *cache.RedisClient) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 先执行请求
		if err := c.Next(); err != nil {
			return err
		}

		// 请求成功后，使相关缓存失效
		if c.Response().StatusCode() < 300 {
			path := c.Path()
			method := c.Method()

			invalidateCacheByPath(c, redis, path, method)
		}

		return nil
	}
}

// CacheByPath 为特定路径创建缓存中间件
func CacheByPath(redis *cache.RedisClient, pathPrefix string, ttl time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !strings.HasPrefix(c.Path(), pathPrefix) {
			return c.Next()
		}

		if c.Method() != "GET" && c.Method() != "HEAD" {
			return c.Next()
		}

		cacheKey := generateCacheKey(c)

		if cached, err := redis.Get(c.Context(), cacheKey); err == nil {
			c.Set("Content-Type", "application/json")
			c.Set("X-Cache", "HIT")
			return c.SendString(cached)
		}

		if err := c.Next(); err != nil {
			return err
		}

		if c.Response().StatusCode() == 200 {
			responseBody := c.Response().Body()
			if len(responseBody) > 0 {
				_ = redis.Set(c.Context(), cacheKey, responseBody, ttl)
				c.Set("X-Cache", "MISS")
			}
		}

		return nil
	}
}

// BypassCache 跳过缓存的中间件标记
func BypassCache(c fiber.Ctx) error {
	c.Locals("bypass_cache", true)
	return c.Next()
}

// shouldBypassCache 检查是否应该跳过缓存
func shouldBypassCache(c fiber.Ctx) bool {
	bypass, ok := c.Locals("bypass_cache").(bool)
	return ok && bypass
}

// isCacheableMethod 检查方法是否可缓存
func isCacheableMethod(method string, cacheableMethods []string) bool {
	for _, m := range cacheableMethods {
		if m == method {
			return true
		}
	}
	return false
}

// shouldSkipPath 检查路径是否应该跳过
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// isCacheableStatusCode 检查状态码是否可缓存
func isCacheableStatusCode(statusCode int, cacheableStatusCodes []int) bool {
	for _, code := range cacheableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// generateCacheKey 生成缓存键
func generateCacheKey(c fiber.Ctx) string {
	// 使用完整的 URL（包含路径和查询参数）
	originalURL := c.OriginalURL()
	cacheKey := fmt.Sprintf("response:%s", originalURL)

	// 如果有认证头，包含用户信息（实现用户级缓存隔离）
	if authHeader := c.Get("Authorization"); authHeader != "" {
		// 使用哈希或简化版本避免暴露完整 token
		if len(authHeader) > 20 {
			cacheKey = fmt.Sprintf("user:%s:%s", authHeader[:20], cacheKey)
		}
	}

	return cacheKey
}

// invalidateCacheByPath 根据路径使缓存失效
func invalidateCacheByPath(c fiber.Ctx, redis *cache.RedisClient, path, method string) {
	// 只在修改操作时使缓存失效
	if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
		return
	}

	var keysToDelete []string

	switch {
	case strings.Contains(path, "/admin/models"):
		keysToDelete = append(keysToDelete, "models:list")

	case strings.Contains(path, "/admin/suppliers"):
		keysToDelete = append(keysToDelete, "suppliers:list")

	case strings.Contains(path, "/admin/memberships"):
		// 删除所有会员套餐缓存
		keysToDelete = append(keysToDelete, "membership:tier:*")

	case strings.Contains(path, "/user/keys"):
		// 从请求中获取 user_id
		if userID := c.Locals("user_id"); userID != nil {
			keysToDelete = append(keysToDelete,
				fmt.Sprintf("user:quota:%s", userID),
				fmt.Sprintf("user:info:%s", userID),
			)
		}
	}

	if len(keysToDelete) > 0 {
		_ = redis.Delete(c.Context(), keysToDelete...)
	}
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits   int64
	Misses int64
}

// CacheMiddlewareWithStats 带统计的缓存中间件
func CacheMiddlewareWithStats(redis *cache.RedisClient, ttl time.Duration, stats *CacheStats) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() != "GET" && c.Method() != "HEAD" {
			return c.Next()
		}

		cacheKey := generateCacheKey(c)

		if cached, err := redis.Get(c.Context(), cacheKey); err == nil {
			stats.Hits++
			c.Set("Content-Type", "application/json")
			c.Set("X-Cache", "HIT")
			return c.SendString(cached)
		}

		stats.Misses++

		if err := c.Next(); err != nil {
			return err
		}

		if c.Response().StatusCode() == 200 {
			responseBody := c.Response().Body()
			if len(responseBody) > 0 {
				_ = redis.Set(c.Context(), cacheKey, responseBody, ttl)
				c.Set("X-Cache", "MISS")
			}
		}

		return nil
	}
}

// CacheControlHeader 添加 Cache-Control 头的中间件
func CacheControlHeader(maxAge int) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}

		if c.Method() == "GET" && c.Response().StatusCode() == 200 {
			c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
		}

		return nil
	}
}

// ETag 支持 ETag 的缓存中间件
func ETag() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() != "GET" && c.Method() != "HEAD" {
			return c.Next()
		}

		// 检查 If-None-Match 头
		ifNoneMatch := c.Get("If-None-Match")
		if ifNoneMatch != "" {
			// 简单实现：基于请求路径生成 ETag
			// 生产环境应该基于响应内容生成
			etag := fmt.Sprintf(`"%s"`, c.Path())
			if ifNoneMatch == etag {
				c.Status(304)
				return nil
			}
		}

		if err := c.Next(); err != nil {
			return err
		}

		// 添加 ETag 头
		if c.Response().StatusCode() == 200 {
			etag := fmt.Sprintf(`"%s"`, c.Path())
			c.Set("ETag", etag)
		}

		return nil
	}
}
