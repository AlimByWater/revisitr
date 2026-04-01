import { useState, useRef, useEffect } from 'react'
import { ChevronLeft, ChevronRight, Calendar } from 'lucide-react'
import { cn } from '@/lib/utils'

const MONTHS_RU = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
]
const DAYS_RU = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

function sameDay(a: Date, b: Date) {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

function fmtDate(d: Date): string {
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

function toYMD(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

interface DatePickerProps {
  /** Value in YYYY-MM-DD format */
  value: string
  onChange: (value: string) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  /** Dropdown alignment: left (default) or right */
  align?: 'left' | 'right'
}

export function DatePicker({
  value,
  onChange,
  placeholder = 'Выберите дату',
  disabled = false,
  className,
  align = 'left',
}: DatePickerProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const parsed = value ? new Date(value + 'T00:00:00') : null
  const [viewMonth, setViewMonth] = useState(parsed?.getMonth() ?? new Date().getMonth())
  const [viewYear, setViewYear] = useState(parsed?.getFullYear() ?? new Date().getFullYear())

  useEffect(() => {
    function handler(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  useEffect(() => {
    if (parsed) {
      setViewMonth(parsed.getMonth())
      setViewYear(parsed.getFullYear())
    }
  }, [value])

  function handleDayClick(d: Date) {
    onChange(toYMD(d))
    setOpen(false)
  }

  function prevMonth() {
    if (viewMonth === 0) { setViewMonth(11); setViewYear(viewYear - 1) }
    else setViewMonth(viewMonth - 1)
  }

  function nextMonth() {
    if (viewMonth === 11) { setViewMonth(0); setViewYear(viewYear + 1) }
    else setViewMonth(viewMonth + 1)
  }

  const firstDay = new Date(viewYear, viewMonth, 1)
  let startDow = firstDay.getDay() - 1
  if (startDow < 0) startDow = 6
  const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate()

  const cells: (Date | null)[] = []
  for (let i = 0; i < startDow; i++) cells.push(null)
  for (let d = 1; d <= daysInMonth; d++) cells.push(new Date(viewYear, viewMonth, d))
  while (cells.length % 7 !== 0) cells.push(null)

  const today = startOfDay(new Date())

  return (
    <div ref={ref} className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => !disabled && setOpen(!open)}
        disabled={disabled}
        className={cn(
          'w-full flex items-center gap-2 border border-neutral-200 rounded px-4 py-2.5 text-sm bg-white cursor-pointer hover:bg-neutral-50 transition-colors text-left',
          disabled && 'cursor-not-allowed bg-neutral-100 border-neutral-300 text-neutral-500 hover:bg-neutral-100',
          !value && 'text-neutral-400',
        )}
      >
        <Calendar className="w-4 h-4 text-neutral-400 shrink-0" />
        <span className="flex-1">{parsed ? fmtDate(parsed) : placeholder}</span>
      </button>

      <div className={cn(
        `absolute top-full mt-1 z-50 w-[260px] ${align === 'right' ? 'right-0' : 'left-0'}`,
        'transition-all duration-200 origin-top',
        open
          ? 'opacity-100 scale-y-100 pointer-events-auto'
          : 'opacity-0 scale-y-95 pointer-events-none',
      )}>
        <div className="bg-white border border-neutral-200 rounded p-4 select-none shadow-lg">
          {/* Month nav */}
          <div className="flex items-center justify-between mb-3">
            <button type="button" onClick={prevMonth} className="w-7 h-7 rounded flex items-center justify-center text-neutral-500 hover:text-neutral-900 hover:bg-neutral-100 transition-colors">
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-sm font-semibold text-neutral-900">
              {MONTHS_RU[viewMonth]} {viewYear}
            </span>
            <button type="button" onClick={nextMonth} className="w-7 h-7 rounded flex items-center justify-center text-neutral-500 hover:text-neutral-900 hover:bg-neutral-100 transition-colors">
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>

          {/* Day headers */}
          <div className="grid grid-cols-7 mb-1">
            {DAYS_RU.map((d) => (
              <div key={d} className="text-center text-[11px] font-medium text-neutral-400 py-1">{d}</div>
            ))}
          </div>

          {/* Days */}
          <div className="grid grid-cols-7">
            {cells.map((date, i) => {
              if (!date) return <div key={`empty-${i}`} />

              const isSelected = parsed && sameDay(date, parsed)
              const isToday = sameDay(date, today)

              return (
                <button
                  key={date.toISOString()}
                  type="button"
                  onClick={() => handleDayClick(date)}
                  className="h-9 text-sm transition-all duration-100 flex items-center justify-center p-0.5"
                >
                  <span className={cn(
                    'w-full h-full flex items-center justify-center rounded transition-all duration-100',
                    isSelected
                      ? 'bg-neutral-900 text-white font-semibold'
                      : isToday
                        ? 'font-semibold text-neutral-900 hover:bg-neutral-100'
                        : 'text-neutral-700 hover:bg-neutral-100',
                  )}>
                    {date.getDate()}
                  </span>
                </button>
              )
            })}
          </div>

          {/* Clear */}
          {value && (
            <button
              type="button"
              onClick={() => { onChange(''); setOpen(false) }}
              className="w-full text-xs text-neutral-400 hover:text-neutral-600 mt-2 py-1 transition-colors"
            >
              Сбросить
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
