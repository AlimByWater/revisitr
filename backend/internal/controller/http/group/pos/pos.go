package pos

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	posUC "revisitr/internal/usecase/pos"
)

type posUsecase interface {
	Create(ctx context.Context, orgID int, req *entity.CreatePOSRequest) (*entity.POSLocation, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.POSLocation, error)
	GetByID(ctx context.Context, id, orgID int) (*entity.POSLocation, error)
	Update(ctx context.Context, id, orgID int, req *entity.UpdatePOSRequest) (*entity.POSLocation, error)
	Delete(ctx context.Context, id, orgID int) error
}

type Group struct {
	uc        posUsecase
	jwtSecret string
}

func New(uc posUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/pos"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleCreate,
		g.handleList,
		g.handleGet,
		g.handleUpdate,
		g.handleDelete,
	}
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		var req entity.CreatePOSRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pos, err := g.uc.Create(c.Request.Context(), orgID, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, pos)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		locations, err := g.uc.GetByOrgID(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, locations)
	}
}

func (g *Group) handleGet() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		pos, err := g.uc.GetByID(c.Request.Context(), id, orgID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, pos)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req entity.UpdatePOSRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pos, err := g.uc.Update(c.Request.Context(), id, orgID, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, pos)
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), id, orgID); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, posUC.ErrPOSNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, posUC.ErrNotPOSOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
