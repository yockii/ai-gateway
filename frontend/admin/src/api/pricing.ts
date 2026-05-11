import client from './client'
import type { EnterprisePricing } from '@/types/models'

export interface CreateEnterprisePricingRequest {
  customer_id: string
  customer_name: string
  model_id: string
  input_price: number
  output_price: number
  min_profit_margin?: number
  max_cost_price?: number
  effective_date: string
  expiry_date?: string
}

export interface UpdateEnterprisePricingRequest {
  customer_id?: string
  customer_name?: string
  model_id?: string
  input_price?: number
  output_price?: number
  min_profit_margin?: number
  max_cost_price?: number
  effective_date?: string
  expiry_date?: string
  is_active?: boolean
}

export const pricingApi = {
  // Enterprise pricing management
  listEnterprise: () => client.get<EnterprisePricing[]>('/enterprise-pricing'),
  createEnterprise: (data: CreateEnterprisePricingRequest) => client.post<EnterprisePricing>('/enterprise-pricing', data),
  updateEnterprise: (id: string, data: UpdateEnterprisePricingRequest) => client.put(`/enterprise-pricing/${id}`, data),
  deleteEnterprise: (id: string) => client.delete(`/enterprise-pricing/${id}`),
  getEnterprise: (id: string) => client.get<EnterprisePricing>(`/enterprise-pricing/${id}`),
}
