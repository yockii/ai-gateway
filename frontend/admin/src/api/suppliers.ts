import client from './client'

export interface CreateSupplierRequest {
  name: string
  display_name: string
  provider: string
}

export const suppliersApi = {
  list: () => client.get<Supplier[]>('/suppliers'),
  create: (data: CreateSupplierRequest) => client.post<Supplier>('/suppliers', data),
  update: (id: string, data: Partial<Supplier>) => client.put(`/suppliers/${id}`, data),
  delete: (id: string) => client.delete(`/suppliers/${id}`),
  testConnection: (id: string) => client.post(`/suppliers/${id}/test`),
}
