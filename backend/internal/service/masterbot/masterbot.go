package masterbot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
)

// Service is the Revisitr master bot (@revisitrbot) — creates managed bots,
// campaign posts, and provides venue management from Telegram.
type Service struct {
	bot    *telego.Bot
	h      *handler
	logger *slog.Logger
	cancel context.CancelFunc
}

type Config struct {
	Token       string
	BotUsername string // filled after GetMe
}

type Deps struct {
	// Legacy (admin bot features)
	AdminLinks adminBotLinksRepo
	Dashboard  dashboardRepository
	Campaigns  campaignsRepository
	Promotions promotionsRepository

	// New (master bot features)
	MasterLinks masterBotLinksRepo
	AuthTokens  masterBotAuthRepo
	Bots        botsRepository
}

func New(cfg Config, deps Deps, logger *slog.Logger) (*Service, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("masterbot: MASTER_BOT_TOKEN is not set")
	}

	tBot, err := telego.NewBot(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("masterbot: create telego bot: %w", err)
	}

	info, err := tBot.GetMe(context.Background())
	if err != nil {
		return nil, fmt.Errorf("masterbot: verify token: %w", err)
	}

	cfg.BotUsername = info.Username
	logger.Info("master bot initialized", "username", info.Username)

	h := newHandler(tBot, cfg, deps, logger)

	return &Service{
		bot:    tBot,
		h:      h,
		logger: logger,
	}, nil
}

func (s *Service) Start(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel

	updates, err := s.bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		cancel()
		return fmt.Errorf("masterbot: start long polling: %w", err)
	}

	s.logger.Info("master bot started, listening for updates")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				s.h.Handle(ctx, update)
			}
		}
	}()

	return nil
}

func (s *Service) Shutdown() {
	if s.cancel != nil {
		s.cancel()
	}
	s.logger.Info("master bot stopped")
}
