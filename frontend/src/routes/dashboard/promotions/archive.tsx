import { cn } from '@/lib/utils'
import { usePromotionsQuery, useChannelAnalyticsQuery } from '@/features/promotions/queries'
import type { Promotion } from '@/features/promotions/types'
import { Archive, BarChart3 } from 'lucide-react'
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
  tag_update: {
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

const channelLabels: Record<string, string> = {
  smm: 'SMM',
  targeting: 'Таргетинг',
  yandex_maps: 'Яндекс Карты',
  flyer: 'Флаер',
  partner: 'Партнёр',
  custom: 'Другой',
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export default function PromotionsArchivePage() {
  const { data: promotions, isLoading, isError, mutate } = usePromotionsQuery()
  const { data: channelAnalytics } = useChannelAnalyticsQuery()

  const archivedPromotions = (promotions ?? []).filter((p) => {
    if (!p.active) return true
    if (p.ends_at && new Date(p.ends_at) < new Date()) return true
    return false
  })

  return (
    <div className="max-w-4xl">
      <div className="mb-6 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          Архив акций
        </h1>
        <p className="text-sm text-neutral-500 mt-1">
          Завершённые и деактивированные акции
        </p>
      </div>

      {channelAnalytics && channelAnalytics.length > 0 && (
        <div className="mb-6 animate-in">
          <h2 className="text-sm font-medium text-neutral-700 mb-3 flex items-center gap-2">
            <BarChart3 className="w-4 h-4 text-neutral-400" />
            Аналитика по каналам
          </h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
            {channelAnalytics.map((ch) => (
              <div
                key={ch.channel}
                className="bg-white rounded-xl border border-surface-border p-4"
              >
                <p className="text-xs font-medium text-neutral-400 uppercase tracking-wider mb-1">
                  {channelLabels[ch.channel] ?? ch.channel}
                </p>
                <p className="text-lg font-bold text-neutral-900 tabular-nums">
                  {ch.total_usages}
                </p>
                <div className="flex items-center gap-3 mt-1.5 text-xs text-neutral-400">
                  <span>{ch.code_count} кодов</span>
                  <span>{ch.unique_clients} клиентов</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="animate-in animate-in-delay-1">
          <TableSkeleton />
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить архив"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : archivedPromotions.length === 0 ? (
        <EmptyState
          icon={Archive}
          title="Архив пуст"
          description="Здесь будут отображаться завершённые и деактивированные акции."
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
                    Причина
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Повторяемость
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Период
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {archivedPromotions.map((promo) => {
                  const promoType = typeConfig[promo.type]
                  const isExpired =
                    promo.ends_at && new Date(promo.ends_at) < new Date()
                  const reason = !promo.active
                    ? { label: 'Деактивирована', className: 'bg-neutral-100 text-neutral-500' }
                    : isExpired
                      ? { label: 'Истекла', className: 'bg-red-100 text-red-700' }
                      : { label: 'Завершена', className: 'bg-neutral-100 text-neutral-500' }

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
                          {promo.usage_limit && (
                            <p className="text-xs text-neutral-400 mt-0.5">
                              Лимит: {promo.usage_limit}
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
                            reason.className,
                          )}
                        >
                          {reason.label}
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
