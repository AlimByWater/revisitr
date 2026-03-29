import { useState } from 'react'
import { TrendingUp, Users, DollarSign, Receipt, BarChart2 } from 'lucide-react'
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from 'recharts'
import { cn } from '@/lib/utils'
import { useSalesAnalyticsQuery } from '@/features/analytics/queries'
import { useBotsQuery } from '@/features/bots/queries'
import {
  MetricSkeleton,
  ChartSkeleton as ChartSkeletonComponent,
} from '@/components/common/LoadingSkeleton'
import { PeriodFilter } from '@/components/common/PeriodFilter'
import { CustomSelect } from '@/components/common/CustomSelect'
import type { AnalyticsFilter } from '@/features/analytics/types'

export default function SalesAnalyticsPage() {
  const [filter, setFilter] = useState<AnalyticsFilter>({ period: '30d' })
  const { data, isLoading, isError } = useSalesAnalyticsQuery(filter)
  const { data: bots } = useBotsQuery()

  return (
    <div>
      <div className="animate-in mb-4">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          Продажи
        </h1>
        <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1">
          Аналитика транзакций и выручки
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

        {bots && bots.length > 0 && (
          <CustomSelect
            value={String(filter.bot_id ?? '')}
            onChange={(v) => setFilter((prev) => ({ ...prev, bot_id: v ? Number(v) : undefined }))}
            options={[
              { value: '', label: 'Все боты' },
              ...bots.map((bot) => ({ value: String(bot.id), label: bot.name })),
            ]}
            placeholder="Все боты"
            width="200px"
          />
        )}
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6 animate-in animate-in-delay-2">
        {isLoading ? (
          <>
            <MetricSkeleton />
            <MetricSkeleton />
            <MetricSkeleton />
            <MetricSkeleton />
          </>
        ) : isError || !data ? (
          <div className="col-span-4 text-sm text-neutral-400 py-4">
            Нет данных за выбранный период
          </div>
        ) : (
          <>
            <StatCard
              label="Транзакции"
              value={data.metrics.transaction_count.toLocaleString('ru-RU')}
              icon={<BarChart2 className="w-4 h-4" />}
            />
            <StatCard
              label="Уникальных клиентов"
              value={data.metrics.unique_clients.toLocaleString('ru-RU')}
              icon={<Users className="w-4 h-4" />}
            />
            <StatCard
              label="Выручка"
              value={formatCurrency(data.metrics.total_amount)}
              icon={<DollarSign className="w-4 h-4" />}
            />
            <StatCard
              label="Средний чек"
              value={formatCurrency(data.metrics.avg_amount)}
              icon={<Receipt className="w-4 h-4" />}
            />
          </>
        )}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6 animate-in animate-in-delay-3">
        <div className="border border-neutral-900 rounded bg-white p-6">
          <h3 className="text-sm font-semibold text-neutral-700 mb-4">
            Выручка по дням
          </h3>
          {isLoading ? (
            <ChartSkeletonComponent />
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <AreaChart data={data?.charts?.revenue ?? []}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis
                  dataKey="day"
                  tickFormatter={formatDay}
                  tick={{ fontSize: 11, fill: '#a3a3a3' }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis
                  tick={{ fontSize: 11, fill: '#a3a3a3' }}
                  axisLine={false}
                  tickLine={false}
                  width={50}
                />
                <Tooltip
                  labelFormatter={(l) => formatDayFull(String(l))}
                  formatter={(v) => [formatCurrency(Number(v)), 'Выручка']}
                  contentStyle={{ borderRadius: 4, border: '1px solid #e5e5e5', fontSize: 13 }}
                  cursor={{ fill: '#f5f5f5' }}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="#EF3219"
                  fill="#EF3219"
                  fillOpacity={0.08}
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>

        <div className="border border-neutral-900 rounded bg-white p-6">
          <h3 className="text-sm font-semibold text-neutral-700 mb-4">
            Транзакции по дням
          </h3>
          {isLoading ? (
            <ChartSkeletonComponent />
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <BarChart data={data?.charts?.transactions ?? []}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis
                  dataKey="day"
                  tickFormatter={formatDay}
                  tick={{ fontSize: 11, fill: '#a3a3a3' }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis
                  tick={{ fontSize: 11, fill: '#a3a3a3' }}
                  axisLine={false}
                  tickLine={false}
                  width={30}
                  allowDecimals={false}
                />
                <Tooltip
                  labelFormatter={(l) => formatDayFull(String(l))}
                  formatter={(v) => [Number(v), 'Транзакции']}
                  contentStyle={{ borderRadius: 4, border: '1px solid #e5e5e5', fontSize: 13 }}
                  cursor={{ fill: '#f5f5f5' }}
                />
                <Bar dataKey="value" fill="#EF3219" radius={[2, 2, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      {/* Loyalty comparison */}
      {data?.comparison && (
        <div className="border border-neutral-900 rounded bg-white p-6 animate-in animate-in-delay-4">
          <h3 className="text-sm font-semibold text-neutral-700 mb-1">
            Программа лояльности
          </h3>
          <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-4">
            Средний чек: участники vs остальные
          </p>
          <div className="flex items-end gap-4">
            <ComparisonBar
              label="Участники"
              value={data.comparison.participants_avg_amount}
              max={Math.max(
                data.comparison.participants_avg_amount,
                data.comparison.non_participants_avg_amount,
              )}
              color="bg-[#EF3219]"
            />
            <ComparisonBar
              label="Без программы"
              value={data.comparison.non_participants_avg_amount}
              max={Math.max(
                data.comparison.participants_avg_amount,
                data.comparison.non_participants_avg_amount,
              )}
              color="bg-neutral-900"
            />
          </div>
        </div>
      )}

      {/* Buy frequency footer stat */}
      {data && (
        <div className="mt-4 flex items-center gap-2 animate-in animate-in-delay-5">
          <TrendingUp className="w-4 h-4 text-neutral-400" />
          <span className="text-sm text-neutral-500">
            Среднее количество покупок на клиента:&nbsp;
            <span className="font-semibold text-neutral-900 font-mono">
              {data.metrics.buy_frequency.toFixed(1)}
            </span>
          </span>
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
    <div className="border border-neutral-900 rounded p-4 bg-white">
      <div className="flex items-center gap-2 text-neutral-400 mb-3">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
      </div>
      <div className="text-3xl font-bold font-mono text-neutral-900 tracking-tight">
        {value}
      </div>
    </div>
  )
}

function ComparisonBar({
  label,
  value,
  max,
  color,
}: {
  label: string
  value: number
  max: number
  color: string
}) {
  const pct = max > 0 ? (value / max) * 100 : 0
  return (
    <div className="flex-1">
      <div className="h-16 bg-neutral-100 rounded overflow-hidden flex items-end">
        <div
          className={cn('w-full rounded transition-all duration-500', color)}
          style={{ height: `${pct}%` }}
        />
      </div>
      <p className="text-xs font-medium text-neutral-600 mt-2">{label}</p>
      <p className="font-mono text-sm font-semibold text-neutral-900">
        {formatCurrency(value)}
      </p>
    </div>
  )
}

function formatDay(str: string) {
  const d = new Date(str)
  return `${d.getDate()}.${String(d.getMonth() + 1).padStart(2, '0')}`
}

function formatDayFull(str: string) {
  return new Date(str).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('ru-RU', {
    style: 'decimal',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(Math.round(value)) + ' ₽'
}
