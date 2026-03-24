package autoaction

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

type AutoActionHook interface {
	OnEvent(ctx context.Context, eventType string, clientID int, data map[string]interface{}) error
}

type EventHook struct {
	scenarios scenariosRepo
	executor  *ActionExecutor
	clients   botClientsRepo
	logger    *slog.Logger
}

func NewEventHook(
	scenarios scenariosRepo,
	executor *ActionExecutor,
	clients botClientsRepo,
	logger *slog.Logger,
) *EventHook {
	return &EventHook{
		scenarios: scenarios,
		executor:  executor,
		clients:   clients,
		logger:    logger,
	}
}

func (h *EventHook) OnEvent(ctx context.Context, eventType string, clientID int, data map[string]interface{}) error {
	scenarios, err := h.scenarios.GetActiveByTriggerType(ctx, eventType)
	if err != nil {
		return fmt.Errorf("get scenarios for event %s: %w", eventType, err)
	}

	if len(scenarios) == 0 {
		return nil
	}

	h.logger.Info("auto-action hook: processing event",
		"event_type", eventType,
		"client_id", clientID,
		"scenario_count", len(scenarios),
	)

	now := time.Now()
	triggerKey := fmt.Sprintf("%s:%d-%02d-%02d", eventType, now.Year(), now.Month(), now.Day())

	for _, scenario := range scenarios {
		client := h.findClient(ctx, scenario.BotID, clientID)
		if client == nil {
			continue
		}

		if err := h.executor.ExecuteWithDedup(ctx, scenario, *client, triggerKey); err != nil {
			h.logger.Error("auto-action hook: execution failed",
				"error", err,
				"scenario_id", scenario.ID,
				"client_id", clientID,
				"event_type", eventType,
			)
		}
	}

	return nil
}

func (h *EventHook) findClient(ctx context.Context, botID, clientID int) *entity.BotClient {
	const batchSize = 100
	offset := 0

	for {
		batch, total, err := h.clients.GetByBotID(ctx, botID, batchSize, offset)
		if err != nil {
			h.logger.Error("auto-action hook: get clients",
				"error", err,
				"bot_id", botID,
			)
			return nil
		}

		for i := range batch {
			if batch[i].ID == clientID {
				return &batch[i]
			}
		}

		offset += batchSize
		if offset >= total {
			break
		}
	}

	return nil
}
