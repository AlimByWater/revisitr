package entity

import "time"

type MasterBotLink struct {
	ID               int       `json:"id"                db:"id"`
	OrgID            int       `json:"org_id"            db:"org_id"`
	TelegramUserID   int64     `json:"telegram_user_id"  db:"telegram_user_id"`
	TelegramUsername string    `json:"telegram_username"  db:"telegram_username"`
	IsActive         bool      `json:"is_active"         db:"is_active"`
	CreatedAt        time.Time `json:"created_at"        db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"        db:"updated_at"`
}

// AuthToken represents a one-time deep link token for master bot activation.
type MasterBotAuthToken struct {
	OrgID  int `json:"org_id"`
	UserID int `json:"user_id"`
}

// CreateManagedBotRequest is the wizard payload for creating a managed bot.
type CreateManagedBotRequest struct {
	Name             string          `json:"name"`
	Username         string          `json:"username"`
	Description      string          `json:"description"`
	WelcomeMessage   string          `json:"welcome_message,omitempty"`
	WelcomeContent   *MessageContent `json:"welcome_content,omitempty"`
	RegistrationForm []FormField     `json:"registration_form,omitempty"`
	Modules          []string        `json:"modules,omitempty"`
}

// CreateManagedBotResponse returned after wizard submission.
type CreateManagedBotResponse struct {
	BotID    int    `json:"bot_id"`
	DeepLink string `json:"deep_link"`
	Status   string `json:"status"`
}

// ActivationLinkResponse returned when generating master bot deep link.
type ActivationLinkResponse struct {
	DeepLink  string    `json:"deep_link"`
	ExpiresAt time.Time `json:"expires_at"`
}
