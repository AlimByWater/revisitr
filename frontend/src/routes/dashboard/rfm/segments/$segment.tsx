import React, { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { ChevronLeft, ChevronUp, ChevronDown, Download, Send, Sprout, TrendingUp, UserCheck, Crown, Gem, AlertTriangle, UserX, Users } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

const SEGMENT_ICONS: Record<string, LucideIcon> = {
  new: Sprout,
  promising: TrendingUp,
  regular: UserCheck,
  vip: Crown,
  rare_valuable: Gem,
  churn_risk: AlertTriangle,
  lost: UserX,
}
import { useRFMSegmentClientsQuery } from '@/features/rfm/queries'
import { RFM_SEGMENT_LABELS, RFM_SEGMENT_COLORS } from '@/features/rfm/types'
import { formatMoney, formatDate, pluralClients, escapeCsvField } from '@/features/rfm/utils'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'

type SortCol = 'r_score' | 'f_score' | 'm_score' | 'last_visit_date' | 'frequency_count' | 'monetary_sum'
type SortOrder = 'asc' | 'desc'

export default function SegmentDetailPage() {
  const navigate = useNavigate()
  const { segment } = useParams<{ segment: string }>()

  const [page, setPage] = useState(1)
  const [sort, setSort] = useState<SortCol>('monetary_sum')
  const [order, setOrder] = useState<SortOrder>('desc')
  const perPage = 20

  const { data, isLoading, isError, mutate } = useRFMSegmentClientsQuery(segment, {
    page,
    per_page: perPage,
    sort,
    order,
  })

  const segmentLabel = RFM_SEGMENT_LABELS[segment || ''] || segment
  const segmentColors = RFM_SEGMENT_COLORS[segment || '']
  const totalPages = data ? Math.ceil(data.total / perPage) : 0

  function handleSort(col: SortCol) {
    if (sort === col) {
      setOrder(order === 'desc' ? 'asc' : 'desc')
    } else {
      setSort(col)
      setOrder('desc')
    }
    setPage(1)
  }

  function handleExportCSV() {
    if (!data?.clients.length) return

    const headers = ['Имя', 'Фамилия', 'Телефон', 'R', 'F', 'M', 'Посл. визит', 'Выручка', 'Визиты']
    const rows = data.clients.map((c) => [
      escapeCsvField(c.first_name),
      escapeCsvField(c.last_name),
      escapeCsvField(c.phone),
      c.r_score ?? '',
      c.f_score ?? '',
      c.m_score ?? '',
      c.last_visit_date ? new Date(c.last_visit_date).toLocaleDateString('ru-RU') : '',
      c.monetary_sum ?? '',
      c.total_visits_lifetime,
    ])

    const csv = [headers, ...rows].map((r) => r.join(',')).join('\n')
    const blob = new Blob(['\ufeff' + csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `rfm-${segment}-page${page}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  function SortIcon({ col }: { col: SortCol }) {
    if (sort !== col) return null
    return order === 'asc'
      ? <ChevronUp className="w-3 h-3 inline ml-0.5" />
      : <ChevronDown className="w-3 h-3 inline ml-0.5" />
  }

  return (
    <div>
      {/* Back + header */}
      <div className="mb-6 animate-in">
        <button
          type="button"
          onClick={() => navigate('/dashboard/rfm')}
          className="flex items-center gap-1 text-sm text-neutral-400 hover:text-neutral-600 transition-colors mb-3"
        >
          <ChevronLeft className="w-4 h-4" />
          RFM-сегментация
        </button>

        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            {(() => {
              const Icon = SEGMENT_ICONS[segment || '']
              return Icon ? <Icon className="w-6 h-6 shrink-0" style={{ color: segmentColors?.color }} /> : null
            })()}
            <div>
              <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
                {segmentLabel}
              </h1>
              {data && (
                <p className="text-sm text-neutral-500">
                  {data.total} {pluralClients(data.total)}
                </p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleExportCSV}
              disabled={!data?.clients.length}
              className={cn(
                'flex items-center gap-2 py-2 px-4 rounded',
                'border border-neutral-900 text-sm font-medium text-neutral-600',
                'hover:bg-neutral-50 hover:border-neutral-300',
                'transition-all duration-150',
                'disabled:opacity-40 disabled:cursor-not-allowed',
              )}
            >
              <Download className="w-4 h-4" />
              <span className="hidden sm:inline">Экспорт CSV</span>
            </button>
            <button
              type="button"
              onClick={() => navigate(`/dashboard/campaigns/create?segment=${segment}`)}
              className={cn(
                'flex items-center gap-2 py-2 px-4 rounded',
                'bg-accent text-white text-sm font-medium',
                'hover:bg-accent-hover active:bg-accent/80',
                'transition-all duration-150',
                'shadow-none',
              )}
            >
              <Send className="w-4 h-4" />
              <span className="hidden sm:inline">Запустить сценарий</span>
            </button>
          </div>
        </div>
      </div>

      {/* Table */}
      {isLoading ? (
        <TableSkeleton rows={10} />
      ) : isError ? (
        <ErrorState title="Не удалось загрузить клиентов" onRetry={() => mutate()} />
      ) : !data?.clients.length ? (
        <div className="flex flex-col items-center justify-center py-24 text-center animate-in">
          <div className="w-16 h-16 rounded bg-neutral-100 flex items-center justify-center mb-4">
            <Users className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5 tracking-tight">Нет клиентов</h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">В этом сегменте пока нет клиентов</p>
        </div>
      ) : (
        <div className="bg-white rounded border border-neutral-900 overflow-hidden animate-in">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr>
                  {[
                    { key: null, label: 'Имя', align: 'left' as const },
                    { key: 'r_score' as SortCol, label: 'R' },
                    { key: 'f_score' as SortCol, label: 'F' },
                    { key: 'm_score' as SortCol, label: 'M' },
                    { key: 'last_visit_date' as SortCol, label: 'Посл. визит' },
                    { key: 'monetary_sum' as SortCol, label: 'Выручка' },
                    { key: 'frequency_count' as SortCol, label: 'Визиты' },
                  ].map((col) => (
                    <th
                      key={col.label}
                      className={cn(
                        'px-4 py-3 text-xs font-medium text-neutral-400 uppercase tracking-wider whitespace-nowrap',
                        col.align === 'left' ? 'text-left' : 'text-right',
                        col.key && 'cursor-pointer hover:text-neutral-600 select-none',
                      )}
                      onClick={col.key ? () => handleSort(col.key!) : undefined}
                    >
                      {col.label}
                      {col.key && <SortIcon col={col.key} />}
                    </th>
                  ))}
                </tr>
                <tr><td colSpan={7} className="p-0"><div className="mx-4 border-t border-neutral-200" /></td></tr>
              </thead>
              <tbody>
                {data.clients.map((client, i) => (
                  <React.Fragment key={client.id}>
                  <tr
                    className={cn(
                      'hover:bg-neutral-50 transition-colors cursor-pointer',
                      'animate-in',
                      `animate-in-delay-${Math.min(i + 1, 5)}`,
                    )}
                    onClick={() => window.open(`/revisitr/dashboard/clients/${client.id}`, '_blank')}
                  >
                    <td className="px-4 py-3 font-medium text-neutral-900 whitespace-nowrap">
                      {client.first_name} {client.last_name?.charAt(0)}.
                    </td>
                    <td className="px-4 py-3 text-right">
                      <ScoreBadge value={client.r_score} />
                    </td>
                    <td className="px-4 py-3 text-right">
                      <ScoreBadge value={client.f_score} />
                    </td>
                    <td className="px-4 py-3 text-right">
                      <ScoreBadge value={client.m_score} />
                    </td>
                    <td className="px-4 py-3 text-right text-neutral-500 font-mono tabular-nums whitespace-nowrap">
                      {formatDate(client.last_visit_date)}
                    </td>
                    <td className="px-4 py-3 text-right text-neutral-700 font-mono tabular-nums whitespace-nowrap">
                      {formatMoney(client.monetary_sum)}
                    </td>
                    <td className="px-4 py-3 text-right text-neutral-700 font-mono tabular-nums">
                      {client.total_visits_lifetime}
                    </td>
                  </tr>
                  {i < data.clients.length - 1 && (
                    <tr><td colSpan={7} className="p-0"><div className="mx-4 border-t border-neutral-200" /></td></tr>
                  )}
                </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <>
            <div className="mx-4 border-t border-neutral-200" />
            <div className="flex items-center justify-between px-6 py-3">
              <span className="text-xs text-neutral-400">
                Стр. {page} из {totalPages}
              </span>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1.5 rounded text-sm border border-neutral-900 disabled:opacity-40 hover:bg-neutral-50 transition-colors"
                >
                  ◄
                </button>
                {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                  let p: number
                  if (totalPages <= 5) {
                    p = i + 1
                  } else if (page <= 3) {
                    p = i + 1
                  } else if (page >= totalPages - 2) {
                    p = totalPages - 4 + i
                  } else {
                    p = page - 2 + i
                  }
                  return (
                    <button
                      key={p}
                      type="button"
                      onClick={() => setPage(p)}
                      className={cn(
                        'w-8 h-8 rounded text-sm font-medium transition-colors',
                        page === p
                          ? 'bg-neutral-900 text-white'
                          : 'text-neutral-600 hover:bg-neutral-100',
                      )}
                    >
                      {p}
                    </button>
                  )
                })}
                <button
                  type="button"
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page >= totalPages}
                  className="px-3 py-1.5 rounded text-sm border border-neutral-900 disabled:opacity-40 hover:bg-neutral-50 transition-colors"
                >
                  ►
                </button>
              </div>
            </div>
            </>
          )}
        </div>
      )}
    </div>
  )
}

function ScoreBadge({ value }: { value: number | null }) {
  if (value == null) return <span className="text-neutral-300">—</span>

  const colors: Record<number, string> = {
    5: 'bg-green-100 text-green-700',
    4: 'bg-blue-100 text-blue-700',
    3: 'bg-yellow-100 text-yellow-700',
    2: 'bg-orange-100 text-orange-700',
    1: 'bg-red-100 text-red-700',
  }

  return (
    <span
      className={cn(
        'inline-flex items-center justify-center w-7 h-6 rounded text-xs font-bold tabular-nums',
        colors[value] || 'bg-neutral-100 text-neutral-600',
      )}
    >
      {value}
    </span>
  )
}

