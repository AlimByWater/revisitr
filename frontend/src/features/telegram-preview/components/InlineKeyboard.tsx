import { useState, useCallback } from 'react'
import type { InlineButton } from '../types'

const STYLE_BG: Record<string, { bg: string; text: string }> = {
  primary: { bg: '#3B95F0', text: '#fff' },
  success: { bg: '#59C95D', text: '#fff' },
  danger: { bg: '#E06456', text: '#fff' },
}

const DEFAULT_STYLE = { bg: '#E5E5EA', text: '#333333' }

interface InlineKeyboardProps {
  buttons: InlineButton[][]
  onButtonClick?: (button: InlineButton) => void
}

export function InlineKeyboard({ buttons, onButtonClick }: InlineKeyboardProps) {
  if (!buttons.length) return null

  return (
    <div className="mt-[2px] w-full space-y-[2px]">
      {buttons.map((row, ri) => (
        <div key={ri} className="flex gap-[2px]">
          {row.map((btn, bi) => (
            <ActionButton key={bi} button={btn} onClick={onButtonClick} rowLength={row.length} />
          ))}
        </div>
      ))}
    </div>
  )
}

function ActionButton({
  button,
  onClick,
  rowLength,
}: {
  button: InlineButton
  onClick?: (button: InlineButton) => void
  rowLength: number
}) {
  const [showTooltip, setShowTooltip] = useState(false)
  const colors = button.style ? STYLE_BG[button.style] ?? DEFAULT_STYLE : DEFAULT_STYLE

  const handleClick = useCallback(() => {
    if (button.url) {
      setShowTooltip(true)
      setTimeout(() => setShowTooltip(false), 1500)
      return
    }
    onClick?.(button)
  }, [button, onClick])

  return (
    <button
      type="button"
      className={`relative flex h-[36px] min-w-0 cursor-pointer items-center justify-center overflow-hidden rounded-[11px] px-[6px] ${
        rowLength === 4 ? 'flex-[1_1_0%]' : 'flex-1'
      }`}
      style={{ background: colors.bg }}
      onClick={handleClick}
    >
      <div className="flex min-w-0 items-center gap-[4px]">
        {button.icon_image_url && (
          <img
            src={button.icon_image_url}
            alt=""
            className="h-[16px] w-[16px] shrink-0 rounded-sm object-cover"
          />
        )}
        <span
          className="truncate text-[13px] font-normal leading-[15px] tracking-[-0.16px]"
          style={{ color: colors.text }}
        >
          {button.text}
        </span>
      </div>

      {button.url && (
        <svg className="tg-url-arrow" viewBox="0 0 12 12" fill="none">
          <path d="M3 1h8v8M11 1 3 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      )}

      {showTooltip && <div className="tg-tooltip">Откроет ссылку</div>}
    </button>
  )
}
