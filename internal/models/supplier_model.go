package models

import (
	"time"
)

// SupplierModel 供应商模型关联
type SupplierModel struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	SupplierID    string    `json:"supplier_id" gorm:"index"`
	ModelID       string    `json:"model_id" gorm:"index"`
	IsActive      bool      `json:"is_active" gorm:"index"`
	
	// 成本价格
	InputCost     float64   `json:"input_cost"`
	OutputCost    float64   `json:"output_cost"`
	
	// 变更追踪
	EffectiveDate time.Time `json:"effective_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SupplierModel) TableName() string {
	return "supplier_models"
}

// PriceHistory 价格变更历史
type PriceHistory struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	SupplierModelID string   `json:"supplier_model_id" gorm:"index"`
	SupplierID    string    `json:"supplier_id" gorm:"index"`
	ModelID       string    `json:"model_id" gorm:"index"`
	OldInputCost  float64   `json:"old_input_cost"`
	OldOutputCost float64   `json:"old_output_cost"`
	NewInputCost  float64   `json:"new_input_cost"`
	NewOutputCost float64   `json:"new_output_cost"`
	ChangedAt     time.Time `json:"changed_at"`
	ChangedBy     string    `json:"changed_by"` // 管理员ID
}

// TableName 指定表名
func (PriceHistory) TableName() string {
	return "price_histories"
}
