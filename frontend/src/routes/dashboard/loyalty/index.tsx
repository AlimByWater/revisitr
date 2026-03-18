import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { Heart, Plus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useProgramsQuery, useUpdateProgramMutation } from '@/features/loyalty/queries'
import { CreateProgramModal } from '@/components/loyalty/CreateProgramModal'

export const Route = createFileRoute('/dashboard/loyalty/')({
  component: LoyaltyProgramsPage,
})

function LoyaltyProgramsPage() {
  const navigate = useNavigate()
  const { data: programs, isLoading } = useProgramsQuery()
  const updateMutation = useUpdateProgramMutation()
  const [showCreate, setShowCreate] = useState(false)

  function handleToggleActive(id: number, currentActive: boolean) {
    updateMutation.mutate({ id, data: { is_active: !currentActive } })
  }

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Программы лояльности</h1>
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
          Создать программу
        </button>
      </div>

      {isLoading && (
        <div className="text-sm text-neutral-500">Загрузка...</div>
      )}

      {!isLoading && (!programs || programs.length === 0) && (
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mx-auto mb-4">
            <Heart className="w-8 h-8 text-neutral-400" />
          </div>
          <h2 className="text-lg font-semibold text-neutral-700 mb-2">
            Нет программ лояльности
          </h2>
          <p className="text-sm text-neutral-400 max-w-md mx-auto">
            Создайте первую программу лояльности для ваших клиентов
          </p>
        </div>
      )}

      {programs && programs.length > 0 && (
        <div className="space-y-3">
          {programs.map((program) => (
            <button
              key={program.id}
              type="button"
              onClick={() =>
                navigate({
                  to: '/dashboard/loyalty/$programId',
                  params: { programId: String(program.id) },
                })
              }
              className="w-full text-left bg-white rounded-2xl shadow-sm border border-surface-border p-6 hover:border-neutral-300 transition-colors"
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
                      {program.levels.length}{' '}
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
