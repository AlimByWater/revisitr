export interface AnalyticsFilter {
  period?: string
  from?: string
  to?: string
  bot_id?: number
  segment_id?: number
}

// Sales
export interface SalesMetrics {
  transaction_count: number
  unique_clients: number
  total_amount: number
  avg_amount: number
  buy_frequency: number
}

export interface ChartPoint {
  day: string
  value: number
}

export interface LoyaltyComparison {
  participants_avg_amount: number
  non_participants_avg_amount: number
}

export interface SalesAnalytics {
  metrics: SalesMetrics
  charts: Record<string, ChartPoint[]>
  comparison?: LoyaltyComparison
}

// Loyalty
export interface PieSlice {
  label: string
  value: number
  percent: number
}

export interface FunnelStep {
  step: string
  count: number
  percent: number
}

export interface ClientDemographics {
  by_gender: PieSlice[]
  by_age_group: PieSlice[]
  by_os: PieSlice[]
  loyalty_percent: number
}

export interface LoyaltyAnalytics {
  new_clients: number
  active_clients: number
  bonus_earned: number
  bonus_spent: number
  demographics: ClientDemographics
  bot_funnel: FunnelStep[]
}

// Campaigns
export interface CampaignStat {
  campaign_id: number
  campaign_name: string
  sent: number
  open_rate: number
  conversions: number
}

export interface CampaignAnalytics {
  total_sent: number
  total_opened: number
  open_rate: number
  conversions: number
  conv_rate: number
  by_campaign: CampaignStat[]
}
