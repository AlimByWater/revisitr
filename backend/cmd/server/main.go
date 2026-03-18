package main

import (
	"context"
	"log/slog"
	"os"

	"revisitr/internal/application"
	"revisitr/internal/application/config"
	"revisitr/internal/application/env"
	httpCtrl "revisitr/internal/controller/http"
	"revisitr/internal/controller/http/group/health"
	pgRepo "revisitr/internal/repository/postgres"
	redisRepo "revisitr/internal/repository/redis"
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

	pg := pgRepo.New(&postgresConfig{cfg: cfg})
	rds := redisRepo.New(&redisConfig{cfg: cfg})

	healthGroup := health.New()
	httpModule := httpCtrl.New(&httpConfig{cfg: cfg}, healthGroup)

	app := application.New(
		logger,
		[]application.Repository{pg, rds},
		[]application.Usecase{},
		[]application.Controller{httpModule},
	)

	if err := app.Run(context.Background()); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
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

type httpConfig struct{ cfg *config.Module }

func (c *httpConfig) GetPort() string      { return c.cfg.Http.Port }
func (c *httpConfig) GetJWTSecret() string { return c.cfg.Auth.JWTSecret }
