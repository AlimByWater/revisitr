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
  pending: { label: 'Ожидает', className: 'bg-yellow-100 text-yellow-700' },
  confirmed: { label: 'Подтверждён', className: 'bg-blue-100 text-blue-700' },
  completed: { label: 'Выполнен', className: 'bg-green-100 text-green-700' },
  cancelled: { label: 'Отменён', className: 'bg-red-100 text-red-700' },
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
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Маркетплейс</h1>
          <p className="font-mono text-xs text-neutral-400 uppercase tracking-wider mt-1">
            Товары и заказы за баллы лояльности
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          Новый товар
        </button>
      </div>

      {/* Stats */}
      {stats && (
        <div className="grid gap-4 sm:grid-cols-4">
          <div className="rounded-lg border bg-card p-4">
            <p className="text-sm text-muted-foreground">Товаров</p>
            <p className="text-2xl font-bold tabular-nums">{stats.total_products}</p>
          </div>
          <div className="rounded-lg border bg-card p-4">
            <p className="text-sm text-muted-foreground">Активных</p>
            <p className="text-2xl font-bold tabular-nums">{stats.active_products}</p>
          </div>
          <div className="rounded-lg border bg-card p-4">
            <p className="text-sm text-muted-foreground">Заказов</p>
            <p className="text-2xl font-bold tabular-nums">{stats.total_orders}</p>
          </div>
          <div className="rounded-lg border bg-card p-4">
            <p className="text-sm text-muted-foreground">Потрачено баллов</p>
            <p className="text-2xl font-bold tabular-nums">{stats.total_spent_points.toLocaleString()}</p>
          </div>
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b">
        <button
          onClick={() => setTab('products')}
          className={cn(
            'px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
            tab === 'products' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          Товары
        </button>
        <button
          onClick={() => setTab('orders')}
          className={cn(
            'px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
            tab === 'orders' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          Заказы
        </button>
      </div>

      {/* Create form */}
      {showCreate && (
        <div className="rounded-lg border bg-card p-4 space-y-3">
          <h3 className="font-medium">Новый товар</h3>
          <input
            type="text"
            placeholder="Название"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            className="w-full rounded-md border px-3 py-2 text-sm"
          />
          <textarea
            placeholder="Описание"
            value={newDesc}
            onChange={(e) => setNewDesc(e.target.value)}
            rows={2}
            className="w-full rounded-md border px-3 py-2 text-sm"
          />
          <div className="grid grid-cols-2 gap-3">
            <input
              type="number"
              placeholder="Цена в баллах"
              value={newPrice}
              onChange={(e) => setNewPrice(e.target.value)}
              className="w-full rounded-md border px-3 py-2 text-sm"
            />
            <input
              type="number"
              placeholder="Запас (пусто = безлимит)"
              value={newStock}
              onChange={(e) => setNewStock(e.target.value)}
              className="w-full rounded-md border px-3 py-2 text-sm"
            />
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleCreate}
              disabled={!newName || !newPrice}
              className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              Создать
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
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
              <div key={product.id} className="rounded-lg border bg-card p-4 space-y-2">
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-medium">{product.name}</h3>
                    {product.description && (
                      <p className="text-sm text-muted-foreground line-clamp-2">{product.description}</p>
                    )}
                  </div>
                  <span className={cn(
                    'inline-block h-2 w-2 rounded-full mt-1.5',
                    product.is_active ? 'bg-green-500' : 'bg-gray-300',
                  )} />
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="font-semibold tabular-nums">{product.price_points} баллов</span>
                  <span className="text-muted-foreground">
                    {product.stock === null ? 'Безлимит' : `Остаток: ${product.stock}`}
                  </span>
                </div>
                <button
                  onClick={() => deleteProduct.trigger(product.id)}
                  className="inline-flex items-center gap-1 rounded-md border border-red-200 px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50"
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
          <div className="rounded-lg border overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-muted/50">
                <tr>
                  <th className="text-left px-4 py-2 font-medium">#</th>
                  <th className="text-left px-4 py-2 font-medium">Клиент</th>
                  <th className="text-left px-4 py-2 font-medium">Товары</th>
                  <th className="text-left px-4 py-2 font-medium">Баллы</th>
                  <th className="text-left px-4 py-2 font-medium">Статус</th>
                  <th className="text-right px-4 py-2 font-medium">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {orders.map((order) => {
                  const st = statusConfig[order.status] || statusConfig.pending
                  return (
                    <tr key={order.id} className="hover:bg-muted/30">
                      <td className="px-4 py-2 tabular-nums">{order.id}</td>
                      <td className="px-4 py-2 tabular-nums">#{order.client_id}</td>
                      <td className="px-4 py-2">
                        {order.items.map((i) => `${i.product_name} x${i.quantity}`).join(', ')}
                      </td>
                      <td className="px-4 py-2 tabular-nums font-medium">{order.total_points}</td>
                      <td className="px-4 py-2">
                        <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', st.className)}>
                          {st.label}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-right">
                        <div className="flex gap-1 justify-end">
                          {order.status === 'confirmed' && (
                            <button
                              onClick={() => updateStatus.trigger({ id: order.id, status: 'completed' })}
                              className="inline-flex items-center gap-1 rounded-md border px-2 py-1 text-xs font-medium text-green-600 hover:bg-green-50"
                            >
                              <CheckCircle className="h-3 w-3" />
                              Выполнен
                            </button>
                          )}
                          {(order.status === 'pending' || order.status === 'confirmed') && (
                            <button
                              onClick={() => updateStatus.trigger({ id: order.id, status: 'cancelled' })}
                              className="inline-flex items-center gap-1 rounded-md border border-red-200 px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50"
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
        )
      )}
    </div>
  )
}
