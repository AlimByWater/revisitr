export interface IntegrationConfig {
  api_url?: string
  api_key?: string
  org_id?: string
  username?: string
  password?: string
  sync_interval?: number
}

export interface Integration {
  id: number
  org_id: number
  type: 'iiko' | 'rkeeper' | '1c' | 'mock'
  config: IntegrationConfig
  status: 'active' | 'inactive' | 'error'
  last_sync_at?: string
  created_at: string
  updated_at: string
}

export interface CreateIntegrationRequest {
  type: string
  config: IntegrationConfig
}

export interface UpdateIntegrationRequest {
  config?: IntegrationConfig
  status?: string
}

export interface ExternalOrder {
  id: number
  integration_id: number
  external_id: string
  client_id?: number
  items: OrderItem[]
  total: number
  ordered_at?: string
  synced_at: string
}

export interface OrderItem {
  name: string
  quantity: number
  price: number
}

export interface IntegrationStats {
  total_orders: number
  total_revenue: number
  matched_clients: number
  unmatched_orders: number
}

export interface POSCustomer {
  external_id: string
  phone: string
  name: string
  email?: string
  birthday?: string
  balance: number
  card_number?: string
}

export interface POSMenu {
  categories: MenuCategory[]
}

export interface MenuCategory {
  name: string
  items: POSMenuItem[]
}

export interface POSMenuItem {
  external_id: string
  name: string
  price: number
  description?: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
}
