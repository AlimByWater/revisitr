export interface DashboardMetric {
  value: number
  previous: number
  trend: number
}

export interface DashboardWidgets {
  revenue: DashboardMetric
  avg_check: DashboardMetric
  new_clients: DashboardMetric
  active_clients: DashboardMetric
}

export interface DashboardFilter {
  period?: string
  from?: string
  to?: string
  bot_id?: number
}

export interface ChartPoint {
  date: string
  value: number
}

export interface DashboardCharts {
  revenue: ChartPoint[]
  new_clients: ChartPoint[]
}

export interface DashboardSalesData {
  revenue: number
  avg_check: number
  tx_count: number
  loyalty_avg: number
  non_loyalty_avg: number
}
