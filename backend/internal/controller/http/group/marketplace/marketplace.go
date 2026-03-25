package marketplace

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	marketplaceUC "revisitr/internal/usecase/marketplace"
)

type marketplaceUsecase interface {
	CreateProduct(ctx context.Context, orgID int, req entity.CreateProductRequest) (*entity.MarketplaceProduct, error)
	GetProducts(ctx context.Context, orgID int) ([]entity.MarketplaceProduct, error)
	GetProduct(ctx context.Context, orgID int, productID int) (*entity.MarketplaceProduct, error)
	UpdateProduct(ctx context.Context, orgID int, productID int, req entity.UpdateProductRequest) (*entity.MarketplaceProduct, error)
	DeleteProduct(ctx context.Context, orgID int, productID int) error
	PlaceOrder(ctx context.Context, orgID int, req entity.PlaceOrderRequest) (*entity.MarketplaceOrder, error)
	GetOrders(ctx context.Context, orgID int) ([]entity.MarketplaceOrder, error)
	GetOrder(ctx context.Context, orgID int, orderID int) (*entity.MarketplaceOrder, error)
	UpdateOrderStatus(ctx context.Context, orgID int, orderID int, status string) error
	GetStats(ctx context.Context, orgID int) (*entity.MarketplaceStats, error)
}

type Group struct {
	uc        marketplaceUsecase
	jwtSecret string
}

func New(uc marketplaceUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/marketplace"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetProducts,
		g.handleGetProduct,
		g.handleCreateProduct,
		g.handleUpdateProduct,
		g.handleDeleteProduct,
		g.handlePlaceOrder,
		g.handleGetOrders,
		g.handleGetOrder,
		g.handleUpdateOrderStatus,
		g.handleGetStats,
	}
}

func (g *Group) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, marketplaceUC.ErrProductNotFound), errors.Is(err, marketplaceUC.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, marketplaceUC.ErrProductInactive):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, marketplaceUC.ErrInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, marketplaceUC.ErrInsufficientPoints):
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient loyalty points"})
	case errors.Is(err, marketplaceUC.ErrEmptyOrder):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, marketplaceUC.ErrWrongOrg):
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	default:
		slog.Error("marketplace", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// ── Products ─────────────────────────────────────────────────────────────────

func (g *Group) handleGetProducts() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/products", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		products, err := g.uc.GetProducts(c.Request.Context(), orgID.(int))
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, products)
	}
}

func (g *Group) handleGetProduct() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/products/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}
		p, err := g.uc.GetProduct(c.Request.Context(), orgID.(int), id)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

func (g *Group) handleCreateProduct() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/products", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		var req entity.CreateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p, err := g.uc.CreateProduct(c.Request.Context(), orgID.(int), req)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, p)
	}
}

func (g *Group) handleUpdateProduct() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/products/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}
		var req entity.UpdateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p, err := g.uc.UpdateProduct(c.Request.Context(), orgID.(int), id, req)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

func (g *Group) handleDeleteProduct() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/products/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}
		if err := g.uc.DeleteProduct(c.Request.Context(), orgID.(int), id); err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// ── Orders ───────────────────────────────────────────────────────────────────

func (g *Group) handlePlaceOrder() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/orders", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		var req entity.PlaceOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		order, err := g.uc.PlaceOrder(c.Request.Context(), orgID.(int), req)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, order)
	}
}

func (g *Group) handleGetOrders() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/orders", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		orders, err := g.uc.GetOrders(c.Request.Context(), orgID.(int))
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

func (g *Group) handleGetOrder() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/orders/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		order, err := g.uc.GetOrder(c.Request.Context(), orgID.(int), id)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, order)
	}
}

func (g *Group) handleUpdateOrderStatus() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/orders/:id/status", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		var body struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := g.uc.UpdateOrderStatus(c.Request.Context(), orgID.(int), id, body.Status); err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func (g *Group) handleGetStats() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/stats", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		stats, err := g.uc.GetStats(c.Request.Context(), orgID.(int))
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}
