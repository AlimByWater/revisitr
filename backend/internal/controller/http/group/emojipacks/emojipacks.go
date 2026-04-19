package emojipacks

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	emojipacksUC "revisitr/internal/usecase/emojipacks"
)

type emojiPacksUsecase interface {
	Create(ctx context.Context, orgID int, req entity.CreateEmojiPackRequest) (*entity.EmojiPack, error)
	Get(ctx context.Context, orgID, packID int) (*entity.EmojiPack, error)
	List(ctx context.Context, orgID int) ([]entity.EmojiPack, error)
	Update(ctx context.Context, orgID, packID int, req entity.UpdateEmojiPackRequest) (*entity.EmojiPack, error)
	Delete(ctx context.Context, orgID, packID int) error
	AddItem(ctx context.Context, orgID, packID int, req entity.CreateEmojiItemRequest) (*entity.EmojiItem, error)
	UpdateItem(ctx context.Context, orgID, itemID int, req entity.UpdateEmojiItemRequest) (*entity.EmojiItem, error)
	DeleteItem(ctx context.Context, orgID, itemID int) error
	ReorderItems(ctx context.Context, orgID, packID int, req entity.ReorderEmojiItemsRequest) error
	SyncToTelegram(ctx context.Context, orgID, packID, botID int) (*entity.EmojiPack, error)
}

type Group struct {
	uc        emojiPacksUsecase
	jwtSecret string
}

func New(uc emojiPacksUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/emoji-packs"
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
		g.handleAddItem,
		g.handleUpdateItem,
		g.handleDeleteItem,
		g.handleReorderItems,
		g.handleSyncToTelegram,
	}
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateEmojiPackRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pack, err := g.uc.Create(c.Request.Context(), orgID.(int), req)
		if err != nil {
			slog.Error("create emoji pack", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, pack)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		packs, err := g.uc.List(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list emoji packs", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, packs)
	}
}

func (g *Group) handleGet() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji pack id"})
			return
		}

		pack, err := g.uc.Get(c.Request.Context(), orgID.(int), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, pack)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji pack id"})
			return
		}

		var req entity.UpdateEmojiPackRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pack, err := g.uc.Update(c.Request.Context(), orgID.(int), id, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, pack)
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji pack id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), orgID.(int), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "emoji pack deleted"})
	}
}

func (g *Group) handleAddItem() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/items", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		packID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji pack id"})
			return
		}

		var req entity.CreateEmojiItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := g.uc.AddItem(c.Request.Context(), orgID.(int), packID, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, item)
	}
}

func (g *Group) handleUpdateItem() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/items/:itemId", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		itemID, err := strconv.Atoi(c.Param("itemId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji item id"})
			return
		}

		var req entity.UpdateEmojiItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := g.uc.UpdateItem(c.Request.Context(), orgID.(int), itemID, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

func (g *Group) handleDeleteItem() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/items/:itemId", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		itemID, err := strconv.Atoi(c.Param("itemId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji item id"})
			return
		}

		if err := g.uc.DeleteItem(c.Request.Context(), orgID.(int), itemID); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "emoji item deleted"})
	}
}

func (g *Group) handleReorderItems() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/:id/items/reorder", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		packID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid emoji pack id"})
			return
		}

		var req entity.ReorderEmojiItemsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.ReorderItems(c.Request.Context(), orgID.(int), packID, req); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "items reordered"})
	}
}

func (g *Group) handleSyncToTelegram() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/:id/sync", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		packID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pack id"})
			return
		}

		var req struct {
			BotID int `json:"bot_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pack, err := g.uc.SyncToTelegram(c.Request.Context(), orgID.(int), packID, req.BotID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, pack)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, emojipacksUC.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, emojipacksUC.ErrNotOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		slog.Error("emoji pack handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
