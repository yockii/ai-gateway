package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// KeyManager API Key 管理器
type KeyManager struct {
	db *database.DB
}

// NewKeyManager 创建 Key 管理器
func NewKeyManager(db *database.DB) (*KeyManager, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	manager := &KeyManager{db: db}
	log.Println("✅ Key 管理器初始化成功")
	return manager, nil
}

// CreateKey 创建新的 API Key (per D-11: UUID v4 with sk- prefix)
func (km *KeyManager) CreateKey(ctx context.Context, userID, name string, opts CreateKeyOptions) (*models.UserAPIKey, error) {
	// 生成 UUID v4 格式的 API Key
	keyUUID := uuid.New()
	keyValue := fmt.Sprintf("sk-%s", keyUUID.String())

	key := &models.UserAPIKey{
		ID:                xid.New().String(),
		UserID:            userID,
		KeyValue:          keyValue,
		Name:              name,
		QuotaDaily:        opts.QuotaDaily,
		QuotaMonthly:      opts.QuotaMonthly,
		ConcurrencyLimit:  opts.ConcurrencyLimit,
		ModelConcurrency:  opts.ModelConcurrency,
		ExpiresAt:         opts.ExpiresAt,
		IsActive:          true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := km.db.WithContext(ctx).Create(key).Error; err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	log.Printf("✅ API Key 创建成功: 用户=%s 名称=%s", userID, name)
	return key, nil
}

// ValidateKey 验证 API Key 并返回用户信息
func (km *KeyManager) ValidateKey(ctx context.Context, keyValue string) (*KeyInfo, error) {
	// 检查格式 (per D-11)
	if len(keyValue) < 4 || keyValue[:3] != "sk-" {
		return nil, fmt.Errorf("invalid API key format")
	}

	var key models.UserAPIKey
	err := km.db.WithContext(ctx).
		Where("key_value = ? AND is_active = ?", keyValue, true).
		First(&key).Error

	if err != nil {
		return nil, fmt.Errorf("API key not found or inactive")
	}

	// 检查过期时间
	if !key.ExpiresAt.IsZero() && time.Now().UTC().After(key.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	return &KeyInfo{
		UserID:           key.UserID,
		KeyID:            key.ID,
		ConcurrencyLimit: key.ConcurrencyLimit,
		ModelConcurrency: key.ModelConcurrency,
	}, nil
}

// GetUserKeys 获取用户的所有 API Key (不返回 KeyValue)
func (km *KeyManager) GetUserKeys(ctx context.Context, userID string) ([]models.UserAPIKey, error) {
	var keys []models.UserAPIKey
	err := km.db.WithContext(ctx).
		Select("id", "user_id", "name", "quota_daily", "quota_monthly",
			"concurrency_limit", "model_concurrency", "expires_at", "is_active",
			"created_at", "updated_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&keys).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user keys: %w", err)
	}

	return keys, nil
}

// UpdateKeyStatus 更新 Key 状态
func (km *KeyManager) UpdateKeyStatus(ctx context.Context, keyID, userID string, isActive bool) error {
	result := km.db.WithContext(ctx).
		Model(&models.UserAPIKey{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Update("is_active", isActive)

	if result.Error != nil {
		return fmt.Errorf("failed to update key status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("key not found or access denied")
	}

	status := "enabled"
	if !isActive {
		status = "disabled"
	}
	log.Printf("✅ API Key %s: 用户=%s KeyID=%s", status, userID, keyID)

	return nil
}

// DeleteKey 删除 API Key
func (km *KeyManager) DeleteKey(ctx context.Context, keyID, userID string) error {
	result := km.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", keyID, userID).
		Delete(&models.UserAPIKey{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("key not found or access denied")
	}

	log.Printf("✅ API Key 删除成功: 用户=%s KeyID=%s", userID, keyID)
	return nil
}

// CreateKeyOptions 创建 Key 选项
type CreateKeyOptions struct {
	QuotaDaily        int64            // 每日额度
	QuotaMonthly      int64            // 每月额度
	ConcurrencyLimit  int64            // 默认并发限制
	ModelConcurrency  map[string]int64 // 模型特定并发限制
	ExpiresAt         time.Time        // 过期时间
}

// KeyInfo Key 验证返回信息
type KeyInfo struct {
	UserID           string
	KeyID            string
	ConcurrencyLimit int64
	ModelConcurrency map[string]int64
}

// UpdateKey 更新 API Key 配置
func (km *KeyManager) UpdateKey(ctx context.Context, keyID, userID string, opts UpdateKeyOptions) error {
	result := km.db.WithContext(ctx).Model(&models.UserAPIKey{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Updates(map[string]interface{}{
			"name":              opts.Name,
			"quota_daily":       opts.QuotaDaily,
			"quota_monthly":     opts.QuotaMonthly,
			"concurrency_limit": opts.ConcurrencyLimit,
			"model_concurrency": opts.ModelConcurrency,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("key not found or access denied")
	}

	log.Printf("✅ API Key 更新成功: 用户=%s KeyID=%s", userID, keyID)
	return nil
}

// GetKeyStats 获取 API Key 使用统计（修复 CR-10: 使用范围查询）
func (km *KeyManager) GetKeyStats(ctx context.Context, keyID string) (*KeyStats, error) {
	var key models.UserAPIKey
	err := km.db.WithContext(ctx).Where("id = ?", keyID).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("key not found: %w", err)
	}

	now := time.Now().UTC()

	// 获取今日统计（使用范围查询代替 DATE() 函数）
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayEnd := todayStart.AddDate(0, 0, 1)
	var dailyStats struct {
		Requests int64
		Tokens   int64
		Cost     float64
	}
	km.db.WithContext(ctx).Model(&models.UsageRecord{}).
		Where("key_id = ? AND created_at >= ? AND created_at < ?", keyID, todayStart, todayEnd).
		Select("COUNT(*) as requests, COALESCE(SUM(total_tokens), 0) as tokens, COALESCE(SUM(selling_price), 0) as cost").
		Scan(&dailyStats)

	// 获取本月统计（使用范围查询）
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var monthlyStats struct {
		Requests int64
		Tokens   int64
		Cost     float64
	}
	km.db.WithContext(ctx).Model(&models.UsageRecord{}).
		Where("key_id = ? AND created_at >= ?", keyID, monthStart).
		Select("COUNT(*) as requests, COALESCE(SUM(total_tokens), 0) as tokens, COALESCE(SUM(selling_price), 0) as cost").
		Scan(&monthlyStats)

	// 获取总计
	var totalStats struct {
		Requests int64
		Tokens   int64
		Cost     float64
	}
	km.db.WithContext(ctx).Model(&models.UsageRecord{}).
		Where("key_id = ?", keyID).
		Select("COUNT(*) as requests, COALESCE(SUM(total_tokens), 0) as tokens, COALESCE(SUM(selling_price), 0) as cost").
		Scan(&totalStats)

	return &KeyStats{
		KeyID:          keyID,
		TotalRequests:  totalStats.Requests,
		TotalTokens:    totalStats.Tokens,
		TotalCost:      totalStats.Cost,
		DailyRequests:  dailyStats.Requests,
		DailyTokens:    dailyStats.Tokens,
		DailyCost:      dailyStats.Cost,
		MonthlyRequests: monthlyStats.Requests,
		MonthlyTokens:  monthlyStats.Tokens,
		MonthlyCost:    monthlyStats.Cost,
		QuotaDaily:     key.QuotaDaily,
		QuotaMonthly:   key.QuotaMonthly,
	}, nil
}

// CheckQuota 检查额度限制（修复 CR-10: 使用范围查询）
func (km *KeyManager) CheckQuota(ctx context.Context, keyID string) error {
	var key models.UserAPIKey
	err := km.db.WithContext(ctx).Where("id = ?", keyID).First(&key).Error
	if err != nil {
		return fmt.Errorf("key not found: %w", err)
	}

	now := time.Now().UTC()

	// 检查每日额度（使用范围查询代替 DATE() 函数）
	if key.QuotaDaily > 0 {
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		todayEnd := todayStart.AddDate(0, 0, 1)
		var dailyCount int64
		km.db.WithContext(ctx).Model(&models.UsageRecord{}).
			Where("key_id = ? AND created_at >= ? AND created_at < ?", keyID, todayStart, todayEnd).
			Count(&dailyCount)

		if dailyCount >= key.QuotaDaily {
			return fmt.Errorf("daily quota exceeded: %d/%d", dailyCount, key.QuotaDaily)
		}
	}

	// 检查每月额度（使用范围查询）
	if key.QuotaMonthly > 0 {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		var monthlyCount int64
		km.db.WithContext(ctx).Model(&models.UsageRecord{}).
			Where("key_id = ? AND created_at >= ?", keyID, monthStart).
			Count(&monthlyCount)

		if monthlyCount >= key.QuotaMonthly {
			return fmt.Errorf("monthly quota exceeded: %d/%d", monthlyCount, key.QuotaMonthly)
		}
	}

	return nil
}

// RecordUsage 记录使用量（由 gateway 在请求完成后调用）
func (km *KeyManager) RecordUsage(ctx context.Context, keyID string, usage interface{}) error {
	// 这个方法主要用于更新使用记录的 key_id 字段
	// 实际的 UsageRecord 创建由 gateway 层负责
	return nil
}

// UpdateKeyOptions 更新 Key 选项
type UpdateKeyOptions struct {
	Name             string
	QuotaDaily       int64
	QuotaMonthly     int64
	ConcurrencyLimit int64
	ModelConcurrency map[string]int64
}

// KeyStats Key 使用统计
type KeyStats struct {
	KeyID          string
	TotalRequests  int64
	TotalTokens    int64
	TotalCost      float64
	DailyRequests  int64
	DailyTokens    int64
	DailyCost      float64
	MonthlyRequests int64
	MonthlyTokens  int64
	MonthlyCost    float64
	QuotaDaily     int64
	QuotaMonthly   int64
}
