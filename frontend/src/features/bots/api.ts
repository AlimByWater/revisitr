import { api } from '@/lib/api'
import type { Bot, BotSettings, CreateBotRequest } from './types'

export const botsApi = {
  list: async (): Promise<Bot[]> => {
    const response = await api.get<Bot[]>('/bots')
    return response.data
  },

  getById: async (id: number): Promise<Bot> => {
    const response = await api.get<Bot>(`/bots/${id}`)
    return response.data
  },

  create: async (data: CreateBotRequest): Promise<Bot> => {
    const response = await api.post<Bot>('/bots', data)
    return response.data
  },

  update: async (id: number, data: Partial<Bot>): Promise<Bot> => {
    const response = await api.patch<Bot>(`/bots/${id}`, data)
    return response.data
  },

  remove: async (id: number): Promise<void> => {
    await api.delete(`/bots/${id}`)
  },

  getSettings: async (id: number): Promise<BotSettings> => {
    const response = await api.get<BotSettings>(`/bots/${id}/settings`)
    return response.data
  },

  updateSettings: async (id: number, data: Partial<BotSettings>): Promise<void> => {
    await api.patch(`/bots/${id}/settings`, data)
  },
}
