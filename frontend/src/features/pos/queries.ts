import { useApiQuery, useApiMutation } from '../../lib/swr'
import { list, getById, create, update, remove } from './api'
import type { CreatePOSRequest, POSLocation } from './types'

export function usePOSQuery() {
  return useApiQuery('pos-locations', list)
}

export function usePOSDetailQuery(id: number) {
  return useApiQuery(id ? `pos-locations-${id}` : null, () => getById(id))
}

export function useCreatePOSMutation() {
  return useApiMutation(
    'pos-locations/create',
    (data: CreatePOSRequest) => create(data),
    ['pos-locations'],
  )
}

export function useUpdatePOSMutation() {
  return useApiMutation(
    'pos-locations/update',
    ({ id, data }: { id: number; data: Partial<POSLocation> }) =>
      update(id, data),
    ['pos-locations'],
  )
}

export function useDeletePOSMutation() {
  return useApiMutation(
    'pos-locations/delete',
    (id: number) => remove(id),
    ['pos-locations'],
  )
}
