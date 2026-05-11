package auth

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Type     string `json:"type"` // "admin" or "user"
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secretKey     string
	adminDuration time.Duration
	userDuration  time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		adminDuration: 24 * time.Hour, // 管理员 token 24小时
		userDuration:  30 * 24 * time.Hour, // 用户 token 30天
	}
}

// GenerateAdminToken 生成管理员 JWT token
func (j *JWTManager) GenerateAdminToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Type:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.adminDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// GenerateUserToken 生成用户 JWT token
func (j *JWTManager) GenerateUserToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Type:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.userDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken 验证 JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetAdminFromContext 从 Fiber 上下文获取管理员信息
func GetAdminFromContext(c fiber.Ctx) (*Claims, error) {
	claims, ok := c.Locals("admin").(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("admin not found in context")
	}
	return claims, nil
}
