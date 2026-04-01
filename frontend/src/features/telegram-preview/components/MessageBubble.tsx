import { cn } from '@/lib/utils'
import type { ReactNode } from 'react'

interface MessageBubbleProps {
  children: ReactNode
  theme?: 'light' | 'dark'
  showTail?: boolean
}

export function MessageBubble({ children, theme = 'light', showTail = false }: MessageBubbleProps) {
  const isDark = theme === 'dark'
  const now = new Date()
  const timestamp = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}`

  return (
    <div className="flex justify-start">
      <div
        className={cn(
          'tg-bubble',
          isDark && 'tg-bubble-dark',
          showTail && 'tg-bubble-tail'
        )}
      >
        <span className="whitespace-pre-wrap">{children}</span>
        <span className={cn('tg-timestamp', isDark && 'tg-timestamp-dark')}>
          {timestamp}
        </span>
      </div>
    </div>
  )
}
