package autoaction

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

type botClientsRepo interface {
	GetByBotID(ctx context.Context, botID, limit, offset int) ([]entity.BotClient, int, error)
}

type AutoActionScheduler struct {
	scenarios scenariosRepo
	executor  *ActionExecutor
	clients   botClientsRepo
	logger    *slog.Logger
}

func NewAutoActionScheduler(
	scenarios scenariosRepo,
	executor *ActionExecutor,
	clients botClientsRepo,
	logger *slog.Logger,
) *AutoActionScheduler {
	return &AutoActionScheduler{
		scenarios: scenarios,
		executor:  executor,
		clients:   clients,
		logger:    logger,
	}
}

func (s *AutoActionScheduler) Evaluate(ctx context.Context) error {
	scenarios, err := s.scenarios.GetActiveDateBased(ctx)
	if err != nil {
		return fmt.Errorf("get active date-based scenarios: %w", err)
	}

	s.logger.Info("auto-action scheduler: evaluating scenarios", "count", len(scenarios))

	for _, scenario := range scenarios {
		s.evaluateScenario(ctx, scenario)
	}
	return nil
}

func (s *AutoActionScheduler) evaluateScenario(ctx context.Context, scenario entity.AutoScenario) {
	switch scenario.TriggerType {
	case "birthday":
		s.evaluateBirthday(ctx, scenario)
	case "holiday":
		s.evaluateHoliday(ctx, scenario)
	case "inactive_days":
		s.evaluateInactivity(ctx, scenario)
	default:
		s.logger.Warn("auto-action scheduler: unknown trigger type",
			"scenario_id", scenario.ID,
			"trigger_type", scenario.TriggerType,
		)
	}
}

func (s *AutoActionScheduler) evaluateBirthday(ctx context.Context, scenario entity.AutoScenario) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	daysBefore := 0
	if scenario.Timing.DaysBefore != nil {
		daysBefore = *scenario.Timing.DaysBefore
	}
	daysAfter := 0
	if scenario.Timing.DaysAfter != nil {
		daysAfter = *scenario.Timing.DaysAfter
	}

	clients := s.getAllClients(ctx, scenario.BotID)
	for _, client := range clients {
		if client.BirthDate == nil {
			continue
		}

		bd := *client.BirthDate
		birthdayThisYear := time.Date(now.Year(), bd.Month(), bd.Day(), 0, 0, 0, 0, now.Location())

		windowStart := birthdayThisYear.AddDate(0, 0, -daysBefore)
		windowEnd := birthdayThisYear.AddDate(0, 0, daysAfter)

		if (today.Equal(windowStart) || today.After(windowStart)) && (today.Equal(windowEnd) || today.Before(windowEnd)) {
			triggerKey := fmt.Sprintf("birthday:%d", now.Year())
			if err := s.executor.ExecuteWithDedup(ctx, scenario, client, triggerKey); err != nil {
				s.logger.Error("auto-action scheduler: birthday execution failed",
					"error", err,
					"scenario_id", scenario.ID,
					"client_id", client.ID,
				)
			}
		}
	}
}

func (s *AutoActionScheduler) evaluateHoliday(ctx context.Context, scenario entity.AutoScenario) {
	now := time.Now()

	targetMonth := 0
	targetDay := 0

	if scenario.Timing.Month != nil {
		targetMonth = *scenario.Timing.Month
	}
	if scenario.Timing.Day != nil {
		targetDay = *scenario.Timing.Day
	}

	if targetMonth == 0 || targetDay == 0 {
		s.logger.Warn("auto-action scheduler: holiday scenario missing month/day",
			"scenario_id", scenario.ID,
		)
		return
	}

	if int(now.Month()) != targetMonth || now.Day() != targetDay {
		return
	}

	triggerKey := fmt.Sprintf("holiday:%d-%02d-%02d", now.Year(), targetMonth, targetDay)

	clients := s.getAllClients(ctx, scenario.BotID)
	for _, client := range clients {
		if err := s.executor.ExecuteWithDedup(ctx, scenario, client, triggerKey); err != nil {
			s.logger.Error("auto-action scheduler: holiday execution failed",
				"error", err,
				"scenario_id", scenario.ID,
				"client_id", client.ID,
			)
		}
	}
}

func (s *AutoActionScheduler) evaluateInactivity(ctx context.Context, scenario entity.AutoScenario) {
	now := time.Now()

	inactiveDays := 0
	if scenario.TriggerConfig.Days != nil {
		inactiveDays = *scenario.TriggerConfig.Days
	}
	if inactiveDays <= 0 {
		s.logger.Warn("auto-action scheduler: inactive_days scenario missing days config",
			"scenario_id", scenario.ID,
		)
		return
	}

	threshold := now.AddDate(0, 0, -inactiveDays)

	clients := s.getAllClients(ctx, scenario.BotID)
	for _, client := range clients {
		// Use registered_at as a proxy for last activity
		// A proper implementation would check a last_activity column
		if client.RegisteredAt.Before(threshold) {
			triggerKey := fmt.Sprintf("inactive:%d:%d", now.Year(), now.YearDay()/7)
			if err := s.executor.ExecuteWithDedup(ctx, scenario, client, triggerKey); err != nil {
				s.logger.Error("auto-action scheduler: inactivity execution failed",
					"error", err,
					"scenario_id", scenario.ID,
					"client_id", client.ID,
				)
			}
		}
	}
}

func (s *AutoActionScheduler) getAllClients(ctx context.Context, botID int) []entity.BotClient {
	const batchSize = 100
	var all []entity.BotClient

	offset := 0
	for {
		batch, total, err := s.clients.GetByBotID(ctx, botID, batchSize, offset)
		if err != nil {
			s.logger.Error("auto-action scheduler: get clients",
				"error", err,
				"bot_id", botID,
				"offset", offset,
			)
			break
		}
		all = append(all, batch...)
		offset += batchSize
		if offset >= total {
			break
		}
	}
	return all
}
