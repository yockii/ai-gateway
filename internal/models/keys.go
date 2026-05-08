package models

import "time"

// UserAPIKey 用户 API Key (per D-11, D-12, D-13)
type UserAPIKey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index"`
	KeyValue    string    `json:"-" gorm:"uniqueIndex"` // Hidden from JSON, unique in DB
	Name        string    `json:"name"`                  // User-defined name for the key

	// Quota limits
	QuotaDaily   int64 `json:"quota_daily"`   // Daily token quota
	QuotaMonthly int64 `json:"quota_monthly"` // Monthly token quota

	// Concurrency limits (per D-12, D-13)
	ConcurrencyLimit  int64            `json:"concurrency_limit"`  // Default concurrent requests
	ModelConcurrency  map[string]int64 `json:"model_concurrency"`  // Model-specific: {"gpt-4": 5, "gpt-3.5-turbo": 10}
	LastCalculatedAt *time.Time       `json:"last_calculated_at"` // For concurrency tracking

	// Time constraints
	ExpiresAt time.Time  `json:"expires_at"`
	IsActive  bool       `json:"is_active" gorm:"index"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
