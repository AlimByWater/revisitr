import { useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import {
  useCreateCampaignMutation,
  usePreviewAudienceMutation,
} from '@/features/campaigns/queries'
import { ArrowLeft, Eye, Send } from 'lucide-react'
import type { AudienceFilter } from '@/features/campaigns/types'

export default function CreateCampaignPage() {
  const navigate = useNavigate()
  const { data: bots } = useBotsQuery()
  const createMutation = useCreateCampaignMutation()
  const previewMutation = usePreviewAudienceMutation()

  const [name, setName] = useState('')
  const [botId, setBotId] = useState<number | ''>('')
  const [message, setMessage] = useState('')
  const [audienceFilter, setAudienceFilter] = useState<AudienceFilter>({})

  const isValid = name.trim() !== '' && botId !== '' && message.trim() !== ''

  function handleBotChange(value: string) {
    const id = value ? Number(value) : ''
    setBotId(id)
    if (id) {
      setAudienceFilter((prev) => ({ ...prev, bot_id: id as number }))
    } else {
      setAudienceFilter((prev) => {
        const { bot_id, ...rest } = prev
        return rest
      })
    }
  }

  function handlePreview() {
    previewMutation.mutate(audienceFilter)
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!isValid) return

    createMutation.mutate(
      {
        bot_id: botId as number,
        name: name.trim(),
        message: message.trim(),
        audience_filter: audienceFilter,
      },
      {
        onSuccess: () => {
          navigate('/dashboard/campaigns')
        },
      },
    )
  }

  return (
    <div className="max-w-2xl">
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate('/dashboard/campaigns')}
          type="button"
          className="p-2 rounded-lg hover:bg-neutral-100 transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-neutral-500" />
        </button>
        <h1 className="font-serif text-2xl font-bold text-neutral-900 tracking-tight">
          Создать рассылку
        </h1>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 space-y-6">
          {/* Bot selector */}
          <div>
            <label
              htmlFor="bot"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Бот
            </label>
            <select
              id="bot"
              value={botId}
              onChange={(e) => handleBotChange(e.target.value)}
              className={cn(
                'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                'text-sm text-neutral-900 bg-white',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              )}
            >
              <option value="">Выберите бота</option>
              {bots?.map((bot) => (
                <option key={bot.id} value={bot.id}>
                  {bot.name} (@{bot.username})
                </option>
              ))}
            </select>
          </div>

          {/* Campaign name */}
          <div>
            <label
              htmlFor="name"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Название рассылки
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Например: Акция выходного дня"
              className={cn(
                'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                'text-sm text-neutral-900 placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              )}
            />
          </div>

          {/* Message */}
          <div>
            <label
              htmlFor="message"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Сообщение
            </label>
            <textarea
              id="message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              rows={6}
              placeholder="Текст сообщения для клиентов..."
              className={cn(
                'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                'text-sm text-neutral-900 placeholder:text-neutral-400 resize-none',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              )}
            />
          </div>

          {/* Audience preview */}
          <div className="bg-neutral-50 rounded-xl p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-neutral-700">
                  Аудитория
                </p>
                <p className="text-xs text-neutral-400 mt-0.5">
                  {botId
                    ? 'Клиенты выбранного бота'
                    : 'Выберите бота для просмотра аудитории'}
                </p>
              </div>
              <div className="flex items-center gap-3">
                {previewMutation.isSuccess && (
                  <span className="text-sm font-semibold text-neutral-900">
                    {previewMutation.data} клиентов
                  </span>
                )}
                <button
                  type="button"
                  onClick={handlePreview}
                  disabled={!botId || previewMutation.isPending}
                  className={cn(
                    'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium',
                    'border border-neutral-200 text-neutral-700',
                    'hover:bg-white transition-colors',
                    'disabled:opacity-50 disabled:cursor-not-allowed',
                  )}
                >
                  <Eye className="w-4 h-4" />
                  Просмотр
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 mt-6">
          <button
            type="button"
            onClick={() => navigate('/dashboard/campaigns')}
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
            disabled={!isValid || createMutation.isPending}
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
            {createMutation.isPending ? 'Создание...' : 'Создать'}
          </button>
        </div>

        {createMutation.isError && (
          <p className="text-sm text-red-600 mt-3 text-right">
            Не удалось создать рассылку. Попробуйте ещё раз.
          </p>
        )}
      </form>
    </div>
  )
}
