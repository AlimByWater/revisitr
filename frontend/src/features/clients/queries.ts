import { useApiQuery, useApiMutation } from '../../lib/swr'
import { clientsApi } from './api'
import type { ClientFilter, UpdateClientRequest } from './types'

export function useClientsQuery(filter: ClientFilter) {
  return useApiQuery(
    `clients-${JSON.stringify(filter)}`,
    () => clientsApi.list(filter),
    { keepPreviousData: true },
  )
}

export function useClientProfileQuery(id: number) {
  return useApiQuery(id ? `clients-${id}` : null, () =>
    clientsApi.getById(id),
  )
}

export function useClientStatsQuery() {
  return useApiQuery('clients-stats', clientsApi.getStats)
}

export function useUpdateTagsMutation() {
  return useApiMutation(
    'clients/update-tags',
    ({ id, data }: { id: number; data: UpdateClientRequest }) =>
      clientsApi.updateTags(id, data),
    ['clients'],
  )
}
