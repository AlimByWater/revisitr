package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type BotSettings struct {
	Modules          []string    `json:"modules"`
	Buttons          []BotButton `json:"buttons"`
	RegistrationForm []FormField `json:"registration_form"`
	WelcomeMessage   string      `json:"welcome_message"`
}

func (s *BotSettings) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	case nil:
		*s = BotSettings{}
		return nil
	default:
		return fmt.Errorf("BotSettings.Scan: unsupported type %T", src)
	}
}

func (s BotSettings) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("BotSettings.Value: %w", err)
	}
	return b, nil
}

type BotButton struct {
	Label string `json:"label"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type FormField struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type Bot struct {
	ID        int         `db:"id" json:"id"`
	OrgID     int         `db:"org_id" json:"org_id"`
	Name      string      `db:"name" json:"name"`
	Token     string      `db:"token" json:"-"`
	Username  string      `db:"username" json:"username"`
	Status    string      `db:"status" json:"status"`
	Settings  BotSettings `db:"settings" json:"settings"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
}

type CreateBotRequest struct {
	Name  string `json:"name" binding:"required"`
	Token string `json:"token" binding:"required"`
}

type UpdateBotRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

type UpdateBotSettingsRequest struct {
	Modules          *[]string    `json:"modules,omitempty"`
	Buttons          *[]BotButton `json:"buttons,omitempty"`
	RegistrationForm *[]FormField `json:"registration_form,omitempty"`
	WelcomeMessage   *string      `json:"welcome_message,omitempty"`
}
