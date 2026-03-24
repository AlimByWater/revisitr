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
	analyticsGroup "revisitr/internal/controller/http/group/analytics"
	authGroup "revisitr/internal/controller/http/group/auth"
	botsGroup "revisitr/internal/controller/http/group/bots"
	campaignsGroup "revisitr/internal/controller/http/group/campaigns"
	clientsGroup "revisitr/internal/controller/http/group/clients"
	dashboardGroup "revisitr/internal/controller/http/group/dashboard"
	filesGroup "revisitr/internal/controller/http/group/files"
	"revisitr/internal/controller/http/group/health"
	integrationsGroup "revisitr/internal/controller/http/group/integrations"
	loyaltyGroup "revisitr/internal/controller/http/group/loyalty"
	posGroup "revisitr/internal/controller/http/group/pos"
	promotionsGroup "revisitr/internal/controller/http/group/promotions"
	segmentsGroup "revisitr/internal/controller/http/group/segments"
	"revisitr/internal/controller/scheduler"
	minioRepo "revisitr/internal/repository/minio"
	pgRepo "revisitr/internal/repository/postgres"
	redisRepo "revisitr/internal/repository/redis"
	autoactionService "revisitr/internal/service/autoaction"
	posService "revisitr/internal/service/pos"
	rfmService "revisitr/internal/service/rfm"
	analyticsUC "revisitr/internal/usecase/analytics"
	authUC "revisitr/internal/usecase/auth"
	botsUC "revisitr/internal/usecase/bots"
	campaignsUC "revisitr/internal/usecase/campaigns"
	clientsUC "revisitr/internal/usecase/clients"
	dashboardUC "revisitr/internal/usecase/dashboard"
	integrationsUC "revisitr/internal/usecase/integrations"
	loyaltyUC "revisitr/internal/usecase/loyalty"
	posUC "revisitr/internal/usecase/pos"
	promotionsUC "revisitr/internal/usecase/promotions"
	segmentsUC "revisitr/internal/usecase/segments"
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

	// ── Repositories ──────────────────────────────────────────────────────────

	usersRepo := pgRepo.NewUsers(pg)
	sessionsRepo := redisRepo.NewSessions(rds)
	botsRepo := pgRepo.NewBots(pg)
	botClientsRepo := pgRepo.NewBotClients(pg)
	loyaltyRepo := pgRepo.NewLoyalty(pg)
	clientsRepo := pgRepo.NewClients(pg)
	dashboardRepo := pgRepo.NewDashboard(pg)
	campaignsRepo := pgRepo.NewCampaigns(pg)
	scenariosRepo := pgRepo.NewAutoScenarios(pg)
	posRepo := pgRepo.NewPOS(pg)

	// Phase 3 repos
	analyticsRepo := pgRepo.NewAnalytics(pg)
	segmentsRepo := pgRepo.NewSegments(pg)
	promotionsRepo := pgRepo.NewPromotions(pg)
	integrationsRepo := pgRepo.NewIntegrations(pg)

	// ── Services ──────────────────────────────────────────────────────────────

	posSyncSvc := posService.NewSyncService(integrationsRepo, botClientsRepo, logger)
	rfmSvc := rfmService.New(botClientsRepo, loyaltyRepo, logger)

	// Auto-action services
	actionExecutor := autoactionService.NewActionExecutor(scenariosRepo, logger)
	autoActionScheduler := autoactionService.NewAutoActionScheduler(scenariosRepo, actionExecutor, botClientsRepo, logger)

	// ── Usecases ──────────────────────────────────────────────────────────────

	authUsecase := authUC.New(&authConfig{cfg: cfg}, usersRepo, sessionsRepo)
	botsUsecase := botsUC.New(botsRepo, botClientsRepo)
	loyaltyUsecase := loyaltyUC.New(loyaltyRepo)
	clientsUsecase := clientsUC.New(clientsRepo, clientsUC.WithBotClients(botClientsRepo))
	dashboardUsecase := dashboardUC.New(dashboardRepo)
	campaignsUsecase := campaignsUC.New(campaignsRepo, scenariosRepo, clientsRepo)
	posUsecase := posUC.New(posRepo)

	// Phase 3 usecases
	analyticsUsecase := analyticsUC.New(analyticsRepo)
	segmentsUsecase := segmentsUC.New(segmentsRepo, clientsRepo)
	promotionsUsecase := promotionsUC.New(promotionsRepo)
	integrationsUsecase := integrationsUC.New(integrationsRepo, posSyncSvc)

	// ── MinIO ─────────────────────────────────────────────────────────────

	minioClient, err := minioRepo.New(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		logger.Error("minio init failed", "error", err)
		os.Exit(1)
	}

	jwtSecret := cfg.Auth.JWTSecret

	// ── Controller groups ─────────────────────────────────────────────────────

	healthGrp := health.New()
	authGrp := authGroup.New(authUsecase)
	botsGrp := botsGroup.New(botsUsecase, jwtSecret)
	loyaltyGrp := loyaltyGroup.New(loyaltyUsecase, jwtSecret)
	clientsGrp := clientsGroup.New(clientsUsecase, jwtSecret)
	dashboardGrp := dashboardGroup.New(dashboardUsecase, jwtSecret,
		dashboardGroup.WithSalesUsecase(integrationsUsecase),
	)
	campaignsGrp := campaignsGroup.New(campaignsUsecase, jwtSecret)
	posGrp := posGroup.New(posUsecase, jwtSecret)

	filesGrp := filesGroup.New(minioClient, cfg.MinIO.Bucket, jwtSecret)

	// Phase 3 groups
	analyticsGrp := analyticsGroup.New(analyticsUsecase, jwtSecret)
	segmentsGrp := segmentsGroup.New(segmentsUsecase, jwtSecret)
	promotionsGrp := promotionsGroup.New(promotionsUsecase, jwtSecret)
	integrationsGrp := integrationsGroup.New(integrationsUsecase, jwtSecret)

	httpModule := httpCtrl.New(&httpConfig{cfg: cfg},
		healthGrp, authGrp, botsGrp, loyaltyGrp, posGrp, clientsGrp,
		dashboardGrp, campaignsGrp,
		analyticsGrp, segmentsGrp, promotionsGrp, integrationsGrp,
		filesGrp,
	)

	// ── Scheduler ─────────────────────────────────────────────────────────────

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	sched := scheduler.New(logger)
	sched.Register(scheduler.Task{
		Name:     "refresh_materialized_views",
		Interval: 6 * time.Hour,
		Fn:       func(ctx context.Context) error { return analyticsRepo.RefreshMaterializedViews(ctx) },
	})
	sched.Register(scheduler.Task{
		Name:     "rfm_recalculate",
		Interval: 24 * time.Hour,
		Fn:       func(ctx context.Context) error { return rfmSvc.RecalculateAll(ctx, 0) },
	})
	sched.Register(scheduler.Task{
		Name:     "pos_sync",
		Interval: 4 * time.Hour,
		Fn:       func(ctx context.Context) error { return posSyncSvc.SyncAll(ctx) },
	})
	sched.Register(scheduler.Task{
		Name:     "loyalty_demotion",
		Interval: 24 * time.Hour,
		Fn:       func(ctx context.Context) error { return loyaltyUsecase.DemoteClients(ctx) },
	})
	sched.Register(scheduler.Task{
		Name:     "expire_reserves",
		Interval: 5 * time.Minute,
		Fn: func(ctx context.Context) error {
			n, err := loyaltyRepo.ExpireOldReserves(ctx)
			if err != nil {
				return err
			}
			if n > 0 {
				logger.Info("expired reserves", "count", n)
			}
			return nil
		},
	})
	sched.Register(scheduler.Task{
		Name:     "auto_actions_evaluate",
		Interval: 1 * time.Hour,
		Fn:       func(ctx context.Context) error { return autoActionScheduler.Evaluate(ctx) },
	})
	go sched.Run(appCtx)

	// ── Application ───────────────────────────────────────────────────────────

	app := application.New(
		logger,
		[]application.Repository{pg, rds},
		[]application.Usecase{
			authUsecase, botsUsecase, loyaltyUsecase, posUsecase,
			clientsUsecase, dashboardUsecase, campaignsUsecase,
			analyticsUsecase, segmentsUsecase, promotionsUsecase, integrationsUsecase,
		},
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

func (c *authConfig) GetJWTSecret() string       { return c.cfg.Auth.JWTSecret }
func (c *authConfig) GetTokenTTL() time.Duration { return c.cfg.Auth.TokenTTL }
