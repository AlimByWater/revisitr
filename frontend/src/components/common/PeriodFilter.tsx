import { useState, useRef, useEffect, useMemo } from 'react'
import { ChevronDown, ChevronLeft, ChevronRight, Check, Calendar } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PeriodFilterProps {
  period: string
  from?: string
  to?: string
  onPeriodChange: (period: string) => void
  onRangeChange: (from: string, to: string) => void
}

const PRESET_GROUPS: { value: string; label: string }[][] = [
  [
    { value: 'today', label: 'Сегодня' },
    { value: 'yesterday', label: 'Вчера' },
    { value: '7d', label: 'Последние 7 дней' },
    { value: '30d', label: 'Последние 30 дней' },
    { value: '365d', label: 'Последний год' },
  ],
  [
    { value: 'this_week', label: 'Эта неделя' },
    { value: 'this_month', label: 'Этот месяц' },
    { value: 'this_year', label: 'Этот год' },
  ],
  [
    { value: 'custom', label: 'Произвольный' },
  ],
]

const ALL_PRESETS = PRESET_GROUPS.flat()

function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

function getPresetRange(preset: string): [Date, Date] {
  const now = new Date()
  const today = startOfDay(now)

  switch (preset) {
    case 'today':
      return [new Date(today), new Date(today)]
    case 'yesterday': {
      const y = new Date(today)
      y.setDate(y.getDate() - 1)
      return [y, y]
    }
    case 'this_week': {
      const day = today.getDay()
      const diff = day === 0 ? 6 : day - 1
      const start = new Date(today)
      start.setDate(today.getDate() - diff)
      return [start, new Date(today)]
    }
    case 'this_month':
      return [new Date(today.getFullYear(), today.getMonth(), 1), new Date(today)]
    case 'this_year':
      return [new Date(today.getFullYear(), 0, 1), new Date(today)]
    case '7d': {
      const s = new Date(today)
      s.setDate(today.getDate() - 6)
      return [s, new Date(today)]
    }
    case '30d': {
      const s = new Date(today)
      s.setDate(today.getDate() - 29)
      return [s, new Date(today)]
    }
    case '365d': {
      const s = new Date(today)
      s.setDate(today.getDate() - 364)
      return [s, new Date(today)]
    }
    default:
      return [new Date(today), new Date(today)]
  }
}

function fmt(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function fmtDisplay(d: Date): string {
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

function fmtMobile(d: Date): string {
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' })
}

export function PeriodFilter({ period, from, to, onPeriodChange, onRangeChange }: PeriodFilterProps) {
  const [presetOpen, setPresetOpen] = useState(false)
  const [calendarOpen, setCalendarOpen] = useState(false)
  const presetRef = useRef<HTMLDivElement>(null)
  const calendarRef = useRef<HTMLDivElement>(null)

  const selectedLabel = ALL_PRESETS.find((p) => p.value === period)?.label ?? 'Период'

  const displayRange = useMemo(() => {
    if (from && to) return { from: new Date(from + 'T00:00:00'), to: new Date(to + 'T00:00:00') }
    if (period !== 'custom') {
      const [f, t] = getPresetRange(period)
      return { from: f, to: t }
    }
    const [f, t] = getPresetRange('30d')
    return { from: f, to: t }
  }, [period, from, to])

  useEffect(() => {
    function handler(e: MouseEvent) {
      if (presetRef.current && !presetRef.current.contains(e.target as Node)) setPresetOpen(false)
      if (calendarRef.current && !calendarRef.current.contains(e.target as Node)) setCalendarOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  function handlePresetSelect(value: string) {
    onPeriodChange(value)
    if (value !== 'custom') {
      const [f, t] = getPresetRange(value)
      onRangeChange(fmt(f), fmt(t))
    }
    setPresetOpen(false)
  }

  function handleRangeSelect(f: Date, t: Date) {
    onRangeChange(fmt(f), fmt(t))
    for (const preset of ALL_PRESETS) {
      if (preset.value === 'custom') continue
      const [pf, pt] = getPresetRange(preset.value)
      if (fmt(pf) === fmt(f) && fmt(pt) === fmt(t)) {
        onPeriodChange(preset.value)
        setCalendarOpen(false)
        return
      }
    }
    onPeriodChange('custom')
    setCalendarOpen(false)
  }

  return (
    <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full sm:w-auto">
      {/* Period preset dropdown */}
      <div ref={presetRef} className="relative">
        <button
          type="button"
          onClick={() => { setPresetOpen(!presetOpen); setCalendarOpen(false) }}
          className="flex items-center gap-3 border border-neutral-900 rounded px-4 py-2 text-sm font-medium text-neutral-900 bg-white cursor-pointer hover:bg-neutral-50 transition-colors w-full sm:w-[200px]"
        >
          <span className="flex-1 text-left">{selectedLabel}</span>
          <ChevronDown className={cn('w-4 h-4 text-neutral-500 transition-transform duration-200', presetOpen && 'rotate-180')} />
        </button>

        <div className={cn(
          'absolute top-full left-0 mt-1 w-full sm:w-[200px] bg-white border border-neutral-900 rounded py-1 z-50',
          'transition-all duration-150 origin-top',
          presetOpen
            ? 'opacity-100 scale-y-100 pointer-events-auto'
            : 'opacity-0 scale-y-95 pointer-events-none',
        )}>
          {PRESET_GROUPS.map((group, gi) => (
            <div key={gi}>
              {gi > 0 && <div className="my-1 mx-3 border-t border-neutral-200" />}
              {group.map((p) => (
                <button
                  key={p.value}
                  type="button"
                  onClick={() => handlePresetSelect(p.value)}
                  className={cn(
                    'w-full flex items-center justify-between px-4 py-1.5 text-sm text-left transition-colors',
                    p.value === period
                      ? 'font-semibold text-neutral-900 bg-neutral-50'
                      : 'text-neutral-600 hover:bg-neutral-50 hover:text-neutral-900',
                  )}
                >
                  <span>{p.label}</span>
                  {p.value === period && <Check className="w-3.5 h-3.5 text-neutral-900" />}
                </button>
              ))}
            </div>
          ))}
        </div>
      </div>

      {/* Date range display / calendar trigger */}
      <div ref={calendarRef} className="relative">
        <button
          type="button"
          onClick={() => { setCalendarOpen(!calendarOpen); setPresetOpen(false) }}
          className="flex items-center gap-2 border border-neutral-900 rounded px-4 py-2 text-sm text-neutral-900 bg-white cursor-pointer hover:bg-neutral-50 transition-colors w-full sm:w-auto"
        >
          <Calendar className="w-4 h-4 text-neutral-500 shrink-0" />
          <span className="hidden sm:inline">
            {fmtDisplay(displayRange.from)}
            <span className="text-neutral-400 mx-1.5">&mdash;</span>
            {fmtDisplay(displayRange.to)}
          </span>
          <span className="sm:hidden">
            {fmtMobile(displayRange.from)}
            <span className="text-neutral-400 mx-1">&mdash;</span>
            {fmtMobile(displayRange.to)}
          </span>
        </button>

        <div className={cn(
          'absolute top-full right-0 mt-1 z-50 w-full',
          'transition-all duration-200 origin-top',
          calendarOpen
            ? 'opacity-100 scale-y-100 pointer-events-auto'
            : 'opacity-0 scale-y-95 pointer-events-none',
        )}>
          <RangeCalendar
            from={displayRange.from}
            to={displayRange.to}
            onSelect={handleRangeSelect}
          />
        </div>
      </div>
    </div>
  )
}

// ── Range Calendar ──

const MONTHS_RU = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
]
const DAYS_RU = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

function sameDay(a: Date, b: Date) {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

function inRange(d: Date, from: Date, to: Date) {
  const t = d.getTime()
  return t >= from.getTime() && t <= to.getTime()
}

function RangeCalendar({
  from,
  to,
  onSelect,
}: {
  from: Date
  to: Date
  onSelect: (from: Date, to: Date) => void
}) {
  const [viewMonth, setViewMonth] = useState(to.getMonth())
  const [viewYear, setViewYear] = useState(to.getFullYear())
  const [rangeStart, setRangeStart] = useState<Date | null>(null)
  const [hoverDate, setHoverDate] = useState<Date | null>(null)

  // Sync calendar view when props change (preset selected)
  useEffect(() => {
    setViewMonth(to.getMonth())
    setViewYear(to.getFullYear())
    setRangeStart(null)
    setHoverDate(null)
  }, [from.getTime(), to.getTime()])

  const effectiveFrom = rangeStart ?? from
  const effectiveTo = rangeStart
    ? hoverDate
      ? hoverDate >= rangeStart ? hoverDate : rangeStart
      : rangeStart
    : to
  const effectiveFromNorm = effectiveFrom <= effectiveTo ? effectiveFrom : effectiveTo
  const effectiveToNorm = effectiveFrom <= effectiveTo ? effectiveTo : effectiveFrom

  function handleDayClick(d: Date) {
    if (!rangeStart) {
      setRangeStart(d)
    } else {
      const f = d >= rangeStart ? rangeStart : d
      const t = d >= rangeStart ? d : rangeStart
      setRangeStart(null)
      setHoverDate(null)
      onSelect(f, t)
    }
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
    <div className="bg-white border border-neutral-900 rounded p-4 w-full select-none">
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

          const isFuture = date.getTime() > today.getTime()
          const isToday = sameDay(date, today)
          const isStart = sameDay(date, effectiveFromNorm)
          const isEnd = sameDay(date, effectiveToNorm)
          const isInRange = !isFuture && inRange(date, effectiveFromNorm, effectiveToNorm)
          const isEdge = isStart || isEnd

          return (
            <button
              key={date.toISOString()}
              type="button"
              disabled={isFuture}
              onClick={() => !isFuture && handleDayClick(date)}
              onMouseEnter={() => rangeStart && !isFuture && setHoverDate(date)}
              className={cn(
                'h-9 text-sm transition-all duration-100 flex items-center justify-center p-0.5',
                isFuture && 'cursor-not-allowed',
                isInRange && !isEdge && 'bg-neutral-100',
                isStart && !isEnd && 'bg-neutral-100 rounded-l',
                isEnd && !isStart && 'bg-neutral-100 rounded-r',
                isStart && isEnd && '',
              )}
            >
              <span className={cn(
                'w-full h-full flex items-center justify-center rounded transition-all duration-100',
                isFuture
                  ? 'text-neutral-300'
                  : isEdge
                    ? 'bg-neutral-900 text-white font-semibold'
                    : isToday
                      ? 'font-semibold text-neutral-900'
                      : 'text-neutral-700 hover:bg-neutral-100',
              )}>
                {date.getDate()}
              </span>
            </button>
          )
        })}
      </div>

      {rangeStart && (
        <p className="text-[11px] text-neutral-400 text-center mt-2 animate-in">
          Выберите конечную дату
        </p>
      )}
    </div>
  )
}
