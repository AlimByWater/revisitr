import { useState } from 'react'
import { UserPlus, Users, Star, Minus } from 'lucide-react'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import { cn } from '@/lib/utils'
import { useLoyaltyAnalyticsQuery } from '@/features/analytics/queries'
import { useBotsQuery } from '@/features/bots/queries'
import {
  MetricSkeleton,
  ChartSkeleton as ChartSkeletonComponent,
} from '@/components/common/LoadingSkeleton'
import type { AnalyticsFilter, PieSlice } from '@/features/analytics/types'

const PERIODS = [
  { value: '7d', label: '7д' },
  { value: '30d', label: '30д' },
  { value: '90d', label: '90д' },
] as const

const PIE_COLORS = ['#E85D3A', '#1a1a1a', '#888888', '#d4d4d4', '#a3a3a3', '#525252']

export default function LoyaltyAnalyticsPage() {
  const [filter, setFilter] = useState<AnalyticsFilter>({ period: '30d' })
  const { data, isLoading, isError } = useLoyaltyAnalyticsQuery(filter)
  const { data: bots } = useBotsQuery()

  return (
    <div>
      <div className="animate-in mb-4">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 mb-1 tracking-tight">
          Лояльность
        </h1>
        <p className="font-mono text-neutral-300 text-xs uppercase tracking-wider">
          Клиентская база и программа лояльности
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

      {/* Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6 animate-in animate-in-delay-1">
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
              label="Новые клиенты"
              value={data.new_clients.toLocaleString('ru-RU')}
              icon={<UserPlus className="w-4 h-4" />}
            />
            <StatCard
              label="Активные"
              value={data.active_clients.toLocaleString('ru-RU')}
              icon={<Users className="w-4 h-4" />}
            />
            <StatCard
              label="Начислено бонусов"
              value={Math.round(data.bonus_earned).toLocaleString('ru-RU')}
              icon={<Star className="w-4 h-4" />}
            />
            <StatCard
              label="Списано бонусов"
              value={Math.round(data.bonus_spent).toLocaleString('ru-RU')}
              icon={<Minus className="w-4 h-4" />}
            />
          </>
        )}
      </div>

      {/* Demographics + Funnel */}
      {!isLoading && !isError && data && (
        <>
          {/* Demographics */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6 animate-in animate-in-delay-2">
            <PieCard
              title="По полу"
              slices={data.demographics.by_gender}
            />
            <PieCard
              title="По возрасту"
              slices={data.demographics.by_age_group}
            />
            <PieCard
              title="По платформе"
              slices={data.demographics.by_os}
            />
          </div>

          {/* Loyalty % */}
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-6 animate-in animate-in-delay-2">
            <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-1">
              Охват лояльностью
            </p>
            <h3 className="text-sm font-semibold text-neutral-700 mb-3">
              Доля клиентов в программе лояльности
            </h3>
            <div className="flex items-center gap-4">
              <div className="flex-1 h-3 bg-neutral-100 rounded-full overflow-hidden">
                <div
                  className="h-full bg-accent rounded-full transition-all duration-700"
                  style={{ width: `${data.demographics.loyalty_percent}%` }}
                />
              </div>
              <span className="font-mono text-xl font-bold text-neutral-900 tabular-nums shrink-0">
                {data.demographics.loyalty_percent.toFixed(1)}%
              </span>
            </div>
          </div>

          {/* Bot funnel */}
          {data.bot_funnel.length > 0 && (
            <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 animate-in animate-in-delay-3">
              <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-0.5">
                Воронка
              </p>
              <h3 className="text-sm font-semibold text-neutral-700 mb-4">
                Шаги взаимодействия с ботом
              </h3>
              <ResponsiveContainer width="100%" height={Math.max(200, data.bot_funnel.length * 48)}>
                <BarChart
                  data={data.bot_funnel}
                  layout="vertical"
                  margin={{ left: 8, right: 32, top: 0, bottom: 0 }}
                >
                  <XAxis type="number" tick={{ fontSize: 11, fill: '#a3a3a3' }} axisLine={false} tickLine={false} />
                  <YAxis
                    type="category"
                    dataKey="step"
                    tick={{ fontSize: 12, fill: '#525252' }}
                    axisLine={false}
                    tickLine={false}
                    width={100}
                  />
                  <Tooltip
                    formatter={(v) => [Number(v).toLocaleString('ru-RU'), 'Клиентов']}
                    contentStyle={{ borderRadius: 12, border: '1px solid #e5e5e5', fontSize: 13 }}
                  />
                  <Bar dataKey="count" fill="#E85D3A" radius={[0, 4, 4, 0]} opacity={0.85} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </>
      )}

      {isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-in animate-in-delay-2">
          <ChartSkeletonComponent />
          <ChartSkeletonComponent />
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
      <div className="text-3xl font-bold font-mono text-neutral-900 tracking-tight">
        {value}
      </div>
    </div>
  )
}

function PieCard({ title, slices }: { title: string; slices: PieSlice[] }) {
  if (!slices || slices.length === 0) {
    return (
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
        <h3 className="text-sm font-semibold text-neutral-700 mb-4">{title}</h3>
        <p className="text-sm text-neutral-400 text-center py-6">Нет данных</p>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
      <h3 className="text-sm font-semibold text-neutral-700 mb-4">{title}</h3>
      <ResponsiveContainer width="100%" height={160}>
        <PieChart>
          <Pie
            data={slices}
            dataKey="value"
            nameKey="label"
            cx="50%"
            cy="50%"
            outerRadius={60}
            strokeWidth={0}
          >
            {slices.map((_, i) => (
              <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
            ))}
          </Pie>
          <Tooltip
            formatter={(v, name) => [Number(v).toLocaleString('ru-RU'), name]}
            contentStyle={{ borderRadius: 12, border: '1px solid #e5e5e5', fontSize: 12 }}
          />
        </PieChart>
      </ResponsiveContainer>
      <div className="space-y-1 mt-2">
        {slices.map((slice, i) => (
          <div key={i} className="flex items-center justify-between text-xs">
            <div className="flex items-center gap-1.5">
              <div
                className="w-2 h-2 rounded-full shrink-0"
                style={{ background: PIE_COLORS[i % PIE_COLORS.length] }}
              />
              <span className="text-neutral-600 truncate max-w-[80px]">{slice.label}</span>
            </div>
            <span className="font-mono text-neutral-500 tabular-nums">
              {slice.percent.toFixed(0)}%
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
