package billing

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/entity"
	billingUC "revisitr/internal/usecase/billing"
)

type billingUsecase interface {
	GetTariffs(ctx context.Context) ([]entity.Tariff, error)
	GetCurrentSubscription(ctx context.Context, orgID int) (*entity.SubscriptionWithTariff, error)
	Subscribe(ctx context.Context, orgID int, tariffSlug string) (*entity.Subscription, error)
	ChangePlan(ctx context.Context, orgID int, tariffSlug string) (*entity.Subscription, error)
	CancelSubscription(ctx context.Context, orgID int) error
	GetInvoices(ctx context.Context, orgID int) ([]entity.Invoice, error)
	GetInvoice(ctx context.Context, orgID int, invoiceID int) (*entity.Invoice, error)
	ProcessPayment(ctx context.Context, orgID int, req entity.ProcessPaymentRequest) error
	ConfirmPayment(ctx context.Context, providerPaymentID string) error
}

type Group struct {
	uc        billingUsecase
	jwtSecret string
}

func New(uc billingUsecase, jwtSecret string) *Group {
	return &Group{uc: uc, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/billing"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleGetTariffs,
		g.handleGetSubscription,
		g.handleSubscribe,
		g.handleChangePlan,
		g.handleCancelSubscription,
		g.handleGetInvoices,
		g.handleGetInvoice,
		g.handleProcessPayment,
	}
}

// PublicHandlers returns handlers that do not require JWT (e.g. payment webhooks).
func (g *Group) PublicHandlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handlePaymentWebhook,
	}
}

func (g *Group) handleGetTariffs() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/tariffs", func(c *gin.Context) {
		tariffs, err := g.uc.GetTariffs(c.Request.Context())
		if err != nil {
			slog.Error("billing get tariffs", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, tariffs)
	}
}

func (g *Group) handleGetSubscription() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/subscription", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		sub, err := g.uc.GetCurrentSubscription(c.Request.Context(), orgID.(int))
		if err != nil {
			if errors.Is(err, billingUC.ErrSubscriptionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription"})
				return
			}
			slog.Error("billing get subscription", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

func (g *Group) handleSubscribe() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/subscription", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.CreateSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sub, err := g.uc.Subscribe(c.Request.Context(), orgID.(int), req.TariffSlug)
		if err != nil {
			switch {
			case errors.Is(err, billingUC.ErrAlreadySubscribed):
				c.JSON(http.StatusConflict, gin.H{"error": "already subscribed"})
			case errors.Is(err, billingUC.ErrTariffNotFound):
				c.JSON(http.StatusBadRequest, gin.H{"error": "tariff not found"})
			default:
				slog.Error("billing subscribe", "error", err, "org_id", orgID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		c.JSON(http.StatusCreated, sub)
	}
}

func (g *Group) handleChangePlan() (string, string, gin.HandlerFunc) {
	return http.MethodPatch, "/subscription", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.ChangeSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sub, err := g.uc.ChangePlan(c.Request.Context(), orgID.(int), req.TariffSlug)
		if err != nil {
			switch {
			case errors.Is(err, billingUC.ErrSubscriptionNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription"})
			case errors.Is(err, billingUC.ErrTariffNotFound):
				c.JSON(http.StatusBadRequest, gin.H{"error": "tariff not found"})
			default:
				slog.Error("billing change plan", "error", err, "org_id", orgID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

func (g *Group) handleCancelSubscription() (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/subscription", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		err := g.uc.CancelSubscription(c.Request.Context(), orgID.(int))
		if err != nil {
			if errors.Is(err, billingUC.ErrSubscriptionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription"})
				return
			}
			slog.Error("billing cancel", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "subscription canceled"})
	}
}

func (g *Group) handleGetInvoices() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/invoices", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		invoices, err := g.uc.GetInvoices(c.Request.Context(), orgID.(int))
		if err != nil {
			slog.Error("billing get invoices", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, invoices)
	}
}

func (g *Group) handleGetInvoice() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/invoices/:invoiceId", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		invoiceID, err := strconv.Atoi(c.Param("invoiceId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice id"})
			return
		}

		invoice, err := g.uc.GetInvoice(c.Request.Context(), orgID.(int), invoiceID)
		if err != nil {
			if errors.Is(err, billingUC.ErrInvoiceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
				return
			}
			slog.Error("billing get invoice", "error", err, "org_id", orgID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, invoice)
	}
}

func (g *Group) handleProcessPayment() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/payments", func(c *gin.Context) {
		orgID, _ := c.Get("org_id")

		var req entity.ProcessPaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := g.uc.ProcessPayment(c.Request.Context(), orgID.(int), req)
		if err != nil {
			switch {
			case errors.Is(err, billingUC.ErrInvoiceNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			case errors.Is(err, billingUC.ErrInvoiceAlreadyPaid):
				c.JSON(http.StatusConflict, gin.H{"error": "invoice already paid"})
			case errors.Is(err, billingUC.ErrAmountMismatch):
				c.JSON(http.StatusBadRequest, gin.H{"error": "payment amount does not match invoice"})
			default:
				slog.Error("billing process payment", "error", err, "org_id", orgID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "pending"})
	}
}

func (g *Group) handlePaymentWebhook() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/webhook", func(c *gin.Context) {
		// TODO: verify provider signature (YooKassa/CloudPayments) before processing
		var body struct {
			Event  string `json:"event"`
			Object struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"object"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if body.Object.Status != "succeeded" {
			c.JSON(http.StatusOK, gin.H{"status": "ignored"})
			return
		}

		if err := g.uc.ConfirmPayment(c.Request.Context(), body.Object.ID); err != nil {
			if errors.Is(err, billingUC.ErrPaymentNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
				return
			}
			slog.Error("billing webhook", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
