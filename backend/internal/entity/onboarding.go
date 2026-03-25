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
		if err := json.Unmarshal(v, s); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(v), s); err != nil {
			return err
		}
	case nil:
		*s = OnboardingState{}
	default:
		return fmt.Errorf("OnboardingState.Scan: unsupported type %T", src)
	}
	if s.Steps == nil {
		s.Steps = make(map[string]OnboardingStep)
	}
	return nil
}

type UpdateOnboardingRequest struct {
	Step      string `json:"step" binding:"required"`
	Completed bool   `json:"completed"`
	Skipped   bool   `json:"skipped"`
	EntityID  *int   `json:"entity_id,omitempty"`
}
