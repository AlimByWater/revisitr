package campaigns

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

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
	Schedule(ctx context.Context, orgID, id int, at time.Time) error
	CancelScheduled(ctx context.Context, orgID, id int) error
	RecordClick(ctx context.Context, campaignID, clientID int, buttonIdx *int, url *string) error
	GetAnalytics(ctx context.Context, orgID, id int) (*entity.CampaignAnalyticsDetail, error)
	PreviewAudience(ctx context.Context, orgID int, filter entity.AudienceFilter) (int, error)
	CreateScenario(ctx context.Context, orgID int, req *entity.CreateScenarioRequest) (*entity.AutoScenario, error)
	ListScenarios(ctx context.Context, orgID int) ([]entity.AutoScenario, error)
	UpdateScenario(ctx context.Context, orgID, id int, req *entity.UpdateScenarioRequest) error
	DeleteScenario(ctx context.Context, orgID, id int) error
	ToggleScenario(ctx context.Context, orgID, id int, active bool) error
	GetTemplates(ctx context.Context) ([]entity.AutoScenario, error)
	CloneTemplate(ctx context.Context, orgID int, templateKey string, botID int) (*entity.AutoScenario, error)
	GetActionLog(ctx context.Context, orgID, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error)
	// A/B testing
	CreateABTest(ctx context.Context, orgID, campaignID int, req *entity.CreateABTestRequest) ([]entity.CampaignVariant, error)
	GetVariants(ctx context.Context, orgID, campaignID int) ([]entity.CampaignVariant, error)
	GetABResults(ctx context.Context, orgID, campaignID int) (*entity.ABTestResults, error)
	PickWinner(ctx context.Context, orgID, campaignID, variantID int) error
	// Campaign templates
	CreateCampaignTemplate(ctx context.Context, orgID int, req *entity.CreateCampaignTemplateRequest) (*entity.CampaignTemplate, error)
	ListCampaignTemplates(ctx context.Context, orgID int) ([]entity.CampaignTemplate, error)
	GetCampaignTemplate(ctx context.Context, orgID, id int) (*entity.CampaignTemplate, error)
	UpdateCampaignTemplate(ctx context.Context, orgID, id int, req *entity.UpdateCampaignTemplateRequest) error
	DeleteCampaignTemplate(ctx context.Context, orgID, id int) error
	CreateCampaignFromTemplate(ctx context.Context, orgID, templateID, botID int) (*entity.Campaign, error)
}

type Option func(*Group)

func WithFeatureGate(gate gin.HandlerFunc) Option {
	return func(g *Group) { g.featureGate = gate }
}

type Group struct {
	uc          campaignsUsecase
	jwtSecret   string
	featureGate gin.HandlerFunc
}

func New(uc campaignsUsecase, jwtSecret string, opts ...Option) *Group {
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
		g.handleGetTemplates,
		// Campaign template routes
		g.handleListCampaignTemplates,
		g.handleCreateCampaignTemplate,
		g.handleGetCampaignTemplate,
		g.handleUpdateCampaignTemplate,
		g.handleDeleteCampaignTemplate,
		g.handleCreateFromTemplate,
		// Parameterized scenario routes
		g.handleCloneTemplate,
		g.handleUpdateScenario,
		g.handleDeleteScenario,
		g.handleGetActionLog,
		// Parameterized campaign routes
		g.handleGet,
		g.handleUpdate,
		g.handleDelete,
		g.handleSend,
		g.handleSchedule,
		g.handleCancelSchedule,
		g.handleGetAnalytics,
		g.handleRecordClick,
		// A/B testing routes
		g.handleCreateABTest,
		g.handleGetVariants,
		g.handleGetABResults,
		g.handlePickWinner,
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

func (g *Group) handleSchedule() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/schedule", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		var req struct {
			ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.Schedule(c.Request.Context(), orgID.(int), id, req.ScheduledAt); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "campaign scheduled"})
	}
}

func (g *Group) handleCancelSchedule() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id/schedule", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		if err := g.uc.CancelScheduled(c.Request.Context(), orgID.(int), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "campaign schedule cancelled"})
	}
}

func (g *Group) handleGetAnalytics() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/analytics", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		analytics, err := g.uc.GetAnalytics(c.Request.Context(), orgID.(int), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, analytics)
	}
}

func (g *Group) handleRecordClick() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/click", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		var req struct {
			ClientID  int    `json:"client_id" binding:"required"`
			ButtonIdx *int   `json:"button_idx"`
			URL       *string `json:"url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.RecordClick(c.Request.Context(), id, req.ClientID, req.ButtonIdx, req.URL); err != nil {
			slog.Error("record click", "error", err, "campaign_id", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "click recorded"})
	}
}

func (g *Group) handleGetTemplates() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/scenarios/templates", func(c *gin.Context) {
		templates, err := g.uc.GetTemplates(c.Request.Context())
		if err != nil {
			slog.Error("get templates", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, templates)
	}
}

func (g *Group) handleCloneTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/scenarios/templates/:key/clone", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		key := c.Param("key")

		var req struct {
			BotID int `json:"bot_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		scenario, err := g.uc.CloneTemplate(c.Request.Context(), orgID.(int), key, req.BotID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, scenario)
	}
}

func (g *Group) handleGetActionLog() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/scenarios/:id/log", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		logs, total, err := g.uc.GetActionLog(c.Request.Context(), orgID.(int), id, limit, offset)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, entity.PaginatedResponse[entity.AutoActionLog]{
			Items: logs,
			Total: total,
		})
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
	case errors.Is(err, campaignsUC.ErrCampaignNotDraft):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrCampaignNotScheduled):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrCampaignNotSendable):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrScenarioNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrNotScenarioOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrInvalidVariantPct):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrVariantNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrTemplateNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrTemplateIsSystem):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, campaignsUC.ErrNotTemplateOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		slog.Error("campaign handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// ── A/B Testing Handlers ─────────────────────────────────────────────────────

func (g *Group) handleCreateABTest() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/variants", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		var req entity.CreateABTestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		variants, err := g.uc.CreateABTest(c.Request.Context(), orgID.(int), id, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, variants)
	}
}

func (g *Group) handleGetVariants() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/variants", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		variants, err := g.uc.GetVariants(c.Request.Context(), orgID.(int), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, variants)
	}
}

func (g *Group) handleGetABResults() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/ab-results", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		results, err := g.uc.GetABResults(c.Request.Context(), orgID.(int), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

func (g *Group) handlePickWinner() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/variants/:vid/winner", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
			return
		}

		vid, err := strconv.Atoi(c.Param("vid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant id"})
			return
		}

		if err := g.uc.PickWinner(c.Request.Context(), orgID.(int), id, vid); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "winner selected"})
	}
}

// ── Campaign Template Handlers ───────────────────────────────────────────────

func (g *Group) handleListCampaignTemplates() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/templates", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		templates, err := g.uc.ListCampaignTemplates(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list campaign templates", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, templates)
	}
}

func (g *Group) handleCreateCampaignTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/templates", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateCampaignTemplateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		t, err := g.uc.CreateCampaignTemplate(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create campaign template", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, t)
	}
}

func (g *Group) handleGetCampaignTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/templates/:tid", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		tid, err := strconv.Atoi(c.Param("tid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
			return
		}

		t, err := g.uc.GetCampaignTemplate(c.Request.Context(), orgID.(int), tid)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, t)
	}
}

func (g *Group) handleUpdateCampaignTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/templates/:tid", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		tid, err := strconv.Atoi(c.Param("tid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
			return
		}

		var req entity.UpdateCampaignTemplateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateCampaignTemplate(c.Request.Context(), orgID.(int), tid, &req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "template updated"})
	}
}

func (g *Group) handleDeleteCampaignTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/templates/:tid", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		tid, err := strconv.Atoi(c.Param("tid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
			return
		}

		if err := g.uc.DeleteCampaignTemplate(c.Request.Context(), orgID.(int), tid); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
	}
}

func (g *Group) handleCreateFromTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/templates/:tid/use", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		tid, err := strconv.Atoi(c.Param("tid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
			return
		}

		var req struct {
			BotID int `json:"bot_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		campaign, err := g.uc.CreateCampaignFromTemplate(c.Request.Context(), orgID.(int), tid, req.BotID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, campaign)
	}
}
