import { useNavigate, useParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  useCampaignQuery,
  useSendCampaignMutation,
  useDeleteCampaignMutation,
} from '@/features/campaigns/queries'
import { ArrowLeft, Send, Trash2, Users, CheckCircle, XCircle } from 'lucide-react'
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

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function CampaignDetailPage() {
  const navigate = useNavigate()
  const { campaignId } = useParams<{ campaignId: string }>()
  const id = Number(campaignId)

  const { data: campaign, isLoading, isError } = useCampaignQuery(id)
  const sendMutation = useSendCampaignMutation()
  const deleteMutation = useDeleteCampaignMutation()

  function handleSend() {
    if (!confirm('Отправить рассылку? Это действие нельзя отменить.')) return
    sendMutation.mutate(id)
  }

  function handleDelete() {
    if (!confirm('Удалить рассылку?')) return
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

  if (isError || !campaign) {
    return (
      <div className="max-w-3xl">
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-red-600">
            Рассылка не найдена или произошла ошибка.
          </p>
        </div>
      </div>
    )
  }

  const status = statusConfig[campaign.status]
  const isDraft = campaign.status === 'draft'
  const isSent = campaign.status === 'sent'

  return (
    <div className="max-w-3xl">
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
            {campaign.name}
          </h1>
        </div>
        <span
          className={cn(
            'text-xs font-medium px-2.5 py-1 rounded-full',
            status.className,
          )}
        >
          {status.label}
        </span>
      </div>

      {/* Campaign info */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-4">
        <h2 className="text-sm font-medium text-neutral-400 uppercase tracking-wider mb-4">
          Детали рассылки
        </h2>
        <div className="space-y-4">
          <div>
            <p className="text-xs text-neutral-400 mb-1">Сообщение</p>
            <p className="text-sm text-neutral-900 whitespace-pre-wrap bg-neutral-50 rounded-lg p-3">
              {campaign.message}
            </p>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-xs text-neutral-400 mb-1">Создано</p>
              <p className="text-sm text-neutral-700">
                {formatDate(campaign.created_at)}
              </p>
            </div>
            {campaign.sent_at && (
              <div>
                <p className="text-xs text-neutral-400 mb-1">Отправлено</p>
                <p className="text-sm text-neutral-700">
                  {formatDate(campaign.sent_at)}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Stats cards (if sent) */}
      {isSent && (
        <div className="grid grid-cols-3 gap-4 mb-4">
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-5">
            <div className="flex items-center gap-2 mb-2">
              <Users className="w-4 h-4 text-neutral-400" />
              <span className="text-xs text-neutral-400">Всего</span>
            </div>
            <p className="text-2xl font-bold font-mono text-neutral-900 tracking-tight">
              {campaign.stats.total}
            </p>
          </div>
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-5">
            <div className="flex items-center gap-2 mb-2">
              <CheckCircle className="w-4 h-4 text-green-500" />
              <span className="text-xs text-neutral-400">Доставлено</span>
            </div>
            <p className="text-2xl font-bold font-mono text-green-600 tracking-tight">
              {campaign.stats.sent}
            </p>
          </div>
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-5">
            <div className="flex items-center gap-2 mb-2">
              <XCircle className="w-4 h-4 text-red-500" />
              <span className="text-xs text-neutral-400">Ошибки</span>
            </div>
            <p className="text-2xl font-bold font-mono text-red-600 tracking-tight">
              {campaign.stats.failed}
            </p>
          </div>
        </div>
      )}

      {/* Actions */}
      {isDraft && (
        <div className="flex items-center gap-3">
          <button
            onClick={handleSend}
            disabled={sendMutation.isPending}
            type="button"
            className={cn(
              'flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium',
              'bg-accent text-white',
              'hover:bg-accent/90 active:bg-accent/80',
              'transition-colors',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'focus:outline-none focus:ring-2 focus:ring-accent/20',
            )}
          >
            <Send className="w-4 h-4" />
            {sendMutation.isPending ? 'Отправка...' : 'Отправить'}
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
      )}
    </div>
  )
}
