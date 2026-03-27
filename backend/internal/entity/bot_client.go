package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Tags []string

func (t *Tags) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	case nil:
		*t = Tags{}
		return nil
	default:
		return fmt.Errorf("Tags.Scan: unsupported type %T", src)
	}
}

func (t Tags) Value() (driver.Value, error) {
	if t == nil {
		t = Tags{}
	}
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("Tags.Value: %w", err)
	}
	return b, nil
}

type BotClient struct {
	ID           int        `db:"id" json:"id"`
	BotID        int        `db:"bot_id" json:"bot_id"`
	TelegramID   int64      `db:"telegram_id" json:"telegram_id"`
	Username     string     `db:"username" json:"username,omitempty"`
	FirstName    string     `db:"first_name" json:"first_name"`
	LastName     string     `db:"last_name" json:"last_name,omitempty"`
	Phone            string     `db:"phone" json:"phone"`
	PhoneNormalized  *string    `db:"phone_normalized" json:"phone_normalized,omitempty"`
	QRCode           *string    `db:"qr_code" json:"qr_code,omitempty"`
	Gender           *string    `db:"gender" json:"gender,omitempty"`
	BirthDate    *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	City         *string    `db:"city" json:"city,omitempty"`
	OS           *string    `db:"os" json:"os,omitempty"`
	Tags         Tags            `db:"tags" json:"tags"`
	Data         json.RawMessage `db:"data"          json:"-"`
	RegisteredAt time.Time       `db:"registered_at" json:"registered_at"`
	// RFM fields
	RFMSegment   *string    `db:"rfm_segment"    json:"rfm_segment,omitempty"`
	RFMUpdatedAt *time.Time `db:"rfm_updated_at" json:"rfm_updated_at,omitempty"`

	// RFM scores and raw metrics (migration 00030)
	RScore           *int       `db:"r_score"              json:"r_score,omitempty"`
	FScore           *int       `db:"f_score"              json:"f_score,omitempty"`
	MScore           *int       `db:"m_score"              json:"m_score,omitempty"`
	RecencyDays      *int       `db:"recency_days"         json:"recency_days,omitempty"`
	FrequencyCount   *int       `db:"frequency_count"      json:"frequency_count,omitempty"`
	MonetarySum      *float64   `db:"monetary_sum"         json:"monetary_sum,omitempty"`
	TotalVisitsLife  int        `db:"total_visits_lifetime" json:"total_visits_lifetime"`
	LastVisitDate    *time.Time `db:"last_visit_date"      json:"last_visit_date,omitempty"`
}
