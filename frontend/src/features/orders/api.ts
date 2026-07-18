import { api } from '@/lib/api'
import type { Order, OrderSource } from './types'

export const ordersApi = {
  listOrgOrders: async (params?: {
    source?: OrderSource
    status?: string
  }): Promise<Order[]> => {
    const response = await api.get<Order[]>('/orders', {
      params: {
        ...(params?.source ? { source: params.source } : {}),
        ...(params?.status ? { status: params.status } : {}),
      },
    })
    // Backend returns JSON null when there are no orders.
    return response.data ?? []
  },

  listOrders: async (
    botId: number,
    params?: { source?: OrderSource; status?: string },
  ): Promise<Order[]> => {
    const response = await api.get<Order[]>(`/orders/bots/${botId}`, {
      params: {
        ...(params?.source ? { source: params.source } : {}),
        ...(params?.status ? { status: params.status } : {}),
      },
    })
    // Backend returns JSON null when there are no orders.
    return response.data ?? []
  },

  updateOrderStatus: async (orderId: number, status: string): Promise<void> => {
    await api.patch(`/orders/${orderId}`, { status })
  },
}
