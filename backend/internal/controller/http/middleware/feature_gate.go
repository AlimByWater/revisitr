package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FeatureChecker checks whether an organization has access to a feature.
type FeatureChecker interface {
	HasFeature(ctx context.Context, orgID int, feature string) (bool, error)
}

// FeatureGate returns middleware that blocks access unless the org's subscription
// includes the named feature. If checker is nil, access is always granted (graceful
// degradation when billing is not yet configured).
func FeatureGate(checker FeatureChecker, feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if checker == nil {
			c.Next()
			return
		}

		orgIDVal, exists := c.Get("org_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing org_id"})
			return
		}
		orgID := orgIDVal.(int)

		allowed, err := checker.HasFeature(c.Request.Context(), orgID, feature)
		if err != nil {
			// On error, allow access (fail-open) to avoid blocking the entire platform
			// due to a billing database issue. Log the error.
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "feature not available on your plan",
				"feature": feature,
			})
			return
		}

		c.Next()
	}
}
