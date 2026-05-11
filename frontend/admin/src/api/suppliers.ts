import client from './client'
import type { Supplier, SupplierApiKey, SupplierModel, HealthCheckResult, PriceHistory } from '@/types/models'

export interface CreateSupplierRequest {
  name: string
  display_name: string
  provider: string
}

export interface CreateApiKeyRequest {
  name: string
  api_key: string
  priority?: number
  is_primary?: boolean
  max_requests?: number
  expire_at?: string
}

export interface UpdateApiKeyRequest {
  name?: string
  priority?: number
  max_requests?: number
  is_active?: boolean
  expire_at?: string
}

export interface AddModelRequest {
  model_id: string
  input_cost: number
  output_cost: number
}

export const suppliersApi = {
  // Basic supplier operations
  list: () => client.get<Supplier[]>('/suppliers'),
  create: (data: CreateSupplierRequest) => client.post<Supplier>('/suppliers', data),
  update: (id: string, data: Partial<Supplier>) => client.put(`/suppliers/${id}`, data),
  delete: (id: string) => client.delete(`/suppliers/${id}`),

  // API Key management
  listApiKeys: (supplierId: string) => client.get<SupplierApiKey[]>(`/suppliers/${supplierId}/api-keys`),
  createApiKey: (supplierId: string, data: CreateApiKeyRequest) => client.post<SupplierApiKey>(`/suppliers/${supplierId}/api-keys`, data),
  updateApiKey: (supplierId: string, keyId: string, data: UpdateApiKeyRequest) => client.put(`/suppliers/${supplierId}/api-keys/${keyId}`, data),
  deleteApiKey: (supplierId: string, keyId: string) => client.delete(`/suppliers/${supplierId}/api-keys/${keyId}`),
  setPrimary: (supplierId: string, keyId: string) => client.patch(`/suppliers/${supplierId}/api-keys/${keyId}/set-primary`),
  rotateApiKey: (supplierId: string, keyId: string) => client.post(`/suppliers/${supplierId}/api-keys/${keyId}/rotate`),
  getApiKeyStats: (supplierId: string, keyId: string) => client.get(`/suppliers/${supplierId}/api-keys/${keyId}/stats`),

  // Model association management
  listModels: (supplierId: string) => client.get<SupplierModel[]>(`/suppliers/${supplierId}/models`),
  addModel: (supplierId: string, data: AddModelRequest) => client.post(`/suppliers/${supplierId}/models`, data),
  updateModelCost: (supplierId: string, modelId: string, data: { input_cost: number; output_cost: number }) =>
    client.put(`/suppliers/${supplierId}/models/${modelId}`, data),
  removeModel: (supplierId: string, modelId: string) => client.delete(`/suppliers/${supplierId}/models/${modelId}`),
  getPriceHistory: (supplierId: string, modelId: string) => client.get<PriceHistory[]>(`/suppliers/${supplierId}/models/${modelId}/history`),

  // Health check
  getHealth: (supplierId: string) => client.get<HealthCheckResult>(`/suppliers/${supplierId}/health`),
  getHealthHistory: (supplierId: string, limit = 50) => client.get<HealthCheckHistory[]>(`/suppliers/${supplierId}/health/history`, { params: { limit } }),
  triggerHealthCheck: (supplierId: string) => client.post(`/suppliers/${supplierId}/health/check`),
}
