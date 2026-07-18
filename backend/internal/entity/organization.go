package entity

import "time"

type Organization struct {
	ID                   int             `db:"id" json:"id"`
	Name                 string          `db:"name" json:"name"`
	OwnerID              int             `db:"owner_id" json:"owner_id"`
	OnboardingCompleted  bool            `db:"onboarding_completed" json:"onboarding_completed"`
	OnboardingState      OnboardingState `db:"onboarding_state" json:"onboarding_state"`
	Timezone             string          `db:"timezone" json:"timezone"` // IANA, e.g. "Europe/Moscow"
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
}

type UpdateOrganizationRequest struct {
	Timezone *string `json:"timezone,omitempty"`
}
