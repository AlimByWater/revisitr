import { useApiQuery } from '@/lib/swr'
import { lunchApi } from './api'

export function useLunchProgramQuery(botId: number) {
  return useApiQuery(botId ? `lunch-program-${botId}` : null, () => lunchApi.getProgram(botId))
}

export function useLunchOrdersQuery(botId: number, status?: string) {
  return useApiQuery(
    botId ? `lunch-orders-${botId}-${status ?? 'all'}` : null,
    () => lunchApi.listOrders(botId, status),
    { refreshInterval: 15000 },
  )
}
