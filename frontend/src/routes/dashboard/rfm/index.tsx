import React, { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { RefreshCw, Settings, ArrowRight, Sprout, TrendingUp, UserCheck, Crown, Gem, AlertTriangle, UserX } from 'lucide-react'
import { useRFMDashboardQuery, useRFMActiveTemplateQuery, useRFMRecalculateMutation } from '@/features/rfm/queries'
import { RFM_SEGMENT_LABELS, RFM_SEGMENT_COLORS, RFM_SEGMENTS } from '@/features/rfm/types'
import type { LucideIcon } from 'lucide-react'

const SEGMENT_ICONS: Record<string, LucideIcon> = {
  new: Sprout,
  promising: TrendingUp,
  regular: UserCheck,
  vip: Crown,
  rare_valuable: Gem,
  churn_risk: AlertTriangle,
  lost: UserX,
}
import { formatMoney, formatDate } from '@/features/rfm/utils'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'

export default function RFMDashboardPage() {
  const navigate = useNavigate()
  const { data: dashboard, isLoading, isError, mutate } = useRFMDashboardQuery()
  const { data: activeTemplate, isLoading: loadingTemplate } = useRFMActiveTemplateQuery()
  const recalcMutation = useRFMRecalculateMutation()

  // Redirect to onboarding if no template configured
  useEffect(() => {
    if (!loadingTemplate && activeTemplate && !activeTemplate.template) {
      navigate('/dashboard/rfm/onboarding', { replace: true })
    }
  }, [activeTemplate, loadingTemplate, navigate])

  if (isLoading || loadingTemplate) {
    return (
      <div>
        <div className="flex items-center justify-between mb-6">
          <div className="shimmer h-8 w-48 rounded" />
          <div className="shimmer h-10 w-32 rounded" />
        </div>
        <TableSkeleton rows={7} />
      </div>
    )
  }

  if (isError) {
    return (
      <div>
        <ErrorState
          title="Не удалось загрузить RFM-данные"
          onRetry={() => mutate()}
        />
      </div>
    )
  }

  const segments = dashboard?.segments ?? []
  const config = dashboard?.config
  const totalClients = segments.reduce((sum, s) => sum + s.client_count, 0)

  // Sort segments in display order
  const sortedSegments = [...segments].sort((a, b) => {
    const order = RFM_SEGMENTS as readonly string[]
    return order.indexOf(a.segment) - order.indexOf(b.segment)
  })

  function handleRecalculate() {
    recalcMutation.mutate(undefined as void)
  }

  return (
    <div>
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 animate-in">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
            RFM-сегментация
          </h1>
          <div className="flex items-center gap-3 mt-1 text-sm text-neutral-500">
            {activeTemplate?.template && (
              <span>
                Шаблон: <span className="font-medium text-neutral-700">{activeTemplate.template.name}</span>
              </span>
            )}
            {config?.last_calc_at && (
              <>
                <span className="text-neutral-300">·</span>
                <span>Обновлено: {formatDate(config.last_calc_at)}</span>
              </>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => navigate('/dashboard/rfm/template')}
            className={cn(
              'flex items-center gap-2 py-2 px-3 rounded',
              'border border-neutral-900 text-sm font-medium text-neutral-600',
              'hover:bg-neutral-50',
              'transition-all duration-150',
            )}
          >
            <Settings className="w-4 h-4" />
            <span className="hidden sm:inline">Сменить</span>
          </button>
          <button
            type="button"
            onClick={handleRecalculate}
            disabled={recalcMutation.isPending}
            className={cn(
              'flex items-center gap-2 py-2 px-4 rounded',
              'bg-accent text-white text-sm font-medium',
              'hover:bg-accent-hover active:bg-accent/80',
              'transition-all duration-150',
              'disabled:opacity-50 disabled:cursor-not-allowed',
            )}
          >
            <RefreshCw className={cn('w-4 h-4', recalcMutation.isPending && 'animate-spin')} />
            Обновить
          </button>
        </div>
      </div>

      {/* Recalc feedback */}
      {recalcMutation.isSuccess && (
        <div className="mb-4 text-sm text-green-700 bg-green-50 p-3 rounded animate-in">
          RFM-сегменты пересчитаны. Данные обновятся в течение нескольких секунд.
        </div>
      )}
      {recalcMutation.isError && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded animate-in">
          Ошибка при пересчёте. Попробуйте снова.
        </div>
      )}

      {/* Segments table */}
      {segments.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 text-center animate-in">
          <div className="w-16 h-16 rounded bg-neutral-100 flex items-center justify-center mb-4">
            <RefreshCw className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5 tracking-tight">Данных пока нет</h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed mb-4">Запустите пересчёт RFM-сегментов</p>
          <button
            type="button"
            onClick={handleRecalculate}
            disabled={recalcMutation.isPending}
            className={cn(
              'inline-flex items-center gap-2 py-2.5 px-5 rounded',
              'bg-accent text-white text-sm font-medium',
              'hover:bg-accent-hover transition-all duration-150',
            )}
          >
            <RefreshCw className={cn('w-4 h-4', recalcMutation.isPending && 'animate-spin')} />
            Пересчитать
          </button>
        </div>
      ) : (
        <div className="bg-white rounded border border-neutral-900 overflow-hidden animate-in">
          {/* Table header */}
          <div className="grid grid-cols-[1fr_80px_60px_100px_120px_40px] gap-2 px-6 py-4 text-xs font-medium text-neutral-400 uppercase tracking-wider">
            <span>Сегмент</span>
            <span className="text-right">Клиенты</span>
            <span className="text-right">%</span>
            <span className="text-right">Ср. чек</span>
            <span className="text-right">Выручка</span>
            <span />
          </div>
          <div className="mx-6 border-t border-neutral-200" />

          {/* Rows */}
          {sortedSegments.map((seg, i) => {
            const colors = RFM_SEGMENT_COLORS[seg.segment]
            const label = RFM_SEGMENT_LABELS[seg.segment] || seg.segment
            return (
              <React.Fragment key={seg.segment}>
              <button
                type="button"
                onClick={() => navigate(`/dashboard/rfm/segments/${seg.segment}`)}
                className={cn(
                  'w-full grid grid-cols-[1fr_80px_60px_100px_120px_40px] gap-2 px-6 py-4',
                  'hover:bg-neutral-50 transition-colors duration-150 text-left group',
                  'animate-in',
                  `animate-in-delay-${Math.min(i + 1, 5)}`,
                )}
              >
                <div className="flex items-center gap-3">
                  {(() => {
                    const Icon = SEGMENT_ICONS[seg.segment]
                    return Icon ? <Icon className="w-4.5 h-4.5 shrink-0" style={{ color: colors?.color }} /> : <span className="w-4.5 h-4.5 rounded-full" style={{ background: colors?.color }} />
                  })()}
                  <span className="text-sm font-medium text-neutral-900">{label}</span>
                </div>
                <span className="text-sm font-mono tabular-nums text-neutral-700 text-right self-center">
                  {seg.client_count}
                </span>
                <span className="text-sm font-mono tabular-nums text-neutral-500 text-right self-center">
                  {Math.round(seg.percentage)}%
                </span>
                <span className="text-sm font-mono tabular-nums text-neutral-700 text-right self-center">
                  {formatMoney(seg.avg_check)}
                </span>
                <span className="text-sm font-mono tabular-nums text-neutral-700 text-right self-center">
                  {formatMoney(seg.total_check)}
                </span>
                <div className="flex items-center justify-end self-center">
                  <ArrowRight className="w-4 h-4 text-neutral-300 group-hover:text-neutral-500 transition-colors" />
                </div>
              </button>
              {i < sortedSegments.length - 1 && (
                <div className="mx-6 border-t border-neutral-200" />
              )}
            </React.Fragment>
            )
          })}

          <div className="mx-6 border-t border-neutral-200" />
          {/* Total row */}
          <div className="grid grid-cols-[1fr_80px_60px_100px_120px_40px] gap-2 px-6 py-4 bg-neutral-50 text-sm font-semibold text-neutral-700">
            <span>Всего</span>
            <span className="text-right font-mono tabular-nums">{totalClients}</span>
            <span />
            <span />
            <span className="text-right font-mono tabular-nums">
              {formatMoney(segments.reduce((sum, s) => sum + s.total_check, 0))}
            </span>
            <span />
          </div>
        </div>
      )}
    </div>
  )
}
