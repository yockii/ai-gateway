package models

import "time"

// ModelMapping 对外模型到供应商模型的映射 (per D-10)
type ModelMapping struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ExternalModelID string   `json:"external_model_id" gorm:"index"` // Our external model name
	SupplierID     string    `json:"supplier_id" gorm:"index"`       // Which supplier
	ActualModelName string   `json:"actual_model_name"`              // Supplier's model identifier

	// Routing configuration
	Priority       int       `json:"priority"`        // Lower = higher priority
	IsBackup       bool      `json:"is_backup"`       // true = only used if primaries fail
	Weight         float64   `json:"weight"`          // Load balancing weight (0-100)
	MaxQPS         int       `json:"max_qps"`         // QPS limit for this route

	// Health check
	HealthCheckURL  string    `json:"health_check_url"`
	TimeoutSeconds  int       `json:"timeout_seconds"`
	RetryCount      int       `json:"retry_count"`

	// Status
	IsActive       bool      `json:"is_active" gorm:"index"`
	Status         string    `json:"status"`          // active, degraded, error

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
