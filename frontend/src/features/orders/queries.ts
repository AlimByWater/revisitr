import { useApiQuery } from '@/lib/swr'
import { ordersApi } from './api'
import type { OrderSource } from './types'

export function useOrdersQuery(botId: number, source?: OrderSource, status?: string) {
  return useApiQuery(
    botId ? `orders-${botId}-${source ?? 'all'}-${status ?? 'all'}` : null,
    () => ordersApi.listOrders(botId, { source, status }),
    { refreshInterval: 15000 },
  )
}

export function useOrgOrdersQuery(source?: OrderSource, status?: string) {
  return useApiQuery(
    `org-orders-${source ?? 'all'}-${status ?? 'all'}`,
    () => ordersApi.listOrgOrders({ source, status }),
    { refreshInterval: 15000 },
  )
}
