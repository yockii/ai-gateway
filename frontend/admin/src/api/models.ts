import client from './client'

export interface CreateModelRequest {
  name: string
  display_name: string
  model_type: string
}

export const modelsApi = {
  list: () => client.get<ExternalModel[]>('/models'),
  create: (data: CreateModelRequest) => client.post<ExternalModel>('/models', data),
  update: (id: string, data: Partial<ExternalModel>) => client.put(`/models/${id}`, data),
  delete: (id: string) => client.delete(`/models/${id}`),
  toggleStatus: (id: string) => client.patch(`/models/${id}/status`),
}
