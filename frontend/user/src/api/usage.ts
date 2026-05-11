import client from './client'
import type { UsageListParams, UsageListResponse, DashboardStats } from '@/types/api'

export const usageApi = {
  list: (params: UsageListParams) => client.get<UsageListResponse>('/user/usage', { params }),
  getStats: () => client.get<DashboardStats>('/user/usage/stats'),
}
