import { useState } from 'react'
import { cn, getApiErrorMessage } from '@/lib/utils'
import {
  usePromotionsQuery,
  useCreatePromotionMutation,
  useUpdatePromotionMutation,
  useDeletePromotionMutation,
} from '@/features/promotions/queries'
import type {
  Promotion,
  CreatePromotionRequest,
} from '@/features/promotions/types'
import { Tag, Plus, X, Trash2, ToggleLeft, ToggleRight } from 'lucide-react'
import { CustomSelect } from '@/components/common/CustomSelect'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'

const typeConfig: Record<Promotion['type'], { label: string; className: string }> = {
  discount: {
    label: 'Скидка',
    className: 'bg-blue-100 text-blue-700',
  },
  bonus: {
    label: 'Бонус',
    className: 'bg-purple-100 text-purple-700',
  },
  tag: {
    label: 'Тег',
    className: 'bg-amber-100 text-amber-700',
  },
  campaign: {
    label: 'Рассылка',
    className: 'bg-cyan-100 text-cyan-700',
  },
}

const recurrenceLabels: Record<Promotion['recurrence'], string> = {
  one_time: 'Разовая',
  daily: 'Ежедневно',
  weekly: 'Еженедельно',
  monthly: 'Ежемесячно',
}

function getStatusInfo(promo: Promotion): { label: string; className: string } {
  if (!promo.active) {
    return { label: 'Неактивна', className: 'bg-neutral-100 text-neutral-500' }
  }
  if (promo.ends_at && new Date(promo.ends_at) < new Date()) {
    return { label: 'Истекла', className: 'bg-red-100 text-red-700' }
  }
  if (promo.starts_at && new Date(promo.starts_at) > new Date()) {
    return { label: 'Запланирована', className: 'bg-blue-100 text-blue-700' }
  }
  return { label: 'Активна', className: 'bg-green-100 text-green-700' }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function CreatePromotionModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('')
  const [type, setType] = useState<string>('discount')
  const [discountPercent, setDiscountPercent] = useState<string>('')
  const [bonusAmount, setBonusAmount] = useState<string>('')
  const [minAmount, setMinAmount] = useState<string>('')
  const [usageLimit, setUsageLimit] = useState<string>('')
  const [recurrence, setRecurrence] = useState<string>('one_time')
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [combinable, setCombinable] = useState(false)

  const createPromotion = useCreatePromotionMutation()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const data: CreatePromotionRequest = {
      name,
      type,
      conditions: {
        ...(minAmount ? { min_amount: Number(minAmount) } : {}),
      },
      result: {
        ...(type === 'discount' && discountPercent
          ? { discount_percent: Number(discountPercent) }
          : {}),
        ...(type === 'bonus' && bonusAmount
          ? { bonus_amount: Number(bonusAmount) }
          : {}),
      },
      recurrence,
      ...(startsAt ? { starts_at: new Date(startsAt).toISOString() } : {}),
      ...(endsAt ? { ends_at: new Date(endsAt).toISOString() } : {}),
      ...(usageLimit ? { usage_limit: Number(usageLimit) } : {}),
      combinable,
    }

    try {
      await createPromotion.mutateAsync(data)
      onClose()
    } catch {
      // error is available via createPromotion.error
    }
  }

  const inputClass = cn(
    'w-full px-4 py-2.5 rounded-lg border border-surface-border',
    'text-sm placeholder:text-neutral-400',
    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
    'transition-colors',
    'disabled:opacity-50 disabled:cursor-not-allowed',
  )

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="create-promotion-title"
    >
      <div className="bg-white rounded-2xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h2
            id="create-promotion-title"
            className="text-lg font-semibold text-neutral-900"
          >
            Создать акцию
          </h2>
          <button
            onClick={onClose}
            type="button"
            className="p-1 rounded-lg text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
            aria-label="Закрыть"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label
              htmlFor="promo-name"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Название
            </label>
            <input
              id="promo-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Скидка 10% на первый заказ"
              required
              maxLength={200}
              autoFocus
              disabled={createPromotion.isPending}
              className={inputClass}
            />
          </div>

          <div>
            <label
              htmlFor="promo-type"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Тип
            </label>
            <CustomSelect
              value={type}
              onChange={(v) => setType(v)}
              options={[
                { value: 'discount', label: 'Скидка' },
                { value: 'bonus', label: 'Бонус' },
                { value: 'tag', label: 'Тег' },
                { value: 'campaign', label: 'Рассылка' },
              ]}
            />
          </div>

          {type === 'discount' && (
            <div>
              <label
                htmlFor="promo-discount"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Процент скидки
              </label>
              <input
                id="promo-discount"
                type="number"
                value={discountPercent}
                onChange={(e) => setDiscountPercent(e.target.value)}
                placeholder="10"
                min={1}
                max={100}
                disabled={createPromotion.isPending}
                className={inputClass}
              />
            </div>
          )}

          {type === 'bonus' && (
            <div>
              <label
                htmlFor="promo-bonus"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Сумма бонуса
              </label>
              <input
                id="promo-bonus"
                type="number"
                value={bonusAmount}
                onChange={(e) => setBonusAmount(e.target.value)}
                placeholder="500"
                min={1}
                disabled={createPromotion.isPending}
                className={inputClass}
              />
            </div>
          )}

          <div>
            <label
              htmlFor="promo-min-amount"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Мин. сумма заказа{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <input
              id="promo-min-amount"
              type="number"
              value={minAmount}
              onChange={(e) => setMinAmount(e.target.value)}
              placeholder="1000"
              min={0}
              disabled={createPromotion.isPending}
              className={inputClass}
            />
          </div>

          <div>
            <label
              htmlFor="promo-recurrence"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Повторяемость
            </label>
            <CustomSelect
              value={recurrence}
              onChange={(v) => setRecurrence(v)}
              options={[
                { value: 'one_time', label: 'Разовая' },
                { value: 'daily', label: 'Ежедневно' },
                { value: 'weekly', label: 'Еженедельно' },
                { value: 'monthly', label: 'Ежемесячно' },
              ]}
            />
          </div>

          <div>
            <label
              htmlFor="promo-usage-limit"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Лимит использований{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <input
              id="promo-usage-limit"
              type="number"
              value={usageLimit}
              onChange={(e) => setUsageLimit(e.target.value)}
              placeholder="100"
              min={1}
              disabled={createPromotion.isPending}
              className={inputClass}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label
                htmlFor="promo-starts-at"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Дата начала
              </label>
              <input
                id="promo-starts-at"
                type="date"
                value={startsAt}
                onChange={(e) => setStartsAt(e.target.value)}
                disabled={createPromotion.isPending}
                className={inputClass}
              />
            </div>
            <div>
              <label
                htmlFor="promo-ends-at"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Дата окончания
              </label>
              <input
                id="promo-ends-at"
                type="date"
                value={endsAt}
                onChange={(e) => setEndsAt(e.target.value)}
                disabled={createPromotion.isPending}
                className={inputClass}
              />
            </div>
          </div>

          <div className="flex items-center gap-3">
            <input
              id="promo-combinable"
              type="checkbox"
              checked={combinable}
              onChange={(e) => setCombinable(e.target.checked)}
              disabled={createPromotion.isPending}
              className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
            />
            <label
              htmlFor="promo-combinable"
              className="text-sm font-medium text-neutral-700"
            >
              Можно совмещать с другими акциями
            </label>
          </div>

          {createPromotion.isError && (
            <p className="text-sm text-red-600">
              {getApiErrorMessage(
                createPromotion.error,
                'Не удалось создать акцию. Попробуйте снова.',
              )}
            </p>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={createPromotion.isPending}
              className={cn(
                'flex-1 py-2.5 px-4 rounded-lg',
                'border border-surface-border text-sm font-medium text-neutral-700',
                'hover:bg-neutral-50 active:bg-neutral-100',
                'transition-colors',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={createPromotion.isPending}
              className={cn(
                'flex-1 py-2.5 px-4 rounded-lg',
                'bg-accent text-white text-sm font-medium',
                'hover:bg-accent/90 active:bg-accent/80',
                'transition-colors',
                'focus:outline-none focus:ring-2 focus:ring-accent/20',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              {createPromotion.isPending ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function PromotionsPage() {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const { data: promotions, isLoading, isError, mutate } = usePromotionsQuery()
  const updatePromotion = useUpdatePromotionMutation()
  const deletePromotion = useDeletePromotionMutation()

  const activePromotions = (promotions ?? []).filter((p) => {
    if (!p.active) return true
    if (p.ends_at && new Date(p.ends_at) < new Date()) return false
    return true
  })

  const handleToggleActive = async (promo: Promotion) => {
    try {
      await updatePromotion.mutateAsync({
        id: promo.id,
        data: { active: !promo.active },
      })
      mutate()
    } catch {
      // error handled silently
    }
  }

  const handleDelete = async (promo: Promotion) => {
    if (!confirm(`Удалить акцию "${promo.name}"?`)) return
    try {
      await deletePromotion.mutateAsync(promo.id)
      mutate()
    } catch {
      // error handled silently
    }
  }

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6 animate-in">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
            Акции
          </h1>
          <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1">
            Управление акциями и специальными предложениями
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          type="button"
          className={cn(
            'flex items-center gap-2 py-2.5 px-4 rounded-lg',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
            'shadow-sm shadow-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          <span className="hidden sm:inline">Создать акцию</span>
          <span className="sm:hidden">Создать</span>
        </button>
      </div>

      {isLoading ? (
        <div className="animate-in animate-in-delay-1">
          <TableSkeleton />
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить акции"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : activePromotions.length === 0 ? (
        <EmptyState
          icon={Tag}
          title="У вас пока нет акций"
          description="Создайте акцию, чтобы привлечь клиентов скидками, бонусами и специальными предложениями."
          actionLabel="Создать акцию"
          onAction={() => setShowCreateModal(true)}
        />
      ) : (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden animate-in animate-in-delay-1">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Название
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Тип
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Статус
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Повторяемость
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Период
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Действия
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {activePromotions.map((promo) => {
                  const status = getStatusInfo(promo)
                  const promoType = typeConfig[promo.type]
                  return (
                    <tr
                      key={promo.id}
                      className="hover:bg-neutral-50 transition-colors"
                    >
                      <td className="px-6 py-4">
                        <div>
                          <span className="text-sm font-medium text-neutral-900">
                            {promo.name}
                          </span>
                          {promo.result.discount_percent && (
                            <p className="text-xs text-neutral-400 mt-0.5">
                              {promo.result.discount_percent}% скидка
                            </p>
                          )}
                          {promo.result.bonus_amount && (
                            <p className="text-xs text-neutral-400 mt-0.5">
                              +{promo.result.bonus_amount} бонусов
                            </p>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 hidden sm:table-cell">
                        <span
                          className={cn(
                            'text-xs font-medium px-2 py-1 rounded-full',
                            promoType.className,
                          )}
                        >
                          {promoType.label}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={cn(
                            'text-xs font-medium px-2 py-1 rounded-full',
                            status.className,
                          )}
                        >
                          {status.label}
                        </span>
                      </td>
                      <td className="px-6 py-4 hidden md:table-cell">
                        <span className="text-sm text-neutral-500">
                          {recurrenceLabels[promo.recurrence]}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden sm:table-cell">
                        <span className="text-sm font-mono text-neutral-400 tabular-nums">
                          {formatDate(promo.starts_at)} — {formatDate(promo.ends_at)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            type="button"
                            onClick={() => handleToggleActive(promo)}
                            className="p-1.5 rounded-lg text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
                            title={promo.active ? 'Деактивировать' : 'Активировать'}
                            aria-label={promo.active ? 'Деактивировать' : 'Активировать'}
                          >
                            {promo.active ? (
                              <ToggleRight className="w-4 h-4 text-green-600" />
                            ) : (
                              <ToggleLeft className="w-4 h-4" />
                            )}
                          </button>
                          <button
                            type="button"
                            onClick={() => handleDelete(promo)}
                            className="p-1.5 rounded-lg text-neutral-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                            title="Удалить"
                            aria-label="Удалить"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showCreateModal && (
        <CreatePromotionModal onClose={() => setShowCreateModal(false)} />
      )}
    </div>
  )
}
