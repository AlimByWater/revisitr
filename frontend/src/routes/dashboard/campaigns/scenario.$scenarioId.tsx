import { useNavigate, useParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  useScenarioQuery,
  useScenariosQuery,
  useUpdateScenarioMutation,
  useDeleteScenarioMutation,
  useActionLogQuery,
} from '@/features/campaigns/queries'
import { ArrowLeft, RotateCcw, Trash2, Zap, Power } from 'lucide-react'
import type { AutoScenario } from '@/features/campaigns/types'

const triggerLabels: Record<AutoScenario['trigger_type'], string> = {
  inactive_days: 'Не был N дней',
  visit_count: 'N-й визит',
  bonus_threshold: 'Порог бонусов',
  level_up: 'Новый уровень',
  birthday: 'День рождения',
  holiday: 'Дата',
  registration: 'Регистрация',
  level_change: 'Смена уровня',
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatTriggerConfig(scenario: AutoScenario): string {
  const tc = scenario.trigger_config
  switch (scenario.trigger_type) {
    case 'inactive_days':
      return `${tc.days} дней`
    case 'visit_count':
      return `${tc.count}-й визит`
    case 'bonus_threshold':
      return `${tc.threshold} бонусов`
    case 'holiday':
      return tc.month && tc.day
        ? `${tc.day}.${String(tc.month).padStart(2, '0')}`
        : ''
    default:
      return ''
  }
}

export default function ScenarioDetailPage() {
  const navigate = useNavigate()
  const { scenarioId } = useParams<{ scenarioId: string }>()
  const id = Number(scenarioId)

  const { data: scenario, isLoading, isError } = useScenarioQuery(id)
  const { mutate: revalidateScenarios } = useScenariosQuery()
  const updateMutation = useUpdateScenarioMutation()
  const deleteMutation = useDeleteScenarioMutation()
  const { data: actionLogData } = useActionLogQuery(id)

  const actionLogs = actionLogData?.items ?? []

  function handleToggleActive() {
    if (!scenario) return
    updateMutation.mutate(
      { id, data: { is_active: !scenario.is_active } },
      { onSuccess: () => revalidateScenarios() },
    )
  }

  function handleDelete() {
    if (!confirm('Удалить сценарий?')) return
    deleteMutation.mutate(id, {
      onSuccess: () => navigate('/dashboard/campaigns'),
    })
  }

  if (isLoading) {
    return (
      <div className="max-w-3xl">
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-neutral-300 border-t-neutral-900 rounded-full animate-spin" />
        </div>
      </div>
    )
  }

  if (isError || !scenario) {
    return (
      <div className="max-w-3xl">
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-red-600">
            Сценарий не найден или произошла ошибка.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-3xl">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate('/dashboard/campaigns')}
          type="button"
          className="p-2 rounded-lg hover:bg-neutral-100 transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-neutral-500" />
        </button>
        <div className="flex-1">
          <h1 className="font-serif text-2xl font-bold text-neutral-900 tracking-tight">
            {scenario.name}
          </h1>
        </div>
        <span
          className={cn(
            'text-xs font-medium px-2.5 py-1 rounded-full',
            scenario.is_active
              ? 'bg-green-100 text-green-700'
              : 'bg-neutral-100 text-neutral-500',
          )}
        >
          {scenario.is_active ? 'Активен' : 'Неактивен'}
        </span>
      </div>

      {/* Message */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-4">
        <h2 className="text-sm font-medium text-neutral-400 uppercase tracking-wider mb-4">
          Сообщение
        </h2>
        <p className="text-sm text-neutral-900 whitespace-pre-wrap bg-neutral-50 rounded-lg p-3">
          {scenario.message || '(нет сообщения)'}
        </p>
      </div>

      {/* Trigger info */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-4">
        <h2 className="text-sm font-medium text-neutral-400 uppercase tracking-wider mb-4">
          Триггер
        </h2>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-neutral-400 mb-1">Тип</p>
            <span className="inline-flex items-center gap-1 text-sm font-medium text-violet-700">
              <Zap className="w-3.5 h-3.5" />
              {triggerLabels[scenario.trigger_type]}
            </span>
          </div>
          {formatTriggerConfig(scenario) && (
            <div>
              <p className="text-xs text-neutral-400 mb-1">Значение</p>
              <p className="text-sm font-mono text-neutral-700">
                {formatTriggerConfig(scenario)}
              </p>
            </div>
          )}
          {scenario.timing?.days_before !== undefined && (
            <div>
              <p className="text-xs text-neutral-400 mb-1">Дней до</p>
              <p className="text-sm font-mono text-neutral-700">
                {scenario.timing.days_before}
              </p>
            </div>
          )}
          {scenario.timing?.days_after !== undefined && (
            <div>
              <p className="text-xs text-neutral-400 mb-1">Дней после</p>
              <p className="text-sm font-mono text-neutral-700">
                {scenario.timing.days_after}
              </p>
            </div>
          )}
          <div>
            <p className="text-xs text-neutral-400 mb-1">Создан</p>
            <p className="text-sm font-mono text-neutral-700 tabular-nums">
              {formatDate(scenario.created_at)}
            </p>
          </div>
        </div>
      </div>

      {/* Action log */}
      {actionLogs.length > 0 && (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-4">
          <h2 className="text-sm font-medium text-neutral-400 uppercase tracking-wider mb-4">
            Журнал действий
          </h2>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-3 py-2">
                    Дата
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-3 py-2">
                    Клиент
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-3 py-2">
                    Результат
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {actionLogs.map((log) => (
                  <tr key={log.id}>
                    <td className="px-3 py-2">
                      <span className="text-xs font-mono text-neutral-500 tabular-nums">
                        {formatDate(log.executed_at)}
                      </span>
                    </td>
                    <td className="px-3 py-2">
                      <span className="text-xs text-neutral-700">
                        #{log.client_id}
                      </span>
                    </td>
                    <td className="px-3 py-2">
                      <span
                        className={cn(
                          'text-xs font-medium px-2 py-0.5 rounded-full',
                          log.result === 'success' && 'bg-green-100 text-green-700',
                          log.result === 'failed' && 'bg-red-100 text-red-700',
                          log.result === 'skipped' && 'bg-neutral-100 text-neutral-500',
                        )}
                      >
                        {log.result === 'success' ? 'Успех' : log.result === 'failed' ? 'Ошибка' : 'Пропущен'}
                      </span>
                      {log.error_msg && (
                        <span className="ml-2 text-xs text-red-400">{log.error_msg}</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-3">
        <button
          onClick={handleToggleActive}
          disabled={updateMutation.isPending}
          type="button"
          className={cn(
            'flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium',
            scenario.is_active
              ? 'border border-neutral-200 text-neutral-700 hover:bg-neutral-50'
              : 'bg-accent text-white hover:bg-accent/90',
            'transition-colors',
            'disabled:opacity-50 disabled:cursor-not-allowed',
          )}
        >
          <Power className="w-4 h-4" />
          {updateMutation.isPending
            ? 'Обновление...'
            : scenario.is_active
              ? 'Деактивировать'
              : 'Активировать'}
        </button>
        <button
          type="button"
          onClick={() =>
            navigate(`/dashboard/campaigns/create?clone=${id}&type=scenario`)
          }
          className={cn(
            'flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium',
            'border border-neutral-200 text-neutral-700',
            'hover:bg-neutral-50 transition-colors',
          )}
        >
          <RotateCcw className="w-4 h-4" />
          Повторить
        </button>
        <button
          onClick={handleDelete}
          disabled={deleteMutation.isPending}
          type="button"
          className={cn(
            'flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium',
            'border border-red-200 text-red-600',
            'hover:bg-red-50 transition-colors',
            'disabled:opacity-50 disabled:cursor-not-allowed',
          )}
        >
          <Trash2 className="w-4 h-4" />
          Удалить
        </button>
      </div>
    </div>
  )
}
