package models

import (
	"time"
)

// UsageRecord 使用记录
type UsageRecord struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"user_id" gorm:"index"`
	ModelID    string    `json:"model_id" gorm:"index"`
	SupplierID string    `json:"supplier_id" gorm:"index"`

	// Token 使用量
	InputTokens  int32 `json:"input_tokens"`
	OutputTokens int32 `json:"output_tokens"`
	TotalTokens  int32 `json:"total_tokens"`

	// 费用信息
	CostPrice   float64 `json:"cost_price"`
	SellingPrice float64 `json:"selling_price"`
	Profit      float64 `json:"profit"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// Supplier 供应商配置
type Supplier struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	DisplayName string    `json:"display_name"`
	Provider    string    `json:"provider"` // openai, anthropic, etc.
	IsActive    bool      `json:"is_active" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SupplierCostPricing 供应商成本定价
type SupplierCostPricing struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	SupplierID    string    `json:"supplier_id" gorm:"index"`
	ModelID       string    `json:"model_id" gorm:"index"`
	InputCost     float64   `json:"input_cost"`
	OutputCost    float64   `json:"output_cost"`
	EffectiveDate time.Time `json:"effective_date"`
	IsActive      bool      `json:"is_active" gorm:"index"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserGroup 用户群体定价
type UserGroup struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	DisplayName string    `json:"display_name"`
	IsActive    bool      `json:"is_active" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserGroupPricing 用户群体定价
type UserGroupPricing struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	UserGroupID      string    `json:"user_group_id" gorm:"index"`
	ModelID          string    `json:"model_id" gorm:"index"`
	InputPrice       float64   `json:"input_price"`
	OutputPrice      float64   `json:"output_price"`
	MinProfitMargin  float64   `json:"min_profit_margin"`
	EffectiveDate    time.Time `json:"effective_date"`
	IsActive         bool      `json:"is_active" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ExternalModel 对外模型配置
type ExternalModel struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	DisplayName string    `json:"display_name"`
	ModelType   string    `json:"model_type"` // chat, completion, image, video, tts, stt, embedding, rerank
	IsActive    bool      `json:"is_active" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User 用户
type User struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Email       string    `json:"email" gorm:"uniqueIndex"`
	Password    string    `json:"-"` // 不在 JSON 中显示
	Name        string    `json:"name"`
	UserGroupID string    `json:"user_group_id" gorm:"index"`
	APIKey      string    `json:"-" gorm:"uniqueIndex"` // API Key，不在 JSON 中显示
	IsActive    bool      `json:"is_active" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Admin 管理员
type Admin struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Password  string    `json:"-"` // 不在 JSON 中显示
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
