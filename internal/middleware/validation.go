package middleware

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/pkg/api"
)

// Validator 请求验证器
type Validator struct {
	db *database.DB
}

// NewValidator 创建验证器
func NewValidator(db *database.DB) *Validator {
	return &Validator{db: db}
}

// ValidateModel 验证模型是否存在且可用
func (v *Validator) ValidateModel() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 对于不需要模型验证的端点跳过
		if c.Path() == "/v1/models" || c.Path() == "/v1/bills" {
			return c.Next()
		}

		// 从请求体或查询参数获取模型
		modelID := c.Query("model", "")
		if modelID == "" {
			// 尝试从请求体解析
			var req struct {
				Model string `json:"model"`
			}
			if err := c.Bind().Body(&req); err == nil {
				modelID = req.Model
			}
		}

		if modelID != "" {
			// 验证模型存在且激活
			var model models.ExternalModel
			err := v.db.Where("name = ? AND is_active = ?", modelID, true).
				First(&model).Error

			if err != nil {
				log.Printf("模型验证失败: 模型=%s 错误=%v", modelID, err)
				return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
					Error: api.ErrorDetail{
						Message: fmt.Sprintf("Model '%s' not found or unavailable", modelID),
						Type:    "invalid_request_error",
						Code:    fiber.StatusBadRequest,
					},
				})
			}
		}

		return c.Next()
	}
}

// ValidateChatRequest 验证聊天请求格式
func (v *Validator) ValidateChatRequest() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 只对聊天端点验证
		if c.Path() != "/v1/chat/completions" {
			return c.Next()
		}

		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Stream bool `json:"stream"`
		}

		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Invalid request body",
					Type:    "invalid_request_error",
					Code:    fiber.StatusBadRequest,
				},
			})
		}

		// 验证必填字段
		if req.Model == "" {
			return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "model is required",
					Type:    "invalid_request_error",
					Code:    fiber.StatusBadRequest,
				},
			})
		}

		if len(req.Messages) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "messages cannot be empty",
					Type:    "invalid_request_error",
					Code:    fiber.StatusBadRequest,
				},
			})
		}

		// 验证消息格式
		validRoles := map[string]bool{
			"system":    true,
			"user":      true,
			"assistant": true,
		}

		for i, msg := range req.Messages {
			if msg.Role == "" || !validRoles[msg.Role] {
				return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
					Error: api.ErrorDetail{
						Message: fmt.Sprintf("message %d: invalid role", i),
						Type:    "invalid_request_error",
						Code:    fiber.StatusBadRequest,
					},
				})
			}
			if msg.Content == "" {
				return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
					Error: api.ErrorDetail{
						Message: fmt.Sprintf("message %d: content is required", i),
						Type:    "invalid_request_error",
						Code:    fiber.StatusBadRequest,
					},
				})
			}
		}

		// 存储解析后的请求供后续使用
		c.Locals("validated_request", req)

		return c.Next()
	}
}

// 全局验证器实例
var defaultValidator *Validator

// ValidateModel 使用默认验证器
func ValidateModel() fiber.Handler {
	if defaultValidator == nil {
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}
	return defaultValidator.ValidateModel()
}

// ValidateChatRequest 使用默认验证器
func ValidateChatRequest() fiber.Handler {
	if defaultValidator == nil {
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}
	return defaultValidator.ValidateChatRequest()
}
