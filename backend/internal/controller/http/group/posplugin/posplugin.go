package posplugin

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	posUC "revisitr/internal/usecase/posplugin"
)

type pluginUsecase interface {
	AuthenticateKey(ctx context.Context, rawKey string) (*entity.PluginKey, error)
	Identify(ctx context.Context, key *entity.PluginKey, code string, orderTotal float64) (*posUC.IdentifyResult, error)
	Redeem(ctx context.Context, key *entity.PluginKey, session, orderID string, amount float64) (*posUC.OpResult, error)
	Accrue(ctx context.Context, key *entity.PluginKey, session, orderID string, amount float64) (*posUC.OpResult, error)
	Config(ctx context.Context, key *entity.PluginKey) (*posUC.ConfigResult, error)
	SubmitOrder(ctx context.Context, key *entity.PluginKey, req posUC.SubmitOrderRequest) error
}

type Group struct {
	uc pluginUsecase
}

func New(uc pluginUsecase) *Group {
	return &Group{uc: uc}
}

func (g *Group) Path() string {
	return "/api/v1/pos-plugin"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.APIKeyAuth(g.uc)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleIdentify,
		g.handleRedeem,
		g.handleAccrue,
		g.handleSubmitOrder,
		g.handleConfig,
	}
}

type identifyRequest struct {
	Code       string  `json:"code" binding:"required"`
	OrderTotal float64 `json:"order_total"`
}

type opRequest struct {
	Session string  `json:"session" binding:"required"`
	OrderID string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount"`
}

func (g *Group) handleIdentify() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/identify", func(c *gin.Context) {
		key := c.MustGet("plugin_key").(*entity.PluginKey)

		var req identifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := g.uc.Identify(c.Request.Context(), key, req.Code, req.OrderTotal)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func (g *Group) handleRedeem() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/redeem", func(c *gin.Context) {
		key := c.MustGet("plugin_key").(*entity.PluginKey)

		var req opRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := g.uc.Redeem(c.Request.Context(), key, req.Session, req.OrderID, req.Amount)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func (g *Group) handleAccrue() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/accrue", func(c *gin.Context) {
		key := c.MustGet("plugin_key").(*entity.PluginKey)

		var req opRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := g.uc.Accrue(c.Request.Context(), key, req.Session, req.OrderID, req.Amount)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func (g *Group) handleSubmitOrder() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/order", func(c *gin.Context) {
		key := c.MustGet("plugin_key").(*entity.PluginKey)
		var req posUC.SubmitOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := g.uc.SubmitOrder(c.Request.Context(), key, req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save order"})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (g *Group) handleConfig() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/config", func(c *gin.Context) {
		key := c.MustGet("plugin_key").(*entity.PluginKey)

		res, err := g.uc.Config(c.Request.Context(), key)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, posUC.ErrGuestNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, posUC.ErrSessionInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, posUC.ErrInvalidAmount):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, posUC.ErrInsufficient):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, posUC.ErrRateLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	default:
		slog.Error("pos-plugin handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
