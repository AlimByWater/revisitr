package promotions

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	promotionsUC "revisitr/internal/usecase/promotions"
)

type promotionsUsecase interface {
	Create(ctx context.Context, orgID int, req *entity.CreatePromotionRequest) (*entity.Promotion, error)
	List(ctx context.Context, orgID int) ([]entity.Promotion, error)
	GetByID(ctx context.Context, id, orgID int) (*entity.Promotion, error)
	Update(ctx context.Context, id, orgID int, req *entity.UpdatePromotionRequest) (*entity.Promotion, error)
	Delete(ctx context.Context, id, orgID int) error
	CreatePromoCode(ctx context.Context, orgID int, req *entity.CreatePromoCodeRequest) (*entity.PromoCode, error)
	ListPromoCodes(ctx context.Context, orgID int) ([]entity.PromoCode, error)
	DeactivatePromoCode(ctx context.Context, id, orgID int) error
	ValidatePromoCode(ctx context.Context, orgID int, code string, clientID int, orderAmount float64) (*entity.PromoCodeValidation, error)
	ActivatePromoCode(ctx context.Context, orgID int, code string, clientID int) (*entity.PromoResult, error)
	GetChannelAnalytics(ctx context.Context, orgID int) ([]entity.PromoChannelAnalytics, error)
	GenerateCode(ctx context.Context, orgID int) (string, error)
	GetPromoCodesByPromotion(ctx context.Context, promotionID, orgID int) ([]entity.PromoCode, error)
}

type Group struct {
	uc        promotionsUsecase
	jwtSecret string
}

func New(uc promotionsUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/promotions"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		// Static routes first
		g.handleCreate,
		g.handleList,
		g.handleListPromoCodes,
		g.handleCreatePromoCode,
		g.handleValidatePromoCode,
		g.handleActivatePromoCode,
		g.handleGeneratePromoCode,
		g.handleChannelAnalytics,
		// Parameterized routes
		g.handleGet,
		g.handleUpdate,
		g.handleDelete,
		g.handleDeactivatePromoCode,
		g.handleGetPromoCodesByPromotion,
	}
}

func (g *Group) handleCreate() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreatePromotionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p, err := g.uc.Create(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create promotion", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, p)
	}
}

func (g *Group) handleList() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		promotions, err := g.uc.List(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list promotions", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, promotions)
	}
}

func (g *Group) handleGet() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promotion id"})
			return
		}

		p, err := g.uc.GetByID(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

func (g *Group) handleUpdate() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promotion id"})
			return
		}

		var req entity.UpdatePromotionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p, err := g.uc.Update(c.Request.Context(), id, orgID.(int), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

func (g *Group) handleDelete() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promotion id"})
			return
		}

		if err := g.uc.Delete(c.Request.Context(), id, orgID.(int)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "promotion deleted"})
	}
}

func (g *Group) handleListPromoCodes() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/promo-codes", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		codes, err := g.uc.ListPromoCodes(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("list promo codes", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, codes)
	}
}

func (g *Group) handleCreatePromoCode() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/promo-codes", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreatePromoCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pc, err := g.uc.CreatePromoCode(c.Request.Context(), orgID.(int), &req)
		if err != nil {
			slog.Error("create promo code", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, pc)
	}
}

func (g *Group) handleDeactivatePromoCode() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/promo-codes/:id", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promo code id"})
			return
		}

		if err := g.uc.DeactivatePromoCode(c.Request.Context(), id, orgID.(int)); err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "promo code deactivated"})
	}
}

func (g *Group) handleValidatePromoCode() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/promo-codes/validate", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req struct {
			Code        string  `json:"code"         binding:"required"`
			ClientID    int     `json:"client_id"    binding:"required"`
			OrderAmount float64 `json:"order_amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := g.uc.ValidatePromoCode(c.Request.Context(), orgID.(int), req.Code, req.ClientID, req.OrderAmount)
		if err != nil {
			slog.Error("validate promo code", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (g *Group) handleActivatePromoCode() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/promo-codes/activate", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req struct {
			Code     string `json:"code"      binding:"required"`
			ClientID int    `json:"client_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := g.uc.ActivatePromoCode(c.Request.Context(), orgID.(int), req.Code, req.ClientID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (g *Group) handleGeneratePromoCode() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/promo-codes/generate", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		code, err := g.uc.GenerateCode(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("generate promo code", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": code})
	}
}

func (g *Group) handleChannelAnalytics() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/promo-codes/analytics", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		analytics, err := g.uc.GetChannelAnalytics(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("get channel analytics", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, analytics)
	}
}

func (g *Group) handleGetPromoCodesByPromotion() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:id/codes", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid promotion id"})
			return
		}

		codes, err := g.uc.GetPromoCodesByPromotion(c.Request.Context(), id, orgID.(int))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, codes)
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, promotionsUC.ErrPromotionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, promotionsUC.ErrNotPromotionOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, promotionsUC.ErrPromoCodeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, promotionsUC.ErrPromoCodeInactive),
		errors.Is(err, promotionsUC.ErrPromoCodeNotActive),
		errors.Is(err, promotionsUC.ErrPromoCodeExpired),
		errors.Is(err, promotionsUC.ErrPromoCodeLimitReached),
		errors.Is(err, promotionsUC.ErrPerUserLimitExceeded),
		errors.Is(err, promotionsUC.ErrMinAmountNotMet):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		slog.Error("promotion handler error", "error", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
