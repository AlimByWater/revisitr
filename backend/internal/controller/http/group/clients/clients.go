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
}

type Group struct {
	uc        clientsUsecase
	jwtSecret string
}

func New(uc clientsUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/clients"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleList,
		g.handleStats,
		g.handleCount,
		g.handleGet,
		g.handleUpdate,
	}
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

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, clientsUC.ErrClientNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		slog.Error("client handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
