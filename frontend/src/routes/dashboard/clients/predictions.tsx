import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  usePredictionsQuery,
  usePredictionSummaryQuery,
  useHighChurnQuery,
} from '@/features/segments/queries'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'
import { AlertTriangle, TrendingUp, TrendingDown, Users, BarChart3 } from 'lucide-react'

function riskColor(risk: number): string {
  if (risk >= 0.7) return 'text-red-600'
  if (risk >= 0.4) return 'text-amber-600'
  return 'text-green-600'
}

function riskBg(risk: number): string {
  if (risk >= 0.7) return 'bg-red-50'
  if (risk >= 0.4) return 'bg-amber-50'
  return 'bg-green-50'
}

function trendIcon(trend: string) {
  if (trend === 'increasing') return <TrendingUp className="w-3.5 h-3.5 text-green-500" />
  if (trend === 'declining') return <TrendingDown className="w-3.5 h-3.5 text-red-500" />
  return <span className="text-xs text-neutral-400">—</span>
}

const TREND_LABELS: Record<string, string> = {
  increasing: 'Растёт',
  stable: 'Стабильно',
  declining: 'Снижается',
}

export default function PredictionsPage() {
  const { data: summary, isLoading: summaryLoading, isError: summaryError, mutate: mutateSummary } =
    usePredictionSummaryQuery()
  const [tab, setTab] = useState<'all' | 'high-churn'>('all')
  const { data: predictions, isLoading: predsLoading } = usePredictionsQuery(50, 0)
  const { data: highChurn, isLoading: churnLoading } = useHighChurnQuery(0.7)

  const isLoading = summaryLoading || predsLoading
  const items = tab === 'high-churn' ? highChurn : predictions

  if (isLoading) {
    return (
      <div>
        <div className="mb-6">
          <div className="h-8 w-48 shimmer rounded" />
          <div className="h-4 w-64 shimmer rounded mt-2" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[0, 1, 2, 3].map((i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      </div>
    )
  }

  if (summaryError) {
    return (
      <ErrorState
        title="Ошибка загрузки"
        message="Не удалось загрузить предиктивную аналитику."
        onRetry={() => mutateSummary()}
      />
    )
  }

  return (
    <div>
      <div className="mb-6 animate-in">
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">
          Предиктивная аналитика
        </h1>
        <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">
          Прогнозы оттока и потенциала клиентов
        </p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8 animate-in animate-in-delay-1">
        <SummaryCard
          icon={<Users className="w-4 h-4" />}
          label="Проанализировано"
          value={summary?.total_predicted ?? 0}
        />
        <SummaryCard
          icon={<AlertTriangle className="w-4 h-4" />}
          label="Высокий риск оттока"
          value={summary?.high_churn_count ?? 0}
        />
        <SummaryCard
          icon={<BarChart3 className="w-4 h-4" />}
          label="Средний риск оттока"
          value={`${((summary?.avg_churn_risk ?? 0) * 100).toFixed(0)}%`}
        />
        <SummaryCard
          icon={<TrendingUp className="w-4 h-4" />}
          label="Потенциал допродаж"
          value={summary?.high_upsell_count ?? 0}
        />
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-neutral-200">
        <button
          type="button"
          onClick={() => setTab('all')}
          className={cn(
            'px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors',
            tab === 'all'
              ? 'border-accent text-accent'
              : 'border-transparent text-neutral-500 hover:text-neutral-700',
          )}
        >
          Все клиенты
        </button>
        <button
          type="button"
          onClick={() => setTab('high-churn')}
          className={cn(
            'px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors',
            tab === 'high-churn'
              ? 'border-red-500 text-red-600'
              : 'border-transparent text-neutral-500 hover:text-neutral-700',
          )}
        >
          Риск оттока
          {(summary?.high_churn_count ?? 0) > 0 && (
            <span className="ml-1.5 text-xs bg-red-100 text-red-600 rounded-full px-1.5 py-0.5">
              {summary?.high_churn_count}
            </span>
          )}
        </button>
      </div>

      {/* Predictions table */}
      {!items || items.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <BarChart3 className="w-8 h-8 text-neutral-400 mb-4" />
          <h3 className="font-display text-xl font-bold text-neutral-800 mb-1.5">
            Нет данных для анализа
          </h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
            Предиктивная аналитика рассчитывается автоматически при наличии данных о транзакциях
          </p>
        </div>
      ) : (
        <div className="bg-white rounded border border-neutral-900 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-200 bg-neutral-50/50">
                  <th className="text-center py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    ID клиента
                  </th>
                  <th className="text-center py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Риск оттока
                  </th>
                  <th className="text-center py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Потенциал
                  </th>
                  <th className="text-center py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Прогноз LTV
                  </th>
                  <th className="text-center py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Визиты
                  </th>
                  <th className="text-center py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Тренд
                  </th>
                  <th className="text-center py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Расчёт
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-200">
                {items.map((p) => (
                  <tr
                    key={p.id}
                    className="hover:bg-neutral-50/50 transition-colors cursor-pointer"
                    onClick={() => window.open(`/dashboard/clients/${p.client_id}`, '_blank')}
                  >
                    <td className="py-3 px-4 text-center font-medium text-neutral-900">#{p.client_id}</td>
                    <td className="py-3 px-4 text-center">
                      <span
                        className={cn(
                          'font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border',
                          p.churn_risk >= 0.7
                            ? 'bg-red-500/10 text-red-700 border-red-500/30'
                            : p.churn_risk >= 0.4
                              ? 'bg-amber-500/10 text-amber-700 border-amber-500/30'
                              : 'bg-emerald-500/10 text-emerald-700 border-emerald-500/30',
                        )}
                      >
                        {(p.churn_risk * 100).toFixed(0)}%
                      </span>
                    </td>
                    <td className="py-3 px-4 text-center text-xs text-neutral-600 font-mono tabular-nums">
                      {(p.upsell_score * 100).toFixed(0)}%
                    </td>
                    <td className="py-3 px-4 text-center tabular-nums text-neutral-700 font-mono">
                      {p.predicted_value.toFixed(0)} ₽
                    </td>
                    <td className="py-3 px-4 text-center tabular-nums text-neutral-600 font-mono">
                      {p.factors.total_orders}
                    </td>
                    <td className="py-3 px-4 text-center">
                      <div className="inline-flex items-center gap-1">
                        {trendIcon(p.factors.visit_trend)}
                        <span className="text-xs text-neutral-500">
                          {TREND_LABELS[p.factors.visit_trend] ?? p.factors.visit_trend}
                        </span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-center text-xs text-neutral-400 font-mono">
                      {new Date(p.computed_at).toLocaleDateString('ru-RU')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

function SummaryCard({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode
  label: string
  value: string | number
}) {
  return (
    <div className="bg-white rounded border border-neutral-900 p-5">
      <div className="flex items-center gap-2 text-neutral-400 mb-3">
        <span className="text-neutral-400">{icon}</span>
        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
      </div>
      <p className="text-2xl font-bold font-mono text-neutral-900 tabular-nums">{value}</p>
    </div>
  )
}
