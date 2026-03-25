package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Onboarding struct {
	pg *Module
}

func NewOnboarding(pg *Module) *Onboarding {
	return &Onboarding{pg: pg}
}

func (r *Onboarding) GetByOrgID(ctx context.Context, orgID int) (*entity.Organization, error) {
	var org entity.Organization
	err := r.pg.DB().GetContext(ctx, &org,
		"SELECT * FROM organizations WHERE id = $1", orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("onboarding.GetByOrgID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("onboarding.GetByOrgID: %w", err)
	}
	return &org, nil
}

func (r *Onboarding) UpdateState(ctx context.Context, orgID int, state entity.OnboardingState) error {
	stateVal, err := state.Value()
	if err != nil {
		return fmt.Errorf("onboarding.UpdateState value: %w", err)
	}

	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE organizations SET onboarding_state = $1 WHERE id = $2",
		stateVal, orgID)
	if err != nil {
		return fmt.Errorf("onboarding.UpdateState: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("onboarding.UpdateState: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Onboarding) SetCompleted(ctx context.Context, orgID int, completed bool) error {
	result, err := r.pg.DB().ExecContext(ctx,
		"UPDATE organizations SET onboarding_completed = $1 WHERE id = $2",
		completed, orgID)
	if err != nil {
		return fmt.Errorf("onboarding.SetCompleted: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("onboarding.SetCompleted: %w", sql.ErrNoRows)
	}
	return nil
}
