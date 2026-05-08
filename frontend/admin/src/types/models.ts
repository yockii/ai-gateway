export interface AdminUser {
  id: string
  email: string
  name: string
  role: 'superadmin' | 'admin' | 'operator'
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface User {
  id: string
  email: string
  name: string
  user_group_id: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ExternalModel {
  id: string
  name: string
  display_name: string
  model_type: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Supplier {
  id: string
  name: string
  display_name: string
  provider: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface MembershipTier {
  id: string
  name: string
  display_name: string
  level: number
  is_active: boolean
  created_at: string
}

export interface SystemMetrics {
  qps: number
  avg_latency: number
  error_rate: number
  active_users: number
  total_requests: number
  cpu_usage: number
  memory_usage: number
}

export interface Alert {
  id: string
  type: 'error' | 'warning' | 'info'
  message: string
  created_at: string
  resolved: boolean
}
