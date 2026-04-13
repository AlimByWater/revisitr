package bots

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"revisitr/internal/entity"
)

type managedBotDeps interface {
	StoreAuthToken(ctx context.Context, token string, data entity.MasterBotAuthToken) error
	CreatePendingBot(ctx context.Context, orgID int, req *entity.CreateManagedBotRequest) (*entity.Bot, error)
	GetBotStatus(ctx context.Context, botID, orgID int) (string, error)
}

// WithManagedBots adds managed bot endpoints to the group.
func WithManagedBots(deps managedBotDeps, masterBotUsername string) Option {
	return func(g *Group) {
		g.managed = deps
		g.masterBotUsername = masterBotUsername
	}
}

func (g *Group) handleActivationLink() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/activation-link", func(c *gin.Context) {
		if g.managed == nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "managed bots not configured"})
			return
		}

		orgID, _ := c.Get("org_id")
		userID, _ := c.Get("user_id")

		// Generate one-time token
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		token := hex.EncodeToString(b)

		authData := entity.MasterBotAuthToken{
			OrgID:  orgID.(int),
			UserID: userID.(int),
		}

		if err := g.managed.StoreAuthToken(c.Request.Context(), token, authData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		username := g.masterBotUsername
		if username == "" {
			username = "revisitrbot"
		}

		deepLink := fmt.Sprintf("https://t.me/%s?start=%s", username, token)

		c.JSON(http.StatusCreated, entity.ActivationLinkResponse{
			DeepLink:  deepLink,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		})
	}
}

func (g *Group) handleCreateManaged() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/create-managed", func(c *gin.Context) {
		if g.managed == nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "managed bots not configured"})
			return
		}

		orgID, _ := c.Get("org_id")

		var req entity.CreateManagedBotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate username
		req.Username = strings.TrimPrefix(req.Username, "@")
		if !strings.HasSuffix(strings.ToLower(req.Username), "bot") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username must end with 'bot'"})
			return
		}

		bot, err := g.managed.CreatePendingBot(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		username := g.masterBotUsername
		if username == "" {
			username = "revisitrbot"
		}

		deepLink := fmt.Sprintf("https://t.me/newbot/%s/%s?name=%s", username, req.Username, req.Name)

		c.JSON(http.StatusCreated, entity.CreateManagedBotResponse{
			BotID:    bot.ID,
			DeepLink: deepLink,
			Status:   bot.Status,
		})
	}
}

func (g *Group) handleGetBotStatus() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/status", func(c *gin.Context) {
		if g.managed == nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "managed bots not configured"})
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
			return
		}

		orgID, _ := c.Get("org_id")

		status, err := g.managed.GetBotStatus(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": status})
	}
}
