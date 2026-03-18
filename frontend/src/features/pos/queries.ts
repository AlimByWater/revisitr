import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { list, getById, create, update, remove } from './api'
import type { CreatePOSRequest, POSLocation } from './types'

export function usePOSQuery() {
  return useQuery({
    queryKey: ['pos-locations'],
    queryFn: list,
  })
}

export function usePOSDetailQuery(id: number) {
  return useQuery({
    queryKey: ['pos-locations', id],
    queryFn: () => getById(id),
    enabled: !!id,
  })
}

export function useCreatePOSMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreatePOSRequest) => create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pos-locations'] })
    },
  })
}

export function useUpdatePOSMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<POSLocation> }) =>
      update(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['pos-locations'] })
      queryClient.invalidateQueries({
        queryKey: ['pos-locations', variables.id],
      })
    },
  })
}

export function useDeletePOSMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pos-locations'] })
    },
  })
}
