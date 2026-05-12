package models

import (
	"time"
)

// SupplierApiKey 供应商 API Key
type SupplierApiKey struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	SupplierID        string    `json:"supplier_id" gorm:"index"`
	Name              string    `json:"name"`                        // 密钥名称
	KeyValueEncrypted string    `json:"-" gorm:"not null"`           // 加密后的密钥
	KeyPrefix         string    `json:"key_prefix"`                  // 用于显示
	
	// 优先级管理
	Priority          int       `json:"priority"`                    // 数字越小优先级越高
	IsPrimary         bool      `json:"is_primary" gorm:"index"`     // 是否为主密钥
	
	// 使用限制
	MaxRequests       int64     `json:"max_requests"`                // 最大请求数
	CurrentRequests   int64     `json:"current_requests"`            // 当前请求数
	
	// 状态管理
	IsActive          bool      `json:"is_active" gorm:"index"`
	LastUsedAt        *time.Time `json:"last_used_at"`
	ExpireAt          *time.Time `json:"expire_at"`
	
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SupplierApiKey) TableName() string {
	return "supplier_api_keys"
}

// IsExpired 检查是否过期
func (k *SupplierApiKey) IsExpired() bool {
	if k.ExpireAt == nil {
		return false
	}
	return time.Now().After(*k.ExpireAt)
}

// ShouldRotate 检查是否应该轮换
func (k *SupplierApiKey) ShouldRotate() bool {
	// 检查是否过期
	if k.IsExpired() {
		return true
	}
	// 检查请求数是否达到阈值
	if k.MaxRequests > 0 && k.CurrentRequests >= k.MaxRequests {
		return true
	}
	return false
}
