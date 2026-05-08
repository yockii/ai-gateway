package models

import "time"

// MembershipTier 会员等级 (per D-14)
type MembershipTier struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`        // basic, premium, vip
	DisplayName string    `json:"display_name"`                   // "普通会员", "高级会员"
	Level       int       `json:"level"`                          // Higher = more benefits
	IsActive    bool      `json:"is_active" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MembershipDiscount 会员折扣配置 (per D-15)
type MembershipDiscount struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	MembershipTierID string    `json:"membership_tier_id" gorm:"index"`
	ModelID          string    `json:"model_id" gorm:"index"`
	DiscountRate     float64   `json:"discount_rate"` // 0.0-1.0 (e.g., 0.1 = 10% off)
	IsActive         bool      `json:"is_active" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UserMembership 用户会员关联 (per D-16 - immediate effect)
type UserMembership struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	UserID           string    `json:"user_id" gorm:"index"`
	MembershipTierID string    `json:"membership_tier_id" gorm:"index"`
	EffectiveAt      time.Time `json:"effective_at"` // When membership becomes active
	ExpiresAt        *time.Time `json:"expires_at"`  // NULL = lifetime
	IsActive         bool      `json:"is_active" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
