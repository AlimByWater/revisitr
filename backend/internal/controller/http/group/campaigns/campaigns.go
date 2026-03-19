package campaigns

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	campaignsUC "revisitr/internal/usecase/campaigns"
)

type campaignsUsecase interface {
	Create(ctx context.Context, orgID int, req *entity.CreateCampaignRequest) (*entity.Campaign, error)
	List(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error)
	GetByID(ctx context.Context, orgID, id int) (*entity.Campaign, error)
	Update(ctx context.Context, orgID, id int, req *entity.UpdateCampaignRequest) error
	Delete(ctx context.Context, orgID, id int) error
	Send(ctx context.Context, orgID, id int) error
	PreviewAudience(ctx context.Context, orgID int, filter entity.AudienceFilter) (int, error)
	CreateScenario(ctx context.Context, orgID int, req *entity.CreateScenarioRequest) (*entity.AutoScenario, error)
	ListScenarios(ctx context.Context, orgID int) ([]entity.AutoScenario, error)
	UpdateScenario(ctx context.Context, orgID, id int, req *entity.UpdateScenarioRequest) error
	DeleteScenario(ctx context.Context, orgID, id int) error
	ToggleScenario(ctx context.Context, orgID, id int, active bool) error
}

type Group struct {
	uc        campaignsUsecase
	jwtSecret string
}

func New(uc campaignsUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/campaigns"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		// Static routes first
		g.handleCreate,
		g.handleList,
		g.handlePreviewAudience,
		g.handleListScenarios,
		g.handleCreateScenario,
		// Parameterized scenario routes
		g.handleUpdateScenario,
		g.handleDeleteScenario,
		// Parameterized campaign routes
		g.handleGet,
		g.handleUpdate,
		g.handleDelete,
		g.handleSend,
	}
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateCampaignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		campaign, err := g.uc.Create(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create campaign", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, campaign)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		campaigns, total, err := g.uc.List(c.Request.Context(), orgID.(int), limit, offset)
		if err != nil {
			slog.Error("list campaigns", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, entity.PaginatedResponse[entity.Campaign]{
			Items: campaigns,
			Total: total,
		})
	}
}

func (g *Group) handlePreviewAudience() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/preview", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var filter entity.AudienceFilter
		if err := c.ShouldBindJSON(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := g.uc.PreviewAudience(c.Request.Context(), orgID.(int), filter)
		if err != nil {
			slog.Error("preview audience", "error", err, "org_id", orgID)
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		campaign, err := g.uc.GetByID(c.Request.Context(), orgID.(int), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, campaign)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		var req entity.UpdateCampaignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.Update(c.Request.Context(), orgID.(int), id, &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "campaign updated"})
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), orgID.(int), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "campaign deleted"})
	}
}

func (g *Group) handleSend() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/send", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		if err := g.uc.Send(c.Request.Context(), orgID.(int), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "campaign sent"})
	}
}

func (g *Group) handleListScenarios() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/scenarios", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		scenarios, err := g.uc.ListScenarios(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list scenarios", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, scenarios)
	}
}

func (g *Group) handleCreateScenario() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/scenarios", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateScenarioRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		scenario, err := g.uc.CreateScenario(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create scenario", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, scenario)
	}
}

func (g *Group) handleUpdateScenario() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/scenarios/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
			return
		}

		var req entity.UpdateScenarioRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateScenario(c.Request.Context(), orgID.(int), id, &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "scenario updated"})
	}
}

func (g *Group) handleDeleteScenario() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/scenarios/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
			return
		}

		if err := g.uc.DeleteScenario(c.Request.Context(), orgID.(int), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "scenario deleted"})
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, campaignsUC.ErrCampaignNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrNotCampaignOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrCampaignAlreadySent):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrScenarioNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrNotScenarioOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		slog.Error("campaign handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
