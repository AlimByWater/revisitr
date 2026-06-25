import { useApiQuery, useApiMutation } from '@/lib/swr'
import { integrationsApi } from './api'
import type { CreateIntegrationRequest, UpdateIntegrationRequest } from './types'

export function useIntegrationsQuery() {
  return useApiQuery('integrations', integrationsApi.list)
}

export function useIntegrationQuery(id: number) {
  return useApiQuery(id ? `integrations-${id}` : null, () =>
    integrationsApi.getById(id),
  )
}

export function useIntegrationStatsQuery(id: number) {
  return useApiQuery(id ? `integrations-${id}-stats` : null, () =>
    integrationsApi.getStats(id),
  )
}

export function useIntegrationOrdersQuery(
  id: number,
  limit = 20,
  offset = 0,
) {
  return useApiQuery(
    id ? `integrations-${id}-orders-${limit}-${offset}` : null,
    () => integrationsApi.getOrders(id, limit, offset),
    { keepPreviousData: true },
  )
}

export function useIntegrationLinkedClientsQuery(
  id: number,
  limit = 20,
  offset = 0,
) {
  return useApiQuery(
    id ? `integrations-${id}-linked-clients-${limit}-${offset}` : null,
    () => integrationsApi.getLinkedClients(id, limit, offset),
    { keepPreviousData: true },
  )
}

export function useIntegrationMenuQuery(id: number) {
  return useApiQuery(id ? `integrations-${id}-menu` : null, () =>
    integrationsApi.getMenu(id),
  )
}

export function useCreateIntegrationMutation() {
  return useApiMutation(
    'integrations/create',
    (data: CreateIntegrationRequest) => integrationsApi.create(data),
    ['integrations'],
  )
}

export function useDiscoverIntegrationMutation() {
  return useApiMutation(
    'integrations/discover',
    (data: CreateIntegrationRequest) => integrationsApi.discover(data),
    [],
  )
}

export function useUpdateIntegrationMutation() {
  return useApiMutation(
    'integrations/update',
    ({ id, data }: { id: number; data: UpdateIntegrationRequest }) =>
      integrationsApi.update(id, data),
    ['integrations'],
  )
}

export function useDeleteIntegrationMutation() {
  return useApiMutation(
    'integrations/delete',
    (id: number) => integrationsApi.remove(id),
    ['integrations'],
  )
}

export function useSyncIntegrationMutation() {
  return useApiMutation(
    'integrations/sync',
    (id: number) => integrationsApi.sync(id),
    ['integrations'],
  )
}

export function useTestConnectionMutation() {
  return useApiMutation(
    'integrations/test',
    (id: number) => integrationsApi.testConnection(id),
    [],
  )
}


