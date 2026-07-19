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

  uploadLogo: async (file: File): Promise<string> => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await api.post<{ url: string }>('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data.url
  },

  getDownloadURL: (serial: string): string => {
    const base = import.meta.env.VITE_API_URL || '/api/v1'
    return `${base}/wallet/passes/${serial}/download`
  },

  getGoogleSaveURL: async (serial: string): Promise<string> => {
    const response = await api.get<{ url: string }>(`/wallet/passes/${serial}/google-save`)
    return response.data.url
  },
}
