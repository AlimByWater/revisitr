package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type OnboardingState struct {
	CurrentStep int                       `json:"current_step"`
	Steps       map[string]OnboardingStep `json:"steps"`
}

type OnboardingStep struct {
	Completed bool `json:"completed"`
	Skipped   bool `json:"skipped"`
	EntityID  *int `json:"entity_id,omitempty"`
}

func (s OnboardingState) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("OnboardingState.Value: %w", err)
	}
	return b, nil
}

func (s *OnboardingState) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	case nil:
		*s = OnboardingState{}
		return nil
	default:
		return fmt.Errorf("OnboardingState.Scan: unsupported type %T", src)
	}
}

type UpdateOnboardingRequest struct {
	Step      int  `json:"step" binding:"required,min=1,max=6"`
	Completed bool `json:"completed"`
	Skipped   bool `json:"skipped"`
	EntityID  *int `json:"entity_id,omitempty"`
}
