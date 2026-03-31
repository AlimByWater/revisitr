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
  holiday: 'Дата',
  registration: 'Регистрация',
  level_change: 'Смена уровня',
}

type TabType = 'active' | 'archive'

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

  // Active tab: only active scenarios
  // Archive tab: sent/completed/failed campaigns + inactive scenarios (no drafts)
  const { activeScenarios, activeCampaigns, archiveRows } = useMemo(() => {
    const active = scenarios.filter((s) => s.is_active)
    const inactive = scenarios.filter((s) => !s.is_active)

    // Active also includes draft/scheduled/sending campaigns
    const activeCampaigns = campaigns.filter((c) =>
      ['draft', 'scheduled', 'sending'].includes(c.status),
    )

    const terminalCampaigns = campaigns.filter((c) =>
      ['sent', 'completed', 'failed'].includes(c.status),
    )

    type ArchiveRow =
      | { kind: 'campaign'; data: Campaign }
      | { kind: 'scenario'; data: AutoScenario }

    const archive: ArchiveRow[] = [
      ...terminalCampaigns.map((c) => ({ kind: 'campaign' as const, data: c })),
      ...inactive.map((s) => ({ kind: 'scenario' as const, data: s })),
    ]
    archive.sort(
      (a, b) =>
        new Date(b.data.created_at).getTime() - new Date(a.data.created_at).getTime(),
    )

    active.sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    )

    return { activeScenarios: active, activeCampaigns, archiveRows: archive }
  }, [campaigns, scenarios])

  function handleRetry() {
    campaignsMutate()
    scenariosMutate()
  }

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          Все рассылки
        </h1>
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
          Активные ({activeScenarios.length + activeCampaigns.length})
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
      ) : tab === 'active' ? (
        /* ── Active tab: active scenarios + draft/scheduled/sending campaigns ── */
        activeScenarios.length === 0 && activeCampaigns.length === 0 ? (
          <EmptyState
            icon={Mail}
            title="Нет активных рассылок"
            description="Создайте рассылку или авто-сценарий, чтобы начать отправку сообщений клиентам."
            actionLabel="Создать рассылку"
            onAction={() => navigate('/dashboard/campaigns/create')}
            variant="campaigns"
          />
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
                    <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                      Дата
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-surface-border">
                  {activeCampaigns.map((c) => {
                    const status = statusConfig[c.status]
                    return (
                      <tr
                        key={`c-${c.id}`}
                        className="hover:bg-neutral-50 transition-colors cursor-pointer"
                        onClick={() => navigate(`/dashboard/campaigns/${c.id}`)}
                      >
                        <td className="px-6 py-4">
                          <span className="text-sm font-medium text-neutral-900">
                            {c.name}
                          </span>
                        </td>
                        <td className="px-6 py-4 hidden sm:table-cell">
                          <span className="inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full bg-neutral-100 text-neutral-600">
                            Ручная
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <span
                            className={cn(
                              'text-xs font-medium px-2 py-1 rounded-full',
                              status.className,
                            )}
                          >
                            {status.label}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-right hidden sm:table-cell">
                          <span className="text-sm font-mono text-neutral-400 tabular-nums">
                            {formatDate(c.created_at)}
                          </span>
                        </td>
                      </tr>
                    )
                  })}
                  {activeScenarios.map((s) => (
                    <tr
                      key={`s-${s.id}`}
                      className="hover:bg-neutral-50 transition-colors cursor-pointer"
                      onClick={() => navigate(`/dashboard/campaigns/scenario/${s.id}`)}
                    >
                      <td className="px-6 py-4">
                        <span className="text-sm font-medium text-neutral-900">
                          {s.name}
                        </span>
                      </td>
                      <td className="px-6 py-4 hidden sm:table-cell">
                        <span className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full bg-violet-100 text-violet-700">
                          <Zap className="w-3 h-3" />
                          Авто
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="inline-flex items-center gap-1 text-xs font-medium px-2 py-1 rounded-full bg-green-100 text-green-700">
                          Активен
                        </span>
                        <span className="ml-2 text-xs text-neutral-400">
                          {triggerLabels[s.trigger_type]}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden sm:table-cell">
                        <span className="text-sm font-mono text-neutral-400 tabular-nums">
                          {formatDate(s.created_at)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )
      ) : (
        /* ── Archive tab: campaigns + inactive scenarios ──────────── */
        archiveRows.length === 0 ? (
          <EmptyState
            icon={Mail}
            title="Архив пуст"
            description="Завершённые рассылки и неактивные сценарии появятся здесь."
          />
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
                    <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                      Действие
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-surface-border">
                  {archiveRows.map((row) => {
                    if (row.kind === 'campaign') {
                      const c = row.data
                      const status = statusConfig[c.status]
                      return (
                        <tr
                          key={`c-${c.id}`}
                          className="hover:bg-neutral-50 transition-colors cursor-pointer"
                          onClick={() => navigate(`/dashboard/campaigns/${c.id}`)}
                        >
                          <td className="px-6 py-4">
                            <span className="text-sm font-medium text-neutral-900">
                              {c.name}
                            </span>
                          </td>
                          <td className="px-6 py-4 hidden sm:table-cell">
                            <span className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full bg-neutral-100 text-neutral-600">
                              Ручная
                            </span>
                          </td>
                          <td className="px-6 py-4">
                            <span
                              className={cn(
                                'text-xs font-medium px-2 py-1 rounded-full',
                                status.className,
                              )}
                            >
                              {status.label}
                            </span>
                          </td>
                          <td className="px-6 py-4 text-right hidden md:table-cell">
                            <span className="text-sm font-mono text-neutral-500 tabular-nums">
                              {c.stats.total}
                            </span>
                          </td>
                          <td className="px-6 py-4 text-right hidden md:table-cell">
                            <span className="text-sm font-mono text-neutral-500 tabular-nums">
                              {c.stats.sent}
                            </span>
                          </td>
                          <td className="px-6 py-4 text-right hidden sm:table-cell">
                            <span className="text-sm font-mono text-neutral-400 tabular-nums">
                              {formatDate(c.created_at)}
                            </span>
                          </td>
                          <td className="px-6 py-4 text-right hidden sm:table-cell">
                            <button
                              type="button"
                              onClick={(e) => {
                                e.stopPropagation()
                                navigate(`/dashboard/campaigns/create?clone=${c.id}&type=campaign`)
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
                        </tr>
                      )
                    }

                    // Inactive scenario row
                    const s = row.data
                    return (
                      <tr
                        key={`s-${s.id}`}
                        className="hover:bg-neutral-50 transition-colors cursor-pointer"
                        onClick={() => navigate(`/dashboard/campaigns/scenario/${s.id}`)}
                      >
                        <td className="px-6 py-4">
                          <span className="text-sm font-medium text-neutral-900">
                            {s.name}
                          </span>
                        </td>
                        <td className="px-6 py-4 hidden sm:table-cell">
                          <span className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full bg-violet-100 text-violet-700">
                            <Zap className="w-3 h-3" />
                            Авто
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <span className="text-xs font-medium px-2 py-1 rounded-full bg-neutral-100 text-neutral-500">
                            Неактивен
                          </span>
                          <span className="ml-2 text-xs text-neutral-400">
                            {triggerLabels[s.trigger_type]}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-right hidden md:table-cell">
                          <span className="text-sm font-mono text-neutral-400">&mdash;</span>
                        </td>
                        <td className="px-6 py-4 text-right hidden md:table-cell">
                          <span className="text-sm font-mono text-neutral-400">&mdash;</span>
                        </td>
                        <td className="px-6 py-4 text-right hidden sm:table-cell">
                          <span className="text-sm font-mono text-neutral-400 tabular-nums">
                            {formatDate(s.created_at)}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-right hidden sm:table-cell">
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation()
                              navigate(`/dashboard/campaigns/create?clone=${s.id}&type=scenario`)
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
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>
        )
      )}
    </div>
  )
}
