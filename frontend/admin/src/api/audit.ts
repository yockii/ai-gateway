import client from './client'
import type { AuditLog } from '@/types/models'

export interface AuditLogFilter {
  entity_type?: string
  entity_id?: string
  action?: string
  admin_id?: string
  start_time?: string
  end_time?: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface AuditLogListResponse {
  data: AuditLog[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export const auditApi = {
  listLogs: (filter: AuditLogFilter = {}) => client.get<AuditLogListResponse>('/audit/logs', { params: filter }),
  getLogDetail: (id: string) => client.get<AuditLog>(`/audit/logs/${id}`),
  getEntityHistory: (entityType: string, entityId: string) =>
    client.get<AuditLog[]>('/audit/history', {
      params: { entity_type: entityType, entity_id: entityId },
    }),
}
