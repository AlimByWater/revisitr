import { Link, useParams } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { ArrowLeft, Plus, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useProgramQuery,
  useUpdateProgramMutation,
  useUpdateLevelsMutation,
} from '@/features/loyalty/queries'
import { deleteLevel } from '@/features/loyalty/api'
import type { LoyaltyLevel, ProgramConfig } from '@/features/loyalty/types'

export default function ProgramDetailPage() {
  const { programId } = useParams<{ programId: string }>()
  const id = Number(programId)

  const { data: program, isLoading, mutate: refetch } = useProgramQuery(id)
  const updateProgram = useUpdateProgramMutation()
  const updateLevels = useUpdateLevelsMutation(id)

  const [config, setConfig] = useState<ProgramConfig>({
    welcome_bonus: 0,
    currency_name: '',
  })
  const [levels, setLevels] = useState<Omit<LoyaltyLevel, 'program_id'>[]>([])
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (program) {
      setConfig(program.config)
      setLevels(
        (program.levels ?? []).map(({ program_id: _, ...rest }) => rest),
      )
    }
  }, [program])

  async function handleSave() {
    setSaving(true)
    try {
      await updateProgram.mutateAsync({ id, data: { config } })
      await updateLevels.mutateAsync(levels)
    } finally {
      setSaving(false)
    }
  }

  function addLevel() {
    const nextOrder = levels.length > 0
      ? Math.max(...levels.map((l) => l.sort_order)) + 1
      : 1
    setLevels([
      ...levels,
      {
        id: 0,
        name: '',
        threshold: 0,
        reward_percent: 0,
        sort_order: nextOrder,
      },
    ])
  }

  function updateLevel(
    index: number,
    field: keyof Omit<LoyaltyLevel, 'id' | 'program_id'>,
    value: string | number,
  ) {
    setLevels((prev) =>
      prev.map((l, i) => (i === index ? { ...l, [field]: value } : l)),
    )
  }

  async function removeLevel(index: number) {
    const level = levels[index]
    if (level.id > 0) {
      await deleteLevel(id, level.id)
      refetch()
    }
    setLevels((prev) => prev.filter((_, i) => i !== index))
  }

  if (isLoading) {
    return <div className="text-sm text-neutral-500">Загрузка...</div>
  }

  if (!program) {
    return <div className="text-sm text-neutral-500">Программа не найдена</div>
  }

  return (
    <div className="max-w-4xl">
      <Link
        to="/dashboard/loyalty"
        className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-900 transition-colors mb-4"
      >
        <ArrowLeft className="w-4 h-4" />
        Назад к программам
      </Link>

      <div className="flex items-center gap-3 mb-8">
        <h1 className="text-2xl font-bold">{program.name}</h1>
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

      {/* Config section */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-6">
        <h2 className="text-base font-semibold mb-4">Настройки</h2>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label
              htmlFor="welcome_bonus"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Приветственный бонус
            </label>
            <input
              id="welcome_bonus"
              type="number"
              min={0}
              value={config.welcome_bonus}
              onChange={(e) =>
                setConfig((c) => ({
                  ...c,
                  welcome_bonus: Number(e.target.value),
                }))
              }
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>
          <div>
            <label
              htmlFor="currency_name"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Название валюты
            </label>
            <input
              id="currency_name"
              type="text"
              value={config.currency_name}
              onChange={(e) =>
                setConfig((c) => ({ ...c, currency_name: e.target.value }))
              }
              placeholder="баллы"
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>
        </div>
      </div>

      {/* Levels section */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-base font-semibold">Уровни</h2>
          <button
            type="button"
            onClick={addLevel}
            className={cn(
              'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg',
              'text-sm font-medium text-neutral-700',
              'border border-surface-border hover:bg-neutral-50 transition-colors',
            )}
          >
            <Plus className="w-4 h-4" />
            Добавить уровень
          </button>
        </div>

        {levels.length === 0 ? (
          <p className="text-sm text-neutral-400 text-center py-6">
            Нет уровней. Добавьте первый уровень программы.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-surface-border text-left">
                  <th className="pb-2 font-medium text-neutral-500">
                    Название
                  </th>
                  <th className="pb-2 font-medium text-neutral-500">Порог</th>
                  <th className="pb-2 font-medium text-neutral-500">
                    Вознаграждение %
                  </th>
                  <th className="pb-2 font-medium text-neutral-500">
                    Порядок
                  </th>
                  <th className="pb-2 w-10" />
                </tr>
              </thead>
              <tbody>
                {levels.map((level, index) => (
                  <tr
                    key={level.id || `new-${index}`}
                    className="border-b border-surface-border last:border-0"
                  >
                    <td className="py-2 pr-2">
                      <input
                        type="text"
                        value={level.name}
                        onChange={(e) =>
                          updateLevel(index, 'name', e.target.value)
                        }
                        placeholder="Название уровня"
                        aria-label="Название уровня"
                        className={cn(
                          'w-full px-3 py-1.5 rounded-lg border border-surface-border',
                          'text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                        )}
                      />
                    </td>
                    <td className="py-2 pr-2">
                      <input
                        type="number"
                        min={0}
                        value={level.threshold}
                        onChange={(e) =>
                          updateLevel(
                            index,
                            'threshold',
                            Number(e.target.value),
                          )
                        }
                        aria-label="Порог"
                        className={cn(
                          'w-24 px-3 py-1.5 rounded-lg border border-surface-border',
                          'text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                        )}
                      />
                    </td>
                    <td className="py-2 pr-2">
                      <input
                        type="number"
                        min={0}
                        max={100}
                        value={level.reward_percent}
                        onChange={(e) =>
                          updateLevel(
                            index,
                            'reward_percent',
                            Number(e.target.value),
                          )
                        }
                        aria-label="Вознаграждение в процентах"
                        className={cn(
                          'w-24 px-3 py-1.5 rounded-lg border border-surface-border',
                          'text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                        )}
                      />
                    </td>
                    <td className="py-2 pr-2">
                      <input
                        type="number"
                        min={0}
                        value={level.sort_order}
                        onChange={(e) =>
                          updateLevel(
                            index,
                            'sort_order',
                            Number(e.target.value),
                          )
                        }
                        aria-label="Порядок сортировки"
                        className={cn(
                          'w-20 px-3 py-1.5 rounded-lg border border-surface-border',
                          'text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                        )}
                      />
                    </td>
                    <td className="py-2">
                      <button
                        type="button"
                        onClick={() => removeLevel(index)}
                        className="p-1.5 text-neutral-400 hover:text-red-600 transition-colors rounded-lg hover:bg-red-50"
                        aria-label="Удалить уровень"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Save button */}
      <div className="mt-6">
        <button
          type="button"
          onClick={handleSave}
          disabled={saving}
          className={cn(
            'px-6 py-2.5 rounded-lg text-sm font-medium',
            'bg-accent text-white hover:bg-accent/90 transition-colors',
            'disabled:opacity-50 disabled:cursor-not-allowed',
          )}
        >
          {saving ? 'Сохранение...' : 'Сохранить изменения'}
        </button>
      </div>
    </div>
  )
}
