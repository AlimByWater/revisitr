import { api } from '@/lib/api'
import type {
  WalletConfig,
  SaveWalletConfigRequest,
  WalletPass,
  IssueWalletPassRequest,
  WalletStats,
} from './types'

export const walletApi = {
  getConfigs: async (): Promise<WalletConfig[]> => {
    const response = await api.get<WalletConfig[]>('/wallet/configs')
    return response.data
  },

  getConfig: async (platform: string): Promise<WalletConfig> => {
    const response = await api.get<WalletConfig>(`/wallet/configs/${platform}`)
    return response.data
  },

  saveConfig: async (platform: string, data: SaveWalletConfigRequest): Promise<WalletConfig> => {
    const response = await api.put<WalletConfig>(`/wallet/configs/${platform}`, data)
    return response.data
  },

  deleteConfig: async (platform: string): Promise<void> => {
    await api.delete(`/wallet/configs/${platform}`)
  },

  getPasses: async (): Promise<WalletPass[]> => {
    const response = await api.get<WalletPass[]>('/wallet/passes')
    return response.data
  },

  issuePass: async (data: IssueWalletPassRequest): Promise<WalletPass> => {
    const response = await api.post<WalletPass>('/wallet/passes', data)
    return response.data
  },

  revokePass: async (id: number): Promise<void> => {
    await api.post(`/wallet/passes/${id}/revoke`)
  },

  getStats: async (): Promise<WalletStats> => {
    const response = await api.get<WalletStats>('/wallet/stats')
    return response.data
  },
}
