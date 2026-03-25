package wallet

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	walletUC "revisitr/internal/usecase/wallet"
)

type walletUsecase interface {
	GetConfigs(ctx context.Context, orgID int) ([]entity.WalletConfig, error)
	GetConfig(ctx context.Context, orgID int, platform string) (*entity.WalletConfig, error)
	SaveConfig(ctx context.Context, orgID int, req entity.SaveWalletConfigRequest) (*entity.WalletConfig, error)
	DeleteConfig(ctx context.Context, orgID int, platform string) error
	IssuePass(ctx context.Context, orgID int, req entity.IssueWalletPassRequest) (*entity.WalletPass, error)
	GetPass(ctx context.Context, serial string) (*entity.WalletPass, error)
	GetPasses(ctx context.Context, orgID int) ([]entity.WalletPass, error)
	RegisterPushToken(ctx context.Context, serial string, authToken string, pushToken string) error
	RevokePass(ctx context.Context, orgID int, passID int) error
	GetStats(ctx context.Context, orgID int) (*entity.WalletStats, error)
}

type Group struct {
	uc        walletUsecase
	jwtSecret string
}

func New(uc walletUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/wallet"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetConfigs,
		g.handleGetConfig,
		g.handleSaveConfig,
		g.handleDeleteConfig,
		g.handleIssuePasses,
		g.handleGetPasses,
		g.handleRevokePass,
		g.handleGetStats,
		// Public endpoints (no auth) handled separately
		g.handleRegisterPushToken,
	}
}

func (g *Group) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, walletUC.ErrConfigNotFound), errors.Is(err, walletUC.ErrPassNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, walletUC.ErrPlatformDisabled):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, walletUC.ErrPassAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, walletUC.ErrInvalidPlatform):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		slog.Error("wallet", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// ── Config ───────────────────────────────────────────────────────────────────

func (g *Group) handleGetConfigs() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/configs", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		configs, err := g.uc.GetConfigs(c.Request.Context(), orgID.(int))
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, configs)
	}
}

func (g *Group) handleGetConfig() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/configs/:platform", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		platform := c.Param("platform")
		cfg, err := g.uc.GetConfig(c.Request.Context(), orgID.(int), platform)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, cfg)
	}
}

func (g *Group) handleSaveConfig() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/configs/:platform", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		platform := c.Param("platform")

		var req entity.SaveWalletConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.Platform = platform

		cfg, err := g.uc.SaveConfig(c.Request.Context(), orgID.(int), req)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, cfg)
	}
}

func (g *Group) handleDeleteConfig() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/configs/:platform", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		platform := c.Param("platform")
		if err := g.uc.DeleteConfig(c.Request.Context(), orgID.(int), platform); err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// ── Passes ───────────────────────────────────────────────────────────────────

func (g *Group) handleIssuePasses() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/passes", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.IssueWalletPassRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pass, err := g.uc.IssuePass(c.Request.Context(), orgID.(int), req)
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, pass)
	}
}

func (g *Group) handleGetPasses() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/passes", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		passes, err := g.uc.GetPasses(c.Request.Context(), orgID.(int))
		if err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, passes)
	}
}

func (g *Group) handleRevokePass() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/passes/:id/revoke", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pass id"})
			return
		}
		if err := g.uc.RevokePass(c.Request.Context(), orgID.(int), id); err != nil {
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

// ── Public (Apple/Google callback) ───────────────────────────────────────────

func (g *Group) handleRegisterPushToken() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/passes/:id/push", func(c *gin.Context) {
		serial := c.Param("id")
		authToken := c.GetHeader("Authorization")

		var req entity.RegisterPushTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.RegisterPushToken(c.Request.Context(), serial, authToken, req.PushToken); err != nil {
			g.handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
