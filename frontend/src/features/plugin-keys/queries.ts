import { useApiQuery, useApiMutation } from '@/lib/swr'
import { pluginKeysApi } from './api'

export function usePluginKeysQuery(integrationId: number) {
  return useApiQuery(
    integrationId ? `plugin-keys-${integrationId}` : null,
    () => pluginKeysApi.list(integrationId),
  )
}

export function useCreatePluginKeyMutation(integrationId: number) {
  return useApiMutation(
    'plugin-keys/create',
    (label: string) => pluginKeysApi.create(integrationId, label),
    [`plugin-keys-${integrationId}`],
  )
}

export function useRevokePluginKeyMutation(integrationId: number) {
  return useApiMutation(
    'plugin-keys/revoke',
    (id: number) => pluginKeysApi.revoke(id),
    [`plugin-keys-${integrationId}`],
  )
}
