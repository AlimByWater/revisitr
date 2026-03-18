import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { Store, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { usePOSQuery } from '@/features/pos/queries'
import { CreatePOSModal } from '@/components/pos/CreatePOSModal'

export const Route = createFileRoute('/dashboard/pos/')({
  component: POSListPage,
})

function POSListPage() {
  const navigate = useNavigate()
  const { data: locations, isLoading } = usePOSQuery()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Точки продаж</h1>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className={cn(
            'inline-flex items-center gap-2 px-4 py-2.5 rounded-lg',
            'bg-neutral-900 text-white text-sm font-medium',
            'hover:bg-neutral-800 transition-colors',
          )}
        >
          <Plus className="w-4 h-4" />
          Добавить точку
        </button>
      </div>

      {isLoading && (
        <div className="text-sm text-neutral-500">Загрузка...</div>
      )}

      {!isLoading && (!locations || locations.length === 0) && (
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mx-auto mb-4">
            <Store className="w-8 h-8 text-neutral-400" />
          </div>
          <h2 className="text-lg font-semibold text-neutral-700 mb-2">
            Нет точек продаж
          </h2>
          <p className="text-sm text-neutral-400 max-w-md mx-auto">
            Добавьте первую точку продаж вашего заведения
          </p>
        </div>
      )}

      {locations && locations.length > 0 && (
        <div className="space-y-3">
          {locations.map((loc) => (
            <button
              key={loc.id}
              type="button"
              onClick={() =>
                navigate({
                  to: '/dashboard/pos/$posId',
                  params: { posId: String(loc.id) },
                })
              }
              className="w-full text-left bg-white rounded-2xl shadow-sm border border-surface-border p-6 hover:border-neutral-300 transition-colors"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-base font-semibold text-neutral-900">
                    {loc.name}
                  </h3>
                  {loc.address && (
                    <p className="text-sm text-neutral-500 mt-1">
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
