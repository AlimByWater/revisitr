import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import {
  useScenariosQuery,
  useCreateScenarioMutation,
  useUpdateScenarioMutation,
  useDeleteScenarioMutation,
} from '@/features/campaigns/queries'
import { Plus, Trash2, Zap } from 'lucide-react'
import type { AutoScenario } from '@/features/campaigns/types'

const triggerLabels: Record<AutoScenario['trigger_type'], string> = {
  inactive_days: 'Не был N дней',
  visit_count: 'N-й визит',
  bonus_threshold: 'Порог бонусов',
  level_up: 'Новый уровень',
  birthday: 'День рождения',
}

export default function ScenariosPage() {
  const { data: scenarios, isLoading, isError } = useScenariosQuery()
  const { data: bots } = useBotsQuery()
  const createMutation = useCreateScenarioMutation()
  const updateMutation = useUpdateScenarioMutation()
  const deleteMutation = useDeleteScenarioMutation()

  const [showForm, setShowForm] = useState(false)
  const [formName, setFormName] = useState('')
  const [formBotId, setFormBotId] = useState<number | ''>('')
  const [formTrigger, setFormTrigger] = useState('')
  const [formDays, setFormDays] = useState('')
  const [formCount, setFormCount] = useState('')
  const [formThreshold, setFormThreshold] = useState('')
  const [formMessage, setFormMessage] = useState('')

  function resetForm() {
    setFormName('')
    setFormBotId('')
    setFormTrigger('')
    setFormDays('')
    setFormCount('')
    setFormThreshold('')
    setFormMessage('')
    setShowForm(false)
  }

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!formName || !formBotId || !formTrigger || !formMessage) return

    const triggerConfig: { days?: number; count?: number; threshold?: number } =
      {}
    if (formDays) triggerConfig.days = Number(formDays)
    if (formCount) triggerConfig.count = Number(formCount)
    if (formThreshold) triggerConfig.threshold = Number(formThreshold)

    createMutation.mutate(
      {
        bot_id: formBotId as number,
        name: formName.trim(),
        trigger_type: formTrigger,
        trigger_config: triggerConfig,
        message: formMessage.trim(),
      },
      { onSuccess: resetForm },
    )
  }

  function handleToggle(scenario: AutoScenario) {
    updateMutation.mutate({
      id: scenario.id,
      data: { is_active: !scenario.is_active },
    })
  }

  function handleDelete(id: number) {
    if (!confirm('Удалить сценарий?')) return
    deleteMutation.mutate(id)
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl">
        <div className="flex items-center justify-between mb-6">
          <h1 className="font-serif text-2xl font-bold text-neutral-900 tracking-tight">
            Авто-сценарии
          </h1>
        </div>
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-neutral-300 border-t-neutral-900 rounded-full animate-spin" />
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="max-w-4xl">
        <div className="flex items-center justify-between mb-6">
          <h1 className="font-serif text-2xl font-bold text-neutral-900 tracking-tight">
            Авто-сценарии
          </h1>
        </div>
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-red-600">
            Не удалось загрузить сценарии. Попробуйте обновить страницу.
          </p>
        </div>
      </div>
    )
  }

  const hasScenarios = scenarios && scenarios.length > 0

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="font-serif text-2xl font-bold text-neutral-900 tracking-tight">Авто-сценарии</h1>
        <button
          onClick={() => setShowForm(true)}
          type="button"
          className={cn(
            'flex items-center gap-2 py-2.5 px-4 rounded-lg',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent/90 active:bg-accent/80',
            'transition-colors',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          Добавить сценарий
        </button>
      </div>

      {/* Create form */}
      {showForm && (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-4">
          <h2 className="text-lg font-semibold text-neutral-900 mb-4">
            Новый сценарий
          </h2>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label
                  htmlFor="scenario-name"
                  className="block text-sm font-medium text-neutral-700 mb-1.5"
                >
                  Название
                </label>
                <input
                  id="scenario-name"
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder="Название сценария"
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                    'text-sm text-neutral-900 placeholder:text-neutral-400',
                    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  )}
                />
              </div>
              <div>
                <label
                  htmlFor="scenario-bot"
                  className="block text-sm font-medium text-neutral-700 mb-1.5"
                >
                  Бот
                </label>
                <select
                  id="scenario-bot"
                  value={formBotId}
                  onChange={(e) =>
                    setFormBotId(e.target.value ? Number(e.target.value) : '')
                  }
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                    'text-sm text-neutral-900 bg-white',
                    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  )}
                >
                  <option value="">Выберите бота</option>
                  {bots?.map((bot) => (
                    <option key={bot.id} value={bot.id}>
                      {bot.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label
                  htmlFor="scenario-trigger"
                  className="block text-sm font-medium text-neutral-700 mb-1.5"
                >
                  Триггер
                </label>
                <select
                  id="scenario-trigger"
                  value={formTrigger}
                  onChange={(e) => setFormTrigger(e.target.value)}
                  className={cn(
                    'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                    'text-sm text-neutral-900 bg-white',
                    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  )}
                >
                  <option value="">Выберите триггер</option>
                  {Object.entries(triggerLabels).map(([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                {formTrigger === 'inactive_days' && (
                  <>
                    <label
                      htmlFor="trigger-days"
                      className="block text-sm font-medium text-neutral-700 mb-1.5"
                    >
                      Количество дней
                    </label>
                    <input
                      id="trigger-days"
                      type="number"
                      value={formDays}
                      onChange={(e) => setFormDays(e.target.value)}
                      placeholder="30"
                      className={cn(
                        'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                        'text-sm text-neutral-900 placeholder:text-neutral-400',
                        'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                      )}
                    />
                  </>
                )}
                {formTrigger === 'visit_count' && (
                  <>
                    <label
                      htmlFor="trigger-count"
                      className="block text-sm font-medium text-neutral-700 mb-1.5"
                    >
                      Номер визита
                    </label>
                    <input
                      id="trigger-count"
                      type="number"
                      value={formCount}
                      onChange={(e) => setFormCount(e.target.value)}
                      placeholder="5"
                      className={cn(
                        'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                        'text-sm text-neutral-900 placeholder:text-neutral-400',
                        'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                      )}
                    />
                  </>
                )}
                {formTrigger === 'bonus_threshold' && (
                  <>
                    <label
                      htmlFor="trigger-threshold"
                      className="block text-sm font-medium text-neutral-700 mb-1.5"
                    >
                      Порог бонусов
                    </label>
                    <input
                      id="trigger-threshold"
                      type="number"
                      value={formThreshold}
                      onChange={(e) => setFormThreshold(e.target.value)}
                      placeholder="100"
                      className={cn(
                        'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                        'text-sm text-neutral-900 placeholder:text-neutral-400',
                        'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                      )}
                    />
                  </>
                )}
              </div>
            </div>

            <div>
              <label
                htmlFor="scenario-message"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Сообщение
              </label>
              <textarea
                id="scenario-message"
                value={formMessage}
                onChange={(e) => setFormMessage(e.target.value)}
                rows={4}
                placeholder="Текст сообщения..."
                className={cn(
                  'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                  'text-sm text-neutral-900 placeholder:text-neutral-400 resize-none',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                )}
              />
            </div>

            <div className="flex items-center justify-end gap-3">
              <button
                type="button"
                onClick={resetForm}
                className={cn(
                  'px-4 py-2.5 rounded-lg text-sm font-medium',
                  'border border-neutral-200 text-neutral-700',
                  'hover:bg-neutral-50 transition-colors',
                )}
              >
                Отмена
              </button>
              <button
                type="submit"
                disabled={createMutation.isPending}
                className={cn(
                  'px-4 py-2.5 rounded-lg text-sm font-medium',
                  'bg-accent text-white',
                  'hover:bg-accent/90 active:bg-accent/80',
                  'transition-colors',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20',
                )}
              >
                {createMutation.isPending ? 'Создание...' : 'Создать'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Scenarios list */}
      {!hasScenarios && !showForm ? (
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mx-auto mb-4">
            <Zap className="w-8 h-8 text-neutral-400" />
          </div>
          <h2 className="text-lg font-semibold text-neutral-700 mb-2">
            Нет авто-сценариев
          </h2>
          <p className="text-sm text-neutral-400 max-w-md mx-auto mb-6">
            Настройте автоматические сценарии для отправки сообщений клиентам по
            триггерам
          </p>
          <button
            onClick={() => setShowForm(true)}
            type="button"
            className={cn(
              'inline-flex items-center gap-2 py-2.5 px-4 rounded-lg',
              'bg-accent text-white text-sm font-medium',
              'hover:bg-accent/90 active:bg-accent/80',
              'transition-colors',
              'focus:outline-none focus:ring-2 focus:ring-accent/20',
            )}
          >
            <Plus className="w-4 h-4" />
            Добавить сценарий
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {scenarios?.map((scenario) => (
            <div
              key={scenario.id}
              className="bg-white rounded-2xl shadow-sm border border-surface-border p-5"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <h3 className="font-semibold text-neutral-900">
                      {scenario.name}
                    </h3>
                    <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-neutral-100 text-neutral-500">
                      {triggerLabels[scenario.trigger_type]}
                    </span>
                  </div>
                  <p className="text-sm text-neutral-500 line-clamp-2">
                    {scenario.message}
                  </p>
                </div>
                <div className="flex items-center gap-3 ml-4">
                  {/* Toggle switch */}
                  <button
                    onClick={() => handleToggle(scenario)}
                    type="button"
                    className={cn(
                      'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                      scenario.is_active ? 'bg-accent' : 'bg-neutral-200',
                    )}
                    aria-label={
                      scenario.is_active ? 'Деактивировать' : 'Активировать'
                    }
                  >
                    <span
                      className={cn(
                        'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
                        scenario.is_active ? 'translate-x-6' : 'translate-x-1',
                      )}
                    />
                  </button>
                  <button
                    onClick={() => handleDelete(scenario.id)}
                    type="button"
                    className="p-1.5 rounded-lg text-neutral-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                    aria-label="Удалить сценарий"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
