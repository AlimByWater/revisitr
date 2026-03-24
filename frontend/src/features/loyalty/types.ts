export interface ProgramConfig {
  welcome_bonus: number
  currency_name: string
}

export interface LoyaltyProgram {
  id: number
  org_id: number
  name: string
  type: 'bonus' | 'discount'
  config: ProgramConfig
  is_active: boolean
  created_at: string
  updated_at: string
  levels?: LoyaltyLevel[]
}

export interface LoyaltyLevel {
  id: number
  program_id: number
  name: string
  threshold: number
  reward_percent: number
  reward_type: 'percent' | 'fixed'
  reward_amount: number
  sort_order: number
}

export interface CreateProgramRequest {
  name: string
  type: 'bonus' | 'discount'
  config: ProgramConfig
}
