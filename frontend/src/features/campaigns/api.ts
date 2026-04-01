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
  CampaignVariant,
  CreateABTestRequest,
  ABTestResults,
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

  // A/B testing
  createABTest: async (id: number, data: CreateABTestRequest): Promise<CampaignVariant[]> => {
    const response = await api.post<CampaignVariant[]>(`/campaigns/${id}/variants`, data)
    return response.data
  },

  getVariants: async (id: number): Promise<CampaignVariant[]> => {
    const response = await api.get<CampaignVariant[]>(`/campaigns/${id}/variants`)
    return response.data
  },

  getABResults: async (id: number): Promise<ABTestResults> => {
    const response = await api.get<ABTestResults>(`/campaigns/${id}/ab-results`)
    return response.data
  },

  pickWinner: async (campaignId: number, variantId: number): Promise<void> => {
    await api.post(`/campaigns/${campaignId}/variants/${variantId}/winner`)
  },

  // File upload
  uploadFile: async (file: File): Promise<string> => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await api.post<{ url: string }>('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
    })
    return response.data.url
  },
}
