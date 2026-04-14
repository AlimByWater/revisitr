import { Link, useParams } from 'react-router-dom'
import { useState, useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'
import { useBotQuery } from '@/features/bots/queries'
import { botsApi } from '@/features/bots/api'
import { menusApi } from '@/features/menus/api'
import { usePOSQuery } from '@/features/pos/queries'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Check,
  Settings,
  Puzzle,
  Store,
  Eye,
  GripVertical,
  ChevronDown,
  ChevronRight,
  Link2,
} from 'lucide-react'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { Bot, BotSettings, BotButton, FormField } from '@/features/bots/types'
import type { POSLocation } from '@/features/pos/types'
import { TelegramPreview, MessageContentEditor } from '@/features/telegram-preview'
import { campaignsApi } from '@/features/campaigns/api'
import type { MessageContent } from '@/features/telegram-preview'

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const TABS = [
  { id: 'connection', label: 'Подключение', icon: Link2 },
  { id: 'general', label: 'Основное', icon: Settings },
  { id: 'modules', label: 'Модули', icon: Puzzle },
  { id: 'preview', label: 'Превью', icon: Eye },
] as const

type TabId = (typeof TABS)[number]['id']

const statusConfig = {
  active: { label: 'Активен', className: 'bg-green-500 text-white' },
  inactive: { label: 'Неактивен', className: 'bg-neutral-200 text-neutral-500' },
  error: { label: 'Ошибка', className: 'bg-red-500 text-white' },
} as const

const inputClassName = cn(
  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
  'text-sm placeholder:text-neutral-400',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors',
  'disabled:opacity-50 disabled:cursor-not-allowed',
)

const MODULE_DEFS = [
  { key: 'loyalty', label: 'Лояльность', description: 'Начисление и списание бонусов' },
  { key: 'menu', label: 'Меню', description: 'Показ меню заведения в боте' },
  { key: 'marketplace', label: 'Маркетплейс', description: 'Каталог товаров для заказа' },
  { key: 'feedback', label: 'Обратная связь', description: 'Сбор отзывов от клиентов' },
  { key: 'booking', label: 'Бронирование', description: 'Бронирование столиков' },
] as const

const FORM_PRESETS: FormField[] = [
  { name: 'first_name', label: 'Имя', type: 'text', required: true },
  { name: 'phone', label: 'Телефон', type: 'phone', required: true },
  { name: 'birthday', label: 'Дата рождения', type: 'date', required: false },
  { name: 'city', label: 'Город', type: 'text', required: false },
]

const TEMPLATE_VARIABLES = ['{first_name}'] as const

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function buttonContentFromValue(value: string): MessageContent {
  return {
    parts: [{ type: 'text', text: value, parse_mode: 'Markdown' }],
  }
}

function normalizeButton(button: BotButton): BotButton {
  return {
    ...button,
    type: 'text',
    content:
      button.content && button.content.parts?.length > 0
        ? button.content
        : buttonContentFromValue(button.value ?? ''),
  }
}

function buttonValueFromContent(content?: MessageContent, fallback = ''): string {
  if (!content || !content.parts?.length) return fallback

  const firstTextPart = content.parts.find((part) => part.type === 'text' && part.text?.trim())
  if (firstTextPart?.text) return firstTextPart.text

  const firstCaptionPart = content.parts.find((part) => part.text?.trim())
  if (firstCaptionPart?.text) return firstCaptionPart.text

  return fallback
}

function welcomeMessageFromContent(content?: MessageContent, fallback = ''): string {
  if (!content || !content.parts?.length) return fallback

  const firstTextLike = content.parts.find((part) => part.text?.trim())
  if (firstTextLike?.text) return firstTextLike.text

  return fallback
}

function buttonSummary(content?: MessageContent, fallback = ''): string {
  const source = content && content.parts?.length > 0 ? content : buttonContentFromValue(fallback)
  const text = buttonValueFromContent(source, fallback).replace(/\s+/g, ' ').trim()
  const mediaCount = source.parts.filter((part) => part.type !== 'text').length
  const inlineButtons = (source.buttons || []).reduce((sum, row) => sum + row.length, 0)

  const details = [
    source.parts.length > 1 ? `${source.parts.length} блока` : null,
    mediaCount > 0 ? `${mediaCount} медиа` : null,
    inlineButtons > 0 ? `${inlineButtons} inline-кнопок` : null,
  ].filter(Boolean)

  if (!text && details.length === 0) return 'Ответ пока не настроен'
  if (!text) return details.join(' · ')
  return [text.slice(0, 90) + (text.length > 90 ? '…' : ''), details.join(' · ')].filter(Boolean).join(' · ')
}

/** Per-tab save hook returning state + handler */
function useSaveAction(_id: number, updater: () => Promise<void>) {
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
      setSaveError('Не удалось сохранить. Попробуйте снова.')
    } finally {
      setIsSaving(false)
    }
  }

  return { isSaving, saveError, saveSuccess, save }
}

function SaveButton({
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
    <div className="flex items-center justify-between mt-6 pt-5 border-t border-surface-border">
      <div>
        {saveError && (
          <div className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-red-50 border border-red-100">
            <span className="text-red-500 font-mono text-xs">&#x2715;</span>
            <p className="text-sm text-red-600">{saveError}</p>
          </div>
        )}
        {saveSuccess && (
          <div className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-green-50 border border-green-100">
            <Check className="w-3.5 h-3.5 text-green-600" />
            <p className="text-sm text-green-600 font-medium">Сохранено</p>
          </div>
        )}
      </div>
      <button
        type="button"
        onClick={onSave}
        disabled={isSaving}
        className={cn(
          'py-2.5 px-6 rounded-xl',
          'bg-accent text-white text-sm font-semibold',
          'hover:bg-accent-hover active:bg-accent/80',
          'transition-all duration-150',
          'focus:outline-none focus:ring-2 focus:ring-accent/20',
          'shadow-sm shadow-accent/20',
          'disabled:opacity-50 disabled:cursor-not-allowed',
        )}
      >
        {isSaving ? 'Сохранение...' : 'Сохранить'}
      </button>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Sortable item wrapper
// ---------------------------------------------------------------------------

function SortableItem({
  id,
  children,
}: {
  id: string
  children: (props: {
    listeners: ReturnType<typeof useSortable>['listeners']
    attributes: ReturnType<typeof useSortable>['attributes']
    setNodeRef: ReturnType<typeof useSortable>['setNodeRef']
  }) => React.ReactNode
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  return (
    <div ref={setNodeRef} style={style}>
      {children({ listeners, attributes, setNodeRef })}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Page Component
// ---------------------------------------------------------------------------

export default function BotDetailPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)
  const { data: bot, isLoading, isError } = useBotQuery(isNaN(id) || id <= 0 ? 0 : id)
  const [activeTab, setActiveTab] = useState<TabId>('connection')

  const [settings, setSettings] = useState<BotSettings | null>(null)

  useEffect(() => {
    if (bot?.settings) {
      setSettings({
        welcome_message: bot.settings.welcome_message ?? '',
        welcome_content: bot.settings.welcome_content,
        modules: bot.settings.modules ?? [],
        buttons: (bot.settings.buttons ?? []).map(normalizeButton),
        registration_form: bot.settings.registration_form ?? [],
      })
    }
  }, [bot])

  if (isLoading) {
    return (
      <div className="max-w-4xl">
        <div className="mb-6 animate-in">
          <div className="h-4 w-32 shimmer rounded" />
        </div>
        <div className="space-y-6">
          {[0, 1, 2].map((i) => (
            <div key={i} className={cn('animate-in', `animate-in-delay-${i + 1}`)}>
              <CardSkeleton />
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (isNaN(id) || id <= 0 || isError || !bot) {
    return (
      <div className="max-w-4xl">
        <Link
          to="/dashboard/bots"
          className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Назад к списку
        </Link>
        <ErrorState
          title="Не удалось загрузить бота"
          message="Проверьте подключение к серверу и попробуйте снова."
        />
      </div>
    )
  }

  const status = statusConfig[bot.status as keyof typeof statusConfig] ?? {
    label: bot.status,
    className: 'bg-neutral-100 text-neutral-500',
  }

  return (
    <div className="max-w-4xl">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6 animate-in">
        <Link
          to="/dashboard/bots"
          className="p-2 rounded-lg text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h1 className="font-serif text-2xl font-bold text-neutral-900 truncate">
              {bot.name}
            </h1>
            <span
              className={cn(
                'font-mono text-[10px] font-semibold px-2.5 py-1 rounded-full uppercase tracking-wider shrink-0',
                status.className,
              )}
            >
              {status.label}
            </span>
          </div>
          <p className="text-sm text-neutral-400 mt-0.5 truncate">
            {bot.username ? `@${bot.username}` : '---'}
            <span className="mx-2 text-neutral-300">|</span>
            Создан {formatDate(bot.created_at)}
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-surface-border overflow-x-auto animate-in animate-in-delay-1">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex items-center gap-1.5 px-3 py-2.5 text-sm font-medium whitespace-nowrap',
              'border-b-2 transition-colors',
              activeTab === tab.id
                ? 'border-accent text-accent'
                : 'border-transparent text-neutral-500 hover:text-neutral-700',
            )}
          >
            <tab.icon className="w-4 h-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="animate-in animate-in-delay-2">
        {settings && (
          <>
            {activeTab === 'connection' && <ConnectionTab bot={bot} botId={id} />}
            {activeTab === 'general' && (
              <GeneralTab botId={id} settings={settings} setSettings={setSettings} botName={bot.name} />
            )}
            {activeTab === 'modules' && (
              <ModulesTab botId={id} settings={settings} setSettings={setSettings} />
            )}
            {activeTab === 'preview' && <PreviewTab settings={settings} botName={bot.name} />}
          </>
        )}
      </div>
    </div>
  )
}

// ===========================================================================
// Tab: Подключение (Connection) — General info + POS
// ===========================================================================

function ConnectionTab({ bot, botId }: { bot: Bot; botId: number }) {
  const [advanced, setAdvanced] = useState(false)

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
      {/* Section: Bot Info */}
      <div className="flex items-center justify-between mb-5">
        <h2 className="text-lg font-semibold text-neutral-900">
          <span className="block font-mono text-[10px] uppercase tracking-widest text-neutral-400 font-normal mb-0.5">
            Информация
          </span>
          Общие данные
        </h2>
        <button
          type="button"
          onClick={() => setAdvanced(!advanced)}
          className={cn(
            'text-xs font-medium px-3 py-1.5 rounded-lg transition-colors',
            advanced
              ? 'bg-accent/10 text-accent'
              : 'bg-neutral-100 text-neutral-500 hover:bg-neutral-200',
          )}
        >
          {advanced ? 'По умолчанию' : 'Расширенный'}
        </button>
      </div>

      <div className="space-y-4">
        <InfoRow label="Название" value={bot.name} />
        <InfoRow label="Username" value={bot.username ? `@${bot.username}` : '---'} />
        <InfoRow label="Статус" value={bot.status} />
        <InfoRow label="Дата создания" value={formatDate(bot.created_at)} />
        {bot.program_id && (
          <div className="flex items-center justify-between py-2">
            <span className="text-sm text-neutral-500">Программа лояльности</span>
            <Link
              to={`/dashboard/loyalty/${bot.program_id}`}
              className="text-sm text-accent hover:underline font-medium"
            >
              Программа #{bot.program_id}
            </Link>
          </div>
        )}

        {advanced && (
          <>
            <p className="text-xs text-neutral-400 leading-relaxed">
              Техническая информация о боте. ID — уникальный идентификатор бота в системе. Org ID — идентификатор вашей организации. Токен — секретный ключ для связи с Telegram API.
            </p>
            {bot.token_masked && (
              <div>
                <span className="block text-sm text-neutral-500 mb-1">Токен</span>
                <div
                  className={cn(
                    'px-4 py-2.5 rounded-lg bg-neutral-50 border border-surface-border',
                    'font-mono text-sm text-neutral-600 select-none',
                  )}
                >
                  {bot.token_masked}
                </div>
              </div>
            )}
            <div>
              <span className="block text-sm text-neutral-500 mb-1">ID бота</span>
              <div className="font-mono text-sm text-neutral-600">{bot.id}</div>
            </div>
            <div>
              <span className="block text-sm text-neutral-500 mb-1">ID организации</span>
              <div className="font-mono text-sm text-neutral-600">{bot.org_id}</div>
            </div>
          </>
        )}
      </div>

      {/* Section divider */}
      <div className="border-t border-surface-border my-8"></div>

      {/* Section: POS */}
      <h3 className="text-base font-semibold text-neutral-900 mb-4">POS-точки бота</h3>
      <POSSection botId={botId} />
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-surface-border/50 last:border-0">
      <span className="text-sm text-neutral-500">{label}</span>
      <span className="text-sm font-medium text-neutral-900">{value}</span>
    </div>
  )
}

/** POS section embedded inside ConnectionTab */
function POSSection({ botId }: { botId: number }) {
  const { data: posLocations, isLoading: posLoading } = usePOSQuery()
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [loaded, setLoaded] = useState(false)
  const { isSaving, saveError, saveSuccess, save } = useSaveAction(botId, () =>
    menusApi.setBotPOSLocations(botId, selectedIds),
  )

  // Load current bindings
  useEffect(() => {
    if (loaded) return
    menusApi.getBotPOSLocations(botId).then((res) => {
      setSelectedIds(res.pos_ids ?? [])
      setLoaded(true)
    }).catch(() => {
      setLoaded(true)
    })
  }, [botId, loaded])

  if (posLoading || !loaded) {
    return <p className="text-sm text-neutral-500">Загрузка POS-точек...</p>
  }

  const locations = posLocations ?? []

  if (locations.length === 0) {
    return (
      <div className="text-center py-6">
        <Store className="w-10 h-10 text-neutral-300 mx-auto mb-3" />
        <p className="text-sm text-neutral-500 mb-4">Нет POS-точек</p>
        <Link
          to="/dashboard/pos"
          className={cn(
            'inline-flex items-center gap-1.5 py-2 px-4 rounded-lg text-sm font-medium',
            'bg-accent text-white hover:bg-accent-hover transition-colors',
          )}
        >
          Управление POS
          <ChevronRight className="w-4 h-4" />
        </Link>
      </div>
    )
  }

  const togglePos = (posId: number) => {
    setSelectedIds((prev) =>
      prev.includes(posId) ? prev.filter((x) => x !== posId) : [...prev, posId],
    )
  }

  return (
    <>
      <p className="text-sm text-neutral-400 mb-5">
        Привяжите точки продаж к боту. Если выбрано несколько, гость сможет выбрать нужную точку при запуске бота через кнопки.
      </p>

      <div className="space-y-2">
        {locations.map((loc: POSLocation) => (
          <label
            key={loc.id}
            className={cn(
              'flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-colors',
              selectedIds.includes(loc.id) ? 'bg-accent/5 border border-accent/20' : 'bg-neutral-50 border border-transparent',
            )}
          >
            <input
              type="checkbox"
              checked={selectedIds.includes(loc.id)}
              onChange={() => togglePos(loc.id)}
              disabled={isSaving}
              className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
            />
            <div>
              <p className="text-sm font-medium text-neutral-900">{loc.name}</p>
              {loc.address && <p className="text-xs text-neutral-500">{loc.address}</p>}
            </div>
          </label>
        ))}
      </div>

      <SaveButton isSaving={isSaving} saveError={saveError} saveSuccess={saveSuccess} onSave={save} />
    </>
  )
}

// ===========================================================================
// Tab: Основное (General) — Messages + Buttons + Form combined
// ===========================================================================

function GeneralTab({
  botId,
  settings,
  setSettings,
  botName,
}: {
  botId: number
  settings: BotSettings
  setSettings: React.Dispatch<React.SetStateAction<BotSettings | null>>
  botName: string
}) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const [expandedButtons, setExpandedButtons] = useState<string[]>([])

  const { isSaving, saveError, saveSuccess, save } = useSaveAction(botId, () =>
    botsApi.updateSettings(botId, {
      welcome_message: settings.welcome_message,
      welcome_content: settings.welcome_content,
      buttons: settings.buttons.map((button) => ({
        ...button,
        type: 'text',
        value: buttonValueFromContent(button.content, button.value),
        content: button.content,
      })),
      registration_form: settings.registration_form,
    }),
  )

  // --- Buttons dnd ---
  const buttonSensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor),
  )

  const buttonIds = settings.buttons.map((_, i) => `btn-${i}`)

  const handleButtonDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = buttonIds.indexOf(String(active.id))
    const newIndex = buttonIds.indexOf(String(over.id))
    setSettings((s) =>
      s ? { ...s, buttons: arrayMove(s.buttons, oldIndex, newIndex) } : s,
    )
  }

  const addButton = () => {
    if (settings.buttons.length >= 10) return
    setSettings((s) =>
      s
        ? {
            ...s,
            buttons: [
              ...s.buttons,
              { label: '', type: 'text', value: '', content: buttonContentFromValue('') },
            ],
          }
        : s,
    )
  }

  const updateButton = (index: number, field: keyof BotButton, value: string) => {
    setSettings((s) => {
      if (!s) return s
      const updated = s.buttons.map((btn, i) => (i === index ? { ...btn, [field]: value } : btn))
      return { ...s, buttons: updated }
    })
  }

  const removeButton = (index: number) => {
    setSettings((s) => (s ? { ...s, buttons: s.buttons.filter((_, i) => i !== index) } : s))
  }

  const updateButtonContent = (index: number, content: MessageContent) => {
    setSettings((s) => {
      if (!s) return s
      const updated = s.buttons.map((btn, i) =>
        i === index
          ? {
              ...btn,
              content,
              value: buttonValueFromContent(content, btn.value),
            }
          : btn,
      )
      return { ...s, buttons: updated }
    })
  }

  const toggleButtonExpanded = (id: string) => {
    setExpandedButtons((current) =>
      current.includes(id) ? current.filter((item) => item !== id) : [...current, id],
    )
  }

  // --- Form dnd ---
  const formSensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor),
  )

  const fieldIds = settings.registration_form.map((_, i) => `field-${i}`)

  const handleFieldDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = fieldIds.indexOf(String(active.id))
    const newIndex = fieldIds.indexOf(String(over.id))
    setSettings((s) =>
      s ? { ...s, registration_form: arrayMove(s.registration_form, oldIndex, newIndex) } : s,
    )
  }

  const addField = () => {
    setSettings((s) =>
      s
        ? {
            ...s,
            registration_form: [
              ...s.registration_form,
              { name: '', label: '', type: 'text', required: false },
            ],
          }
        : s,
    )
  }

  const addPreset = (preset: FormField) => {
    setSettings((s) =>
      s ? { ...s, registration_form: [...s.registration_form, { ...preset }] } : s,
    )
  }

  const updateField = (index: number, field: keyof FormField, value: string | boolean) => {
    setSettings((s) => {
      if (!s) return s
      const updated = s.registration_form.map((f, i) =>
        i === index ? { ...f, [field]: value } : f,
      )
      return { ...s, registration_form: updated }
    })
  }

  const removeField = (index: number) => {
    setSettings((s) =>
      s ? { ...s, registration_form: s.registration_form.filter((_, i) => i !== index) } : s,
    )
  }

  // --- Template variables ---
  const insertVariable = (variable: string) => {
    const ta = textareaRef.current
    if (!ta) return
    const start = ta.selectionStart
    const end = ta.selectionEnd
    const text = settings.welcome_message
    const newText = text.substring(0, start) + variable + text.substring(end)
    setSettings((s) => (s ? { ...s, welcome_message: newText } : s))
    requestAnimationFrame(() => {
      ta.focus()
      const pos = start + variable.length
      ta.setSelectionRange(pos, pos)
    })
  }

  // Welcome content state (composite message)
  const welcomeContent: import('@/features/telegram-preview').MessageContent = settings.welcome_content
    && settings.welcome_content.parts?.length > 0
    ? settings.welcome_content
    : {
        parts: settings.welcome_message
          ? [{ type: 'text' as const, text: settings.welcome_message, parse_mode: 'Markdown' as const }]
          : [{ type: 'text' as const, text: '', parse_mode: 'Markdown' as const }],
      }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
      {/* Section 1: Welcome Message */}
      <h3 className="text-base font-semibold text-neutral-900 mb-4">Приветственное сообщение</h3>
      <p className="text-sm text-neutral-400 mb-5">
        Составное сообщение, которое клиент видит при запуске бота. Добавляйте текст, фото, видео, стикеры и кнопки.
      </p>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        {/* Left: Editor */}
        <div>
          {/* Template variable chips */}
          <div className="flex flex-wrap gap-2 mb-3">
            {TEMPLATE_VARIABLES.map((v) => (
              <button
                key={v}
                type="button"
                onClick={() => insertVariable(v)}
                className="px-2.5 py-1 rounded-md text-xs font-mono bg-accent/10 text-accent hover:bg-accent/20 transition-colors"
              >
                {v}
              </button>
            ))}
          </div>
          <MessageContentEditor
            value={welcomeContent}
            onChange={(content) => {
              setSettings((s) => s ? {
                ...s,
                welcome_content: content,
                welcome_message: welcomeMessageFromContent(content, s.welcome_message),
              } : s)
            }}
            onUpload={campaignsApi.uploadFile}
            maxParts={5}
          />
        </div>

        {/* Right: Live Preview */}
        <div className="flex justify-center">
          <TelegramPreview
            botName={botName}
            content={welcomeContent}
            showFrame
          />
        </div>
      </div>

      {/* Section divider */}
      <div className="border-t border-surface-border my-8"></div>

      {/* Section 2: Buttons */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-base font-semibold text-neutral-900">Кнопки бота</h3>
        <button
          type="button"
          onClick={addButton}
          disabled={isSaving || settings.buttons.length >= 10}
          className={cn(
            'flex items-center gap-1.5 text-sm font-medium text-accent',
            'hover:text-accent/80 transition-colors',
            'disabled:opacity-50 disabled:cursor-not-allowed',
          )}
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>
      <p className="text-xs text-neutral-400 mb-4">
        Сейчас используются только обычные кнопки меню Telegram. Для каждой кнопки можно собрать полноценный ответ:
        текст, фото, видео, документы, несколько сообщений подряд и inline-кнопки внутри ответа.
      </p>

      {settings.buttons.length === 0 ? (
        <p className="text-sm text-neutral-400 text-center py-4">Нет кнопок меню</p>
      ) : (
        <DndContext sensors={buttonSensors} collisionDetection={closestCenter} onDragEnd={handleButtonDragEnd}>
          <SortableContext items={buttonIds} strategy={verticalListSortingStrategy}>
            <div className="space-y-3">
              {settings.buttons.map((button, index) => (
                <SortableItem key={buttonIds[index]} id={buttonIds[index]}>
                  {({ listeners, attributes }) => (
                    <div className="rounded-xl border border-surface-border bg-neutral-50 overflow-hidden">
                      <div className="flex items-start gap-3 p-3">
                        <button
                          type="button"
                          className="mt-1 text-neutral-400 hover:text-neutral-600 cursor-grab active:cursor-grabbing touch-none"
                          {...listeners}
                          {...attributes}
                          aria-label="Перетащить"
                        >
                          <GripVertical className="w-4 h-4" />
                        </button>

                        <button
                          type="button"
                          onClick={() => toggleButtonExpanded(buttonIds[index])}
                          className="flex-1 text-left min-w-0"
                        >
                          <div className="flex items-start justify-between gap-3">
                            <div className="min-w-0">
                              <div className="text-sm font-semibold text-neutral-900">
                                {button.label || `Кнопка ${index + 1}`}
                              </div>
                              <div className="text-xs text-neutral-500 mt-1 break-words">
                                {buttonSummary(button.content, button.value)}
                              </div>
                            </div>
                            <ChevronDown
                              className={cn(
                                'w-4 h-4 mt-0.5 text-neutral-400 transition-transform shrink-0',
                                expandedButtons.includes(buttonIds[index]) && 'rotate-180',
                              )}
                            />
                          </div>
                        </button>

                        <button
                          type="button"
                          onClick={() => removeButton(index)}
                          disabled={isSaving}
                          className="mt-1 p-2 rounded-lg text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
                          aria-label={`Удалить кнопку ${index + 1}`}
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>

                      {expandedButtons.includes(buttonIds[index]) && (
                        <div className="border-t border-surface-border bg-white p-4">
                          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
                            <div className="space-y-3">
                              <input
                                type="text"
                                value={button.label}
                                onChange={(e) => updateButton(index, 'label', e.target.value)}
                                placeholder="Название кнопки в клавиатуре"
                                disabled={isSaving}
                                className={inputClassName}
                                aria-label={`Название кнопки ${index + 1}`}
                              />
                              <MessageContentEditor
                                value={button.content && button.content.parts?.length > 0
                                  ? button.content
                                  : buttonContentFromValue(button.value)}
                                onChange={(content) => updateButtonContent(index, content)}
                                onUpload={campaignsApi.uploadFile}
                                maxParts={5}
                              />
                            </div>

                            <div className="flex justify-center">
                              <TelegramPreview
                                botName={button.label || botName}
                                content={button.content && button.content.parts?.length > 0
                                  ? button.content
                                  : buttonContentFromValue(button.value)}
                                showFrame
                              />
                            </div>
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </SortableItem>
              ))}
            </div>
          </SortableContext>
        </DndContext>
      )}

      {settings.buttons.length >= 10 && (
        <p className="text-xs text-neutral-400 mt-2">Максимум 10 кнопок</p>
      )}

      {/* Section divider */}
      <div className="border-t border-surface-border my-8"></div>

      {/* Section 3: Registration Form */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-base font-semibold text-neutral-900">Форма регистрации</h3>
        <button
          type="button"
          onClick={addField}
          disabled={isSaving}
          className={cn(
            'flex items-center gap-1.5 text-sm font-medium text-accent',
            'hover:text-accent/80 transition-colors',
            'disabled:opacity-50 disabled:cursor-not-allowed',
          )}
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {/* Presets */}
      <div className="flex flex-wrap gap-2 mb-4">
        {FORM_PRESETS.map((preset) => (
          <button
            key={preset.name}
            type="button"
            onClick={() => addPreset(preset)}
            disabled={isSaving}
            className={cn(
              'px-3 py-1.5 rounded-lg text-xs font-medium transition-colors',
              'bg-neutral-100 text-neutral-600 hover:bg-neutral-200',
              'disabled:opacity-50 disabled:cursor-not-allowed',
            )}
          >
            + {preset.label}
          </button>
        ))}
      </div>

      {settings.registration_form.length === 0 ? (
        <p className="text-sm text-neutral-400 text-center py-4">Нет полей анкеты</p>
      ) : (
        <DndContext sensors={formSensors} collisionDetection={closestCenter} onDragEnd={handleFieldDragEnd}>
          <SortableContext items={fieldIds} strategy={verticalListSortingStrategy}>
            <div className="space-y-3">
              {settings.registration_form.map((field, index) => (
                <SortableItem key={fieldIds[index]} id={fieldIds[index]}>
                  {({ listeners, attributes }) => (
                    <div className="flex items-start gap-3 p-3 rounded-lg bg-neutral-50">
                      <button
                        type="button"
                        className="mt-2.5 text-neutral-400 hover:text-neutral-600 cursor-grab active:cursor-grabbing touch-none"
                        {...listeners}
                        {...attributes}
                        aria-label="Перетащить"
                      >
                        <GripVertical className="w-4 h-4" />
                      </button>
                      <div className="flex-1 grid grid-cols-2 gap-3">
                        <input
                          type="text"
                          value={field.name}
                          onChange={(e) => updateField(index, 'name', e.target.value)}
                          placeholder="Ключ поля (латиница)"
                          disabled={isSaving}
                          className={inputClassName}
                          aria-label={`Ключ поля ${index + 1}`}
                        />
                        <input
                          type="text"
                          value={field.label}
                          onChange={(e) => updateField(index, 'label', e.target.value)}
                          placeholder="Название для пользователя"
                          disabled={isSaving}
                          className={inputClassName}
                          aria-label={`Название поля ${index + 1}`}
                        />
                        <select
                          value={field.type}
                          onChange={(e) => updateField(index, 'type', e.target.value)}
                          disabled={isSaving}
                          className={inputClassName}
                          aria-label={`Тип поля ${index + 1}`}
                        >
                          <option value="text">Текст</option>
                          <option value="email">Email</option>
                          <option value="phone">Телефон</option>
                          <option value="date">Дата</option>
                          <option value="select">Выбор</option>
                        </select>
                        <label className="flex items-center gap-2 px-4 py-2.5">
                          <input
                            type="checkbox"
                            checked={field.required}
                            onChange={(e) => updateField(index, 'required', e.target.checked)}
                            disabled={isSaving}
                            className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
                          />
                          <span className="text-sm text-neutral-700">Обязательное</span>
                        </label>
                        {field.type === 'select' && (
                          <div className="col-span-2">
                            <input
                              type="text"
                              value={(field.options ?? []).join(', ')}
                              onChange={(e) => {
                                const opts = e.target.value.split(',').map((s) => s.trim()).filter(Boolean)
                                setSettings((s) => {
                                  if (!s) return s
                                  const updated = s.registration_form.map((f, i) =>
                                    i === index ? { ...f, options: opts } : f,
                                  )
                                  return { ...s, registration_form: updated }
                                })
                              }}
                              placeholder="Варианты через запятую (будут отображаться как кнопки)"
                              disabled={isSaving}
                              className={inputClassName}
                            />
                            <p className="text-[11px] text-neutral-400 mt-1">Варианты выбора отображаются как кнопки в боте</p>
                          </div>
                        )}
                        {(field.type === 'email' || field.type === 'phone' || field.type === 'date') && (
                          <p className="col-span-2 text-[11px] text-neutral-400">
                            {field.type === 'email' && 'Бот проверит формат email перед принятием ответа'}
                            {field.type === 'phone' && 'Бот проверит формат номера телефона перед принятием ответа'}
                            {field.type === 'date' && 'Бот проверит формат даты (ДД.ММ.ГГГГ) перед принятием ответа'}
                          </p>
                        )}
                      </div>
                      <button
                        type="button"
                        onClick={() => removeField(index)}
                        disabled={isSaving}
                        className="mt-2.5 p-2 rounded-lg text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
                        aria-label={`Удалить поле ${index + 1}`}
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                </SortableItem>
              ))}
            </div>
          </SortableContext>
        </DndContext>
      )}

      {/* Single combined save button */}
      <SaveButton isSaving={isSaving} saveError={saveError} saveSuccess={saveSuccess} onSave={save} />
    </div>
  )
}

// ===========================================================================
// Tab: Модули (Modules)
// ===========================================================================

function ModulesTab({
  botId,
  settings,
  setSettings,
}: {
  botId: number
  settings: BotSettings
  setSettings: React.Dispatch<React.SetStateAction<BotSettings | null>>
}) {
  const { isSaving, saveError, saveSuccess, save } = useSaveAction(botId, () =>
    botsApi.updateSettings(botId, { modules: settings.modules }),
  )

  const toggleModule = (moduleKey: string) => {
    setSettings((s) => {
      if (!s) return s
      const modules = s.modules.includes(moduleKey)
        ? s.modules.filter((m) => m !== moduleKey)
        : [...s.modules, moduleKey]
      return { ...s, modules }
    })
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
      <h2 className="text-lg font-semibold text-neutral-900 mb-5">
        <span className="block font-mono text-[10px] uppercase tracking-widest text-neutral-400 font-normal mb-0.5">
          Функционал
        </span>
        Модули бота
      </h2>

      <div className="grid gap-3 sm:grid-cols-2">
        {MODULE_DEFS.map((mod) => {
          const isActive = settings.modules.includes(mod.key)
          const configHref = mod.key === 'menu' ? '/dashboard/menus'
            : mod.key === 'marketplace' ? '/dashboard/marketplace'
            : null
          return (
            <div
              key={mod.key}
              className={cn(
                'p-4 rounded-xl border transition-colors',
                isActive
                  ? 'border-accent/30 bg-accent/5'
                  : 'border-surface-border bg-neutral-50',
              )}
            >
              <div className="flex items-center justify-between">
                <div className="min-w-0 mr-3">
                  <p className="text-sm font-medium text-neutral-900">{mod.label}</p>
                  <p className="text-xs text-neutral-500 mt-0.5">{mod.description}</p>
                </div>
                <button
                  type="button"
                  role="switch"
                  aria-checked={isActive}
                  aria-label={`${mod.label} модуль`}
                  onClick={() => toggleModule(mod.key)}
                  disabled={isSaving}
                  className={cn(
                    'relative shrink-0 w-10 h-6 rounded-full transition-colors',
                    'focus:outline-none focus:ring-2 focus:ring-accent/20',
                    'disabled:opacity-50 disabled:cursor-not-allowed',
                    isActive ? 'bg-accent' : 'bg-neutral-300',
                  )}
                >
                  <span
                    className={cn(
                      'absolute top-0.5 left-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform',
                      isActive && 'translate-x-4',
                    )}
                  />
                </button>
              </div>
              {isActive && configHref && (
                <Link
                  to={configHref}
                  className="inline-flex items-center gap-1.5 mt-3 text-xs text-accent hover:text-accent/80 font-medium transition-colors"
                >
                  Настроить <ChevronRight className="w-3 h-3" />
                </Link>
              )}
            </div>
          )
        })}
      </div>

      {/* Anchor links for active modules */}
      {settings.modules.length > 0 && (
        <div className="mt-5 pt-5 border-t border-surface-border">
          <p className="text-xs font-medium text-neutral-400 uppercase tracking-wider mb-3">Активные модули</p>
          <div className="flex flex-wrap gap-2">
            {settings.modules.map((mod) => {
              const def = MODULE_DEFS.find((d) => d.key === mod)
              return (
                <a
                  key={mod}
                  href={`#module-${mod}`}
                  className="text-xs font-medium px-3 py-1.5 rounded-lg bg-accent/10 text-accent hover:bg-accent/20 transition-colors"
                >
                  {def?.label ?? mod}
                </a>
              )
            })}
          </div>
        </div>
      )}

      {/* Inline module config sections */}
      {settings.modules.includes('menu') && (
        <div id="module-menu" className="mt-5 pt-5 border-t border-surface-border scroll-mt-20">
          <h3 className="text-sm font-semibold text-neutral-900 mb-2">Настройка: Меню</h3>
          <p className="text-xs text-neutral-400 mb-3">Меню заведения, которое будет отображаться в боте.</p>
          <Link
            to="/dashboard/menus"
            className="inline-flex items-center gap-1.5 text-sm text-accent hover:text-accent/80 font-medium transition-colors"
          >
            Управление меню <ChevronRight className="w-3.5 h-3.5" />
          </Link>
        </div>
      )}

      {settings.modules.includes('marketplace') && (
        <div id="module-marketplace" className="mt-5 pt-5 border-t border-surface-border scroll-mt-20">
          <h3 className="text-sm font-semibold text-neutral-900 mb-2">Настройка: Маркетплейс</h3>
          <p className="text-xs text-neutral-400 mb-3">Каталог товаров для заказа через бота.</p>
          <Link
            to="/dashboard/marketplace"
            className="inline-flex items-center gap-1.5 text-sm text-accent hover:text-accent/80 font-medium transition-colors"
          >
            Управление товарами <ChevronRight className="w-3.5 h-3.5" />
          </Link>
        </div>
      )}

      <SaveButton isSaving={isSaving} saveError={saveError} saveSuccess={saveSuccess} onSave={save} />
    </div>
  )
}

// ===========================================================================
// Tab: Превью (Preview) — Telegram iOS mockup
// ===========================================================================

function PreviewTab({ settings, botName }: { settings: BotSettings; botName: string }) {
  // Build MessageContent from settings for preview
  const content: import('@/features/telegram-preview').MessageContent =
    settings.welcome_content && settings.welcome_content.parts?.length > 0
      ? settings.welcome_content
      : {
          parts: settings.welcome_message
            ? [{
                type: 'text' as const,
                text: settings.welcome_message
                  .replace(/\{first_name\}/g, 'Александр')
                  .replace(/\{bonus_balance\}/g, '1 250')
                  .replace(/\{loyalty_level\}/g, 'Gold'),
                parse_mode: 'Markdown' as const,
              }]
            : [],
        }

  return (
    <div className="flex justify-center">
      <TelegramPreview
        botName={botName}
        content={content}
        showFrame
      />
    </div>
  )
}
