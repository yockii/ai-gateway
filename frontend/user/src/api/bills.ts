import client from './client'
import type { ExportResponse } from '@/types/api'
import type { Bill } from '@/types/models'

export const billsApi = {
  list: () => client.get<Bill[]>('/user/bills'),
  getById: (id: string) => client.get<Bill>(`/user/bills/${id}`),
  export: (id: string, format: 'pdf' | 'csv') =>
    client.post<ExportResponse>(`/user/bills/${id}/export`, { format }),
}
