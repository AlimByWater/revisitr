import { cn } from '@/lib/utils'
import {
  usePromotionsQuery,
  useUpdatePromotionMutation,
  useDeletePromotionMutation,
} from '@/features/promotions/queries'
import type { Promotion } from '@/features/promotions/types'
import { Tag, Plus, Trash2 } from 'lucide-react'
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
  tag_update: {
    label: 'Тег',
    className: 'bg-amber-100 text-amber-700',
  },
  campaign: {
    label: 'Рассылка',
    className: 'bg-cyan-100 text-cyan-700',
  },
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
  })
}

function formatValue(promo: Promotion): string {
  if (promo.result.discount_percent) return `${promo.result.discount_percent}%`
  if (promo.result.bonus_amount) return `+${promo.result.bonus_amount}`
  return ''
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
    <div>
      <div className="flex items-center justify-between mb-6 animate-in">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
            Акции
          </h1>
          <p className="font-mono text-xs text-neutral-400 uppercase tracking-wider mt-1">
            Управление акциями и специальными предложениями
          </p>
        </div>
        <button
          onClick={() => navigate('/dashboard/promotions/create')}
          type="button"
          className={cn(
            'flex items-center gap-2 py-2.5 px-4 rounded',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
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
        <div className="bg-white rounded border border-neutral-900 overflow-hidden animate-in animate-in-delay-1">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-neutral-200">
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Название
                  </th>
                  <th className="text-center text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Тип
                  </th>
                  <th className="text-center text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Статус
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Период
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 w-24">
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-200">
                {activePromotions.map((promo) => {
                  const status = getStatusInfo(promo)
                  const promoType = typeConfig[promo.type]
                  const valueStr = formatValue(promo)
                  return (
                    <tr
                      key={promo.id}
                      className="hover:bg-neutral-50 transition-colors"
                    >
                      <td className="px-6 py-4">
                        <span className="text-sm font-medium text-neutral-900">
                          {promo.name}
                        </span>
                        {valueStr && (
                          <span className="ml-2 text-xs text-neutral-400">{valueStr}</span>
                        )}
                      </td>
                      <td className="px-6 py-4 text-center hidden sm:table-cell">
                        <span
                          className={cn(
                            'text-xs font-medium px-2 py-1 rounded',
                            promoType.className,
                          )}
                        >
                          {promoType.label}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-center">
                        <span
                          className={cn(
                            'text-xs font-medium px-2 py-1 rounded',
                            status.className,
                          )}
                        >
                          {status.label}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden md:table-cell">
                        <span className="text-sm font-mono text-neutral-400 tabular-nums whitespace-nowrap">
                          {formatDate(promo.starts_at)} — {formatDate(promo.ends_at)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex items-center justify-end gap-3">
                          <button
                            type="button"
                            onClick={() => handleToggleActive(promo)}
                            className="relative w-9 h-5 rounded-full transition-colors duration-200 focus:outline-none"
                            style={{ backgroundColor: promo.active ? '#EF3219' : '#d4d4d4' }}
                            title={promo.active ? 'Деактивировать' : 'Активировать'}
                            aria-label={promo.active ? 'Деактивировать' : 'Активировать'}
                          >
                            <span
                              className={cn(
                                'absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform duration-200 shadow-sm',
                                promo.active ? 'left-[18px]' : 'left-0.5',
                              )}
                            />
                          </button>
                          <button
                            type="button"
                            onClick={() => handleDelete(promo)}
                            className="p-1.5 rounded text-neutral-300 hover:text-red-500 hover:bg-red-50 transition-colors"
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
