import client from './client'

export interface CreateTierRequest {
  name: string
  display_name: string
  level: number
}

export const plansApi = {
  listTiers: () => client.get<MembershipTier[]>('/plans/tiers'),
  createTier: (data: CreateTierRequest) => client.post<MembershipTier>('/plans/tiers', data),
  updateTier: (id: string, data: Partial<MembershipTier>) => client.put(`/plans/tiers/${id}`, data),
  deleteTier: (id: string) => client.delete(`/plans/tiers/${id}`),
}
