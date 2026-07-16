package lunch

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	lunchUC "revisitr/internal/usecase/lunch"
)

type lunchUsecase interface {
	GetProgram(ctx context.Context, orgID, botID int) (*entity.LunchProgram, error)
	UpsertProgram(ctx context.Context, orgID, botID int, req entity.UpsertLunchProgramRequest) (*entity.LunchProgram, error)
	CreateCourse(ctx context.Context, orgID, botID int, req entity.CreateLunchCourseRequest) (*entity.LunchCourse, error)
	UpdateCourse(ctx context.Context, orgID, courseID int, req entity.UpdateLunchCourseRequest) error
	DeleteCourse(ctx context.Context, orgID, courseID int) error
	CreateFormat(ctx context.Context, orgID, botID int, req entity.CreateLunchFormatRequest) (*entity.LunchFormat, error)
	UpdateFormat(ctx context.Context, orgID, formatID int, req entity.UpdateLunchFormatRequest) error
	DeleteFormat(ctx context.Context, orgID, formatID int) error
	SetAvailability(ctx context.Context, orgID, botID int, slots []entity.LunchAvailability) error
	ListOrders(ctx context.Context, orgID, botID int, status string) ([]entity.LunchOrder, error)
	UpdateOrderStatus(ctx context.Context, orgID, orderID int, status string) error
}

type Group struct {
	uc        lunchUsecase
	jwtSecret string
}

func New(uc lunchUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

// Path deviates from the spec's /bots/{botId}/lunch/... nesting: the bots
// group already owns /api/v1/bots with the :id param, so lunch lives in its
// own group (same precedent as /api/v1/module-settings).
func (g *Group) Path() string {
	return "/api/v1/lunch"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetProgram,
		g.handleUpsertProgram,
		g.handleCreateCourse,
		g.handleUpdateCourse,
		g.handleDeleteCourse,
		g.handleCreateFormat,
		g.handleUpdateFormat,
		g.handleDeleteFormat,
		g.handleSetAvailability,
		g.handleListOrders,
		g.handleUpdateOrderStatus,
	}
}

func pathID(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name + " id"})
		return 0, false
	}
	return id, true
}

func (g *Group) handleGetProgram() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/bots/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		botID, ok := pathID(c, "bot")
		if !ok {
			return
		}

		program, err := g.uc.GetProgram(c.Request.Context(), orgID.(int), botID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, program)
	}
}

func (g *Group) handleUpsertProgram() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/bots/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		botID, ok := pathID(c, "bot")
		if !ok {
			return
		}

		var req entity.UpsertLunchProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		program, err := g.uc.UpsertProgram(c.Request.Context(), orgID.(int), botID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, program)
	}
}

func (g *Group) handleCreateCourse() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/bots/:id/courses", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		botID, ok := pathID(c, "bot")
		if !ok {
			return
		}

		var req entity.CreateLunchCourseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		course, err := g.uc.CreateCourse(c.Request.Context(), orgID.(int), botID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, course)
	}
}

func (g *Group) handleUpdateCourse() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/courses/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		courseID, ok := pathID(c, "course")
		if !ok {
			return
		}

		var req entity.UpdateLunchCourseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateCourse(c.Request.Context(), orgID.(int), courseID, req); err != nil {
			handleError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (g *Group) handleDeleteCourse() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/courses/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		courseID, ok := pathID(c, "course")
		if !ok {
			return
		}

		if err := g.uc.DeleteCourse(c.Request.Context(), orgID.(int), courseID); err != nil {
			handleError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (g *Group) handleCreateFormat() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/bots/:id/formats", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		botID, ok := pathID(c, "bot")
		if !ok {
			return
		}

		var req entity.CreateLunchFormatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		format, err := g.uc.CreateFormat(c.Request.Context(), orgID.(int), botID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, format)
	}
}

func (g *Group) handleUpdateFormat() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/formats/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		formatID, ok := pathID(c, "format")
		if !ok {
			return
		}

		var req entity.UpdateLunchFormatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateFormat(c.Request.Context(), orgID.(int), formatID, req); err != nil {
			handleError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (g *Group) handleDeleteFormat() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/formats/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		formatID, ok := pathID(c, "format")
		if !ok {
			return
		}

		if err := g.uc.DeleteFormat(c.Request.Context(), orgID.(int), formatID); err != nil {
			handleError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (g *Group) handleSetAvailability() (string, string, gin.HandlerFunc) {
	return http.MethodPut, "/bots/:id/availability", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		botID, ok := pathID(c, "bot")
		if !ok {
			return
		}

		var req entity.SetLunchAvailabilityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.SetAvailability(c.Request.Context(), orgID.(int), botID, req.Slots); err != nil {
			handleError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (g *Group) handleListOrders() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/bots/:id/orders", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		botID, ok := pathID(c, "bot")
		if !ok {
			return
		}

		orders, err := g.uc.ListOrders(c.Request.Context(), orgID.(int), botID, c.Query("status"))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

func (g *Group) handleUpdateOrderStatus() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/orders/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")
		orderID, ok := pathID(c, "order")
		if !ok {
			return
		}

		var req entity.UpdateLunchOrderStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := g.uc.UpdateOrderStatus(c.Request.Context(), orgID.(int), orderID, req.Status); err != nil {
			handleError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, lunchUC.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, lunchUC.ErrNotOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, lunchUC.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		slog.Error("lunch handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
