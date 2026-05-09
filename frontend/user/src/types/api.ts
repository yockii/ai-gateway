import type { User, UserAPIKey, UsageRecord } from './models'

export interface ApiResponse<T> {
  data: T
  message?: string
}

export interface ApiError {
  message: string
  code?: string
  details?: unknown
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface LoginResponse {
  user: User
  token: string
}

export interface CreateKeyRequest {
  name: string
  quota_daily?: number
  concurrency_limit?: number
}

export interface CreateKeyResponse {
  key: string
  data: UserAPIKey
}

export interface UsageListParams {
  page?: number
  limit?: number
  start_date?: string
  end_date?: string
}

export interface UsageListResponse {
  data: UsageRecord[]
  total: number
}

export interface ExportResponse {
  download_url: string
}

export interface DashboardStats {
  total_keys: number
  monthly_requests: number
  monthly_cost: number
  monthly_tokens: number
}
