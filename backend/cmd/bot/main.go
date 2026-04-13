package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"revisitr/internal/application/config"
	"revisitr/internal/application/env"
	pgRepo "revisitr/internal/repository/postgres"
	redisRepo "revisitr/internal/repository/redis"
	"revisitr/internal/service/botmanager"
	"revisitr/internal/service/campaign"
	"revisitr/internal/service/eventbus"
	tgService "revisitr/internal/service/telegram"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	envModule := &env.Module{}
	if err := envModule.Init(); err != nil {
		logger.Warn("no .env file loaded, using environment variables", "error", err)
	}

	cfg := config.NewFromEnv()

	// Init PostgreSQL
	pg := pgRepo.New(&postgresConfig{cfg: cfg})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := pg.Init(ctx, logger); err != nil {
		logger.Error("postgres init failed", "error", err)
		os.Exit(1)
	}
	defer pg.Close()

	// Init Redis
	rds := redisRepo.New(&redisConfig{cfg: cfg})
	if err := rds.Init(ctx, logger); err != nil {
		logger.Error("redis init failed", "error", err)
		os.Exit(1)
	}
	defer rds.Close()

	// Create repositories
	botsRepo := pgRepo.NewBots(pg)
	botClientsRepo := pgRepo.NewBotClients(pg)
	loyaltyRepo := pgRepo.NewLoyalty(pg)
	posRepo := pgRepo.NewPOS(pg)

	campaignsRepo := pgRepo.NewCampaigns(pg)
	scenariosRepo := pgRepo.NewAutoScenarios(pg)

	// Create Telegram sender for rich messages
	baseURL := cfg.GetBaseURL()
	tgSender := tgService.NewSender(baseURL, logger)

	// Create and start bot manager
	var mgrOpts []botmanager.ManagerOption
	mgrOpts = append(mgrOpts, botmanager.WithTelegramSender(tgSender))
	if cfg.TelegramAPIURL != "" {
		logger.Info("using custom Telegram API server", "url", cfg.TelegramAPIURL)
		mgrOpts = append(mgrOpts, botmanager.WithAPIServer(cfg.TelegramAPIURL))
	}
	mgr := botmanager.New(botsRepo, botClientsRepo, loyaltyRepo, posRepo, logger, mgrOpts...)

	if err := mgr.Start(ctx); err != nil {
		logger.Error("bot manager start failed", "error", err)
		os.Exit(1)
	}

	logger.Info("bot service started", "active_bots", mgr.ActiveCount())

	// Start event bus subscriber (listens for hot reload events from API server)
	subscriber := eventbus.NewSubscriber(rds.Client(), logger)
	go subscriber.Listen(ctx, mgr)

	// Start auto-scenario scheduler
	scheduler := campaign.NewScheduler(scenariosRepo, botsRepo, botClientsRepo, loyaltyRepo, logger)
	go scheduler.Run(ctx)

	// Campaign sender with rich message support
	_ = campaign.NewSender(campaignsRepo, botsRepo, botClientsRepo, logger,
		campaign.WithTelegramSender(tgSender),
	)

	<-ctx.Done()
	logger.Info("shutting down bot service")
	mgr.Shutdown()
	logger.Info("bot service stopped")
}

type postgresConfig struct{ cfg *config.Module }

func (c *postgresConfig) GetHost() string     { return c.cfg.Postgres.Host }
func (c *postgresConfig) GetPort() string     { return c.cfg.Postgres.Port }
func (c *postgresConfig) GetUser() string     { return c.cfg.Postgres.User }
func (c *postgresConfig) GetPassword() string { return c.cfg.Postgres.Password }
func (c *postgresConfig) GetDatabase() string { return c.cfg.Postgres.Database }
func (c *postgresConfig) GetSSLMode() string  { return c.cfg.Postgres.SSLMode }

type redisConfig struct{ cfg *config.Module }

func (c *redisConfig) GetHost() string     { return c.cfg.Redis.Host }
func (c *redisConfig) GetPort() string     { return c.cfg.Redis.Port }
func (c *redisConfig) GetPassword() string { return c.cfg.Redis.Password }
