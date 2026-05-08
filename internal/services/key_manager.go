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
