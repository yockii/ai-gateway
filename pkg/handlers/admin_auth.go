package handlers

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/pkg/api"
)

// AdminLogin 管理员登录
func (h *Handler) AdminLogin(c fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	// 验证请求
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Email and password are required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	ctx := context.Background()
	loginReq := &services.AdminLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.adminService.Login(ctx, loginReq)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Invalid email or password",
				Type:    "authentication_error",
				Code:    fiber.StatusUnauthorized,
			},
		})
	}

	return c.JSON(resp)
}

// CreateAdmin 创建管理员（仅限已登录管理员）
func (h *Handler) CreateAdmin(c fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
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

	// 验证请求
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Email, password and name are required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	// 密码强度检查
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Password must be at least 8 characters",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	ctx := context.Background()
	createReq := &services.CreateAdminRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	admin, err := h.adminService.CreateAdmin(ctx, createReq)
	if err != nil {
		if err.Error() == "email already exists" {
			return c.Status(fiber.StatusConflict).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Email already exists",
					Type:    "duplicate_error",
					Code:    fiber.StatusConflict,
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Failed to create admin",
				Type:    "server_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         admin.ID,
		"email":      admin.Email,
		"name":       admin.Name,
		"is_active":  admin.IsActive,
		"created_at": admin.CreatedAt,
	})
}

// GetMe 获取当前管理员信息
func (h *Handler) GetMe(c fiber.Ctx) error {
	adminID := c.Locals("admin_id").(string)
	email := c.Locals("admin_email").(string)

	return c.JSON(fiber.Map{
		"id":    adminID,
		"email": email,
		"type":  "admin",
	})
}

// ListUsers 获取用户列表（管理员）
func (h *Handler) ListUsers(c fiber.Ctx) error {
	search := c.Query("search", "")

	db := h.gateway.GetDB()
	var users []models.User
	var err error
	if search != "" {
		err = db.Where("email LIKE ? OR name LIKE ?", "%"+search+"%", "%"+search+"%").Find(&users).Error
	} else {
		err = db.Find(&users).Error
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	return c.JSON(fiber.Map{
		"object": "list",
		"data":   users,
	})
}
