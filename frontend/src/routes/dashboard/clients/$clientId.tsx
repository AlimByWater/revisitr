import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  useClientProfileQuery,
  useUpdateTagsMutation,
} from '@/features/clients/queries'
import type { LoyaltyTransaction } from '@/features/clients/types'
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
} from 'lucide-react'

export const Route = createFileRoute('/dashboard/clients/$clientId')({
  component: ClientDetailPage,
})

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

function ClientDetailPage() {
  const { clientId } = Route.useParams()
  const id = Number(clientId)

  const { data: client, isLoading } = useClientProfileQuery(id)
  const updateTagsMutation = useUpdateTagsMutation()

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

      <h1 className="text-2xl font-bold text-neutral-900 mb-6">{fullName}</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
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
                <p className="text-xl font-bold text-neutral-900 tabular-nums">
                  {formatBalance(client.loyalty_balance)}
                </p>
              </div>
              <div className="bg-neutral-50 rounded-xl p-4">
                <div className="flex items-center gap-2 mb-1">
                  <TrendingUp className="w-4 h-4 text-neutral-400" />
                  <span className="text-xs text-neutral-500">Покупок</span>
                </div>
                <p className="text-xl font-bold text-neutral-900 tabular-nums">
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
              <p className="text-xl font-bold text-neutral-900 tabular-nums">
                {formatBalance(client.total_purchases)}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Tags */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-6">
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
                onClick={() => handleRemoveTag(tag)}
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
                onChange={(e) => setNewTag(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleAddTag()
                  if (e.key === 'Escape') {
                    setShowTagInput(false)
                    setNewTag('')
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
                onClick={handleAddTag}
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
              onClick={() => setShowTagInput(true)}
              className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium border border-dashed border-neutral-300 text-neutral-500 hover:border-neutral-400 hover:text-neutral-700 transition-colors"
            >
              <Plus className="w-3 h-3" />
              Добавить тег
            </button>
          )}
        </div>
      </div>

      {/* Transactions */}
      {client.transactions && client.transactions.length > 0 && (
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
              {client.transactions.map((tx) => {
                const config = transactionTypeConfig[tx.type]
                return (
                  <tr key={tx.id} className="hover:bg-neutral-50 transition-colors">
                    <td className="px-4 py-3 text-sm text-neutral-600">
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
                    <td className="px-4 py-3 text-sm text-right font-medium tabular-nums">
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
                    <td className="px-4 py-3 text-sm text-right text-neutral-600 tabular-nums">
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
      )}
    </div>
  )
}
