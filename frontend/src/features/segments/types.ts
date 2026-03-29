export interface Segment {
  id: number
  org_id: number
  name: string
  type: 'rfm' | 'custom'
  filter: SegmentFilter
  auto_assign: boolean
  client_count?: number
  created_at: string
  updated_at: string
}

export interface SegmentFilter {
  gender?: string
  age_from?: number
  age_to?: number
  min_visits?: number
  max_visits?: number
  min_spend?: number
  max_spend?: number
  tags?: string[]
  rfm_category?: string
  bot_id?: number
  search?: string
  city?: string
  os?: string
  level_id?: number
  registered_from?: string
  registered_to?: string
  min_balance?: number
  max_balance?: number
  min_spent_points?: number
  max_spent_points?: number
}

export interface CreateSegmentRequest {
  name: string
  type: 'rfm' | 'custom'
  filter: SegmentFilter
  auto_assign: boolean
}

export interface UpdateSegmentRequest {
  name?: string
  filter?: SegmentFilter
  auto_assign?: boolean
}

export interface SegmentRule {
  id: number
  segment_id: number
  field: string
  operator: string
  value: unknown
  created_at: string
}

export interface CreateSegmentRuleRequest {
  field: string
  operator: string
  value: unknown
}

export interface ClientPrediction {
  id: number
  org_id: number
  client_id: number
  churn_risk: number
  upsell_score: number
  predicted_value: number
  factors: PredictionFactors
  computed_at: string
}

export interface PredictionFactors {
  days_since_last_visit: number
  visit_trend: 'increasing' | 'stable' | 'declining'
  spend_trend: 'increasing' | 'stable' | 'declining'
  avg_check: number
  total_orders: number
  loyalty_level: string
}

export interface PredictionSummary {
  high_churn_count: number
  avg_churn_risk: number
  high_upsell_count: number
  total_predicted: number
}

export const RULE_FIELDS: Record<string, string> = {
  days_since_visit: 'Дней с последнего визита',
  total_orders: 'Всего заказов',
  avg_check: 'Средний чек',
  loyalty_level: 'Уровень лояльности',
  total_spend: 'Общая сумма покупок',
  visit_frequency: 'Частота визитов',
}

export const RULE_OPERATORS: Record<string, string> = {
  eq: '=',
  neq: '≠',
  gt: '>',
  gte: '≥',
  lt: '<',
  lte: '≤',
  in: 'в списке',
  not_in: 'не в списке',
  between: 'между',
}
