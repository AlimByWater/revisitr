export interface ClientProfile {
  id: number
  bot_id: number
  telegram_id: number
  username?: string
  first_name: string
  last_name?: string
  phone?: string
  gender?: string
  birth_date?: string
  city?: string
  os?: string
  phone_normalized?: string
  qr_code?: string
  tags: string[]
  registered_at: string
  bot_name: string
  loyalty_balance: number
  loyalty_level?: string
  total_purchases: number
  purchase_count: number
  transactions?: LoyaltyTransaction[]
}

export interface LoyaltyTransaction {
  id: number
  client_id: number
  program_id: number
  type: 'earn' | 'spend' | 'adjust'
  amount: number
  balance_after: number
  description?: string
  created_at: string
}

export interface ClientFilter {
  bot_id?: number
  segment?: string
  search?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  limit: number
  offset: number
}

export interface ClientStats {
  total_clients: number
  total_balance: number
  new_this_month: number
  active_this_week: number
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
}

export interface UpdateClientRequest {
  tags?: string[]
}
