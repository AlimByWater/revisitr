import { useState } from 'react'
import { cn } from '@/lib/utils'
import { ordersApi } from '@/features/orders/api'
import {
  ORDER_SOURCE_LABELS,
  ORDER_STATUS_LABELS,
  ORDER_STATUS_STYLES,
} from '@/features/orders/labels'
import { useOrgOrdersQuery } from '@/features/orders/queries'
import type { OrderSource } from '@/features/orders/types'
import { ErrorState } from '@/components/common/ErrorState'

const priceFormatter = new Intl.NumberFormat('ru-RU', {
  style: 'currency',
  currency: 'RUB',
  maximumFractionDigits: 0,
})

const SOURCE_FILTERS: { value: OrderSource | undefined; label: string }[] = [
  { value: undefined, label: 'Все источники' },
  { value: 'lunch', label: ORDER_SOURCE_LABELS.lunch },
  { value: 'menu', label: ORDER_SOURCE_LABELS.menu },
]

export default function OrdersPage() {
  const [showAll, setShowAll] = useState(false)
  const [source, setSource] = useState<OrderSource | undefined>(undefined)
  const { data: orders = [], isError, mutate } = useOrgOrdersQuery(source, showAll ? undefined : 'new')
  const [busyOrderId, setBusyOrderId] = useState<number | null>(null)

  const changeStatus = async (orderId: number, status: string) => {
    setBusyOrderId(orderId)
    try {
      await ordersApi.updateOrderStatus(orderId, status)
      await mutate()
    } finally {
      setBusyOrderId(null)
    }
  }

  return (
    <div className="max-w-4xl">
      <div className="animate-in mb-6">
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Заказы</h1>
        <p className="mt-2 max-w-2xl text-sm text-neutral-500">
          Все заказы гостей через ботов — из любых модулей. Список обновляется автоматически.
        </p>
      </div>

      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-1">
          {SOURCE_FILTERS.map((filter) => (
            <button
              key={filter.label}
              type="button"
              onClick={() => setSource(filter.value)}
              className={cn(
                'inline-flex min-h-11 items-center rounded px-3 text-sm font-medium transition-colors',
                source === filter.value
                  ? 'bg-neutral-900 text-white'
                  : 'text-neutral-500 hover:bg-neutral-100 hover:text-neutral-700',
              )}
            >
              {filter.label}
            </button>
          ))}
        </div>
        <button
          type="button"
          onClick={() => setShowAll((current) => !current)}
          className="inline-flex min-h-11 shrink-0 items-center rounded text-sm font-medium text-accent hover:text-accent/80 transition-colors"
        >
          {showAll ? 'Только новые' : 'Показать все'}
        </button>
      </div>

      <div className="space-y-2">
        {isError && (
          <ErrorState
            title="Не удалось загрузить заказы"
            message="Проверьте подключение к серверу и попробуйте снова."
          />
        )}
        {!isError && orders.length === 0 && (
          <p className="text-sm text-neutral-500">{showAll ? 'Заказов пока нет.' : 'Новых заказов нет.'}</p>
        )}
        {orders.map((order) => (
          <div key={order.id} className="rounded border border-neutral-200 bg-white px-3 py-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm font-medium text-neutral-900">
                    №{order.id} · Стол {order.table_num} · {priceFormatter.format(order.total_price)}
                  </span>
                  <span className="rounded bg-neutral-100 px-1.5 py-0.5 text-xs font-medium text-neutral-600">
                    {ORDER_SOURCE_LABELS[order.source] ?? order.source}
                  </span>
                  <span className={cn('rounded px-1.5 py-0.5 text-xs font-medium', ORDER_STATUS_STYLES[order.status])}>
                    {ORDER_STATUS_LABELS[order.status]}
                  </span>
                </div>
                <div className="mt-1 text-xs text-neutral-500">
                  {[order.bot_name, order.format_name, new Date(order.created_at).toLocaleString('ru-RU')]
                    .filter(Boolean)
                    .join(' · ')}
                </div>
                <div className="mt-1 text-xs text-neutral-500">
                  {order.items
                    .map((item) => (item.course_title ? `${item.course_title}: ${item.item_name}` : item.item_name))
                    .join(' · ')}
                </div>
              </div>
              {order.status === 'new' && (
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    disabled={busyOrderId === order.id}
                    onClick={() => changeStatus(order.id, 'sent')}
                    className={cn(
                      'inline-flex min-h-11 items-center rounded px-3 text-sm font-medium',
                      'bg-accent text-white hover:bg-accent/90 transition-colors',
                      'disabled:cursor-not-allowed disabled:opacity-50',
                    )}
                  >
                    Отработан
                  </button>
                  <button
                    type="button"
                    disabled={busyOrderId === order.id}
                    onClick={() => changeStatus(order.id, 'cancelled')}
                    className="inline-flex min-h-11 items-center rounded px-3 text-sm text-neutral-500 hover:bg-red-50 hover:text-red-600 transition-colors disabled:opacity-50"
                  >
                    Отменить
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
