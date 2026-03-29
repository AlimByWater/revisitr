import { useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import {
  useCreateCampaignMutation,
  useCreateScenarioMutation,
  usePreviewAudienceMutation,
} from '@/features/campaigns/queries'
import { ArrowLeft, Send, Zap } from 'lucide-react'
import type { SegmentFilter } from '@/features/segments/types'
import type { AudienceFilter } from '@/features/campaigns/types'
import { ClientFilterBuilder } from '@/components/filters/ClientFilterBuilder'

type Format = 'manual' | 'auto'

type TriggerType =
  | 'inactive_days'
  | 'visit_count'
  | 'bonus_threshold'
  | 'level_up'
  | 'birthday'
  | 'date'
  | 'registration'
  | 'level_change'

const TRIGGER_OPTIONS: { value: TriggerType; label: string }[] = [
  { value: 'inactive_days', label: 'Не был N дней' },
  { value: 'visit_count', label: 'N-й визит' },
  { value: 'bonus_threshold', label: 'Порог бонусов' },
  { value: 'level_up', label: 'Новый уровень' },
  { value: 'birthday', label: 'День рождения' },
  { value: 'date', label: 'Конкретная дата' },
  { value: 'registration', label: 'Регистрация' },
  { value: 'level_change', label: 'Смена уровня' },
]

const inputClass = cn(
  'w-full max-w-sm px-3 py-2.5 rounded-lg border border-neutral-200',
  'text-sm text-neutral-900 placeholder:text-neutral-400 bg-white',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
)

const labelClass = 'block text-sm font-medium text-neutral-700 mb-1.5'

const blockHeaderClass =
  'font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-4'

export default function CreateCampaignPage() {
  const navigate = useNavigate()
  const { data: bots } = useBotsQuery()
  const createCampaignMutation = useCreateCampaignMutation()
  const createScenarioMutation = useCreateScenarioMutation()
  const previewMutation = usePreviewAudienceMutation()

  const [format, setFormat] = useState<Format>('manual')
  const [name, setName] = useState('')
  const [botId, setBotId] = useState<number | ''>('')
  const [message, setMessage] = useState('')
  const [audienceFilter, setAudienceFilter] = useState<SegmentFilter>({})

  // Auto-scenario state
  const [triggerType, setTriggerType] = useState<TriggerType>('inactive_days')
  const [triggerDays, setTriggerDays] = useState<number | ''>('')
  const [triggerCount, setTriggerCount] = useState<number | ''>('')
  const [triggerThreshold, setTriggerThreshold] = useState<number | ''>('')
  const [triggerDate, setTriggerDate] = useState('')
  const [daysBefore, setDaysBefore] = useState<number | ''>('')
  const [daysAfter, setDaysAfter] = useState<number | ''>('')

  const isManual = format === 'manual'
  const activeMutation = isManual ? createCampaignMutation : createScenarioMutation

  const baseValid = name.trim() !== '' && botId !== '' && message.trim() !== ''
  const isValid = isManual ? baseValid : baseValid && isTriggerConfigValid()

  function isTriggerConfigValid(): boolean {
    switch (triggerType) {
      case 'inactive_days':
        return triggerDays !== '' && triggerDays > 0
      case 'visit_count':
        return triggerCount !== '' && triggerCount > 0
      case 'bonus_threshold':
        return triggerThreshold !== '' && triggerThreshold > 0
      case 'date':
        return triggerDate !== ''
      default:
        return true
    }
  }

  function handleBotChange(value: string) {
    const id = value ? Number(value) : ''
    setBotId(id)
    if (id) {
      setAudienceFilter((prev) => ({ ...prev, bot_id: id as number }))
    } else {
      setAudienceFilter((prev) => {
        const { bot_id: _, ...rest } = prev
        return rest
      })
    }
  }

  /** Convert SegmentFilter to AudienceFilter for the create campaign request */
  function toAudienceFilter(filter: SegmentFilter): AudienceFilter {
    const af: AudienceFilter = {}
    if (filter.bot_id) af.bot_id = filter.bot_id
    if (filter.tags && filter.tags.length > 0) af.tags = filter.tags
    if (filter.level_id) af.level_id = filter.level_id
    return af
  }

  function handlePreview() {
    previewMutation.mutate(audienceFilter)
  }

  function buildTriggerConfig() {
    switch (triggerType) {
      case 'inactive_days':
        return { days: triggerDays as number }
      case 'visit_count':
        return { count: triggerCount as number }
      case 'bonus_threshold':
        return { threshold: triggerThreshold as number }
      case 'date': {
        const d = new Date(triggerDate)
        return { month: d.getMonth() + 1, day: d.getDate() }
      }
      default:
        return {}
    }
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!isValid) return

    if (isManual) {
      createCampaignMutation.mutate(
        {
          bot_id: botId as number,
          name: name.trim(),
          message: message.trim(),
          audience_filter: toAudienceFilter({
            ...audienceFilter,
            bot_id: botId as number,
          }),
        },
        { onSuccess: () => navigate('/dashboard/campaigns') },
      )
    } else {
      const timing: { days_before?: number; days_after?: number } = {}
      if (daysBefore !== '') timing.days_before = daysBefore
      if (daysAfter !== '') timing.days_after = daysAfter

      createScenarioMutation.mutate(
        {
          bot_id: botId as number,
          name: name.trim(),
          // "date" in UI maps to "holiday" for API backward compatibility
          trigger_type: triggerType === 'date' ? 'holiday' : triggerType,
          trigger_config: buildTriggerConfig(),
          message: message.trim(),
          timing: Object.keys(timing).length > 0 ? timing : undefined,
        },
        { onSuccess: () => navigate('/dashboard/campaigns') },
      )
    }
  }

  return (
    <div className="max-w-2xl">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate('/dashboard/campaigns')}
          type="button"
          className="p-2 rounded-lg hover:bg-neutral-100 transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-neutral-500" />
        </button>
        <h1 className="font-serif text-2xl font-bold text-neutral-900 tracking-tight">
          {isManual ? 'Создать рассылку' : 'Создать авто-сценарий'}
        </h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* ── Block 1: Format Selection ─────────────────────────────── */}
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 space-y-5">
          <div>
            <p className={blockHeaderClass}>Формат рассылки</p>
            <p className="text-sm text-neutral-500 mb-3">
              Выберите тип рассылки
            </p>

            <select
              value={format}
              onChange={(e) => setFormat(e.target.value as Format)}
              className={inputClass}
            >
              <option value="manual">Ручная рассылка</option>
              <option value="auto">Авто-сценарий</option>
            </select>
          </div>

          {/* Bot selector */}
          <div>
            <label htmlFor="bot" className={labelClass}>
              Бот
            </label>
            <select
              id="bot"
              value={botId}
              onChange={(e) => handleBotChange(e.target.value)}
              className={inputClass}
            >
              <option value="">Выберите бота</option>
              {bots?.map((bot) => (
                <option key={bot.id} value={bot.id}>
                  {bot.name} (@{bot.username})
                </option>
              ))}
            </select>
          </div>

          {/* Campaign name */}
          <div>
            <label htmlFor="name" className={labelClass}>
              {isManual ? 'Название рассылки' : 'Название сценария'}
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={
                isManual
                  ? 'Например: Акция выходного дня'
                  : 'Например: Возврат неактивных клиентов'
              }
              className={inputClass}
            />
          </div>
        </div>

        {/* ── Block 2: Auto-scenario Settings ──────────────────────── */}
        {!isManual && (
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 space-y-5">
            <p className={blockHeaderClass}>Настройки сценария</p>

            {/* Trigger type */}
            <div>
              <label htmlFor="trigger-type" className={labelClass}>
                Триггер
              </label>
              <select
                id="trigger-type"
                value={triggerType}
                onChange={(e) => setTriggerType(e.target.value as TriggerType)}
                className={inputClass}
              >
                {TRIGGER_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Trigger config fields (conditional) */}
            {triggerType === 'inactive_days' && (
              <div>
                <label htmlFor="trigger-days" className={labelClass}>
                  Количество дней
                </label>
                <input
                  id="trigger-days"
                  type="number"
                  min={1}
                  value={triggerDays}
                  onChange={(e) =>
                    setTriggerDays(e.target.value ? Number(e.target.value) : '')
                  }
                  placeholder="7"
                  className={inputClass}
                />
              </div>
            )}

            {triggerType === 'visit_count' && (
              <div>
                <label htmlFor="trigger-count" className={labelClass}>
                  Номер визита
                </label>
                <input
                  id="trigger-count"
                  type="number"
                  min={1}
                  value={triggerCount}
                  onChange={(e) =>
                    setTriggerCount(
                      e.target.value ? Number(e.target.value) : '',
                    )
                  }
                  placeholder="5"
                  className={inputClass}
                />
              </div>
            )}

            {triggerType === 'bonus_threshold' && (
              <div>
                <label htmlFor="trigger-threshold" className={labelClass}>
                  Порог бонусов
                </label>
                <input
                  id="trigger-threshold"
                  type="number"
                  min={1}
                  value={triggerThreshold}
                  onChange={(e) =>
                    setTriggerThreshold(
                      e.target.value ? Number(e.target.value) : '',
                    )
                  }
                  placeholder="1000"
                  className={inputClass}
                />
              </div>
            )}

            {triggerType === 'date' && (
              <div>
                <label htmlFor="trigger-date" className={labelClass}>
                  Выберите дату
                </label>
                <input
                  id="trigger-date"
                  type="date"
                  value={triggerDate}
                  onChange={(e) => setTriggerDate(e.target.value)}
                  className={inputClass}
                />
              </div>
            )}

            {/* Timing section */}
            <div>
              <p className="text-sm font-medium text-neutral-700 mb-3">
                Время отправки (опционально)
              </p>
              <div className="flex gap-4">
                <div className="flex-1 max-w-[180px]">
                  <label
                    htmlFor="days-before"
                    className="block text-xs text-neutral-500 mb-1"
                  >
                    Дней до
                  </label>
                  <input
                    id="days-before"
                    type="number"
                    min={0}
                    value={daysBefore}
                    onChange={(e) =>
                      setDaysBefore(
                        e.target.value ? Number(e.target.value) : '',
                      )
                    }
                    placeholder="0"
                    className={inputClass}
                  />
                </div>
                <div className="flex-1 max-w-[180px]">
                  <label
                    htmlFor="days-after"
                    className="block text-xs text-neutral-500 mb-1"
                  >
                    Дней после
                  </label>
                  <input
                    id="days-after"
                    type="number"
                    min={0}
                    value={daysAfter}
                    onChange={(e) =>
                      setDaysAfter(
                        e.target.value ? Number(e.target.value) : '',
                      )
                    }
                    placeholder="0"
                    className={inputClass}
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* ── Block 3: Message Editor ──────────────────────────────── */}
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 space-y-5">
          <p className={blockHeaderClass}>Сообщение</p>

          {/* Message textarea */}
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label htmlFor="message" className="block text-sm font-medium text-neutral-700">
                Текст сообщения
              </label>
              <span
                className={cn(
                  'text-xs tabular-nums',
                  message.length > 3800 ? 'text-red-500' : 'text-neutral-400',
                )}
              >
                {message.length} / 4096
              </span>
            </div>
            <textarea
              id="message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              rows={6}
              maxLength={4096}
              placeholder="Текст сообщения. Поддерживается форматирование Telegram: *жирный*, _курсив_"
              className={cn(
                'w-full px-3 py-2.5 rounded-lg border border-neutral-200',
                'text-sm text-neutral-900 placeholder:text-neutral-400 resize-none',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              )}
            />
          </div>

          {/* Audience filter */}
          <div>
            <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-2">
              Аудитория
            </p>
            <ClientFilterBuilder
              value={audienceFilter}
              onChange={(f) =>
                setAudienceFilter(
                  botId ? { ...f, bot_id: botId as number } : f,
                )
              }
              hiddenFields={['bot_id']}
              previewCount={
                previewMutation.isSuccess ? previewMutation.data : null
              }
              onPreview={botId ? handlePreview : undefined}
              isPreviewing={previewMutation.isPending}
            />
          </div>
        </div>

        {/* ── Bottom Actions ───────────────────────────────────────── */}
        <div className="flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate('/dashboard/campaigns')}
            className={cn(
              'px-4 py-2.5 rounded-lg text-sm font-medium',
              'border border-neutral-200 text-neutral-700',
              'hover:bg-neutral-50 transition-colors',
            )}
          >
            Отмена
          </button>
          <button
            type="submit"
            disabled={!isValid || activeMutation.isPending}
            className={cn(
              'flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium',
              'bg-accent text-white',
              'hover:bg-accent/90 active:bg-accent/80',
              'transition-colors',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'focus:outline-none focus:ring-2 focus:ring-accent/20',
            )}
          >
            {isManual ? (
              <Send className="w-4 h-4" />
            ) : (
              <Zap className="w-4 h-4" />
            )}
            {activeMutation.isPending
              ? 'Создание...'
              : isManual
                ? 'Создать рассылку'
                : 'Создать авто-сценарий'}
          </button>
        </div>

        {activeMutation.isError && (
          <p className="text-sm text-red-600 mt-3 text-right">
            {isManual
              ? 'Не удалось создать рассылку. Попробуйте ещё раз.'
              : 'Не удалось создать сценарий. Попробуйте ещё раз.'}
          </p>
        )}
      </form>
    </div>
  )
}
