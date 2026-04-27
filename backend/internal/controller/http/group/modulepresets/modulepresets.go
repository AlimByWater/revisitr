package modulepresets

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	modulepresetsUC "revisitr/internal/usecase/modulepresets"
)

type presetsUsecase interface {
	ListPresets(ctx context.Context, moduleKey string) ([]entity.ModulePreset, error)
	GetBotModuleSettings(ctx context.Context, botID, orgID int, moduleKey string) (*entity.BotModuleSettings, error)
	SelectPreset(ctx context.Context, botID, orgID int, moduleKey, presetKey string) error
	UpdateCustomizations(ctx context.Context, botID, orgID int, moduleKey string, customizations json.RawMessage) error
	ResetPreset(ctx context.Context, botID, orgID int, moduleKey string) error
}

type Group struct {
	uc        presetsUsecase
	jwtSecret string
}

func New(uc presetsUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/module-settings"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

// PublicHandlers returns unauthenticated routes (preset catalog).
func (g *Group) PublicHandlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleListPresets,
	}
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetBotModuleSettings,
		g.handleSelectPreset,
		g.handleUpdateCustomizations,
		g.handleResetPreset,
	}
}

func (g *Group) handleListPresets() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/catalog/:moduleKey", func(c *gin.Context) {
		moduleKey := c.Param("moduleKey")
		presets, err := g.uc.ListPresets(c.Request.Context(), moduleKey)
		if err != nil {
			slog.Error("list presets", "error", err, "module_key", moduleKey)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		slog.Info("list presets result", "module_key", moduleKey, "count", len(presets))
		if presets == nil {
			presets = []entity.ModulePreset{}
		}
		data, jsonErr := json.Marshal(presets)
		if jsonErr != nil {
			slog.Error("marshal presets", "error", jsonErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "marshal error"})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", data)
	}
}

func (g *Group) handleGetBotModuleSettings() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/:moduleKey", func(c *gin.Context) {
		orgID, botID, ok := g.extractIDs(c)
		if !ok {
			return
		}
		moduleKey := c.Param("moduleKey")

		settings, err := g.uc.GetBotModuleSettings(c.Request.Context(), botID, orgID, moduleKey)
		if err != nil {
			handleError(c, err)
			return
		}
		if settings == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no module settings found"})
			return
		}
		c.JSON(http.StatusOK, settings)
	}
}

type selectPresetRequest struct {
	PresetKey string `json:"preset_key" binding:"required"`
}

func (g *Group) handleSelectPreset() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/:id/:moduleKey/preset", func(c *gin.Context) {
		orgID, botID, ok := g.extractIDs(c)
		if !ok {
			return
		}
		moduleKey := c.Param("moduleKey")

		var req selectPresetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.SelectPreset(c.Request.Context(), botID, orgID, moduleKey, req.PresetKey); err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

type updateCustomizationsRequest struct {
	Customizations json.RawMessage `json:"customizations" binding:"required"`
}

func (g *Group) handleUpdateCustomizations() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id/:moduleKey/customizations", func(c *gin.Context) {
		orgID, botID, ok := g.extractIDs(c)
		if !ok {
			return
		}
		moduleKey := c.Param("moduleKey")

		var req updateCustomizationsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateCustomizations(c.Request.Context(), botID, orgID, moduleKey, req.Customizations); err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func (g *Group) handleResetPreset() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/:moduleKey/reset", func(c *gin.Context) {
		orgID, botID, ok := g.extractIDs(c)
		if !ok {
			return
		}
		moduleKey := c.Param("moduleKey")

		if err := g.uc.ResetPreset(c.Request.Context(), botID, orgID, moduleKey); err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func (g *Group) extractIDs(c *gin.Context) (orgID, botID int, ok bool) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, 0, false
	}
	orgID = orgIDVal.(int)

	botID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
		return 0, 0, false
	}
	return orgID, botID, true
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, modulepresetsUC.ErrBotNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, modulepresetsUC.ErrNotBotOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, modulepresetsUC.ErrPresetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, modulepresetsUC.ErrInvalidPreset):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		slog.Error("module presets handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
