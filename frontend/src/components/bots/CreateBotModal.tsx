import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useCreateBotMutation } from '@/features/bots/queries'
import { X } from 'lucide-react'

interface CreateBotModalProps {
  onClose: () => void
}

export function CreateBotModal({ onClose }: CreateBotModalProps) {
  const [name, setName] = useState('')
  const [token, setToken] = useState('')
  const createBot = useCreateBotMutation()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      await createBot.mutateAsync({ name, token })
      onClose()
    } catch {
      // error is available via createBot.error
    }
  }

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="create-bot-title"
    >
      <div className="bg-white rounded-2xl p-6 w-full max-w-md mx-4">
        <div className="flex items-center justify-between mb-6">
          <h2 id="create-bot-title" className="text-lg font-semibold text-neutral-900">
            Создать бота
          </h2>
          <button
            onClick={onClose}
            type="button"
            className="p-1 rounded-lg text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
            aria-label="Закрыть"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label
              htmlFor="bot-name"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Имя бота
            </label>
            <input
              id="bot-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Мой ресторан"
              required
              disabled={createBot.isPending}
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            />
          </div>

          <div>
            <label
              htmlFor="bot-token"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Токен
            </label>
            <input
              id="bot-token"
              type="text"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="Токен от @BotFather"
              required
              disabled={createBot.isPending}
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            />
            <p className="mt-1.5 text-xs text-neutral-400">
              Получите токен у @BotFather в Telegram
            </p>
          </div>

          {createBot.isError && (
            <p className="text-sm text-red-600">
              {createBot.error instanceof Error
                ? createBot.error.message
                : 'Не удалось создать бота. Попробуйте снова.'}
            </p>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={createBot.isPending}
              className={cn(
                'flex-1 py-2.5 px-4 rounded-lg',
                'border border-surface-border text-sm font-medium text-neutral-700',
                'hover:bg-neutral-50 active:bg-neutral-100',
                'transition-colors',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={createBot.isPending}
              className={cn(
                'flex-1 py-2.5 px-4 rounded-lg',
                'bg-accent text-white text-sm font-medium',
                'hover:bg-accent/90 active:bg-accent/80',
                'transition-colors',
                'focus:outline-none focus:ring-2 focus:ring-accent/20',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              {createBot.isPending ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
