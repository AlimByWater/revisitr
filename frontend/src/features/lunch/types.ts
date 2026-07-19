import type { MenuItem } from '@/features/menus/types'

export type LunchPriceMode = 'fixed' | 'sum_of_items' | 'base_plus_surcharge'

export interface LunchCourseItem {
  course_id: number
  menu_item_id: number
  surcharge: number
  menu_item?: MenuItem
}

export interface LunchCourse {
  id: number
  program_id: number
  code: string
  title: string
  menu_category_id: number
  sort_order: number
  items: LunchCourseItem[]
}

export interface LunchFormat {
  id: number
  program_id: number
  name: string
  price_mode: LunchPriceMode
  base_price: number
  is_active: boolean
  sort_order: number
  course_ids: number[]
}

export interface LunchAvailabilitySlot {
  id?: number
  program_id?: number
  weekday: number // ISO: 1 = понедельник … 7 = воскресенье
  time_from: string
  time_to: string
}

export interface LunchProgram {
  id: number
  bot_id: number
  name: string
  description: string
  is_active: boolean
  created_at: string
  updated_at: string
  courses: LunchCourse[]
  formats: LunchFormat[]
  availability: LunchAvailabilitySlot[]
}

export interface LunchOrderItem {
  id: number
  lunch_order_id: number
  course_id?: number
  course_title: string
  menu_item_id?: number
  item_name: string
  price: number
  surcharge: number
}

export interface LunchOrder {
  id: number
  bot_id: number
  bot_client_id: number
  format_id?: number
  format_name: string
  table_num: string
  total_price: number
  status: 'new' | 'sent' | 'cancelled'
  created_at: string
  items: LunchOrderItem[]
}

export interface UpsertLunchProgramRequest {
  name?: string
  description?: string
  is_active?: boolean
}

export interface LunchCourseItemRequest {
  menu_item_id: number
  surcharge: number
}

export interface SaveLunchCourseRequest {
  code: string
  title: string
  menu_category_id: number
  sort_order?: number
  items: LunchCourseItemRequest[]
}

export interface SaveLunchFormatRequest {
  name: string
  price_mode: LunchPriceMode
  base_price: number
  is_active: boolean
  sort_order?: number
  course_ids: number[]
}
