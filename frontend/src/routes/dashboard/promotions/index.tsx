import { cn } from '@/lib/utils'
import {
  usePromotionsQuery,
  useUpdatePromotionMutation,
  useDeletePromotionMutation,
} from '@/features/promotions/queries'
import type { Promotion } from '@/features/promotions/types'
import { Tag, Plus, Trash2, ToggleLeft, ToggleRight } from 'lucide-react'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'
import { useNavigate } from 'react-router-dom'

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

export default function PromotionsPage() {
  const navigate = useNavigate()
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
          <p className="text-sm text-neutral-500 mt-1">
            Управление акциями и специальными предложениями
          </p>
        </div>
        <button
          onClick={() => navigate('/dashboard/promotions/create')}
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
          onAction={() => navigate('/dashboard/promotions/create')}
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
    </div>
  )
}
