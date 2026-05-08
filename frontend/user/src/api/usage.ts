import client from './client'
import type { UsageListParams, UsageListResponse, DashboardStats } from '@/types/api'

export const usageApi = {
  list: (params: UsageListParams) => client.get<UsageListResponse>('/usage', { params }),
  getStats: () => client.get<DashboardStats>('/usage/stats'),
}
