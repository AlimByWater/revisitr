import { useNavigate, useSearchParams } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import {
  useCampaignsQuery,
  useCampaignQuery,
  useCreateCampaignMutation,
  useUpdateCampaignMutation,
  useDeleteCampaignMutation,
  useSendCampaignMutation,
  useCreateScenarioMutation,
  usePreviewAudienceMutation,
  useScenarioQuery,
  useUploadFileMutation,
} from '@/features/campaigns/queries'
import { ArrowLeft, Send, Zap, Upload, X, FileText, Save, Trash2 } from 'lucide-react'
import type { SegmentFilter } from '@/features/segments/types'
import type { AudienceFilter, Campaign } from '@/features/campaigns/types'
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
  { value: 'date', label: 'Дата' },
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

type TabType = 'create' | 'drafts'

export default function CreateCampaignPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const { data: bots } = useBotsQuery()
  const createCampaignMutation = useCreateCampaignMutation()
  const updateCampaignMutation = useUpdateCampaignMutation()
  const deleteCampaignMutation = useDeleteCampaignMutation()
  const sendCampaignMutation = useSendCampaignMutation()
  const createScenarioMutation = useCreateScenarioMutation()
  const previewMutation = usePreviewAudienceMutation()
  const uploadMutation = useUploadFileMutation()

  const [activeTab, setActiveTab] = useState<TabType>('create')
  const [format, setFormat] = useState<Format>('manual')
  const [name, setName] = useState('')
  const [botId, setBotId] = useState<number | ''>('')
  const [message, setMessage] = useState('')
  const [audienceFilter, setAudienceFilter] = useState<SegmentFilter>({})
  const [mediaUrl, setMediaUrl] = useState('')

  // Auto-scenario state
  const [triggerType, setTriggerType] = useState<TriggerType>('inactive_days')
  const [triggerDays, setTriggerDays] = useState<number | ''>('')
  const [triggerCount, setTriggerCount] = useState<number | ''>('')
  const [triggerThreshold, setTriggerThreshold] = useState<number | ''>('')
  const [triggerDate, setTriggerDate] = useState('')
  const [daysBefore, setDaysBefore] = useState<number | ''>('')
  const [daysAfter, setDaysAfter] = useState<number | ''>('')

  // Server draft editing
  const [editingDraftId, setEditingDraftId] = useState<number | null>(null)

  // Server drafts from campaigns list
  const { data: campaignsData, mutate: revalidateCampaigns } = useCampaignsQuery()
  const drafts = (campaignsData?.items ?? []).filter(
    (c: Campaign) => c.status === 'draft',
  )

  // Clone support
  const cloneId = searchParams.get('clone')
  const cloneType = searchParams.get('type') as 'campaign' | 'scenario' | null
  const cloneCampaignId = cloneType === 'campaign' && cloneId ? Number(cloneId) : 0
  const cloneScenarioId = cloneType === 'scenario' && cloneId ? Number(cloneId) : 0
  const { data: cloneCampaign } = useCampaignQuery(cloneCampaignId)
  const { data: cloneScenario } = useScenarioQuery(cloneScenarioId)

  // Load clone data
  useEffect(() => {
    if (cloneType === 'campaign' && cloneCampaign) {
      setFormat('manual')
      setName(`${cloneCampaign.name} (копия)`)
      setBotId(cloneCampaign.bot_id)
      setMessage(cloneCampaign.message)
      setMediaUrl(cloneCampaign.media_url || '')
      if (cloneCampaign.audience_filter) {
        const af = cloneCampaign.audience_filter
        const sf: SegmentFilter = {}
        if (af.bot_id) sf.bot_id = af.bot_id
        if (af.tags) sf.tags = af.tags
        if (af.level_id) sf.level_id = af.level_id
        setAudienceFilter(sf)
      }
      setSearchParams({}, { replace: true })
    }
  }, [cloneCampaign, cloneType, setSearchParams])

  useEffect(() => {
    if (cloneType === 'scenario' && cloneScenario) {
      setFormat('auto')
      setName(`${cloneScenario.name} (копия)`)
      setBotId(cloneScenario.bot_id)
      setMessage(cloneScenario.message)
      const tt = cloneScenario.trigger_type === 'holiday' ? 'date' : cloneScenario.trigger_type
      setTriggerType(tt as TriggerType)
      const tc = cloneScenario.trigger_config
      if (tc.days) setTriggerDays(tc.days)
      if (tc.count) setTriggerCount(tc.count)
      if (tc.threshold) setTriggerThreshold(tc.threshold)
      if (tc.month && tc.day) {
        const year = new Date().getFullYear()
        const dateStr = `${year}-${String(tc.month).padStart(2, '0')}-${String(tc.day).padStart(2, '0')}`
        setTriggerDate(dateStr)
      }
      if (cloneScenario.timing) {
        if (cloneScenario.timing.days_before !== undefined) setDaysBefore(cloneScenario.timing.days_before)
        if (cloneScenario.timing.days_after !== undefined) setDaysAfter(cloneScenario.timing.days_after)
      }
      setSearchParams({}, { replace: true })
    }
  }, [cloneScenario, cloneType, setSearchParams])

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

  function resetForm() {
    setEditingDraftId(null)
    setName('')
    setFormat('manual')
    setBotId('')
    setMessage('')
    setAudienceFilter({})
    setMediaUrl('')
    setTriggerType('inactive_days')
    setTriggerDays('')
    setTriggerCount('')
    setTriggerThreshold('')
    setTriggerDate('')
    setDaysBefore('')
    setDaysAfter('')
  }

  // ── Server draft operations ─────────────────────────────────────────────────

  function handleSaveDraft() {
    if (!name.trim() || botId === '') return

    const campaignData = {
      name: name.trim(),
      message: message.trim(),
      audience_filter: toAudienceFilter({
        ...audienceFilter,
        bot_id: botId as number,
      }),
      media_url: mediaUrl || undefined,
    }

    if (editingDraftId) {
      // Update existing draft
      updateCampaignMutation.mutate(
        { id: editingDraftId, data: campaignData },
        { onSuccess: () => revalidateCampaigns() },
      )
    } else {
      // Create new draft
      createCampaignMutation.mutate(
        { bot_id: botId as number, ...campaignData },
        {
          onSuccess: (created) => {
            setEditingDraftId(created.id)
            revalidateCampaigns()
          },
        },
      )
    }
  }

  function loadDraft(draft: Campaign) {
    setEditingDraftId(draft.id)
    setFormat('manual')
    setName(draft.name)
    setBotId(draft.bot_id)
    setMessage(draft.message)
    setMediaUrl(draft.media_url || '')
    if (draft.audience_filter) {
      const af = draft.audience_filter
      const sf: SegmentFilter = {}
      if (af.bot_id) sf.bot_id = af.bot_id
      if (af.tags) sf.tags = af.tags
      if (af.level_id) sf.level_id = af.level_id
      setAudienceFilter(sf)
    }
    setActiveTab('create')
  }

  function handleDeleteDraft(id: number) {
    deleteCampaignMutation.mutate(id, {
      onSuccess: () => {
        if (editingDraftId === id) resetForm()
        revalidateCampaigns()
      },
    })
  }

  // ── File attachment ─────────────────────────────────────────────────────────

  function handleFileUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    uploadMutation.mutate(file, {
      onSuccess: (url) => setMediaUrl(url),
    })
    e.target.value = ''
  }

  function isImageUrl(url: string) {
    return /\.(jpg|jpeg|png|gif|webp)(\?|$)/i.test(url)
  }

  // ── Submit ──────────────────────────────────────────────────────────────────

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!isValid) return

    if (isManual) {
      if (editingDraftId) {
        // Update draft then send
        updateCampaignMutation.mutate(
          {
            id: editingDraftId,
            data: {
              name: name.trim(),
              message: message.trim(),
              audience_filter: toAudienceFilter({
                ...audienceFilter,
                bot_id: botId as number,
              }),
              media_url: mediaUrl || undefined,
            },
          },
          {
            onSuccess: () => {
              sendCampaignMutation.mutate(editingDraftId, {
                onSuccess: () => navigate('/dashboard/campaigns'),
              })
            },
          },
        )
      } else {
        // Create new campaign (status=draft), then send
        createCampaignMutation.mutate(
          {
            bot_id: botId as number,
            name: name.trim(),
            message: message.trim(),
            audience_filter: toAudienceFilter({
              ...audienceFilter,
              bot_id: botId as number,
            }),
            media_url: mediaUrl || undefined,
          },
          {
            onSuccess: (created) => {
              sendCampaignMutation.mutate(created.id, {
                onSuccess: () => navigate('/dashboard/campaigns'),
              })
            },
          },
        )
      }
    } else {
      const timing: { days_before?: number; days_after?: number } = {}
      if (daysBefore !== '') timing.days_before = daysBefore
      if (daysAfter !== '') timing.days_after = daysAfter

      createScenarioMutation.mutate(
        {
          bot_id: botId as number,
          name: name.trim(),
          trigger_type: triggerType === 'date' ? 'holiday' : triggerType,
          trigger_config: buildTriggerConfig(),
          message: message.trim(),
          timing: Object.keys(timing).length > 0 ? timing : undefined,
        },
        {
          onSuccess: () => navigate('/dashboard/campaigns'),
        },
      )
    }
  }

  const isSaving =
    createCampaignMutation.isPending ||
    updateCampaignMutation.isPending ||
    sendCampaignMutation.isPending ||
    createScenarioMutation.isPending

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
          Создать рассылку
        </h1>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 p-1 bg-neutral-100 rounded-lg w-fit mb-6 animate-in">
        <button
          type="button"
          onClick={() => setActiveTab('create')}
          className={cn(
            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-150',
            activeTab === 'create'
              ? 'bg-white text-neutral-900 shadow-sm'
              : 'text-neutral-500 hover:text-neutral-700',
          )}
        >
          Создать
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('drafts')}
          className={cn(
            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-150',
            activeTab === 'drafts'
              ? 'bg-white text-neutral-900 shadow-sm'
              : 'text-neutral-500 hover:text-neutral-700',
          )}
        >
          Черновики ({drafts.length})
        </button>
      </div>

      {activeTab === 'drafts' ? (
        /* ── Drafts Tab ──────────────────────────────────────────── */
        <div className="space-y-3">
          {drafts.length === 0 ? (
            <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-12 text-center">
              <p className="text-sm text-neutral-400">Нет сохранённых черновиков</p>
            </div>
          ) : (
            drafts.map((draft) => (
              <div
                key={draft.id}
                className={cn(
                  'bg-white rounded-2xl shadow-sm border border-surface-border p-4',
                  'flex items-center gap-4 hover:bg-neutral-50 transition-colors',
                )}
              >
                <button
                  type="button"
                  onClick={() => loadDraft(draft)}
                  className="flex-1 text-left min-w-0"
                >
                  <p className="text-sm font-medium text-neutral-900 truncate">
                    {draft.name || 'Без названия'}
                  </p>
                  <p className="text-xs text-neutral-400 mt-0.5">
                    {new Date(draft.updated_at).toLocaleDateString('ru-RU', {
                      day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit',
                    })}
                  </p>
                </button>
                <button
                  type="button"
                  onClick={() => handleDeleteDraft(draft.id)}
                  disabled={deleteCampaignMutation.isPending}
                  className="p-2 rounded-lg text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            ))
          )}
        </div>
      ) : (
        /* ── Create Tab ──────────────────────────────────────────── */
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Editing draft indicator */}
          {editingDraftId && (
            <div className="flex items-center gap-2 px-3 py-2 bg-amber-50 border border-amber-200 rounded-lg">
              <span className="text-xs text-amber-700">
                Редактирование черновика #{editingDraftId}
              </span>
              <button
                type="button"
                onClick={resetForm}
                className="ml-auto text-xs text-amber-600 hover:text-amber-800 underline"
              >
                Новая рассылка
              </button>
            </div>
          )}

          {/* ── Block 1: Создать ─────────────────────────────────── */}
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 space-y-5">
            <p className={blockHeaderClass}>Создать</p>

            {/* Campaign name */}
            <div>
              <label htmlFor="name" className={labelClass}>
                Название
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

            {/* Format */}
            <div>
              <label htmlFor="format" className={labelClass}>
                Формат
              </label>
              <select
                id="format"
                value={format}
                onChange={(e) => {
                  setFormat(e.target.value as Format)
                  // Clear server draft when switching to auto (no drafts for scenarios)
                  if (e.target.value === 'auto' && editingDraftId) {
                    setEditingDraftId(null)
                  }
                }}
                className={inputClass}
              >
                <option value="manual">Ручная рассылка</option>
                <option value="auto">Авто-сценарий</option>
              </select>
            </div>
          </div>

          {/* ── Block 2: Настройки ───────────────────────────────── */}
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 space-y-5">
            <p className={blockHeaderClass}>Настройки</p>

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

            {/* Auto-scenario specific fields */}
            {!isManual && (
              <>
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

                {/* Trigger config fields */}
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
                        setTriggerCount(e.target.value ? Number(e.target.value) : '')
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
                        setTriggerThreshold(e.target.value ? Number(e.target.value) : '')
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

                {/* Timing */}
                <div>
                  <p className="text-sm font-medium text-neutral-700 mb-3">
                    Время отправки (опционально)
                  </p>
                  <div className="flex gap-4">
                    <div className="flex-1 max-w-[180px]">
                      <label htmlFor="days-before" className="block text-xs text-neutral-500 mb-1">
                        Дней до
                      </label>
                      <input
                        id="days-before"
                        type="number"
                        min={0}
                        value={daysBefore}
                        onChange={(e) =>
                          setDaysBefore(e.target.value ? Number(e.target.value) : '')
                        }
                        placeholder="0"
                        className={inputClass}
                      />
                    </div>
                    <div className="flex-1 max-w-[180px]">
                      <label htmlFor="days-after" className="block text-xs text-neutral-500 mb-1">
                        Дней после
                      </label>
                      <input
                        id="days-after"
                        type="number"
                        min={0}
                        value={daysAfter}
                        onChange={(e) =>
                          setDaysAfter(e.target.value ? Number(e.target.value) : '')
                        }
                        placeholder="0"
                        className={inputClass}
                      />
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>

          {/* ── Block 3: Редактировать сообщение ─────────────────── */}
          <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 space-y-5">
            <p className={blockHeaderClass}>Редактировать сообщение</p>

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

            {/* File attachment */}
            <div>
              <p className="text-sm font-medium text-neutral-700 mb-2">Вложение</p>
              {mediaUrl ? (
                <div className="flex items-center gap-3 p-3 bg-neutral-50 rounded-lg border border-neutral-200">
                  {isImageUrl(mediaUrl) ? (
                    <img
                      src={mediaUrl}
                      alt="Вложение"
                      className="w-12 h-12 rounded-lg object-cover"
                    />
                  ) : (
                    <div className="w-12 h-12 rounded-lg bg-neutral-200 flex items-center justify-center">
                      <FileText className="w-5 h-5 text-neutral-500" />
                    </div>
                  )}
                  <span className="flex-1 text-sm text-neutral-700 truncate">
                    {mediaUrl.split('/').pop()}
                  </span>
                  <button
                    type="button"
                    onClick={() => setMediaUrl('')}
                    className="p-1.5 rounded-lg text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ) : (
                <label
                  className={cn(
                    'flex items-center gap-2 px-4 py-2.5 rounded-lg border border-dashed border-neutral-300',
                    'text-sm text-neutral-500 cursor-pointer',
                    'hover:bg-neutral-50 hover:border-neutral-400 transition-colors',
                    uploadMutation.isPending && 'opacity-50 pointer-events-none',
                  )}
                >
                  {uploadMutation.isPending ? (
                    <div className="w-4 h-4 border-2 border-neutral-300 border-t-neutral-600 rounded-full animate-spin" />
                  ) : (
                    <Upload className="w-4 h-4" />
                  )}
                  {uploadMutation.isPending ? 'Загрузка...' : 'Прикрепить файл'}
                  <input
                    type="file"
                    className="hidden"
                    accept="image/*,.pdf,.doc,.docx"
                    onChange={handleFileUpload}
                    disabled={uploadMutation.isPending}
                  />
                  <span className="ml-auto text-xs text-neutral-400">до 50 МБ</span>
                </label>
              )}
              {uploadMutation.isError && (
                <p className="text-xs text-red-500 mt-1">Не удалось загрузить файл</p>
              )}
            </div>
          </div>

          {/* ── Bottom Actions ───────────────────────────────────── */}
          <div className="flex items-center justify-end gap-3">
            {isManual && (
              <button
                type="button"
                onClick={handleSaveDraft}
                disabled={!name.trim() || botId === '' || isSaving}
                className={cn(
                  'flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium',
                  'border border-neutral-200 text-neutral-700',
                  'hover:bg-neutral-50 transition-colors',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
                )}
              >
                <Save className="w-4 h-4" />
                {updateCampaignMutation.isPending ? 'Сохранение...' : 'Сохранить черновик'}
              </button>
            )}
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
              disabled={!isValid || isSaving}
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
              {isSaving
                ? 'Создание...'
                : isManual
                  ? 'Отправить рассылку'
                  : 'Создать авто-сценарий'}
            </button>
          </div>

          {(activeMutation.isError || sendCampaignMutation.isError) && (
            <p className="text-sm text-red-600 mt-3 text-right">
              {isManual
                ? 'Не удалось создать рассылку. Попробуйте ещё раз.'
                : 'Не удалось создать сценарий. Попробуйте ещё раз.'}
            </p>
          )}
        </form>
      )}
    </div>
  )
}
