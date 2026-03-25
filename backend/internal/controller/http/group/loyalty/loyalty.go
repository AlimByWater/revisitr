package loyalty

import (
	"context"
	"errors"
	"log/slog"
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
	CalculateBonus(ctx context.Context, clientID, programID int, checkAmount float64) (float64, error)
	EarnFromCheck(ctx context.Context, clientID, programID int, checkAmount float64) (*entity.ClientLoyalty, error)
	ReservePoints(ctx context.Context, clientID, programID int, amount float64) (int, error)
	ConfirmReserve(ctx context.Context, reserveID int) (*entity.ClientLoyalty, error)
	CancelReserve(ctx context.Context, reserveID int) error
}

type Option func(*Group)

func WithFeatureGate(gate gin.HandlerFunc) Option {
	return func(g *Group) { g.featureGate = gate }
}

type Group struct {
	uc          loyaltyUsecase
	jwtSecret   string
	featureGate gin.HandlerFunc
}

func New(uc loyaltyUsecase, jwtSecret string, opts ...Option) *Group {
	g := &Group{uc: uc, jwtSecret: jwtSecret}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func (g *Group) ExtraMiddleware() []gin.HandlerFunc {
	if g.featureGate != nil {
		return []gin.HandlerFunc{g.featureGate}
	}
	return nil
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
		g.handleEarnFromCheck,
		g.handleReservePoints,
		g.handleConfirmReserve,
		g.handleCancelReserve,
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
			slog.Error("create loyalty program", "error", err, "org_id", orgID)
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
			slog.Error("list loyalty programs", "error", err, "org_id", orgID)
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

type earnFromCheckRequest struct {
	ClientID    int     `json:"client_id" binding:"required"`
	ProgramID   int     `json:"program_id" binding:"required"`
	CheckAmount float64 `json:"check_amount" binding:"required,gt=0"`
}

type reserveRequest struct {
	ClientID  int     `json:"client_id" binding:"required"`
	ProgramID int     `json:"program_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
}

func (g *Group) handleEarnFromCheck() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/earn-from-check", func(c *gin.Context) {
		var req earnFromCheckRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cl, err := g.uc.EarnFromCheck(c.Request.Context(), req.ClientID, req.ProgramID, req.CheckAmount)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, cl)
	}
}

func (g *Group) handleReservePoints() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/reserve", func(c *gin.Context) {
		var req reserveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		reserveID, err := g.uc.ReservePoints(c.Request.Context(), req.ClientID, req.ProgramID, req.Amount)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"reserve_id": reserveID})
	}
}

func (g *Group) handleConfirmReserve() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/reserve/:id/confirm", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reserve id"})
			return
		}

		cl, err := g.uc.ConfirmReserve(c.Request.Context(), id)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, cl)
	}
}

func (g *Group) handleCancelReserve() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/reserve/:id/cancel", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reserve id"})
			return
		}

		if err := g.uc.CancelReserve(c.Request.Context(), id); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "reserve cancelled"})
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
	case errors.Is(err, loyaltyUC.ErrReserveExpired):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, loyaltyUC.ErrReserveNotPending):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		slog.Error("loyalty handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
