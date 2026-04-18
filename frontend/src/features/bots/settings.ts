import type { MessageContent } from '@/features/telegram-preview'
import type { POSLocation } from '@/features/pos/types'
import type {
  BookingModuleConfig,
  BookingTimeSlot,
  Bot,
  BotButton,
  BotSettings,
  FormField,
  ModuleConfigs,
} from './types'

export const MODULE_DEFS = [
  { key: 'loyalty', label: 'Лояльность', description: 'Начисление и списание бонусов' },
  { key: 'menu', label: 'Меню', description: 'Показ меню заведения в боте' },
  { key: 'feedback', label: 'Связаться', description: 'Вопросы гостей и обратная связь' },
  { key: 'booking', label: 'Бронирование', description: 'Бронирование столиков' },
] as const

export const STANDARD_FIELD_PRESETS: FormField[] = [
  { name: 'first_name', label: 'Имя', type: 'text', required: true },
  { name: 'phone', label: 'Телефон', type: 'phone', required: true },
  { name: 'birthday', label: 'Дата рождения', type: 'date', required: false },
  { name: 'city', label: 'Город', type: 'text', required: false },
  { name: 'email', label: 'E-mail', type: 'email', required: false },
  {
    name: 'gender',
    label: 'Пол',
    type: 'select',
    required: false,
    options: ['Женский', 'Мужской', 'Не хочу отвечать'],
  },
] as const

export const STANDARD_FIELD_NAMES = new Set(
  STANDARD_FIELD_PRESETS.map((field) => field.name),
)

export interface BotRequirement {
  id: string
  label: string
  tab: 'connection' | 'general' | 'modules'
}

const DEFAULT_BOOKING_PARTY_OPTIONS = ['1', '2', '3-5', '6+']
const DEFAULT_FEEDBACK_PROMPT = 'Напишите ваш вопрос:'
const DEFAULT_FEEDBACK_SUCCESS = 'Ваше сообщение отправлено.'

function defaultBookingIntro(): MessageContent {
  return {
    parts: [
      {
        type: 'text',
        text: 'Забронировать столик можно по телефону или кнопкой ниже.',
        parse_mode: 'Markdown',
      },
    ],
  }
}

function withFallbackItems<T>(items: T[] | undefined, fallback: T[]): T[] {
  return items && items.length > 0 ? items : fallback
}

function normalizeFeedbackConfig(configs: ModuleConfigs['feedback'] | undefined) {
  return {
    prompt_message: configs?.prompt_message ?? DEFAULT_FEEDBACK_PROMPT,
    success_message: configs?.success_message ?? DEFAULT_FEEDBACK_SUCCESS,
  }
}

function applyBookingDefaults(
  config: BookingModuleConfig | undefined,
  posLocations: POSLocation[],
  boundPosIds: number[],
): BookingModuleConfig {
  const normalized = {
    intro_content: config?.intro_content ?? defaultBookingIntro(),
    date_from_days: config?.date_from_days ?? 0,
    date_to_days: config?.date_to_days ?? 7,
    time_slots: config?.time_slots ?? [],
    party_size_options: config?.party_size_options ?? [],
    pos_ids: config?.pos_ids ?? [],
  }

  return {
    ...normalized,
    time_slots: withFallbackItems(
      normalized.time_slots,
      buildDefaultBookingSlots(posLocations, boundPosIds),
    ),
    party_size_options: withFallbackItems(
      normalized.party_size_options,
      DEFAULT_BOOKING_PARTY_OPTIONS,
    ),
    pos_ids: withFallbackItems(normalized.pos_ids, boundPosIds),
  }
}

export function isStandardField(field: Pick<FormField, 'name'>): boolean {
  return STANDARD_FIELD_NAMES.has(field.name)
}

export function canAddPreset(
  fields: FormField[],
  presetName: string,
): boolean {
  return !fields.some((field) => field.name === presetName)
}

export function getEditableButtons(buttons: BotButton[]): BotButton[] {
  return buttons.filter((button) => !button.is_system && !button.managed_by_module)
}

export function getManagedButtonSummary(modules: string[]): string[] {
  const summaries: string[] = ['Контакты', 'На главную']

  if (modules.includes('loyalty')) summaries.unshift('Лояльность')
  if (modules.includes('menu')) summaries.push('Меню')
  if (modules.includes('booking')) summaries.push('Бронирование')
  if (modules.includes('feedback')) summaries.push('Связаться')

  return summaries
}

export function normalizeBotSettings(
  settings: Partial<BotSettings> | undefined,
): BotSettings {
  return {
    modules: settings?.modules ?? [],
    buttons: (settings?.buttons ?? []).map((button) => ({
      ...button,
      managed_by_module: button.managed_by_module ?? null,
      is_system: button.is_system ?? Boolean(button.managed_by_module),
    })),
    registration_form: settings?.registration_form ?? [],
    welcome_message: settings?.welcome_message ?? '',
    welcome_content: settings?.welcome_content,
    module_configs: normalizeModuleConfigs(settings?.module_configs),
    pos_selector_enabled: settings?.pos_selector_enabled ?? false,
    contacts_pos_ids: settings?.contacts_pos_ids ?? [],
  }
}

export function normalizeModuleConfigs(
  configs: ModuleConfigs | undefined,
): ModuleConfigs {
  return {
    menu: {
      unavailable_message: configs?.menu?.unavailable_message ?? '',
    },
    booking: {
      intro_content: configs?.booking?.intro_content ?? defaultBookingIntro(),
      date_from_days: configs?.booking?.date_from_days ?? 0,
      date_to_days: configs?.booking?.date_to_days ?? 7,
      time_slots: configs?.booking?.time_slots ?? [],
      party_size_options: configs?.booking?.party_size_options ?? [],
      pos_ids: configs?.booking?.pos_ids ?? [],
    },
    feedback: normalizeFeedbackConfig(configs?.feedback),
  }
}

export function withModuleDefaults(
  moduleKey: string,
  current: ModuleConfigs | undefined,
  posLocations: POSLocation[],
  boundPosIds: number[],
): ModuleConfigs {
  const configs = normalizeModuleConfigs(current)

  if (moduleKey === 'booking') {
    return {
      ...configs,
      booking: applyBookingDefaults(configs.booking, posLocations, boundPosIds),
    }
  }

  if (moduleKey === 'feedback') {
    return {
      ...configs,
      feedback: normalizeFeedbackConfig(configs.feedback),
    }
  }

  return configs
}

export function deriveBotRequirements(args: {
  bot: Pick<Bot, 'name' | 'username' | 'program_id' | 'created_by_telegram_id'>
  settings: BotSettings
  boundPosIds: number[]
}): BotRequirement[] {
  const requirements: BotRequirement[] = []
  const { bot, settings, boundPosIds } = args

  if (!bot.name?.trim()) {
    requirements.push({
      id: 'bot-name',
      label: 'Выберите название',
      tab: 'connection',
    })
  }

  if (!bot.username?.trim()) {
    requirements.push({
      id: 'bot-link',
      label: 'Выберите ссылку',
      tab: 'connection',
    })
  }

  if (boundPosIds.length === 0) {
    requirements.push({
      id: 'bot-pos',
      label: 'Привяжите точку продаж',
      tab: 'connection',
    })
  }

  const welcomeHasText = Boolean(settings.welcome_message?.trim())
  const welcomeHasContent = Boolean(settings.welcome_content?.parts?.length)
  if (!welcomeHasText && !welcomeHasContent) {
    requirements.push({
      id: 'welcome',
      label: 'Заполните приветственное сообщение',
      tab: 'general',
    })
  }

  if (settings.modules.includes('loyalty') && !bot.program_id) {
    requirements.push({
      id: 'loyalty-program',
      label: 'Выберите программу лояльности',
      tab: 'connection',
    })
  }

  if (settings.modules.includes('feedback') && !bot.created_by_telegram_id) {
    requirements.push({
      id: 'feedback-recipient',
      label: 'Подключите админский Telegram',
      tab: 'modules',
    })
  }

  return requirements
}

export function normalizeTimeInput(value: string): string {
  const digits = value.replace(/\D/g, '').slice(0, 4)

  if (digits.length <= 2) return digits
  if (digits.length === 3) return `${digits.slice(0, 1)}:${digits.slice(1)}`

  const hours = Number(digits.slice(0, 2))
  const minutes = Number(digits.slice(2, 4))

  if (Number.isNaN(hours) || Number.isNaN(minutes)) return value

  const normalizedHours = String(Math.max(0, Math.min(hours, 23))).padStart(2, '0')
  const normalizedMinutes = String(Math.max(0, Math.min(minutes, 59))).padStart(2, '0')

  return `${normalizedHours}:${normalizedMinutes}`
}

export function buildDefaultBookingSlots(
  posLocations: POSLocation[],
  boundPosIds: number[],
): BookingTimeSlot[] {
  const selectedPOS = posLocations.find((pos) => boundPosIds.includes(pos.id))
  const schedule = selectedPOS?.schedule

  if (schedule) {
    const firstOpenDay = Object.values(schedule).find(
      (day) => !day.closed && day.open && day.close,
    )

    if (firstOpenDay && firstOpenDay.open < firstOpenDay.close) {
      const slots = buildHourlySlots(firstOpenDay.open, firstOpenDay.close)
      if (slots.length > 0) return slots
    }
  }

  return buildHourlySlots('10:00', '20:00')
}

export function buildHourlySlots(start: string, end: string): BookingTimeSlot[] {
  const startHour = Number(start.split(':')[0] ?? 0)
  const endHour = Number(end.split(':')[0] ?? 0)

  if (Number.isNaN(startHour) || Number.isNaN(endHour) || endHour <= startHour) {
    return []
  }

  return Array.from({ length: endHour - startHour }, (_, index) => {
    const hour = startHour + index
    return {
      start: `${String(hour).padStart(2, '0')}:00`,
      end: `${String(hour + 1).padStart(2, '0')}:00`,
    }
  })
}

export function normalizeBookingConfig(
  config: BookingModuleConfig | undefined,
  posLocations: POSLocation[],
  boundPosIds: number[],
): BookingModuleConfig {
  return applyBookingDefaults(config, posLocations, boundPosIds)
}
