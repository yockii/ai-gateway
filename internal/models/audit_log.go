package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AuditLog 审计日志
type AuditLog struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	AdminID    string    `json:"admin_id" gorm:"index"`
	AdminName  string    `json:"admin_name"`
	EntityType string    `json:"entity_type" gorm:"index"` // supplier, pricing, api_key, model, user
	EntityID   string    `json:"entity_id" gorm:"index"`
	Action     string    `json:"action" gorm:"index"` // create, update, delete, rotate, set_primary
	Changes    ChangeLog `json:"changes" gorm:"type:jsonb"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Timestamp  time.Time `json:"timestamp" gorm:"index"`
}

// ChangeLog 变更日志
type ChangeLog struct {
	Before map[string]interface{} `json:"before,omitempty"`
	After  map[string]interface{} `json:"after,omitempty"`
}

// Scan 实现 sql.Scanner 接口
func (c *ChangeLog) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value 实现 driver.Valuer 接口
func (c ChangeLog) Value() (driver.Value, error) {
	if c.Before == nil && c.After == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// AuditLogFilter 审计日志过滤器
type AuditLogFilter struct {
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	AdminID    string    `json:"admin_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Keyword    string    `json:"keyword"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}
