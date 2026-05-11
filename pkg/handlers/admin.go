package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/pkg/api"
)

// CreateModel 创建对外模型 (管理员)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) CreateModel(c fiber.Ctx) error {
	var req api.CreateModelRequest
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
	if err := h.validateCreateModelRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	ctx := context.Background()

	// TODO: 调用服务层创建模型
	_ = ctx

	log.Printf("创建模型: 名称=%s 类型=%s", req.Name, req.ModelType)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         fmt.Sprintf("model-%d", time.Now().Unix()),
		"name":       req.Name,
		"display_name": req.DisplayName,
		"model_type": req.ModelType,
		"created_at": time.Now().UTC(),
	})
}

// UpdateModel 更新模型配置 (管理员)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) UpdateModel(c fiber.Ctx) error {
	modelID := c.Params("id")
	if modelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Model ID is required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	var req api.UpdateModelRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Invalid request body",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	ctx := context.Background()
	_ = ctx

	log.Printf("更新模型: ID=%s", modelID)

	return c.JSON(fiber.Map{
		"id":         modelID,
		"updated_at": time.Now().UTC(),
	})
}

// DeleteModel 删除模型 (管理员)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) DeleteModel(c fiber.Ctx) error {
	modelID := c.Params("id")
	if modelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Model ID is required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	log.Printf("删除模型: ID=%s", modelID)

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// AdminListModels 列出所有模型 (管理员视图)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) AdminListModels(c fiber.Ctx) error {
	// TODO: 查询数据库获取所有模型

	return c.JSON(fiber.Map{
		"object": "list",
		"data": []fiber.Map{
			{
				"id":          "model-001",
				"name":        "gpt-4",
				"display_name": "GPT-4",
				"model_type":  "chat",
				"is_active":   true,
				"created_at":  time.Now().UTC(),
			},
		},
	})
}

// CreateSupplier 创建供应商 (管理员)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) CreateSupplier(c fiber.Ctx) error {
	var req api.CreateSupplierRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Invalid request body",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	log.Printf("创建供应商: 名称=%s 提供商=%s", req.Name, req.Provider)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         fmt.Sprintf("supplier-%d", time.Now().Unix()),
		"name":       req.Name,
		"display_name": req.DisplayName,
		"provider":   req.Provider,
		"created_at": time.Now().UTC(),
	})
}

// UpdateSupplier 更新供应商 (管理员)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) UpdateSupplier(c fiber.Ctx) error {
	supplierID := c.Params("id")
	if supplierID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Supplier ID is required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	var req api.UpdateModelRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Invalid request body",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	log.Printf("更新供应商: ID=%s", supplierID)

	return c.JSON(fiber.Map{
		"id":         supplierID,
		"updated_at": time.Now().UTC(),
	})
}

// DeleteSupplier 删除供应商 (管理员)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) DeleteSupplier(c fiber.Ctx) error {
	supplierID := c.Params("id")
	if supplierID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Supplier ID is required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	log.Printf("删除供应商: ID=%s", supplierID)

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ListSuppliers 列出所有供应商 (管理员)
// 权限验证由 RequireAdmin 中间件处理
func (h *Handler) ListSuppliers(c fiber.Ctx) error {
	// TODO: 查询数据库获取所有供应商

	return c.JSON(fiber.Map{
		"object": "list",
		"data": []fiber.Map{
			{
				"id":          "supplier-001",
				"name":        "openai",
				"display_name": "OpenAI",
				"provider":    "openai",
				"is_active":   true,
				"created_at":  time.Now().UTC(),
			},
		},
	})
}

// validateCreateModelRequest 验证创建模型请求
func (h *Handler) validateCreateModelRequest(req *api.CreateModelRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	validTypes := map[string]bool{
		"chat":        true,
		"completion":  true,
		"image":       true,
		"video":       true,
		"tts":         true,
		"stt":         true,
		"embedding":   true,
		"rerank":      true,
	}

	if !validTypes[req.ModelType] {
		return fmt.Errorf("invalid model_type: %s", req.ModelType)
	}

	return nil
}
