package loyalty

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	loyaltyUC "revisitr/internal/usecase/loyalty"
)

type loyaltyUsecase interface {
	CreateProgram(ctx context.Context, orgID int, req *entity.CreateProgramRequest) (*entity.LoyaltyProgram, error)
	GetPrograms(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
	GetProgram(ctx context.Context, id, orgID int) (*entity.LoyaltyProgram, error)
	UpdateProgram(ctx context.Context, id, orgID int, req *entity.UpdateProgramRequest) (*entity.LoyaltyProgram, error)
	CreateLevel(ctx context.Context, programID, orgID int, req *entity.CreateLevelRequest) (*entity.LoyaltyLevel, error)
	UpdateLevels(ctx context.Context, programID, orgID int, req *entity.BatchUpdateLevelsRequest) ([]entity.LoyaltyLevel, error)
	DeleteLevel(ctx context.Context, programID, orgID, levelID int) error
}

type Group struct {
	uc        loyaltyUsecase
	jwtSecret string
}

func New(uc loyaltyUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/loyalty"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleCreateProgram,
		g.handleListPrograms,
		g.handleGetProgram,
		g.handleUpdateProgram,
		g.handleCreateLevel,
		g.handleBatchUpdateLevels,
		g.handleDeleteLevel,
	}
}

func (g *Group) handleCreateProgram() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/programs", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		var req entity.CreateProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		program, err := g.uc.CreateProgram(c.Request.Context(), orgID, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, program)
	}
}

func (g *Group) handleListPrograms() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/programs", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		programs, err := g.uc.GetPrograms(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, programs)
	}
}

func (g *Group) handleGetProgram() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/programs/:id", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid program id"})
			return
		}

		program, err := g.uc.GetProgram(c.Request.Context(), id, orgID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, program)
	}
}

func (g *Group) handleUpdateProgram() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/programs/:id", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid program id"})
			return
		}

		var req entity.UpdateProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		program, err := g.uc.UpdateProgram(c.Request.Context(), id, orgID, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, program)
	}
}

func (g *Group) handleCreateLevel() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/programs/:id/levels", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		programID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid program id"})
			return
		}

		var req entity.CreateLevelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		level, err := g.uc.CreateLevel(c.Request.Context(), programID, orgID, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, level)
	}
}

func (g *Group) handleBatchUpdateLevels() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/programs/:id/levels", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		programID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid program id"})
			return
		}

		var req entity.BatchUpdateLevelsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		levels, err := g.uc.UpdateLevels(c.Request.Context(), programID, orgID, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, levels)
	}
}

func (g *Group) handleDeleteLevel() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/programs/:programId/levels/:levelId", func(c *gin.Context) {
		orgID := c.GetInt("org_id")

		programID, err := strconv.Atoi(c.Param("programId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid program id"})
			return
		}

		levelID, err := strconv.Atoi(c.Param("levelId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid level id"})
			return
		}

		if err := g.uc.DeleteLevel(c.Request.Context(), programID, orgID, levelID); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "level deleted"})
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, loyaltyUC.ErrProgramNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, loyaltyUC.ErrNotProgramOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, loyaltyUC.ErrInsufficientPoints):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
