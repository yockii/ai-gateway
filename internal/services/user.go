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

// UserService 用户服务
type UserService struct {
	db           *database.DB
	tokenManager *auth.TokenManager
}

// NewUserService 创建用户服务
func NewUserService(db *database.DB, redis *cache.RedisClient) *UserService {
	return &UserService{
		db:           db,
		tokenManager: auth.NewTokenManager(redis),
	}
}

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *UserRegisterRequest) (*models.User, error) {
	// 检查邮箱是否已存在
	var existingUser models.User
	err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// 哈希密码
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 生成 API Key
	apiKey := "sk-" + xid.New().String()

	user := &models.User{
		ID:         xid.New().String(),
		Email:      req.Email,
		Password:   hashedPassword,
		Name:       req.Name,
		APIKey:     apiKey,
		UserGroupID: "default", // 默认用户组
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logging.Info("用户注册成功", zap.String("user_id", user.ID), zap.String("email", user.Email))

	return user, nil
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *UserLoginRequest) (*UserLoginResponse, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error
	if err != nil {
		logging.Warn("用户登录失败: 邮箱不存在或未激活", zap.String("email", req.Email))
		return nil, errors.New("invalid credentials")
	}

	// 验证密码
	if !auth.CheckPassword(req.Password, user.Password) {
		logging.Warn("用户登录失败: 密码错误", zap.String("email", req.Email))
		return nil, errors.New("invalid credentials")
	}

	// 生成 xid token (30天有效)
	token, err := s.tokenManager.GenerateToken(ctx, user.ID, user.Email, "user", 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logging.Info("用户登录成功", zap.String("user_id", user.ID), zap.String("email", user.Email))

	return &UserLoginResponse{
		Token:     token,
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		APIKey:    user.APIKey,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidateToken 验证 token
func (s *UserService) ValidateToken(ctx context.Context, tokenString string) (*auth.TokenInfo, error) {
	return s.tokenManager.ValidateToken(ctx, tokenString)
}
