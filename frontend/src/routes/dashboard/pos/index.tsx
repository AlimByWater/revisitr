import { Link } from 'react-router-dom'
import { useState } from 'react'
import { Store, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { usePOSQuery } from '@/features/pos/queries'
import { CreatePOSModal } from '@/components/pos/CreatePOSModal'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { Button } from '@/components/common/Button'

function formatScheduleSummary(schedule?: Record<string, { open?: string; close?: string; closed?: boolean }>): string {
  if (!schedule) return ''
  const workdays = ['mon', 'tue', 'wed', 'thu', 'fri']
  const weekend = ['sat', 'sun']
  const allWorkOpen = workdays.every((d) => schedule[d] && !schedule[d].closed)
  const noWeekend = weekend.every((d) => !schedule[d] || schedule[d].closed)
  if (allWorkOpen && noWeekend) {
    const open = schedule['mon']?.open ?? '09:00'
    const close = schedule['mon']?.close ?? '22:00'
    return `Пн–Пт ${open.slice(0, 5)}–${close.slice(0, 5)}`
  }
  const openDays = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'].filter((d) => schedule[d] && !schedule[d].closed)
  if (openDays.length === 0) return 'Выходной'
  if (openDays.length === 7) {
    const open = schedule['mon']?.open ?? '09:00'
    const close = schedule['mon']?.close ?? '22:00'
    return `Ежедневно ${open.slice(0, 5)}–${close.slice(0, 5)}`
  }
  return `${openDays.length} дн/нед`
}

export default function POSListPage() {
  const { data: locations, isLoading, isError, mutate } = usePOSQuery()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div>
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Точки продаж</h1>
        <Button
          variant="primary"
          leftIcon={<Plus className="w-4 h-4" />}
          onClick={() => setShowCreate(true)}
        >
          <span className="hidden sm:inline">Добавить точку</span>
          <span className="sm:hidden">Добавить</span>
        </Button>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[0, 1, 2].map((i) => (
            <div key={i} className={cn('animate-in', `animate-in-delay-${i + 1}`)}>
              <CardSkeleton />
            </div>
          ))}
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить точки продаж"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : !locations || locations.length === 0 ? (
        <EmptyState
          icon={Store}
          title="Нет точек продаж"
          description="Добавьте первую точку продаж вашего заведения — адрес, телефон и график работы."
          actionLabel="Добавить точку"
          onAction={() => setShowCreate(true)}
          variant="pos"
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {locations.map((loc, i) => (
            <Link
              key={loc.id}
              to={`/dashboard/pos/${loc.id}`}
              className={cn(
                'bg-white rounded border border-neutral-900 p-6',
                'cursor-pointer hover:scale-[1.02] transition-transform duration-150',
                'group animate-in',
                `animate-in-delay-${Math.min(i + 1, 5)}`,
              )}
            >
              <div className="flex items-start justify-between mb-3">
                <h3 className="text-base font-semibold text-neutral-900">
                  {loc.name}
                </h3>
                <span
                  className={cn(
                    'font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border shrink-0',
                    loc.is_active
                      ? 'bg-emerald-500/10 text-emerald-700 border-emerald-500/30'
                      : 'bg-neutral-100 text-neutral-600 border-neutral-300',
                  )}
                >
                  {loc.is_active ? 'Активна' : 'Неактивна'}
                </span>
              </div>
              {loc.address && (
                <p className="text-sm text-neutral-500 mb-1">{loc.address}</p>
              )}
              {loc.phone && (
                <p className="text-sm text-neutral-400">{loc.phone}</p>
              )}
              <div className="mt-4 pt-4 border-t border-neutral-200 flex items-center justify-between text-neutral-400 font-mono text-[11px] uppercase tracking-wider tabular-nums">
                <span className="flex items-center gap-1.5">
                  <Store className="w-3.5 h-3.5" />
                  {formatScheduleSummary(loc.schedule as any) || 'График не задан'}
                </span>
              </div>
            </Link>
          ))}

        </div>
      )}

      {showCreate && (
        <CreatePOSModal onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}

