import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clientsApi } from './api'
import type { ClientFilter, UpdateClientRequest } from './types'

export function useClientsQuery(filter: ClientFilter) {
  return useQuery({
    queryKey: ['clients', filter],
    queryFn: () => clientsApi.list(filter),
    placeholderData: (prev) => prev,
  })
}

export function useClientProfileQuery(id: number) {
  return useQuery({
    queryKey: ['clients', id],
    queryFn: () => clientsApi.getById(id),
    enabled: !!id,
  })
}

export function useClientStatsQuery() {
  return useQuery({
    queryKey: ['clients', 'stats'],
    queryFn: clientsApi.getStats,
  })
}

export function useUpdateTagsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateClientRequest }) =>
      clientsApi.updateTags(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['clients', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['clients'] })
    },
  })
}
