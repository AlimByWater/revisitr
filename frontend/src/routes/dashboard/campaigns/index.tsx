import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useCampaignsQuery, useScenariosQuery } from '@/features/campaigns/queries'
import { Mail, Plus, RotateCcw, Zap } from 'lucide-react'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'
import type { Campaign, AutoScenario } from '@/features/campaigns/types'

const statusConfig: Record<
  Campaign['status'],
  { label: string; className: string }
> = {
  draft: {
    label: 'Черновик',
    className: 'bg-neutral-100 text-neutral-500',
  },
  scheduled: {
    label: 'Запланировано',
    className: 'bg-blue-100 text-blue-700',
  },
  sending: {
    label: 'Отправляется',
    className: 'bg-yellow-100 text-yellow-700',
  },
  sent: {
    label: 'Отправлено',
    className: 'bg-green-100 text-green-700',
  },
  completed: {
    label: 'Завершено',
    className: 'bg-emerald-100 text-emerald-700',
  },
  failed: {
    label: 'Ошибка',
    className: 'bg-red-100 text-red-700',
  },
}

const triggerLabels: Record<AutoScenario['trigger_type'], string> = {
  inactive_days: 'Не был N дней',
  visit_count: 'N-й визит',
  bonus_threshold: 'Порог бонусов',
  level_up: 'Новый уровень',
  birthday: 'День рождения',
  holiday: 'Праздник',
  registration: 'Регистрация',
  level_change: 'Смена уровня',
}

type TabType = 'active' | 'archive'

interface CampaignRow {
  kind: 'campaign'
  id: number
  name: string
  type: 'manual' | 'auto'
  status: Campaign['status']
  statusLabel: string
  statusClassName: string
  total: number
  sent: number
  date: string
}

interface ScenarioRow {
  kind: 'scenario'
  id: number
  name: string
  triggerType: AutoScenario['trigger_type']
  isActive: boolean
  date: string
}

type ListRow = CampaignRow | ScenarioRow

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export default function CampaignsPage() {
  const navigate = useNavigate()
  const [tab, setTab] = useState<TabType>('active')

  const {
    data: campaignsData,
    isLoading: campaignsLoading,
    isError: campaignsError,
    mutate: campaignsMutate,
  } = useCampaignsQuery()

  const {
    data: scenariosData,
    isLoading: scenariosLoading,
    isError: scenariosError,
    mutate: scenariosMutate,
  } = useScenariosQuery()

  const isLoading = campaignsLoading || scenariosLoading
  const isError = campaignsError || scenariosError

  const campaigns = campaignsData?.items ?? []
  const scenarios = (scenariosData ?? []) as AutoScenario[]

  const { activeRows, archiveRows } = useMemo(() => {
    const active: ListRow[] = []
    const archive: ListRow[] = []

    for (const c of campaigns) {
      const status = statusConfig[c.status]
      const row: CampaignRow = {
        kind: 'campaign',
        id: c.id,
        name: c.name,
        type: c.type,
        status: c.status,
        statusLabel: status.label,
        statusClassName: status.className,
        total: c.stats.total,
        sent: c.stats.sent,
        date: c.created_at,
      }

      if (['sent', 'completed', 'failed'].includes(c.status)) {
        archive.push(row)
      } else {
        active.push(row)
      }
    }

    for (const s of scenarios) {
      const row: ScenarioRow = {
        kind: 'scenario',
        id: s.id,
        name: s.name,
        triggerType: s.trigger_type,
        isActive: s.is_active,
        date: s.created_at,
      }

      if (s.is_active) {
        active.push(row)
      } else {
        archive.push(row)
      }
    }

    active.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
    archive.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())

    return { activeRows: active, archiveRows: archive }
  }, [campaigns, scenarios])

  const rows = tab === 'active' ? activeRows : archiveRows

  function handleRetry() {
    campaignsMutate()
    scenariosMutate()
  }

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Рассылки</h1>
        <button
          onClick={() => navigate('/dashboard/campaigns/create')}
          type="button"
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
          <span className="hidden sm:inline">Создать рассылку</span>
          <span className="sm:hidden">Создать</span>
        </button>
      </div>

      {/* Tab switcher */}
      <div className="flex gap-1 p-1 bg-neutral-100 rounded-lg w-fit mb-6 animate-in">
        <button
          type="button"
          onClick={() => setTab('active')}
          className={cn(
            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-150',
            tab === 'active'
              ? 'bg-white text-neutral-900 shadow-sm'
              : 'text-neutral-500 hover:text-neutral-700',
          )}
        >
          Активные ({activeRows.length})
        </button>
        <button
          type="button"
          onClick={() => setTab('archive')}
          className={cn(
            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-150',
            tab === 'archive'
              ? 'bg-white text-neutral-900 shadow-sm'
              : 'text-neutral-500 hover:text-neutral-700',
          )}
        >
          Архив ({archiveRows.length})
        </button>
      </div>

      {isLoading ? (
        <div className="animate-in animate-in-delay-1">
          <TableSkeleton />
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить рассылки"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={handleRetry}
        />
      ) : rows.length === 0 ? (
        tab === 'active' ? (
          <EmptyState
            icon={Mail}
            title="Нет активных рассылок"
            description="Создайте рассылку или авто-сценарий, чтобы начать отправку сообщений клиентам."
            actionLabel="Создать рассылку"
            onAction={() => navigate('/dashboard/campaigns/create')}
            variant="campaigns"
          />
        ) : (
          <EmptyState
            icon={Mail}
            title="Архив пуст"
            description="Завершённые рассылки и неактивные сценарии появятся здесь."
          />
        )
      ) : (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden animate-in animate-in-delay-1">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Название
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Тип
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Статус
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Охват
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Отправлено
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Дата
                  </th>
                  {tab === 'archive' && (
                    <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                      Действие
                    </th>
                  )}
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {rows.map((row) => {
                  if (row.kind === 'campaign') {
                    return (
                      <tr
                        key={`c-${row.id}`}
                        className="hover:bg-neutral-50 transition-colors cursor-pointer"
                        onClick={() => navigate(`/dashboard/campaigns/${row.id}`)}
                      >
                        <td className="px-6 py-4">
                          <span className="text-sm font-medium text-neutral-900">
                            {row.name}
                          </span>
                        </td>
                        <td className="px-6 py-4 hidden sm:table-cell">
                          <span className={cn(
                            'inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full',
                            row.type === 'auto'
                              ? 'bg-violet-100 text-violet-700'
                              : 'bg-neutral-100 text-neutral-600',
                          )}>
                            {row.type === 'auto' && <Zap className="w-3 h-3" />}
                            {row.type === 'manual' ? 'Ручная' : 'Авто'}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <span
                            className={cn(
                              'text-xs font-medium px-2 py-1 rounded-full',
                              row.statusClassName,
                            )}
                          >
                            {row.statusLabel}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-right hidden md:table-cell">
                          <span className="text-sm font-mono text-neutral-500 tabular-nums">
                            {row.total}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-right hidden md:table-cell">
                          <span className="text-sm font-mono text-neutral-500 tabular-nums">
                            {row.sent}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-right hidden sm:table-cell">
                          <span className="text-sm font-mono text-neutral-400 tabular-nums">
                            {formatDate(row.date)}
                          </span>
                        </td>
                        {tab === 'archive' && (
                          <td className="px-6 py-4 text-right hidden sm:table-cell">
                            <button
                              type="button"
                              onClick={(e) => {
                                e.stopPropagation()
                                navigate(`/dashboard/campaigns/create?clone=${row.id}`)
                              }}
                              className={cn(
                                'inline-flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-lg',
                                'text-neutral-600 bg-neutral-100 hover:bg-neutral-200',
                                'transition-colors duration-150',
                              )}
                            >
                              <RotateCcw className="w-3 h-3" />
                              Повторить
                            </button>
                          </td>
                        )}
                      </tr>
                    )
                  }

                  // Scenario row
                  return (
                    <tr
                      key={`s-${row.id}`}
                      className="hover:bg-neutral-50 transition-colors cursor-pointer"
                      onClick={() => navigate(`/dashboard/campaigns/${row.id}`)}
                    >
                      <td className="px-6 py-4">
                        <span className="text-sm font-medium text-neutral-900">
                          {row.name}
                        </span>
                      </td>
                      <td className="px-6 py-4 hidden sm:table-cell">
                        <span className={cn(
                          'inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full',
                          'bg-violet-100 text-violet-700',
                        )}>
                          <Zap className="w-3 h-3" />
                          Авто
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={cn(
                            'text-xs font-medium px-2 py-1 rounded-full',
                            row.isActive
                              ? 'bg-green-100 text-green-700'
                              : 'bg-neutral-100 text-neutral-500',
                          )}
                        >
                          {row.isActive ? 'Активен' : 'Неактивен'}
                        </span>
                        <span className="ml-2 text-xs text-neutral-400">
                          {triggerLabels[row.triggerType]}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden md:table-cell">
                        <span className="text-sm font-mono text-neutral-400">
                          &mdash;
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden md:table-cell">
                        <span className="text-sm font-mono text-neutral-400">
                          &mdash;
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden sm:table-cell">
                        <span className="text-sm font-mono text-neutral-400 tabular-nums">
                          {formatDate(row.date)}
                        </span>
                      </td>
                      {tab === 'archive' && (
                        <td className="px-6 py-4 text-right hidden sm:table-cell">
                          {/* No repeat action for scenarios */}
                        </td>
                      )}
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
