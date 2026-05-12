package models

import "time"

// SupplierFailureEvent 供应商失败事件
type SupplierFailureEvent struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SupplierID string    `json:"supplier_id" gorm:"index"`
	ModelID    string    `json:"model_id" gorm:"index"`
	ErrorType  string    `json:"error_type"`  // timeout, connection_error, rate_limit, api_error
	ErrorMsg   string    `json:"error_msg"`
	StatusCode int       `json:"status_code"`
	Timestamp  time.Time `json:"timestamp" gorm:"index"`
}

// FailoverEvent 故障转移事件
type FailoverEvent struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	FromSupplierID string    `json:"from_supplier_id" gorm:"index"`
	ToSupplierID   string    `json:"to_supplier_id" gorm:"index"`
	ModelID        string    `json:"model_id" gorm:"index"`
	Reason         string    `json:"reason"` // unhealthy, timeout, error
	Timestamp      time.Time `json:"timestamp" gorm:"index"`
}

// HealthCheckHistory 健康检查历史
type HealthCheckHistory struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SupplierID string    `json:"supplier_id" gorm:"index"`
	IsHealthy  bool      `json:"is_healthy"`
	Latency    int64     `json:"latency"` // 毫秒
	Error      string    `json:"error"`
	Timestamp  time.Time `json:"timestamp" gorm:"index"`
}
