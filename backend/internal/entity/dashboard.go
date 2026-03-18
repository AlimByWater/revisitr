package entity

import "time"

type DashboardMetric struct {
	Value    float64 `json:"value"`
	Previous float64 `json:"previous"`
	Trend    float64 `json:"trend"` // percentage change
}

type DashboardWidgets struct {
	Revenue       DashboardMetric `json:"revenue"`
	AvgCheck      DashboardMetric `json:"avg_check"`
	NewClients    DashboardMetric `json:"new_clients"`
	ActiveClients DashboardMetric `json:"active_clients"`
}

type DashboardFilter struct {
	Period string     `form:"period"` // 7d, 30d, 90d
	From   *time.Time `form:"from"`
	To     *time.Time `form:"to"`
	BotID  *int       `form:"bot_id"`
}

type ChartPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type DashboardCharts struct {
	Revenue    []ChartPoint `json:"revenue"`
	NewClients []ChartPoint `json:"new_clients"`
}
