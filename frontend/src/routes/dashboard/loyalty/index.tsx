import { useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { Heart, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useProgramsQuery, useUpdateProgramMutation } from '@/features/loyalty/queries'
import { CreateProgramModal } from '@/components/loyalty/CreateProgramModal'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'

export default function LoyaltyProgramsPage() {
  const navigate = useNavigate()
  const { data: programs, isLoading, isError, mutate } = useProgramsQuery()
  const updateMutation = useUpdateProgramMutation()
  const [showCreate, setShowCreate] = useState(false)

  function handleToggleActive(id: number, currentActive: boolean) {
    updateMutation.mutate({ id, data: { is_active: !currentActive } })
  }

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Программы лояльности</h1>
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
          <span className="hidden sm:inline">Создать программу</span>
          <span className="sm:hidden">Создать</span>
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
          title="Не удалось загрузить программы"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : !programs || programs.length === 0 ? (
        <EmptyState
          icon={Heart}
          title="Нет программ лояльности"
          description="Создайте первую программу лояльности для ваших клиентов — бонусную или скидочную."
          actionLabel="Создать программу"
          onAction={() => setShowCreate(true)}
          variant="loyalty"
        />
      ) : (
        <div className="space-y-3">
          {programs.map((program, i) => (
            <button
              key={program.id}
              type="button"
              onClick={() =>
                navigate(`/dashboard/loyalty/${program.id}`)
              }
              className={cn(
                'w-full text-left bg-white rounded-2xl shadow-sm border border-surface-border p-6',
                'hover:border-neutral-300 hover:shadow-md',
                'transition-all duration-200 group',
                'animate-in',
                `animate-in-delay-${Math.min(i + 1, 5)}`,
              )}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <h3 className="text-base font-semibold text-neutral-900">
                    {program.name}
                  </h3>
                  <span
                    className={cn(
                      'text-xs font-medium px-2 py-0.5 rounded-full',
                      program.type === 'bonus'
                        ? 'bg-blue-50 text-blue-700'
                        : 'bg-purple-50 text-purple-700',
                    )}
                  >
                    {program.type === 'bonus' ? 'Бонусная' : 'Скидочная'}
                  </span>
                </div>

                <div className="flex items-center gap-4">
                  {program.levels && (
                    <span className="text-xs text-neutral-500">
                      <span className="font-mono tabular-nums">{program.levels.length}</span>{' '}
                      {pluralLevels(program.levels.length)}
                    </span>
                  )}
                  <label
                    className="relative inline-flex items-center cursor-pointer"
                    onClick={(e) => e.stopPropagation()}
                    aria-label={
                      program.is_active
                        ? 'Деактивировать программу'
                        : 'Активировать программу'
                    }
                  >
                    <input
                      type="checkbox"
                      checked={program.is_active}
                      onChange={() =>
                        handleToggleActive(program.id, program.is_active)
                      }
                      className="sr-only peer"
                    />
                    <div
                      className={cn(
                        'w-9 h-5 rounded-full transition-colors',
                        'peer-checked:bg-accent bg-neutral-300',
                        'after:content-[""] after:absolute after:top-0.5 after:start-[2px]',
                        'after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all',
                        'peer-checked:after:translate-x-full',
                      )}
                    />
                  </label>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}

      {showCreate && (
        <CreateProgramModal onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}

function pluralLevels(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return 'уровень'
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return 'уровня'
  return 'уровней'
}
