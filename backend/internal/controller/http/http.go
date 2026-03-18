package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"revisitr/internal/controller/http/middleware"

	"github.com/gin-gonic/gin"
)

type config interface {
	GetPort() string
	GetJWTSecret() string
}

type group interface {
	Path() string
	Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc)
	Auth() gin.HandlerFunc
}

type Module struct {
	cfg    config
	server *http.Server
	groups []group
	logger *slog.Logger
}

func New(cfg config, groups ...group) *Module {
	return &Module{
		cfg:    cfg,
		groups: groups,
	}
}

func (m *Module) Init(_ context.Context, stop context.CancelFunc, logger *slog.Logger) error {
	m.logger = logger

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.RedirectTrailingSlash = false

	engine.Use(
		middleware.Recovery(logger),
		middleware.CORS(),
		middleware.Logger(logger),
	)

	for _, g := range m.groups {
		rg := engine.Group(g.Path())
		authMW := g.Auth()
		if authMW != nil {
			rg.Use(authMW)
		}

		for _, handlerFn := range g.Handlers() {
			method, path, handler := handlerFn()
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

	m.server = &http.Server{
		Addr:         fmt.Sprintf(":%s", m.cfg.GetPort()),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

func (m *Module) Run() {
	m.logger.Info("http server starting", "addr", m.server.Addr)
	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		m.logger.Error("http server error", "error", err)
	}
}

func (m *Module) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m.logger.Info("http server shutting down")
	return m.server.Shutdown(ctx)
}
