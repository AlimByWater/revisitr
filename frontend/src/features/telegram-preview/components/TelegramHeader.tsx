import { cn } from '@/lib/utils'
import { ChevronLeft } from 'lucide-react'

interface TelegramHeaderProps {
  botName: string
  botAvatar?: string
  theme?: 'light' | 'dark'
}

export function TelegramHeader({ botName, botAvatar, theme = 'light' }: TelegramHeaderProps) {
  const isDark = theme === 'dark'

  return (
    <div
      className={cn(
        'flex items-center gap-3 px-4 py-2.5',
        isDark ? 'bg-[#17212B]' : 'bg-white border-b border-gray-200'
      )}
    >
      <ChevronLeft
        className={cn('w-5 h-5 flex-shrink-0', isDark ? 'text-[#6AB2F2]' : 'text-[#3A8EEE]')}
      />
      <div
        className={cn(
          'w-9 h-9 rounded-full flex-shrink-0 flex items-center justify-center text-white font-semibold text-sm',
          'bg-gradient-to-br from-[#5B9BD5] to-[#3A7BD5]'
        )}
      >
        {botAvatar ? (
          <img src={botAvatar} alt={botName} className="w-full h-full rounded-full object-cover" />
        ) : (
          botName.charAt(0).toUpperCase()
        )}
      </div>
      <div className="min-w-0 flex-1">
        <div
          className={cn(
            'text-sm font-semibold truncate',
            isDark ? 'text-white' : 'text-black'
          )}
        >
          {botName}
        </div>
        <div className={cn('text-xs', isDark ? 'text-[#6D7F8F]' : 'text-gray-500')}>bot</div>
      </div>
    </div>
  )
}
