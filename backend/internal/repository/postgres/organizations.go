package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Organizations struct {
	pg *Module
}

func NewOrganizations(pg *Module) *Organizations {
	return &Organizations{pg: pg}
}

func (r *Organizations) GetByID(ctx context.Context, orgID int) (*entity.Organization, error) {
	var org entity.Organization
	err := r.pg.DB().GetContext(ctx, &org,
		"SELECT * FROM organizations WHERE id = $1", orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("organizations.GetByID: %w", err)
	}
	return &org, nil
}

func (r *Organizations) GetTimezone(ctx context.Context, orgID int) (string, error) {
	var tz string
	err := r.pg.DB().GetContext(ctx, &tz,
		"SELECT timezone FROM organizations WHERE id = $1", orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("organizations.GetTimezone: %w", err)
	}
	return tz, nil
}

func (r *Organizations) UpdateTimezone(ctx context.Context, orgID int, timezone string) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"UPDATE organizations SET timezone = $2 WHERE id = $1", orgID, timezone)
	if err != nil {
		return fmt.Errorf("organizations.UpdateTimezone: %w", err)
	}
	return nil
}
