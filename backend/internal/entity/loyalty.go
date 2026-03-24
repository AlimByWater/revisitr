package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type ProgramConfig struct {
	WelcomeBonus int    `json:"welcome_bonus"`
	CurrencyName string `json:"currency_name"`
}

func (c *ProgramConfig) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	case nil:
		*c = ProgramConfig{}
		return nil
	default:
		return fmt.Errorf("ProgramConfig.Scan: unsupported type %T", src)
	}
}

func (c ProgramConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("ProgramConfig.Value: %w", err)
	}
	return b, nil
}

type LoyaltyProgram struct {
	ID        int           `db:"id" json:"id"`
	OrgID     int           `db:"org_id" json:"org_id"`
	Name      string        `db:"name" json:"name"`
	Type      string        `db:"type" json:"type"`
	Config    ProgramConfig `db:"config" json:"config"`
	IsActive  bool          `db:"is_active" json:"is_active"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt time.Time     `db:"updated_at" json:"updated_at"`
	Levels    []LoyaltyLevel `db:"-" json:"levels,omitempty"`
}

type LoyaltyLevel struct {
	ID            int     `db:"id" json:"id"`
	ProgramID     int     `db:"program_id" json:"program_id"`
	Name          string  `db:"name" json:"name"`
	Threshold     int     `db:"threshold" json:"threshold"`
	RewardPercent float64 `db:"reward_percent" json:"reward_percent"`
	RewardType    string  `db:"reward_type" json:"reward_type"`
	RewardAmount  float64 `db:"reward_amount" json:"reward_amount"`
	SortOrder     int     `db:"sort_order" json:"sort_order"`
}

type ClientLoyalty struct {
	ID          int       `db:"id" json:"id"`
	ClientID    int       `db:"client_id" json:"client_id"`
	ProgramID   int       `db:"program_id" json:"program_id"`
	LevelID     *int      `db:"level_id" json:"level_id"`
	Balance     float64   `db:"balance" json:"balance"`
	TotalEarned float64   `db:"total_earned" json:"total_earned"`
	TotalSpent  float64   `db:"total_spent" json:"total_spent"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type CreateProgramRequest struct {
	Name   string        `json:"name" binding:"required"`
	Type   string        `json:"type" binding:"required,oneof=bonus discount"`
	Config ProgramConfig `json:"config"`
}

type UpdateProgramRequest struct {
	Name     *string        `json:"name,omitempty"`
	IsActive *bool          `json:"is_active,omitempty"`
	Config   *ProgramConfig `json:"config,omitempty"`
}

type CreateLevelRequest struct {
	Name          string  `json:"name" binding:"required"`
	Threshold     int     `json:"threshold" binding:"min=0"`
	RewardPercent float64 `json:"reward_percent" binding:"min=0,max=100"`
	RewardType    string  `json:"reward_type"`
	RewardAmount  float64 `json:"reward_amount" binding:"min=0"`
	SortOrder     int     `json:"sort_order"`
}

type BatchUpdateLevelsRequest struct {
	Levels []LoyaltyLevel `json:"levels" binding:"required"`
}
