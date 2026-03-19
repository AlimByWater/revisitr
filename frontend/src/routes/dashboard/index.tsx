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
import { useBotsQuery } from '@/features/bots/queries'
import { useTheme } from '@/contexts/ThemeContext'
import { MetricSkeleton, ChartSkeleton as ChartSkeletonComponent } from '@/components/common/LoadingSkeleton'
import type { DashboardFilter, DashboardMetric } from '@/features/dashboard/types'

const PERIODS = [
  { value: '7d', label: '7д' },
  { value: '30d', label: '30д' },
  { value: '90d', label: '90д' },
] as const

export default function DashboardHome() {
  const [filter, setFilter] = useState<DashboardFilter>({ period: '30d' })
  const { theme } = useTheme()
  const isAurora = theme === 'aurora'

  const { data: widgets, isLoading: widgetsLoading } =
    useDashboardWidgetsQuery(filter)
  const { data: charts, isLoading: chartsLoading } =
    useDashboardChartsQuery(filter)
  const { data: bots } = useBotsQuery()

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
          className="font-display text-3xl font-bold tracking-tight mb-1"
          style={{ color: 'var(--color-text-primary)' }}
        >
          Дашборд
        </h1>
        <p className="font-mono text-xs uppercase tracking-wider" style={{ color: 'var(--color-text-muted)' }}>
          Обзор ключевых показателей
        </p>
      </div>

      {/* Filter bar */}
      <div className="flex items-center gap-3 flex-wrap mb-8">
        <div
          className="flex rounded-xl border p-1"
          style={{
            background: 'var(--color-surface-card)',
            borderColor: 'var(--color-surface-border)',
          }}
        >
          {PERIODS.map((p) => (
            <button
              key={p.value}
              onClick={() => setFilter((prev) => ({ ...prev, period: p.value }))}
              className={cn(
                'px-3 py-1.5 text-sm font-medium rounded-lg transition-all duration-200',
                filter.period === p.value
                  ? isAurora
                    ? 'bg-violet-500/20 text-violet-300'
                    : 'bg-neutral-900 text-white'
                  : isAurora
                    ? 'text-white/40 hover:text-white/70'
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
            className="rounded-xl border px-3 py-2 text-sm outline-none"
            style={{
              background: 'var(--color-surface-card)',
              borderColor: 'var(--color-surface-border)',
              color: 'var(--color-text-primary)',
            }}
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
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6 animate-in animate-in-delay-1">
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
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-in animate-in-delay-2">

        <div
          className={cn('rounded-2xl border p-6 transition-all duration-300', isAurora && 'glass-card')}
          style={{
            background: isAurora ? undefined : 'var(--color-surface-card)',
            borderColor: isAurora ? undefined : 'var(--color-surface-border)',
            boxShadow: isAurora ? undefined : 'var(--shadow-card)',
          }}
        >
          <h3
            className="text-sm font-semibold mb-4"
            style={{ color: 'var(--color-text-secondary)' }}
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
          className={cn('rounded-2xl border p-6 transition-all duration-300', isAurora && 'glass-card')}
          style={{
            background: isAurora ? undefined : 'var(--color-surface-card)',
            borderColor: isAurora ? undefined : 'var(--color-surface-border)',
            boxShadow: isAurora ? undefined : 'var(--shadow-card)',
          }}
        >
          <h3
            className="text-sm font-semibold mb-4"
            style={{ color: 'var(--color-text-secondary)' }}
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
        'rounded-2xl border p-5 transition-all duration-200 hover:-translate-y-0.5',
        isAurora && 'glass-card',
      )}
      style={{
        background: isAurora ? undefined : 'var(--color-surface-card)',
        borderColor: isAurora ? undefined : 'var(--color-surface-border)',
        boxShadow: isAurora ? undefined : 'var(--shadow-card)',
      }}
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
