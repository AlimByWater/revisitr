package menus

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	menusUC "revisitr/internal/usecase/menus"
)

type menusUsecase interface {
	Create(ctx context.Context, orgID int, req entity.CreateMenuRequest) (*entity.Menu, error)
	Get(ctx context.Context, orgID, menuID int) (*entity.Menu, error)
	List(ctx context.Context, orgID int) ([]entity.Menu, error)
	Update(ctx context.Context, orgID, menuID int, req entity.UpdateMenuRequest) error
	Delete(ctx context.Context, orgID, menuID int) error
	AddCategory(ctx context.Context, orgID, menuID int, req entity.CreateMenuCategoryRequest) (*entity.MenuCategory, error)
	UpdateCategory(ctx context.Context, orgID, categoryID int, req entity.UpdateMenuCategoryRequest) (*entity.MenuCategory, error)
	AddItem(ctx context.Context, orgID, menuID, categoryID int, req entity.CreateMenuItemRequest) (*entity.MenuItem, error)
	UpdateItem(ctx context.Context, orgID, itemID int, req entity.UpdateMenuItemRequest) (*entity.MenuItem, error)
	GetClientOrderStats(ctx context.Context, clientID int) (*entity.ClientOrderStats, error)
	SetBotPOSLocations(ctx context.Context, botID int, posIDs []int) error
	GetBotPOSLocations(ctx context.Context, botID int) ([]int, error)
}

type Group struct {
	uc        menusUsecase
	jwtSecret string
}

func New(uc menusUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/menus"
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
		g.handleAddCategory,
		g.handleUpdateCategory,
		g.handleAddItem,
		g.handleUpdateItem,
	}
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		menu, err := g.uc.Create(c.Request.Context(), orgID.(int), req)
		if err != nil {
			slog.Error("create menu", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, menu)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		menus, err := g.uc.List(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list menus", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, menus)
	}
}

func (g *Group) handleGet() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu id"})
			return
		}

		menu, err := g.uc.Get(c.Request.Context(), orgID.(int), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, menu)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu id"})
			return
		}

		var req entity.UpdateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.Update(c.Request.Context(), orgID.(int), id, req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "menu updated"})
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), orgID.(int), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "menu deleted"})
	}
}

func (g *Group) handleAddCategory() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/categories", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		menuID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu id"})
			return
		}

		var req entity.CreateMenuCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cat, err := g.uc.AddCategory(c.Request.Context(), orgID.(int), menuID, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, cat)
	}
}

func (g *Group) handleAddItem() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/categories/:catId/items", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		menuID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu id"})
			return
		}
		catID, err := strconv.Atoi(c.Param("catId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
			return
		}

		var req entity.CreateMenuItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := g.uc.AddItem(c.Request.Context(), orgID.(int), menuID, catID, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, item)
	}
}

func (g *Group) handleUpdateCategory() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/categories/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
			return
		}

		var req entity.UpdateMenuCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		category, err := g.uc.UpdateCategory(c.Request.Context(), orgID.(int), id, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, category)
	}
}

func (g *Group) handleUpdateItem() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/items/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
			return
		}

		var req entity.UpdateMenuItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := g.uc.UpdateItem(c.Request.Context(), orgID.(int), id, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, menusUC.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, menusUC.ErrNotOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		slog.Error("menu handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
