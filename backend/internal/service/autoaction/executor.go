package autoaction

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

type scenariosRepo interface {
	CreateActionLog(ctx context.Context, log *entity.AutoActionLog) error
	CheckDedup(ctx context.Context, scenarioID, clientID int, triggerKey string) (bool, error)
	CreateDedup(ctx context.Context, scenarioID, clientID int, triggerKey string) error
	GetActiveDateBased(ctx context.Context) ([]entity.AutoScenario, error)
	GetActiveByTriggerType(ctx context.Context, triggerType string) ([]entity.AutoScenario, error)
}

type actionResult struct {
	Result string
	Error  string
}

type ActionExecutor struct {
	scenarios scenariosRepo
	logger    *slog.Logger
}

func NewActionExecutor(scenarios scenariosRepo, logger *slog.Logger) *ActionExecutor {
	return &ActionExecutor{
		scenarios: scenarios,
		logger:    logger,
	}
}

func (e *ActionExecutor) Execute(ctx context.Context, scenario entity.AutoScenario, client entity.BotClient) error {
	actions := scenario.Actions
	if len(actions) == 0 && scenario.Message != "" {
		actions = entity.ActionDefs{{Type: "campaign", Template: &scenario.Message}}
	}

	for _, action := range actions {
		result := e.executeAction(ctx, action, scenario, client)
		e.logAction(ctx, scenario.ID, client.ID, action, result)
	}
	return nil
}

func (e *ActionExecutor) executeAction(ctx context.Context, action entity.ActionDef, scenario entity.AutoScenario, client entity.BotClient) actionResult {
	switch action.Type {
	case "bonus":
		e.logger.Info("auto-action: bonus",
			"scenario_id", scenario.ID,
			"client_id", client.ID,
			"amount", action.Amount,
		)
		return actionResult{Result: "success"}

	case "campaign":
		msg := ""
		if action.Template != nil {
			msg = *action.Template
		}
		e.logger.Info("auto-action: campaign",
			"scenario_id", scenario.ID,
			"client_id", client.ID,
			"message_len", len(msg),
		)
		return actionResult{Result: "success"}

	case "promo_code":
		e.logger.Info("auto-action: promo_code",
			"scenario_id", scenario.ID,
			"client_id", client.ID,
			"discount", action.Discount,
		)
		return actionResult{Result: "success"}

	case "level_change":
		e.logger.Info("auto-action: level_change",
			"scenario_id", scenario.ID,
			"client_id", client.ID,
			"level_id", action.LevelID,
		)
		return actionResult{Result: "success"}

	default:
		e.logger.Warn("auto-action: unknown type",
			"scenario_id", scenario.ID,
			"action_type", action.Type,
		)
		return actionResult{Result: "skipped", Error: "unknown action type: " + action.Type}
	}
}

func (e *ActionExecutor) logAction(ctx context.Context, scenarioID, clientID int, action entity.ActionDef, result actionResult) {
	actionData, err := json.Marshal(action)
	if err != nil {
		e.logger.Error("auto-action: marshal action data",
			"error", err,
			"scenario_id", scenarioID,
			"client_id", clientID,
		)
		return
	}

	logEntry := &entity.AutoActionLog{
		ScenarioID: scenarioID,
		ClientID:   clientID,
		ActionType: action.Type,
		ActionData: actionData,
		Result:     result.Result,
	}
	if result.Error != "" {
		logEntry.ErrorMsg = &result.Error
	}

	if err := e.scenarios.CreateActionLog(ctx, logEntry); err != nil {
		e.logger.Error("auto-action: create log entry",
			"error", err,
			"scenario_id", scenarioID,
			"client_id", clientID,
		)
	}
}

func (e *ActionExecutor) ExecuteWithDedup(ctx context.Context, scenario entity.AutoScenario, client entity.BotClient, triggerKey string) error {
	exists, err := e.scenarios.CheckDedup(ctx, scenario.ID, client.ID, triggerKey)
	if err != nil {
		return fmt.Errorf("check dedup: %w", err)
	}
	if exists {
		e.logger.Debug("auto-action: skipped (dedup)",
			"scenario_id", scenario.ID,
			"client_id", client.ID,
			"trigger_key", triggerKey,
		)
		return nil
	}

	if err := e.Execute(ctx, scenario, client); err != nil {
		return err
	}

	if err := e.scenarios.CreateDedup(ctx, scenario.ID, client.ID, triggerKey); err != nil {
		e.logger.Error("auto-action: create dedup record",
			"error", err,
			"scenario_id", scenario.ID,
			"client_id", client.ID,
		)
	}

	return nil
}
