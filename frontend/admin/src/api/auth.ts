import client from './client'
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
  login: (data: AdminLoginRequest) => client.post<AdminLoginResponse>('/login', data),
  logout: () => client.post('/logout'),
  profile: () => client.get<AdminUser>('/profile'),
}
