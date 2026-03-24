import { api } from '@/lib/api'
import type {
  DashboardWidgets,
  DashboardCharts,
  DashboardFilter,
  DashboardSalesData,
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

  getSalesData: async (from: string, to: string): Promise<DashboardSalesData> => {
    const response = await api.get<DashboardSalesData>('/dashboard/sales', {
      params: { from, to },
    })
    return response.data
  },
}
