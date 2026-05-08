export interface User {
  id: string
  email: string
  name: string
  user_group_id: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UserAPIKey {
  id: string
  user_id: string
  name: string
  quota_daily: number
  quota_monthly: number
  concurrency_limit: number
  model_concurrency: Record<string, number>
  expires_at: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UsageRecord {
  id: string
  request_id: string
  user_id: string
  model_id: string
  supplier_id: string
  input_tokens: number
  output_tokens: number
  total_tokens: number
  cost_price: number
  selling_price: number
  profit: number
  created_at: string
}

export interface Bill {
  id: string
  user_id: string
  period: string
  start_date: string
  end_date: string
  items: BillModelDetail[]
  total_requests: number
  total_cost: number
  total_revenue: number
  total_profit: number
  status: string
  created_at: string
  updated_at: string
}

export interface BillModelDetail {
  model_id: string
  request_count: number
  input_tokens: number
  output_tokens: number
  total_tokens: number
  total_cost: number
  total_revenue: number
  total_profit: number
}

export interface MembershipPlan {
  id: string
  name: string
  description: string
  price_monthly: number
  price_yearly: number
  features: string[]
  is_active: boolean
}

export interface UserMembership {
  id: string
  user_id: string
  plan_id: string
  status: 'active' | 'expired' | 'cancelled'
  current_period_start: string
  current_period_end: string
  auto_renew: boolean
}
