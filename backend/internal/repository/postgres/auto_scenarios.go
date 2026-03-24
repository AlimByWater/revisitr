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
		INSERT INTO auto_scenarios (org_id, bot_id, name, trigger_type, trigger_config, message, actions, timing, conditions, is_template, template_key, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	configVal, err := scenario.TriggerConfig.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Create config value: %w", err)
	}

	actionsVal, err := scenario.Actions.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Create actions value: %w", err)
	}

	timingVal, err := scenario.Timing.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Create timing value: %w", err)
	}

	conditionsVal, err := scenario.Conditions.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Create conditions value: %w", err)
	}

	err = r.pg.DB().QueryRowContext(ctx, query,
		scenario.OrgID, scenario.BotID, scenario.Name, scenario.TriggerType,
		configVal, scenario.Message, actionsVal, timingVal, conditionsVal,
		scenario.IsTemplate, scenario.TemplateKey, scenario.IsActive,
	).Scan(&scenario.ID, &scenario.CreatedAt, &scenario.UpdatedAt)
	if err != nil {
		return fmt.Errorf("autoScenarios.Create: %w", err)
	}

	return nil
}

func (r *AutoScenarios) GetByOrgID(ctx context.Context, orgID int) ([]entity.AutoScenario, error) {
	var scenarios []entity.AutoScenario
	err := r.pg.DB().SelectContext(ctx, &scenarios,
		"SELECT * FROM auto_scenarios WHERE org_id = $1 AND is_template = false ORDER BY created_at DESC", orgID)
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

	actionsVal, err := scenario.Actions.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Update actions value: %w", err)
	}

	timingVal, err := scenario.Timing.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Update timing value: %w", err)
	}

	conditionsVal, err := scenario.Conditions.Value()
	if err != nil {
		return fmt.Errorf("autoScenarios.Update conditions value: %w", err)
	}

	query := `
		UPDATE auto_scenarios
		SET name = $1, trigger_config = $2, message = $3, is_active = $4,
		    actions = $5, timing = $6, conditions = $7, updated_at = NOW()
		WHERE id = $8`

	result, err := r.pg.DB().ExecContext(ctx, query,
		scenario.Name, configVal, scenario.Message, scenario.IsActive,
		actionsVal, timingVal, conditionsVal, scenario.ID,
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

func (r *AutoScenarios) GetTemplates(ctx context.Context) ([]entity.AutoScenario, error) {
	var scenarios []entity.AutoScenario
	err := r.pg.DB().SelectContext(ctx, &scenarios,
		`SELECT id, COALESCE(org_id, 0) AS org_id, COALESCE(bot_id, 0) AS bot_id,
		        name, trigger_type, trigger_config, message, actions, timing,
		        conditions, is_template, template_key, is_active, created_at, updated_at
		 FROM auto_scenarios WHERE is_template = true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("autoScenarios.GetTemplates: %w", err)
	}
	return scenarios, nil
}

func (r *AutoScenarios) GetActiveDateBased(ctx context.Context) ([]entity.AutoScenario, error) {
	var scenarios []entity.AutoScenario
	err := r.pg.DB().SelectContext(ctx, &scenarios,
		`SELECT * FROM auto_scenarios
		 WHERE is_active = true AND is_template = false
		   AND trigger_type IN ('birthday', 'holiday', 'inactive_days')`)
	if err != nil {
		return nil, fmt.Errorf("autoScenarios.GetActiveDateBased: %w", err)
	}
	return scenarios, nil
}

func (r *AutoScenarios) CreateActionLog(ctx context.Context, log *entity.AutoActionLog) error {
	query := `
		INSERT INTO auto_action_log (scenario_id, client_id, action_type, action_data, result, error_msg)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, executed_at`

	err := r.pg.DB().QueryRowContext(ctx, query,
		log.ScenarioID, log.ClientID, log.ActionType, log.ActionData, log.Result, log.ErrorMsg,
	).Scan(&log.ID, &log.ExecutedAt)
	if err != nil {
		return fmt.Errorf("autoScenarios.CreateActionLog: %w", err)
	}
	return nil
}

func (r *AutoScenarios) GetActionLog(ctx context.Context, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error) {
	var total int
	err := r.pg.DB().GetContext(ctx, &total,
		"SELECT COUNT(*) FROM auto_action_log WHERE scenario_id = $1", scenarioID)
	if err != nil {
		return nil, 0, fmt.Errorf("autoScenarios.GetActionLog count: %w", err)
	}

	var logs []entity.AutoActionLog
	err = r.pg.DB().SelectContext(ctx, &logs,
		`SELECT * FROM auto_action_log WHERE scenario_id = $1
		 ORDER BY executed_at DESC LIMIT $2 OFFSET $3`,
		scenarioID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("autoScenarios.GetActionLog: %w", err)
	}
	return logs, total, nil
}

func (r *AutoScenarios) CheckDedup(ctx context.Context, scenarioID, clientID int, triggerKey string) (bool, error) {
	var exists bool
	err := r.pg.DB().GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM auto_action_dedup
		 WHERE scenario_id = $1 AND client_id = $2 AND trigger_key = $3)`,
		scenarioID, clientID, triggerKey)
	if err != nil {
		return false, fmt.Errorf("autoScenarios.CheckDedup: %w", err)
	}
	return exists, nil
}

func (r *AutoScenarios) CreateDedup(ctx context.Context, scenarioID, clientID int, triggerKey string) error {
	_, err := r.pg.DB().ExecContext(ctx,
		`INSERT INTO auto_action_dedup (scenario_id, client_id, trigger_key)
		 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		scenarioID, clientID, triggerKey)
	if err != nil {
		return fmt.Errorf("autoScenarios.CreateDedup: %w", err)
	}
	return nil
}

func (r *AutoScenarios) GetActiveByTriggerType(ctx context.Context, triggerType string) ([]entity.AutoScenario, error) {
	var scenarios []entity.AutoScenario
	err := r.pg.DB().SelectContext(ctx, &scenarios,
		`SELECT * FROM auto_scenarios
		 WHERE is_active = true AND is_template = false AND trigger_type = $1`,
		triggerType)
	if err != nil {
		return nil, fmt.Errorf("autoScenarios.GetActiveByTriggerType: %w", err)
	}
	return scenarios, nil
}
