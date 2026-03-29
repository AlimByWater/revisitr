import { useState } from 'react'
import { Send, Eye, MousePointerClick } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCampaignAnalyticsQuery } from '@/features/analytics/queries'
import { MetricSkeleton } from '@/components/common/LoadingSkeleton'
import { PeriodFilter } from '@/components/common/PeriodFilter'
import type { AnalyticsFilter } from '@/features/analytics/types'

export default function MailingsAnalyticsPage() {
  const [filter, setFilter] = useState<AnalyticsFilter>({ period: '30d' })
  const { data, isLoading, isError } = useCampaignAnalyticsQuery(filter)

  return (
    <div>
      <div className="animate-in mb-4">
        <h1 className="text-2xl font-bold text-neutral-900 tracking-tight mb-1">
          Рассылки
        </h1>
        <p className="text-sm text-neutral-400 mt-1">
          Эффективность рассылок
        </p>
      </div>

      {/* Filter bar */}
      <div className="relative z-20 flex items-center gap-3 flex-wrap mb-8 animate-in animate-in-delay-1">
        <PeriodFilter
          period={filter.period ?? '30d'}
          from={filter.from}
          to={filter.to}
          onPeriodChange={(p) => setFilter((prev) => ({ ...prev, period: p }))}
          onRangeChange={(from, to) => setFilter((prev) => ({ ...prev, from, to }))}
        />
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-6 animate-in animate-in-delay-2">
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
        <div className="border border-neutral-900 rounded bg-white overflow-hidden animate-in animate-in-delay-3">
          <div className="px-6 py-4 border-b border-neutral-200 flex items-center justify-between">
            <div>
              <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-0.5">
                Детализация
              </p>
              <h3 className="text-sm font-semibold text-neutral-700">По рассылкам</h3>
            </div>
            <span className="font-mono text-[11px] text-neutral-400 tabular-nums">
              {(data.by_campaign ?? []).length} рассылок
            </span>
          </div>

          {(data.by_campaign ?? []).length === 0 ? (
            <div className="px-6 py-12 text-center text-sm text-neutral-400">
              Нет рассылок за выбранный период
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left border-b border-neutral-200">
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
                  {(data.by_campaign ?? []).map((row) => (
                    <tr
                      key={row.campaign_id}
                      className="border-b border-neutral-200 last:border-0 hover:bg-neutral-50 transition-colors"
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
    <div className="border border-neutral-900 rounded bg-white p-5 transition-all duration-200 hover:-translate-y-0.5 hover:shadow-[0_6px_16px_rgba(0,0,0,0.07)]">
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
