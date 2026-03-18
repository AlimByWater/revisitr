import { api } from '@/lib/api'
import type {
  ClientProfile,
  ClientFilter,
  ClientStats,
  PaginatedResponse,
  UpdateClientRequest,
} from './types'

export const clientsApi = {
  list: async (filter: ClientFilter): Promise<PaginatedResponse<ClientProfile>> => {
    const response = await api.get<PaginatedResponse<ClientProfile>>('/clients', {
      params: filter,
    })
    return response.data
  },

  getById: async (id: number): Promise<ClientProfile> => {
    const response = await api.get<ClientProfile>(`/clients/${id}`)
    return response.data
  },

  getStats: async (): Promise<ClientStats> => {
    const response = await api.get<ClientStats>('/clients/stats')
    return response.data
  },

  updateTags: async (id: number, data: UpdateClientRequest): Promise<void> => {
    await api.patch(`/clients/${id}`, data)
  },

  countByFilter: async (
    filter: Partial<ClientFilter>,
  ): Promise<{ count: number }> => {
    const response = await api.get<{ count: number }>('/clients/count', {
      params: filter,
    })
    return response.data
  },
}
