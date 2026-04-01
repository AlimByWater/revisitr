import { cn } from '@/lib/utils'
import { ExternalLink } from 'lucide-react'
import type { InlineButton } from '../types'

interface InlineKeyboardProps {
  buttons: InlineButton[][]
  theme?: 'light' | 'dark'
}

export function InlineKeyboard({ buttons, theme = 'light' }: InlineKeyboardProps) {
  const isDark = theme === 'dark'

  if (!buttons.length) return null

  return (
    <div className="flex justify-start mt-1">
      <div className="max-w-[85%] w-full space-y-1">
        {buttons.map((row, ri) => (
          <div key={ri} className="flex gap-1">
            {row.map((btn, bi) => (
              <div
                key={bi}
                className={cn(
                  'tg-inline-btn flex-1',
                  isDark && 'tg-inline-btn-dark'
                )}
              >
                <span className="truncate">{btn.text}</span>
                {btn.url && (
                  <ExternalLink className="w-3 h-3 ml-1 flex-shrink-0 opacity-60" />
                )}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}
