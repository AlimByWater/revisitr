import { useNavigate, useSearchParams } from 'react-router-dom'
import { useState, useMemo } from 'react'
import { cn } from '@/lib/utils'
import { useClientsQuery, useClientStatsQuery } from '@/features/clients/queries'
import { useBotsQuery } from '@/features/bots/queries'
import { RFM_SEGMENT_LABELS, RFM_SEGMENT_COLORS } from '@/features/rfm/types'
import { CustomSelect } from '@/components/common/CustomSelect'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton, MetricSkeleton } from '@/components/common/LoadingSkeleton'
import { Pagination } from '@/components/common/Pagination'
import type { ClientProfile, ClientFilter } from '@/features/clients/types'
import {
  Users,
  Search,
  Wallet,
  UserPlus,
  Activity,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  X,
} from 'lucide-react'

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function formatBalance(value: number): string {
  return value.toLocaleString('ru-RU')
}

interface Column {
  id: string
  header: string
  render: (row: ClientProfile) => React.ReactNode
}

const columns: Column[] = [
  {
    id: 'name',
    header: 'Имя',
    render: (row) => (
      <span className="font-medium text-neutral-900">
        {[row.first_name, row.last_name].filter(Boolean).join(' ')}
      </span>
    ),
  },
  {
    id: 'bot_name',
    header: 'Бот',
    render: (row) => <span className="text-neutral-600">{row.bot_name}</span>,
  },
  {
    id: 'loyalty_level',
    header: 'Уровень',
    render: (row) =>
      row.loyalty_level ? (
        <span className="font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border bg-neutral-100 text-neutral-600 border-neutral-300">
          {row.loyalty_level}
        </span>
      ) : (
        <span className="text-neutral-400">—</span>
      ),
  },
  {
    id: 'loyalty_balance',
    header: 'Баланс',
    render: (row) => (
      <span className="font-mono font-medium text-neutral-900 tabular-nums">
        {formatBalance(row.loyalty_balance)}
      </span>
    ),
  },
  {
    id: 'purchase_count',
    header: 'Покупки',
    render: (row) => (
      <span className="font-mono text-neutral-600 tabular-nums">{row.purchase_count}</span>
    ),
  },
  {
    id: 'rfm_segment',
    header: 'RFM сегмент',
    render: (row) => {
      if (!row.rfm_segment) return <span className="text-neutral-400">--</span>
      const color = RFM_SEGMENT_COLORS[row.rfm_segment]
      const label = RFM_SEGMENT_LABELS[row.rfm_segment] ?? row.rfm_segment
      return (
        <span
          className={cn(
            'font-mono text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded',
            color?.bg ?? 'bg-neutral-100',
            color?.text ?? 'text-neutral-700',
          )}
        >
          {label}
        </span>
      )
    },
  },
  {
    id: 'registered_at',
    header: 'Зарегистрирован',
    render: (row) => (
      <span className="font-mono text-neutral-500 tabular-nums">{formatDate(row.registered_at)}</span>
    ),
  },
]

function StatCard({
  label,
  value,
  icon: Icon,
}: {
  label: string
  value: string | number
  icon: React.ComponentType<{ className?: string }>
}) {
  return (
    <div className="bg-white rounded border border-neutral-900 p-5">
      <div className="flex items-center gap-2 text-neutral-400 mb-3">
        <Icon className="w-4 h-4" />
        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
      </div>
      <p className="text-3xl font-bold font-mono text-neutral-900 tracking-tight tabular-nums">{value}</p>
    </div>
  )
}

export default function ClientsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const rfmSegment = searchParams.get('rfm_segment') ?? undefined
  const { data: bots } = useBotsQuery()
  const { data: stats } = useClientStatsQuery()

  const [search, setSearch] = useState('')
  const [botFilter, setBotFilter] = useState<number | undefined>(undefined)
  const [sortBy, setSortBy] = useState<string | undefined>(undefined)
  const [sortDesc, setSortDesc] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const pageSize = 20

  const clearRfmFilter = () => {
    setSearchParams((prev) => {
      prev.delete('rfm_segment')
      return prev
    })
    setPageIndex(0)
  }

  const filter: ClientFilter = useMemo(
    () => ({
      search: search || undefined,
      bot_id: botFilter,
      segment: rfmSegment,
      sort_by: sortBy,
      sort_order: sortBy ? (sortDesc ? 'desc' : 'asc') : undefined,
      limit: pageSize,
      offset: pageIndex * pageSize,
    }),
    [search, botFilter, rfmSegment, sortBy, sortDesc, pageIndex],
  )

  const { data, isLoading, isError, mutate } = useClientsQuery(filter)

  const pageCount = data ? Math.ceil(data.total / pageSize) : 0

  function handleSort(columnId: string) {
    if (sortBy === columnId) {
      if (sortDesc) {
        setSortBy(undefined)
        setSortDesc(false)
      } else {
        setSortDesc(true)
      }
    } else {
      setSortBy(columnId)
      setSortDesc(false)
    }
    setPageIndex(0)
  }

  function getSortIcon(columnId: string) {
    if (sortBy !== columnId) return <ArrowUpDown className="w-3.5 h-3.5 opacity-40" />
    return sortDesc ? <ArrowDown className="w-3.5 h-3.5" /> : <ArrowUp className="w-3.5 h-3.5" />
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Клиенты</h1>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6 animate-in animate-in-delay-1">
        <StatCard
          label="Всего клиентов"
          value={stats ? formatBalance(stats.total_clients) : '—'}
          icon={Users}
        />
        <StatCard
          label="Общий баланс"
          value={stats ? formatBalance(stats.total_balance) : '—'}
          icon={Wallet}
        />
        <StatCard
          label="Новых за месяц"
          value={stats ? formatBalance(stats.new_this_month) : '—'}
          icon={UserPlus}
        />
        <StatCard
          label="Активных за неделю"
          value={stats ? formatBalance(stats.active_this_week) : '—'}
          icon={Activity}
        />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 mb-4 animate-in animate-in-delay-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-neutral-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPageIndex(0)
            }}
            placeholder="Поиск по имени или телефону..."
            className={cn(
              'w-full pl-9 pr-4 py-2.5 rounded border border-neutral-200 bg-white',
              'text-sm placeholder:text-neutral-400',
              'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              'transition-colors',
            )}
          />
        </div>
        <CustomSelect
          value={String(botFilter ?? '')}
          onChange={(v) => { setBotFilter(v ? Number(v) : undefined); setPageIndex(0) }}
          options={[
            { value: '', label: 'Все боты' },
            ...(bots?.map((bot) => ({ value: String(bot.id), label: bot.name })) ?? []),
          ]}
          placeholder="Все боты"
          width="200px"
          light
        />

        {rfmSegment && (
          <div
            className={cn(
              'flex items-center gap-1.5 px-3 py-2 rounded text-sm font-medium',
              RFM_SEGMENT_COLORS[rfmSegment]?.bg ?? 'bg-neutral-100',
              RFM_SEGMENT_COLORS[rfmSegment]?.text ?? 'text-neutral-700',
            )}
          >
            <span>RFM: {RFM_SEGMENT_LABELS[rfmSegment] ?? rfmSegment}</span>
            <button
              type="button"
              onClick={clearRfmFilter}
              className="ml-1 hover:opacity-70 transition-opacity"
              aria-label="Сбросить фильтр по RFM-сегменту"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        )}
      </div>

      {/* Table */}
      {!data && isLoading ? (
        <div className="animate-in animate-in-delay-3">
          <TableSkeleton />
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить клиентов"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : !data?.items.length ? (
        <EmptyState
          icon={Users}
          title="Клиентов пока нет"
          description="Клиенты появятся после того, как начнут взаимодействовать с вашим Telegram-ботом."
          variant="clients"
        />
      ) : (
        <>
          <div className="bg-white rounded border border-neutral-900 overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="border-b border-neutral-200">
                  {columns.map((col) => (
                    <th
                      key={col.id}
                      className="text-left text-[11px] font-medium text-neutral-400 uppercase tracking-wider px-6 py-4 font-mono"
                    >
                      <button
                        type="button"
                        className={cn(
                          'flex items-center gap-1 hover:text-neutral-900 transition-colors',
                          sortBy === col.id && 'text-neutral-900',
                        )}
                        onClick={() => handleSort(col.id)}
                      >
                        {col.header}
                        {getSortIcon(col.id)}
                      </button>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-200">
                {data.items.map((row) => (
                  <tr
                    key={row.id}
                    className="hover:bg-neutral-50 cursor-pointer transition-colors"
                    onClick={() => navigate(`/dashboard/clients/${row.id}`)}
                  >
                    {columns.map((col) => (
                      <td key={col.id} className="px-6 py-4 text-sm">
                        {col.render(row)}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <Pagination
            page={pageIndex + 1}
            pageCount={pageCount || 1}
            onChange={(p) => setPageIndex(p - 1)}
            total={data.total}
            itemsLabel="клиентов"
          />
        </>
      )}
    </div>
  )
}
