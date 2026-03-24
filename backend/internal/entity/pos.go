package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type DaySchedule struct {
	Open   string `json:"open"`
	Close  string `json:"close"`
	Closed bool   `json:"closed,omitempty"`
}

type Schedule map[string]DaySchedule

func (s *Schedule) Scan(src interface{}) error {
	if src == nil {
		*s = make(Schedule)
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Schedule.Scan: expected []byte, got %T", src)
	}
	return json.Unmarshal(b, s)
}

func (s Schedule) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type POSLocation struct {
	ID        int       `db:"id" json:"id"`
	OrgID     int       `db:"org_id" json:"org_id"`
	BotID     *int      `db:"bot_id" json:"bot_id"`
	Name      string    `db:"name" json:"name"`
	Address   string    `db:"address" json:"address,omitempty"`
	Phone     string    `db:"phone" json:"phone,omitempty"`
	Schedule  Schedule  `db:"schedule" json:"schedule"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreatePOSRequest struct {
	Name     string   `json:"name" binding:"required"`
	Address  string   `json:"address"`
	Phone    string   `json:"phone"`
	Schedule Schedule `json:"schedule"`
	BotID    *int     `json:"bot_id"`
}

type UpdatePOSRequest struct {
	Name     *string   `json:"name,omitempty"`
	Address  *string   `json:"address,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
	Schedule *Schedule `json:"schedule,omitempty"`
	IsActive *bool     `json:"is_active,omitempty"`
	BotID    *int      `json:"bot_id,omitempty"`
}
