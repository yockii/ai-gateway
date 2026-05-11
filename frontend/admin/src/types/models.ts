export interface AdminUser {
  id: string
  email: string
  name: string
  role?: 'superadmin' | 'admin' | 'operator'
  is_active?: boolean
  created_at?: string
  updated_at?: string
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

// Phase 6: Supplier API Key
export interface SupplierApiKey {
  id: string
  supplier_id: string
  name: string
  key_prefix: string
  priority: number
  is_primary: boolean
  max_requests: number
  current_requests: number
  is_active: boolean
  last_used_at: string | null
  expire_at: string | null
  created_at: string
  updated_at: string
}

// Phase 6: Supplier Model
export interface SupplierModel {
  id: string
  supplier_id: string
  model_id: string
  input_cost: number
  output_cost: number
  is_active: boolean
  effective_date: string
  created_at: string
  updated_at: string
}

// Phase 6: Health Check Result
export interface HealthCheckResult {
  supplier_id: string
  is_healthy: boolean
  latency: number
  error: string
  timestamp: string
}

// Phase 5: Enterprise Pricing
export interface EnterprisePricing {
  id: string
  customer_id: string
  customer_name: string
  model_id: string
  input_price: number
  output_price: number
  min_profit_margin: number
  max_cost_price: number
  effective_date: string
  expiry_date: string | null
  is_active: boolean
  created_at: string
  updated_at: string
}

// Phase 6: Price History
export interface PriceHistory {
  id: string
  supplier_model_id: string
  input_cost: number
  output_cost: number
  changed_at: string
  changed_by: string
  reason?: string
}

// Phase 8: Supplier Event
export interface SupplierFailureEvent {
  id: string
  supplier_id: string
  model_id: string
  error_type: string
  error_msg: string
  status_code: number
  timestamp: string
}

export interface FailoverEvent {
  id: string
  from_supplier_id: string
  to_supplier_id: string
  model_id: string
  reason: string
  timestamp: string
}

export interface HealthCheckHistory {
  id: string
  supplier_id: string
  is_healthy: boolean
  latency: number
  error: string
  timestamp: string
}

// Phase 9: Audit Log
export interface AuditLog {
  id: string
  admin_id: string
  admin_name: string
  entity_type: 'supplier' | 'pricing' | 'api_key' | 'model' | 'user'
  entity_id: string
  action: 'create' | 'update' | 'delete' | 'rotate' | 'set_primary'
  changes: {
    before?: Record<string, any>
    after?: Record<string, any>
  }
  ip_address: string
  user_agent: string
  timestamp: string
}

export interface HealthStatus {
  is_healthy: boolean
  latency: number
  last_check: string
}
