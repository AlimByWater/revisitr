package rfm

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
)

type rfmUsecase interface {
	GetDashboard(ctx context.Context, orgID int) (*entity.RFMDashboard, error)
	Recalculate(ctx context.Context, orgID int) error
	GetConfig(ctx context.Context, orgID int) (*entity.RFMConfig, error)
	UpdateConfig(ctx context.Context, orgID int, req entity.UpdateRFMConfigRequest) (*entity.RFMConfig, error)
}

type Option func(*Group)

func WithFeatureGate(gate gin.HandlerFunc) Option {
	return func(g *Group) { g.featureGate = gate }
}

type Group struct {
	uc          rfmUsecase
	jwtSecret   string
	featureGate gin.HandlerFunc
}

func New(uc rfmUsecase, jwtSecret string, opts ...Option) *Group {
	g := &Group{uc: uc, jwtSecret: jwtSecret}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Group) ExtraMiddleware() []gin.HandlerFunc {
	if g.featureGate != nil {
		return []gin.HandlerFunc{g.featureGate}
	}
	return nil
}

func (g *Group) Path() string {
	return "/api/v1/rfm"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleDashboard,
		g.handleRecalculate,
		g.handleGetConfig,
		g.handleUpdateConfig,
	}
}

func (g *Group) handleDashboard() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/dashboard", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		dashboard, err := g.uc.GetDashboard(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("rfm dashboard", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, dashboard)
	}
}

func (g *Group) handleRecalculate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/recalculate", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		if err := g.uc.Recalculate(c.Request.Context(), orgID.(int)); err != nil {
			slog.Error("rfm recalculate", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "recalculation started"})
	}
}

func (g *Group) handleGetConfig() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/config", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		cfg, err := g.uc.GetConfig(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("rfm get config", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, cfg)
	}
}

func (g *Group) handleUpdateConfig() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/config", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.UpdateRFMConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cfg, err := g.uc.UpdateConfig(c.Request.Context(), orgID.(int), req)
		if err != nil {
			slog.Error("rfm update config", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, cfg)
	}
}
