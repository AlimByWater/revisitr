import { useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { Store, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { usePOSQuery } from '@/features/pos/queries'
import { CreatePOSModal } from '@/components/pos/CreatePOSModal'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'

export default function POSListPage() {
  const navigate = useNavigate()
  const { data: locations, isLoading, isError, mutate } = usePOSQuery()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div>
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Точки продаж</h1>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className={cn(
            'flex items-center gap-2 py-2.5 px-4 rounded',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
            '',
          )}
        >
          <Plus className="w-4 h-4" />
          <span className="hidden sm:inline">Добавить точку</span>
          <span className="sm:hidden">Добавить</span>
        </button>
      </div>

      {isLoading ? (
        <div className="space-y-3">
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
        <div className="space-y-3">
          {locations.map((loc, i) => (
            <button
              key={loc.id}
              type="button"
              onClick={() =>
                navigate(`/dashboard/pos/${loc.id}`)
              }
              className={cn(
                'w-full text-left bg-white rounded border border-neutral-900 p-6',
                'hover:border-neutral-300 hover:shadow-md',
                'transition-all duration-200 group',
                'animate-in',
                `animate-in-delay-${Math.min(i + 1, 5)}`,
              )}
            >
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-base font-semibold text-neutral-900">
                    {loc.name}
                  </h3>
                  {loc.address && (
                    <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1">
                      {loc.address}
                    </p>
                  )}
                  {loc.phone && (
                    <p className="text-sm text-neutral-400 mt-0.5">
                      {loc.phone}
                    </p>
                  )}
                </div>
                <span
                  className={cn(
                    'text-xs font-medium px-2 py-0.5 rounded-full',
                    loc.is_active
                      ? 'bg-green-50 text-green-700'
                      : 'bg-neutral-100 text-neutral-500',
                  )}
                >
                  {loc.is_active ? 'Активна' : 'Неактивна'}
                </span>
              </div>
            </button>
          ))}
        </div>
      )}

      {showCreate && (
        <CreatePOSModal onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}

