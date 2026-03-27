import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { RefreshCw, Settings, ArrowRight } from 'lucide-react'
import { useRFMDashboardQuery, useRFMActiveTemplateQuery, useRFMRecalculateMutation } from '@/features/rfm/queries'
import { RFM_SEGMENT_LABELS, RFM_SEGMENT_COLORS, RFM_SEGMENTS } from '@/features/rfm/types'
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
      <div className="max-w-5xl">
        <div className="flex items-center justify-between mb-6">
          <div className="shimmer h-8 w-48 rounded" />
          <div className="shimmer h-10 w-32 rounded-lg" />
        </div>
        <TableSkeleton rows={7} />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="max-w-5xl">
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
    <div className="max-w-5xl">
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
              'flex items-center gap-2 py-2 px-3 rounded-lg',
              'border border-neutral-200 text-sm font-medium text-neutral-600',
              'hover:bg-neutral-50 hover:border-neutral-300',
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
              'flex items-center gap-2 py-2 px-4 rounded-lg',
              'bg-accent text-white text-sm font-medium',
              'hover:bg-accent-hover active:bg-accent/80',
              'transition-all duration-150',
              'shadow-sm shadow-accent/20',
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
        <div className="mb-4 text-sm text-green-700 bg-green-50 p-3 rounded-lg animate-in">
          RFM-сегменты пересчитаны. Данные обновятся в течение нескольких секунд.
        </div>
      )}
      {recalcMutation.isError && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded-lg animate-in">
          Ошибка при пересчёте. Попробуйте снова.
        </div>
      )}

      {/* Segments table */}
      {segments.length === 0 ? (
        <div className="bg-white rounded-2xl border border-surface-border py-16 px-8 text-center animate-in">
          <p className="text-neutral-500 mb-4">Данных пока нет. Запустите пересчёт RFM-сегментов.</p>
          <button
            type="button"
            onClick={handleRecalculate}
            disabled={recalcMutation.isPending}
            className={cn(
              'inline-flex items-center gap-2 py-2.5 px-5 rounded-lg',
              'bg-accent text-white text-sm font-medium',
              'hover:bg-accent-hover transition-all duration-150',
              'shadow-sm shadow-accent/20',
            )}
          >
            <RefreshCw className={cn('w-4 h-4', recalcMutation.isPending && 'animate-spin')} />
            Пересчитать
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-2xl border border-surface-border overflow-hidden animate-in">
          {/* Table header */}
          <div className="grid grid-cols-[1fr_80px_60px_100px_120px_40px] gap-2 px-6 py-3 border-b border-surface-border text-xs font-medium text-neutral-400 uppercase tracking-wider">
            <span>Сегмент</span>
            <span className="text-right">Клиенты</span>
            <span className="text-right">%</span>
            <span className="text-right">Ср. чек</span>
            <span className="text-right">Выручка</span>
            <span />
          </div>

          {/* Rows */}
          {sortedSegments.map((seg, i) => {
            const colors = RFM_SEGMENT_COLORS[seg.segment]
            const label = RFM_SEGMENT_LABELS[seg.segment] || seg.segment
            return (
              <button
                key={seg.segment}
                type="button"
                onClick={() => navigate(`/dashboard/rfm/segments/${seg.segment}`)}
                className={cn(
                  'w-full grid grid-cols-[1fr_80px_60px_100px_120px_40px] gap-2 px-6 py-4',
                  'border-b border-surface-border last:border-0',
                  'hover:bg-neutral-50 transition-colors duration-150 text-left group',
                  'animate-in',
                  `animate-in-delay-${Math.min(i + 1, 5)}`,
                )}
              >
                <div className="flex items-center gap-3">
                  <span className="text-base">{colors?.icon || '●'}</span>
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
            )
          })}

          {/* Total row */}
          <div className="grid grid-cols-[1fr_80px_60px_100px_120px_40px] gap-2 px-6 py-3 bg-neutral-50 text-xs font-medium text-neutral-500">
            <span>Всего</span>
            <span className="text-right font-mono tabular-nums">{totalClients}</span>
            <span className="text-right font-mono tabular-nums">100%</span>
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
