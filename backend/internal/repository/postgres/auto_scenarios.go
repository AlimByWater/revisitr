package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"revisitr/internal/entity"
)

type AutoScenarios struct {
	pg *Module
}

func NewAutoScenarios(pg *Module) *AutoScenarios {
	return &AutoScenarios{pg: pg}
}

func (r *AutoScenarios) Create(ctx context.Context, scenario *entity.AutoScenario) error {
	query := `
		INSERT INTO auto_scenarios (org_id, bot_id, name, trigger_type, trigger_config, message, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	configVal, err := scenario.TriggerConfig.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Create config value: %w", err)
	}

	err = r.pg.DB().QueryRowContext(ctx, query,
		scenario.OrgID, scenario.BotID, scenario.Name, scenario.TriggerType,
		configVal, scenario.Message, scenario.IsActive,
	).Scan(&scenario.ID, &scenario.CreatedAt, &scenario.UpdatedAt)
	if err != nil {
		return fmt.Errorf("autoScenarios.Create: %w", err)
	}

	return nil
}

func (r *AutoScenarios) GetByOrgID(ctx context.Context, orgID int) ([]entity.AutoScenario, error) {
	var scenarios []entity.AutoScenario
	err := r.pg.DB().SelectContext(ctx, &scenarios,
		"SELECT * FROM auto_scenarios WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("autoScenarios.GetByOrgID: %w", err)
	}
	return scenarios, nil
}

func (r *AutoScenarios) GetByID(ctx context.Context, id int) (*entity.AutoScenario, error) {
	var scenario entity.AutoScenario
	err := r.pg.DB().GetContext(ctx, &scenario, "SELECT * FROM auto_scenarios WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("autoScenarios.GetByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("autoScenarios.GetByID: %w", err)
	}
	return &scenario, nil
}

func (r *AutoScenarios) Update(ctx context.Context, scenario *entity.AutoScenario) error {
	configVal, err := scenario.TriggerConfig.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Update config value: %w", err)
	}

	query := `
		UPDATE auto_scenarios
		SET name = $1, trigger_config = $2, message = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5`

	result, err := r.pg.DB().ExecContext(ctx, query,
		scenario.Name, configVal, scenario.Message, scenario.IsActive, scenario.ID,
	)
	if err != nil {
		return fmt.Errorf("autoScenarios.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("autoScenarios.Update rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("autoScenarios.Update: %w", sql.ErrNoRows)
	}

	return nil
}

func (r *AutoScenarios) Delete(ctx context.Context, id int) error {
	result, err := r.pg.DB().ExecContext(ctx, "DELETE FROM auto_scenarios WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("autoScenarios.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("autoScenarios.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("autoScenarios.Delete: %w", sql.ErrNoRows)
	}

	return nil
}
