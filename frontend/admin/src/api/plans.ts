import client from './client'
import type { MembershipTier } from '@/types/models'

export interface CreateTierRequest {
  name: string
  display_name: string
  level: number
}

export const plansApi = {
  listTiers: () => client.get<MembershipTier[]>('/plans'),
  createTier: (data: CreateTierRequest) => client.post<MembershipTier>('/plans', data),
  updateTier: (id: string, data: Partial<MembershipTier>) => client.put(`/plans/${id}`, data),
  deleteTier: (id: string) => client.delete(`/plans/${id}`),
}
