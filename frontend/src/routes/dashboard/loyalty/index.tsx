import { Link, useSearchParams } from 'react-router-dom'
import { useState } from 'react'
import { ArrowLeft, Heart, Plus, Users } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useProgramsQuery, useUpdateProgramMutation } from '@/features/loyalty/queries'
import { CreateProgramModal } from '@/components/loyalty/CreateProgramModal'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { Button } from '@/components/common/Button'

export default function LoyaltyProgramsPage() {
  const [searchParams] = useSearchParams()
  const botId = searchParams.get('botId')
  const { data: programs, isLoading, isError, mutate } = useProgramsQuery()
  const updateMutation = useUpdateProgramMutation()
  const [showCreate, setShowCreate] = useState(false)

  function handleToggleActive(id: number, currentActive: boolean) {
    updateMutation.mutate({ id, data: { is_active: !currentActive } })
  }

  return (
    <div>
      {botId && (
        <Link
          to={`/dashboard/bots/${botId}?tab=modules`}
          className="mb-4 inline-flex min-h-11 items-center gap-1.5 rounded text-sm text-neutral-500 transition-colors hover:text-neutral-700"
        >
          <ArrowLeft className="h-4 w-4" />
          Назад к модулям бота
        </Link>
      )}

      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Программы лояльности</h1>
        <Button
          variant="primary"
          leftIcon={<Plus className="w-4 h-4" />}
          onClick={() => setShowCreate(true)}
        >
          <span className="hidden sm:inline">Создать программу</span>
          <span className="sm:hidden">Создать</span>
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
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {programs.map((program, i) => (
            <Link
              key={program.id}
              to={`/dashboard/loyalty/${program.id}${botId ? `?botId=${botId}` : ''}`}
              className={cn(
                'bg-white rounded border border-neutral-900 p-6',
                'cursor-pointer hover:scale-[1.02] transition-transform duration-150',
                'group animate-in',
                `animate-in-delay-${Math.min(i + 1, 5)}`,
              )}
            >
              <div className="flex items-start justify-between mb-3">
                <h3 className="text-base font-semibold text-neutral-900">
                  {program.name}
                </h3>
                <span
                  className={cn(
                    'font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border shrink-0',
                    program.is_active
                      ? 'bg-emerald-500/10 text-emerald-700 border-emerald-500/30'
                      : 'bg-neutral-100 text-neutral-600 border-neutral-300',
                  )}
                >
                  {program.is_active ? 'Активна' : 'Неактивна'}
                </span>
              </div>

              <div className="flex items-center gap-2 mb-3">
                <span
                  className={cn(
                    'font-mono text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded',
                    program.type === 'bonus'
                      ? 'bg-accent text-white'
                      : 'bg-neutral-900 text-white',
                  )}
                >
                  {program.type === 'bonus' ? 'Бонусная' : 'Скидочная'}
                </span>
                {program.levels && program.levels.length > 0 && (
                  <span className="text-xs text-neutral-400">
                    <span className="font-mono tabular-nums">{program.levels.length}</span>{' '}
                    {pluralLevels(program.levels.length)}
                  </span>
                )}
              </div>

              <div className="flex items-center justify-between mt-4 pt-4 border-t border-neutral-200">
                <div className="flex items-center gap-1.5 text-neutral-400">
                  <Users className="w-3.5 h-3.5" />
                  <span className="font-mono text-[11px] uppercase tracking-wider tabular-nums">
                    {(program.client_count ?? 0).toLocaleString('ru-RU')} клиентов
                  </span>
                </div>
                <label
                  className="relative inline-flex items-center cursor-pointer"
                  onClick={(e) => e.preventDefault()}
                  aria-label={program.is_active ? 'Деактивировать программу' : 'Активировать программу'}
                >
                  <input
                    type="checkbox"
                    checked={program.is_active}
                    onChange={() => handleToggleActive(program.id, program.is_active)}
                    onClick={(e) => e.stopPropagation()}
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
            </Link>
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
