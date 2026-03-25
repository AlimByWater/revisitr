package analytics

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	analyticsUC "revisitr/internal/usecase/analytics"
)

type analyticsUsecase interface {
	GetSalesAnalytics(ctx context.Context, orgID int, f entity.AnalyticsFilter) (*entity.SalesAnalytics, error)
	GetLoyaltyAnalytics(ctx context.Context, orgID int, f entity.AnalyticsFilter) (*entity.LoyaltyAnalytics, error)
	GetCampaignAnalytics(ctx context.Context, orgID int, f entity.AnalyticsFilter) (*entity.CampaignAnalytics, error)
}

type Option func(*Group)

func WithFeatureGate(gate gin.HandlerFunc) Option {
	return func(g *Group) { g.featureGate = gate }
}

type Group struct {
	uc          analyticsUsecase
	jwtSecret   string
	featureGate gin.HandlerFunc
}

func New(uc analyticsUsecase, jwtSecret string, opts ...Option) *Group {
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
	return "/api/v1/analytics"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleSales,
		g.handleLoyalty,
		g.handleCampaigns,
	}
}

func (g *Group) handleSales() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/sales", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		f := parseFilter(c)

		result, err := g.uc.GetSalesAnalytics(c.Request.Context(), orgID.(int), f)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (g *Group) handleLoyalty() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/loyalty", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		f := parseFilter(c)

		result, err := g.uc.GetLoyaltyAnalytics(c.Request.Context(), orgID.(int), f)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (g *Group) handleCampaigns() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/campaigns", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		f := parseFilter(c)

		result, err := g.uc.GetCampaignAnalytics(c.Request.Context(), orgID.(int), f)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func parseFilter(c *gin.Context) entity.AnalyticsFilter {
	var f entity.AnalyticsFilter

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			f.From = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			f.To = t.Add(24*time.Hour - time.Second)
		}
	}
	if botIDStr := c.Query("bot_id"); botIDStr != "" {
		if id, err := strconv.Atoi(botIDStr); err == nil {
			f.BotID = &id
		}
	}
	if segIDStr := c.Query("segment_id"); segIDStr != "" {
		if id, err := strconv.Atoi(segIDStr); err == nil {
			f.SegmentID = &id
		}
	}

	return f
}

func handleError(c *gin.Context, err error) {
	switch err {
	case analyticsUC.ErrInvalidDateRange:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		slog.Error("analytics handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
