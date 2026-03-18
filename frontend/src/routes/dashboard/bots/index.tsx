import { Link } from 'react-router-dom'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import { CreateBotModal } from '@/components/bots/CreateBotModal'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { Bot as BotIcon, Plus, Users } from 'lucide-react'
import type { Bot } from '@/features/bots/types'

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

export default function BotsPage() {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const { data: bots, isLoading, isError, mutate } = useBotsQuery()

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Боты</h1>
        <button
          onClick={() => setShowCreateModal(true)}
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
          <span className="hidden sm:inline">Создать бота</span>
          <span className="sm:hidden">Создать</span>
        </button>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[0, 1, 2, 3].map((i) => (
            <div key={i} className={cn('animate-in', `animate-in-delay-${i + 1}`)}>
              <CardSkeleton />
            </div>
          ))}
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить ботов"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : !bots || bots.length === 0 ? (
        <EmptyState
          icon={BotIcon}
          title="У вас пока нет ботов"
          description="Создайте Telegram-бота для вашего ресторана, чтобы начать работу с программой лояльности и рассылками."
          actionLabel="Создать бота"
          onAction={() => setShowCreateModal(true)}
          variant="bots"
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {bots.map((bot, i) => {
            const status = statusConfig[bot.status]
            return (
              <Link
                key={bot.id}
                to={`/dashboard/bots/${bot.id}`}
                className={cn(
                  'bg-white rounded-2xl shadow-sm border border-surface-border p-6',
                  'hover:border-neutral-300 hover:shadow-md',
                  'transition-all duration-200 group',
                  'animate-in',
                  `animate-in-delay-${Math.min(i + 1, 5)}`,
                )}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-neutral-100 flex items-center justify-center group-hover:bg-accent/10 transition-colors duration-200">
                      <BotIcon className="w-5 h-5 text-neutral-500 group-hover:text-accent transition-colors duration-200" />
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
                    <span className="font-mono tabular-nums">{bot.client_count ?? 0} клиентов</span>
                  </div>
                  <span className="font-mono tabular-nums">{formatDate(bot.created_at)}</span>
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
