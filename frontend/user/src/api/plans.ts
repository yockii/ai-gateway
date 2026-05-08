import client from './client'
import type { MembershipPlan, UserMembership } from '@/types/models'

export const plansApi = {
  list: () => client.get<MembershipPlan[]>('/plans'),
  getCurrent: () => client.get<UserMembership>('/plans/current'),
  subscribe: (planId: string, interval: 'monthly' | 'yearly') =>
    client.post<{ client_secret: string }>(`/plans/${planId}/subscribe`, { interval }),
  cancel: () => client.post('/plans/cancel'),
}
