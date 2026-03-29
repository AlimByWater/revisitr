import { useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { Plug, Plus, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useIntegrationsQuery } from '@/features/integrations/queries'
import { CreateIntegrationModal } from '@/components/integrations/CreateIntegrationModal'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'

const TYPE_LABELS: Record<string, string> = {
  iiko: 'iiko',
  rkeeper: 'r-keeper',
  '1c': '1C',
  mock: 'Mock',
}

const STATUS_STYLES: Record<string, { bg: string; text: string; label: string }> = {
  active: { bg: 'bg-green-50', text: 'text-green-700', label: 'Активна' },
  inactive: { bg: 'bg-neutral-100', text: 'text-neutral-500', label: 'Неактивна' },
  error: { bg: 'bg-red-50', text: 'text-red-700', label: 'Ошибка' },
}

function formatDate(dateStr?: string) {
  if (!dateStr) return 'Никогда'
  return new Date(dateStr).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function IntegrationsPage() {
  const navigate = useNavigate()
  const { data: integrations, isLoading, isError, mutate } = useIntegrationsQuery()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          Интеграции
        </h1>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className={cn(
            'flex items-center gap-2 py-2.5 px-4 rounded-lg',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
            'shadow-sm shadow-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          <span className="hidden sm:inline">Добавить</span>
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
          title="Не удалось загрузить интеграции"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : !integrations || integrations.length === 0 ? (
        <EmptyState
          icon={Plug}
          title="Нет интеграций"
          description="Подключите POS-систему (iiko, r-keeper) или создайте Mock-интеграцию для тестирования."
          actionLabel="Добавить интеграцию"
          onAction={() => setShowCreate(true)}
          variant="pos"
        />
      ) : (
        <div className="space-y-3">
          {integrations.map((intg, i) => {
            const status = STATUS_STYLES[intg.status] || STATUS_STYLES.inactive
            return (
              <button
                key={intg.id}
                type="button"
                onClick={() => navigate(`/dashboard/integrations/${intg.id}`)}
                className={cn(
                  'w-full text-left bg-white rounded-2xl shadow-sm border border-surface-border p-6',
                  'hover:border-neutral-300 hover:shadow-md',
                  'transition-all duration-200 group',
                  'animate-in',
                  `animate-in-delay-${Math.min(i + 1, 5)}`,
                )}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="text-base font-semibold text-neutral-900">
                        {TYPE_LABELS[intg.type] || intg.type}
                      </h3>
                      {intg.type === 'mock' && (
                        <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-orange-100 text-orange-700">
                          DEV
                        </span>
                      )}
                    </div>
                    {intg.config.api_url && (
                      <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1 truncate max-w-md">
                        {intg.config.api_url}
                      </p>
                    )}
                    <div className="flex items-center gap-3 mt-1.5 text-xs text-neutral-400">
                      <span className="flex items-center gap-1">
                        <RefreshCw className="w-3 h-3" />
                        Синхр: {formatDate(intg.last_sync_at)}
                      </span>
                    </div>
                  </div>
                  <span
                    className={cn(
                      'text-xs font-medium px-2 py-0.5 rounded-full',
                      status.bg,
                      status.text,
                    )}
                  >
                    {status.label}
                  </span>
                </div>
              </button>
            )
          })}
        </div>
      )}

      {showCreate && <CreateIntegrationModal onClose={() => setShowCreate(false)} />}
    </div>
  )
}
