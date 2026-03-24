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
	Month     *int `json:"month,omitempty"`
	Day       *int `json:"day,omitempty"`
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

type ActionTiming struct {
	DaysBefore *int `json:"days_before,omitempty"`
	DaysAfter  *int `json:"days_after,omitempty"`
	Month      *int `json:"month,omitempty"`
	Day        *int `json:"day,omitempty"`
}

func (t *ActionTiming) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, t)
	case string:
		return json.Unmarshal([]byte(v), t)
	case nil:
		*t = ActionTiming{}
		return nil
	default:
		return fmt.Errorf("ActionTiming.Scan: unsupported type %T", src)
	}
}

func (t ActionTiming) Value() (driver.Value, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("ActionTiming.Value: %w", err)
	}
	return b, nil
}

type ActionCondition struct {
	Type      string   `json:"type,omitempty"`
	MinAmount *float64 `json:"min_amount,omitempty"`
	LevelID   *int     `json:"level_id,omitempty"`
	SegmentID *int     `json:"segment_id,omitempty"`
}

func (c *ActionCondition) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = ActionCondition{}
		return nil
	default:
		return fmt.Errorf("ActionCondition.Scan: unsupported type %T", src)
	}
}

func (c ActionCondition) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("ActionCondition.Value: %w", err)
	}
	return b, nil
}

type ActionDef struct {
	Type       string  `json:"type"`
	Amount     *int    `json:"amount,omitempty"`
	TemplateID *int    `json:"template_id,omitempty"`
	Template   *string `json:"template,omitempty"`
	Discount   *int    `json:"discount,omitempty"`
	LevelID    *int    `json:"level_id,omitempty"`
}

type ActionDefs []ActionDef

func (d *ActionDefs) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, d)
	case string:
		return json.Unmarshal([]byte(v), d)
	case nil:
		*d = ActionDefs{}
		return nil
	default:
		return fmt.Errorf("ActionDefs.Scan: unsupported type %T", src)
	}
}

func (d ActionDefs) Value() (driver.Value, error) {
	if d == nil {
		d = ActionDefs{}
	}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("ActionDefs.Value: %w", err)
	}
	return b, nil
}

type AutoActionLog struct {
	ID         int             `db:"id" json:"id"`
	ScenarioID int             `db:"scenario_id" json:"scenario_id"`
	ClientID   int             `db:"client_id" json:"client_id"`
	ActionType string          `db:"action_type" json:"action_type"`
	ActionData json.RawMessage `db:"action_data" json:"action_data"`
	Result     string          `db:"result" json:"result"`
	ErrorMsg   *string         `db:"error_msg" json:"error_msg,omitempty"`
	ExecutedAt time.Time       `db:"executed_at" json:"executed_at"`
}

type AutoScenario struct {
	ID            int             `db:"id" json:"id"`
	OrgID         int             `db:"org_id" json:"org_id"`
	BotID         int             `db:"bot_id" json:"bot_id"`
	Name          string          `db:"name" json:"name"`
	TriggerType   string          `db:"trigger_type" json:"trigger_type"`
	TriggerConfig TriggerConfig   `db:"trigger_config" json:"trigger_config"`
	Message       string          `db:"message" json:"message"`
	Actions       ActionDefs      `db:"actions" json:"actions"`
	Timing        ActionTiming    `db:"timing" json:"timing"`
	Conditions    ActionCondition `db:"conditions" json:"conditions"`
	IsTemplate    bool            `db:"is_template" json:"is_template"`
	TemplateKey   *string         `db:"template_key" json:"template_key,omitempty"`
	IsActive      bool            `db:"is_active" json:"is_active"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at" json:"updated_at"`
}

type CreateScenarioRequest struct {
	BotID         int             `json:"bot_id" binding:"required"`
	Name          string          `json:"name" binding:"required"`
	TriggerType   string          `json:"trigger_type" binding:"required,oneof=inactive_days visit_count bonus_threshold level_up birthday holiday"`
	TriggerConfig TriggerConfig   `json:"trigger_config"`
	Message       string          `json:"message"`
	Actions       ActionDefs      `json:"actions,omitempty"`
	Timing        ActionTiming    `json:"timing,omitempty"`
	Conditions    ActionCondition `json:"conditions,omitempty"`
}

type UpdateScenarioRequest struct {
	Name          *string          `json:"name,omitempty"`
	TriggerConfig *TriggerConfig   `json:"trigger_config,omitempty"`
	Message       *string          `json:"message,omitempty"`
	IsActive      *bool            `json:"is_active,omitempty"`
	Actions       *ActionDefs      `json:"actions,omitempty"`
	Timing        *ActionTiming    `json:"timing,omitempty"`
	Conditions    *ActionCondition `json:"conditions,omitempty"`
}
