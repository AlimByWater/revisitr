import { api } from '@/lib/api'
import type {
  Campaign,
  AutoScenario,
  AutoActionLog,
  CampaignAnalyticsDetail,
  CreateCampaignRequest,
  UpdateCampaignRequest,
  CreateScenarioRequest,
  UpdateScenarioRequest,
  AudienceFilter,
  PaginatedResponse,
} from './types'

export const campaignsApi = {
  list: async (
    limit = 20,
    offset = 0,
  ): Promise<PaginatedResponse<Campaign>> => {
    const response = await api.get<PaginatedResponse<Campaign>>('/campaigns', {
      params: { limit, offset },
    })
    return response.data
  },

  getById: async (id: number): Promise<Campaign> => {
    const response = await api.get<Campaign>(`/campaigns/${id}`)
    return response.data
  },

  create: async (data: CreateCampaignRequest): Promise<Campaign> => {
    const response = await api.post<Campaign>('/campaigns', data)
    return response.data
  },

  update: async (id: number, data: UpdateCampaignRequest): Promise<void> => {
    await api.patch(`/campaigns/${id}`, data)
  },

  remove: async (id: number): Promise<void> => {
    await api.delete(`/campaigns/${id}`)
  },

  send: async (id: number): Promise<void> => {
    await api.post(`/campaigns/${id}/send`)
  },

  previewAudience: async (filter: AudienceFilter): Promise<number> => {
    const response = await api.post<{ count: number }>(
      '/campaigns/preview',
      filter,
    )
    return response.data.count
  },

  // Scenarios
  listScenarios: async (): Promise<AutoScenario[]> => {
    const response = await api.get<AutoScenario[]>('/campaigns/scenarios')
    return response.data
  },

  createScenario: async (
    data: CreateScenarioRequest,
  ): Promise<AutoScenario> => {
    const response = await api.post<AutoScenario>(
      '/campaigns/scenarios',
      data,
    )
    return response.data
  },

  updateScenario: async (
    id: number,
    data: UpdateScenarioRequest,
  ): Promise<void> => {
    await api.patch(`/campaigns/scenarios/${id}`, data)
  },

  deleteScenario: async (id: number): Promise<void> => {
    await api.delete(`/campaigns/scenarios/${id}`)
  },

  // Campaign scheduling & analytics
  schedule: async (id: number, scheduledAt: string): Promise<void> => {
    await api.post(`/campaigns/${id}/schedule`, { scheduled_at: scheduledAt })
  },

  cancelSchedule: async (id: number): Promise<void> => {
    await api.delete(`/campaigns/${id}/schedule`)
  },

  getAnalytics: async (id: number): Promise<CampaignAnalyticsDetail> => {
    const response = await api.get<CampaignAnalyticsDetail>(
      `/campaigns/${id}/analytics`,
    )
    return response.data
  },

  recordClick: async (
    id: number,
    clientId: number,
    buttonIdx?: number,
    url?: string,
  ): Promise<void> => {
    await api.post(`/campaigns/${id}/click`, {
      client_id: clientId,
      button_idx: buttonIdx,
      url,
    })
  },

  // Auto-action templates
  getTemplates: async (): Promise<AutoScenario[]> => {
    const response = await api.get<AutoScenario[]>(
      '/campaigns/scenarios/templates',
    )
    return response.data
  },

  cloneTemplate: async (
    key: string,
    botId: number,
  ): Promise<AutoScenario> => {
    const response = await api.post<AutoScenario>(
      `/campaigns/scenarios/templates/${key}/clone`,
      { bot_id: botId },
    )
    return response.data
  },

  getActionLog: async (
    scenarioId: number,
    limit = 20,
    offset = 0,
  ): Promise<PaginatedResponse<AutoActionLog>> => {
    const response = await api.get<PaginatedResponse<AutoActionLog>>(
      `/campaigns/scenarios/${scenarioId}/log`,
      { params: { limit, offset } },
    )
    return response.data
  },
}
