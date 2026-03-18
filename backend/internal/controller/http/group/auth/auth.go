package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/entity"
	authUC "revisitr/internal/usecase/auth"
)

type authUsecase interface {
	Register(ctx context.Context, req *entity.RegisterRequest) (*entity.AuthResponse, error)
	Login(ctx context.Context, req *entity.LoginRequest) (*entity.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*entity.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type Group struct {
	uc authUsecase
}

func New(uc authUsecase) *Group {
	return &Group{uc: uc}
}

func (g *Group) Path() string {
	return "/api/v1/auth"
}

func (g *Group) Auth() gin.HandlerFunc {
	return nil
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleRegister,
		g.handleLogin,
		g.handleRefresh,
		g.handleLogout,
	}
}

func (g *Group) handleRegister() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/register", func(c *gin.Context) {
		var req entity.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := g.uc.Register(c.Request.Context(), &req)
		if err != nil {
			if errors.Is(err, authUC.ErrUserExists) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func (g *Group) handleLogin() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/login", func(c *gin.Context) {
		var req entity.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := g.uc.Login(c.Request.Context(), &req)
		if err != nil {
			if errors.Is(err, authUC.ErrInvalidCredentials) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (g *Group) handleRefresh() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/refresh", func(c *gin.Context) {
		var req entity.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tokens, err := g.uc.Refresh(c.Request.Context(), req.RefreshToken)
		if err != nil {
			if errors.Is(err, authUC.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, tokens)
	}
}

func (g *Group) handleLogout() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/logout", func(c *gin.Context) {
		var req entity.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
	}
}
