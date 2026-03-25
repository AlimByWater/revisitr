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
      <div className="mb-6">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          Предиктивная аналитика
        </h1>
        <p className="text-sm text-neutral-500 mt-1">
          Прогнозы оттока и потенциала клиентов
        </p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <SummaryCard
          icon={<Users className="w-5 h-5" />}
          label="Проанализировано"
          value={summary?.total_predicted ?? 0}
          color="text-neutral-700"
          bg="bg-neutral-50"
        />
        <SummaryCard
          icon={<AlertTriangle className="w-5 h-5" />}
          label="Высокий риск оттока"
          value={summary?.high_churn_count ?? 0}
          color="text-red-700"
          bg="bg-red-50"
        />
        <SummaryCard
          icon={<BarChart3 className="w-5 h-5" />}
          label="Средний риск оттока"
          value={`${((summary?.avg_churn_risk ?? 0) * 100).toFixed(0)}%`}
          color="text-amber-700"
          bg="bg-amber-50"
        />
        <SummaryCard
          icon={<TrendingUp className="w-5 h-5" />}
          label="Потенциал допродаж"
          value={summary?.high_upsell_count ?? 0}
          color="text-green-700"
          bg="bg-green-50"
        />
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-surface-border">
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
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mb-4">
            <BarChart3 className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">
            Нет данных для анализа
          </h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
            Предиктивная аналитика рассчитывается автоматически при наличии данных о транзакциях
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-2xl border border-surface-border overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-surface-border bg-neutral-50/50">
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    ID клиента
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Риск оттока
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Потенциал
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Прогноз LTV
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Визиты
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Тренд
                  </th>
                  <th className="text-left py-3 px-4 text-xs font-medium text-neutral-500 uppercase tracking-wider">
                    Расчёт
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {items.map((p) => (
                  <tr key={p.id} className="hover:bg-neutral-50/50 transition-colors">
                    <td className="py-3 px-4 font-medium text-neutral-900">#{p.client_id}</td>
                    <td className="py-3 px-4">
                      <span
                        className={cn(
                          'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                          riskBg(p.churn_risk),
                          riskColor(p.churn_risk),
                        )}
                      >
                        {(p.churn_risk * 100).toFixed(0)}%
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-xs text-neutral-600">
                        {(p.upsell_score * 100).toFixed(0)}%
                      </span>
                    </td>
                    <td className="py-3 px-4 tabular-nums text-neutral-700">
                      {p.predicted_value.toFixed(0)} ₽
                    </td>
                    <td className="py-3 px-4 tabular-nums text-neutral-600">
                      {p.factors.total_orders}
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-1">
                        {trendIcon(p.factors.visit_trend)}
                        <span className="text-xs text-neutral-500">
                          {TREND_LABELS[p.factors.visit_trend] ?? p.factors.visit_trend}
                        </span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-xs text-neutral-400">
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
  color,
  bg,
}: {
  icon: React.ReactNode
  label: string
  value: string | number
  color: string
  bg: string
}) {
  return (
    <div className="bg-white rounded-2xl border border-surface-border p-5">
      <div className="flex items-center gap-2 mb-3">
        <div className={cn('w-8 h-8 rounded-lg flex items-center justify-center', bg, color)}>
          {icon}
        </div>
        <span className="text-sm text-neutral-500">{label}</span>
      </div>
      <p className="text-2xl font-bold text-neutral-900 tabular-nums">{value}</p>
    </div>
  )
}
