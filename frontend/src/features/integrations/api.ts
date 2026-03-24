import { api } from '@/lib/api'
import type {
  Integration,
  IntegrationAggregate,
  DashboardAggregates,
  CreateIntegrationRequest,
  UpdateIntegrationRequest,
  ExternalOrder,
  IntegrationStats,
  POSCustomer,
  POSMenu,
  PaginatedResponse,
} from './types'

export const integrationsApi = {
  list: async (): Promise<Integration[]> => {
    const response = await api.get<Integration[]>('/integrations')
    return response.data
  },

  getById: async (id: number): Promise<Integration> => {
    const response = await api.get<Integration>(`/integrations/${id}`)
    return response.data
  },

  create: async (data: CreateIntegrationRequest): Promise<Integration> => {
    const response = await api.post<Integration>('/integrations', data)
    return response.data
  },

  update: async (
    id: number,
    data: UpdateIntegrationRequest,
  ): Promise<Integration> => {
    const response = await api.patch<Integration>(`/integrations/${id}`, data)
    return response.data
  },

  remove: async (id: number): Promise<void> => {
    await api.delete(`/integrations/${id}`)
  },

  sync: async (id: number): Promise<void> => {
    await api.post(`/integrations/${id}/sync`)
  },

  testConnection: async (id: number): Promise<void> => {
    await api.post(`/integrations/${id}/test`)
  },

  getOrders: async (
    id: number,
    limit = 20,
    offset = 0,
  ): Promise<PaginatedResponse<ExternalOrder>> => {
    const response = await api.get<PaginatedResponse<ExternalOrder>>(
      `/integrations/${id}/orders`,
      { params: { limit, offset } },
    )
    return response.data
  },

  getCustomers: async (
    id: number,
    limit = 20,
    offset = 0,
    search = '',
  ): Promise<POSCustomer[]> => {
    const response = await api.get<POSCustomer[]>(
      `/integrations/${id}/customers`,
      { params: { limit, offset, search } },
    )
    return response.data
  },

  getMenu: async (id: number): Promise<POSMenu> => {
    const response = await api.get<POSMenu>(`/integrations/${id}/menu`)
    return response.data
  },

  getStats: async (id: number): Promise<IntegrationStats> => {
    const response = await api.get<IntegrationStats>(
      `/integrations/${id}/stats`,
    )
    return response.data
  },

  getAggregates: async (
    id: number,
    from: string,
    to: string,
  ): Promise<IntegrationAggregate[]> => {
    const response = await api.get<IntegrationAggregate[]>(
      `/integrations/${id}/aggregates`,
      { params: { from, to } },
    )
    return response.data
  },

  getSalesData: async (
    from: string,
    to: string,
  ): Promise<DashboardAggregates> => {
    const response = await api.get<DashboardAggregates>('/dashboard/sales', {
      params: { from, to },
    })
    return response.data
  },
}
