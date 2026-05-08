import client from './client'
import type { ExportResponse } from '@/types/api'
import type { Bill } from '@/types/models'

export const billsApi = {
  list: () => client.get<Bill[]>('/bills'),
  getById: (id: string) => client.get<Bill>(`/bills/${id}`),
  export: (id: string, format: 'pdf' | 'csv') =>
    client.post<ExportResponse>(`/bills/${id}/export`, { format }),
}
