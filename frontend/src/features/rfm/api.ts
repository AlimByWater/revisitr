import { api } from '@/lib/api'
import type {
  RFMDashboard,
  RFMConfig,
  UpdateRFMConfigRequest,
  RFMTemplate,
  SetTemplateRequest,
  ActiveTemplateResponse,
  OnboardingQuestion,
  TemplateRecommendation,
  SegmentClientsResponse,
} from './types'

export const rfmApi = {
  getDashboard: async (): Promise<RFMDashboard> => {
    const response = await api.get<RFMDashboard>('/rfm/dashboard')
    return response.data
  },

  recalculate: async (): Promise<void> => {
    await api.post('/rfm/recalculate')
  },

  getConfig: async (): Promise<RFMConfig> => {
    const response = await api.get<RFMConfig>('/rfm/config')
    return response.data
  },

  updateConfig: async (data: UpdateRFMConfigRequest): Promise<RFMConfig> => {
    const response = await api.patch<RFMConfig>('/rfm/config', data)
    return response.data
  },

  // Templates
  listTemplates: async (): Promise<RFMTemplate[]> => {
    const response = await api.get<{ templates: RFMTemplate[] }>('/rfm/templates')
    return response.data.templates
  },

  getActiveTemplate: async (): Promise<ActiveTemplateResponse> => {
    const response = await api.get<ActiveTemplateResponse>('/rfm/template')
    return response.data
  },

  setTemplate: async (data: SetTemplateRequest): Promise<{ message: string; template: RFMTemplate }> => {
    const response = await api.put<{ message: string; template: RFMTemplate }>('/rfm/template', data)
    return response.data
  },

  // Onboarding
  getOnboardingQuestions: async (): Promise<OnboardingQuestion[]> => {
    const response = await api.get<{ questions: OnboardingQuestion[] }>('/rfm/onboarding/questions')
    return response.data.questions
  },

  recommendTemplate: async (answers: number[]): Promise<TemplateRecommendation> => {
    const response = await api.post<TemplateRecommendation>('/rfm/onboarding/recommend', { answers })
    return response.data
  },

  // Segment detail
  getSegmentClients: async (
    segment: string,
    params: { page?: number; per_page?: number; sort?: string; order?: string } = {},
  ): Promise<SegmentClientsResponse> => {
    const response = await api.get<SegmentClientsResponse>(`/rfm/segments/${segment}/clients`, {
      params,
    })
    return response.data
  },
}
