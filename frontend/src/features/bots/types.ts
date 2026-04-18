export interface Bot {
  id: number
  org_id: number
  name: string
  username: string
  token_masked?: string
  status: 'active' | 'inactive' | 'pending' | 'error'
  settings: BotSettings
  created_at: string
  updated_at: string
  client_count?: number
  program_id?: number
  is_managed?: boolean
  managed_bot_id?: number
  created_by_telegram_id?: number
}

export interface BotSettings {
  modules: string[]
  buttons: BotButton[]
  registration_form: FormField[]
  welcome_message: string
  welcome_content?: import('@/features/telegram-preview').MessageContent
  module_configs?: ModuleConfigs
  pos_selector_enabled?: boolean
  contacts_pos_ids?: number[]
}

export interface BotButton {
  label: string
  type: string
  value: string
  content?: import('@/features/telegram-preview').MessageContent
  managed_by_module?: string | null
  is_system?: boolean
}

export interface FormField {
  name: string
  label: string
  type: string
  required: boolean
  options?: string[]
}

export interface ModuleConfigs {
  menu?: MenuModuleConfig
  booking?: BookingModuleConfig
  feedback?: FeedbackModuleConfig
}

export interface MenuModuleConfig {
  unavailable_message?: string
}

export interface BookingTimeSlot {
  start: string
  end: string
}

export interface BookingModuleConfig {
  intro_content?: import('@/features/telegram-preview').MessageContent
  date_from_days?: number
  date_to_days?: number
  time_slots?: BookingTimeSlot[]
  party_size_options?: string[]
  pos_ids?: number[]
}

export interface FeedbackModuleConfig {
  prompt_message?: string
  success_message?: string
}

export interface CreateBotRequest {
  name: string
  token: string
  program_id?: number
}

export interface CreateManagedBotRequest {
  name: string
  username: string
  description: string
  welcome_message?: string
  registration_form?: FormField[]
  modules?: string[]
}

export interface CreateManagedBotResponse {
  bot_id: number
  deep_link: string
  status: string
}

export interface ActivationLinkResponse {
  deep_link: string
  expires_at: string
}

export interface PostCode {
  id: number
  org_id: number
  code: string
  content: PostCodeContent
  created_at: string
  updated_at: string
}

export interface PostCodeContent {
  text?: string
  media_urls?: string[]
  media_type?: string
  buttons?: { text: string; url?: string; data?: string }[][]
}
