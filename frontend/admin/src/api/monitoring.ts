import client from './client'
import type { SystemMetrics, Alert } from '@/types/models'

export const monitoringApi = {
  getMetrics: () => client.get<SystemMetrics>('/monitoring/metrics'),
  getAlerts: () => client.get<Alert[]>('/monitoring/alerts'),
  getLogs: (params: { page?: number; limit?: number }) =>
    client.get('/monitoring/logs', { params }),
}
