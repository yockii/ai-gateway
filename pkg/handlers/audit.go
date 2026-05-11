package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
)

// SetAuditService 设置审计日志服务
func (h *Handler) SetAuditService(svc *services.AuditService) {
	h.auditService = svc
}

// ListAuditLogs 获取审计日志列表
func (h *Handler) ListAuditLogs(c fiber.Ctx) error {
	var filter models.AuditLogFilter

	// 解析查询参数
	filter.EntityType = c.Query("entity_type", "")
	filter.EntityID = c.Query("entity_id", "")
	filter.Action = c.Query("action", "")
	filter.AdminID = c.Query("admin_id", "")
	filter.Keyword = c.Query("keyword", "")

	// 解析时间范围
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = t
		}
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	filter.Page = page
	filter.PageSize = pageSize

	logs, total, err := h.auditService.ListLogs(c.Context(), filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data": logs,
		"pagination": fiber.Map{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetAuditLogDetail 获取审计日志详情
func (h *Handler) GetAuditLogDetail(c fiber.Ctx) error {
	id := c.Params("id")

	log, err := h.auditService.GetLogByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Audit log not found"})
	}

	return c.JSON(fiber.Map{"data": log})
}

// GetEntityAuditHistory 获取实体审计历史
func (h *Handler) GetEntityAuditHistory(c fiber.Ctx) error {
	entityType := c.Query("entity_type")
	entityID := c.Query("entity_id")

	if entityType == "" || entityID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "entity_type and entity_id are required"})
	}

	history, err := h.auditService.GetEntityHistory(c.Context(), entityType, entityID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": history})
}

// Helper functions for recording audit logs in handlers

// getAdminInfo 从 context 获取管理员信息
func getAdminInfo(c fiber.Ctx) (adminID, adminName string) {
	adminID, _ = c.Locals("admin_id").(string)
	adminName, _ = c.Locals("admin_name").(string)
	return
}

// getClientInfo 获取客户端信息
func getClientInfo(c fiber.Ctx) (ipAddress, userAgent string) {
	ipAddress = c.IP()
	userAgent = c.Get("User-Agent")
	return
}

// recordAudit 记录审计日志的辅助函数
func recordAudit(handler *Handler, c fiber.Ctx, entityType, entityID, action string, changes models.ChangeLog) {
	if handler.auditService == nil {
		return
	}

	adminID, adminName := getAdminInfo(c)
	ipAddress, userAgent := getClientInfo(c)

	handler.auditService.LogActionAsync(c.Context(), adminID, adminName, entityType, entityID, action, changes, ipAddress, userAgent)
}
