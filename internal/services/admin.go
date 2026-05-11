package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/auth"
	"github.com/yockii/ai-gateway/internal/cache"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/logging"
	"go.uber.org/zap"
)

// AdminService 管理员服务
type AdminService struct {
	db           *database.DB
	tokenManager *auth.TokenManager
}

// NewAdminService 创建管理员服务
func NewAdminService(db *database.DB, redis *cache.RedisClient) *AdminService {
	return &AdminService{
		db:           db,
		tokenManager: auth.NewTokenManager(redis),
	}
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	Token     string    `json:"token"`
	AdminID   string    `json:"admin_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login 管理员登录
func (s *AdminService) Login(ctx context.Context, req *AdminLoginRequest) (*AdminLoginResponse, error) {
	var admin models.Admin
	err := s.db.WithContext(ctx).Where("email = ? AND is_active = ?", req.Email, true).First(&admin).Error
	if err != nil {
		logging.Warn("管理员登录失败: 邮箱不存在或未激活", zap.String("email", req.Email))
		return nil, errors.New("invalid credentials")
	}

	// 验证密码
	if !auth.CheckPassword(req.Password, admin.Password) {
		logging.Warn("管理员登录失败: 密码错误", zap.String("email", req.Email))
		return nil, errors.New("invalid credentials")
	}

	// 生成 xid token (24小时有效)
	token, err := s.tokenManager.GenerateToken(ctx, admin.ID, admin.Email, "admin", 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logging.Info("管理员登录成功", zap.String("admin_id", admin.ID), zap.String("email", admin.Email))

	return &AdminLoginResponse{
		Token:     token,
		AdminID:   admin.ID,
		Email:     admin.Email,
		Name:      admin.Name,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// CreateAdmin 创建管理员（仅限已登录管理员操作）
func (s *AdminService) CreateAdmin(ctx context.Context, req *CreateAdminRequest) (*models.Admin, error) {
	// 检查邮箱是否已存在
	var existingAdmin models.Admin
	err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingAdmin).Error
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// 哈希密码
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &models.Admin{
		ID:        xid.New().String(),
		Email:     req.Email,
		Password:  hashedPassword,
		Name:      req.Name,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(admin).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	logging.Info("创建管理员成功", zap.String("admin_id", admin.ID), zap.String("email", admin.Email))

	return admin, nil
}

// GetByID 根据 ID 获取管理员
func (s *AdminService) GetByID(ctx context.Context, id string) (*models.Admin, error) {
	var admin models.Admin
	err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// ValidateToken 验证 token
func (s *AdminService) ValidateToken(ctx context.Context, tokenString string) (*auth.TokenInfo, error) {
	return s.tokenManager.ValidateToken(ctx, tokenString)
}
