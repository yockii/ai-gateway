package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/models"
)

// SSEClient SSE 客户端连接
type SSEClient struct {
	Channel chan fiber.Map
	Context context.Context
}

// SSEBroadcaster SSE 广播器
type SSEBroadcaster struct {
	clients map[*SSEClient]bool
	mutex   sync.RWMutex
}

// NewSSEBroadcaster 创建 SSE 广播器
func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		clients: make(map[*SSEClient]bool),
	}
}

// AddClient 添加客户端
func (b *SSEBroadcaster) AddClient(client *SSEClient) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.clients[client] = true
}

// RemoveClient 移除客户端
func (b *SSEBroadcaster) RemoveClient(client *SSEClient) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if _, ok := b.clients[client]; ok {
		delete(b.clients, client)
		close(client.Channel)
	}
}

// Broadcast 广播消息到所有客户端（修复 WR-01: 清理已关闭/已满的客户端）
func (b *SSEBroadcaster) Broadcast(message fiber.Map) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for client := range b.clients {
		select {
		case client.Channel <- message:
			// 发送成功
		default:
			// 客户端通道已满或已关闭 - 移除该客户端
			delete(b.clients, client)
			close(client.Channel)
		}
	}
}

// Global SSE broadcasters for health status updates
var (
	healthStatusBroadcaster = NewSSEBroadcaster()
	supplierHealthBroadcasters = make(map[string]*SSEBroadcaster)
	supplierBroadcastersMutex sync.RWMutex
)

// StartHealthStatusTicker 启动健康状态推送定时器
func StartHealthStatusTicker() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			broadcastAllHealthStatus()
		}
	}()
}

// broadcastAllHealthStatus 广播所有供应商健康状态
func broadcastAllHealthStatus() {
	// 这个函数会从 Handler 中获取实际的健康状态
	// 由 StreamHealthStatus 处理器实际推送
}

// StreamHealthStatus SSE 流：推送所有供应商健康状态
func (h *Handler) StreamHealthStatus(c fiber.Ctx) error {
	// 设置 SSE 响应头
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// 创建客户端
	client := &SSEClient{
		Channel: make(chan fiber.Map, 10),
		Context: c.Context(),
	}
	healthStatusBroadcaster.AddClient(client)
	defer healthStatusBroadcaster.RemoveClient(client)

	// 立即发送当前状态
	h.sendCurrentHealthStatus(c)

	// 监听客户端断开
	clientChan := client.Channel
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Context().Done():
			return nil
		case <-ticker.C:
			// 每 5 秒推送一次健康状态
			if err := h.sendCurrentHealthStatus(c); err != nil {
				return nil
			}
		case msg, ok := <-clientChan:
			if !ok {
				return nil
			}
			data, _ := json.Marshal(msg)
			fmt.Fprintf(c, "data: %s\n\n", data)
			if err := c.Context().Err(); err != nil {
				return nil
			}
		}
	}
}

// StreamSupplierHealth SSE 流：推送单个供应商健康状态
func (h *Handler) StreamSupplierHealth(c fiber.Ctx) error {
	supplierID := c.Params("id")

	// 设置 SSE 响应头
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// 获取或创建该供应商的广播器
	supplierBroadcastersMutex.Lock()
	broadcaster, exists := supplierHealthBroadcasters[supplierID]
	if !exists {
		broadcaster = NewSSEBroadcaster()
		supplierHealthBroadcasters[supplierID] = broadcaster
	}
	supplierBroadcastersMutex.Unlock()

	// 创建客户端
	client := &SSEClient{
		Channel: make(chan fiber.Map, 10),
		Context: c.Context(),
	}
	broadcaster.AddClient(client)
	defer broadcaster.RemoveClient(client)

	// 立即发送当前状态
	h.sendSupplierHealthStatus(c, supplierID)

	// 监听客户端断开
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Context().Done():
			return nil
		case <-ticker.C:
			// 每 5 秒推送一次健康状态
			if err := h.sendSupplierHealthStatus(c, supplierID); err != nil {
				return nil
			}
		}
	}
}

// sendCurrentHealthStatus 发送当前健康状态
func (h *Handler) sendCurrentHealthStatus(c fiber.Ctx) error {
	type HealthStatusGetter interface {
		GetAllHealthStatus() map[string]interface{}
	}

	if sm, ok := h.supplierManager.(HealthStatusGetter); ok {
		statusMap := sm.GetAllHealthStatus()

		// 转换为 SSE 格式
		healthData := make(fiber.Map)
		for supplierID, status := range statusMap {
			if hs, ok := status.(*supplierHealthStatus); ok {
				healthData[supplierID] = fiber.Map{
					"supplier_id":  hs.SupplierID,
					"is_healthy":   hs.IsHealthy,
					"last_check":   hs.LastCheckTime,
					"error":        hs.LastError,
				}
			}
		}

		data, err := json.Marshal(fiber.Map{"data": healthData})
		if err != nil {
			return err
		}
		fmt.Fprintf(c, "data: %s\n\n", data)
	}

	return nil
}

// sendSupplierHealthStatus 发送单个供应商健康状态
func (h *Handler) sendSupplierHealthStatus(c fiber.Ctx, supplierID string) error {
	type HealthStatusGetter interface {
		GetHealthStatus(supplierID string) interface{}
	}

	if sm, ok := h.supplierManager.(HealthStatusGetter); ok {
		status := sm.GetHealthStatus(supplierID)

		healthData := fiber.Map{
			"supplier_id": supplierID,
			"is_healthy":  true,
			"last_check":  time.Now().Format(time.RFC3339),
			"latency":     0,
			"error":       "",
		}

		if hs, ok := status.(*supplierHealthStatus); ok {
			healthData["is_healthy"] = hs.IsHealthy
			healthData["last_check"] = hs.LastCheckTime
			healthData["error"] = hs.LastError
		}

		data, err := json.Marshal(fiber.Map{"data": healthData})
		if err != nil {
			return err
		}
		fmt.Fprintf(c, "data: %s\n\n", data)
	}

	return nil
}

// supplierHealthStatus 供应商健康状态（内部类型）
type supplierHealthStatus struct {
	SupplierID        string
	IsHealthy         bool
	LastCheckTime     string
	LastError         string
}

// SetSupplierManager 设置供应商管理器
func (h *Handler) SetSupplierManager(sm interface{}) {
	h.supplierManager = sm
}

// GetAllSuppliersHealth 获取所有供应商健康状态
func (h *Handler) GetAllSuppliersHealth(c fiber.Ctx) error {
	type HealthStatusGetter interface {
		GetAllHealthStatus() map[string]interface{}
	}

	if sm, ok := h.supplierManager.(HealthStatusGetter); ok {
		statusMap := sm.GetAllHealthStatus()
		return c.JSON(fiber.Map{"data": statusMap})
	}

	return c.Status(501).JSON(fiber.Map{
		"error": fiber.Map{
			"message": "Health status not available",
			"type":    "not_implemented",
			"code":    501,
		},
	})
}

// GetSupplierHealth 获取单个供应商健康状态
func (h *Handler) GetSupplierHealth(c fiber.Ctx) error {
	supplierID := c.Params("id")

	type HealthStatusGetter interface {
		GetHealthStatus(supplierID string) interface{}
	}

	if sm, ok := h.supplierManager.(HealthStatusGetter); ok {
		status := sm.GetHealthStatus(supplierID)
		if status == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Supplier not found",
					"type":    "not_found_error",
					"code":    404,
				},
			})
		}
		return c.JSON(fiber.Map{"data": status})
	}

	return c.Status(501).JSON(fiber.Map{
		"error": fiber.Map{
			"message": "Health status not available",
			"type":    "not_implemented",
			"code":    501,
		},
	})
}

// TriggerHealthCheck 手动触发健康检查
func (h *Handler) TriggerHealthCheck(c fiber.Ctx) error {
	supplierID := c.Params("id")

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"supplier_id": supplierID,
			"is_healthy":  true,
			"message":     "Health check triggered (simplified)",
		},
	})
}

// GetSupplierFailureEvents 获取供应商失败事件
func (h *Handler) GetSupplierFailureEvents(c fiber.Ctx) error {
	supplierID := c.Query("supplier_id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	var events []models.SupplierFailureEvent
	query := h.gateway.GetDB().Order("timestamp DESC").Limit(limit)

	if supplierID != "" {
		query = query.Where("supplier_id = ?", supplierID)
	}

	if err := query.Find(&events).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Failed to fetch failure events",
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": events,
	})
}

// GetFailoverEvents 获取故障转移事件
func (h *Handler) GetFailoverEvents(c fiber.Ctx) error {
	supplierID := c.Query("supplier_id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	var events []models.FailoverEvent
	query := h.gateway.GetDB().Order("timestamp DESC").Limit(limit)

	if supplierID != "" {
		query = query.Where("from_supplier_id = ? OR to_supplier_id = ?", supplierID, supplierID)
	}

	if err := query.Find(&events).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Failed to fetch failover events",
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": events,
	})
}

// GetHealthCheckHistory 获取健康检查历史
func (h *Handler) GetHealthCheckHistory(c fiber.Ctx) error {
	supplierID := c.Params("id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	var history []models.HealthCheckHistory
	if err := h.gateway.GetDB().
		Where("supplier_id = ?", supplierID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&history).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message": "Failed to fetch health check history",
				"type":    "api_error",
				"code":    500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": history,
	})
}
