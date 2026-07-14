package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/entity"
)

type KeyAuthenticator interface {
	AuthenticateKey(ctx context.Context, rawKey string) (*entity.PluginKey, error)
}

// APIKeyAuth authenticates a POS plugin request via the X-API-Key header. On
// success it sets org_id, integration_id, and plugin_key on the context.
func APIKeyAuth(a KeyAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-Key")
		if rawKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
			return
		}

		key, err := a.AuthenticateKey(c.Request.Context(), rawKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		c.Set("org_id", key.OrgID)
		c.Set("integration_id", key.IntegrationID)
		c.Set("plugin_key", key)
		c.Next()
	}
}
