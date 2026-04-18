package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type BotSettings struct {
	Modules            []string        `json:"modules"`
	Buttons            []BotButton     `json:"buttons"`
	RegistrationForm   []FormField     `json:"registration_form"`
	WelcomeMessage     string          `json:"welcome_message"`           // Legacy
	WelcomeContent     *MessageContent `json:"welcome_content,omitempty"` // New: composite welcome
	ModuleConfigs      ModuleConfigs   `json:"module_configs,omitempty"`
	PosSelectorEnabled bool            `json:"pos_selector_enabled,omitempty"`
	ContactsPOSIDs     []int           `json:"contacts_pos_ids,omitempty"`
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
	Label   string          `json:"label"`
	Type    string          `json:"type"`
	Value   string          `json:"value"`
	Content *MessageContent `json:"content,omitempty"`
}

type FormField struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
}

type ModuleConfigs struct {
	Menu     MenuModuleConfig     `json:"menu,omitempty"`
	Booking  BookingModuleConfig  `json:"booking,omitempty"`
	Feedback FeedbackModuleConfig `json:"feedback,omitempty"`
}

type MenuModuleConfig struct {
	UnavailableMessage string `json:"unavailable_message,omitempty"`
}

type BookingModuleConfig struct {
	IntroContent     *MessageContent   `json:"intro_content,omitempty"`
	DateFromDays     int               `json:"date_from_days,omitempty"`
	DateToDays       int               `json:"date_to_days,omitempty"`
	TimeSlots        []BookingTimeSlot `json:"time_slots,omitempty"`
	PartySizeOptions []string          `json:"party_size_options,omitempty"`
	POSIDs           []int             `json:"pos_ids,omitempty"`
}

type BookingTimeSlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type FeedbackModuleConfig struct {
	PromptMessage  string `json:"prompt_message,omitempty"`
	SuccessMessage string `json:"success_message,omitempty"`
}

type Bot struct {
	ID                  int         `db:"id" json:"id"`
	OrgID               int         `db:"org_id" json:"org_id"`
	ProgramID           *int        `db:"program_id" json:"program_id"`
	Name                string      `db:"name" json:"name"`
	Token               string      `db:"token" json:"-"`
	Username            string      `db:"username" json:"username"`
	Status              string      `db:"status" json:"status"` // "active" | "inactive" | "pending" | "error"
	Settings            BotSettings `db:"settings" json:"settings"`
	CreatedAt           time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time   `db:"updated_at" json:"updated_at"`
	TokenMasked         string      `db:"-" json:"token_masked,omitempty"`
	IsManagedBot        bool        `db:"is_managed" json:"is_managed"`
	ManagedBotID        *int64      `db:"managed_bot_id" json:"managed_bot_id,omitempty"`
	CreatedByTelegramID *int64      `db:"created_by_telegram_id" json:"created_by_telegram_id,omitempty"`
}

// MaskToken returns a partially masked token for safe display.
// For tokens longer than 10 characters, it shows the first 5 and last 5 characters.
// For shorter non-empty tokens, it shows the first 2 characters followed by "...".
func MaskToken(token string) string {
	if len(token) > 10 {
		return token[:5] + "..." + token[len(token)-5:]
	}
	if len(token) > 0 {
		return token[:2] + "..."
	}
	return ""
}

type CreateBotRequest struct {
	Name      string `json:"name" binding:"required"`
	Token     string `json:"token" binding:"required"`
	ProgramID *int   `json:"program_id"`
}

type UpdateBotRequest struct {
	Name      *string `json:"name,omitempty"`
	Status    *string `json:"status,omitempty"`
	ProgramID *int    `json:"program_id,omitempty"`
}

type UpdateBotSettingsRequest struct {
	Modules            *[]string       `json:"modules,omitempty"`
	Buttons            *[]BotButton    `json:"buttons,omitempty"`
	RegistrationForm   *[]FormField    `json:"registration_form,omitempty"`
	WelcomeMessage     *string         `json:"welcome_message,omitempty"`
	WelcomeContent     *MessageContent `json:"welcome_content,omitempty"`
	ModuleConfigs      *ModuleConfigs  `json:"module_configs,omitempty"`
	PosSelectorEnabled *bool           `json:"pos_selector_enabled,omitempty"`
	ContactsPOSIDs     *[]int          `json:"contacts_pos_ids,omitempty"`
}
