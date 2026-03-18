import { useApiQuery, useApiMutation } from '../../lib/swr'
import { botsApi } from './api'
import type { CreateBotRequest, Bot } from './types'

export function useBotsQuery() {
  return useApiQuery('bots', botsApi.list)
}

export function useBotQuery(id: number) {
  return useApiQuery(id ? `bots-${id}` : null, () => botsApi.getById(id))
}

export function useCreateBotMutation() {
  return useApiMutation(
    'bots/create',
    (data: CreateBotRequest) => botsApi.create(data),
    ['bots'],
  )
}

export function useUpdateBotMutation() {
  return useApiMutation(
    'bots/update',
    ({ id, data }: { id: number; data: Partial<Bot> }) =>
      botsApi.update(id, data),
    ['bots'],
  )
}

export function useDeleteBotMutation() {
  return useApiMutation(
    'bots/delete',
    (id: number) => botsApi.remove(id),
    ['bots'],
  )
}
