package bots

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	botsUC "revisitr/internal/usecase/bots"
)

type botsUsecase interface {
	Create(ctx context.Context, orgID int, req *entity.CreateBotRequest) (*entity.Bot, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Bot, error)
	GetByID(ctx context.Context, id, orgID int) (*entity.Bot, error)
	Update(ctx context.Context, id, orgID int, req *entity.UpdateBotRequest) (*entity.Bot, error)
	Delete(ctx context.Context, id, orgID int) error
	GetSettings(ctx context.Context, id, orgID int) (*entity.BotSettings, error)
	UpdateSettings(ctx context.Context, id, orgID int, req *entity.UpdateBotSettingsRequest) error
}

type posLocationsUsecase interface {
	SetBotPOSLocations(ctx context.Context, botID int, posIDs []int) error
	GetBotPOSLocations(ctx context.Context, botID int) ([]int, error)
}

type Group struct {
	uc        botsUsecase
	posLoc    posLocationsUsecase
	jwtSecret string
}

type Option func(*Group)

func WithPOSLocations(uc posLocationsUsecase) Option {
	return func(g *Group) { g.posLoc = uc }
}

func New(uc botsUsecase, jwtSecret string, opts ...Option) *Group {
	g := &Group{uc: uc, jwtSecret: jwtSecret}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Group) Path() string {
	return "/api/v1/bots"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	handlers := []func() (string, string, gin.HandlerFunc){
		g.handleCreate,
		g.handleList,
		g.handleGet,
		g.handleUpdate,
		g.handleDelete,
		g.handleGetSettings,
		g.handleUpdateSettings,
	}
	if g.posLoc != nil {
		handlers = append(handlers, g.handleGetPOSLocations, g.handleSetPOSLocations)
	}
	return handlers
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateBotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		bot, err := g.uc.Create(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create bot", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, bot)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		bots, err := g.uc.GetByOrgID(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list bots", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, bots)
	}
}

func (g *Group) handleGet() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		bot, err := g.uc.GetByID(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, bot)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		var req entity.UpdateBotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		bot, err := g.uc.Update(c.Request.Context(), id, orgID.(int), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, bot)
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), id, orgID.(int)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "bot deleted"})
	}
}

func (g *Group) handleGetSettings() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/settings", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		settings, err := g.uc.GetSettings(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, settings)
	}
}

func (g *Group) handleUpdateSettings() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id/settings", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		var req entity.UpdateBotSettingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateSettings(c.Request.Context(), id, orgID.(int), &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
	}
}

func (g *Group) handleGetPOSLocations() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/pos-locations", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		posIDs, err := g.posLoc.GetBotPOSLocations(c.Request.Context(), id)
		if err != nil {
			slog.Error("get bot pos locations", "error", err, "bot_id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"pos_ids": posIDs})
	}
}

func (g *Group) handleSetPOSLocations() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/:id/pos-locations", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		var req struct {
			POSIDs []int `json:"pos_ids" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.posLoc.SetBotPOSLocations(c.Request.Context(), id, req.POSIDs); err != nil {
			slog.Error("set bot pos locations", "error", err, "bot_id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "pos locations updated"})
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, botsUC.ErrBotNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, botsUC.ErrNotBotOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		slog.Error("bot handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
