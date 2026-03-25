package entity

import "time"

// RFMConfig — per-org configuration for RFM calculation.
type RFMConfig struct {
	ID               int        `db:"id"                json:"id"`
	OrgID            int        `db:"org_id"            json:"org_id"`
	PeriodDays       int        `db:"period_days"       json:"period_days"`
	RecalcInterval   string     `db:"recalc_interval"   json:"recalc_interval"`
	LastCalcAt       *time.Time `db:"last_calc_at"      json:"last_calc_at,omitempty"`
	ClientsProcessed int        `db:"clients_processed" json:"clients_processed"`
	CreatedAt        time.Time  `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"        json:"updated_at"`
}

type UpdateRFMConfigRequest struct {
	PeriodDays     *int    `json:"period_days,omitempty"`
	RecalcInterval *string `json:"recalc_interval,omitempty"`
}

// RFMHistory — snapshot of segment distribution for trend tracking.
type RFMHistory struct {
	ID           int       `db:"id"            json:"id"`
	OrgID        int       `db:"org_id"        json:"org_id"`
	Segment      string    `db:"segment"       json:"segment"`
	ClientCount  int       `db:"client_count"  json:"client_count"`
	CalculatedAt time.Time `db:"calculated_at" json:"calculated_at"`
}

// RFMDashboard — response for the RFM dashboard endpoint.
type RFMDashboard struct {
	Segments []RFMSegmentSummary `json:"segments"`
	Trends   []RFMHistory        `json:"trends"`
	Config   *RFMConfig          `json:"config,omitempty"`
}

type RFMSegmentSummary struct {
	Segment     string  `json:"segment"      db:"segment"`
	ClientCount int     `json:"client_count" db:"client_count"`
	Percentage  float64 `json:"percentage"`
}
