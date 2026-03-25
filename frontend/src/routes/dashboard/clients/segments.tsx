import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useRFMDashboardQuery, useRFMConfigQuery, useRFMRecalculateMutation, useRFMUpdateConfigMutation } from '@/features/rfm/queries'
import { RFM_SEGMENT_LABELS, RFM_SEGMENT_COLORS } from '@/features/rfm/types'
import type { RFMConfig, RFMHistory } from '@/features/rfm/types'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'
import { RefreshCw, Settings, Users, TrendingUp, BarChart3 } from 'lucide-react'
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
  champions: '#22c55e',
  loyal: '#3b82f6',
  potential_loyalist: '#8b5cf6',
  new_customers: '#06b6d4',
  at_risk: '#f59e0b',
  cant_lose: '#ef4444',
  hibernating: '#9ca3af',
  lost: '#4b5563',
}

function buildChartData(trends: RFMHistory[]) {
  return [...trends].reverse().map((t) => ({
    date: new Date(t.calculated_at).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' }),
    ...t.distribution,
  }))
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function SegmentsPage() {
  const navigate = useNavigate()
  const { data: dashboard, isLoading, isError, mutate } = useRFMDashboardQuery()
  const recalculate = useRFMRecalculateMutation()
  const [showConfig, setShowConfig] = useState(false)

  const handleRecalculate = async () => {
    await recalculate.mutate(undefined)
    mutate()
  }

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <div className="h-8 w-48 shimmer rounded" />
          <div className="h-4 w-64 shimmer rounded mt-2" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[0, 1, 2, 3, 4, 5].map((i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <ErrorState
        title="Ошибка загрузки"
        message="Не удалось загрузить данные RFM-сегментации."
        onRetry={() => mutate()}
      />
    )
  }

  const segments = dashboard?.segments ?? []
  const trends = dashboard?.trends ?? []
  const config = dashboard?.config

  return (
    <div>
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
            RFM-сегментация
          </h1>
          <p className="text-sm text-neutral-500 mt-1">
            Анализ клиентской базы по Recency, Frequency, Monetary
          </p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setShowConfig(!showConfig)}
            className={cn(
              'flex items-center gap-1.5 py-2 px-3 rounded-lg text-sm font-medium',
              'border border-surface-border text-neutral-700',
              'hover:bg-neutral-50 transition-colors',
            )}
          >
            <Settings className="w-4 h-4" />
            Настройки
          </button>
          <button
            type="button"
            onClick={handleRecalculate}
            disabled={recalculate.isPending}
            className={cn(
              'flex items-center gap-1.5 py-2 px-4 rounded-lg text-sm font-medium',
              'bg-accent text-white hover:bg-accent-hover',
              'disabled:opacity-50 transition-all',
            )}
          >
            <RefreshCw className={cn('w-4 h-4', recalculate.isPending && 'animate-spin')} />
            {recalculate.isPending ? 'Пересчёт...' : 'Пересчитать'}
          </button>
        </div>
      </div>

      {recalculate.isSuccess && (
        <div className="mb-4 text-sm text-green-700 bg-green-50 p-3 rounded-lg animate-in">
          RFM-сегменты успешно пересчитаны.
        </div>
      )}

      {recalculate.isError && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded-lg animate-in">
          Ошибка при пересчёте. Попробуйте снова.
        </div>
      )}

      {showConfig && <RFMConfigPanel config={config} onClose={() => setShowConfig(false)} />}

      {segments.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mb-4">
            <Users className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">
            Нет данных для сегментации
          </h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed mb-4">
            Нажмите «Пересчитать» для первичного расчёта RFM-сегментов
          </p>
        </div>
      ) : (
        <>
          {/* Segment cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 mb-8">
            {segments.map((seg) => {
              const color = RFM_SEGMENT_COLORS[seg.segment] ?? { bg: 'bg-neutral-50', text: 'text-neutral-700', accent: 'bg-neutral-500' }
              const label = RFM_SEGMENT_LABELS[seg.segment] ?? seg.segment
              return (
                <div
                  key={seg.segment}
                  role="button"
                  tabIndex={0}
                  onClick={() => navigate(`/dashboard/clients?rfm_segment=${seg.segment}`)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault()
                      navigate(`/dashboard/clients?rfm_segment=${seg.segment}`)
                    }
                  }}
                  className="bg-white rounded-2xl border border-surface-border p-5 hover:shadow-sm transition-shadow cursor-pointer"
                >
                  <div className="flex items-center gap-2 mb-3">
                    <div className={cn('w-2.5 h-2.5 rounded-full', color.accent)} />
                    <h3 className={cn('text-sm font-semibold', color.text)}>{label}</h3>
                  </div>
                  <p className="text-3xl font-bold text-neutral-900 tabular-nums">
                    {seg.client_count}
                  </p>
                  <div className="flex items-center justify-between mt-2">
                    <span className="text-xs text-neutral-500">клиентов</span>
                    <span className={cn('text-sm font-medium tabular-nums', color.text)}>
                      {seg.percentage.toFixed(1)}%
                    </span>
                  </div>
                  <div className="mt-3 h-1.5 bg-neutral-100 rounded-full overflow-hidden">
                    <div
                      className={cn('h-full rounded-full transition-all', color.accent)}
                      style={{ width: `${Math.min(seg.percentage, 100)}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>

          {/* Trends chart */}
          {trends.length > 0 && (
            <section className="bg-white rounded-2xl border border-surface-border p-6">
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp className="w-5 h-5 text-neutral-400" />
                <h2 className="text-lg font-semibold text-neutral-900">Динамика сегментов</h2>
              </div>
              <ResponsiveContainer width="100%" height={300}>
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
                    contentStyle={{ borderRadius: 12, border: '1px solid #e5e5e5', fontSize: 13 }}
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
              <div className="flex flex-wrap gap-3 mt-4">
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

function RFMConfigPanel({ config, onClose }: { config?: RFMConfig | null; onClose: () => void }) {
  const { data: currentConfig } = useRFMConfigQuery()
  const updateConfig = useRFMUpdateConfigMutation()

  const cfg = currentConfig ?? config
  const [recencyDays, setRecencyDays] = useState(cfg?.recency_days ?? 365)
  const [minTransactions, setMinTransactions] = useState(cfg?.min_transactions ?? 1)

  const handleSave = async () => {
    await updateConfig.mutate({ recency_days: recencyDays, min_transactions: minTransactions })
  }

  const inputClassName = cn(
    'w-full px-3 py-2 rounded-lg border border-surface-border text-sm',
    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  )

  return (
    <div className="bg-white rounded-2xl border border-surface-border p-6 mb-6 animate-in">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-neutral-900">
          <span className="block font-mono text-[10px] uppercase tracking-widest text-neutral-400 font-normal mb-0.5">RFM</span>
          Настройки расчёта
        </h2>
        <button
          type="button"
          onClick={onClose}
          className="text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
        >
          Закрыть
        </button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
        <div>
          <label htmlFor="recency-days" className="block text-sm font-medium text-neutral-700 mb-1">
            Период анализа (дни)
          </label>
          <input
            id="recency-days"
            type="number"
            value={recencyDays}
            onChange={(e) => setRecencyDays(Number(e.target.value))}
            min={30}
            max={730}
            className={inputClassName}
          />
          <p className="text-xs text-neutral-400 mt-1">Анализировать транзакции за последние N дней</p>
        </div>
        <div>
          <label htmlFor="min-tx" className="block text-sm font-medium text-neutral-700 mb-1">
            Мин. транзакций
          </label>
          <input
            id="min-tx"
            type="number"
            value={minTransactions}
            onChange={(e) => setMinTransactions(Number(e.target.value))}
            min={1}
            max={100}
            className={inputClassName}
          />
          <p className="text-xs text-neutral-400 mt-1">Минимум транзакций для включения в расчёт</p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={handleSave}
          disabled={updateConfig.isPending}
          className={cn(
            'py-2 px-4 rounded-lg text-sm font-medium',
            'bg-accent text-white hover:bg-accent-hover',
            'disabled:opacity-50 transition-all',
          )}
        >
          {updateConfig.isPending ? 'Сохранение...' : 'Сохранить'}
        </button>
        {updateConfig.isSuccess && (
          <span className="text-sm text-green-600">Сохранено</span>
        )}
      </div>
    </div>
  )
}
