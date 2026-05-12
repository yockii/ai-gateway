package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/crypto"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"gorm.io/gorm"
)

type SupplierApiKeyService struct {
	db     *database.DB
	crypto *crypto.EncryptionService
}

func NewSupplierApiKeyService(db *database.DB) (*SupplierApiKeyService, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	cryptoSvc, err := crypto.NewEncryptionService()
	if err != nil {
		return nil, fmt.Errorf("failed to create crypto service: %w", err)
	}
	service := &SupplierApiKeyService{db: db, crypto: cryptoSvc}
	log.Println("✅ 供应商 API Key 服务初始化成功")
	return service, nil
}

type ApiKeyStats struct {
	KeyID           string     `json:"key_id"`
	KeyName         string     `json:"key_name"`
	CurrentRequests int64      `json:"current_requests"`
	MaxRequests     int64      `json:"max_requests"`
	UsagePercentage float64    `json:"usage_percentage"`
	LastUsedAt      *time.Time `json:"last_used_at"`
	IsPrimary       bool       `json:"is_primary"`
	IsActive        bool       `json:"is_active"`
}

func (s *SupplierApiKeyService) CreateApiKey(ctx context.Context, key *models.SupplierApiKey, plaintextKey string) error {
	encrypted, err := s.crypto.Encrypt(plaintextKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt key: %w", err)
	}
	key.KeyPrefix = crypto.GenerateKeyPrefix(plaintextKey)
	key.KeyValueEncrypted = encrypted
	if key.ID == "" {
		key.ID = xid.New().String()
	}
	now := time.Now().UTC()
	key.CreatedAt = now
	key.UpdatedAt = now
	if key.IsPrimary {
		s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).Where("supplier_id = ? AND is_primary = ?", key.SupplierID, true).Update("is_primary", false)
	} else {
		var count int64
		s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).Where("supplier_id = ? AND is_primary = ?", key.SupplierID, true).Count(&count)
		if count == 0 {
			key.IsPrimary = true
		}
	}
	if err := s.db.WithContext(ctx).Create(key).Error; err != nil {
		return fmt.Errorf("failed to create api key: %w", err)
	}
	log.Printf("✅ API Key 创建成功: 供应商=%s 名称=%s", key.SupplierID, key.Name)
	return nil
}

func (s *SupplierApiKeyService) ListApiKeys(ctx context.Context, supplierID string) ([]*models.SupplierApiKey, error) {
	var keys []*models.SupplierApiKey
	err := s.db.WithContext(ctx).Where("supplier_id = ?", supplierID).Order("priority ASC, created_at DESC").Find(&keys).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	return keys, nil
}

func (s *SupplierApiKeyService) GetBestApiKey(ctx context.Context, supplierID string) (*models.SupplierApiKey, string, error) {
	var keys []*models.SupplierApiKey
	err := s.db.WithContext(ctx).Where("supplier_id = ? AND is_active = ?", supplierID, true).Order("is_primary DESC, priority ASC").Find(&keys).Error
	if err != nil {
		return nil, "", fmt.Errorf("failed to get api keys: %w", err)
	}
	now := time.Now().UTC()
	for _, key := range keys {
		if key.ExpireAt != nil && now.After(*key.ExpireAt) {
			continue
		}
		if key.MaxRequests > 0 && key.CurrentRequests >= key.MaxRequests {
			continue
		}
		plaintext, err := s.crypto.Decrypt(key.KeyValueEncrypted)
		if err != nil {
			log.Printf("警告: 解密密钥失败: %s", err)
			continue
		}
		return key, plaintext, nil
	}
	return nil, "", fmt.Errorf("no available api key for supplier: %s", supplierID)
}

func (s *SupplierApiKeyService) GetApiKeyStats(ctx context.Context, keyID string) (*ApiKeyStats, error) {
	var key models.SupplierApiKey
	err := s.db.WithContext(ctx).Where("id = ?", keyID).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}
	stats := &ApiKeyStats{
		KeyID:           key.ID,
		KeyName:         key.Name,
		CurrentRequests: key.CurrentRequests,
		MaxRequests:     key.MaxRequests,
		LastUsedAt:      key.LastUsedAt,
		IsPrimary:       key.IsPrimary,
		IsActive:        key.IsActive,
	}
	if key.MaxRequests > 0 {
		stats.UsagePercentage = float64(key.CurrentRequests) / float64(key.MaxRequests) * 100
	}
	return stats, nil
}

// SetPrimaryApiKey 设置主密钥（修复 CR-02: 在事务内完成所有操作）
func (s *SupplierApiKeyService) SetPrimaryApiKey(ctx context.Context, keyID string) error {
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 在事务内获取密钥
	var key models.SupplierApiKey
	err := tx.Where("id = ?", keyID).First(&key).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get api key: %w", err)
	}

	// 重置所有主密钥状态
	result := tx.Model(&models.SupplierApiKey{}).
		Where("supplier_id = ? AND is_primary = ?", key.SupplierID, true).
		Update("is_primary", false)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to reset primary keys: %w", result.Error)
	}

	// 设置新主密钥
	key.IsPrimary = true
	key.UpdatedAt = time.Now().UTC()
	if err := tx.Save(&key).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update primary key: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("✅ 主密钥设置成功: %s", keyID)
	return nil
}

func (s *SupplierApiKeyService) RotateApiKey(ctx context.Context, supplierID string) error {
	var currentKey models.SupplierApiKey
	err := s.db.WithContext(ctx).Where("supplier_id = ? AND is_primary = ?", supplierID, true).First(&currentKey).Error
	if err != nil {
		return fmt.Errorf("failed to get current primary key: %w", err)
	}
	var backupKey models.SupplierApiKey
	err = s.db.WithContext(ctx).Where("supplier_id = ? AND id != ? AND is_active = ?", supplierID, currentKey.ID, true).Order("priority ASC").First(&backupKey).Error
	if err != nil {
		return fmt.Errorf("no backup key available for rotation: %w", err)
	}
	return s.SetPrimaryApiKey(ctx, backupKey.ID)
}

// IncrementUsage 增加使用计数（修复 CR-12: 使用原子操作）
func (s *SupplierApiKeyService) IncrementUsage(ctx context.Context, keyID string) error {
	now := time.Now().UTC()
	result := s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).
		Where("id = ?", keyID).
		Updates(map[string]interface{}{
			"current_requests": gorm.Expr("current_requests + ?", 1),
			"last_used_at":     now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to increment usage: %w", result.Error)
	}
	return nil
}

func (s *SupplierApiKeyService) ShouldRotate(ctx context.Context, keyID string) (bool, error) {
	var key models.SupplierApiKey
	err := s.db.WithContext(ctx).Where("id = ?", keyID).First(&key).Error
	if err != nil {
		return false, fmt.Errorf("failed to get api key: %w", err)
	}
	return key.ShouldRotate(), nil
}

func (s *SupplierApiKeyService) UpdateApiKey(ctx context.Context, key *models.SupplierApiKey) error {
	key.UpdatedAt = time.Now().UTC()
	result := s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).Where("id = ?", key.ID).Updates(key)
	if result.Error != nil {
		return fmt.Errorf("failed to update api key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("api key not found: %s", key.ID)
	}
	log.Printf("✅ API Key 更新成功: %s", key.ID)
	return nil
}

func (s *SupplierApiKeyService) DeleteApiKey(ctx context.Context, keyID string) error {
	var key models.SupplierApiKey
	err := s.db.WithContext(ctx).Where("id = ?", keyID).First(&key).Error
	if err != nil {
		return fmt.Errorf("failed to get api key: %w", err)
	}
	if key.IsPrimary {
		var count int64
		s.db.WithContext(ctx).Model(&models.SupplierApiKey{}).Where("supplier_id = ? AND id != ?", key.SupplierID, keyID).Count(&count)
		if count == 0 {
			return fmt.Errorf("cannot delete primary key: no other keys available")
		}
	}
	result := s.db.WithContext(ctx).Where("id = ?", keyID).Delete(&models.SupplierApiKey{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete api key: %w", result.Error)
	}
	log.Printf("✅ API Key 删除成功: %s", keyID)
	return nil
}

// GetApiKeyByID 根据 ID 获取 API Key（用于审计日志）
func (s *SupplierApiKeyService) GetApiKeyByID(ctx context.Context, keyID string) (*models.SupplierApiKey, error) {
	var key models.SupplierApiKey
	err := s.db.WithContext(ctx).Where("id = ?", keyID).First(&key).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}
	return &key, nil
}
