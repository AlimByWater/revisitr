import { api } from '@/lib/api'
import type {
  SalesAnalytics,
  LoyaltyAnalytics,
  CampaignAnalytics,
  AnalyticsFilter,
} from './types'

export const analyticsApi = {
  getSales: async (filter: AnalyticsFilter): Promise<SalesAnalytics> => {
    const response = await api.get<SalesAnalytics>('/analytics/sales', {
      params: filter,
    })
    return response.data
  },

  getLoyalty: async (filter: AnalyticsFilter): Promise<LoyaltyAnalytics> => {
    const response = await api.get<LoyaltyAnalytics>('/analytics/loyalty', {
      params: filter,
    })
    return response.data
  },

  getCampaigns: async (
    filter: AnalyticsFilter,
  ): Promise<CampaignAnalytics> => {
    const response = await api.get<CampaignAnalytics>('/analytics/campaigns', {
      params: filter,
    })
    return response.data
  },
}
