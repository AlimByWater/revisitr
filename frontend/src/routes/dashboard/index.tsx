import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import {
  TrendingUp,
  TrendingDown,
  Minus,
  Users,
  UserPlus,
  DollarSign,
  Receipt,
} from 'lucide-react'
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
import {
  useDashboardWidgetsQuery,
  useDashboardChartsQuery,
} from '@/features/dashboard/queries'
import { useBotsQuery } from '@/features/bots/queries'
import type { DashboardFilter, DashboardMetric } from '@/features/dashboard/types'

export const Route = createFileRoute('/dashboard/')({
  component: DashboardHome,
})

const PERIODS = [
  { value: '7d', label: '7д' },
  { value: '30d', label: '30д' },
  { value: '90d', label: '90д' },
] as const

function DashboardHome() {
  const [filter, setFilter] = useState<DashboardFilter>({ period: '30d' })

  const { data: widgets, isLoading: widgetsLoading } =
    useDashboardWidgetsQuery(filter)
  const { data: charts, isLoading: chartsLoading } =
    useDashboardChartsQuery(filter)
  const { data: bots } = useBotsQuery()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-neutral-900 mb-1">Дашборд</h1>
        <p className="text-neutral-500 text-sm">
          Обзор ключевых показателей
        </p>
      </div>

      {/* Filter bar */}
      <div className="flex items-center gap-3 flex-wrap">
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

        {bots && bots.length > 0 && (
          <select
            value={filter.bot_id ?? ''}
            onChange={(e) =>
              setFilter((prev) => ({
                ...prev,
                bot_id: e.target.value ? Number(e.target.value) : undefined,
              }))
            }
            className="bg-white rounded-xl border border-surface-border px-3 py-2 text-sm text-neutral-700 outline-none"
          >
            <option value="">Все боты</option>
            {bots.map((bot) => (
              <option key={bot.id} value={bot.id}>
                {bot.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* Widget cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <MetricCard
          label="Выручка"
          metric={widgets?.revenue}
          loading={widgetsLoading}
          icon={<DollarSign className="w-4 h-4" />}
          format="currency"
        />
        <MetricCard
          label="Ср. чек"
          metric={widgets?.avg_check}
          loading={widgetsLoading}
          icon={<Receipt className="w-4 h-4" />}
          format="currency"
        />
        <MetricCard
          label="Новые клиенты"
          metric={widgets?.new_clients}
          loading={widgetsLoading}
          icon={<UserPlus className="w-4 h-4" />}
          format="number"
        />
        <MetricCard
          label="Активные"
          metric={widgets?.active_clients}
          loading={widgetsLoading}
          icon={<Users className="w-4 h-4" />}
          format="number"
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
          <h3 className="text-sm font-semibold text-neutral-700 mb-4">
            Выручка по дням
          </h3>
          {chartsLoading ? (
            <ChartSkeleton />
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <AreaChart data={charts?.revenue ?? []}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis
                  dataKey="date"
                  tickFormatter={formatDateShort}
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
                  labelFormatter={(label) => formatDateFull(String(label))}
                  formatter={(value) => [
                    formatCurrency(Number(value)),
                    'Выручка',
                  ]}
                  contentStyle={{
                    borderRadius: 12,
                    border: '1px solid #e5e5e5',
                    fontSize: 13,
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="#171717"
                  fill="#f5f5f5"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>

        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
          <h3 className="text-sm font-semibold text-neutral-700 mb-4">
            Новые клиенты
          </h3>
          {chartsLoading ? (
            <ChartSkeleton />
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <BarChart data={charts?.new_clients ?? []}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis
                  dataKey="date"
                  tickFormatter={formatDateShort}
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
                  labelFormatter={(label) => formatDateFull(String(label))}
                  formatter={(value) => [Number(value), 'Клиенты']}
                  contentStyle={{
                    borderRadius: 12,
                    border: '1px solid #e5e5e5',
                    fontSize: 13,
                  }}
                />
                <Bar dataKey="value" fill="#171717" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </div>
  )
}

function MetricCard({
  label,
  metric,
  loading,
  icon,
  format,
}: {
  label: string
  metric?: DashboardMetric
  loading: boolean
  icon: React.ReactNode
  format: 'currency' | 'number'
}) {
  if (loading || !metric) {
    return (
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 animate-pulse">
        <div className="h-4 bg-neutral-100 rounded w-20 mb-3" />
        <div className="h-7 bg-neutral-100 rounded w-24 mb-2" />
        <div className="h-4 bg-neutral-100 rounded w-16" />
      </div>
    )
  }

  const formattedValue =
    format === 'currency' ? formatCurrency(metric.value) : Math.round(metric.value).toLocaleString('ru-RU')

  const trend = metric.trend
  const trendColor =
    trend > 0 ? 'text-green-600' : trend < 0 ? 'text-red-500' : 'text-neutral-400'
  const TrendIcon = trend > 0 ? TrendingUp : trend < 0 ? TrendingDown : Minus

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
      <div className="flex items-center gap-2 text-neutral-400 mb-3">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wide">
          {label}
        </span>
      </div>
      <div className="text-2xl font-bold text-neutral-900 mb-1">
        {formattedValue}
      </div>
      <div className={cn('flex items-center gap-1 text-sm', trendColor)}>
        <TrendIcon className="w-3.5 h-3.5" />
        <span className="font-medium">
          {trend > 0 ? '+' : ''}
          {trend.toFixed(1)}%
        </span>
      </div>
    </div>
  )
}

function ChartSkeleton() {
  return (
    <div className="h-[240px] bg-neutral-50 rounded-xl animate-pulse flex items-center justify-center">
      <span className="text-neutral-300 text-sm">Загрузка...</span>
    </div>
  )
}

function formatDateShort(dateStr: string) {
  const date = new Date(dateStr)
  return `${date.getDate()}.${String(date.getMonth() + 1).padStart(2, '0')}`
}

function formatDateFull(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleDateString('ru-RU', {
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
  }).format(Math.round(value))
}
