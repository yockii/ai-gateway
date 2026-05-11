import client from './client'
import type { MembershipPlan, UserMembership } from '@/types/models'

export const plansApi = {
  list: () => client.get<MembershipPlan[]>('/public/plans'),
  getCurrent: () => client.get<UserMembership>('/user/plans/current'),
  subscribe: (planId: string, interval: 'monthly' | 'yearly') =>
    client.post<{ client_secret: string }>(`/user/plans/${planId}/subscribe`, { interval }),
  cancel: () => client.post('/user/plans/cancel'),
}
