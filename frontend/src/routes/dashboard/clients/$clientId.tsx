import { Link, useParams } from 'react-router-dom'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  useClientProfileQuery,
  useUpdateTagsMutation,
} from '@/features/clients/queries'
import { useClientOrderStatsQuery } from '@/features/menus/queries'
import type { LoyaltyTransaction } from '@/features/clients/types'
import type { ClientOrderStats } from '@/features/menus/types'
import {
  ArrowLeft,
  User,
  Phone,
  MapPin,
  Monitor,
  Calendar,
  Heart,
  Wallet,
  TrendingUp,
  TrendingDown,
  X,
  Plus,
  Hash,
  Receipt,
  ShoppingCart,
} from 'lucide-react'

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatBalance(value: number): string {
  return value.toLocaleString('ru-RU')
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(amount)
}

const transactionTypeConfig: Record<
  LoyaltyTransaction['type'],
  { label: string; className: string }
> = {
  earn: {
    label: 'Начисление',
    className: 'bg-green-100 text-green-700',
  },
  spend: {
    label: 'Списание',
    className: 'bg-orange-100 text-orange-700',
  },
  adjust: {
    label: 'Корректировка',
    className: 'bg-blue-100 text-blue-700',
  },
}

const genderLabels: Record<string, string> = {
  male: 'Мужской',
  female: 'Женский',
}

const TABS = [
  { id: 'profile', label: 'Профиль', icon: User },
  { id: 'transactions', label: 'Транзакции', icon: Receipt },
  { id: 'orders', label: 'Заказы POS', icon: ShoppingCart },
] as const

type TabId = (typeof TABS)[number]['id']

function InfoRow({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value?: string | null
}) {
  if (!value) return null
  return (
    <div className="flex items-center gap-3 py-2">
      <Icon className="w-4 h-4 text-neutral-400 shrink-0" />
      <span className="text-sm text-neutral-500 w-28 shrink-0">{label}</span>
      <span className="text-sm text-neutral-900">{value}</span>
    </div>
  )
}

export default function ClientDetailPage() {
  const { clientId } = useParams<{ clientId: string }>()
  const id = Number(clientId)

  const { data: client, isLoading } = useClientProfileQuery(id)
  const updateTagsMutation = useUpdateTagsMutation()

  const [activeTab, setActiveTab] = useState<TabId>('profile')
  const [newTag, setNewTag] = useState('')
  const [showTagInput, setShowTagInput] = useState(false)

  function handleAddTag() {
    if (!client || !newTag.trim()) return
    const updated = [...(client.tags ?? []), newTag.trim()]
    updateTagsMutation.mutate({ id, data: { tags: updated } })
    setNewTag('')
    setShowTagInput(false)
  }

  function handleRemoveTag(tag: string) {
    if (!client) return
    const updated = (client.tags ?? []).filter((t) => t !== tag)
    updateTagsMutation.mutate({ id, data: { tags: updated } })
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl">
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-neutral-300 border-t-neutral-900 rounded-full animate-spin" />
        </div>
      </div>
    )
  }

  if (!client) {
    return (
      <div className="max-w-4xl">
        <Link
          to="/dashboard/clients"
          className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-900 transition-colors mb-4"
        >
          <ArrowLeft className="w-4 h-4" />
          Назад к клиентам
        </Link>
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-neutral-500">Клиент не найден</p>
        </div>
      </div>
    )
  }

  const fullName = [client.first_name, client.last_name]
    .filter(Boolean)
    .join(' ')

  return (
    <div className="max-w-4xl">
      <Link
        to="/dashboard/clients"
        className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-900 transition-colors mb-4"
      >
        <ArrowLeft className="w-4 h-4" />
        Назад к клиентам
      </Link>

      <h1 className="font-serif font-serif text-3xl font-bold text-neutral-900 tracking-tight mb-6">{fullName}</h1>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-surface-border">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex items-center gap-1.5 px-3 py-2.5 text-sm font-medium border-b-2 transition-colors',
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
      {activeTab === 'profile' && (
        <ProfileTab
          client={client}
          fullName={fullName}
          newTag={newTag}
          showTagInput={showTagInput}
          onNewTagChange={setNewTag}
          onShowTagInput={setShowTagInput}
          onAddTag={handleAddTag}
          onRemoveTag={handleRemoveTag}
        />
      )}
      {activeTab === 'transactions' && (
        <TransactionsTab transactions={client.transactions} />
      )}
      {activeTab === 'orders' && (
        <OrdersTab clientId={id} />
      )}
    </div>
  )
}

// --- Profile Tab ---
function ProfileTab({
  client,
  fullName,
  newTag,
  showTagInput,
  onNewTagChange,
  onShowTagInput,
  onAddTag,
  onRemoveTag,
}: {
  client: {
    phone?: string | null
    username?: string | null
    city?: string | null
    os?: string | null
    gender?: string | null
    birth_date?: string | null
    registered_at: string
    loyalty_level?: string | null
    loyalty_balance: number
    purchase_count: number
    total_purchases: number
    tags?: string[] | null
  }
  fullName: string
  newTag: string
  showTagInput: boolean
  onNewTagChange: (v: string) => void
  onShowTagInput: (v: boolean) => void
  onAddTag: () => void
  onRemoveTag: (tag: string) => void
}) {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Info card */}
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
          <h2 className="text-base font-semibold text-neutral-900 mb-4">
            Информация
          </h2>
          <div className="divide-y divide-surface-border">
            <InfoRow icon={User} label="Имя" value={fullName} />
            <InfoRow icon={Phone} label="Телефон" value={client.phone} />
            <InfoRow
              icon={Hash}
              label="Telegram"
              value={client.username ? `@${client.username}` : undefined}
            />
            <InfoRow icon={MapPin} label="Город" value={client.city} />
            <InfoRow icon={Monitor} label="ОС" value={client.os} />
            <InfoRow
              icon={User}
              label="Пол"
              value={
                client.gender ? (genderLabels[client.gender] ?? client.gender) : undefined
              }
            />
            <InfoRow
              icon={Calendar}
              label="Дата рождения"
              value={client.birth_date ? formatDate(client.birth_date) : undefined}
            />
            <InfoRow
              icon={Calendar}
              label="Регистрация"
              value={formatDate(client.registered_at)}
            />
          </div>
        </div>

        {/* Loyalty card */}
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
          <h2 className="text-base font-semibold text-neutral-900 mb-4">
            Лояльность
          </h2>
          <div className="space-y-4">
            {client.loyalty_level && (
              <div className="flex items-center gap-2">
                <Heart className="w-4 h-4 text-neutral-400" />
                <span className="text-sm text-neutral-500">Уровень</span>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-accent/10 text-accent">
                  {client.loyalty_level}
                </span>
              </div>
            )}
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-neutral-50 rounded-xl p-4">
                <div className="flex items-center gap-2 mb-1">
                  <Wallet className="w-4 h-4 text-neutral-400" />
                  <span className="text-xs text-neutral-500">Баланс</span>
                </div>
                <p className="text-xl font-bold font-mono text-neutral-900 tabular-nums tracking-tight">
                  {formatBalance(client.loyalty_balance)}
                </p>
              </div>
              <div className="bg-neutral-50 rounded-xl p-4">
                <div className="flex items-center gap-2 mb-1">
                  <TrendingUp className="w-4 h-4 text-neutral-400" />
                  <span className="text-xs text-neutral-500">Покупок</span>
                </div>
                <p className="text-xl font-bold font-mono text-neutral-900 tabular-nums tracking-tight">
                  {client.purchase_count}
                </p>
              </div>
            </div>
            <div className="bg-neutral-50 rounded-xl p-4">
              <div className="flex items-center gap-2 mb-1">
                <TrendingDown className="w-4 h-4 text-neutral-400" />
                <span className="text-xs text-neutral-500">
                  Сумма покупок
                </span>
              </div>
              <p className="text-xl font-bold font-mono text-neutral-900 tabular-nums tracking-tight">
                {formatBalance(client.total_purchases)}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Tags */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
        <h2 className="text-base font-semibold text-neutral-900 mb-4">Теги</h2>
        <div className="flex flex-wrap items-center gap-2">
          {(client.tags ?? []).map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-neutral-100 text-neutral-700"
            >
              {tag}
              <button
                type="button"
                onClick={() => onRemoveTag(tag)}
                className="hover:text-red-600 transition-colors"
                aria-label={`Удалить тег ${tag}`}
              >
                <X className="w-3 h-3" />
              </button>
            </span>
          ))}
          {showTagInput ? (
            <div className="flex items-center gap-1">
              <input
                type="text"
                value={newTag}
                onChange={(e) => onNewTagChange(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') onAddTag()
                  if (e.key === 'Escape') {
                    onShowTagInput(false)
                    onNewTagChange('')
                  }
                }}
                placeholder="Новый тег..."
                autoFocus
                className={cn(
                  'px-2.5 py-1 rounded-lg border border-surface-border text-xs',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                )}
              />
              <button
                type="button"
                onClick={onAddTag}
                disabled={!newTag.trim()}
                className={cn(
                  'px-2 py-1 rounded-lg text-xs font-medium',
                  'bg-neutral-900 text-white hover:bg-neutral-800 transition-colors',
                  'disabled:opacity-40 disabled:cursor-not-allowed',
                )}
              >
                Добавить
              </button>
            </div>
          ) : (
            <button
              type="button"
              onClick={() => onShowTagInput(true)}
              className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium border border-dashed border-neutral-300 text-neutral-500 hover:border-neutral-400 hover:text-neutral-700 transition-colors"
            >
              <Plus className="w-3 h-3" />
              Добавить тег
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

// --- Transactions Tab ---
function TransactionsTab({
  transactions,
}: {
  transactions?: LoyaltyTransaction[] | null
}) {
  if (!transactions || transactions.length === 0) {
    return (
      <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
        <p className="text-sm text-neutral-500">Нет транзакций</p>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden">
      <div className="px-6 py-4 border-b border-surface-border">
        <h2 className="text-base font-semibold text-neutral-900">
          История транзакций
        </h2>
      </div>
      <table className="w-full">
        <thead>
          <tr className="border-b border-surface-border">
            <th className="text-left text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
              Дата
            </th>
            <th className="text-left text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
              Тип
            </th>
            <th className="text-right text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
              Сумма
            </th>
            <th className="text-right text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
              Баланс после
            </th>
            <th className="text-left text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
              Описание
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-surface-border">
          {transactions.map((tx) => {
            const config = transactionTypeConfig[tx.type]
            return (
              <tr key={tx.id} className="hover:bg-neutral-50 transition-colors">
                <td className="px-4 py-3 text-sm font-mono text-neutral-600 tabular-nums">
                  {formatDateTime(tx.created_at)}
                </td>
                <td className="px-4 py-3">
                  <span
                    className={cn(
                      'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                      config.className,
                    )}
                  >
                    {config.label}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-right font-mono font-medium tabular-nums">
                  <span
                    className={
                      tx.type === 'earn'
                        ? 'text-green-700'
                        : tx.type === 'spend'
                          ? 'text-orange-700'
                          : 'text-blue-700'
                    }
                  >
                    {tx.type === 'earn' ? '+' : tx.type === 'spend' ? '-' : ''}
                    {formatBalance(Math.abs(tx.amount))}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-right font-mono text-neutral-600 tabular-nums">
                  {formatBalance(tx.balance_after)}
                </td>
                <td className="px-4 py-3 text-sm text-neutral-500">
                  {tx.description || '—'}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

// --- Orders Tab ---
function OrdersTab({ clientId }: { clientId: number }) {
  const { data: stats, isLoading, isError } = useClientOrderStatsQuery(clientId)

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-white rounded-xl border border-surface-border p-4 animate-pulse">
              <div className="h-3 w-20 bg-neutral-200 rounded mb-2" />
              <div className="h-6 w-16 bg-neutral-200 rounded" />
            </div>
          ))}
        </div>
        <div className="h-48 bg-neutral-100 rounded-2xl animate-pulse" />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
        <p className="text-sm text-red-600">Ошибка загрузки данных POS</p>
      </div>
    )
  }

  if (!stats || stats.total_orders === 0) {
    return (
      <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
        <ShoppingCart className="w-8 h-8 text-neutral-300 mx-auto mb-3" />
        <p className="text-sm text-neutral-500">
          Нет данных POS. Подключите интеграцию для отображения заказов.
        </p>
      </div>
    )
  }

  return <OrderStatsContent stats={stats} />
}

function OrderStatsContent({ stats }: { stats: ClientOrderStats }) {
  return (
    <div className="space-y-6">
      {/* Summary cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-xs text-neutral-500 mb-1">Всего заказов</p>
          <p className="text-xl font-semibold text-neutral-900">{stats.total_orders}</p>
        </div>
        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-xs text-neutral-500 mb-1">Общая сумма</p>
          <p className="text-xl font-semibold text-neutral-900">{formatCurrency(stats.total_amount)}</p>
        </div>
        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-xs text-neutral-500 mb-1">Средний чек</p>
          <p className="text-xl font-semibold text-neutral-900">{formatCurrency(stats.avg_amount)}</p>
        </div>
        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-xs text-neutral-500 mb-1">Последний заказ</p>
          <p className="text-xl font-semibold text-neutral-900">
            {stats.last_order_at ? formatDate(stats.last_order_at) : '—'}
          </p>
        </div>
      </div>

      {/* Top items table */}
      {stats.top_items && stats.top_items.length > 0 && (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden">
          <div className="px-6 py-4 border-b border-surface-border">
            <h2 className="text-base font-semibold text-neutral-900">
              Популярные позиции
            </h2>
          </div>
          <table className="w-full">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
                  Позиция
                </th>
                <th className="text-right text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
                  Кол-во заказов
                </th>
                <th className="text-right text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
                  Количество
                </th>
                <th className="text-right text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">
                  Сумма
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {stats.top_items.map((item) => (
                <tr key={item.name} className="hover:bg-neutral-50 transition-colors">
                  <td className="px-4 py-3 text-sm text-neutral-900 font-medium">
                    {item.name}
                  </td>
                  <td className="px-4 py-3 text-sm text-right font-mono text-neutral-600 tabular-nums">
                    {item.order_count}
                  </td>
                  <td className="px-4 py-3 text-sm text-right font-mono text-neutral-600 tabular-nums">
                    {item.total_qty}
                  </td>
                  <td className="px-4 py-3 text-sm text-right font-mono font-medium text-neutral-900 tabular-nums">
                    {formatCurrency(item.total_sum)}
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
