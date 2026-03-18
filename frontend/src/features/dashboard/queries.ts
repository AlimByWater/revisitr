import { useApiQuery } from '../../lib/swr'
import { dashboardApi } from './api'
import type { DashboardFilter } from './types'

export function useDashboardWidgetsQuery(filter: DashboardFilter) {
  return useApiQuery(`dashboard-widgets-${JSON.stringify(filter)}`, () =>
    dashboardApi.getWidgets(filter),
  )
}

export function useDashboardChartsQuery(filter: DashboardFilter) {
  return useApiQuery(`dashboard-charts-${JSON.stringify(filter)}`, () =>
    dashboardApi.getCharts(filter),
  )
}
