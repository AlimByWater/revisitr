import { useMemo, useState } from 'react'
import { cn } from '@/lib/utils'
import {
  usePromotionsQuery,
  useUpdatePromotionMutation,
  useDeletePromotionMutation,
} from '@/features/promotions/queries'
import type { Promotion } from '@/features/promotions/types'
import { Tag, Plus, Trash2, Search } from 'lucide-react'
import { CustomSelect } from '@/components/common/CustomSelect'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'
import { useNavigate, Link } from 'react-router-dom'
import { Button } from '@/components/common/Button'

const typeConfig: Record<Promotion['type'], { label: string; className: string }> = {
  discount: {
    label: 'Скидка',
    className: 'bg-accent/10 text-accent border-accent/30',
  },
  bonus: {
    label: 'Бонус',
    className: 'bg-violet-500/10 text-violet-700 border-violet-500/30',
  },
  tag_update: {
    label: 'Тег',
    className: 'bg-amber-500/10 text-amber-700 border-amber-500/30',
  },
  campaign: {
    label: 'Рассылка',
    className: 'bg-sky-500/10 text-sky-700 border-sky-500/30',
  },
}

function getStatusInfo(promo: Promotion): { label: string; className: string } {
  if (!promo.active) {
    return { label: 'Неактивна', className: 'bg-neutral-100 text-neutral-600 border border-neutral-300' }
  }
  if (promo.ends_at && new Date(promo.ends_at) < new Date()) {
    return { label: 'Истекла', className: 'bg-red-500/10 text-red-700 border border-red-500/30' }
  }
  if (promo.starts_at && new Date(promo.starts_at) > new Date()) {
    return { label: 'Запланирована', className: 'bg-amber-500/10 text-amber-700 border border-amber-500/30' }
  }
  return { label: 'Активна', className: 'bg-emerald-500/10 text-emerald-700 border border-emerald-500/30' }
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

  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('')

  const activePromotions = (promotions ?? []).filter((p) => {
    if (!p.active) return true
    if (p.ends_at && new Date(p.ends_at) < new Date()) return false
    return true
  })

  const filteredPromotions = useMemo(() => {
    return activePromotions.filter((p) => {
      if (typeFilter && p.type !== typeFilter) return false
      if (search.trim()) {
        const q = search.trim().toLowerCase()
        if (!p.name.toLowerCase().includes(q)) return false
      }
      return true
    })
  }, [activePromotions, search, typeFilter])

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
          <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">
            Акции
          </h1>
          <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">
            Управление акциями и специальными предложениями
          </p>
        </div>
        <Button
          variant="primary"
          leftIcon={<Plus className="w-4 h-4" />}
          onClick={() => navigate('/dashboard/promotions/create')}
        >
          <span className="hidden sm:inline">Создать акцию</span>
          <span className="sm:hidden">Создать</span>
        </Button>
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
        <>
          <div className="flex items-center gap-3 mb-4 flex-wrap animate-in animate-in-delay-1">
            <div className="relative flex-1 max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-neutral-400" />
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Поиск по названию..."
                className={cn(
                  'w-full pl-9 pr-4 py-2.5 rounded border border-neutral-200 bg-white',
                  'text-sm placeholder:text-neutral-400',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  'transition-colors',
                )}
              />
            </div>
            <CustomSelect
              value={typeFilter}
              onChange={(v) => setTypeFilter(v)}
              options={[
                { value: '', label: 'Все типы' },
                { value: 'discount', label: 'Скидка' },
                { value: 'bonus', label: 'Бонус' },
                { value: 'tag_update', label: 'Тег' },
                { value: 'campaign', label: 'Рассылка' },
              ]}
              placeholder="Все типы"
              width="180px"
              light
            />
            <span className="text-xs font-mono text-neutral-400 tabular-nums">
              {filteredPromotions.length} / {activePromotions.length}
            </span>
          </div>

          {filteredPromotions.length === 0 ? (
            <div className="bg-white rounded border border-neutral-900 px-6 py-12 text-center text-sm text-neutral-400 animate-in">
              Ничего не найдено по выбранным фильтрам
            </div>
          ) : (
        <div className="bg-white rounded border border-neutral-900 overflow-hidden animate-in animate-in-delay-1">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-neutral-200">
                  <th className="text-left text-[11px] font-mono font-medium text-neutral-400 uppercase tracking-wider px-6 py-4">
                    Название
                  </th>
                  <th className="text-center text-[11px] font-mono font-medium text-neutral-400 uppercase tracking-wider px-6 py-4 hidden sm:table-cell">
                    Тип
                  </th>
                  <th className="text-center text-[11px] font-mono font-medium text-neutral-400 uppercase tracking-wider px-6 py-4">
                    Статус
                  </th>
                  <th className="text-right text-[11px] font-mono font-medium text-neutral-400 uppercase tracking-wider px-6 py-4 hidden md:table-cell">
                    Период
                  </th>
                  <th className="text-right text-[11px] font-mono font-medium text-neutral-400 uppercase tracking-wider px-6 py-4 w-24">
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-200">
                {filteredPromotions.map((promo) => {
                  const status = getStatusInfo(promo)
                  const promoType = typeConfig[promo.type]
                  const valueStr = formatValue(promo)
                  return (
                    <tr
                      key={promo.id}
                      className="hover:bg-neutral-50 transition-colors cursor-pointer"
                      onClick={() => navigate(`/dashboard/promotions/${promo.id}`)}
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
                            'inline-block font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border',
                            promoType.className,
                          )}
                        >
                          {promoType.label}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-center">
                        <span
                          className={cn(
                            'font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border',
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
                            className={cn(
                              'relative w-9 h-5 rounded-full transition-colors duration-200 focus:outline-none',
                              promo.active ? 'bg-accent' : 'bg-neutral-300',
                            )}
                            title={promo.active ? 'Деактивировать' : 'Активировать'}
                            aria-label={promo.active ? 'Деактивировать' : 'Активировать'}
                          >
                            <span
                              className={cn(
                                'absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform duration-200',
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
        </>
      )}
    </div>
  )
}
