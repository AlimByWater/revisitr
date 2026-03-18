import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import { CreateBotModal } from '@/components/bots/CreateBotModal'
import { Bot as BotIcon, Plus, Users } from 'lucide-react'
import type { Bot } from '@/features/bots/types'

export const Route = createFileRoute('/dashboard/bots/')({
  component: BotsPage,
})

const statusConfig: Record<Bot['status'], { label: string; className: string }> = {
  active: {
    label: 'Активен',
    className: 'bg-green-100 text-green-700',
  },
  inactive: {
    label: 'Неактивен',
    className: 'bg-neutral-100 text-neutral-500',
  },
  error: {
    label: 'Ошибка',
    className: 'bg-red-100 text-red-700',
  },
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function BotsPage() {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const { data: bots, isLoading, isError } = useBotsQuery()

  if (isLoading) {
    return (
      <div className="max-w-4xl">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-neutral-900">Боты</h1>
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
          <h1 className="text-2xl font-bold text-neutral-900">Боты</h1>
        </div>
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-red-600">
            Не удалось загрузить список ботов. Попробуйте обновить страницу.
          </p>
        </div>
      </div>
    )
  }

  const hasBots = bots && bots.length > 0

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-neutral-900">Боты</h1>
        <button
          onClick={() => setShowCreateModal(true)}
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
          Создать бота
        </button>
      </div>

      {!hasBots ? (
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mx-auto mb-4">
            <BotIcon className="w-8 h-8 text-neutral-400" />
          </div>
          <h2 className="text-lg font-semibold text-neutral-700 mb-2">
            У вас пока нет ботов
          </h2>
          <p className="text-sm text-neutral-400 max-w-md mx-auto mb-6">
            Создайте Telegram-бота для вашего ресторана, чтобы начать работу с программой лояльности
          </p>
          <button
            onClick={() => setShowCreateModal(true)}
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
            Создать бота
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {bots.map((bot) => {
            const status = statusConfig[bot.status]
            return (
              <Link
                key={bot.id}
                to="/dashboard/bots/$botId"
                params={{ botId: String(bot.id) }}
                className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 hover:border-neutral-300 transition-colors group"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-neutral-100 flex items-center justify-center group-hover:bg-neutral-200 transition-colors">
                      <BotIcon className="w-5 h-5 text-neutral-500" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-neutral-900">{bot.name}</h3>
                      <p className="text-sm text-neutral-400">@{bot.username}</p>
                    </div>
                  </div>
                  <span
                    className={cn(
                      'text-xs font-medium px-2 py-1 rounded-full',
                      status.className,
                    )}
                  >
                    {status.label}
                  </span>
                </div>

                <div className="flex items-center justify-between text-sm text-neutral-400 mt-4 pt-4 border-t border-surface-border">
                  <div className="flex items-center gap-1.5">
                    <Users className="w-4 h-4" />
                    <span>{bot.client_count ?? 0} клиентов</span>
                  </div>
                  <span>{formatDate(bot.created_at)}</span>
                </div>
              </Link>
            )
          })}
        </div>
      )}

      {showCreateModal && (
        <CreateBotModal onClose={() => setShowCreateModal(false)} />
      )}
    </div>
  )
}
