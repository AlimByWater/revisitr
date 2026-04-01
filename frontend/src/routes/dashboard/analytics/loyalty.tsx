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
import { useLoyaltyAnalyticsQuery } from '@/features/analytics/queries'
import { useBotsQuery } from '@/features/bots/queries'
import {
  MetricSkeleton,
  ChartSkeleton as ChartSkeletonComponent,
} from '@/components/common/LoadingSkeleton'
import { PeriodFilter } from '@/components/common/PeriodFilter'
import { CustomSelect } from '@/components/common/CustomSelect'
import type { AnalyticsFilter, PieSlice } from '@/features/analytics/types'

const PIE_COLORS = ['#EF3219', '#171717', '#525252', '#a3a3a3', '#d4d4d4', '#f5f5f5']

export default function LoyaltyAnalyticsPage() {
  const [filter, setFilter] = useState<AnalyticsFilter>({ period: '30d' })
  const { data, isLoading, isError } = useLoyaltyAnalyticsQuery(filter)
  const { data: bots } = useBotsQuery()

  return (
    <div>
      <div className="animate-in mb-4">
        <h1 className="text-2xl font-bold text-neutral-900 mb-1 tracking-tight">
          Лояльность
        </h1>
        <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1">
          Клиентская база и программа лояльности
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
            light
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
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6 animate-in animate-in-delay-3">
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
          <div className="bg-white rounded border border-neutral-900 p-6 mb-6 animate-in animate-in-delay-4">
            <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-1">
              Охват лояльностью
            </p>
            <h3 className="text-sm font-semibold text-neutral-700 mb-3">
              Доля клиентов в программе лояльности
            </h3>
            <div className="flex items-center gap-4">
              <div className="flex-1 h-3 bg-neutral-100 rounded overflow-hidden">
                <div
                  className="h-full bg-[#EF3219] rounded transition-all duration-700"
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
            <div className="bg-white rounded border border-neutral-900 p-6 animate-in animate-in-delay-5">
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
                    axisLine={{ stroke: '#171717', strokeWidth: 1 }}
                    tickLine={false}
                    width={140}
                  />
                  <Tooltip
                    formatter={(v) => [Number(v).toLocaleString('ru-RU'), 'Клиентов']}
                    contentStyle={{ borderRadius: 4, border: '1px solid #e5e5e5', fontSize: 13 }}
                    cursor={{ fill: '#f5f5f5' }}
                  />
                  <Bar dataKey="count" fill="#EF3219" radius={[0, 2, 2, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </>
      )}

      {isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-in animate-in-delay-3">
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
    <div className="bg-white rounded border border-neutral-900 p-5">
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
      <div className="bg-white rounded border border-neutral-900 p-6">
        <h3 className="text-sm font-semibold text-neutral-700 mb-4">{title}</h3>
        <p className="text-sm text-neutral-400 text-center py-6">Нет данных</p>
      </div>
    )
  }

  return (
    <div className="bg-white rounded border border-neutral-900 p-6">
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
            contentStyle={{ borderRadius: 4, border: '1px solid #e5e5e5', fontSize: 12 }}
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
