import client, { publicClient } from './client'
import type { AdminUser } from '@/types/models'

export interface AdminLoginRequest {
  email: string
  password: string
}

export interface AdminLoginResponse {
  admin: AdminUser
  token: string
}

export const authApi = {
  // 登录使用公开 API（不需要 JWT）
  login: (data: AdminLoginRequest) => publicClient.post<AdminLoginResponse>('/admin/login', data),
  // 登出和管理员信息使用需要认证的 API
  logout: () => client.post('/logout'),
  profile: () => client.get<AdminUser>('/me'),
}
