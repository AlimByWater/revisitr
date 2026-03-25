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
	"revisitr/internal/service/adminbot"
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

	// Repositories
	adminBotRepo := pgRepo.NewAdminBot(pg)
	dashboardRepo := pgRepo.NewDashboard(pg)
	campaignsRepo := pgRepo.NewCampaigns(pg)
	promotionsRepo := pgRepo.NewPromotions(pg)

	// Create and start admin bot service
	svc, err := adminbot.New(
		cfg.AdminBot.Token,
		adminBotRepo,
		dashboardRepo,
		campaignsRepo,
		promotionsRepo,
		nil, // analyticsRepo — placeholder for future use
		logger,
	)
	if err != nil {
		logger.Error("admin bot init failed", "error", err)
		os.Exit(1)
	}

	if err := svc.Start(ctx); err != nil {
		logger.Error("admin bot start failed", "error", err)
		os.Exit(1)
	}

	logger.Info("admin bot service started")

	<-ctx.Done()
	logger.Info("shutting down admin bot service")
	svc.Shutdown()
	logger.Info("admin bot service stopped")
}

type postgresConfig struct{ cfg *config.Module }

func (c *postgresConfig) GetHost() string     { return c.cfg.Postgres.Host }
func (c *postgresConfig) GetPort() string     { return c.cfg.Postgres.Port }
func (c *postgresConfig) GetUser() string     { return c.cfg.Postgres.User }
func (c *postgresConfig) GetPassword() string { return c.cfg.Postgres.Password }
func (c *postgresConfig) GetDatabase() string { return c.cfg.Postgres.Database }
func (c *postgresConfig) GetSSLMode() string  { return c.cfg.Postgres.SSLMode }
