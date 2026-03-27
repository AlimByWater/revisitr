package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Deprecated: v1 RFM segment categories. Use RFMSegment* constants from rfm.go instead.
const (
	RFMChampions   = "champions"   // Deprecated: use RFMSegmentVIP
	RFMLoyal       = "loyal"       // Deprecated: use RFMSegmentRegular
	RFMAtRisk      = "at_risk"     // Deprecated: use RFMSegmentChurnRisk
	RFMCantLose    = "cant_lose"   // Deprecated: use RFMSegmentRareValue
	RFMHibernating = "hibernating" // Deprecated: use RFMSegmentPromising
	RFMLost        = "lost"        // Deprecated: use RFMSegmentLost
)

type Segment struct {
	ID          int           `json:"id"           db:"id"`
	OrgID       int           `json:"org_id"       db:"org_id"`
	Name        string        `json:"name"         db:"name"`
	Type        string        `json:"type"         db:"type"` // "rfm"|"custom"
	Filter      SegmentFilter `json:"filter"       db:"filter"`
	AutoAssign  bool          `json:"auto_assign"  db:"auto_assign"`
	ClientCount *int          `json:"client_count,omitempty" db:"-"`
	CreatedAt   time.Time     `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"   db:"updated_at"`
}

type SegmentFilter struct {
	Gender      *string  `json:"gender,omitempty"`
	AgeFrom     *int     `json:"age_from,omitempty"`
	AgeTo       *int     `json:"age_to,omitempty"`
	MinVisits   *int     `json:"min_visits,omitempty"`
	MaxVisits   *int     `json:"max_visits,omitempty"`
	MinSpend    *float64 `json:"min_spend,omitempty"`
	MaxSpend    *float64 `json:"max_spend,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	RFMCategory *string  `json:"rfm_category,omitempty"`
}

func (f SegmentFilter) Value() (driver.Value, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("SegmentFilter.Value: %w", err)
	}
	return b, nil
}

func (f *SegmentFilter) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, f)
	case string:
		return json.Unmarshal([]byte(v), f)
	case nil:
		*f = SegmentFilter{}
		return nil
	default:
		return fmt.Errorf("SegmentFilter.Scan: unsupported type %T", src)
	}
}

type CreateSegmentRequest struct {
	Name       string        `json:"name"        binding:"required"`
	Type       string        `json:"type"        binding:"required,oneof=rfm custom"`
	Filter     SegmentFilter `json:"filter"`
	AutoAssign bool          `json:"auto_assign"`
}

type UpdateSegmentRequest struct {
	Name       *string        `json:"name"`
	Filter     *SegmentFilter `json:"filter"`
	AutoAssign *bool          `json:"auto_assign"`
}

// ── Advanced Segmentation ────────────────────────────────────────────────────

// SegmentRule represents a single behavioral rule within a segment.
type SegmentRule struct {
	ID        int             `db:"id" json:"id"`
	SegmentID int             `db:"segment_id" json:"segment_id"`
	Field     string          `db:"field" json:"field"`       // days_since_visit, total_orders, avg_check, loyalty_level
	Operator  string          `db:"operator" json:"operator"` // eq, neq, gt, gte, lt, lte, in, not_in, between
	Value     json.RawMessage `db:"value" json:"value"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

type CreateSegmentRuleRequest struct {
	Field    string          `json:"field" binding:"required"`
	Operator string          `json:"operator" binding:"required,oneof=eq neq gt gte lt lte in not_in between"`
	Value    json.RawMessage `json:"value" binding:"required"`
}

// ClientPrediction stores ML/heuristic predictions for a client.
type ClientPrediction struct {
	ID             int              `db:"id" json:"id"`
	OrgID          int              `db:"org_id" json:"org_id"`
	ClientID       int              `db:"client_id" json:"client_id"`
	ChurnRisk      float32          `db:"churn_risk" json:"churn_risk"`
	UpsellScore    float32          `db:"upsell_score" json:"upsell_score"`
	PredictedValue float32          `db:"predicted_value" json:"predicted_value"`
	Factors        PredictionFactors `db:"factors" json:"factors"`
	ComputedAt     time.Time        `db:"computed_at" json:"computed_at"`
}

type PredictionFactors struct {
	DaysSinceLastVisit int     `json:"days_since_last_visit"`
	VisitTrend         string  `json:"visit_trend"`   // increasing, stable, declining
	SpendTrend         string  `json:"spend_trend"`   // increasing, stable, declining
	AvgCheck           float64 `json:"avg_check"`
	TotalOrders        int     `json:"total_orders"`
	LoyaltyLevel       string  `json:"loyalty_level"`
}

func (f PredictionFactors) Value() (driver.Value, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("PredictionFactors.Value: %w", err)
	}
	return b, nil
}

func (f *PredictionFactors) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, f)
	case string:
		return json.Unmarshal([]byte(v), f)
	case nil:
		*f = PredictionFactors{}
		return nil
	default:
		return fmt.Errorf("PredictionFactors.Scan: unsupported type %T", src)
	}
}

type PredictionSummary struct {
	HighChurnCount  int     `json:"high_churn_count" db:"high_churn_count"`
	AvgChurnRisk    float32 `json:"avg_churn_risk" db:"avg_churn_risk"`
	HighUpsellCount int     `json:"high_upsell_count" db:"high_upsell_count"`
	TotalPredicted  int     `json:"total_predicted" db:"total_predicted"`
}
