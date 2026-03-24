package dashboard

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
)

type salesUsecase interface {
	GetDashboardData(ctx context.Context, orgID int, from, to time.Time) (*entity.DashboardAggregates, error)
}

type dashboardUsecase interface {
	GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error)
	GetCharts(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardCharts, error)
}

type Group struct {
	uc        dashboardUsecase
	salesUC   salesUsecase
	jwtSecret string
}

func New(uc dashboardUsecase, jwtSecret string, opts ...func(*Group)) *Group {
	g := &Group{uc: uc, jwtSecret: jwtSecret}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func WithSalesUsecase(salesUC salesUsecase) func(*Group) {
	return func(g *Group) {
		g.salesUC = salesUC
	}
}

func (g *Group) Path() string {
	return "/api/v1/dashboard"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	handlers := []func() (string, string, gin.HandlerFunc){
		g.handleWidgets,
		g.handleCharts,
	}
	if g.salesUC != nil {
		handlers = append(handlers, g.handleSales)
	}
	return handlers
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

func (g *Group) handleSales() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/sales", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		from, err := time.Parse("2006-01-02", c.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02")))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date format, use YYYY-MM-DD"})
			return
		}

		to, err := time.Parse("2006-01-02", c.DefaultQuery("to", time.Now().Format("2006-01-02")))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date format, use YYYY-MM-DD"})
			return
		}

		data, err := g.salesUC.GetDashboardData(c.Request.Context(), orgID.(int), from, to)
		if err != nil {
			slog.Error("get dashboard sales", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, data)
	}
}
