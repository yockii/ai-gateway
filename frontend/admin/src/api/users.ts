import client from './client'

export interface UserListParams {
  page?: number
  limit?: number
  search?: string
}

export interface UserListResponse {
  data: User[]
  total: number
}

export const usersApi = {
  list: (params: UserListParams) => client.get<UserListResponse>('/users', { params }),
  getById: (id: string) => client.get<User>(`/users/${id}`),
  update: (id: string, data: Partial<User>) => client.put(`/users/${id}`, data),
  delete: (id: string) => client.delete(`/users/${id}`),
  toggleStatus: (id: string) => client.patch(`/users/${id}/status`),
}
