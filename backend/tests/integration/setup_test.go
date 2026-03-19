//go:build integration

package integration_test

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	analyticsGroup "revisitr/internal/controller/http/group/analytics"
	authGroup "revisitr/internal/controller/http/group/auth"
	botsGroup "revisitr/internal/controller/http/group/bots"
	campaignsGroup "revisitr/internal/controller/http/group/campaigns"
	clientsGroup "revisitr/internal/controller/http/group/clients"
	dashboardGroup "revisitr/internal/controller/http/group/dashboard"
	"revisitr/internal/controller/http/group/health"
	integrationsGroup "revisitr/internal/controller/http/group/integrations"
	loyaltyGroup "revisitr/internal/controller/http/group/loyalty"
	posGroup "revisitr/internal/controller/http/group/pos"
	promotionsGroup "revisitr/internal/controller/http/group/promotions"
	segmentsGroup "revisitr/internal/controller/http/group/segments"
	pgRepo "revisitr/internal/repository/postgres"
	redisRepo "revisitr/internal/repository/redis"
	posService "revisitr/internal/service/pos"
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

const (
	testJWTSecret = "integration-test-secret"
	testDBHost    = "localhost"
	testDBPort    = "5433"
	testDBUser    = "revisitr"
	testDBPass    = "devpassword"
	testDBName    = "revisitr"
	testDBSSL     = "disable"
	testRedisHost = "localhost"
	testRedisPort = "6380"

	// All test records use this email domain for easy cleanup
	testEmailDomain = "@test.revisitr.local"
)

var (
	srv    *httptest.Server
	pgMod  *pgRepo.Module
	rdsMod *redisRepo.Module
)

// ── config adapters ───────────────────────────────────────────────────────────

type pgCfg struct{}

func (pgCfg) GetHost() string     { return envOr("POSTGRES_HOST", testDBHost) }
func (pgCfg) GetPort() string     { return envOr("POSTGRES_PORT", testDBPort) }
func (pgCfg) GetUser() string     { return envOr("POSTGRES_USER", testDBUser) }
func (pgCfg) GetPassword() string { return envOr("POSTGRES_PASSWORD", testDBPass) }
func (pgCfg) GetDatabase() string { return envOr("POSTGRES_DATABASE", testDBName) }
func (pgCfg) GetSSLMode() string  { return testDBSSL }

type rdsCfg struct{}

func (rdsCfg) GetHost() string     { return envOr("REDIS_HOST", testRedisHost) }
func (rdsCfg) GetPort() string     { return envOr("REDIS_PORT", testRedisPort) }
func (rdsCfg) GetPassword() string { return envOr("REDIS_PASSWORD", "") }

type authCfg struct{}

func (authCfg) GetJWTSecret() string       { return testJWTSecret }
func (authCfg) GetTokenTTL() time.Duration { return time.Hour }

// ── TestMain ─────────────────────────────────────────────────────────────────

func TestMain(m *testing.M) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// PostgreSQL
	pgMod = pgRepo.New(pgCfg{})
	if err := pgMod.Init(ctx, logger); err != nil {
		log.Fatalf("postgres init: %v", err)
	}

	// Redis
	rdsMod = redisRepo.New(rdsCfg{})
	if err := rdsMod.Init(ctx, logger); err != nil {
		log.Fatalf("redis init: %v", err)
	}

	// Repositories
	usersRepo := pgRepo.NewUsers(pgMod)
	sessionsRepo := redisRepo.NewSessions(rdsMod)
	botsRepo := pgRepo.NewBots(pgMod)
	botClientsRepo := pgRepo.NewBotClients(pgMod)
	loyaltyRepo := pgRepo.NewLoyalty(pgMod)
	clientsRepo := pgRepo.NewClients(pgMod)
	dashboardRepo := pgRepo.NewDashboard(pgMod)
	campaignsRepo := pgRepo.NewCampaigns(pgMod)
	scenariosRepo := pgRepo.NewAutoScenarios(pgMod)
	posRepo := pgRepo.NewPOS(pgMod)
	analyticsRepo := pgRepo.NewAnalytics(pgMod)
	segmentsRepo := pgRepo.NewSegments(pgMod)
	promotionsRepo := pgRepo.NewPromotions(pgMod)
	integrationsRepo := pgRepo.NewIntegrations(pgMod)

	// Usecases
	authUsecase := authUC.New(authCfg{}, usersRepo, sessionsRepo)
	botsUsecase := botsUC.New(botsRepo, botClientsRepo)
	loyaltyUsecase := loyaltyUC.New(loyaltyRepo)
	clientsUsecase := clientsUC.New(clientsRepo)
	dashboardUsecase := dashboardUC.New(dashboardRepo)
	campaignsUsecase := campaignsUC.New(campaignsRepo, scenariosRepo, clientsRepo)
	posUsecase := posUC.New(posRepo)
	analyticsUsecase := analyticsUC.New(analyticsRepo)
	segmentsUsecase := segmentsUC.New(segmentsRepo, clientsRepo)
	promotionsUsecase := promotionsUC.New(promotionsRepo)
	posSyncSvc := posService.NewSyncService(integrationsRepo)
	integrationsUsecase := integrationsUC.New(integrationsRepo, posSyncSvc)

	for _, uc := range []interface{ Init(context.Context, *slog.Logger) error }{
		authUsecase, botsUsecase, loyaltyUsecase,
		clientsUsecase, dashboardUsecase, campaignsUsecase, posUsecase,
	} {
		if err := uc.Init(ctx, logger); err != nil {
			log.Fatalf("usecase init: %v", err)
		}
	}

	// Groups
	groups := []group{
		health.New(),
		authGroup.New(authUsecase),
		botsGroup.New(botsUsecase, testJWTSecret),
		loyaltyGroup.New(loyaltyUsecase, testJWTSecret),
		clientsGroup.New(clientsUsecase, testJWTSecret),
		dashboardGroup.New(dashboardUsecase, testJWTSecret),
		campaignsGroup.New(campaignsUsecase, testJWTSecret),
		posGroup.New(posUsecase, testJWTSecret),
		analyticsGroup.New(analyticsUsecase, testJWTSecret),
		segmentsGroup.New(segmentsUsecase, testJWTSecret),
		promotionsGroup.New(promotionsUsecase, testJWTSecret),
		integrationsGroup.New(integrationsUsecase, testJWTSecret),
	}

	// HTTP test server
	gin.SetMode(gin.TestMode)
	engine := buildEngine(groups...)
	srv = httptest.NewServer(engine)

	code := m.Run()

	srv.Close()
	pgMod.Close()
	rdsMod.Close()
	os.Exit(code)
}

// ── engine builder ────────────────────────────────────────────────────────────

type group interface {
	Path() string
	Handlers() []func() (string, string, gin.HandlerFunc)
	Auth() gin.HandlerFunc
}

func buildEngine(groups ...group) *gin.Engine {
	engine := gin.New()
	engine.RedirectTrailingSlash = false

	for _, g := range groups {
		rg := engine.Group(g.Path())
		if auth := g.Auth(); auth != nil {
			rg.Use(auth)
		}
		for _, fn := range g.Handlers() {
			method, path, handler := fn()
			switch method {
			case http.MethodGet:
				rg.GET(path, handler)
			case http.MethodPost:
				rg.POST(path, handler)
			case http.MethodPut:
				rg.PUT(path, handler)
			case http.MethodPatch:
				rg.PATCH(path, handler)
			case http.MethodDelete:
				rg.DELETE(path, handler)
			default:
				rg.Handle(method, path, handler)
			}
		}
	}

	return engine
}

// ── helpers ───────────────────────────────────────────────────────────────────

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
