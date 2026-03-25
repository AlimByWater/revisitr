export interface TariffFeatures {
  loyalty: boolean
  campaigns: boolean
  promotions: boolean
  integrations: boolean
  analytics: boolean
  rfm: boolean
  advanced_campaigns: boolean
}

export interface TariffLimits {
  max_clients: number
  max_bots: number
  max_campaigns_per_month: number
  max_pos: number
}

export interface Tariff {
  id: number
  name: string
  slug: string
  price: number // kopeks
  currency: string
  interval: string
  features: TariffFeatures
  limits: TariffLimits
  active: boolean
  sort_order: number
  created_at: string
}

export interface Subscription {
  id: number
  org_id: number
  tariff_id: number
  status: 'trialing' | 'active' | 'past_due' | 'canceled' | 'expired'
  current_period_start: string
  current_period_end: string
  canceled_at?: string
  created_at: string
  updated_at: string
}

export interface SubscriptionWithTariff extends Subscription {
  tariff_name: string
  tariff_slug: string
  tariff_price: number
  tariff_features: TariffFeatures
  tariff_limits: TariffLimits
}

export interface Invoice {
  id: number
  org_id: number
  subscription_id?: number
  amount: number
  currency: string
  status: 'pending' | 'paid' | 'failed' | 'refunded'
  due_date: string
  paid_at?: string
  created_at: string
}

export interface Payment {
  id: number
  invoice_id: number
  org_id: number
  amount: number
  currency: string
  provider: string
  provider_payment_id?: string
  status: 'pending' | 'succeeded' | 'failed' | 'refunded'
  created_at: string
}

export interface CreateSubscriptionRequest {
  tariff_slug: string
}

export interface ChangeSubscriptionRequest {
  tariff_slug: string
}
