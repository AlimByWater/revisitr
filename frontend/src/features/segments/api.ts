import { api } from '@/lib/api'
import type {
  Segment,
  CreateSegmentRequest,
  UpdateSegmentRequest,
  SegmentRule,
  CreateSegmentRuleRequest,
  ClientPrediction,
  PredictionSummary,
} from './types'
import type { PaginatedResponse } from '@/features/clients/types'
import type { ClientProfile } from '@/features/clients/types'

export const segmentsApi = {
  list: async (): Promise<Segment[]> => {
    const response = await api.get<Segment[]>('/segments')
    return response.data
  },

  getById: async (id: number): Promise<Segment> => {
    const response = await api.get<Segment>(`/segments/${id}`)
    return response.data
  },

  create: async (data: CreateSegmentRequest): Promise<Segment> => {
    const response = await api.post<Segment>('/segments', data)
    return response.data
  },

  update: async (id: number, data: UpdateSegmentRequest): Promise<Segment> => {
    const response = await api.patch<Segment>(`/segments/${id}`, data)
    return response.data
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/segments/${id}`)
  },

  getClients: async (
    segmentId: number,
    limit = 20,
    offset = 0,
  ): Promise<PaginatedResponse<ClientProfile>> => {
    const response = await api.get<PaginatedResponse<ClientProfile>>(
      `/segments/${segmentId}/clients`,
      { params: { limit, offset } },
    )
    return response.data
  },

  recalculate: async (segmentId: number): Promise<void> => {
    await api.post(`/segments/${segmentId}/recalculate`)
  },

  previewCount: async (filter: SegmentFilter): Promise<{ count: number }> => {
    const response = await api.post<{ count: number }>('/segments/preview', filter)
    return response.data
  },

  // Rules
  getRules: async (segmentId: number): Promise<SegmentRule[]> => {
    const response = await api.get<SegmentRule[]>(`/segments/${segmentId}/rules`)
    return response.data
  },

  addRule: async (segmentId: number, data: CreateSegmentRuleRequest): Promise<SegmentRule> => {
    const response = await api.post<SegmentRule>(`/segments/${segmentId}/rules`, data)
    return response.data
  },

  deleteRule: async (segmentId: number, ruleId: number): Promise<void> => {
    await api.delete(`/segments/${segmentId}/rules/${ruleId}`)
  },

  // Predictions
  getPredictions: async (
    limit = 20,
    offset = 0,
  ): Promise<ClientPrediction[]> => {
    const response = await api.get<ClientPrediction[]>('/segments/predictions', {
      params: { limit, offset },
    })
    return response.data
  },

  getPredictionSummary: async (): Promise<PredictionSummary> => {
    const response = await api.get<PredictionSummary>('/segments/predictions/summary')
    return response.data
  },

  getHighChurnClients: async (threshold = 0.7): Promise<ClientPrediction[]> => {
    const response = await api.get<ClientPrediction[]>('/segments/predictions/high-churn', {
      params: { threshold },
    })
    return response.data
  },
}
