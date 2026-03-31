import type { SegmentFilter } from '@/features/segments/types'

export interface PromotionTrigger {
  type: string  // purchase, purchase_product, purchase_min_items, receipt_sum, event
  product_id?: number
  min_items?: number
  min_amount?: number
  event_type?: string  // birthday, registration, activation, last_purchase
}

export interface PromotionAction {
  type: string  // discount, bonus, data_update, campaign
  discount_percent?: number
  bonus_amount?: number
  tag_add?: string
  level_id?: number
  campaign_id?: number
  message?: string
  media_url?: string
}

export interface PromotionConditions {
  min_amount?: number
  segment_id?: number
  min_visits?: number
}

export interface PromotionResult {
  discount_percent?: number
  bonus_amount?: number
  tag_add?: string
  campaign_id?: number
}

export interface Promotion {
  id: number
  org_id: number
  name: string
  type: 'discount' | 'bonus' | 'tag_update' | 'campaign'
  conditions: PromotionConditions
  result: PromotionResult
  recurrence: 'one_time' | 'daily' | 'weekly' | 'monthly'
  starts_at?: string
  ends_at?: string
  usage_limit?: number
  combinable: boolean
  active: boolean
  created_at: string
  filter: SegmentFilter
  triggers: PromotionTrigger[]
  actions: PromotionAction[]
  combinable_with_loyalty: boolean
}

export interface CreatePromotionRequest {
  name: string
  type: string
  conditions: PromotionConditions
  result: PromotionResult
  recurrence?: string
  starts_at?: string
  ends_at?: string
  usage_limit?: number
  combinable?: boolean
  filter?: SegmentFilter
  triggers?: PromotionTrigger[]
  actions?: PromotionAction[]
  combinable_with_loyalty?: boolean
}

export interface UpdatePromotionRequest {
  name?: string
  conditions?: PromotionConditions
  result?: PromotionResult
  recurrence?: string
  starts_at?: string
  ends_at?: string
  usage_limit?: number
  combinable?: boolean
  active?: boolean
  filter?: SegmentFilter
  triggers?: PromotionTrigger[]
  actions?: PromotionAction[]
  combinable_with_loyalty?: boolean
}

export interface PromoCodeConditions {
  min_amount?: number
}

export interface PromoCode {
  id: number
  org_id: number
  promotion_id?: number
  code: string
  discount_percent?: number
  bonus_amount?: number
  starts_at?: string
  ends_at?: string
  conditions: PromoCodeConditions
  channel?: string
  per_user_limit?: number
  description?: string
  usage_count: number
  usage_limit?: number
  active: boolean
  created_at: string
}

export interface CreatePromoCodeRequest {
  promotion_id?: number
  code?: string
  discount_percent?: number
  bonus_amount?: number
  starts_at?: string
  ends_at?: string
  conditions?: PromoCodeConditions
  usage_limit?: number
  channel?: string
  per_user_limit?: number
  description?: string
}

export interface PromoCodeValidation {
  valid: boolean
  reason?: string
  promo?: PromoResult
}

export interface PromoResult {
  code: string
  discount_percent?: number
  bonus_amount?: number
}

export interface PromoChannelAnalytics {
  channel: string
  code_count: number
  total_usages: number
  unique_clients: number
}
