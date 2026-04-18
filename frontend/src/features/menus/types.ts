export interface Menu {
  id: number
  org_id: number
  integration_id?: number
  name: string
  source: 'manual' | 'pos_import'
  intro_content?: import('@/features/telegram-preview').MessageContent
  last_synced_at?: string
  created_at: string
  updated_at: string
  categories?: MenuCategory[]
  bindings?: MenuPOSBinding[]
}

export interface MenuCategory {
  id: number
  menu_id: number
  name: string
  icon_emoji?: string
  icon_image_url?: string
  sort_order: number
  created_at: string
  items?: MenuItem[]
}

export interface MenuItem {
  id: number
  category_id: number
  name: string
  description?: string
  price: number
  weight?: string
  image_url?: string
  tags: string[]
  external_id?: string
  is_available: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface MenuPOSBinding {
  menu_id: number
  pos_id: number
  pos_name?: string
  is_active: boolean
  created_at?: string
  updated_at?: string
}

export interface CreateMenuRequest {
  name: string
}

export interface UpdateMenuRequest {
  name?: string
  intro_content?: import('@/features/telegram-preview').MessageContent
  bindings?: MenuPOSBindingRequest[]
}

export interface CreateMenuCategoryRequest {
  name: string
  icon_emoji?: string
  icon_image_url?: string
  sort_order?: number
}

export interface UpdateMenuCategoryRequest {
  name?: string
  icon_emoji?: string
  icon_image_url?: string
  sort_order?: number
}

export interface CreateMenuItemRequest {
  name: string
  description?: string
  price: number
  weight?: string
  image_url?: string
  tags?: string[]
}

export interface UpdateMenuItemRequest {
  name?: string
  description?: string
  price?: number
  weight?: string
  image_url?: string
  tags?: string[]
  is_available?: boolean
  sort_order?: number
}

export interface MenuPOSBindingRequest {
  pos_id: number
  is_active: boolean
}

export interface ClientOrderStats {
  total_orders: number
  total_amount: number
  avg_amount: number
  last_order_at?: string
  top_items: TopOrderItem[]
}

export interface TopOrderItem {
  name: string
  order_count: number
  total_qty: number
  total_sum: number
}
