package masterbot

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
)

type adminBotRepo interface {
	GetByUserID(ctx context.Context, userID int) (*entity.AdminBotLink, error)
	CreateLinkCode(ctx context.Context, userID int, orgID int, role string, code string, expiresAt time.Time) error
	DeleteLink(ctx context.Context, userID int) error
}

type Group struct {
	repo      adminBotRepo
	jwtSecret string
}

func New(repo adminBotRepo, jwtSecret string) *Group {
	return &Group{repo: repo, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/admin-bot"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetStatus,
		g.handleGenerateCode,
		g.handleUnlink,
	}
}

func (g *Group) handleGetStatus() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/status", func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		link, err := g.repo.GetByUserID(c.Request.Context(), userID.(int))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
				c.JSON(http.StatusOK, entity.AdminBotStatus{Linked: false})
				return
			}
			slog.Error("admin bot status", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, entity.AdminBotStatus{
			Linked:     link.TelegramID != nil,
			TelegramID: link.TelegramID,
			Role:       link.Role,
		})
	}
}

func (g *Group) handleGenerateCode() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/link-code", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		orgID, _ := c.Get("org_id")

		// Generate 6-char code
		b := make([]byte, 3)
		if _, err := rand.Read(b); err != nil {
			slog.Error("generate link code", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		code := strings.ToUpper(hex.EncodeToString(b))
		expiresAt := time.Now().Add(10 * time.Minute)

		if err := g.repo.CreateLinkCode(c.Request.Context(), userID.(int), orgID.(int), "owner", code, expiresAt); err != nil {
			slog.Error("create link code", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, entity.GenerateLinkCodeResponse{
			Code:      code,
			ExpiresAt: expiresAt,
		})
	}
}

func (g *Group) handleUnlink() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/link", func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		if err := g.repo.DeleteLink(c.Request.Context(), userID.(int)); err != nil {
			slog.Error("unlink admin bot", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "admin bot unlinked"})
	}
}
