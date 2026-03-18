import { api } from '@/lib/api'
import type {
  DashboardWidgets,
  DashboardCharts,
  DashboardFilter,
} from './types'

export const dashboardApi = {
  getWidgets: async (filter: DashboardFilter): Promise<DashboardWidgets> => {
    const response = await api.get<DashboardWidgets>('/dashboard/widgets', {
      params: filter,
    })
    return response.data
  },

  getCharts: async (filter: DashboardFilter): Promise<DashboardCharts> => {
    const response = await api.get<DashboardCharts>('/dashboard/charts', {
      params: filter,
    })
    return response.data
  },
}
