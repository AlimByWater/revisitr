import { api } from '@/lib/api'
import type { RFMDashboard, RFMConfig, UpdateRFMConfigRequest } from './types'

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
}
