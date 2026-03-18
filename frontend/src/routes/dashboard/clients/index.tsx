import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useState, useMemo } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
  type SortingState,
  type PaginationState,
} from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { useClientsQuery, useClientStatsQuery } from '@/features/clients/queries'
import { useBotsQuery } from '@/features/bots/queries'
import type { ClientProfile, ClientFilter } from '@/features/clients/types'
import {
  Users,
  Search,
  Wallet,
  UserPlus,
  Activity,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
} from 'lucide-react'

export const Route = createFileRoute('/dashboard/clients/')({
  component: ClientsPage,
})

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

const columnHelper = createColumnHelper<ClientProfile>()

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
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
      <div className="flex items-center gap-3 mb-2">
        <div className="w-10 h-10 rounded-xl bg-neutral-100 flex items-center justify-center">
          <Icon className="w-5 h-5 text-neutral-500" />
        </div>
        <span className="text-sm text-neutral-500">{label}</span>
      </div>
      <p className="text-2xl font-bold text-neutral-900">{value}</p>
    </div>
  )
}

function ClientsPage() {
  const navigate = useNavigate()
  const { data: bots } = useBotsQuery()
  const { data: stats } = useClientStatsQuery()

  const [search, setSearch] = useState('')
  const [botFilter, setBotFilter] = useState<number | undefined>(undefined)
  const [sorting, setSorting] = useState<SortingState>([])
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20,
  })

  const filter: ClientFilter = useMemo(
    () => ({
      search: search || undefined,
      bot_id: botFilter,
      sort_by: sorting[0]?.id,
      sort_order: sorting[0] ? (sorting[0].desc ? 'desc' : 'asc') : undefined,
      limit: pagination.pageSize,
      offset: pagination.pageIndex * pagination.pageSize,
    }),
    [search, botFilter, sorting, pagination],
  )

  const { data, isLoading, isError } = useClientsQuery(filter)

  const columns = useMemo(
    () => [
      columnHelper.accessor(
        (row) =>
          [row.first_name, row.last_name].filter(Boolean).join(' '),
        {
          id: 'name',
          header: 'Имя',
          cell: (info) => (
            <span className="font-medium text-neutral-900">
              {info.getValue()}
            </span>
          ),
        },
      ),
      columnHelper.accessor('bot_name', {
        header: 'Бот',
        cell: (info) => (
          <span className="text-neutral-600">{info.getValue()}</span>
        ),
      }),
      columnHelper.accessor('loyalty_level', {
        header: 'Уровень',
        cell: (info) => {
          const level = info.getValue()
          if (!level) return <span className="text-neutral-400">—</span>
          return (
            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-neutral-100 text-neutral-700">
              {level}
            </span>
          )
        },
      }),
      columnHelper.accessor('loyalty_balance', {
        header: 'Баланс',
        cell: (info) => (
          <span className="font-medium text-neutral-900 tabular-nums">
            {formatBalance(info.getValue())}
          </span>
        ),
      }),
      columnHelper.accessor('purchase_count', {
        header: 'Покупки',
        cell: (info) => (
          <span className="text-neutral-600 tabular-nums">
            {info.getValue()}
          </span>
        ),
      }),
      columnHelper.accessor('registered_at', {
        header: 'Зарегистрирован',
        cell: (info) => (
          <span className="text-neutral-500">{formatDate(info.getValue())}</span>
        ),
      }),
    ],
    [],
  )

  const pageCount = data ? Math.ceil(data.total / pagination.pageSize) : 0

  const table = useReactTable({
    data: data?.items ?? [],
    columns,
    pageCount,
    state: { sorting, pagination },
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
  })

  return (
    <div className="max-w-6xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-neutral-900">Клиенты</h1>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
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
      <div className="flex items-center gap-3 mb-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-neutral-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPagination((prev) => ({ ...prev, pageIndex: 0 }))
            }}
            placeholder="Поиск по имени или телефону..."
            className={cn(
              'w-full pl-9 pr-4 py-2.5 rounded-lg border border-surface-border',
              'text-sm placeholder:text-neutral-400',
              'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              'transition-colors',
            )}
          />
        </div>
        <select
          value={botFilter ?? ''}
          onChange={(e) => {
            setBotFilter(e.target.value ? Number(e.target.value) : undefined)
            setPagination((prev) => ({ ...prev, pageIndex: 0 }))
          }}
          aria-label="Фильтр по боту"
          className={cn(
            'px-4 py-2.5 rounded-lg border border-surface-border',
            'text-sm bg-white',
            'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
            'transition-colors',
          )}
        >
          <option value="">Все боты</option>
          {bots?.map((bot) => (
            <option key={bot.id} value={bot.id}>
              {bot.name}
            </option>
          ))}
        </select>
      </div>

      {/* Table */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-neutral-300 border-t-neutral-900 rounded-full animate-spin" />
        </div>
      ) : isError ? (
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-red-600">
            Не удалось загрузить список клиентов. Попробуйте обновить страницу.
          </p>
        </div>
      ) : !data?.items.length ? (
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mx-auto mb-4">
            <Users className="w-8 h-8 text-neutral-400" />
          </div>
          <h2 className="text-lg font-semibold text-neutral-700 mb-2">
            Клиентов пока нет
          </h2>
          <p className="text-sm text-neutral-400 max-w-md mx-auto">
            Клиенты появятся после того, как начнут взаимодействовать с вашим Telegram-ботом
          </p>
        </div>
      ) : (
        <>
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden">
            <table className="w-full">
              <thead>
                {table.getHeaderGroups().map((headerGroup) => (
                  <tr
                    key={headerGroup.id}
                    className="border-b border-surface-border"
                  >
                    {headerGroup.headers.map((header) => {
                      const sorted = header.column.getIsSorted()
                      return (
                        <th
                          key={header.id}
                          className="text-left text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3"
                        >
                          {header.isPlaceholder ? null : (
                            <button
                              type="button"
                              className={cn(
                                'flex items-center gap-1 hover:text-neutral-900 transition-colors',
                                sorted && 'text-neutral-900',
                              )}
                              onClick={header.column.getToggleSortingHandler()}
                            >
                              {flexRender(
                                header.column.columnDef.header,
                                header.getContext(),
                              )}
                              {sorted === 'asc' ? (
                                <ArrowUp className="w-3.5 h-3.5" />
                              ) : sorted === 'desc' ? (
                                <ArrowDown className="w-3.5 h-3.5" />
                              ) : (
                                <ArrowUpDown className="w-3.5 h-3.5 opacity-40" />
                              )}
                            </button>
                          )}
                        </th>
                      )
                    })}
                  </tr>
                ))}
              </thead>
              <tbody className="divide-y divide-surface-border">
                {table.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    className="hover:bg-neutral-50 cursor-pointer transition-colors"
                    onClick={() =>
                      navigate({
                        to: '/dashboard/clients/$clientId',
                        params: { clientId: String(row.original.id) },
                      })
                    }
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-4 py-3 text-sm">
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext(),
                        )}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="flex items-center justify-between mt-4">
            <p className="text-sm text-neutral-500">
              Всего {formatBalance(data.total)} клиентов
            </p>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={() => table.setPageIndex(0)}
                disabled={!table.getCanPreviousPage()}
                aria-label="Первая страница"
                className={cn(
                  'p-2 rounded-lg text-neutral-500 hover:bg-neutral-100 transition-colors',
                  'disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent',
                )}
              >
                <ChevronsLeft className="w-4 h-4" />
              </button>
              <button
                type="button"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
                aria-label="Предыдущая страница"
                className={cn(
                  'p-2 rounded-lg text-neutral-500 hover:bg-neutral-100 transition-colors',
                  'disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent',
                )}
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <span className="px-3 py-2 text-sm text-neutral-700">
                {pagination.pageIndex + 1} / {pageCount || 1}
              </span>
              <button
                type="button"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
                aria-label="Следующая страница"
                className={cn(
                  'p-2 rounded-lg text-neutral-500 hover:bg-neutral-100 transition-colors',
                  'disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent',
                )}
              >
                <ChevronRight className="w-4 h-4" />
              </button>
              <button
                type="button"
                onClick={() => table.setPageIndex(pageCount - 1)}
                disabled={!table.getCanNextPage()}
                aria-label="Последняя страница"
                className={cn(
                  'p-2 rounded-lg text-neutral-500 hover:bg-neutral-100 transition-colors',
                  'disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent',
                )}
              >
                <ChevronsRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
