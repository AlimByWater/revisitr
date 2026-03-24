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
  type: 'discount' | 'bonus' | 'tag' | 'campaign'
  conditions: PromotionConditions
  result: PromotionResult
  recurrence: 'one_time' | 'daily' | 'weekly' | 'monthly'
  starts_at?: string
  ends_at?: string
  usage_limit?: number
  combinable: boolean
  active: boolean
  created_at: string
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
