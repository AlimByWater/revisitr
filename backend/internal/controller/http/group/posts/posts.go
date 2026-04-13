package posts

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
)

type postCodesRepo interface {
	GetByCode(ctx context.Context, orgID int, code string) (*entity.PostCode, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.PostCode, error)
	Delete(ctx context.Context, orgID int, code string) error
}

type Group struct {
	repo      postCodesRepo
	jwtSecret string
}

func New(repo postCodesRepo, jwtSecret string) *Group {
	return &Group{repo: repo, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/posts"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleList,
		g.handleGetByCode,
		g.handleDelete,
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		codes, err := g.repo.GetByOrgID(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list post codes", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, codes)
	}
}

func (g *Group) handleGetByCode() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:code", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		code := c.Param("code")

		pc, err := g.repo.GetByCode(c.Request.Context(), orgID.(int), code)
		if err != nil {
			slog.Error("get post code", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if pc == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		c.JSON(http.StatusOK, pc)
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:code", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		code := c.Param("code")

		err := g.repo.Delete(c.Request.Context(), orgID.(int), code)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
				return
			}
			slog.Error("delete post code", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "post deleted"})
	}
}
