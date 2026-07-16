import { api } from '@/lib/api'
import type { PluginKey, CreateKeyResponse } from './types'

// POS-plugin API keys, scoped per integration (iiko, r-keeper, …).
export const pluginKeysApi = {
  list: async (integrationId: number): Promise<PluginKey[]> => {
    const response = await api.get<PluginKey[] | null>('/pos-plugin/admin/keys', {
      params: { integration_id: integrationId },
    })
    return response.data ?? []
  },

  create: async (
    integrationId: number,
    label: string,
  ): Promise<CreateKeyResponse> => {
    const response = await api.post<CreateKeyResponse>(
      '/pos-plugin/admin/keys',
      { integration_id: integrationId, label },
    )
    return response.data
  },

  revoke: async (id: number): Promise<void> => {
    await api.delete(`/pos-plugin/admin/keys/${id}`)
  },
}
