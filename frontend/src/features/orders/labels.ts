import type { Order, OrderSource } from './types'

export const ORDER_STATUS_LABELS: Record<Order['status'], string> = {
  new: 'Новый',
  sent: 'Отработан',
  cancelled: 'Отменён',
}

export const ORDER_STATUS_STYLES: Record<Order['status'], string> = {
  new: 'bg-accent/10 text-accent',
  sent: 'bg-green-100 text-green-700',
  cancelled: 'bg-neutral-100 text-neutral-500',
}

export const ORDER_SOURCE_LABELS: Record<OrderSource, string> = {
  lunch: 'Ланч',
  menu: 'Меню',
}
