package account

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	accountUC "revisitr/internal/usecase/account"
)

type accountUsecase interface {
	GetOrganization(ctx context.Context, orgID int) (*entity.Organization, error)
	UpdateOrganization(ctx context.Context, orgID int, req entity.UpdateOrganizationRequest) (*entity.Organization, error)
}

type Group struct {
	uc        accountUsecase
	jwtSecret string
}

func New(uc accountUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/account"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetOrg,
		g.handleUpdateOrg,
	}
}

func (g *Group) handleGetOrg() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/org", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		org, err := g.uc.GetOrganization(c.Request.Context(), orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, org)
	}
}

func (g *Group) handleUpdateOrg() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/org", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.UpdateOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		org, err := g.uc.UpdateOrganization(c.Request.Context(), orgID.(int), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, org)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, accountUC.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, accountUC.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		slog.Error("account handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
