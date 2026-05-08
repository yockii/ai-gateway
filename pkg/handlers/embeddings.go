package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/middleware"
	"github.com/yockii/ai-gateway/pkg/api"
)

// CreateEmbedding 创建嵌入向量 (per D-03)
func (h *Handler) CreateEmbedding(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req api.EmbeddingRequest
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
	if err := h.validateEmbeddingRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 调用 Gateway
	resp, err := h.gateway.CreateEmbedding(ctx, userID, &req)
	if err != nil {
		log.Printf("嵌入生成错误: 用户=%s 模型=%s 错误=%v", userID, req.Model, err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Failed to create embedding",
				Type:    "api_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	return c.JSON(resp)
}

// validateEmbeddingRequest 验证嵌入请求
func (h *Handler) validateEmbeddingRequest(req *api.EmbeddingRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	// 验证输入
	if req.Input == nil || len(req.Input) == 0 {
		return fmt.Errorf("input is required")
	}

	// 验证输入数量
	if len(req.Input) > 2048 {
		return fmt.Errorf("too many inputs (max 2048)")
	}

	// 验证编码格式
	if req.EncodingFormat != "" && req.EncodingFormat != "float" && req.EncodingFormat != "base64" {
		return fmt.Errorf("invalid encoding_format: %s", req.EncodingFormat)
	}

	return nil
}
