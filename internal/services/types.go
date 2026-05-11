package services

import "time"

// CommonLoginRequest 通用登录请求
type CommonLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CommonLoginResponse 通用登录响应
type CommonLoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
