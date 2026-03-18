import { useQuery } from '@tanstack/react-query'
import { dashboardApi } from './api'
import type { DashboardFilter } from './types'

export function useDashboardWidgetsQuery(filter: DashboardFilter) {
  return useQuery({
    queryKey: ['dashboard', 'widgets', filter],
    queryFn: () => dashboardApi.getWidgets(filter),
  })
}

export function useDashboardChartsQuery(filter: DashboardFilter) {
  return useQuery({
    queryKey: ['dashboard', 'charts', filter],
    queryFn: () => dashboardApi.getCharts(filter),
  })
}
