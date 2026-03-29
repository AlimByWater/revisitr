import { useState } from 'react'
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
import { useTheme } from '@/contexts/ThemeContext'
import { MetricSkeleton, ChartSkeleton as ChartSkeletonComponent } from '@/components/common/LoadingSkeleton'
import { PeriodFilter } from '@/components/common/PeriodFilter'
import type { DashboardFilter, DashboardMetric } from '@/features/dashboard/types'

export default function DashboardHome() {
  const [filter, setFilter] = useState<DashboardFilter>({ period: '30d' })
  const { theme } = useTheme()
  const isAurora = theme === 'aurora'

  const { data: widgets, isLoading: widgetsLoading } =
    useDashboardWidgetsQuery(filter)
  const { data: charts, isLoading: chartsLoading } =
    useDashboardChartsQuery(filter)

  const accentColor = isAurora ? '#8B5CF6' : '#E85D3A'
  const gridColor = isAurora ? 'rgba(255,255,255,0.06)' : '#f0f0f0'
  const tickColor = isAurora ? 'rgba(255,255,255,0.3)' : '#a3a3a3'
  const tooltipBg = isAurora ? 'rgba(15,11,26,0.9)' : '#fff'
  const tooltipBorder = isAurora ? 'rgba(255,255,255,0.1)' : '#e5e5e5'
  const tooltipColor = isAurora ? 'rgba(255,255,255,0.9)' : undefined

  return (
    <div>
      <div className="animate-in mb-4">
        <h1
          className="text-2xl font-bold text-neutral-900 tracking-tight mb-1"
        >
          Дашборд
        </h1>
        <p className="text-sm text-neutral-400 mt-1">
          Обзор ключевых показателей
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

      {/* Widget cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6 animate-in animate-in-delay-2">
        <MetricCard
          label="Выручка"
          metric={widgets?.revenue}
          loading={widgetsLoading}
          icon={<DollarSign className="w-4 h-4" />}
          format="currency"
          isAurora={isAurora}
        />
        <MetricCard
          label="Ср. чек"
          metric={widgets?.avg_check}
          loading={widgetsLoading}
          icon={<Receipt className="w-4 h-4" />}
          format="currency"
          isAurora={isAurora}
        />
        <MetricCard
          label="Новые клиенты"
          metric={widgets?.new_clients}
          loading={widgetsLoading}
          icon={<UserPlus className="w-4 h-4" />}
          format="number"
          isAurora={isAurora}
        />
        <MetricCard
          label="Активные"
          metric={widgets?.active_clients}
          loading={widgetsLoading}
          icon={<Users className="w-4 h-4" />}
          format="number"
          isAurora={isAurora}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-in animate-in-delay-3">

        <div
          className={cn('border border-neutral-900 rounded bg-white p-6 transition-all duration-300', isAurora && 'glass-card')}
        >
          <h3
            className="text-sm font-semibold mb-4 text-neutral-600"
          >
            Выручка по дням
          </h3>
          {chartsLoading ? (
            <ChartSkeleton />
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <AreaChart data={charts?.revenue ?? []}>
                <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
                <XAxis
                  dataKey="date"
                  tickFormatter={formatDateShort}
                  tick={{ fontSize: 11, fill: tickColor }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis
                  tick={{ fontSize: 11, fill: tickColor }}
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
                    border: `1px solid ${tooltipBorder}`,
                    fontSize: 13,
                    background: tooltipBg,
                    color: tooltipColor,
                  }}
                  labelStyle={tooltipColor ? { color: tooltipColor } : undefined}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke={accentColor}
                  fill={accentColor}
                  fillOpacity={isAurora ? 0.15 : 0.08}
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>

        <div
          className={cn('border border-neutral-900 rounded bg-white p-6 transition-all duration-300', isAurora && 'glass-card')}
        >
          <h3
            className="text-sm font-semibold mb-4 text-neutral-600"
          >
            Новые клиенты
          </h3>
          {chartsLoading ? (
            <ChartSkeleton />
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <BarChart data={charts?.new_clients ?? []}>
                <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
                <XAxis
                  dataKey="date"
                  tickFormatter={formatDateShort}
                  tick={{ fontSize: 11, fill: tickColor }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis
                  tick={{ fontSize: 11, fill: tickColor }}
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
                    border: `1px solid ${tooltipBorder}`,
                    fontSize: 13,
                    background: tooltipBg,
                    color: tooltipColor,
                  }}
                  labelStyle={tooltipColor ? { color: tooltipColor } : undefined}
                />
                <Bar dataKey="value" fill={accentColor} radius={[4, 4, 0, 0]} opacity={0.85} />
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
  isAurora,
}: {
  label: string
  metric?: DashboardMetric
  loading: boolean
  icon: React.ReactNode
  format: 'currency' | 'number'
  isAurora: boolean
}) {
  if (loading || !metric) {
    return <MetricSkeleton />
  }

  const formattedValue =
    format === 'currency' ? formatCurrency(metric.value) : Math.round(metric.value).toLocaleString('ru-RU')

  const trend = metric.trend
  const trendColor =
    trend > 0
      ? isAurora ? 'text-emerald-400' : 'text-green-600'
      : trend < 0
        ? isAurora ? 'text-red-400' : 'text-red-500'
        : isAurora ? 'text-white/30' : 'text-neutral-400'
  const TrendIcon = trend > 0 ? TrendingUp : trend < 0 ? TrendingDown : Minus

  return (
    <div
      className={cn(
        'border border-neutral-900 rounded bg-white p-5 transition-all duration-200 hover:-translate-y-0.5',
        isAurora && 'glass-card',
      )}
    >
      <div className="flex items-center gap-2 mb-3" style={{ color: 'var(--color-text-muted)' }}>
        {icon}
        <span className="text-xs font-medium uppercase tracking-wide">
          {label}
        </span>
      </div>
      <div
        className="text-3xl font-bold font-mono mb-1 tracking-tight"
        style={{ color: 'var(--color-text-primary)' }}
      >
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
  return <ChartSkeletonComponent />
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
