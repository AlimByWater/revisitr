export interface AudienceFilter {
  bot_id?: number
  tags?: string[]
  segment_id?: number
  level_id?: number
  client_ids?: number[]
}

export interface CampaignStats {
  total: number
  sent: number
  failed: number
}

export interface CampaignButton {
  text: string
  url: string
}

export interface Campaign {
  id: number
  org_id: number
  bot_id: number
  name: string
  type: 'manual' | 'auto'
  status: 'draft' | 'scheduled' | 'sending' | 'sent' | 'completed' | 'failed'
  audience_filter: AudienceFilter
  message: string
  media_url?: string
  buttons: CampaignButton[]
  tracking_mode: 'utm' | 'buttons' | 'both' | 'none'
  scheduled_at?: string
  sent_at?: string
  stats: CampaignStats
  created_at: string
  updated_at: string
}

export interface CampaignClick {
  id: number
  campaign_id: number
  client_id: number
  button_idx?: number
  url?: string
  clicked_at: string
}

export interface CampaignAnalyticsDetail {
  total: number
  sent: number
  failed: number
  clicked: number
  click_rate: number
}

export interface ActionDef {
  type: 'bonus' | 'campaign' | 'promo_code' | 'level_change'
  amount?: number
  template_id?: number
  template?: string
  discount?: number
  level_id?: number
}

export interface ActionTiming {
  days_before?: number
  days_after?: number
  month?: number
  day?: number
}

export interface ActionCondition {
  type?: string
  min_amount?: number
  level_id?: number
  segment_id?: number
}

export interface AutoScenario {
  id: number
  org_id: number
  bot_id: number
  name: string
  trigger_type:
    | 'inactive_days'
    | 'visit_count'
    | 'bonus_threshold'
    | 'level_up'
    | 'birthday'
    | 'holiday'
    | 'registration'
    | 'level_change'
  trigger_config: { days?: number; count?: number; threshold?: number; month?: number; day?: number }
  message: string
  actions: ActionDef[]
  timing: ActionTiming
  conditions: ActionCondition
  is_template: boolean
  template_key?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface AutoActionLog {
  id: number
  scenario_id: number
  client_id: number
  action_type: string
  action_data: Record<string, unknown>
  result: 'success' | 'failed' | 'skipped'
  error_msg?: string
  executed_at: string
}

export interface CreateCampaignRequest {
  bot_id: number
  name: string
  message: string
  audience_filter: AudienceFilter
  media_url?: string
  scheduled_at?: string
}

export interface UpdateCampaignRequest {
  name?: string
  message?: string
  audience_filter?: AudienceFilter
  media_url?: string
  scheduled_at?: string
}

export interface CreateScenarioRequest {
  bot_id: number
  name: string
  trigger_type: string
  trigger_config: { days?: number; count?: number; threshold?: number; month?: number; day?: number }
  message?: string
  actions?: ActionDef[]
  timing?: ActionTiming
  conditions?: ActionCondition
}

export interface UpdateScenarioRequest {
  name?: string
  trigger_config?: { days?: number; count?: number; threshold?: number; month?: number; day?: number }
  message?: string
  is_active?: boolean
  actions?: ActionDef[]
  timing?: ActionTiming
  conditions?: ActionCondition
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
}
