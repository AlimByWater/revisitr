export interface RFMConfig {
  id: number
  org_id: number
  recency_days: number
  min_transactions: number
  period_days: number
  recalc_interval: string
  last_calc_at?: string
  clients_processed: number
  created_at: string
  updated_at: string
}

export interface RFMSegmentSummary {
  segment: string
  client_count: number
  percentage: number
}

export interface RFMHistory {
  id: number
  org_id: number
  total_clients: number
  distribution: Record<string, number>
  calculated_at: string
}

export interface RFMDashboard {
  segments: RFMSegmentSummary[]
  trends: RFMHistory[]
  config?: RFMConfig
}

export interface UpdateRFMConfigRequest {
  recency_days?: number
  min_transactions?: number
  period_days?: number
  recalc_interval?: string
}

export const RFM_SEGMENT_LABELS: Record<string, string> = {
  champions: 'Чемпионы',
  loyal: 'Лояльные',
  potential_loyalist: 'Потенциально лояльные',
  new_customers: 'Новые клиенты',
  at_risk: 'В зоне риска',
  cant_lose: 'Нельзя потерять',
  hibernating: 'Спящие',
  lost: 'Потерянные',
}

export const RFM_SEGMENT_COLORS: Record<string, { bg: string; text: string; accent: string }> = {
  champions: { bg: 'bg-green-50', text: 'text-green-700', accent: 'bg-green-500' },
  loyal: { bg: 'bg-blue-50', text: 'text-blue-700', accent: 'bg-blue-500' },
  potential_loyalist: { bg: 'bg-violet-50', text: 'text-violet-700', accent: 'bg-violet-500' },
  new_customers: { bg: 'bg-cyan-50', text: 'text-cyan-700', accent: 'bg-cyan-500' },
  at_risk: { bg: 'bg-amber-50', text: 'text-amber-700', accent: 'bg-amber-500' },
  cant_lose: { bg: 'bg-red-50', text: 'text-red-700', accent: 'bg-red-500' },
  hibernating: { bg: 'bg-neutral-100', text: 'text-neutral-600', accent: 'bg-neutral-400' },
  lost: { bg: 'bg-neutral-100', text: 'text-neutral-500', accent: 'bg-neutral-600' },
}
