import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useRFMDashboardQuery } from '@/features/rfm/queries'
import { RFM_SEGMENT_LABELS, RFM_SEGMENT_COLORS } from '@/features/rfm/types'
import type { RFMHistory } from '@/features/rfm/types'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'
import { Users, TrendingUp, Sprout, UserCheck, Crown, Gem, AlertTriangle, UserX } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { PeriodFilter } from '@/components/common/PeriodFilter'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'

const SEGMENT_HEX_COLORS: Record<string, string> = {
  new: '#10b981',
  promising: '#3b82f6',
  regular: '#8b5cf6',
  vip: '#f59e0b',
  rare_valuable: '#a855f7',
  churn_risk: '#f97316',
  lost: '#ef4444',
}

const SEGMENT_ICONS: Record<string, LucideIcon> = {
  new: Sprout,
  promising: TrendingUp,
  regular: UserCheck,
  vip: Crown,
  rare_valuable: Gem,
  churn_risk: AlertTriangle,
  lost: UserX,
}

function buildChartData(trends: RFMHistory[]) {
  const grouped: Record<string, Record<string, number>> = {}
  for (const t of trends) {
    const date = new Date(t.calculated_at).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })
    if (!grouped[date]) grouped[date] = {}
    grouped[date][t.segment] = t.client_count
  }
  return Object.entries(grouped).map(([date, counts]) => ({ date, ...counts }))
}

function formatMoney(val: number): string {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 }).format(val) + ' ₽'
}

export default function SegmentsPage() {
  const [filter, setFilter] = useState({ period: '30d', from: undefined as string | undefined, to: undefined as string | undefined })
  const { data: dashboard, isLoading, isError, mutate } = useRFMDashboardQuery()

  const segments = dashboard?.segments ?? []
  const trends = dashboard?.trends ?? []
  const totalClients = segments.reduce((s, seg) => s + seg.client_count, 0)

  function handleSegmentClick(segKey: string) {
    window.open(`/revisitr/dashboard/rfm/segments/${segKey}`, '_blank')
  }

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <div className="h-8 w-48 shimmer rounded" />
          <div className="h-4 w-64 shimmer rounded mt-2" />
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[0, 1, 2, 3, 4, 5, 6].map((i) => <CardSkeleton key={i} />)}
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <ErrorState
        title="Ошибка загрузки"
        message="Не удалось загрузить данные RFM."
        onRetry={() => mutate()}
      />
    )
  }

  return (
    <div>
      <div className="animate-in mb-4">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          RFM
        </h1>
        <p className="font-mono text-xs text-neutral-400 uppercase tracking-wider mt-1">
          Анализ клиентской базы по Recency, Frequency, Monetary
        </p>
      </div>

      {/* Period filter */}
      <div className="relative z-20 flex items-center gap-3 flex-wrap mb-8 animate-in animate-in-delay-1">
        <PeriodFilter
          period={filter.period}
          from={filter.from}
          to={filter.to}
          onPeriodChange={(p) => setFilter((prev) => ({ ...prev, period: p }))}
          onRangeChange={(from, to) => setFilter((prev) => ({ ...prev, from, to }))}
        />
      </div>

      {segments.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 rounded bg-neutral-100 flex items-center justify-center mb-4">
            <Users className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">
            Нет данных
          </h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
            RFM-сегменты пока не рассчитаны. Настройте сегментацию в разделе Клиенты → RFM-сегменты.
          </p>
        </div>
      ) : (
        <>
          {/* Segment cards */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-8 animate-in animate-in-delay-2">
            {segments.map((seg) => {
              const color = SEGMENT_HEX_COLORS[seg.segment]
              const label = RFM_SEGMENT_LABELS[seg.segment] ?? seg.segment
              const Icon = SEGMENT_ICONS[seg.segment]
              return (
                <button
                  key={seg.segment}
                  type="button"
                  onClick={() => handleSegmentClick(seg.segment)}
                  className="bg-white rounded border border-neutral-900 p-4 text-left hover:bg-neutral-50 transition-colors group"
                >
                  <div className="flex items-center gap-2 mb-3">
                    {Icon && <Icon className="w-4 h-4 shrink-0" style={{ color }} />}
                    <span className="text-xs font-medium text-neutral-500 uppercase tracking-wide">{label}</span>
                  </div>
                  <p className="text-2xl font-bold font-mono text-neutral-900 tabular-nums tracking-tight">
                    {seg.client_count}
                  </p>
                  <div className="flex items-center justify-between mt-1">
                    <span className="text-xs text-neutral-400">{Math.round(seg.percentage)}%</span>
                    <span className="text-xs font-mono text-neutral-400 tabular-nums">
                      {formatMoney(seg.avg_check)} ср.
                    </span>
                  </div>
                </button>
              )
            })}
          </div>

          {/* Summary row */}
          <div className="flex items-center gap-6 mb-8 animate-in animate-in-delay-3">
            <div className="flex items-center gap-2 text-sm text-neutral-500">
              <Users className="w-4 h-4 text-neutral-400" />
              Всего клиентов: <span className="font-semibold text-neutral-900 font-mono">{totalClients}</span>
            </div>
            <div className="flex items-center gap-2 text-sm text-neutral-500">
              Общая выручка: <span className="font-semibold text-neutral-900 font-mono">{formatMoney(segments.reduce((s, seg) => s + seg.total_check, 0))}</span>
            </div>
          </div>

          {/* Trends chart */}
          {trends.length > 0 && (
            <section className="bg-white rounded border border-neutral-900 p-6 animate-in animate-in-delay-4">
              <h2 className="text-sm font-semibold text-neutral-700 mb-4">Динамика сегментов</h2>
              <ResponsiveContainer width="100%" height={280}>
                <AreaChart data={buildChartData(trends)}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                  <XAxis
                    dataKey="date"
                    tick={{ fontSize: 11, fill: '#a3a3a3' }}
                    axisLine={false}
                    tickLine={false}
                  />
                  <YAxis
                    tick={{ fontSize: 11, fill: '#a3a3a3' }}
                    axisLine={false}
                    tickLine={false}
                    width={40}
                    allowDecimals={false}
                  />
                  <Tooltip
                    contentStyle={{ borderRadius: 4, border: '1px solid #e5e5e5', fontSize: 13 }}
                    cursor={{ fill: '#f5f5f5' }}
                    formatter={(value: unknown, name: unknown) => [
                      String(value),
                      RFM_SEGMENT_LABELS[String(name)] ?? String(name),
                    ]}
                  />
                  {Object.keys(RFM_SEGMENT_LABELS).map((segKey) => (
                    <Area
                      key={segKey}
                      type="monotone"
                      dataKey={segKey}
                      stackId="1"
                      stroke={SEGMENT_HEX_COLORS[segKey] ?? '#9ca3af'}
                      fill={SEGMENT_HEX_COLORS[segKey] ?? '#9ca3af'}
                      fillOpacity={0.6}
                      strokeWidth={1.5}
                    />
                  ))}
                </AreaChart>
              </ResponsiveContainer>
              {/* Legend */}
              <div className="flex flex-wrap justify-center gap-3 mt-4">
                {Object.entries(RFM_SEGMENT_LABELS).map(([key, label]) => (
                  <div key={key} className="flex items-center gap-1.5 text-xs text-neutral-600">
                    <div
                      className="w-2.5 h-2.5 rounded-full"
                      style={{ backgroundColor: SEGMENT_HEX_COLORS[key] ?? '#9ca3af' }}
                    />
                    {label}
                  </div>
                ))}
              </div>
            </section>
          )}
        </>
      )}
    </div>
  )
}
