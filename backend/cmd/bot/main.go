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
	"revisitr/internal/service/botmanager"
	"revisitr/internal/service/campaign"
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

	// Create repositories
	botsRepo := pgRepo.NewBots(pg)
	botClientsRepo := pgRepo.NewBotClients(pg)
	loyaltyRepo := pgRepo.NewLoyalty(pg)
	posRepo := pgRepo.NewPOS(pg)

	campaignsRepo := pgRepo.NewCampaigns(pg)
	scenariosRepo := pgRepo.NewAutoScenarios(pg)

	// Create and start bot manager
	mgr := botmanager.New(botsRepo, botClientsRepo, loyaltyRepo, posRepo, logger)

	if err := mgr.Start(ctx); err != nil {
		logger.Error("bot manager start failed", "error", err)
		os.Exit(1)
	}

	logger.Info("bot service started", "active_bots", mgr.ActiveCount())

	// Start auto-scenario scheduler
	scheduler := campaign.NewScheduler(scenariosRepo, botsRepo, botClientsRepo, loyaltyRepo, logger)
	go scheduler.Run(ctx)

	// Campaign sender available for use
	_ = campaign.NewSender(campaignsRepo, botsRepo, botClientsRepo, logger)

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
