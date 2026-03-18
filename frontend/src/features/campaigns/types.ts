export interface AudienceFilter {
  bot_id?: number
  tags?: string[]
}

export interface CampaignStats {
  total: number
  sent: number
  failed: number
}

export interface Campaign {
  id: number
  org_id: number
  bot_id: number
  name: string
  type: 'manual' | 'auto'
  status: 'draft' | 'scheduled' | 'sending' | 'sent' | 'failed'
  audience_filter: AudienceFilter
  message: string
  media_url?: string
  scheduled_at?: string
  sent_at?: string
  stats: CampaignStats
  created_at: string
  updated_at: string
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
  trigger_config: { days?: number; count?: number; threshold?: number }
  message: string
  is_active: boolean
  created_at: string
  updated_at: string
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
  trigger_config: { days?: number; count?: number; threshold?: number }
  message: string
}

export interface UpdateScenarioRequest {
  name?: string
  trigger_config?: { days?: number; count?: number; threshold?: number }
  message?: string
  is_active?: boolean
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
}
