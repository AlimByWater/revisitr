package posplugin

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	posUC "revisitr/internal/usecase/posplugin"
)

type pluginAdminUsecase interface {
	CreateKey(ctx context.Context, orgID, integrationID int, label string) (string, *entity.PluginKey, error)
	ListKeys(ctx context.Context, orgID, integrationID int) ([]entity.PluginKey, error)
	RevokeKey(ctx context.Context, orgID, keyID int) error
}

type AdminGroup struct {
	uc        pluginAdminUsecase
	jwtSecret string
}

func NewAdmin(uc pluginAdminUsecase, jwtSecret string) *AdminGroup {
	return &AdminGroup{uc: uc, jwtSecret: jwtSecret}
}

func (g *AdminGroup) Path() string {
	return "/api/v1/pos-plugin/admin"
}

func (g *AdminGroup) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *AdminGroup) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleCreateKey,
		g.handleListKeys,
		g.handleRevokeKey,
	}
}

type createKeyRequest struct {
	IntegrationID int    `json:"integration_id" binding:"required"`
	Label         string `json:"label"`
}

func (g *AdminGroup) handleCreateKey() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/keys", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		var req createKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rawKey, k, err := g.uc.CreateKey(c.Request.Context(), orgID, req.IntegrationID, req.Label)
		if err != nil {
			adminError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"key": rawKey, "id": k.ID, "label": k.Label})
	}
}

func (g *AdminGroup) handleListKeys() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/keys", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		integrationID, err := strconv.Atoi(c.Query("integration_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration_id"})
			return
		}

		keys, err := g.uc.ListKeys(c.Request.Context(), orgID, integrationID)
		if err != nil {
			adminError(c, err)
			return
		}

		c.JSON(http.StatusOK, keys)
	}
}

func (g *AdminGroup) handleRevokeKey() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/keys/:id", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		keyID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
			return
		}

		if err := g.uc.RevokeKey(c.Request.Context(), orgID, keyID); err != nil {
			adminError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "key revoked"})
	}
}

func adminError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, posUC.ErrUnauthorizedKey):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
