package adminbot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
)

// Service is the admin Telegram bot — a single-instance bot
// for business owners/managers to manage their venue from Telegram.
type Service struct {
	bot    *telego.Bot
	h      *handler
	logger *slog.Logger
	cancel context.CancelFunc
}

func New(
	token string,
	linksRepo adminBotLinksRepo,
	dashboardRepo dashboardRepository,
	campaignsRepo campaignsRepository,
	promotionsRepo promotionsRepository,
	_ interface{}, // reserved for future analytics repo
	logger *slog.Logger,
) (*Service, error) {
	if token == "" {
		return nil, fmt.Errorf("adminbot: ADMIN_BOT_TOKEN is not set")
	}

	tBot, err := telego.NewBot(token)
	if err != nil {
		return nil, fmt.Errorf("adminbot: create telego bot: %w", err)
	}

	info, err := tBot.GetMe()
	if err != nil {
		return nil, fmt.Errorf("adminbot: verify token: %w", err)
	}

	logger.Info("admin bot initialized", "username", info.Username)

	h := newHandler(tBot, linksRepo, dashboardRepo, campaignsRepo, promotionsRepo, nil, logger)

	return &Service{
		bot:    tBot,
		h:      h,
		logger: logger,
	}, nil
}

func (s *Service) Start(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel

	updates, err := s.bot.UpdatesViaLongPolling(nil)
	if err != nil {
		cancel()
		return fmt.Errorf("adminbot: start long polling: %w", err)
	}

	s.logger.Info("admin bot started, listening for updates")

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
	s.bot.StopLongPolling()
	s.logger.Info("admin bot stopped")
}
