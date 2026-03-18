import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { botsApi } from './api'
import type { CreateBotRequest, Bot } from './types'

export function useBotsQuery() {
  return useQuery({
    queryKey: ['bots'],
    queryFn: botsApi.list,
  })
}

export function useBotQuery(id: number) {
  return useQuery({
    queryKey: ['bots', id],
    queryFn: () => botsApi.getById(id),
    enabled: !!id,
  })
}

export function useCreateBotMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateBotRequest) => botsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bots'] })
    },
  })
}

export function useUpdateBotMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Bot> }) =>
      botsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bots'] })
    },
  })
}

export function useDeleteBotMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => botsApi.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bots'] })
    },
  })
}
