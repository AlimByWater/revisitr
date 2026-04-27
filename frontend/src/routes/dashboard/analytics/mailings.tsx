import { useMemo, useState } from 'react'
import { Send, Eye, MousePointerClick, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCampaignAnalyticsQuery } from '@/features/analytics/queries'
import { MetricSkeleton } from '@/components/common/LoadingSkeleton'
import { PeriodFilter } from '@/components/common/PeriodFilter'
import type { AnalyticsFilter } from '@/features/analytics/types'

type SortKey = 'campaign_name' | 'sent' | 'open_rate' | 'conversions'

export default function MailingsAnalyticsPage() {
  const [filter, setFilter] = useState<AnalyticsFilter>({ period: '30d' })
  const [sortBy, setSortBy] = useState<SortKey | undefined>(undefined)
  const [sortDesc, setSortDesc] = useState(false)
  const { data, isLoading, isError } = useCampaignAnalyticsQuery(filter)

  function handleSort(key: SortKey) {
    if (sortBy === key) {
      if (sortDesc) {
        setSortBy(undefined)
        setSortDesc(false)
      } else {
        setSortDesc(true)
      }
    } else {
      setSortBy(key)
      setSortDesc(false)
    }
  }

  function getSortIcon(key: SortKey) {
    if (sortBy !== key) return <ArrowUpDown className="w-3 h-3 opacity-40" />
    return sortDesc ? <ArrowDown className="w-3 h-3" /> : <ArrowUp className="w-3 h-3" />
  }

  const sortedCampaigns = useMemo(() => {
    const rows = [...(data?.by_campaign ?? [])]
    if (!sortBy) return rows
    const dir = sortDesc ? -1 : 1
    return rows.sort((a, b) => {
      const av = a[sortBy]
      const bv = b[sortBy]
      if (typeof av === 'number' && typeof bv === 'number') return (av - bv) * dir
      return String(av).localeCompare(String(bv), 'ru') * dir
    })
  }, [data?.by_campaign, sortBy, sortDesc])

  return (
    <div>
      <div className="animate-in mb-4">
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight mb-1">
          Рассылки
        </h1>
        <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">
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
              value={data.total_opened.toLocaleString('ru-RU')}
              subValue={`${data.open_rate.toFixed(1)}% открытий`}
              icon={<Eye className="w-4 h-4 text-neutral-400" />}
            />
            <StatCard
              label="Конверсии"
              value={data.conversions.toLocaleString('ru-RU')}
              subValue={`${data.conv_rate.toFixed(1)}% конверсий`}
              icon={<MousePointerClick className="w-4 h-4 text-neutral-400" />}
            />
          </>
        )}
      </div>

      {/* Campaign table */}
      {!isLoading && !isError && data && (
        <div className="border border-neutral-900 rounded bg-white overflow-hidden animate-in animate-in-delay-3">
          <div className="px-6 py-4 border-b border-neutral-200 flex items-center justify-between gap-4 flex-wrap">
            <div>
              <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-0.5">
                Детализация
              </p>
              <h3 className="text-sm font-semibold text-neutral-700">По рассылкам</h3>
            </div>
            <span className="font-mono text-[11px] text-neutral-400 tabular-nums">
              {sortedCampaigns.length}
            </span>
          </div>

          {sortedCampaigns.length === 0 ? (
            <div className="px-6 py-12 text-center text-sm text-neutral-400">
              Нет рассылок за выбранный период
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-neutral-200">
                    <SortableTh
                      label="Рассылка"
                      align="left"
                      active={sortBy === 'campaign_name'}
                      icon={getSortIcon('campaign_name')}
                      onClick={() => handleSort('campaign_name')}
                    />
                    <SortableTh
                      label="Отправлено"
                      align="center"
                      active={sortBy === 'sent'}
                      icon={getSortIcon('sent')}
                      onClick={() => handleSort('sent')}
                    />
                    <SortableTh
                      label="Открываемость"
                      align="center"
                      active={sortBy === 'open_rate'}
                      icon={getSortIcon('open_rate')}
                      onClick={() => handleSort('open_rate')}
                    />
                    <SortableTh
                      label="Конверсии"
                      align="center"
                      active={sortBy === 'conversions'}
                      icon={getSortIcon('conversions')}
                      onClick={() => handleSort('conversions')}
                    />
                  </tr>
                </thead>
                <tbody>
                  {sortedCampaigns.map((row) => (
                    <tr
                      key={row.campaign_id}
                      className="border-b border-neutral-200 last:border-0 hover:bg-neutral-50 transition-colors"
                    >
                      <td className="px-6 py-4 text-neutral-900 font-medium">
                        {row.campaign_name}
                      </td>
                      <td className="px-6 py-4 text-center font-mono tabular-nums text-neutral-700">
                        {row.sent.toLocaleString('ru-RU')}
                      </td>
                      <td className="px-6 py-4 text-center">
                        <RateBadge value={row.open_rate} />
                      </td>
                      <td className="px-6 py-4 text-center font-mono tabular-nums text-neutral-700">
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

function SortableTh({
  label,
  align,
  active,
  icon,
  onClick,
}: {
  label: string
  align: 'left' | 'center' | 'right'
  active: boolean
  icon: React.ReactNode
  onClick: () => void
}) {
  const alignClass = align === 'left' ? 'text-left' : align === 'center' ? 'text-center' : 'text-right'
  return (
    <th className={cn('px-6 py-3 font-medium font-mono text-[11px] uppercase tracking-wider', alignClass)}>
      <button
        type="button"
        onClick={onClick}
        className={cn(
          'inline-flex items-center gap-1 hover:text-neutral-900 transition-colors cursor-pointer',
          active ? 'text-neutral-900' : 'text-neutral-500',
        )}
      >
        {label}
        {icon}
      </button>
    </th>
  )
}

function StatCard({
  label,
  value,
  subValue,
  icon,
}: {
  label: string
  value: string
  subValue?: string
  icon: React.ReactNode
}) {
  return (
    <div className="border border-neutral-900 rounded bg-white p-5">
      <div className="flex items-center gap-2 text-neutral-400 mb-3">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
      </div>
      <div className="font-mono text-3xl font-bold text-neutral-900 tracking-tight leading-tight">
        {value}
      </div>
      {subValue && (
        <p className="text-xs text-neutral-400 mt-1">{subValue}</p>
      )}
    </div>
  )
}

function RateBadge({ value }: { value: number }) {
  const color =
    value >= 50
      ? 'bg-emerald-100 text-emerald-700 border border-emerald-500/30'
      : value >= 25
        ? 'bg-orange-100 text-orange-700 border border-orange-500/30'
        : value >= 10
          ? 'bg-amber-50 text-amber-700 border border-amber-400/30'
          : 'bg-neutral-100 text-neutral-500 border border-neutral-300'
  return (
    <span
      className={cn(
        'inline-block font-mono text-[11px] font-semibold px-2 py-0.5 rounded tabular-nums',
        color,
      )}
    >
      {value.toFixed(1)}%
    </span>
  )
}
