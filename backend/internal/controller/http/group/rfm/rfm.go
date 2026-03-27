package rfm

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
)

type rfmUsecase interface {
	GetDashboard(ctx context.Context, orgID int) (*entity.RFMDashboard, error)
	Recalculate(ctx context.Context, orgID int) error
	GetConfig(ctx context.Context, orgID int) (*entity.RFMConfig, error)
	UpdateConfig(ctx context.Context, orgID int, req entity.UpdateRFMConfigRequest) (*entity.RFMConfig, error)
	ListTemplates() []entity.RFMTemplate
	GetActiveTemplate(ctx context.Context, orgID int) (*entity.RFMConfig, *entity.RFMTemplate, error)
	SetTemplate(ctx context.Context, orgID int, req entity.SetTemplateRequest) (*entity.RFMTemplate, error)
	GetOnboardingQuestions() []entity.OnboardingQuestion
	RecommendTemplate(answers []int) (*entity.TemplateRecommendation, error)
	GetSegmentClients(ctx context.Context, orgID int, segment string, page, perPage int, sortCol, order string) (*entity.SegmentClientsResponse, error)
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
		g.handleListTemplates,
		g.handleGetActiveTemplate,
		g.handleSetTemplate,
		g.handleOnboardingQuestions,
		g.handleOnboardingRecommend,
		g.handleSegmentClients,
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

func (g *Group) handleListTemplates() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/templates", func(c *gin.Context) {
		templates := g.uc.ListTemplates()
		c.JSON(http.StatusOK, gin.H{"templates": templates})
	}
}

func (g *Group) handleGetActiveTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/template", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		cfg, tpl, err := g.uc.GetActiveTemplate(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("rfm get active template", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"active_template_type": cfg.ActiveTemplateType,
			"active_template_key":  cfg.ActiveTemplateKey,
			"template":             tpl,
		})
	}
}

func (g *Group) handleSetTemplate() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/template", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.SetTemplateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tpl, err := g.uc.SetTemplate(c.Request.Context(), orgID.(int), req)
		if err != nil {
			slog.Error("rfm set template", "error", err, "org_id", orgID)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "template updated",
			"template":      tpl,
			"recalculation": "started",
		})
	}
}

func (g *Group) handleOnboardingQuestions() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/onboarding/questions", func(c *gin.Context) {
		questions := g.uc.GetOnboardingQuestions()
		c.JSON(http.StatusOK, gin.H{"questions": questions})
	}
}

func (g *Group) handleOnboardingRecommend() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/onboarding/recommend", func(c *gin.Context) {
		var req struct {
			Answers []int `json:"answers" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := g.uc.RecommendTemplate(req.Answers)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (g *Group) handleSegmentClients() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/segments/:segment/clients", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		segment := c.Param("segment")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
		sortCol := c.DefaultQuery("sort", "monetary_sum")
		order := c.DefaultQuery("order", "desc")

		result, err := g.uc.GetSegmentClients(c.Request.Context(), orgID.(int), segment, page, perPage, sortCol, order)
		if err != nil {
			slog.Error("rfm segment clients", "error", err, "org_id", orgID, "segment", segment)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
