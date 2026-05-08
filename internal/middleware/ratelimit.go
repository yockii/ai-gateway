package middleware

import (
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/pkg/api"
)

// RateLimiter 限流器
type RateLimiter struct {
	// 用户请求记录
	userRequests map[string]*UserRequestStats
	mutex        sync.RWMutex

	// 限流配置
	maxRequestsPerMinute int
	cleanupInterval      time.Duration
}

// UserRequestStats 用户请求统计
type UserRequestStats struct {
	RequestTimes []time.Time
	LastCleanup  time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(maxRequestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		userRequests:         make(map[string]*UserRequestStats),
		maxRequestsPerMinute: maxRequestsPerMinute,
		cleanupInterval:      time.Minute,
	}

	// 启动后台清理任务
	go rl.cleanupExpiredRecords()

	return rl
}

// RateLimit 限流中间件
func (rl *RateLimiter) RateLimit() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			// 如果没有用户 ID，跳过限流检查
			return c.Next()
		}

		// 检查是否超过限流
		allowed, retryAfter := rl.checkRateLimit(userID)
		if !allowed {
			log.Printf("限流触发: 用户=%s IP=%s", userID, c.IP())
			c.Set("Retry-After", retryAfter.String())
			return c.Status(fiber.StatusTooManyRequests).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Rate limit exceeded",
					Type:    "rate_limit_error",
					Code:    fiber.StatusTooManyRequests,
				},
			})
		}

		// 记录本次请求
		rl.recordRequest(userID)

		return c.Next()
	}
}

// checkRateLimit 检查是否超过限流
func (rl *RateLimiter) checkRateLimit(userID string) (bool, time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	stats, exists := rl.userRequests[userID]

	if !exists {
		return true, 0
	}

	// 清理过期的请求记录（超过 1 分钟）
	cutoffTime := now.Add(-time.Minute)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range stats.RequestTimes {
		if reqTime.After(cutoffTime) {
			validRequests = append(validRequests, reqTime)
		}
	}
	stats.RequestTimes = validRequests

	// 检查是否超过限制
	if len(stats.RequestTimes) >= rl.maxRequestsPerMinute {
		// 计算最早请求的过期时间
		oldestRequest := stats.RequestTimes[0]
		retryAfter := time.Minute - now.Sub(oldestRequest)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	return true, 0
}

// recordRequest 记录请求
func (rl *RateLimiter) recordRequest(userID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	stats, exists := rl.userRequests[userID]

	if !exists {
		stats = &UserRequestStats{
			RequestTimes: make([]time.Time, 0),
			LastCleanup:  now,
		}
		rl.userRequests[userID] = stats
	}

	stats.RequestTimes = append(stats.RequestTimes, now)
}

// cleanupExpiredRecords 清理过期记录
func (rl *RateLimiter) cleanupExpiredRecords() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()

		// 清理超过 5 分钟没有活动的用户记录
		for userID, stats := range rl.userRequests {
			if now.Sub(stats.LastCleanup) > 5*time.Minute {
				delete(rl.userRequests, userID)
			}
		}

		rl.mutex.Unlock()
		log.Printf("限流器清理完成，当前活跃用户数: %d", len(rl.userRequests))
	}
}

// 全局限流器实例（默认配置）
var defaultRateLimiter = NewRateLimiter(60) // 每分钟 60 次请求

// RateLimit 使用默认限流器的中间件
func RateLimit() fiber.Handler {
	return defaultRateLimiter.RateLimit()
}
