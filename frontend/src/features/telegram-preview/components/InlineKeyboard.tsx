import { cn } from '@/lib/utils'
import { ExternalLink } from 'lucide-react'
import type { InlineButton } from '../types'

const STYLE_CLASSES: Record<string, { light: string; dark: string }> = {
  primary: { light: 'bg-blue-500 text-white', dark: 'bg-blue-600 text-white' },
  success: { light: 'bg-green-500 text-white', dark: 'bg-green-600 text-white' },
  danger: { light: 'bg-red-500 text-white', dark: 'bg-red-600 text-white' },
}

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
            {row.map((btn, bi) => {
              const styleClass = btn.style && STYLE_CLASSES[btn.style]
                ? STYLE_CLASSES[btn.style][isDark ? 'dark' : 'light']
                : undefined

              return (
                <div
                  key={bi}
                  className={cn(
                    'tg-inline-btn flex-1',
                    isDark && !styleClass && 'tg-inline-btn-dark',
                    styleClass && 'rounded-lg',
                  )}
                  style={styleClass ? undefined : undefined}
                >
                  <div className={cn(
                    'flex items-center justify-center gap-1 w-full px-2 py-1.5 text-xs font-medium rounded-lg',
                    styleClass || (isDark ? 'tg-inline-btn-dark' : 'tg-inline-btn'),
                    styleClass,
                  )}>
                    {btn.icon_image_url && (
                      <img
                        src={btn.icon_image_url}
                        alt=""
                        className="w-3.5 h-3.5 rounded-sm object-cover shrink-0"
                      />
                    )}
                    <span className="truncate">{btn.text}</span>
                    {btn.url && (
                      <ExternalLink className="w-3 h-3 ml-0.5 flex-shrink-0 opacity-60" />
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        ))}
      </div>
    </div>
  )
}
