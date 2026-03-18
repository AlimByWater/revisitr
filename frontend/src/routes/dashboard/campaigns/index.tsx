import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useCampaignsQuery } from '@/features/campaigns/queries'
import { Mail, Plus } from 'lucide-react'
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
  const { data, isLoading, isError } = useCampaignsQuery()

  if (isLoading) {
    return (
      <div className="max-w-4xl">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-neutral-900">Рассылки</h1>
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
          <h1 className="text-2xl font-bold text-neutral-900">Рассылки</h1>
        </div>
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-red-600">
            Не удалось загрузить список рассылок. Попробуйте обновить страницу.
          </p>
        </div>
      </div>
    )
  }

  const campaigns = data?.items ?? []
  const hasCampaigns = campaigns.length > 0

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-neutral-900">Рассылки</h1>
        <button
          onClick={() => navigate('/dashboard/campaigns/create')}
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
          Создать рассылку
        </button>
      </div>

      {!hasCampaigns ? (
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mx-auto mb-4">
            <Mail className="w-8 h-8 text-neutral-400" />
          </div>
          <h2 className="text-lg font-semibold text-neutral-700 mb-2">
            У вас пока нет рассылок
          </h2>
          <p className="text-sm text-neutral-400 max-w-md mx-auto mb-6">
            Создайте рассылку, чтобы отправить сообщение вашим клиентам через
            Telegram-бота
          </p>
          <button
            onClick={() => navigate('/dashboard/campaigns/create')}
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
            Создать рассылку
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                  Название
                </th>
                <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                  Тип
                </th>
                <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                  Статус
                </th>
                <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                  Охват
                </th>
                <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                  Отправлено
                </th>
                <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
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
                    <td className="px-6 py-4">
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
                    <td className="px-6 py-4 text-right">
                      <span className="text-sm text-neutral-500">
                        {campaign.stats.total}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <span className="text-sm text-neutral-500">
                        {campaign.stats.sent}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
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
      )}
    </div>
  )
}
