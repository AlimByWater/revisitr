package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type Loyalty struct {
	pg *Module
}

func NewLoyalty(pg *Module) *Loyalty {
	return &Loyalty{pg: pg}
}

func (r *Loyalty) CreateProgram(ctx context.Context, program *entity.LoyaltyProgram) error {
	query := `
		INSERT INTO loyalty_programs (org_id, name, type, config, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	configVal, err := program.Config.Value()
	if err != nil {
		return fmt.Errorf("loyalty.CreateProgram config value: %w", err)
	}

	err = r.pg.DB().QueryRowContext(ctx, query,
		program.OrgID, program.Name, program.Type, configVal, program.IsActive,
	).Scan(&program.ID, &program.CreatedAt, &program.UpdatedAt)
	if err != nil {
		return fmt.Errorf("loyalty.CreateProgram: %w", err)
	}

	return nil
}

func (r *Loyalty) GetProgramByID(ctx context.Context, id int) (*entity.LoyaltyProgram, error) {
	var program entity.LoyaltyProgram
	err := r.pg.DB().GetContext(ctx, &program, "SELECT * FROM loyalty_programs WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("loyalty.GetProgramByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("loyalty.GetProgramByID: %w", err)
	}

	levels, err := r.GetLevelsByProgramID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loyalty.GetProgramByID levels: %w", err)
	}
	program.Levels = levels

	return &program, nil
}

func (r *Loyalty) GetProgramsByOrgID(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error) {
	var programs []entity.LoyaltyProgram
	err := r.pg.DB().SelectContext(ctx, &programs,
		"SELECT * FROM loyalty_programs WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("loyalty.GetProgramsByOrgID: %w", err)
	}
	return programs, nil
}

func (r *Loyalty) UpdateProgram(ctx context.Context, program *entity.LoyaltyProgram) error {
	configVal, err := program.Config.Value()
	if err != nil {
		return fmt.Errorf("loyalty.UpdateProgram config value: %w", err)
	}

	query := `
		UPDATE loyalty_programs
		SET name = $1, is_active = $2, config = $3, updated_at = NOW()
		WHERE id = $4`

	result, err := r.pg.DB().ExecContext(ctx, query,
		program.Name, program.IsActive, configVal, program.ID)
	if err != nil {
		return fmt.Errorf("loyalty.UpdateProgram: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("loyalty.UpdateProgram rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("loyalty.UpdateProgram: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Loyalty) CreateLevel(ctx context.Context, level *entity.LoyaltyLevel) error {
	query := `
		INSERT INTO loyalty_levels (program_id, name, threshold, reward_percent, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.pg.DB().QueryRowContext(ctx, query,
		level.ProgramID, level.Name, level.Threshold, level.RewardPercent, level.SortOrder,
	).Scan(&level.ID)
	if err != nil {
		return fmt.Errorf("loyalty.CreateLevel: %w", err)
	}

	return nil
}

func (r *Loyalty) GetLevelsByProgramID(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error) {
	var levels []entity.LoyaltyLevel
	err := r.pg.DB().SelectContext(ctx, &levels,
		"SELECT * FROM loyalty_levels WHERE program_id = $1 ORDER BY sort_order", programID)
	if err != nil {
		return nil, fmt.Errorf("loyalty.GetLevelsByProgramID: %w", err)
	}
	return levels, nil
}

func (r *Loyalty) UpdateLevel(ctx context.Context, level *entity.LoyaltyLevel) error {
	query := `
		UPDATE loyalty_levels
		SET name = $1, threshold = $2, reward_percent = $3, sort_order = $4
		WHERE id = $5`

	result, err := r.pg.DB().ExecContext(ctx, query,
		level.Name, level.Threshold, level.RewardPercent, level.SortOrder, level.ID)
	if err != nil {
		return fmt.Errorf("loyalty.UpdateLevel: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("loyalty.UpdateLevel rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("loyalty.UpdateLevel: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Loyalty) DeleteLevel(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM loyalty_levels WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("loyalty.DeleteLevel: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("loyalty.DeleteLevel rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("loyalty.DeleteLevel: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *Loyalty) GetClientLoyalty(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
	var cl entity.ClientLoyalty
	err := r.pg.DB().GetContext(ctx, &cl,
		"SELECT * FROM client_loyalty WHERE client_id = $1 AND program_id = $2", clientID, programID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("loyalty.GetClientLoyalty: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("loyalty.GetClientLoyalty: %w", err)
	}
	return &cl, nil
}

func (r *Loyalty) UpsertClientLoyalty(ctx context.Context, cl *entity.ClientLoyalty) error {
	query := `
		INSERT INTO client_loyalty (client_id, program_id, level_id, balance, total_earned, total_spent, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (client_id, program_id) DO UPDATE SET
			level_id = EXCLUDED.level_id,
			balance = EXCLUDED.balance,
			total_earned = EXCLUDED.total_earned,
			total_spent = EXCLUDED.total_spent,
			updated_at = NOW()
		RETURNING id, updated_at`

	err := r.pg.DB().QueryRowContext(ctx, query,
		cl.ClientID, cl.ProgramID, cl.LevelID, cl.Balance, cl.TotalEarned, cl.TotalSpent,
	).Scan(&cl.ID, &cl.UpdatedAt)
	if err != nil {
		return fmt.Errorf("loyalty.UpsertClientLoyalty: %w", err)
	}

	return nil
}

func (r *Loyalty) CreateTransaction(ctx context.Context, tx *entity.LoyaltyTransaction) error {
	query := `
		INSERT INTO loyalty_transactions (client_id, program_id, type, amount, balance_after, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.pg.DB().QueryRowContext(ctx, query,
		tx.ClientID, tx.ProgramID, tx.Type, tx.Amount, tx.BalanceAfter, tx.Description, tx.CreatedBy,
	).Scan(&tx.ID, &tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("loyalty.CreateTransaction: %w", err)
	}

	return nil
}
