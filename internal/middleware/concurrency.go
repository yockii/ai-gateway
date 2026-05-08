package middleware

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/yockii/ai-gateway/pkg/api"
)

// ConcurrencyLimiter 并发限制器
type ConcurrencyLimiter struct {
	redis *redis.Client
}

// NewConcurrencyLimiter 创建并发限制器
func NewConcurrencyLimiter(redisClient *redis.Client) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		redis: redisClient,
	}
}

// ConcurrentLimit 并发限制中间件
func (cl *ConcurrencyLimiter) ConcurrentLimit() fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		userID := GetUserID(c)
		apiKey := c.Locals("api_key").(string)
		modelID := c.Query("model", "") // May be empty for list endpoints

		// 从上下文获取 KeyInfo (由 KeyManager 设置)
		keyInfo, ok := c.Locals("key_info").(*KeyInfo)
		if !ok {
			// Skip if key info not available
			return c.Next()
		}

		// 检查默认并发限制
		if keyInfo.ConcurrencyLimit > 0 {
			current, err := cl.incrementConcurrency(ctx, apiKey, "")
			if err != nil {
				log.Printf("并发检查错误: %v", err)
				return c.Next() // 失败时放行
			}

			if current > keyInfo.ConcurrencyLimit {
				cl.decrementConcurrency(ctx, apiKey, "")
				log.Printf("并发限制触发: 用户=%s 当前=%d 限制=%d",
					userID, current, keyInfo.ConcurrencyLimit)
				return c.Status(fiber.StatusTooManyRequests).JSON(api.ErrorResponse{
					Error: api.ErrorDetail{
						Message: "Concurrent request limit exceeded",
						Type:    "concurrency_limit_error",
						Code:    fiber.StatusTooManyRequests,
					},
				})
			}
		}

		// 检查模型特定并发限制 (per D-12)
		if modelID != "" && keyInfo.ModelConcurrency != nil {
			if modelLimit, exists := keyInfo.ModelConcurrency[modelID]; exists {
				current, err := cl.incrementConcurrency(ctx, apiKey, modelID)
				if err != nil {
					log.Printf("模型并发检查错误: %v", err)
					return c.Next()
				}

				if current > modelLimit {
					cl.decrementConcurrency(ctx, apiKey, modelID)
					log.Printf("模型并发限制触发: 用户=%s 模型=%s 当前=%d 限制=%d",
						userID, modelID, current, modelLimit)
					return c.Status(fiber.StatusTooManyRequests).JSON(api.ErrorResponse{
						Error: api.ErrorDetail{
							Message: fmt.Sprintf("Concurrent request limit exceeded for model %s", modelID),
							Type:    "model_concurrency_limit_error",
							Code:    fiber.StatusTooManyRequests,
						},
					})
				}
			}
		}

		// 继续处理请求
		return c.Next()
	}
}

// incrementConcurrency 增加并发计数 (per D-13)
func (cl *ConcurrencyLimiter) incrementConcurrency(ctx context.Context, apiKey, modelID string) (int64, error) {
	key := cl.concurrencyKey(apiKey, modelID)
	result, err := cl.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment concurrency: %w", err)
	}

	// 设置过期时间 (防止计数器永久存在)
	if result == 1 {
		cl.redis.Expire(ctx, key, 3600) // 1 hour
	}

	return result, nil
}

// decrementConcurrency 减少并发计数
func (cl *ConcurrencyLimiter) decrementConcurrency(ctx context.Context, apiKey, modelID string) {
	key := cl.concurrencyKey(apiKey, modelID)
	cl.redis.Decr(ctx, key)
}

// concurrencyKey 生成并发计数 Redis key
func (cl *ConcurrencyLimiter) concurrencyKey(apiKey, modelID string) string {
	if modelID == "" {
		return fmt.Sprintf("concurrency:%s", apiKey)
	}
	return fmt.Sprintf("concurrency:%s:model:%s", apiKey, modelID)
}

// GetModelConcurrency 获取当前模型并发数
func (cl *ConcurrencyLimiter) GetModelConcurrency(ctx context.Context, apiKey, modelID string) (int64, error) {
	key := cl.concurrencyKey(apiKey, modelID)
	result, err := cl.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(result, 10, 64)
}

// KeyInfo Key 信息 (从 KeyManager 传递)
type KeyInfo struct {
	UserID           string
	KeyID            string
	ConcurrencyLimit int64
	ModelConcurrency map[string]int64
}

// 全局实例 (需要初始化)
var defaultConcurrencyLimiter *ConcurrencyLimiter

// ConcurrentLimit 使用默认限制器的中间件
func ConcurrentLimit() fiber.Handler {
	if defaultConcurrencyLimiter == nil {
		// 未初始化时跳过
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}
	return defaultConcurrencyLimiter.ConcurrentLimit()
}
