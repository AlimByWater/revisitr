package wallet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
	GetClientsQRCode(ctx context.Context, clientID int) (string, error)
	GetOrgName(ctx context.Context, orgID int) (string, error)
	GenerateGoogleSaveURL(ctx context.Context, orgID int, pass *entity.WalletPass) (string, error)
}

type Group struct {
	uc        walletUsecase
	jwtSecret string
	passGen   *walletUC.PassGenerator
}

func New(uc walletUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret, passGen: walletUC.NewPassGenerator()}
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
		g.handleDownloadPass,
		g.handleGoogleSaveURL,
	}
}

func (g *Group) PublicHandlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleAppleGetPass,
		g.handleAppleLogPass,
		g.handleAppleLogCrash,
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

// ── Google Wallet save URL ──────────────────────────────────────────────────

func (g *Group) handleGoogleSaveURL() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/passes/:serial/google-save", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		serial := c.Param("serial")

		pass, err := g.uc.GetPass(c.Request.Context(), serial)
		if err != nil {
			g.handleError(c, err)
			return
		}

		saveURL, err := g.uc.GenerateGoogleSaveURL(c.Request.Context(), orgID.(int), pass)
		if err != nil {
			slog.Error("wallet google save", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate save url"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": saveURL})
	}
}

// ── Admin pass download ──────────────────────────────────────────────────────

func (g *Group) handleDownloadPass() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/passes/:serial/download", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		serial := c.Param("serial")

		pass, err := g.uc.GetPass(c.Request.Context(), serial)
		if err != nil {
			g.handleError(c, err)
			return
		}

		pkpassData, err := g.generatePass(c, pass, orgID.(int))
		if err != nil {
			slog.Error("wallet generate pass", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pass"})
			return
		}

		c.Data(http.StatusOK, "application/vnd.apple.pkpass", pkpassData)
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (g *Group) generatePass(c *gin.Context, pass *entity.WalletPass, orgID int) ([]byte, error) {
	cfg, err := g.uc.GetConfig(c.Request.Context(), orgID, "apple")
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("apple wallet not configured")
	}

	clientQR, err := g.uc.GetClientsQRCode(c.Request.Context(), pass.ClientID)
	if err != nil {
		slog.Warn("wallet: could not get client qr code", "client_id", pass.ClientID, "error", err)
	}

	orgName, err := g.uc.GetOrgName(c.Request.Context(), orgID)
	if err != nil {
		slog.Warn("wallet: could not get org name", "org_id", orgID, "error", err)
	}

	webServiceURL := cfg.Design.WebServiceURL
	if webServiceURL == "" {
		scheme := "https"
		if c.Request.TLS == nil {
			scheme = "http"
		}
		webServiceURL = fmt.Sprintf("%s://%s/api/v1/wallet", scheme, c.Request.Host)
	}

	return g.passGen.GeneratePass(pass, cfg, clientQR, orgName, webServiceURL)
}

func (g *Group) verifyAppleAuth(c *gin.Context, pass *entity.WalletPass) error {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "ApplePass ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		return fmt.Errorf("invalid auth header")
	}
	token := strings.TrimPrefix(auth, "ApplePass ")
	if token != pass.AuthToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		return fmt.Errorf("invalid auth token")
	}
	return nil
}

// ── Apple Wallet Web Service (public) ────────────────────────────────────────
//
// Apple calls these endpoints after the pass is added to Wallet.
// Auth via Authorization: ApplePass <authenticationToken>

func (g *Group) handleAppleGetPass() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/v1/passes/:passTypeId/:serial", func(c *gin.Context) {
		serial := c.Param("serial")

		pass, err := g.uc.GetPass(c.Request.Context(), serial)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "pass not found"})
			return
		}

		if err := g.verifyAppleAuth(c, pass); err != nil {
			return
		}

		pkpassData, err := g.generatePass(c, pass, pass.OrgID)
		if err != nil {
			slog.Error("wallet apple get pass", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pass"})
			return
		}

		c.Data(http.StatusOK, "application/vnd.apple.pkpass", pkpassData)
	}
}

func (g *Group) handleAppleLogPass() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/v1/passes/:passTypeId/:serial/log", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func (g *Group) handleAppleLogCrash() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/v1/log", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// ── Public push token registration ──────────────────────────────────────────

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
