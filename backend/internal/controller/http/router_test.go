package http_test

import (
	"testing"

	"github.com/gin-gonic/gin"

	httpCtrl "revisitr/internal/controller/http"
	"revisitr/internal/controller/http/group/masterbot"
	"revisitr/internal/controller/http/group/analytics"
	"revisitr/internal/controller/http/group/auth"
	"revisitr/internal/controller/http/group/billing"
	"revisitr/internal/controller/http/group/bots"
	"revisitr/internal/controller/http/group/campaigns"
	"revisitr/internal/controller/http/group/clients"
	"revisitr/internal/controller/http/group/dashboard"
	"revisitr/internal/controller/http/group/files"
	"revisitr/internal/controller/http/group/health"
	"revisitr/internal/controller/http/group/integrations"
	"revisitr/internal/controller/http/group/loyalty"
	"revisitr/internal/controller/http/group/marketplace"
	"revisitr/internal/controller/http/group/menus"
	"revisitr/internal/controller/http/group/onboarding"
	"revisitr/internal/controller/http/group/pos"
	"revisitr/internal/controller/http/group/promotions"
	"revisitr/internal/controller/http/group/rfm"
	"revisitr/internal/controller/http/group/segments"
	"revisitr/internal/controller/http/group/wallet"
)

// stubConfig satisfies the unexported config interface required by httpCtrl.New.
type stubConfig struct{}

func (stubConfig) GetPort() string      { return "0" }
func (stubConfig) GetJWTSecret() string { return "test-secret" }

// TestAllGroupsRegisterWithoutPanic verifies that every controller group can be
// registered on a single Gin engine without triggering route parameter conflicts
// or any other panic. This is a pure unit test — no database, Redis, or external
// services are required because we only exercise route registration, never invoke
// the actual handlers.
func TestAllGroupsRegisterWithoutPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const jwtSecret = "test-secret"

	// Construct all 20 groups with nil usecases / repos / storage.
	// nil is valid here because the constructors only store the dependency;
	// it is never dereferenced during route registration.
	groups := []interface{}{
		health.New(),
		auth.New(nil),
		bots.New(nil, jwtSecret),
		loyalty.New(nil, jwtSecret),
		clients.New(nil, jwtSecret),
		dashboard.New(nil, jwtSecret),
		campaigns.New(nil, jwtSecret),
		pos.New(nil, jwtSecret),
		analytics.New(nil, jwtSecret),
		segments.New(nil, jwtSecret),
		promotions.New(nil, jwtSecret),
		integrations.New(nil, jwtSecret),
		files.New(nil, "test-bucket", jwtSecret),
		menus.New(nil, jwtSecret),
		rfm.New(nil, jwtSecret),
		onboarding.New(nil, jwtSecret),
		billing.New(nil, jwtSecret),
		masterbot.New(nil, jwtSecret),
		wallet.New(nil, jwtSecret),
		marketplace.New(nil, jwtSecret),
	}

	if len(groups) != 20 {
		t.Fatalf("expected 20 controller groups, got %d", len(groups))
	}

	// httpCtrl.New accepts variadic group interface values.
	// We need to convert our slice to the right call signature.
	// Since httpCtrl.New takes ...group (unexported), we replicate
	// the registration logic directly on a gin.Engine instead.
	engine := gin.New()

	for _, g := range groups {
		type groupI interface {
			Path() string
			Handlers() []func() (string, string, gin.HandlerFunc)
			Auth() gin.HandlerFunc
		}
		type groupWithPublicHandlers interface {
			PublicHandlers() []func() (string, string, gin.HandlerFunc)
		}
		type groupWithMiddleware interface {
			ExtraMiddleware() []gin.HandlerFunc
		}

		gi := g.(groupI)

		// Register public handlers (unauthenticated) first — mirrors http.go
		if pub, ok := g.(groupWithPublicHandlers); ok {
			pubGroup := engine.Group(gi.Path())
			for _, handlerFn := range pub.PublicHandlers() {
				method, path, handler := handlerFn()
				registerRoute(t, pubGroup, method, path, handler)
			}
		}

		rg := engine.Group(gi.Path())
		authMW := gi.Auth()
		if authMW != nil {
			rg.Use(authMW)
		}

		if mw, ok := g.(groupWithMiddleware); ok {
			for _, m := range mw.ExtraMiddleware() {
				rg.Use(m)
			}
		}

		for _, handlerFn := range gi.Handlers() {
			method, path, handler := handlerFn()
			registerRoute(t, rg, method, path, handler)
		}
	}

	routes := engine.Routes()
	if len(routes) == 0 {
		t.Fatal("expected at least one registered route, got 0")
	}

	t.Logf("successfully registered %d routes across %d groups", len(routes), len(groups))

	// Log all routes for debugging convenience.
	for _, r := range routes {
		t.Logf("  %s %s", r.Method, r.Path)
	}
}

// registerRoute mirrors the production registerRoute in http.go.
func registerRoute(t *testing.T, rg *gin.RouterGroup, method, path string, handler gin.HandlerFunc) {
	t.Helper()
	switch method {
	case "GET":
		rg.GET(path, handler)
	case "POST":
		rg.POST(path, handler)
	case "PUT":
		rg.PUT(path, handler)
	case "PATCH":
		rg.PATCH(path, handler)
	case "DELETE":
		rg.DELETE(path, handler)
	default:
		rg.Handle(method, path, handler)
	}
}

// Verify that _ = httpCtrl.New compiles (ensures the package is importable).
var _ = httpCtrl.New
