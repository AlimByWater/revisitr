// ── RFM Segment constants (v2: 7 segments) ──────────────────────────────

export const RFM_SEGMENTS = [
  'new',
  'promising',
  'regular',
  'vip',
  'rare_valuable',
  'churn_risk',
  'lost',
] as const

export type RFMSegmentKey = (typeof RFM_SEGMENTS)[number]

export const RFM_SEGMENT_LABELS: Record<string, string> = {
  new: 'Новые',
  promising: 'Перспективные',
  regular: 'Регулярные',
  vip: 'VIP / Ядро',
  rare_valuable: 'Редкие, но ценные',
  churn_risk: 'На грани оттока',
  lost: 'Потерянные',
}

export const RFM_SEGMENT_COLORS: Record<
  string,
  { bg: string; text: string; accent: string; icon: string }
> = {
  new: { bg: 'bg-cyan-50', text: 'text-cyan-700', accent: 'bg-cyan-500', icon: '🟢' },
  promising: { bg: 'bg-blue-50', text: 'text-blue-700', accent: 'bg-blue-500', icon: '🔵' },
  regular: { bg: 'bg-violet-50', text: 'text-violet-700', accent: 'bg-violet-500', icon: '🟣' },
  vip: { bg: 'bg-amber-50', text: 'text-amber-700', accent: 'bg-amber-500', icon: '⭐' },
  rare_valuable: { bg: 'bg-purple-50', text: 'text-purple-700', accent: 'bg-purple-500', icon: '💎' },
  churn_risk: { bg: 'bg-yellow-50', text: 'text-yellow-700', accent: 'bg-yellow-500', icon: '🟡' },
  lost: { bg: 'bg-red-50', text: 'text-red-700', accent: 'bg-red-500', icon: '🔴' },
}

// ── RFM Config ──────────────────────────────────────────────────────────

export interface RFMConfig {
  id: number
  org_id: number
  period_days: number
  recalc_interval: string
  last_calc_at?: string
  clients_processed: number
  created_at: string
  updated_at: string
  active_template_type: string
  active_template_key: string
  custom_template_name?: string
  custom_r_thresholds?: number[]
  custom_f_thresholds?: number[]
}

export interface UpdateRFMConfigRequest {
  period_days?: number
  recalc_interval?: string
}

// ── RFM Templates ───────────────────────────────────────────────────────

export interface RFMTemplate {
  key: string
  name: string
  description: string
  r_thresholds: [number, number, number, number]
  f_thresholds: [number, number, number, number]
}

export interface SetTemplateRequest {
  template_type: 'standard' | 'custom'
  template_key?: string
  custom_name?: string
  r_thresholds?: [number, number, number, number]
  f_thresholds?: [number, number, number, number]
}

export interface ActiveTemplateResponse {
  active_template_type: string
  active_template_key: string
  template: RFMTemplate
}

// ── RFM Dashboard ───────────────────────────────────────────────────────

export interface RFMSegmentSummary {
  segment: string
  client_count: number
  percentage: number
  avg_check: number
  total_check: number
}

export interface RFMHistory {
  id: number
  org_id: number
  segment: string
  client_count: number
  calculated_at: string
}

export interface RFMDashboard {
  segments: RFMSegmentSummary[]
  trends: RFMHistory[]
  config?: RFMConfig
}

// ── Segment Detail ──────────────────────────────────────────────────────

export interface SegmentClientRow {
  id: number
  first_name: string
  last_name: string
  phone: string
  r_score: number | null
  f_score: number | null
  m_score: number | null
  recency_days: number | null
  frequency_count: number | null
  monetary_sum: number | null
  last_visit_date: string | null
  total_visits_lifetime: number
}

export interface SegmentClientsResponse {
  segment: string
  segment_name: string
  total: number
  page: number
  per_page: number
  clients: SegmentClientRow[]
}

// ── Onboarding ──────────────────────────────────────────────────────────

export interface OnboardingAnswer {
  id: number
  text: string
}

export interface OnboardingQuestion {
  id: number
  text: string
  answers: OnboardingAnswer[]
}

export interface TemplateRecommendation {
  recommended: RFMTemplate
  alternative?: RFMTemplate
  all_scores: Record<string, number>
}
