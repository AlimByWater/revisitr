package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type AudienceFilter struct {
	BotID     *int     `json:"bot_id,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	SegmentID *int     `json:"segment_id,omitempty"`
	LevelID   *int     `json:"level_id,omitempty"`
	ClientIDs []int    `json:"client_ids,omitempty"`
}

func (f *AudienceFilter) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, f)
	case string:
		return json.Unmarshal([]byte(v), f)
	case nil:
		*f = AudienceFilter{}
		return nil
	default:
		return fmt.Errorf("AudienceFilter.Scan: unsupported type %T", src)
	}
}

func (f AudienceFilter) Value() (driver.Value, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("AudienceFilter.Value: %w", err)
	}
	return b, nil
}

type CampaignStats struct {
	Total  int `json:"total"`
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
}

func (s *CampaignStats) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	case nil:
		*s = CampaignStats{}
		return nil
	default:
		return fmt.Errorf("CampaignStats.Scan: unsupported type %T", src)
	}
}

func (s CampaignStats) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("CampaignStats.Value: %w", err)
	}
	return b, nil
}

type CampaignButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type CampaignButtons []CampaignButton

func (b *CampaignButtons) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, b)
	case string:
		return json.Unmarshal([]byte(v), b)
	case nil:
		*b = CampaignButtons{}
		return nil
	default:
		return fmt.Errorf("CampaignButtons.Scan: unsupported type %T", src)
	}
}

func (b CampaignButtons) Value() (driver.Value, error) {
	if b == nil {
		return []byte("[]"), nil
	}
	data, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("CampaignButtons.Value: %w", err)
	}
	return data, nil
}

type CampaignClick struct {
	ID         int       `db:"id" json:"id"`
	CampaignID int       `db:"campaign_id" json:"campaign_id"`
	ClientID   int       `db:"client_id" json:"client_id"`
	ButtonIdx  *int      `db:"button_idx" json:"button_idx,omitempty"`
	URL        *string   `db:"url" json:"url,omitempty"`
	ClickedAt  time.Time `db:"clicked_at" json:"clicked_at"`
}

type CampaignAnalyticsDetail struct {
	Total     int     `json:"total"`
	Sent      int     `json:"sent"`
	Failed    int     `json:"failed"`
	Clicked   int     `json:"clicked"`
	ClickRate float64 `json:"click_rate"`
}

type Campaign struct {
	ID             int             `db:"id" json:"id"`
	OrgID          int             `db:"org_id" json:"org_id"`
	BotID          int             `db:"bot_id" json:"bot_id"`
	Name           string          `db:"name" json:"name"`
	Type           string          `db:"type" json:"type"`
	Status         string          `db:"status" json:"status"`
	AudienceFilter AudienceFilter  `db:"audience_filter" json:"audience_filter"`
	Message        string          `db:"message" json:"message"`
	MediaURL       *string         `db:"media_url" json:"media_url,omitempty"`
	Buttons        CampaignButtons `db:"buttons" json:"buttons"`
	TrackingMode   string          `db:"tracking_mode" json:"tracking_mode"`
	ScheduledAt    *time.Time      `db:"scheduled_at" json:"scheduled_at,omitempty"`
	SentAt         *time.Time      `db:"sent_at" json:"sent_at,omitempty"`
	Stats          CampaignStats   `db:"stats" json:"stats"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
}

type CampaignMessage struct {
	ID           int        `db:"id" json:"id"`
	CampaignID   int        `db:"campaign_id" json:"campaign_id"`
	ClientID     int        `db:"client_id" json:"client_id"`
	TelegramID   int64      `db:"telegram_id" json:"telegram_id"`
	Status       string     `db:"status" json:"status"`
	ErrorMessage *string    `db:"error_message" json:"error_message,omitempty"`
	SentAt       *time.Time `db:"sent_at" json:"sent_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type CreateCampaignRequest struct {
	BotID          int            `json:"bot_id" binding:"required"`
	Name           string         `json:"name" binding:"required"`
	Message        string         `json:"message" binding:"required"`
	AudienceFilter AudienceFilter `json:"audience_filter"`
	MediaURL       *string        `json:"media_url,omitempty"`
	ScheduledAt    *time.Time     `json:"scheduled_at,omitempty"`
}

type UpdateCampaignRequest struct {
	Name           *string         `json:"name,omitempty"`
	Message        *string         `json:"message,omitempty"`
	AudienceFilter *AudienceFilter `json:"audience_filter,omitempty"`
	MediaURL       *string         `json:"media_url,omitempty"`
	ScheduledAt    *time.Time      `json:"scheduled_at,omitempty"`
}
