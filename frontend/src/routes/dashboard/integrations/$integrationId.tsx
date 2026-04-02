import { useParams, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import {
  ArrowLeft,
  RefreshCw,
  Wifi,
  Trash2,
  BarChart3,
  ShoppingCart,
  Users,
  UtensilsCrossed,
  Settings,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useIntegrationQuery,
  useIntegrationStatsQuery,
  useIntegrationOrdersQuery,
  useIntegrationCustomersQuery,
  useIntegrationMenuQuery,
  useSyncIntegrationMutation,
  useTestConnectionMutation,
  useDeleteIntegrationMutation,
  useUpdateIntegrationMutation,
} from '@/features/integrations/queries'
import { ErrorState } from '@/components/common/ErrorState'
import type { Integration } from '@/features/integrations/types'

const TYPE_LABELS: Record<string, string> = {
  iiko: 'iiko',
  rkeeper: 'r-keeper',
  '1c': '1C',
  mock: 'Mock',
}

const STATUS_STYLES: Record<
  string,
  { bg: string; text: string; label: string }
> = {
  active: { bg: 'bg-green-50', text: 'text-green-700', label: 'Активна' },
  inactive: {
    bg: 'bg-neutral-100',
    text: 'text-neutral-500',
    label: 'Неактивна',
  },
  error: { bg: 'bg-red-50', text: 'text-red-700', label: 'Ошибка' },
}

const TABS = [
  { id: 'overview', label: 'Обзор', icon: BarChart3 },
  { id: 'orders', label: 'Заказы', icon: ShoppingCart },
  { id: 'customers', label: 'Клиенты', icon: Users },
  { id: 'menu', label: 'Меню', icon: UtensilsCrossed },
  { id: 'settings', label: 'Настройки', icon: Settings },
] as const

type TabId = (typeof TABS)[number]['id']

function formatDate(dateStr?: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(amount)
}

export default function IntegrationDetailPage() {
  const { integrationId } = useParams()
  const navigate = useNavigate()
  const id = Number(integrationId)
  const [activeTab, setActiveTab] = useState<TabId>('overview')

  const {
    data: integration,
    isLoading,
    isError,
    mutate,
  } = useIntegrationQuery(id)

  if (isLoading) {
    return (
      <div className="animate-pulse space-y-4">
        <div className="h-8 w-48 bg-neutral-200 rounded" />
        <div className="h-64 bg-neutral-100 rounded" />
      </div>
    )
  }

  if (isError || !integration) {
    return (
      <ErrorState
        title="Интеграция не найдена"
        message="Проверьте URL или вернитесь к списку интеграций."
        onRetry={() => mutate()}
      />
    )
  }

  const status = STATUS_STYLES[integration.status] || STATUS_STYLES.inactive

  return (
    <div>
      {/* Header */}
      <div className="flex items-center gap-3 mb-6 animate-in">
        <button
          type="button"
          onClick={() => navigate('/dashboard/integrations')}
          className="p-2 rounded text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h1 className="font-serif text-2xl font-bold text-neutral-900">
              {TYPE_LABELS[integration.type] || integration.type}
            </h1>
            {integration.type === 'mock' && (
              <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-orange-100 text-orange-700">
                DEV
              </span>
            )}
            <span
              className={cn(
                'text-xs font-medium px-2 py-0.5 rounded-full',
                status.bg,
                status.text,
              )}
            >
              {status.label}
            </span>
          </div>
          <p className="text-sm text-neutral-500 mt-0.5">
            Последняя синхронизация: {formatDate(integration.last_sync_at)}
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-neutral-200 overflow-x-auto animate-in animate-in-delay-1">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex items-center gap-1.5 px-3 py-2.5 text-sm font-medium whitespace-nowrap',
              'border-b-2 transition-colors',
              activeTab === tab.id
                ? 'border-accent text-accent'
                : 'border-transparent text-neutral-500 hover:text-neutral-700',
            )}
          >
            <tab.icon className="w-4 h-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="animate-in animate-in-delay-2">
        {activeTab === 'overview' && (
          <OverviewTab integrationId={id} integration={integration} />
        )}
        {activeTab === 'orders' && <OrdersTab integrationId={id} />}
        {activeTab === 'customers' && <CustomersTab integrationId={id} />}
        {activeTab === 'menu' && <MenuTab integrationId={id} />}
        {activeTab === 'settings' && (
          <SettingsTab integration={integration} onDeleted={() => navigate('/dashboard/integrations')} />
        )}
      </div>
    </div>
  )
}

// --- Overview Tab ---
function OverviewTab({
  integrationId,
  integration,
}: {
  integrationId: number
  integration: { type: string; status: string }
}) {
  const { data: stats } = useIntegrationStatsQuery(integrationId)
  const syncMutation = useSyncIntegrationMutation()
  const testMutation = useTestConnectionMutation()

  return (
    <div className="space-y-6">
      {/* Action buttons */}
      <div className="flex gap-3">
        <button
          type="button"
          onClick={() => syncMutation.mutate(integrationId)}
          disabled={syncMutation.isPending}
          className={cn(
            'flex items-center gap-2 py-2 px-4 rounded text-sm font-medium',
            'bg-accent text-white hover:bg-accent-hover',
            'disabled:opacity-50 transition-all',
          )}
        >
          <RefreshCw
            className={cn('w-4 h-4', syncMutation.isPending && 'animate-spin')}
          />
          {syncMutation.isPending ? 'Синхронизация...' : 'Синхронизировать'}
        </button>
        <button
          type="button"
          onClick={() => testMutation.mutate(integrationId)}
          disabled={testMutation.isPending}
          className={cn(
            'flex items-center gap-2 py-2 px-4 rounded text-sm font-medium',
            'border border-neutral-900 text-neutral-700',
            'hover:bg-neutral-50 disabled:opacity-50 transition-all',
          )}
        >
          <Wifi className="w-4 h-4" />
          {testMutation.isPending
            ? 'Проверка...'
            : testMutation.isSuccess
              ? 'Подключено'
              : 'Проверить связь'}
        </button>
      </div>

      {(syncMutation.isError || testMutation.isError) && (
        <p className="text-sm text-red-600 bg-red-50 p-3 rounded">
          {integration.type === 'iiko' || integration.type === 'rkeeper'
            ? 'Ошибка подключения к POS-системе. Проверьте настройки.'
            : 'Произошла ошибка. Попробуйте снова.'}
        </p>
      )}

      {syncMutation.isSuccess && (
        <p className="text-sm text-green-700 bg-green-50 p-3 rounded">
          Синхронизация завершена успешно.
        </p>
      )}

      {/* Stats */}
      {stats && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <StatCard label="Заказов" value={stats.total_orders} />
          <StatCard label="Выручка" value={formatCurrency(stats.total_revenue)} />
          <StatCard label="Клиентов" value={stats.matched_clients} />
          <StatCard label="Без привязки" value={stats.unmatched_orders} />
        </div>
      )}
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="bg-white rounded border border-neutral-900 p-4">
      <p className="text-xs text-neutral-500 mb-1">{label}</p>
      <p className="text-xl font-semibold text-neutral-900">{value}</p>
    </div>
  )
}

// --- Orders Tab ---
function OrdersTab({ integrationId }: { integrationId: number }) {
  const [page, setPage] = useState(0)
  const limit = 20
  const { data, isLoading } = useIntegrationOrdersQuery(
    integrationId,
    limit,
    page * limit,
  )

  if (isLoading) {
    return <div className="text-sm text-neutral-500">Загрузка заказов...</div>
  }

  const orders = data?.items || []
  const total = data?.total || 0

  if (orders.length === 0) {
    return (
      <p className="text-sm text-neutral-500 py-8 text-center">
        Заказы пока не синхронизированы. Нажмите &ldquo;Синхронизировать&rdquo; на
        вкладке Обзор.
      </p>
    )
  }

  return (
    <div className="space-y-4">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-neutral-200">
              <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                ID
              </th>
              <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                Дата
              </th>
              <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                Позиции
              </th>
              <th className="text-right py-2 px-3 text-neutral-500 font-medium">
                Сумма
              </th>
              <th className="text-center py-2 px-3 text-neutral-500 font-medium">
                Клиент
              </th>
            </tr>
          </thead>
          <tbody>
            {orders.map((order) => (
              <tr
                key={order.id}
                className="border-b border-neutral-200/50 hover:bg-neutral-50 transition-colors"
              >
                <td className="py-2.5 px-3 text-neutral-600 font-mono text-xs">
                  {order.external_id}
                </td>
                <td className="py-2.5 px-3 text-neutral-700">
                  {formatDate(order.ordered_at)}
                </td>
                <td className="py-2.5 px-3 text-neutral-600">
                  {order.items?.map((it) => it.name).join(', ') || '—'}
                </td>
                <td className="py-2.5 px-3 text-right font-medium text-neutral-900">
                  {formatCurrency(order.total)}
                </td>
                <td className="py-2.5 px-3 text-center">
                  {order.client_id ? (
                    <span className="text-green-600 text-xs font-medium">
                      Привязан
                    </span>
                  ) : (
                    <span className="text-neutral-400 text-xs">—</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {total > limit && (
        <div className="flex justify-center gap-2">
          <button
            type="button"
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1.5 text-sm rounded border border-neutral-900 disabled:opacity-30 hover:bg-neutral-50 transition-colors"
          >
            Назад
          </button>
          <span className="px-3 py-1.5 text-sm text-neutral-500">
            {page * limit + 1}–{Math.min((page + 1) * limit, total)} из {total}
          </span>
          <button
            type="button"
            onClick={() => setPage((p) => p + 1)}
            disabled={(page + 1) * limit >= total}
            className="px-3 py-1.5 text-sm rounded border border-neutral-900 disabled:opacity-30 hover:bg-neutral-50 transition-colors"
          >
            Далее
          </button>
        </div>
      )}
    </div>
  )
}

// --- Customers Tab ---
function CustomersTab({ integrationId }: { integrationId: number }) {
  const [search, setSearch] = useState('')
  const { data: customers, isLoading } = useIntegrationCustomersQuery(
    integrationId,
    50,
    0,
    search,
  )

  return (
    <div className="space-y-4">
      <input
        type="text"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Поиск по имени или телефону..."
        className={cn(
          'w-full rounded border border-neutral-900 px-3 py-2',
          'text-sm placeholder:text-neutral-400',
          'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
        )}
      />

      {isLoading ? (
        <p className="text-sm text-neutral-500">Загрузка клиентов...</p>
      ) : !customers || customers.length === 0 ? (
        <p className="text-sm text-neutral-500 py-8 text-center">
          {search ? 'Ничего не найдено.' : 'Нет клиентов в POS-системе.'}
        </p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-200">
                <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                  Имя
                </th>
                <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                  Телефон
                </th>
                <th className="text-right py-2 px-3 text-neutral-500 font-medium">
                  Баланс
                </th>
                <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                  Карта
                </th>
              </tr>
            </thead>
            <tbody>
              {customers.map((c) => (
                <tr
                  key={c.external_id}
                  className="border-b border-neutral-200/50 hover:bg-neutral-50 transition-colors"
                >
                  <td className="py-2.5 px-3 text-neutral-900 font-medium">
                    {c.name}
                  </td>
                  <td className="py-2.5 px-3 text-neutral-600">{c.phone}</td>
                  <td className="py-2.5 px-3 text-right text-neutral-700">
                    {formatCurrency(c.balance)}
                  </td>
                  <td className="py-2.5 px-3 text-neutral-400 text-xs font-mono">
                    {c.card_number || '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

// --- Menu Tab ---
function MenuTab({ integrationId }: { integrationId: number }) {
  const { data: menu, isLoading } = useIntegrationMenuQuery(integrationId)

  if (isLoading) {
    return <p className="text-sm text-neutral-500">Загрузка меню...</p>
  }

  if (!menu || menu.categories.length === 0) {
    return (
      <p className="text-sm text-neutral-500 py-8 text-center">
        Меню не загружено из POS-системы.
      </p>
    )
  }

  return (
    <div className="space-y-6">
      {menu.categories.map((cat) => (
        <div key={cat.name}>
          <h3 className="text-sm font-semibold text-neutral-900 mb-2">
            {cat.name}
          </h3>
          <div className="bg-white rounded border border-neutral-900 divide-y divide-neutral-200/50">
            {cat.items.map((item) => (
              <div
                key={item.external_id}
                className="flex items-center justify-between px-4 py-2.5"
              >
                <div>
                  <span className="text-sm text-neutral-800">{item.name}</span>
                  {item.description && (
                    <p className="text-xs text-neutral-400 mt-0.5">
                      {item.description}
                    </p>
                  )}
                </div>
                <span className="text-sm font-medium text-neutral-900 ml-4 whitespace-nowrap">
                  {formatCurrency(item.price)}
                </span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

// --- Settings Tab ---
function SettingsTab({
  integration,
  onDeleted,
}: {
  integration: Integration
  onDeleted: () => void
}) {
  const deleteMutation = useDeleteIntegrationMutation()
  const updateMutation = useUpdateIntegrationMutation()
  const [confirmDelete, setConfirmDelete] = useState(false)

  async function handleStatusToggle() {
    const newStatus = integration.status === 'active' ? 'inactive' : 'active'
    await updateMutation.mutateAsync({
      id: integration.id,
      data: { status: newStatus },
    })
  }

  async function handleDelete() {
    await deleteMutation.mutateAsync(integration.id)
    onDeleted()
  }

  return (
    <div className="space-y-6 max-w-lg">
      {/* Config display */}
      <div className="bg-white rounded border border-neutral-900 p-4 space-y-3">
        <h3 className="text-sm font-semibold text-neutral-900">Конфигурация</h3>
        {integration.config.api_url && (
          <div>
            <p className="text-xs text-neutral-500">API URL</p>
            <p className="text-sm text-neutral-700 font-mono break-all">
              {integration.config.api_url}
            </p>
          </div>
        )}
        {integration.config.api_key && (
          <div>
            <p className="text-xs text-neutral-500">API Key</p>
            <p className="text-sm text-neutral-700 font-mono">{'*'.repeat(16)}</p>
          </div>
        )}
        {integration.type === 'mock' && (
          <p className="text-sm text-neutral-500">
            Mock-интеграция не требует настройки.
          </p>
        )}
      </div>

      {/* Status toggle */}
      <div className="flex items-center justify-between bg-white rounded border border-neutral-900 p-4">
        <div>
          <p className="text-sm font-medium text-neutral-900">Статус</p>
          <p className="text-xs text-neutral-500">
            {integration.status === 'active'
              ? 'Интеграция активна и синхронизируется'
              : 'Интеграция приостановлена'}
          </p>
        </div>
        <button
          type="button"
          onClick={handleStatusToggle}
          disabled={updateMutation.isPending}
          className={cn(
            'px-3 py-1.5 text-sm font-medium rounded transition-colors',
            integration.status === 'active'
              ? 'text-neutral-700 border border-neutral-900 hover:bg-neutral-50'
              : 'bg-accent text-white hover:bg-accent-hover',
          )}
        >
          {integration.status === 'active' ? 'Приостановить' : 'Активировать'}
        </button>
      </div>

      {/* Delete */}
      <div className="bg-white rounded border border-red-200 p-4">
        <h3 className="text-sm font-semibold text-red-700 mb-1">
          Удаление интеграции
        </h3>
        <p className="text-xs text-neutral-500 mb-3">
          Все синхронизированные данные (заказы, связи с клиентами) будут удалены.
        </p>
        {!confirmDelete ? (
          <button
            type="button"
            onClick={() => setConfirmDelete(true)}
            className="flex items-center gap-1.5 text-sm text-red-600 hover:text-red-700 transition-colors"
          >
            <Trash2 className="w-4 h-4" />
            Удалить интеграцию
          </button>
        ) : (
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
              className="px-3 py-1.5 text-sm font-medium rounded bg-red-600 text-white hover:bg-red-700 disabled:opacity-50 transition-colors"
            >
              {deleteMutation.isPending ? 'Удаление...' : 'Подтвердить'}
            </button>
            <button
              type="button"
              onClick={() => setConfirmDelete(false)}
              className="px-3 py-1.5 text-sm text-neutral-600 hover:text-neutral-800 transition-colors"
            >
              Отмена
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
