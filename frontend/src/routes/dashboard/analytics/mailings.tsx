import { useState } from 'react'
import { Send, Eye, MousePointerClick } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCampaignAnalyticsQuery } from '@/features/analytics/queries'
import { MetricSkeleton } from '@/components/common/LoadingSkeleton'
import type { AnalyticsFilter } from '@/features/analytics/types'

const PERIODS = [
  { value: '7d', label: '7д' },
  { value: '30d', label: '30д' },
  { value: '90d', label: '90д' },
] as const

export default function MailingsAnalyticsPage() {
  const [filter, setFilter] = useState<AnalyticsFilter>({ period: '30d' })
  const { data, isLoading, isError } = useCampaignAnalyticsQuery(filter)

  return (
    <div>
      <div className="animate-in mb-4">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 mb-1 tracking-tight">
          Рассылки
        </h1>
        <p className="font-mono text-neutral-300 text-xs uppercase tracking-wider">
          Эффективность рассылок
        </p>
      </div>

      {/* Filter bar */}
      <div className="flex items-center gap-3 flex-wrap mb-8">
        <div className="flex bg-white rounded-xl border border-surface-border p-1">
          {PERIODS.map((p) => (
            <button
              key={p.value}
              onClick={() => setFilter((prev) => ({ ...prev, period: p.value }))}
              className={cn(
                'px-3 py-1.5 text-sm font-medium rounded-lg transition-colors',
                filter.period === p.value
                  ? 'bg-neutral-900 text-white'
                  : 'text-neutral-500 hover:text-neutral-700',
              )}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-6 animate-in animate-in-delay-1">
        {isLoading ? (
          <>
            <MetricSkeleton />
            <MetricSkeleton />
            <MetricSkeleton />
          </>
        ) : isError || !data ? (
          <div className="col-span-3 text-sm text-neutral-400 py-4">
            Нет данных за выбранный период
          </div>
        ) : (
          <>
            <StatCard
              label="Отправлено"
              value={data.total_sent.toLocaleString('ru-RU')}
              icon={<Send className="w-4 h-4" />}
            />
            <StatCard
              label="Открыто"
              value={`${data.total_opened.toLocaleString('ru-RU')} (${data.open_rate.toFixed(1)}%)`}
              icon={<Eye className="w-4 h-4" />}
            />
            <StatCard
              label="Конверсии"
              value={`${data.conversions.toLocaleString('ru-RU')} (${data.conv_rate.toFixed(1)}%)`}
              icon={<MousePointerClick className="w-4 h-4" />}
            />
          </>
        )}
      </div>

      {/* Campaign table */}
      {!isLoading && !isError && data && (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden animate-in animate-in-delay-2">
          <div className="px-6 py-4 border-b border-surface-border flex items-center justify-between">
            <div>
              <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-0.5">
                Детализация
              </p>
              <h3 className="text-sm font-semibold text-neutral-700">По рассылкам</h3>
            </div>
            <span className="font-mono text-[11px] text-neutral-400 tabular-nums">
              {data.by_campaign.length} рассылок
            </span>
          </div>

          {data.by_campaign.length === 0 ? (
            <div className="px-6 py-12 text-center text-sm text-neutral-400">
              Нет рассылок за выбранный период
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left border-b border-surface-border">
                    <th className="px-6 py-3 font-medium text-neutral-500 font-mono text-[11px] uppercase tracking-wider">
                      Рассылка
                    </th>
                    <th className="px-6 py-3 font-medium text-neutral-500 font-mono text-[11px] uppercase tracking-wider text-right">
                      Отправлено
                    </th>
                    <th className="px-6 py-3 font-medium text-neutral-500 font-mono text-[11px] uppercase tracking-wider text-right">
                      Открываемость
                    </th>
                    <th className="px-6 py-3 font-medium text-neutral-500 font-mono text-[11px] uppercase tracking-wider text-right">
                      Конверсии
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {data.by_campaign.map((row) => (
                    <tr
                      key={row.campaign_id}
                      className="border-b border-surface-border last:border-0 hover:bg-neutral-50 transition-colors"
                    >
                      <td className="px-6 py-4 text-neutral-900 font-medium">
                        {row.campaign_name}
                      </td>
                      <td className="px-6 py-4 text-right font-mono tabular-nums text-neutral-700">
                        {row.sent.toLocaleString('ru-RU')}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <RateBadge value={row.open_rate} />
                      </td>
                      <td className="px-6 py-4 text-right font-mono tabular-nums text-neutral-700">
                        {row.conversions.toLocaleString('ru-RU')}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function StatCard({
  label,
  value,
  icon,
}: {
  label: string
  value: string
  icon: React.ReactNode
}) {
  return (
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-5 transition-all duration-200 hover:-translate-y-0.5 hover:shadow-[0_6px_16px_rgba(0,0,0,0.07)]">
      <div className="flex items-center gap-2 text-neutral-400 mb-3">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
      </div>
      <div className="text-2xl font-bold font-mono text-neutral-900 tracking-tight leading-tight">
        {value}
      </div>
    </div>
  )
}

function RateBadge({ value }: { value: number }) {
  const color =
    value >= 30
      ? 'bg-green-100 text-green-700'
      : value >= 15
        ? 'bg-yellow-100 text-yellow-700'
        : 'bg-neutral-100 text-neutral-500'
  return (
    <span
      className={cn(
        'inline-block font-mono text-xs font-semibold px-2 py-0.5 rounded-md tabular-nums',
        color,
      )}
    >
      {value.toFixed(1)}%
    </span>
  )
}
