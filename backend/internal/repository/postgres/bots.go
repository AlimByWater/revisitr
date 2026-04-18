package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Bots struct {
	pg *Module
}

func NewBots(pg *Module) *Bots {
	return &Bots{pg: pg}
}

func (r *Bots) Create(ctx context.Context, bot *entity.Bot) error {
	query := `
		INSERT INTO bots (org_id, program_id, name, token, username, status, settings)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	settingsVal, err := bot.Settings.Value()
	if err != nil {
		return fmt.Errorf("bots.Create settings value: %w", err)
	}

	err = r.pg.DB().QueryRowContext(ctx, query,
		bot.OrgID, bot.ProgramID, bot.Name, bot.Token, bot.Username, bot.Status, settingsVal,
	).Scan(&bot.ID, &bot.CreatedAt, &bot.UpdatedAt)
	if err != nil {
		return fmt.Errorf("bots.Create: %w", err)
	}

	return nil
}

func (r *Bots) GetByID(ctx context.Context, id int) (*entity.Bot, error) {
	var bot entity.Bot
	err := r.pg.DB().GetContext(ctx, &bot, "SELECT * FROM bots WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bots.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("bots.GetByID: %w", err)
	}
	return &bot, nil
}

func (r *Bots) GetByOrgID(ctx context.Context, orgID int) ([]entity.Bot, error) {
	var bots []entity.Bot
	err := r.pg.DB().SelectContext(ctx, &bots, "SELECT * FROM bots WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("bots.GetByOrgID: %w", err)
	}
	return bots, nil
}

func (r *Bots) HasPOSLocations(ctx context.Context, botID int) (bool, error) {
	var exists bool
	if err := r.pg.DB().GetContext(ctx, &exists,
		"SELECT EXISTS(SELECT 1 FROM bot_pos_locations WHERE bot_id = $1)", botID); err != nil {
		return false, fmt.Errorf("bots.HasPOSLocations: %w", err)
	}
	return exists, nil
}

func (r *Bots) Update(ctx context.Context, bot *entity.Bot) error {
	query := `UPDATE bots SET name = $1, status = $2, program_id = $3, updated_at = NOW() WHERE id = $4`
	result, err := r.pg.DB().ExecContext(ctx, query, bot.Name, bot.Status, bot.ProgramID, bot.ID)
	if err != nil {
		return fmt.Errorf("bots.Update: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bots.Update rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bots.Update: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Bots) GetByProgramID(ctx context.Context, programID int) ([]entity.Bot, error) {
	var bots []entity.Bot
	err := r.pg.DB().SelectContext(ctx, &bots,
		"SELECT * FROM bots WHERE program_id = $1 ORDER BY created_at DESC", programID)
	if err != nil {
		return nil, fmt.Errorf("bots.GetByProgramID: %w", err)
	}
	return bots, nil
}

func (r *Bots) UpdateSettings(ctx context.Context, id int, settings entity.BotSettings) error {
	settingsVal, err := settings.Value()
	if err != nil {
		return fmt.Errorf("bots.UpdateSettings settings value: %w", err)
	}

	query := `UPDATE bots SET settings = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.pg.DB().ExecContext(ctx, query, settingsVal, id)
	if err != nil {
		return fmt.Errorf("bots.UpdateSettings: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bots.UpdateSettings rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bots.UpdateSettings: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Bots) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM bots WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("bots.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bots.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bots.Delete: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Bots) GetAllActive(ctx context.Context) ([]entity.Bot, error) {
	var bots []entity.Bot
	err := r.pg.DB().SelectContext(ctx, &bots, "SELECT * FROM bots WHERE status = 'active'")
	if err != nil {
		return nil, fmt.Errorf("bots.GetAllActive: %w", err)
	}
	return bots, nil
}
