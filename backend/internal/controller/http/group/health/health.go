package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Group struct{}

func New() *Group {
	return &Group{}
}

func (g *Group) Path() string {
	return ""
}

func (g *Group) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.healthz,
	}
}

func (g *Group) healthz() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
