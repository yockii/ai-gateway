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

// CreateImage 图片生成 (per D-03: OpenAI 兼容)
func (h *Handler) CreateImage(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req api.ImageRequest
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
	if err := h.validateImageRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 调用 Gateway
	resp, err := h.gateway.ImageGeneration(ctx, userID, &req)
	if err != nil {
		log.Printf("图片生成错误: 用户=%s 模型=%s 错误=%v", userID, req.Model, err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Failed to generate image",
				Type:    "api_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	return c.JSON(resp)
}

// CreateImageEdit 图片编辑
func (h *Handler) CreateImageEdit(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// 获取上传的图片
	image, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Image file is required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	mask, _ := c.FormFile("mask") // 可选

	var req api.ImageEditRequest
	req.Model = c.FormValue("model")
	req.Prompt = c.FormValue("prompt")
	req.N = parseIntOrDefault(c.FormValue("n"), 1)
	req.Size = c.FormValue("size")
	req.ResponseFormat = c.FormValue("response_format")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// TODO: 处理图片上传和编辑
	_ = userID
	_ = image
	_ = mask
	_ = ctx

	return c.Status(fiber.StatusNotImplemented).JSON(api.ErrorResponse{
		Error: api.ErrorDetail{
			Message: "Image edit not yet implemented",
			Type:    "not_implemented",
			Code:    fiber.StatusNotImplemented,
		},
	})
}

// CreateImageVariation 图片变体
func (h *Handler) CreateImageVariation(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// 获取上传的图片
	image, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Image file is required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	var req api.ImageVariationRequest
	req.Model = c.FormValue("model")
	req.N = parseIntOrDefault(c.FormValue("n"), 1)
	req.Size = c.FormValue("size")
	req.ResponseFormat = c.FormValue("response_format")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// TODO: 处理图片变体
	_ = userID
	_ = image
	_ = ctx

	return c.Status(fiber.StatusNotImplemented).JSON(api.ErrorResponse{
		Error: api.ErrorDetail{
			Message: "Image variation not yet implemented",
			Type:    "not_implemented",
			Code:    fiber.StatusNotImplemented,
		},
	})
}

// validateImageRequest 验证图片请求
func (h *Handler) validateImageRequest(req *api.ImageRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	// 验证尺寸
	validSizes := map[string]bool{
		"256x256":   true,
		"512x512":   true,
		"1024x1024": true,
		"1792x1024": true,
		"1024x1792": true,
	}

	if req.Size != "" && !validSizes[req.Size] {
		return fmt.Errorf("invalid size: %s", req.Size)
	}

	// 验证 N 参数
	if req.N < 1 || req.N > 10 {
		return fmt.Errorf("n must be between 1 and 10")
	}

	return nil
}

func parseIntOrDefault(s string, defaultVal int) int {
	result := defaultVal
	if s != "" {
		fmt.Sscanf(s, "%d", &result)
	}
	return result
}
