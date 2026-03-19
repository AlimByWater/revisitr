import { useApiQuery } from '@/lib/swr'
import { analyticsApi } from './api'
import type { AnalyticsFilter } from './types'

export function useSalesAnalyticsQuery(filter: AnalyticsFilter) {
  return useApiQuery(
    `analytics-sales-${JSON.stringify(filter)}`,
    () => analyticsApi.getSales(filter),
  )
}

export function useLoyaltyAnalyticsQuery(filter: AnalyticsFilter) {
  return useApiQuery(
    `analytics-loyalty-${JSON.stringify(filter)}`,
    () => analyticsApi.getLoyalty(filter),
  )
}

export function useCampaignAnalyticsQuery(filter: AnalyticsFilter) {
  return useApiQuery(
    `analytics-campaigns-${JSON.stringify(filter)}`,
    () => analyticsApi.getCampaigns(filter),
  )
}
