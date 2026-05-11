package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/gateway"
	"github.com/yockii/ai-gateway/internal/middleware"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/pkg/api"
)

// Handler API 处理器
type Handler struct {
	enterpriseService *services.EnterprisePricingService
	pricingService    *services.PricingService
	gateway           *gateway.Gateway
	adminService      *services.AdminService
	userService       *services.UserService
	supplierApiKeyService *services.SupplierApiKeyService
	supplierModelService  *services.SupplierModelService
	keyManager            *services.KeyManager
	supplierManager       interface{}
	auditService          *services.AuditService
}

// New 创建处理器
func New(gw *gateway.Gateway) *Handler {
	return &Handler{
		gateway: gw,
	}
}

// SetAdminService 设置管理员服务
func (h *Handler) SetAdminService(adminService *services.AdminService) {
	h.adminService = adminService
}

// SetUserService 设置用户服务
func (h *Handler) SetUserService(userService *services.UserService) {
	h.userService = userService
}

// SetEnterpriseService 设置企业定价服务
func (h *Handler) SetEnterpriseService(es *services.EnterprisePricingService) {
	h.enterpriseService = es
}

// SetPricingService 设置统一定价服务
func (h *Handler) SetPricingService(ps *services.PricingService) {
	h.pricingService = ps
}

// SetGateway 设置网关（用于依赖注入）
func (h *Handler) SetGateway(gw *gateway.Gateway) {
	h.gateway = gw
}
// ChatCompletions 聊天完成接口
func (h *Handler) ChatCompletions(c fiber.Ctx) error {
	// 获取用户 ID
	userID := middleware.GetUserID(c)

	// 解析请求
	var req api.ChatCompletionRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Invalid request body",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	// 验证请求
	if err := h.validateChatRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 调用 Gateway
	resp, err := h.gateway.ChatCompletion(ctx, userID, &req)
	if err != nil {
		log.Printf("Chat completion 错误: 用户=%s 模型=%s 错误=%v", userID, req.Model, err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Failed to process request",
				Type:    "api_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	return c.JSON(resp)
}

// Completions 文本完成接口
func (h *Handler) Completions(c fiber.Ctx) error {
	// TODO: 实现 Completions 接口
	return c.Status(fiber.StatusNotImplemented).JSON(api.ErrorResponse{
		Error: api.ErrorDetail{
			Message: "Completions endpoint not yet implemented",
			Type:    "not_implemented",
			Code:    fiber.StatusNotImplemented,
		},
	})
}

// ListModels 列出可用模型（优化版，只查询必要字段）
func (h *Handler) ListModels(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 从数据库查询可用模型（使用优化的查询）
	models, err := h.gateway.GetModels(ctx)
	if err != nil {
		log.Printf("获取模型列表失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Failed to fetch models",
				Type:    "api_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	// 转换为 API 响应格式
	data := make([]api.ModelInfo, len(models))
	for i, model := range models {
		data[i] = api.ModelInfo{
			ID:      model.Name,
			Object:  "model",
			Created: model.CreatedAt.Unix(),
			OwnedBy: "ai-gateway",
		}
	}

	return c.JSON(&api.ModelsResponse{
		Object: "list",
		Data:   data,
	})
}

// GetUsage 获取使用记录
func (h *Handler) GetUsage(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// 获取查询参数
	days := 30 // 默认 30 天
	if daysParam := c.Query("days", ""); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 {
			days = d
		}
	}

	// TODO: 调用定价管理器获取使用统计
	return c.JSON(fiber.Map{
		"user_id": userID,
		"days":    days,
		"message": "Usage statistics not yet implemented",
	})
}

// GetUsageStats 获取仪表板统计数据
func (h *Handler) GetUsageStats(c fiber.Ctx) error {
	_ = c.Locals("user_id").(string) // TODO: 使用 userID 获取用户特定数据

	// TODO: 从数据库获取实际统计数据
	// 返回模拟数据
	return c.JSON(fiber.Map{
		"total_keys":       3,
		"monthly_requests": 12500,
		"monthly_cost":     158.50,
		"monthly_tokens":   2500000,
	})
}

// GenerateBill 生成账单
func (h *Handler) GenerateBill(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// 获取查询参数
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// TODO: 解析日期并调用定价管理器生成账单
	return c.JSON(fiber.Map{
		"user_id":     userID,
		"start_date":  startDate,
		"end_date":    endDate,
		"message":     "Bill generation not yet implemented",
	})
}

// validateChatRequest 验证聊天请求
func (h *Handler) validateChatRequest(req *api.ChatCompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if len(req.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	// 验证消息格式
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message %d: content is required", i)
		}

		// 验证角色类型
		validRoles := map[string]bool{
			"system":    true,
			"user":      true,
			"assistant": true,
		}
		if !validRoles[msg.Role] {
			return fmt.Errorf("message %d: invalid role '%s'", i, msg.Role)
		}
	}

	// 验证温度参数
	if req.Temperature != nil {
		temp := *req.Temperature
		if temp < 0 || temp > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
	}

	// 验证最大 tokens
	if req.MaxTokens != nil {
		maxTokens := *req.MaxTokens
		if maxTokens < 1 || maxTokens > 128000 {
			return fmt.Errorf("max_tokens must be between 1 and 128000")
		}
	}

	return nil
}
