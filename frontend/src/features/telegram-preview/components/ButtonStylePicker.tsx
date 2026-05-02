import { useState, useRef, useEffect } from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'

const BUTTON_STYLES = [
  { value: '' as const, color: 'bg-neutral-400', preview: 'bg-neutral-200/60 text-neutral-600', label: 'Обычная' },
  { value: 'primary' as const, color: 'bg-[#0088FF]', preview: 'bg-[#0088FF] text-white', label: 'Синяя' },
  { value: 'success' as const, color: 'bg-[#34C759]', preview: 'bg-[#34C759] text-white', label: 'Зелёная' },
  { value: 'danger' as const, color: 'bg-[#FF383C]', preview: 'bg-[#FF383C] text-white', label: 'Красная' },
]

type ButtonStyle = '' | 'primary' | 'success' | 'danger'

export function ButtonStylePicker({ value, onChange }: { value: string; onChange: (style: ButtonStyle) => void }) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const currentIdx = BUTTON_STYLES.findIndex((s) => s.value === value)
  const current = BUTTON_STYLES[Math.max(currentIdx, 0)]

  useEffect(() => {
    if (!open) return
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('mousedown', handleClick)
    document.addEventListener('keydown', handleKey)
    return () => {
      document.removeEventListener('mousedown', handleClick)
      document.removeEventListener('keydown', handleKey)
    }
  }, [open])

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          'inline-flex items-center gap-2 rounded-lg border border-surface-border bg-white px-3 py-1.5 text-sm transition-colors',
          'hover:bg-neutral-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/40',
        )}
      >
        <span className={cn('h-3 w-3 shrink-0 rounded-full', current.color)} />
        <span className="text-neutral-700">{current.label}</span>
        <ChevronDown className={cn('h-3.5 w-3.5 text-neutral-400 transition-transform', open && 'rotate-180')} />
      </button>

      {open && (
        <div className="absolute top-full left-0 mt-1 z-50 w-48 rounded-lg border border-surface-border bg-white p-1.5 shadow-lg">
          {BUTTON_STYLES.map((style) => (
            <button
              key={style.value || 'default'}
              type="button"
              onClick={() => { onChange(style.value); setOpen(false) }}
              className={cn(
                'flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors',
                style.value === value
                  ? 'bg-neutral-100 text-neutral-900'
                  : 'text-neutral-600 hover:bg-neutral-50',
              )}
            >
              <span className={cn('h-3 w-3 shrink-0 rounded-full', style.color)} />
              <span className="flex-1 text-left">{style.label}</span>
              <span className={cn('rounded-md px-2 py-0.5 text-[11px] font-medium leading-tight', style.preview)}>
                Кнопка
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
