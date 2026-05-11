package handlers

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/pkg/api"
)

// UserLogin 用户登录
func (h *Handler) UserLogin(c fiber.Ctx) error {
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
	loginReq := &services.UserLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.userService.Login(ctx, loginReq)
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

// UserRegister 用户注册
func (h *Handler) UserRegister(c fiber.Ctx) error {
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
	registerReq := &services.UserRegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	user, err := h.userService.Register(ctx, registerReq)
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
				Message: "Failed to create user",
				Type:    "server_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"api_key":    user.APIKey,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
	})
}

// GetUserMe 获取当前用户信息
func (h *Handler) GetUserMe(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	email := c.Locals("user_email").(string)

	return c.JSON(fiber.Map{
		"id":    userID,
		"email": email,
		"type":  "user",
	})
}
