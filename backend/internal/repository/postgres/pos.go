package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type POS struct {
	pg *Module
}

func NewPOS(pg *Module) *POS {
	return &POS{pg: pg}
}

func (r *POS) Create(ctx context.Context, pos *entity.POSLocation) error {
	query := `
		INSERT INTO pos_locations (org_id, bot_id, name, address, phone, schedule, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	scheduleVal, err := pos.Schedule.Value()
	if err != nil {
		return fmt.Errorf("pos.Create schedule value: %w", err)
	}

	err = r.pg.DB().QueryRowContext(ctx, query,
		pos.OrgID, pos.BotID, pos.Name, pos.Address, pos.Phone, scheduleVal, pos.IsActive,
	).Scan(&pos.ID, &pos.CreatedAt, &pos.UpdatedAt)
	if err != nil {
		return fmt.Errorf("pos.Create: %w", err)
	}

	return nil
}

func (r *POS) GetByID(ctx context.Context, id int) (*entity.POSLocation, error) {
	var pos entity.POSLocation
	err := r.pg.DB().GetContext(ctx, &pos, "SELECT * FROM pos_locations WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pos.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("pos.GetByID: %w", err)
	}
	return &pos, nil
}

func (r *POS) GetByOrgID(ctx context.Context, orgID int) ([]entity.POSLocation, error) {
	var locations []entity.POSLocation
	err := r.pg.DB().SelectContext(ctx, &locations,
		"SELECT * FROM pos_locations WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("pos.GetByOrgID: %w", err)
	}
	return locations, nil
}

func (r *POS) Update(ctx context.Context, pos *entity.POSLocation) error {
	query := `
		UPDATE pos_locations
		SET name = $1, address = $2, phone = $3, schedule = $4, is_active = $5, bot_id = $6, updated_at = NOW()
		WHERE id = $7`

	scheduleVal, err := pos.Schedule.Value()
	if err != nil {
		return fmt.Errorf("pos.Update schedule value: %w", err)
	}

	result, err := r.pg.DB().ExecContext(ctx, query,
		pos.Name, pos.Address, pos.Phone, scheduleVal, pos.IsActive, pos.BotID, pos.ID)
	if err != nil {
		return fmt.Errorf("pos.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("pos.Update rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pos.Update: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *POS) GetByBotID(ctx context.Context, botID int) ([]entity.POSLocation, error) {
	var locations []entity.POSLocation
	err := r.pg.DB().SelectContext(ctx, &locations,
		"SELECT * FROM pos_locations WHERE bot_id = $1 ORDER BY created_at DESC", botID)
	if err != nil {
		return nil, fmt.Errorf("pos.GetByBotID: %w", err)
	}
	return locations, nil
}

func (r *POS) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM pos_locations WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("pos.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("pos.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pos.Delete: %w", sql.ErrNoRows)
	}

	return nil
}
