package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// RFM segment categories
const (
	RFMChampions   = "champions"
	RFMLoyal       = "loyal"
	RFMAtRisk      = "at_risk"
	RFMCantLose    = "cant_lose"
	RFMHibernating = "hibernating"
	RFMLost        = "lost"
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
