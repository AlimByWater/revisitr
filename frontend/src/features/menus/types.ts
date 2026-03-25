export interface Menu {
  id: number
  org_id: number
  integration_id?: number
  name: string
  source: 'manual' | 'pos_import'
  last_synced_at?: string
  created_at: string
  updated_at: string
  categories?: MenuCategory[]
}

export interface MenuCategory {
  id: number
  menu_id: number
  name: string
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
  image_url?: string
  tags: string[]
  external_id?: string
  is_available: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface CreateMenuRequest {
  name: string
}

export interface CreateMenuCategoryRequest {
  name: string
  sort_order?: number
}

export interface CreateMenuItemRequest {
  name: string
  description?: string
  price: number
  image_url?: string
  tags?: string[]
}

export interface UpdateMenuItemRequest {
  name?: string
  description?: string
  price?: number
  image_url?: string
  tags?: string[]
  is_available?: boolean
  sort_order?: number
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
