import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  useMarketplaceProductsQuery,
  useMarketplaceOrdersQuery,
  useMarketplaceStatsQuery,
  useCreateProductMutation,
  useDeleteProductMutation,
  useUpdateOrderStatusMutation,
} from '@/features/marketplace/queries'
import { ShoppingBag, Plus, Package, Trash2, CheckCircle, XCircle } from 'lucide-react'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'

const statusConfig: Record<string, { label: string; className: string }> = {
  pending: { label: 'Ожидает', className: 'bg-amber-500/10 text-amber-700 border border-amber-500/30' },
  confirmed: { label: 'Подтверждён', className: 'bg-accent/10 text-accent border border-accent/30' },
  completed: { label: 'Выполнен', className: 'bg-emerald-500/10 text-emerald-700 border border-emerald-500/30' },
  cancelled: { label: 'Отменён', className: 'bg-red-500/10 text-red-700 border border-red-500/30' },
}

export default function MarketplacePage() {
  const { data: products, isLoading: productsLoading, error: productsError } = useMarketplaceProductsQuery()
  const { data: orders, isLoading: ordersLoading } = useMarketplaceOrdersQuery()
  const { data: stats } = useMarketplaceStatsQuery()
  const createProduct = useCreateProductMutation()
  const deleteProduct = useDeleteProductMutation()
  const updateStatus = useUpdateOrderStatusMutation()

  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [newPrice, setNewPrice] = useState('')
  const [newStock, setNewStock] = useState('')
  const [tab, setTab] = useState<'products' | 'orders'>('products')

  if (productsError) return <ErrorState message="Ошибка загрузки маркетплейса" />

  const handleCreate = async () => {
    if (!newName || !newPrice) return
    await createProduct.trigger({
      name: newName,
      description: newDesc,
      price_points: parseInt(newPrice, 10),
      stock: newStock ? parseInt(newStock, 10) : undefined,
    })
    setShowCreate(false)
    setNewName('')
    setNewDesc('')
    setNewPrice('')
    setNewStock('')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="animate-in">
          <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Revisitr</p>
          <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Маркетплейс</h1>
          <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">
            Товары и заказы за баллы лояльности
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 rounded bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover transition-colors"
        >
          <Plus className="h-4 w-4" />
          Новый товар
        </button>
      </div>

      {/* Stats */}
      {stats && (
        <div className="grid gap-3 sm:grid-cols-4">
          <div className="rounded border border-neutral-900 bg-white p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-neutral-400 mb-2">Товаров</p>
            <p className="text-2xl font-bold font-mono tabular-nums text-neutral-900">{stats.total_products}</p>
          </div>
          <div className="rounded border border-neutral-900 bg-white p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-neutral-400 mb-2">Активных</p>
            <p className="text-2xl font-bold font-mono tabular-nums text-neutral-900">{stats.active_products}</p>
          </div>
          <div className="rounded border border-neutral-900 bg-white p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-neutral-400 mb-2">Заказов</p>
            <p className="text-2xl font-bold font-mono tabular-nums text-neutral-900">{stats.total_orders}</p>
          </div>
          <div className="rounded border border-neutral-900 bg-white p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-neutral-400 mb-2">Потрачено баллов</p>
            <p className="text-2xl font-bold font-mono tabular-nums text-neutral-900">{stats.total_spent_points.toLocaleString()}</p>
          </div>
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b border-neutral-200">
        <button
          onClick={() => setTab('products')}
          className={cn(
            'px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors',
            tab === 'products' ? 'border-accent text-accent' : 'border-transparent text-neutral-500 hover:text-neutral-700',
          )}
        >
          Товары
        </button>
        <button
          onClick={() => setTab('orders')}
          className={cn(
            'px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors',
            tab === 'orders' ? 'border-accent text-accent' : 'border-transparent text-neutral-500 hover:text-neutral-700',
          )}
        >
          Заказы
        </button>
      </div>

      {/* Create form */}
      {showCreate && (
        <div className="rounded border border-neutral-900 bg-white p-5 space-y-3 animate-in">
          <h3 className="text-sm font-semibold text-neutral-900">Новый товар</h3>
          <input
            type="text"
            placeholder="Название"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            className="w-full rounded border border-neutral-200 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent"
          />
          <textarea
            placeholder="Описание"
            value={newDesc}
            onChange={(e) => setNewDesc(e.target.value)}
            rows={2}
            className="w-full rounded border border-neutral-200 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent"
          />
          <div className="grid grid-cols-2 gap-3">
            <input
              type="number"
              placeholder="Цена в баллах"
              value={newPrice}
              onChange={(e) => setNewPrice(e.target.value)}
              className="w-full rounded border border-neutral-200 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent"
            />
            <input
              type="number"
              placeholder="Запас (пусто = безлимит)"
              value={newStock}
              onChange={(e) => setNewStock(e.target.value)}
              className="w-full rounded border border-neutral-200 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent"
            />
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleCreate}
              disabled={!newName || !newPrice}
              className="rounded bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover disabled:opacity-50 transition-colors"
            >
              Создать
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="rounded border border-neutral-200 px-4 py-2 text-sm font-medium text-neutral-700 hover:bg-neutral-50 transition-colors"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {/* Products tab */}
      {tab === 'products' && (
        productsLoading ? <TableSkeleton rows={4} /> :
        !products || products.length === 0 ? (
          <EmptyState
            icon={Package}
            title="Нет товаров"
            description="Добавьте первый товар для маркетплейса"
          />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {products.map((product) => (
              <div key={product.id} className="rounded border border-neutral-900 bg-white p-5 space-y-3">
                <div className="flex items-start justify-between">
                  <div className="min-w-0">
                    <h3 className="text-sm font-semibold text-neutral-900">{product.name}</h3>
                    {product.description && (
                      <p className="text-sm text-neutral-500 line-clamp-2 mt-0.5">{product.description}</p>
                    )}
                  </div>
                  <span className={cn(
                    'inline-block h-2 w-2 rounded-full mt-1.5 shrink-0 ml-2',
                    product.is_active ? 'bg-accent' : 'bg-neutral-300',
                  )} />
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="font-mono font-semibold tabular-nums text-neutral-900">{product.price_points} баллов</span>
                  <span className="text-neutral-500 text-xs">
                    {product.stock === null ? 'Безлимит' : `Остаток: ${product.stock}`}
                  </span>
                </div>
                <button
                  onClick={() => deleteProduct.trigger(product.id)}
                  className="inline-flex items-center gap-1 rounded border border-red-200 px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 transition-colors"
                >
                  <Trash2 className="h-3 w-3" />
                  Удалить
                </button>
              </div>
            ))}
          </div>
        )
      )}

      {/* Orders tab */}
      {tab === 'orders' && (
        ordersLoading ? <TableSkeleton rows={5} /> :
        !orders || orders.length === 0 ? (
          <EmptyState
            icon={ShoppingBag}
            title="Нет заказов"
            description="Заказы появятся когда клиенты начнут покупать товары за баллы"
          />
        ) : (
          <div className="rounded border border-neutral-900 bg-white overflow-hidden">
            <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-200 bg-neutral-50/50">
                  <th className="text-left px-4 py-3 text-xs font-medium text-neutral-500 uppercase tracking-wider">#</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-neutral-500 uppercase tracking-wider">Клиент</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-neutral-500 uppercase tracking-wider">Товары</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-neutral-500 uppercase tracking-wider">Баллы</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-neutral-500 uppercase tracking-wider">Статус</th>
                  <th className="text-right px-4 py-3 text-xs font-medium text-neutral-500 uppercase tracking-wider">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-200">
                {orders.map((order) => {
                  const st = statusConfig[order.status] || statusConfig.pending
                  return (
                    <tr key={order.id} className="hover:bg-neutral-50 transition-colors">
                      <td className="px-4 py-3 tabular-nums text-neutral-700">{order.id}</td>
                      <td className="px-4 py-3 tabular-nums text-neutral-700">#{order.client_id}</td>
                      <td className="px-4 py-3 text-neutral-700">
                        {order.items.map((i) => `${i.product_name} x${i.quantity}`).join(', ')}
                      </td>
                      <td className="px-4 py-3 tabular-nums font-mono font-medium text-neutral-900">{order.total_points}</td>
                      <td className="px-4 py-3">
                        <span className={cn('font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded', st.className)}>
                          {st.label}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex gap-1 justify-end">
                          {order.status === 'confirmed' && (
                            <button
                              onClick={() => updateStatus.trigger({ id: order.id, status: 'completed' })}
                              className="inline-flex items-center gap-1 rounded border px-2 py-1 text-xs font-medium text-emerald-700 border-emerald-200 hover:bg-emerald-50 transition-colors"
                            >
                              <CheckCircle className="h-3 w-3" />
                              Выполнен
                            </button>
                          )}
                          {(order.status === 'pending' || order.status === 'confirmed') && (
                            <button
                              onClick={() => updateStatus.trigger({ id: order.id, status: 'cancelled' })}
                              className="inline-flex items-center gap-1 rounded border border-red-200 px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 transition-colors"
                            >
                              <XCircle className="h-3 w-3" />
                              Отменить
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
            </div>
          </div>
        )
      )}
    </div>
  )
}
