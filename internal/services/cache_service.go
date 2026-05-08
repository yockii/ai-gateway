package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yockii/ai-gateway/internal/cache"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// CacheService 缓存服务层
type CacheService struct {
	redis *cache.RedisClient
	db    *database.DB
}

// NewCacheService 创建缓存服务
func NewCacheService(redis *cache.RedisClient, db *database.DB) *CacheService {
	return &CacheService{redis: redis, db: db}
}

// 缓存键前缀
const (
	CacheKeyModelList      = "models:list"
	CacheKeyUserQuota      = "user:quota:%s"
	CacheKeyUserInfo       = "user:info:%s"
	CacheKeySupplierList   = "suppliers:list"
	CacheKeyMembershipTier = "membership:tier:%s"
	CacheKeyModelMapping   = "model:mapping:%s"
)

// 缓存过期时间
const (
	CacheTTLShort  = 5 * time.Minute  // 短期缓存：频繁变化的数据
	CacheTTLMedium = 30 * time.Minute // 中期缓存：模型列表等
	CacheTTLLong   = 2 * time.Hour    // 长期缓存：配置数据
)

// GetModelList 获取对外模型列表（带缓存）
func (s *CacheService) GetModelList(ctx context.Context) ([]models.ExternalModel, error) {
	// 尝试从缓存获取
	cached, err := s.redis.Get(ctx, CacheKeyModelList)
	if err == nil {
		var modelList []models.ExternalModel
		if err := json.Unmarshal([]byte(cached), &modelList); err == nil {
			return modelList, nil
		}
	}

	// 缓存未命中，从数据库加载
	var modelList []models.ExternalModel
	if err := s.db.Where("is_active = ?", true).Find(&modelList).Error; err != nil {
		return nil, fmt.Errorf("failed to load models from database: %w", err)
	}

	// 写入缓存
	data, _ := json.Marshal(modelList)
	_ = s.redis.Set(ctx, CacheKeyModelList, data, CacheTTLMedium)

	return modelList, nil
}

// InvalidateModelList 使模型列表缓存失效
func (s *CacheService) InvalidateModelList(ctx context.Context) error {
	return s.redis.Delete(ctx, CacheKeyModelList)
}

// GetModelByID 根据 ID 获取模型（带缓存）
func (s *CacheService) GetModelByID(ctx context.Context, modelID string) (*models.ExternalModel, error) {
	models, err := s.GetModelList(ctx)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.ID == modelID {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model not found: %s", modelID)
}

// GetUserQuota 获取用户配额（带缓存）
func (s *CacheService) GetUserQuota(ctx context.Context, userID string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf(CacheKeyUserQuota, userID)

	// 尝试从缓存获取
	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil {
		var quota map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &quota); err == nil {
			return quota, nil
		}
	}

	// 从数据库加载
	var key models.UserAPIKey
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to load user quota: %w", err)
	}

	quota := map[string]interface{}{
		"quota_daily":       key.QuotaDaily,
		"quota_monthly":     key.QuotaMonthly,
		"concurrency_limit": key.ConcurrencyLimit,
		"model_concurrency": key.ModelConcurrency,
	}

	// 写入缓存（短期缓存，配额变化较快）
	data, _ := json.Marshal(quota)
	_ = s.redis.Set(ctx, cacheKey, data, CacheTTLShort)

	return quota, nil
}

// InvalidateUserQuota 使用户配额缓存失效
func (s *CacheService) InvalidateUserQuota(ctx context.Context, userID string) error {
	cacheKey := fmt.Sprintf(CacheKeyUserQuota, userID)
	return s.redis.Delete(ctx, cacheKey)
}

// GetUserInfo 获取用户信息（带缓存）
func (s *CacheService) GetUserInfo(ctx context.Context, userID string) (*models.User, error) {
	cacheKey := fmt.Sprintf(CacheKeyUserInfo, userID)

	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to load user info: %w", err)
	}

	data, _ := json.Marshal(user)
	_ = s.redis.Set(ctx, cacheKey, data, CacheTTLLong)

	return &user, nil
}

// InvalidateUserInfo 使用户信息缓存失效
func (s *CacheService) InvalidateUserInfo(ctx context.Context, userID string) error {
	cacheKey := fmt.Sprintf(CacheKeyUserInfo, userID)
	return s.redis.Delete(ctx, cacheKey)
}

// GetSupplierList 获取供应商列表（带缓存）
func (s *CacheService) GetSupplierList(ctx context.Context) ([]models.Supplier, error) {
	cached, err := s.redis.Get(ctx, CacheKeySupplierList)
	if err == nil {
		var suppliers []models.Supplier
		if err := json.Unmarshal([]byte(cached), &suppliers); err == nil {
			return suppliers, nil
		}
	}

	var suppliers []models.Supplier
	if err := s.db.Where("is_active = ?", true).Find(&suppliers).Error; err != nil {
		return nil, fmt.Errorf("failed to load suppliers: %w", err)
	}

	data, _ := json.Marshal(suppliers)
	_ = s.redis.Set(ctx, CacheKeySupplierList, data, CacheTTLLong)

	return suppliers, nil
}

// InvalidateSupplierList 使供应商列表缓存失效
func (s *CacheService) InvalidateSupplierList(ctx context.Context) error {
	return s.redis.Delete(ctx, CacheKeySupplierList)
}

// GetMembershipTier 获取会员套餐（带缓存）
func (s *CacheService) GetMembershipTier(ctx context.Context, tierID string) (*models.MembershipTier, error) {
	cacheKey := fmt.Sprintf(CacheKeyMembershipTier, tierID)

	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil {
		var tier models.MembershipTier
		if err := json.Unmarshal([]byte(cached), &tier); err == nil {
			return &tier, nil
		}
	}

	var tier models.MembershipTier
	if err := s.db.Where("id = ?", tierID).First(&tier).Error; err != nil {
		return nil, fmt.Errorf("failed to load membership tier: %w", err)
	}

	data, _ := json.Marshal(tier)
	_ = s.redis.Set(ctx, cacheKey, data, CacheTTLLong)

	return &tier, nil
}

// InvalidateMembershipTier 使会员套餐缓存失效
func (s *CacheService) InvalidateMembershipTier(ctx context.Context, tierID string) error {
	cacheKey := fmt.Sprintf(CacheKeyMembershipTier, tierID)
	return s.redis.Delete(ctx, cacheKey)
}

// GetModelMapping 获取模型映射（带缓存）
func (s *CacheService) GetModelMapping(ctx context.Context, modelName string) (*models.ModelMapping, error) {
	cacheKey := fmt.Sprintf(CacheKeyModelMapping, modelName)

	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil {
		var mapping models.ModelMapping
		if err := json.Unmarshal([]byte(cached), &mapping); err == nil {
			return &mapping, nil
		}
	}

	var mapping models.ModelMapping
	if err := s.db.Where("external_model = ?", modelName).First(&mapping).Error; err != nil {
		return nil, fmt.Errorf("failed to load model mapping: %w", err)
	}

	data, _ := json.Marshal(mapping)
	_ = s.redis.Set(ctx, cacheKey, data, CacheTTLLong)

	return &mapping, nil
}

// InvalidateModelMapping 使模型映射缓存失效
func (s *CacheService) InvalidateModelMapping(ctx context.Context, modelName string) error {
	cacheKey := fmt.Sprintf(CacheKeyModelMapping, modelName)
	return s.redis.Delete(ctx, cacheKey)
}

// GetOrCreateUserAPIKey 获取或创建用户 API Key（带缓存）
func (s *CacheService) GetOrCreateUserAPIKey(ctx context.Context, userID string) (*models.UserAPIKey, error) {
	var key models.UserAPIKey
	err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load user API key: %w", err)
	}

	return &key, nil
}

// CheckUserQuota 检查用户配额（带缓存）
func (s *CacheService) CheckUserQuota(ctx context.Context, userID string) (bool, error) {
	quota, err := s.GetUserQuota(ctx, userID)
	if err != nil {
		return false, err
	}

	// 检查日配额
	if quotaDaily, ok := quota["quota_daily"].(float64); ok && quotaDaily > 0 {
		// TODO: 实现实际的配额检查逻辑
		_ = quotaDaily
	}

	return true, nil
}

// GetUsageStats 获取使用统计（短期缓存）
func (s *CacheService) GetUsageStats(ctx context.Context, userID string, period string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("user:stats:%s:%s", userID, period)

	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil {
		var stats map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return stats, nil
		}
	}

	// 从数据库计算统计
	// TODO: 实现实际的使用统计计算
	stats := map[string]interface{}{
		"total_requests": 0,
		"total_tokens":   0,
		"total_cost":     0.0,
	}

	data, _ := json.Marshal(stats)
	_ = s.redis.Set(ctx, cacheKey, data, CacheTTLShort)

	return stats, nil
}

// InvalidateAllUserCache 清除用户所有相关缓存
func (s *CacheService) InvalidateAllUserCache(ctx context.Context, userID string) error {
	cacheKeys := []string{
		fmt.Sprintf(CacheKeyUserQuota, userID),
		fmt.Sprintf(CacheKeyUserInfo, userID),
		fmt.Sprintf("user:stats:%s:*", userID),
	}

	return s.redis.Delete(ctx, cacheKeys...)
}

// Ping 检查 Redis 连接
func (s *CacheService) Ping(ctx context.Context) error {
	return s.redis.Ping(ctx)
}

// GetRedis 获取 Redis 客户端（用于高级操作）
func (s *CacheService) GetRedis() *cache.RedisClient {
	return s.redis
}
