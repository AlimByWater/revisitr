export type OrderSource = 'lunch' | 'menu'

export type OrderStatus = 'new' | 'sent' | 'cancelled'

export interface OrderItem {
  id: number
  order_id: number
  course_id?: number
  course_title: string
  menu_item_id?: number
  item_name: string
  price: number
  surcharge: number
}

export interface Order {
  id: number
  bot_id: number
  bot_client_id: number
  source: OrderSource
  format_id?: number
  format_name: string
  table_num: string
  total_price: number
  status: OrderStatus
  created_at: string
  /** Присутствует только в org-wide выборке (GET /orders). */
  bot_name?: string
  items: OrderItem[]
}
