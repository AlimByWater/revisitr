export interface Bot {
  id: number
  org_id: number
  name: string
  username: string
  token_masked?: string
  status: 'active' | 'inactive' | 'error'
  settings: BotSettings
  created_at: string
  updated_at: string
  client_count?: number
  program_id?: number
}

export interface BotSettings {
  modules: string[]
  buttons: BotButton[]
  registration_form: FormField[]
  welcome_message: string
}

export interface BotButton {
  label: string
  type: string
  value: string
}

export interface FormField {
  name: string
  label: string
  type: string
  required: boolean
}

export interface CreateBotRequest {
  name: string
  token: string
  program_id?: number
}
