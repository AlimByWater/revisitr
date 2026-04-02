import { Link, useParams } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { usePOSDetailQuery, useUpdatePOSMutation } from '@/features/pos/queries'
import type { Schedule, DaySchedule } from '@/features/pos/types'

const DAYS: { key: string; label: string }[] = [
  { key: 'mon', label: 'Пн' },
  { key: 'tue', label: 'Вт' },
  { key: 'wed', label: 'Ср' },
  { key: 'thu', label: 'Чт' },
  { key: 'fri', label: 'Пт' },
  { key: 'sat', label: 'Сб' },
  { key: 'sun', label: 'Вс' },
]

const DEFAULT_DAY: DaySchedule = { open: '09:00', close: '22:00' }

export default function POSDetailPage() {
  const { posId } = useParams<{ posId: string }>()
  const id = Number(posId)

  const { data: location, isLoading } = usePOSDetailQuery(id)
  const updateMutation = useUpdatePOSMutation()

  const [name, setName] = useState('')
  const [address, setAddress] = useState('')
  const [phone, setPhone] = useState('')
  const [schedule, setSchedule] = useState<Schedule>({})

  useEffect(() => {
    if (location) {
      setName(location.name)
      setAddress(location.address)
      setPhone(location.phone)
      setSchedule(location.schedule ?? {})
    }
  }, [location])

  function updateDay(dayKey: string, field: keyof DaySchedule, value: string | boolean) {
    setSchedule((prev) => ({
      ...prev,
      [dayKey]: {
        ...(prev[dayKey] ?? DEFAULT_DAY),
        [field]: value,
      },
    }))
  }

  async function handleSave() {
    await updateMutation.mutateAsync({
      id,
      data: { name, address, phone, schedule },
    })
  }

  if (isLoading) {
    return <div className="text-sm text-neutral-500">Загрузка...</div>
  }

  if (!location) {
    return <div className="text-sm text-neutral-500">Точка не найдена</div>
  }

  return (
    <div>
      <Link
        to="/dashboard/pos"
        className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-900 transition-colors mb-4"
      >
        <ArrowLeft className="w-4 h-4" />
        Назад к точкам продаж
      </Link>

      <h1 className="font-serif font-serif text-3xl font-bold text-neutral-900 tracking-tight mb-8">{location.name}</h1>

      {/* General info */}
      <div className="bg-white rounded border border-neutral-900 p-6 mb-6">
        <h2 className="text-base font-semibold mb-4">
          <span className="block font-mono text-[10px] uppercase tracking-widest text-neutral-400 font-normal mb-0.5">Заведение</span>
          Основная информация
        </h2>
        <div className="space-y-4">
          <div>
            <label
              htmlFor="pos-name"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Название
            </label>
            <input
              id="pos-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={cn(
                'w-full max-w-sm px-4 py-2.5 rounded border border-neutral-200',
                'text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>
          <div>
            <label
              htmlFor="pos-address"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Адрес
            </label>
            <input
              id="pos-address"
              type="text"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="Улица, дом"
              className={cn(
                'w-full px-4 py-2.5 rounded border border-neutral-200',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>
          <div>
            <label
              htmlFor="pos-phone"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Телефон
            </label>
            <input
              id="pos-phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="+7 (999) 123-45-67"
              className={cn(
                'w-full max-w-[220px] px-4 py-2.5 rounded border border-neutral-200',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>
        </div>
      </div>

      {/* Schedule */}
      <div className="bg-white rounded border border-neutral-900 p-6">
        <h2 className="text-base font-semibold mb-4">
          <span className="block font-mono text-[10px] uppercase tracking-widest text-neutral-400 font-normal mb-0.5">Режим работы</span>
          Расписание
        </h2>
        <div className="space-y-3">
          {DAYS.map(({ key, label }) => {
            const day = schedule[key] ?? DEFAULT_DAY
            const isClosed = day.closed ?? false
            return (
              <div key={key} className="flex items-center gap-4">
                <span className="w-8 text-sm font-medium text-neutral-700">
                  {label}
                </span>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={!isClosed}
                    onChange={(e) => updateDay(key, 'closed', !e.target.checked)}
                    className="sr-only peer"
                  />
                  <div
                    className={cn(
                      'w-9 h-5 rounded-full transition-colors',
                      'peer-checked:bg-green-400 bg-neutral-300',
                      'relative after:content-[""] after:absolute after:top-0.5 after:start-[2px]',
                      'after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all',
                      'peer-checked:after:translate-x-full',
                    )}
                  />
                  <span className="text-xs text-neutral-500">
                    {isClosed ? 'Выходной' : 'Открыто'}
                  </span>
                </label>
                <input
                  type="time"
                  value={day.open}
                  onChange={(e) => updateDay(key, 'open', e.target.value)}
                  disabled={isClosed}
                  aria-label={`${label} открытие`}
                  className={cn(
                    'px-3 py-1.5 rounded border border-neutral-200 text-sm',
                    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                    'disabled:opacity-40 disabled:cursor-not-allowed',
                  )}
                />
                <span className="text-xs text-neutral-400">до</span>
                <input
                  type="time"
                  value={day.close}
                  onChange={(e) => updateDay(key, 'close', e.target.value)}
                  disabled={isClosed}
                  aria-label={`${label} закрытие`}
                  className={cn(
                    'px-3 py-1.5 rounded border border-neutral-200 text-sm',
                    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                    'disabled:opacity-40 disabled:cursor-not-allowed',
                  )}
                />
              </div>
            )
          })}
        </div>
      </div>

      {/* Save button */}
      <div className="mt-6 flex items-center gap-4">
        <button
          type="button"
          onClick={handleSave}
          disabled={updateMutation.isPending}
          className={cn(
            'px-6 py-2.5 rounded text-sm font-medium',
            'bg-accent text-white hover:bg-accent/90 transition-colors',
            'disabled:opacity-50 disabled:cursor-not-allowed',
          )}
        >
          {updateMutation.isPending ? 'Сохранение...' : 'Сохранить'}
        </button>
        {updateMutation.isSuccess && (
          <span className="text-sm text-green-600">Изменения сохранены</span>
        )}
        {updateMutation.isError && (
          <span className="text-sm text-red-600">Не удалось сохранить. Попробуйте снова.</span>
        )}
      </div>
    </div>
  )
}
