package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/cache"
)

// TokenInfo 存储在 Redis 中的 token 信息
type TokenInfo struct {
	UserID string
	Email  string
	Type   string // "admin" 或 "user"
}

// TokenManager 基于 Redis 的 token 管理
type TokenManager struct {
	redis *cache.RedisClient
}

// NewTokenManager 创建 token 管理器
func NewTokenManager(redis *cache.RedisClient) *TokenManager {
	return &TokenManager{redis: redis}
}

// GenerateToken 生成新 token
func (m *TokenManager) GenerateToken(ctx context.Context, userID, email, tokenType string, expiration time.Duration) (string, error) {
	token := xid.New().String()

	// 存储到 Redis
	key := m.tokenKey(token)
	info := fmt.Sprintf("%s|%s|%s", userID, email, tokenType)

	if err := m.redis.Set(ctx, key, info, expiration); err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	return token, nil
}

// ValidateToken 验证 token 并返回用户信息
func (m *TokenManager) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	key := m.tokenKey(token)
	info, err := m.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// 解析 info: userID|email|type (使用 strings.Split)
	parts := strings.Split(info, "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	return &TokenInfo{
		UserID: parts[0],
		Email:  parts[1],
		Type:   parts[2],
	}, nil
}

// RevokeToken 撤销 token
func (m *TokenManager) RevokeToken(ctx context.Context, token string) error {
	key := m.tokenKey(token)
	return m.redis.Delete(ctx, key)
}

// tokenKey 生成 Redis key
func (m *TokenManager) tokenKey(token string) string {
	return "auth:token:" + token
}
