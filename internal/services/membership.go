package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// MembershipService 会员服务
type MembershipService struct {
	db *database.DB
}

// NewMembershipService 创建会员服务
func NewMembershipService(db *database.DB) (*MembershipService, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	service := &MembershipService{db: db}
	log.Println("✅ 会员服务初始化成功")
	return service, nil
}

// GetUserMembership 获取用户会员信息
func (ms *MembershipService) GetUserMembership(ctx context.Context, userID string) (*UserMembershipInfo, error) {
	var userMembership models.UserMembership
	err := ms.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		First(&userMembership).Error

	if err != nil {
		// 检查是否有默认会员等级
		var defaultTier models.MembershipTier
		err := ms.db.Where("name = ?", "basic").First(&defaultTier).Error
		if err != nil {
			return nil, fmt.Errorf("user has no membership and no default tier found")
		}

		return &UserMembershipInfo{
			UserID:           userID,
			MembershipTierID: defaultTier.ID,
			TierName:         defaultTier.Name,
			DisplayName:      defaultTier.DisplayName,
			Level:            defaultTier.Level,
			EffectiveAt:      time.Now().UTC(),
		}, nil
	}

	// 查询会员等级信息
	var tier models.MembershipTier
	err = ms.db.WithContext(ctx).Where("id = ?", userMembership.MembershipTierID).First(&tier).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch membership tier: %w", err)
	}

	return &UserMembershipInfo{
		UserID:           userID,
		MembershipTierID: userMembership.MembershipTierID,
		TierName:         tier.Name,
		DisplayName:      tier.DisplayName,
		Level:            tier.Level,
		EffectiveAt:      userMembership.EffectiveAt,
		ExpiresAt:        userMembership.ExpiresAt,
	}, nil
}

// GetModelDiscount 获取模型折扣率 (per D-15: 差异化定价折扣)
func (ms *MembershipService) GetModelDiscount(ctx context.Context, userID, modelID string) (float64, error) {
	// 获取用户会员信息
	membership, err := ms.GetUserMembership(ctx, userID)
	if err != nil {
		return 0, err
	}

	// 查询该会员等级对该模型的折扣
	var discount models.MembershipDiscount
	err = ms.db.WithContext(ctx).
		Where("membership_tier_id = ? AND model_id = ? AND is_active = ?",
			membership.MembershipTierID, modelID, true).
		First(&discount).Error

	if err != nil {
		// 没有特定折扣，返回 0
		return 0, nil
	}

	return discount.DiscountRate, nil
}

// ApplyDiscount 应用折扣计算价格
func (ms *MembershipService) ApplyDiscount(originalPrice, discountRate float64) float64 {
	if discountRate <= 0 || discountRate >= 1 {
		return originalPrice
	}
	return originalPrice * (1 - discountRate)
}

// CalculatePrice 计算用户实际价格 (会员价)
func (ms *MembershipService) CalculatePrice(ctx context.Context, userID, modelID string, basePrice float64) (float64, error) {
	discount, err := ms.GetModelDiscount(ctx, userID, modelID)
	if err != nil {
		return basePrice, err
	}

	return ms.ApplyDiscount(basePrice, discount), nil
}

// UpdateUserMembership 更新用户会员 (per D-16: 即时生效)
func (ms *MembershipService) UpdateUserMembership(ctx context.Context, userID, tierID string) error {
	// 查找现有会员记录
	var existing models.UserMembership
	err := ms.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&existing).Error

	now := time.Now().UTC()

	if err == nil {
		// 更新现有记录
		existing.MembershipTierID = tierID
		existing.EffectiveAt = now // 即时生效
		existing.IsActive = true
		existing.UpdatedAt = now

		if err := ms.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update membership: %w", err)
		}
	} else {
		// 创建新记录
		newMembership := &models.UserMembership{
			ID:               xid.New().String(),
			UserID:           userID,
			MembershipTierID: tierID,
			EffectiveAt:      now, // 即时生效
			IsActive:         true,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := ms.db.WithContext(ctx).Create(newMembership).Error; err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}
	}

	log.Printf("✅ 会员更新成功: 用户=%s 等级=%s", userID, tierID)
	return nil
}

// GetAllMembershipTiers 获取所有会员等级
func (ms *MembershipService) GetAllMembershipTiers(ctx context.Context) ([]models.MembershipTier, error) {
	var tiers []models.MembershipTier
	err := ms.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("level ASC").
		Find(&tiers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch membership tiers: %w", err)
	}

	return tiers, nil
}

// CreateMembershipTier 创建会员等级
func (ms *MembershipService) CreateMembershipTier(ctx context.Context, tier *models.MembershipTier) error {
	tier.ID = xid.New().String()
	tier.CreatedAt = time.Now().UTC()
	tier.UpdatedAt = time.Now().UTC()

	if err := ms.db.WithContext(ctx).Create(tier).Error; err != nil {
		return fmt.Errorf("failed to create membership tier: %w", err)
	}

	log.Printf("✅ 会员等级创建成功: 名称=%s 级别=%d", tier.Name, tier.Level)
	return nil
}

// SetModelDiscount 设置模型折扣
func (ms *MembershipService) SetModelDiscount(ctx context.Context, discount *models.MembershipDiscount) error {
	discount.ID = xid.New().String()
	discount.CreatedAt = time.Now().UTC()
	discount.UpdatedAt = time.Now().UTC()
	discount.IsActive = true

	// 检查是否已存在
	var existing models.MembershipDiscount
	err := ms.db.WithContext(ctx).
		Where("membership_tier_id = ? AND model_id = ?", discount.MembershipTierID, discount.ModelID).
		First(&existing).Error

	if err == nil {
		// 更新现有折扣
		existing.DiscountRate = discount.DiscountRate
		existing.UpdatedAt = time.Now().UTC()
		return ms.db.WithContext(ctx).Save(&existing).Error
	}

	// 创建新折扣
	if err := ms.db.WithContext(ctx).Create(discount).Error; err != nil {
		return fmt.Errorf("failed to create discount: %w", err)
	}

	log.Printf("✅ 模型折扣设置成功: 等级=%s 模型=%s 折扣=%.2f",
		discount.MembershipTierID, discount.ModelID, discount.DiscountRate)
	return nil
}

// UserMembershipInfo 用户会员信息
type UserMembershipInfo struct {
	UserID           string     `json:"user_id"`
	MembershipTierID string     `json:"membership_tier_id"`
	TierName         string     `json:"tier_name"`
	DisplayName      string     `json:"display_name"`
	Level            int        `json:"level"`
	EffectiveAt      time.Time  `json:"effective_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}
