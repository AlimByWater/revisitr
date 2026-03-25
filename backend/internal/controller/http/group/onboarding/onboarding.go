package onboarding

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
)

type onboardingUsecase interface {
	GetState(ctx context.Context, orgID int) (*entity.Organization, error)
	UpdateStep(ctx context.Context, orgID int, req entity.UpdateOnboardingRequest) (*entity.OnboardingState, error)
	Complete(ctx context.Context, orgID int) error
	Reset(ctx context.Context, orgID int) error
}

type Group struct {
	uc        onboardingUsecase
	jwtSecret string
}

func New(uc onboardingUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/onboarding"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetState,
		g.handleUpdateStep,
		g.handleComplete,
		g.handleReset,
	}
}

func (g *Group) handleGetState() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		org, err := g.uc.GetState(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("onboarding get state", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"onboarding_completed": org.OnboardingCompleted,
			"onboarding_state":    org.OnboardingState,
		})
	}
}

func (g *Group) handleUpdateStep() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.UpdateOnboardingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		state, err := g.uc.UpdateStep(c.Request.Context(), orgID.(int), req)
		if err != nil {
			slog.Error("onboarding update step", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, state)
	}
}

func (g *Group) handleComplete() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/complete", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		if err := g.uc.Complete(c.Request.Context(), orgID.(int)); err != nil {
			slog.Error("onboarding complete", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "onboarding completed"})
	}
}

func (g *Group) handleReset() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/reset", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		if err := g.uc.Reset(c.Request.Context(), orgID.(int)); err != nil {
			slog.Error("onboarding reset", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "onboarding reset"})
	}
}
