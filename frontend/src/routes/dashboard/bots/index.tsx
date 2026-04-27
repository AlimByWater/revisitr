import { Link, useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { Bot as BotIcon, Plus, Users } from 'lucide-react'
import type { Bot } from '@/features/bots/types'

const statusConfig: Record<Bot['status'], { label: string; className: string }> = {
  active: {
    label: 'Активен',
    className: 'bg-emerald-500/10 text-emerald-700 border border-emerald-500/20',
  },
  inactive: {
    label: 'Неактивен',
    className: 'bg-neutral-100 text-neutral-500 border border-neutral-200',
  },
  pending: {
    label: 'Ожидает',
    className: 'bg-amber-500/10 text-amber-700 border border-amber-500/20',
  },
  error: {
    label: 'Ошибка',
    className: 'bg-red-500/10 text-red-700 border border-red-500/20',
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
  const navigate = useNavigate()
  const { data: bots, isLoading, isError, mutate } = useBotsQuery()

  return (
    <div>
      <div className="flex items-start justify-between gap-4 mb-6 animate-in">
        <div>
          <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Мои боты</h1>
          <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">Список ботов</p>
        </div>
        <Link
          to="/dashboard/bots/create"
          className={cn(
            'inline-flex items-center gap-2 py-2.5 px-4 rounded shrink-0',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          <span className="hidden sm:inline">Создать бота</span>
          <span className="sm:hidden">Создать</span>
        </Link>

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
          onAction={() => navigate('/dashboard/bots/create')}
          variant="bots"
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 animate-in animate-in-delay-2">
          {bots.map((bot, i) => {
            const status = statusConfig[bot.status]
            return (
              <Link
                key={bot.id}
                to={`/dashboard/bots/${bot.id}`}
                className={cn(
                  'bg-white rounded border border-neutral-900 p-6',
                  'cursor-pointer hover:scale-[1.02] transition-transform duration-150',
                  'group',
                  'animate-in',
                  `animate-in-delay-${Math.min(i + 1, 5)}`,
                )}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded bg-neutral-100 flex items-center justify-center group-hover:bg-accent/10 transition-colors duration-200">
                      <BotIcon className="w-5 h-5 text-neutral-500 group-hover:text-accent transition-colors duration-200" />
                    </div>
                    <div className="min-w-0">
                      <h3 className="font-semibold text-neutral-900 truncate">{bot.name}</h3>
                      <p className="text-sm text-neutral-400 truncate">{bot.username ? `@${bot.username}` : '—'}</p>
                    </div>
                  </div>
                  <span
                    className={cn(
                      'font-mono text-[10px] px-2 py-0.5 rounded uppercase tracking-wider',
                      status.className,
                    )}
                  >
                    {status.label}
                  </span>
                </div>

                <div className="flex items-center justify-between mt-4 pt-4 border-t border-neutral-200">
                  <div className="flex items-center gap-1.5 text-neutral-400">
                    <Users className="w-3.5 h-3.5" />
                    <span className="font-mono text-[11px] uppercase tracking-wider tabular-nums">
                      {(bot.client_count ?? 0).toLocaleString('ru-RU')} клиентов
                    </span>
                  </div>
                  <span className="font-mono text-[11px] uppercase tracking-wider tabular-nums text-neutral-400">
                    {formatDate(bot.created_at)}
                  </span>
                </div>
              </Link>
            )
          })}
        </div>
      )}
    </div>
  )
}
