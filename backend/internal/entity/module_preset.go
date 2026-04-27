package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Menu render modes
const (
	MenuRenderTabs     = "tabs"
	MenuRenderList     = "list"
	MenuRenderCarousel = "carousel"
)

// IsValidMenuRenderMode checks if the given mode is a known menu render mode.
func IsValidMenuRenderMode(mode string) bool {
	switch mode {
	case MenuRenderTabs, MenuRenderList, MenuRenderCarousel:
		return true
	}
	return false
}

// JSONB is a generic JSON column type for sqlx scanning.
type JSONB json.RawMessage

func (j *JSONB) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		cp := make([]byte, len(v))
		copy(cp, v)
		*j = JSONB(cp)
		return nil
	case string:
		*j = JSONB(v)
		return nil
	case nil:
		*j = JSONB("{}")
		return nil
	default:
		return fmt.Errorf("JSONB.Scan: unsupported type %T", src)
	}
}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return []byte(j), nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return []byte(j), nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	*j = JSONB(data)
	return nil
}

// ModulePreset is a predefined template for a bot module's UI/UX.
type ModulePreset struct {
	ID          int       `db:"id" json:"id"`
	ModuleKey   string    `db:"module_key" json:"module_key"`
	PresetKey   string    `db:"preset_key" json:"preset_key"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Definition  JSONB     `db:"definition" json:"definition"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// BotModuleSettings stores the per-bot, per-module preset selection and customizations.
type BotModuleSettings struct {
	BotID          int       `db:"bot_id" json:"bot_id"`
	ModuleKey      string    `db:"module_key" json:"module_key"`
	PresetID       *int      `db:"preset_id" json:"preset_id"`
	PresetKey      string    `db:"preset_key" json:"preset_key"`
	Customized     bool      `db:"customized" json:"customized"`
	Customizations JSONB     `db:"customizations" json:"customizations"`
	Config         JSONB     `db:"config" json:"config"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
