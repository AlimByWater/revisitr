import { api } from '@/lib/api'
import type {
  Bot, BotSettings, CreateBotRequest,
  CreateManagedBotRequest, CreateManagedBotResponse,
  ActivationLinkResponse, PostCode,
  ModulePreset, BotModuleSettings,
} from './types'

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

  // Managed bots
  getActivationLink: async (): Promise<ActivationLinkResponse> => {
    const response = await api.post<ActivationLinkResponse>('/bots/activation-link')
    return response.data
  },

  createManaged: async (data: CreateManagedBotRequest): Promise<CreateManagedBotResponse> => {
    const response = await api.post<CreateManagedBotResponse>('/bots/create-managed', data)
    return response.data
  },

  getBotStatus: async (id: number): Promise<{ status: string }> => {
    const response = await api.get<{ status: string }>(`/bots/${id}/status`)
    return response.data
  },
}

export const modulePresetsApi = {
  listPresets: async (moduleKey: string): Promise<ModulePreset[]> => {
    const response = await api.get<ModulePreset[]>(`/module-settings/catalog/${moduleKey}`)
    return response.data
  },

  getBotModuleSettings: async (botId: number, moduleKey: string): Promise<BotModuleSettings> => {
    const response = await api.get<BotModuleSettings>(`/module-settings/${botId}/${moduleKey}`)
    return response.data
  },

  selectPreset: async (botId: number, moduleKey: string, presetKey: string): Promise<void> => {
    await api.put(`/module-settings/${botId}/${moduleKey}/preset`, { preset_key: presetKey })
  },

  updateCustomizations: async (botId: number, moduleKey: string, customizations: Record<string, unknown>): Promise<void> => {
    await api.patch(`/module-settings/${botId}/${moduleKey}/customizations`, { customizations })
  },

  resetPreset: async (botId: number, moduleKey: string): Promise<void> => {
    await api.post(`/module-settings/${botId}/${moduleKey}/reset`)
  },
}

export const postsApi = {
  list: async (): Promise<PostCode[]> => {
    const response = await api.get<PostCode[]>('/posts')
    return response.data
  },

  getByCode: async (code: string): Promise<PostCode> => {
    const response = await api.get<PostCode>(`/posts/${code}`)
    return response.data
  },

  remove: async (code: string): Promise<void> => {
    await api.delete(`/posts/${code}`)
  },
}
