package clients

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	clientsUC "revisitr/internal/usecase/clients"
)

type clientsUsecase interface {
	List(ctx context.Context, orgID int, filter entity.ClientFilter) (*entity.PaginatedResponse[entity.ClientProfile], error)
	GetProfile(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error)
	UpdateTags(ctx context.Context, orgID, clientID int, req *entity.UpdateClientRequest) error
	GetStats(ctx context.Context, orgID int) (*entity.ClientStats, error)
	CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error)
	IdentifyClient(ctx context.Context, orgID int, phone, qrCode string) (*entity.BotClient, error)
}

type orderStatsUsecase interface {
	GetClientOrderStats(ctx context.Context, clientID int) (*entity.ClientOrderStats, error)
}

type Group struct {
	uc         clientsUsecase
	orderStats orderStatsUsecase
	jwtSecret  string
}

type Option func(*Group)

func WithOrderStats(uc orderStatsUsecase) Option {
	return func(g *Group) { g.orderStats = uc }
}

func New(uc clientsUsecase, jwtSecret string, opts ...Option) *Group {
	g := &Group{uc: uc, jwtSecret: jwtSecret}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Group) Path() string {
	return "/api/v1/clients"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	handlers := []func() (string, string, gin.HandlerFunc){
		g.handleList,
		g.handleStats,
		g.handleCount,
		g.handleIdentify,
		g.handleGet,
		g.handleUpdate,
	}
	if g.orderStats != nil {
		handlers = append(handlers, g.handleOrderStats)
	}
	return handlers
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var filter entity.ClientFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := g.uc.List(c.Request.Context(), orgID.(int), filter)
		if err != nil {
			slog.Error("list clients", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (g *Group) handleStats() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/stats", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		stats, err := g.uc.GetStats(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("get client stats", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (g *Group) handleCount() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/count", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var filter entity.ClientFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := g.uc.CountByFilter(c.Request.Context(), orgID.(int), filter)
		if err != nil {
			slog.Error("count clients", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"count": count})
	}
}

func (g *Group) handleIdentify() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/identify", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		phone := c.Query("phone")
		qrCode := c.Query("qr_code")

		if phone == "" && qrCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone or qr_code query parameter required"})
			return
		}

		client, err := g.uc.IdentifyClient(c.Request.Context(), orgID.(int), phone, qrCode)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, client)
	}
}

func (g *Group) handleGet() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
			return
		}

		profile, err := g.uc.GetProfile(c.Request.Context(), orgID.(int), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
			return
		}

		var req entity.UpdateClientRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateTags(c.Request.Context(), orgID.(int), id, &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "client updated"})
	}
}

func (g *Group) handleOrderStats() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/order-stats", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
			return
		}

		stats, err := g.orderStats.GetClientOrderStats(c.Request.Context(), id)
		if err != nil {
			slog.Error("client order stats", "error", err, "client_id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, clientsUC.ErrClientNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		slog.Error("client handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
