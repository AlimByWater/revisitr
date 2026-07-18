package orders

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	ordersUC "revisitr/internal/usecase/orders"
)

type ordersUsecase interface {
	ListOrders(ctx context.Context, orgID, botID int, source, status string) ([]entity.Order, error)
	ListOrgOrders(ctx context.Context, orgID int, source, status string) ([]entity.Order, error)
	UpdateOrderStatus(ctx context.Context, orgID, orderID int, status string) error
}

type Group struct {
	uc        ordersUsecase
	jwtSecret string
}

func New(uc ordersUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/orders"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleListOrgOrders,
		g.handleListOrders,
		g.handleUpdateOrderStatus,
	}
}

func pathID(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name + " id"})
		return 0, false
	}
	return id, true
}

func (g *Group) handleListOrgOrders() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		orders, err := g.uc.ListOrgOrders(c.Request.Context(), orgID.(int),
			c.Query("source"), c.Query("status"))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

func (g *Group) handleListOrders() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/bots/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		botID, ok := pathID(c, "bot")
		if !ok {
			return
		}

		orders, err := g.uc.ListOrders(c.Request.Context(), orgID.(int), botID,
			c.Query("source"), c.Query("status"))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

func (g *Group) handleUpdateOrderStatus() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		orderID, ok := pathID(c, "order")
		if !ok {
			return
		}

		var req entity.UpdateOrderStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateOrderStatus(c.Request.Context(), orgID.(int), orderID, req.Status); err != nil {
			handleError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ordersUC.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ordersUC.ErrNotOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ordersUC.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		slog.Error("orders handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
