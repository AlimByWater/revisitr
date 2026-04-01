import { useState, useRef, useEffect } from 'react'
import { ChevronDown, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface SelectOption {
  value: string
  label: string
}

export interface SelectGroup {
  options: SelectOption[]
}

interface CustomSelectProps {
  value: string
  onChange: (value: string) => void
  options: SelectOption[]
  /** Optional grouped options (overrides `options` when provided) */
  groups?: SelectGroup[]
  placeholder?: string
  disabled?: boolean
  className?: string
  width?: string
  /** Use lighter border (border-neutral-200) for selects inside bordered containers */
  light?: boolean
}

export function CustomSelect({
  value,
  onChange,
  options,
  groups,
  placeholder = 'Выберите...',
  disabled = false,
  className,
  width,
  light = false,
}: CustomSelectProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const allOptions = groups ? groups.flatMap((g) => g.options) : options
  const selectedLabel = allOptions.find((o) => o.value === value)?.label ?? placeholder

  useEffect(() => {
    function handler(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const renderOptions = (opts: SelectOption[]) =>
    opts.map((o) => (
      <button
        key={o.value}
        type="button"
        onClick={() => { onChange(o.value); setOpen(false) }}
        className={cn(
          'w-full flex items-center justify-between px-4 py-1.5 text-sm text-left transition-colors',
          o.value === value
            ? 'font-semibold text-neutral-900 bg-neutral-50'
            : 'text-neutral-600 hover:bg-neutral-50 hover:text-neutral-900',
        )}
      >
        <span>{o.label}</span>
        {o.value === value && <Check className="w-3.5 h-3.5 text-neutral-900" />}
      </button>
    ))

  return (
    <div ref={ref} className={cn('relative', className)} style={width ? { width } : undefined}>
      <button
        type="button"
        onClick={() => !disabled && setOpen(!open)}
        disabled={disabled}
        className={cn(
          'w-full flex items-center gap-3 border rounded px-4 py-2 text-sm font-medium text-neutral-900 bg-white cursor-pointer hover:bg-neutral-50 transition-colors',
          light ? 'border-neutral-200' : 'border-neutral-900',
          disabled && 'cursor-not-allowed bg-neutral-100 border-neutral-300 text-neutral-500 hover:bg-neutral-100',
        )}
      >
        <span className="flex-1 text-left truncate">{selectedLabel}</span>
        <ChevronDown className={cn('w-4 h-4 text-neutral-500 shrink-0 transition-transform duration-200', open && 'rotate-180')} />
      </button>

      <div className={cn(
        'absolute top-full left-0 mt-1 w-full bg-white border rounded py-1 z-50',
        light ? 'border-neutral-200' : 'border-neutral-900',
        'transition-all duration-150 origin-top',
        open
          ? 'opacity-100 scale-y-100 pointer-events-auto'
          : 'opacity-0 scale-y-95 pointer-events-none',
      )}>
        {groups
          ? groups.map((group, gi) => (
              <div key={gi}>
                {gi > 0 && <div className="my-1 mx-3 border-t border-neutral-200" />}
                {renderOptions(group.options)}
              </div>
            ))
          : renderOptions(allOptions)
        }
      </div>
    </div>
  )
}
