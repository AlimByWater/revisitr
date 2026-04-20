import { cn } from '@/lib/utils'
import { ExternalLink } from 'lucide-react'
import type { InlineButton } from '../types'

const STYLE_COLORS: Record<string, string> = {
  primary: 'bg-[#5AB4F0] text-white',
  success: 'bg-[#59C95D] text-white',
  danger: 'bg-[#E06456] text-white',
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
      <div className="max-w-[85%] w-full space-y-[3px]">
        {buttons.map((row, ri) => (
          <div key={ri} className="flex gap-[3px]">
            {row.map((btn, bi) => {
              const styleColor = btn.style ? STYLE_COLORS[btn.style] : undefined

              return (
                <div
                  key={bi}
                  className={cn(
                    'flex-1 flex items-center justify-center gap-1 px-2 py-[7px] rounded-[8px] text-[13px] font-medium leading-tight',
                    styleColor || (isDark
                      ? 'bg-[#2A2A2A] text-[#5AB4F0]'
                      : 'bg-[#ECF1F7] text-[#3B90C5]'),
                  )}
                >
                  {btn.icon_image_url && (
                    <img
                      src={btn.icon_image_url}
                      alt=""
                      className="w-[15px] h-[15px] rounded-sm object-cover shrink-0"
                    />
                  )}
                  <span className="truncate">{btn.text}</span>
                  {btn.url && (
                    <ExternalLink className="w-[10px] h-[10px] flex-shrink-0 opacity-50" />
                  )}
                </div>
              )
            })}
          </div>
        ))}
      </div>
    </div>
  )
}
