package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"revisitr/internal/application"
	"revisitr/internal/application/config"
	"revisitr/internal/application/env"
	httpCtrl "revisitr/internal/controller/http"
	authGroup "revisitr/internal/controller/http/group/auth"
	botsGroup "revisitr/internal/controller/http/group/bots"
	campaignsGroup "revisitr/internal/controller/http/group/campaigns"
	clientsGroup "revisitr/internal/controller/http/group/clients"
	dashboardGroup "revisitr/internal/controller/http/group/dashboard"
	"revisitr/internal/controller/http/group/health"
	loyaltyGroup "revisitr/internal/controller/http/group/loyalty"
	posGroup "revisitr/internal/controller/http/group/pos"
	pgRepo "revisitr/internal/repository/postgres"
	redisRepo "revisitr/internal/repository/redis"
	authUC "revisitr/internal/usecase/auth"
	botsUC "revisitr/internal/usecase/bots"
	loyaltyUC "revisitr/internal/usecase/loyalty"
	campaignsUC "revisitr/internal/usecase/campaigns"
	clientsUC "revisitr/internal/usecase/clients"
	dashboardUC "revisitr/internal/usecase/dashboard"
	posUC "revisitr/internal/usecase/pos"
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

	usersRepo := pgRepo.NewUsers(pg)
	sessionsRepo := redisRepo.NewSessions(rds)

	authUsecase := authUC.New(&authConfig{cfg: cfg}, usersRepo, sessionsRepo)

	botsRepo := pgRepo.NewBots(pg)
	botClientsRepo := pgRepo.NewBotClients(pg)
	botsUsecase := botsUC.New(botsRepo, botClientsRepo)

	loyaltyRepo := pgRepo.NewLoyalty(pg)
	loyaltyUsecase := loyaltyUC.New(loyaltyRepo)

	clientsRepo := pgRepo.NewClients(pg)
	clientsUsecase := clientsUC.New(clientsRepo)

	dashboardRepo := pgRepo.NewDashboard(pg)
	dashboardUsecase := dashboardUC.New(dashboardRepo)

	campaignsRepo := pgRepo.NewCampaigns(pg)
	scenariosRepo := pgRepo.NewAutoScenarios(pg)
	campaignsUsecase := campaignsUC.New(campaignsRepo, scenariosRepo, clientsRepo)

	posRepo := pgRepo.NewPOS(pg)
	posUsecase := posUC.New(posRepo)

	jwtSecret := cfg.Auth.JWTSecret

	healthGrp := health.New()
	authGrp := authGroup.New(authUsecase)
	botsGrp := botsGroup.New(botsUsecase, jwtSecret)
	loyaltyGrp := loyaltyGroup.New(loyaltyUsecase, jwtSecret)
	clientsGrp := clientsGroup.New(clientsUsecase, jwtSecret)
	dashboardGrp := dashboardGroup.New(dashboardUsecase, jwtSecret)
	campaignsGrp := campaignsGroup.New(campaignsUsecase, jwtSecret)
	posGrp := posGroup.New(posUsecase, jwtSecret)
	httpModule := httpCtrl.New(&httpConfig{cfg: cfg}, healthGrp, authGrp, botsGrp, loyaltyGrp, posGrp, clientsGrp, dashboardGrp, campaignsGrp)

	app := application.New(
		logger,
		[]application.Repository{pg, rds},
		[]application.Usecase{authUsecase, botsUsecase, loyaltyUsecase, posUsecase, clientsUsecase, dashboardUsecase, campaignsUsecase},
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

type authConfig struct{ cfg *config.Module }

func (c *authConfig) GetJWTSecret() string    { return c.cfg.Auth.JWTSecret }
func (c *authConfig) GetTokenTTL() time.Duration { return c.cfg.Auth.TokenTTL }
