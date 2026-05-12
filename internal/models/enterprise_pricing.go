package models

import "time"

// EnterprisePricing 大客户独立定价 (Milestone 2 - Phase 5)
type EnterprisePricing struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	CustomerID       string    `json:"customer_id" gorm:"index"`
	CustomerName     string    `json:"customer_name"`
	ModelID          string    `json:"model_id" gorm:"index"`

	// 定价配置 (per 1M tokens)
	InputPrice       float64   `json:"input_price"`
	OutputPrice      float64   `json:"output_price"`

	// 利润保护
	MinProfitMargin  float64   `json:"min_profit_margin"` // 最低利润率 (0.1 = 10%)
	MaxCostPrice     float64   `json:"max_cost_price"`    // 最高可接受成本价

	// 生效时间
	EffectiveDate    time.Time `json:"effective_date"`
	ExpiryDate       *time.Time `json:"expiry_date"` // NULL = 无限期

	// 状态
	IsActive         bool      `json:"is_active" gorm:"index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedBy        string    `json:"created_by"` // 创建人 (管理员 ID)
	UpdatedBy        string    `json:"updated_by"` // 更新人 (管理员 ID)
}

// TableName 指定表名
func (EnterprisePricing) TableName() string {
	return "enterprise_pricing"
}

// IsActiveAt 检查定价在指定时间是否生效
func (e *EnterprisePricing) IsActiveAt(t time.Time) bool {
	if !e.IsActive {
		return false
	}
	if t.Before(e.EffectiveDate) {
		return false
	}
	if e.ExpiryDate != nil && t.After(*e.ExpiryDate) {
		return false
	}
	return true
}

// EnterprisePricingFilter 大客户定价筛选条件
type EnterprisePricingFilter struct {
	CustomerID   string
	ModelID      string
	IsActive     *bool
	EffectiveBefore *time.Time
	EffectiveAfter  *time.Time
	Limit        int
	Offset       int
}
