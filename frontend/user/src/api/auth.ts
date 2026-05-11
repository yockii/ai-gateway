import client from './client'
import type { LoginRequest, RegisterRequest, LoginResponse } from '@/types/api'
import type { User } from '@/types/models'

export const authApi = {
  login: (data: LoginRequest) => client.post<LoginResponse>('/public/user/login', data),
  register: (data: RegisterRequest) => client.post<LoginResponse>('/public/user/register', data),
  logout: () => client.post('/user/logout'),
  profile: () => client.get<User>('/user/me'),
}
