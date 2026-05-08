import client from './client'
import type { CreateKeyRequest, CreateKeyResponse } from '@/types/api'
import type { UserAPIKey } from '@/types/models'

export const keysApi = {
  list: () => client.get<UserAPIKey[]>('/user/keys'),
  create: (data: CreateKeyRequest) => client.post<CreateKeyResponse>('/user/keys', data),
  delete: (id: string) => client.delete(`/user/keys/${id}`),
  disable: (id: string) => client.patch(`/user/keys/${id}/disable`),
  enable: (id: string) => client.patch(`/user/keys/${id}/enable`),
  getStats: (id: string) => client.get(`/user/keys/${id}/stats`),
}
