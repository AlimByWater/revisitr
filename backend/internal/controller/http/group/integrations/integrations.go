package integrations

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
	posService "revisitr/internal/service/pos"
	integrationsUC "revisitr/internal/usecase/integrations"
)

type integrationsUsecase interface {
	Create(ctx context.Context, orgID int, req *entity.CreateIntegrationRequest) (*entity.Integration, error)
	List(ctx context.Context, orgID int) ([]entity.Integration, error)
	GetByID(ctx context.Context, id, orgID int) (*entity.Integration, error)
	Update(ctx context.Context, id, orgID int, req *entity.UpdateIntegrationRequest) (*entity.Integration, error)
	Delete(ctx context.Context, id, orgID int) error
	SyncNow(ctx context.Context, id, orgID int) error
	TestConnection(ctx context.Context, id, orgID int) error
	GetOrders(ctx context.Context, id, orgID, limit, offset int) ([]entity.ExternalOrder, int, error)
	GetCustomers(ctx context.Context, id, orgID int, opts posService.CustomerListOpts) ([]posService.POSCustomer, error)
	GetMenu(ctx context.Context, id, orgID int) (*posService.POSMenu, error)
	GetStats(ctx context.Context, id, orgID int) (*entity.IntegrationStats, error)
	GetAggregates(ctx context.Context, id, orgID int, from, to time.Time) ([]entity.IntegrationAggregate, error)
	GetDashboardData(ctx context.Context, orgID int, from, to time.Time) (*entity.DashboardAggregates, error)
}

type Option func(*Group)

func WithFeatureGate(gate gin.HandlerFunc) Option {
	return func(g *Group) { g.featureGate = gate }
}

type Group struct {
	uc          integrationsUsecase
	jwtSecret   string
	featureGate gin.HandlerFunc
}

func New(uc integrationsUsecase, jwtSecret string, opts ...Option) *Group {
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
	return "/api/v1/integrations"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleCreate,
		g.handleList,
		g.handleGet,
		g.handleUpdate,
		g.handleDelete,
		g.handleSync,
		g.handleTestConnection,
		g.handleGetOrders,
		g.handleGetCustomers,
		g.handleGetMenu,
		g.handleGetStats,
		g.handleGetAggregates,
	}
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateIntegrationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		intg, err := g.uc.Create(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create integration", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, intg)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		intgs, err := g.uc.List(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list integrations", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, intgs)
	}
}

func (g *Group) handleGet() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		intg, err := g.uc.GetByID(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, intg)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		var req entity.UpdateIntegrationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		intg, err := g.uc.Update(c.Request.Context(), id, orgID.(int), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, intg)
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), id, orgID.(int)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "integration deleted"})
	}
}

func (g *Group) handleSync() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/sync", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		if err := g.uc.SyncNow(c.Request.Context(), id, orgID.(int)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "sync completed"})
	}
}

func (g *Group) handleTestConnection() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/test", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		if err := g.uc.TestConnection(c.Request.Context(), id, orgID.(int)); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "connection successful"})
	}
}

func (g *Group) handleGetOrders() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/orders", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if limit > 100 {
			limit = 100
		}

		orders, total, err := g.uc.GetOrders(c.Request.Context(), id, orgID.(int), limit, offset)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": orders,
			"total": total,
		})
	}
}

func (g *Group) handleGetCustomers() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/customers", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		search := c.Query("search")

		customers, err := g.uc.GetCustomers(c.Request.Context(), id, orgID.(int), posService.CustomerListOpts{
			Limit:  limit,
			Offset: offset,
			Search: search,
		})
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, customers)
	}
}

func (g *Group) handleGetMenu() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/menu", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		menu, err := g.uc.GetMenu(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, menu)
	}
}

func (g *Group) handleGetStats() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/stats", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

		stats, err := g.uc.GetStats(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func (g *Group) handleGetAggregates() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/aggregates", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration id"})
			return
		}

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

		aggs, err := g.uc.GetAggregates(c.Request.Context(), id, orgID.(int), from, to)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, aggs)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, integrationsUC.ErrIntegrationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, integrationsUC.ErrNotIntegrationOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		slog.Error("integration handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
