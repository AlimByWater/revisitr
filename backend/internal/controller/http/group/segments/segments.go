package segments

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	segmentsUC "revisitr/internal/usecase/segments"
)

type segmentsUsecase interface {
	Create(ctx context.Context, orgID int, req *entity.CreateSegmentRequest) (*entity.Segment, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error)
	GetByID(ctx context.Context, id, orgID int) (*entity.Segment, error)
	Update(ctx context.Context, id, orgID int, req *entity.UpdateSegmentRequest) (*entity.Segment, error)
	Delete(ctx context.Context, id, orgID int) error
	GetClients(ctx context.Context, segmentID, orgID, limit, offset int) ([]entity.BotClient, int, error)
	RecalculateCustom(ctx context.Context, segmentID, orgID int) error
	PreviewCount(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error)
	// Advanced segmentation
	AddRule(ctx context.Context, segmentID, orgID int, req entity.CreateSegmentRuleRequest) (*entity.SegmentRule, error)
	GetRules(ctx context.Context, segmentID, orgID int) ([]entity.SegmentRule, error)
	DeleteRule(ctx context.Context, segmentID, orgID, ruleID int) error
	GetPredictions(ctx context.Context, orgID, limit, offset int) ([]entity.ClientPrediction, error)
	GetClientPrediction(ctx context.Context, clientID int) (*entity.ClientPrediction, error)
	GetHighChurnClients(ctx context.Context, orgID int, threshold float32) ([]entity.ClientPrediction, error)
	GetPredictionSummary(ctx context.Context, orgID int) (*entity.PredictionSummary, error)
}

type Group struct {
	uc        segmentsUsecase
	jwtSecret string
}

func New(uc segmentsUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/segments"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleCreate,
		g.handleList,
		g.handlePreview,
		g.handleGet,
		g.handleUpdate,
		g.handleDelete,
		g.handleGetClients,
		g.handleRecalculate,
		// Advanced segmentation
		g.handleAddRule,
		g.handleGetRules,
		g.handleDeleteRule,
		g.handleGetPredictions,
		g.handleGetPredictionSummary,
		g.handleGetHighChurn,
	}
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateSegmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		seg, err := g.uc.Create(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create segment", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, seg)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		segs, err := g.uc.GetByOrgID(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list segments", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, segs)
	}
}

func (g *Group) handlePreview() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/preview", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var f entity.SegmentFilter
		if err := c.ShouldBindJSON(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := g.uc.PreviewCount(c.Request.Context(), orgID.(int), f)
		if err != nil {
			slog.Error("preview segment", "error", err, "org_id", orgID)
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}

		seg, err := g.uc.GetByID(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, seg)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}

		var req entity.UpdateSegmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		seg, err := g.uc.Update(c.Request.Context(), id, orgID.(int), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, seg)
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), id, orgID.(int)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "segment deleted"})
	}
}

func (g *Group) handleGetClients() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/clients", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		clients, total, err := g.uc.GetClients(c.Request.Context(), id, orgID.(int), limit, offset)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, entity.PaginatedResponse[entity.BotClient]{
			Items: clients,
			Total: total,
		})
	}
}

func (g *Group) handleRecalculate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/recalculate", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}

		if err := g.uc.RecalculateCustom(c.Request.Context(), id, orgID.(int)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "segment recalculated"})
	}
}

// ── Advanced Segmentation Handlers ────────────────────────────────────────────

func (g *Group) handleAddRule() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/rules", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}
		var req entity.CreateSegmentRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rule, err := g.uc.AddRule(c.Request.Context(), id, orgID.(int), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, rule)
	}
}

func (g *Group) handleGetRules() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/rules", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}
		rules, err := g.uc.GetRules(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, rules)
	}
}

func (g *Group) handleDeleteRule() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id/rules/:ruleId", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment id"})
			return
		}
		ruleID, err := strconv.Atoi(c.Param("ruleId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
			return
		}
		if err := g.uc.DeleteRule(c.Request.Context(), id, orgID.(int), ruleID); err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func (g *Group) handleGetPredictions() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/predictions", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		preds, err := g.uc.GetPredictions(c.Request.Context(), orgID.(int), limit, offset)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, preds)
	}
}

func (g *Group) handleGetPredictionSummary() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/predictions/summary", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		summary, err := g.uc.GetPredictionSummary(c.Request.Context(), orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, summary)
	}
}

func (g *Group) handleGetHighChurn() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/predictions/high-churn", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		threshold := float32(0.7)
		if t := c.Query("threshold"); t != "" {
			if v, err := strconv.ParseFloat(t, 32); err == nil {
				threshold = float32(v)
			}
		}
		preds, err := g.uc.GetHighChurnClients(c.Request.Context(), orgID.(int), threshold)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, preds)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, segmentsUC.ErrSegmentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, segmentsUC.ErrNotSegmentOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, segmentsUC.ErrNotCustomSegment):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, segmentsUC.ErrPredictionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		slog.Error("segment handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
