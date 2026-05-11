package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/models"
	"gorm.io/gorm"
)

// AuditService 审计日志服务
type AuditService struct {
	db *gorm.DB
}

// NewAuditService 创建审计日志服务
func NewAuditService(db interface{}) (*AuditService, error) {
	// 支持 *gorm.DB 或 *database.DB
	var gormDB *gorm.DB
	if dbVal, ok := db.(*gorm.DB); ok {
		gormDB = dbVal
	} else if dbVal, ok := db.(interface{ DB() *gorm.DB }); ok {
		gormDB = dbVal.DB()
	} else {
		return nil, fmt.Errorf("unsupported db type: %T", db)
	}
	return &AuditService{db: gormDB}, nil
}

// LogAction 记录审计日志
func (s *AuditService) LogAction(ctx context.Context, adminID, adminName, entityType, entityID, action string, changes models.ChangeLog, ipAddress, userAgent string) error {
	log := &models.AuditLog{
		ID:         xid.New().String(),
		AdminID:    adminID,
		AdminName:  adminName,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Changes:    changes,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Timestamp:  time.Now(),
	}

	return s.db.WithContext(ctx).Create(log).Error
}

// LogActionAsync 异步记录审计日志（不阻塞主流程）
func (s *AuditService) LogActionAsync(ctx context.Context, adminID, adminName, entityType, entityID, action string, changes models.ChangeLog, ipAddress, userAgent string) {
	go func() {
		// 使用超时 context 防止 goroutine 泄漏
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.LogAction(timeoutCtx, adminID, adminName, entityType, entityID, action, changes, ipAddress, userAgent); err != nil {
			// 记录失败不应影响主流程，仅打印日志
			fmt.Printf("failed to log audit action: %v\n", err)
		}
	}()
}

// ListLogs 查询审计日志
func (s *AuditService) ListLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64

	query := s.db.WithContext(ctx).Model(&models.AuditLog{})

	// 应用过滤条件
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.EntityID != "" {
		query = query.Where("entity_id = ?", filter.EntityID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.AdminID != "" {
		query = query.Where("admin_id = ?", filter.AdminID)
	}
	if !filter.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndTime)
	}
	if filter.Keyword != "" {
		query = query.Where("entity_id LIKE ? OR admin_name LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// 分页
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// 查询数据（按时间倒序）
	if err := query.Order("timestamp DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return logs, total, nil
}

// GetEntityHistory 获取实体的完整历史记录
func (s *AuditService) GetEntityHistory(ctx context.Context, entityType, entityID string) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog

	query := s.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("timestamp ASC")

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get entity history: %w", err)
	}

	return logs, nil
}

// GetLogByID 根据 ID 获取审计日志详情
func (s *AuditService) GetLogByID(ctx context.Context, id string) (*models.AuditLog, error) {
	var log models.AuditLog
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&log).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	return &log, nil
}

// BuildChangeLog 构建变更日志（用于 update 操作）
func BuildChangeLog(before, after interface{}) (models.ChangeLog, error) {
	changeLog := models.ChangeLog{
		Before: make(map[string]interface{}),
		After:  make(map[string]interface{}),
	}

	if before != nil {
		beforeBytes, err := json.Marshal(before)
		if err != nil {
			return changeLog, fmt.Errorf("failed to marshal before: %w", err)
		}
		if err := json.Unmarshal(beforeBytes, &changeLog.Before); err != nil {
			return changeLog, fmt.Errorf("failed to unmarshal before: %w", err)
		}
	}

	if after != nil {
		afterBytes, err := json.Marshal(after)
		if err != nil {
			return changeLog, fmt.Errorf("failed to marshal after: %w", err)
		}
		if err := json.Unmarshal(afterBytes, &changeLog.After); err != nil {
			return changeLog, fmt.Errorf("failed to unmarshal after: %w", err)
		}
	}

	return changeLog, nil
}
