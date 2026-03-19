package campaign

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type scenariosRepo interface {
	GetByOrgID(ctx context.Context, orgID int) ([]entity.AutoScenario, error)
}

type allBotsRepo interface {
	GetAllActive(ctx context.Context) ([]entity.Bot, error)
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
}

type schedulerClientsRepo interface {
	GetByBotID(ctx context.Context, botID int, limit, offset int) ([]entity.BotClient, int, error)
}

type loyaltyRepo interface {
	GetClientLoyalty(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
	GetProgramsByOrgID(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
}

type Scheduler struct {
	scenarios  scenariosRepo
	bots       allBotsRepo
	botClients schedulerClientsRepo
	loyalty    loyaltyRepo
	logger     *slog.Logger
	interval   time.Duration
}

func NewScheduler(
	scenarios scenariosRepo,
	bots allBotsRepo,
	botClients schedulerClientsRepo,
	loyalty loyaltyRepo,
	logger *slog.Logger,
) *Scheduler {
	return &Scheduler{
		scenarios:  scenarios,
		bots:       bots,
		botClients: botClients,
		loyalty:    loyalty,
		logger:     logger,
		interval:   1 * time.Hour,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.logger.Info("auto-scenario scheduler started", "interval", s.interval)

	// Run immediately once, then on interval
	s.evaluate(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("auto-scenario scheduler stopped")
			return
		case <-ticker.C:
			s.evaluate(ctx)
		}
	}
}

func (s *Scheduler) evaluate(ctx context.Context) {
	bots, err := s.bots.GetAllActive(ctx)
	if err != nil {
		s.logger.Error("scheduler: get active bots", "error", err)
		return
	}

	for _, bot := range bots {
		scenarios, err := s.scenarios.GetByOrgID(ctx, bot.OrgID)
		if err != nil {
			s.logger.Error("scheduler: get scenarios", "error", err, "org_id", bot.OrgID)
			continue
		}

		for _, scenario := range scenarios {
			if !scenario.IsActive || scenario.BotID != bot.ID {
				continue
			}

			s.evaluateScenario(ctx, bot, scenario)
		}
	}
}

func (s *Scheduler) evaluateScenario(ctx context.Context, bot entity.Bot, scenario entity.AutoScenario) {
	switch scenario.TriggerType {
	case "birthday":
		s.evaluateBirthday(ctx, bot, scenario)
	case "inactive_days":
		s.evaluateInactiveDays(ctx, bot, scenario)
	default:
		// visit_count, bonus_threshold, level_up are event-driven, not scheduled
	}
}

func (s *Scheduler) evaluateBirthday(ctx context.Context, bot entity.Bot, scenario entity.AutoScenario) {
	clients := s.getAllBotClients(ctx, bot.ID)
	today := time.Now()

	tBot, err := telego.NewBot(bot.Token)
	if err != nil {
		s.logger.Error("scheduler: create telego bot", "error", err, "bot_id", bot.ID)
		return
	}

	for _, client := range clients {
		if client.BirthDate == nil {
			continue
		}

		bd := *client.BirthDate
		if bd.Month() == today.Month() && bd.Day() == today.Day() {
			msg := s.personalizeMessage(scenario.Message, client)
			tgMsg := tu.Message(tu.ID(client.TelegramID), msg)
			if _, err := tBot.SendMessage(tgMsg); err != nil {
				s.logger.Warn("scheduler: birthday message failed",
					"client_id", client.ID, "error", err)
			} else {
				s.logger.Info("scheduler: birthday message sent",
					"client_id", client.ID, "bot_id", bot.ID)
			}
			time.Sleep(35 * time.Millisecond)
		}
	}
}

func (s *Scheduler) evaluateInactiveDays(ctx context.Context, bot entity.Bot, scenario entity.AutoScenario) {
	if scenario.TriggerConfig.Days == nil {
		return
	}

	days := *scenario.TriggerConfig.Days
	cutoff := time.Now().AddDate(0, 0, -days)

	clients := s.getAllBotClients(ctx, bot.ID)

	tBot, err := telego.NewBot(bot.Token)
	if err != nil {
		s.logger.Error("scheduler: create telego bot", "error", err, "bot_id", bot.ID)
		return
	}

	for _, client := range clients {
		if client.RegisteredAt.Before(cutoff) {
			msg := s.personalizeMessage(scenario.Message, client)
			tgMsg := tu.Message(tu.ID(client.TelegramID), msg)
			if _, err := tBot.SendMessage(tgMsg); err != nil {
				s.logger.Warn("scheduler: inactive message failed",
					"client_id", client.ID, "error", err)
			}
			time.Sleep(35 * time.Millisecond)
		}
	}
}

func (s *Scheduler) getAllBotClients(ctx context.Context, botID int) []entity.BotClient {
	var all []entity.BotClient
	limit := 100
	offset := 0

	for {
		clients, total, err := s.botClients.GetByBotID(ctx, botID, limit, offset)
		if err != nil {
			s.logger.Error("scheduler: get bot clients", "error", err, "bot_id", botID)
			return all
		}
		all = append(all, clients...)
		offset += limit
		if offset >= total {
			break
		}
	}

	return all
}

func (s *Scheduler) personalizeMessage(template string, client entity.BotClient) string {
	name := client.FirstName
	if name == "" {
		name = "клиент"
	}
	msg := strings.ReplaceAll(template, "{name}", name)
	msg = strings.ReplaceAll(msg, "{first_name}", name)
	return msg
}
