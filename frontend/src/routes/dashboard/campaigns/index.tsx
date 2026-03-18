import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useCampaignsQuery } from '@/features/campaigns/queries'
import { Mail, Plus } from 'lucide-react'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'
import type { Campaign } from '@/features/campaigns/types'

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
  failed: {
    label: 'Ошибка',
    className: 'bg-red-100 text-red-700',
  },
}

const typeLabels: Record<Campaign['type'], string> = {
  manual: 'Ручная',
  auto: 'Авто',
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export default function CampaignsPage() {
  const navigate = useNavigate()
  const { data, isLoading, isError, mutate } = useCampaignsQuery()

  const campaigns = data?.items ?? []

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

      {isLoading ? (
        <div className="animate-in animate-in-delay-1">
          <TableSkeleton />
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить рассылки"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : campaigns.length === 0 ? (
        <EmptyState
          icon={Mail}
          title="У вас пока нет рассылок"
          description="Создайте рассылку, чтобы отправить сообщение вашим клиентам через Telegram-бота."
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
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Охват
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Отправлено
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Дата
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {campaigns.map((campaign) => {
                  const status = statusConfig[campaign.status]
                  return (
                    <tr
                      key={campaign.id}
                      className="hover:bg-neutral-50 transition-colors cursor-pointer"
                      onClick={() =>
                        navigate(`/dashboard/campaigns/${campaign.id}`)
                      }
                    >
                      <td className="px-6 py-4">
                        <span className="text-sm font-medium text-neutral-900">
                          {campaign.name}
                        </span>
                      </td>
                      <td className="px-6 py-4 hidden sm:table-cell">
                        <span className="text-sm text-neutral-500">
                          {typeLabels[campaign.type]}
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
                        <span className="text-sm text-neutral-500 tabular-nums">
                          {campaign.stats.total}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden md:table-cell">
                        <span className="text-sm text-neutral-500 tabular-nums">
                          {campaign.stats.sent}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden sm:table-cell">
                        <span className="text-sm text-neutral-400">
                          {formatDate(campaign.created_at)}
                        </span>
                      </td>
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
