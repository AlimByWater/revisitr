package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type TriggerConfig struct {
	Days      *int `json:"days,omitempty"`
	Count     *int `json:"count,omitempty"`
	Threshold *int `json:"threshold,omitempty"`
}

func (c *TriggerConfig) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = TriggerConfig{}
		return nil
	default:
		return fmt.Errorf("TriggerConfig.Scan: unsupported type %T", src)
	}
}

func (c TriggerConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("TriggerConfig.Value: %w", err)
	}
	return b, nil
}

type AutoScenario struct {
	ID            int           `db:"id" json:"id"`
	OrgID         int           `db:"org_id" json:"org_id"`
	BotID         int           `db:"bot_id" json:"bot_id"`
	Name          string        `db:"name" json:"name"`
	TriggerType   string        `db:"trigger_type" json:"trigger_type"`
	TriggerConfig TriggerConfig `db:"trigger_config" json:"trigger_config"`
	Message       string        `db:"message" json:"message"`
	IsActive      bool          `db:"is_active" json:"is_active"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at" json:"updated_at"`
}

type CreateScenarioRequest struct {
	BotID         int           `json:"bot_id" binding:"required"`
	Name          string        `json:"name" binding:"required"`
	TriggerType   string        `json:"trigger_type" binding:"required,oneof=inactive_days visit_count bonus_threshold level_up birthday"`
	TriggerConfig TriggerConfig `json:"trigger_config"`
	Message       string        `json:"message" binding:"required"`
}

type UpdateScenarioRequest struct {
	Name          *string        `json:"name,omitempty"`
	TriggerConfig *TriggerConfig `json:"trigger_config,omitempty"`
	Message       *string        `json:"message,omitempty"`
	IsActive      *bool          `json:"is_active,omitempty"`
}
