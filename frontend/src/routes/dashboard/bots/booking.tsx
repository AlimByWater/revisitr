import { Suspense, lazy, useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, Check, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { botsApi } from '@/features/bots/api'
import { useBotQuery } from '@/features/bots/queries'
import type { BookingModuleConfig, BookingTimeSlot, FormField, ModuleConfigs } from '@/features/bots/types'
import { normalizeBookingConfig, normalizeModuleConfigs, normalizeTimeInput } from '@/features/bots/settings'
import { usePOSQuery } from '@/features/pos/queries'
import { menusApi } from '@/features/menus/api'
import { campaignsApi } from '@/features/campaigns/api'
import type { MessageContent, MessagePlaceholder } from '@/features/telegram-preview'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'

const MessageContentEditor = lazy(() =>
  import('@/features/telegram-preview').then((module) => ({
    default: module.MessageContentEditor,
  })),
)

const inputClassName = cn(
  'w-full px-4 py-2.5 rounded border border-neutral-200',
  'text-sm placeholder:text-neutral-400 bg-white',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors disabled:opacity-50 disabled:cursor-not-allowed',
)

interface BookingDraft {
  intro_content?: MessageContent
  date_from_days: string
  date_to_days: string
  time_slots: BookingTimeSlot[]
  party_size_options: string[]
  pos_ids: number[]
}

function editorFallback() {
  return (
    <div className="space-y-3 rounded border border-neutral-200 bg-white p-4">
      <div className="h-11 rounded bg-neutral-100 animate-pulse" />
      <div className="h-32 rounded bg-neutral-100 animate-pulse" />
      <div className="h-11 rounded bg-neutral-100 animate-pulse" />
    </div>
  )
}

function buildMessagePlaceholders(fields: FormField[]): MessagePlaceholder[] {
  const seen = new Set<string>()
  const placeholders: MessagePlaceholder[] = []

  for (const field of fields) {
    const name = field.name.trim()
    if (!name) continue
    const tokenName = name === 'birthday' ? 'birth_date' : name
    const token = `{${tokenName}}`
    if (seen.has(token)) continue
    placeholders.push({
      token,
      label: field.label.trim() || name,
      fieldType: field.type,
    })
    seen.add(token)
  }

  return placeholders
}

function draftFromConfig(config: BookingModuleConfig): BookingDraft {
  return {
    intro_content: config.intro_content,
    date_from_days: String(config.date_from_days ?? 0),
    date_to_days: String(config.date_to_days ?? 7),
    time_slots: config.time_slots ?? [],
    party_size_options: config.party_size_options ?? [],
    pos_ids: config.pos_ids ?? [],
  }
}

function numberFromDraft(value: string, fallback: number): number {
  if (value.trim() === '') return fallback
  const parsed = Number(value)
  return Number.isFinite(parsed) ? Math.max(0, parsed) : fallback
}

function buildModuleConfigs(configs: ModuleConfigs | undefined, draft: BookingDraft): ModuleConfigs {
  const normalized = normalizeModuleConfigs(configs)
  return {
    ...normalized,
    booking: {
      intro_content: draft.intro_content,
      date_from_days: numberFromDraft(draft.date_from_days, 0),
      date_to_days: numberFromDraft(draft.date_to_days, 7),
      time_slots: draft.time_slots,
      party_size_options: draft.party_size_options.filter((option) => option.trim()),
      pos_ids: draft.pos_ids,
    },
  }
}

function useSaveAction(updater: () => Promise<void>) {
  const [isSaving, setIsSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)

  useEffect(() => {
    if (!saveSuccess) return
    const t = setTimeout(() => setSaveSuccess(false), 3000)
    return () => clearTimeout(t)
  }, [saveSuccess])

  const save = async () => {
    setIsSaving(true)
    setSaveError(null)
    setSaveSuccess(false)
    try {
      await updater()
      setSaveSuccess(true)
    } catch {
      setSaveError('Не удалось сохранить изменения. Попробуйте ещё раз.')
    } finally {
      setIsSaving(false)
    }
  }

  return { isSaving, saveError, saveSuccess, save }
}

function SaveFooter({
  isSaving,
  saveError,
  saveSuccess,
  onSave,
}: {
  isSaving: boolean
  saveError: string | null
  saveSuccess: boolean
  onSave: () => void
}) {
  return (
    <div className="mt-6 flex flex-col gap-3 border-t border-neutral-200 pt-5 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-h-5 text-sm">
        {saveError && <span className="text-red-600">{saveError}</span>}
        {saveSuccess && <span className="text-green-600">Сохранено</span>}
      </div>
      <button
        type="button"
        onClick={onSave}
        disabled={isSaving}
        className={cn(
          'inline-flex min-h-11 items-center justify-center rounded px-5 text-sm font-medium',
          'bg-accent text-white hover:bg-accent/90 transition-colors',
          'disabled:cursor-not-allowed disabled:opacity-50',
        )}
      >
        {isSaving ? 'Сохранение...' : 'Сохранить изменения'}
      </button>
    </div>
  )
}

export default function BotBookingSettingsPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)
  const { data: bot, isLoading, isError, mutate } = useBotQuery(Number.isNaN(id) ? 0 : id)
  const { data: posLocations = [] } = usePOSQuery()
  const [boundPosIds, setBoundPosIds] = useState<number[]>([])
  const [posBindingsLoaded, setPosBindingsLoaded] = useState(false)
  const [draft, setDraft] = useState<BookingDraft | null>(null)

  useEffect(() => {
    if (!id || Number.isNaN(id) || id <= 0) return
    let mounted = true
    menusApi
      .getBotPOSLocations(id)
      .then((response) => {
        if (!mounted) return
        setBoundPosIds(response.pos_ids ?? [])
        setPosBindingsLoaded(true)
      })
      .catch(() => {
        if (!mounted) return
        setBoundPosIds([])
        setPosBindingsLoaded(true)
      })
    return () => {
      mounted = false
    }
  }, [id])

  useEffect(() => {
    if (!bot || !posBindingsLoaded) return
    const normalized = normalizeBookingConfig(
      bot.settings.module_configs?.booking,
      posLocations,
      boundPosIds,
    )
    setDraft(draftFromConfig(normalized))
  }, [bot, boundPosIds, posBindingsLoaded, posLocations])

  const availablePosLocations = useMemo(
    () => posLocations.filter((location) => boundPosIds.includes(location.id)),
    [boundPosIds, posLocations],
  )
  const placeholders = useMemo(
    () => buildMessagePlaceholders(bot?.settings.registration_form ?? []),
    [bot?.settings.registration_form],
  )

  const { isSaving, saveError, saveSuccess, save } = useSaveAction(async () => {
    if (!bot || !draft) return
    await botsApi.updateSettings(id, {
      module_configs: buildModuleConfigs(bot.settings.module_configs, draft),
    })
    mutate()
  })

  if (isLoading || !posBindingsLoaded || !draft) {
    return (
      <div className="max-w-4xl">
        <div className="mb-6 h-4 w-32 shimmer rounded" />
        <CardSkeleton />
      </div>
    )
  }

  if (Number.isNaN(id) || isError || !bot) {
    return (
      <div className="max-w-4xl">
        <Link
          to="/dashboard/bots"
          className="mb-6 inline-flex items-center gap-1.5 text-sm text-neutral-500 transition-colors hover:text-neutral-700"
        >
          <ArrowLeft className="h-4 w-4" />
          Назад к ботам
        </Link>
        <ErrorState title="Не удалось загрузить настройки" message="Проверьте подключение и попробуйте снова." />
      </div>
    )
  }

  return (
    <div className="max-w-4xl">
      <div className="animate-in mb-6">
        <Link
          to={`/dashboard/bots/${id}?tab=modules`}
          className="mb-4 inline-flex min-h-11 items-center gap-1.5 rounded text-sm text-neutral-500 transition-colors hover:text-neutral-700"
        >
          <ArrowLeft className="h-4 w-4" />
          К модулям
        </Link>
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Бронирование</h1>
        <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">Настройки модуля</p>
        <p className="mt-2 max-w-2xl text-sm text-neutral-500">
          Сообщение, доступные даты, слоты времени и точки продаж для бронирования столиков.
        </p>
      </div>

      <div className="space-y-6">
        <section className="rounded border border-neutral-900 bg-white p-4 sm:p-5 md:p-6">
          <div className="mb-5">
            <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Точки продаж</p>
            <h3 className="text-sm font-semibold text-neutral-700">Привязка к точкам продаж</h3>
          </div>
          <section className="rounded border border-neutral-200 bg-neutral-50/70 p-4">
            {availablePosLocations.length === 0 ? (
              <p className="text-sm text-neutral-500">
                Для этого бота пока не привязаны точки продаж во вкладке «Подключение».
              </p>
            ) : (
              <div className="space-y-2">
                {availablePosLocations.map((location) => {
                  const isChecked = draft.pos_ids.includes(location.id)
                  return (
                    <div key={location.id} className="rounded border border-neutral-200 bg-white px-3 py-3">
                      <div className="flex flex-wrap items-center justify-between gap-3">
                        <div>
                          <div className="text-sm font-medium text-neutral-900">{location.name}</div>
                          {location.address && (
                            <div className="text-xs text-neutral-500">{location.address}</div>
                          )}
                        </div>
                        <button
                          type="button"
                          role="switch"
                          aria-checked={isChecked}
                          onClick={() =>
                            setDraft((current) =>
                              current
                                ? {
                                    ...current,
                                    pos_ids: isChecked
                                      ? current.pos_ids.filter((id) => id !== location.id)
                                      : [...current.pos_ids, location.id],
                                  }
                                : current,
                            )
                          }
                          className={cn(
                            'relative h-6 w-10 shrink-0 rounded-full transition-colors cursor-pointer',
                            isChecked ? 'bg-accent' : 'bg-neutral-300',
                          )}
                        >
                          <span
                            className={cn(
                              'absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform',
                              isChecked && 'translate-x-4',
                            )}
                          />
                        </button>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </section>
        </section>

        <section className="rounded border border-neutral-900 bg-white p-4 sm:p-5 md:p-6">
          <div className="mb-5">
            <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Старт</p>
            <h3 className="text-sm font-semibold text-neutral-700">Приветственное сообщение</h3>
            <p className="text-sm text-neutral-500 mt-1">Что увидит гость, когда нажмёт «Забронировать столик».</p>
          </div>
          <Suspense fallback={editorFallback()}>
            <MessageContentEditor
              value={draft.intro_content ?? { parts: [{ type: 'text', text: '', parse_mode: 'Markdown' }] }}
              onChange={(content) => setDraft((current) => (current ? { ...current, intro_content: content } : current))}
              onUpload={campaignsApi.uploadFile}
              maxParts={4}
              placeholders={placeholders}
            />
          </Suspense>
        </section>

        <section className="rounded border border-neutral-900 bg-white p-4 sm:p-5 md:p-6">
          <div className="mb-5">
            <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Период</p>
            <h3 className="text-sm font-semibold text-neutral-700">Когда можно бронировать</h3>
            <p className="text-sm text-neutral-500 mt-1">
              За сколько дней наперёд гость может выбрать дату — например, от 0 (сегодня) до 7 (на неделю вперёд).
            </p>
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="space-y-2">
              <label htmlFor="booking-from-days" className="block text-sm font-medium text-neutral-700">Бронь доступна от</label>
              <div className="flex items-stretch">
                <input
                  id="booking-from-days"
                  type="text"
                  inputMode="numeric"
                  value={draft.date_from_days}
                  onChange={(event) =>
                    setDraft((current) => (current ? { ...current, date_from_days: event.target.value.replace(/\D/g, '') } : current))
                  }
                  className={cn(inputClassName, 'rounded-r-none')}
                />
                <span className="inline-flex items-center px-3 text-sm text-neutral-500 bg-neutral-50 border border-l-0 border-neutral-200 rounded-r">
                  дней
                </span>
              </div>
            </div>
            <div className="space-y-2">
              <label htmlFor="booking-to-days" className="block text-sm font-medium text-neutral-700">Бронь доступна до</label>
              <div className="flex items-stretch">
                <input
                  id="booking-to-days"
                  type="text"
                  inputMode="numeric"
                  value={draft.date_to_days}
                  onChange={(event) =>
                    setDraft((current) => (current ? { ...current, date_to_days: event.target.value.replace(/\D/g, '') } : current))
                  }
                  className={cn(inputClassName, 'rounded-r-none')}
                />
                <span className="inline-flex items-center px-3 text-sm text-neutral-500 bg-neutral-50 border border-l-0 border-neutral-200 rounded-r">
                  дней
                </span>
              </div>
            </div>
          </div>
        </section>

        <section className="rounded border border-neutral-900 bg-white p-4 sm:p-5 md:p-6">
          <div className="mb-5 flex items-start justify-between gap-3">
            <div>
              <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Расписание</p>
              <h3 className="text-sm font-semibold text-neutral-700">Слоты времени</h3>
              <p className="text-sm text-neutral-500 mt-1">Интервалы, на которые гость может забронировать столик.</p>
            </div>
            <button
              type="button"
              onClick={() =>
                setDraft((current) =>
                  current
                    ? { ...current, time_slots: [...current.time_slots, { start: '10:00', end: '11:00' }] }
                    : current,
                )
              }
              className="inline-flex min-h-11 items-center gap-1.5 rounded text-sm font-medium text-accent hover:text-accent/80 transition-colors shrink-0"
            >
              <span>+ Добавить слот</span>
            </button>
          </div>
          <div className="space-y-2">
            {draft.time_slots.map((slot, index) => (
              <div
                key={index}
                className="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] gap-2"
              >
                <input
                  type="text"
                  value={slot.start}
                  onChange={(event) =>
                    setDraft((current) => {
                      if (!current) return current
                      const next = [...current.time_slots]
                      next[index] = { ...slot, start: normalizeTimeInput(event.target.value) }
                      return { ...current, time_slots: next }
                    })
                  }
                  className={inputClassName}
                  placeholder="10:00"
                />
                <input
                  type="text"
                  value={slot.end}
                  onChange={(event) =>
                    setDraft((current) => {
                      if (!current) return current
                      const next = [...current.time_slots]
                      next[index] = { ...slot, end: normalizeTimeInput(event.target.value) }
                      return { ...current, time_slots: next }
                    })
                  }
                  className={inputClassName}
                  placeholder="11:00"
                />
                <button
                  type="button"
                  onClick={() =>
                    setDraft((current) =>
                      current
                        ? { ...current, time_slots: current.time_slots.filter((_, slotIndex) => slotIndex !== index) }
                        : current,
                    )
                  }
                  className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-red-50 hover:text-red-600"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>
        </section>

        <section className="rounded border border-neutral-900 bg-white p-4 sm:p-5 md:p-6">
          <div className="mb-5">
            <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">Гости</p>
            <h3 className="text-sm font-semibold text-neutral-700">Количество гостей</h3>
            <p className="text-sm text-neutral-500 mt-1">Варианты, из которых гость выберет размер компании.</p>
          </div>
          <div className="space-y-2">
            {draft.party_size_options.map((option, index) => (
              <div key={index} className="grid grid-cols-[minmax(0,1fr)_auto] gap-2">
                <input
                  type="text"
                  value={option}
                  onChange={(event) =>
                    setDraft((current) => {
                      if (!current) return current
                      const next = [...current.party_size_options]
                      next[index] = event.target.value
                      return { ...current, party_size_options: next }
                    })
                  }
                  className={inputClassName}
                  placeholder="1, 2, 3-5, 6+"
                />
                <button
                  type="button"
                  onClick={() =>
                    setDraft((current) =>
                      current
                        ? { ...current, party_size_options: current.party_size_options.filter((_, optionIndex) => optionIndex !== index) }
                        : current,
                    )
                  }
                  className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-red-50 hover:text-red-600"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() =>
                setDraft((current) =>
                  current ? { ...current, party_size_options: [...current.party_size_options, ''] } : current,
                )
              }
              className="inline-flex min-h-11 items-center gap-1.5 rounded text-sm font-medium text-accent hover:text-accent/80 transition-colors"
            >
              + Добавить вариант
            </button>
          </div>
        </section>
      </div>

      <SaveFooter isSaving={isSaving} saveError={saveError} saveSuccess={saveSuccess} onSave={save} />
    </div>
  )
}
