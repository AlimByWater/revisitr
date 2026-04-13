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
	"revisitr/internal/service/masterbot"
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
	redis := redisRepo.New(&redisConfig{cfg: cfg})
	if err := redis.Init(ctx, logger); err != nil {
		logger.Error("redis init failed", "error", err)
		os.Exit(1)
	}
	defer redis.Close()

	// Create and start master bot service
	svc, err := masterbot.New(
		masterbot.Config{Token: cfg.MasterBot.Token},
		masterbot.Deps{
			AdminLinks:  pgRepo.NewAdminBot(pg),
			Dashboard:   pgRepo.NewDashboard(pg),
			Campaigns:   pgRepo.NewCampaigns(pg),
			Promotions:  pgRepo.NewPromotions(pg),
			MasterLinks: pgRepo.NewMasterBot(pg),
			AuthTokens:  redisRepo.NewMasterBotAuth(redis),
			Bots:        pgRepo.NewBots(pg),
			PostCodes:   pgRepo.NewPostCodes(pg),
		},
		logger,
	)
	if err != nil {
		logger.Error("master bot init failed", "error", err)
		os.Exit(1)
	}

	if err := svc.Start(ctx); err != nil {
		logger.Error("master bot start failed", "error", err)
		os.Exit(1)
	}

	logger.Info("master bot service started")

	<-ctx.Done()
	logger.Info("shutting down master bot service")
	svc.Shutdown()
	logger.Info("master bot service stopped")
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
