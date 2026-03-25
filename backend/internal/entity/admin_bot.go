package entity

import "time"

type AdminBotLink struct {
	ID                int        `json:"id"                    db:"id"`
	UserID            int        `json:"user_id"               db:"user_id"`
	TelegramID        *int64     `json:"telegram_id,omitempty" db:"telegram_id"`
	OrgID             int        `json:"org_id"                db:"org_id"`
	Role              string     `json:"role"                  db:"role"` // "owner" | "manager"
	LinkedAt          *time.Time `json:"linked_at,omitempty"   db:"linked_at"`
	LinkCode          *string    `json:"-"                     db:"link_code"`
	LinkCodeExpiresAt *time.Time `json:"-"                     db:"link_code_expires_at"`
	CreatedAt         time.Time  `json:"created_at"            db:"created_at"`
}

type GenerateLinkCodeResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AdminBotStatus struct {
	Linked     bool   `json:"linked"`
	TelegramID *int64 `json:"telegram_id,omitempty"`
	Role       string `json:"role,omitempty"`
}
