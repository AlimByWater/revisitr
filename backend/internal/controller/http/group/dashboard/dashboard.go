package dashboard

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
)

type dashboardUsecase interface {
	GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error)
	GetCharts(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardCharts, error)
}

type Group struct {
	uc        dashboardUsecase
	jwtSecret string
}

func New(uc dashboardUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/dashboard"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleWidgets,
		g.handleCharts,
	}
}

func (g *Group) handleWidgets() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/widgets", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var filter entity.DashboardFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		widgets, err := g.uc.GetWidgets(c.Request.Context(), orgID.(int), filter)
		if err != nil {
			slog.Error("get dashboard widgets", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, widgets)
	}
}

func (g *Group) handleCharts() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/charts", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var filter entity.DashboardFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		charts, err := g.uc.GetCharts(c.Request.Context(), orgID.(int), filter)
		if err != nil {
			slog.Error("get dashboard charts", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, charts)
	}
}
