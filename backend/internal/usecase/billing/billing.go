package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrTariffNotFound       = errors.New("tariff not found")
	ErrAlreadySubscribed    = errors.New("organization already has an active subscription")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvoiceAlreadyPaid   = errors.New("invoice already paid")
	ErrAmountMismatch       = errors.New("payment amount does not match invoice")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrFeatureNotAvailable  = errors.New("feature not available on current plan")
	ErrLimitExceeded        = errors.New("plan limit exceeded")
)

type billingRepo interface {
	GetTariffs(ctx context.Context) ([]entity.Tariff, error)
	GetTariffByID(ctx context.Context, id int) (*entity.Tariff, error)
	GetTariffBySlug(ctx context.Context, slug string) (*entity.Tariff, error)
	GetSubscriptionByOrgID(ctx context.Context, orgID int) (*entity.SubscriptionWithTariff, error)
	CreateSubscription(ctx context.Context, sub *entity.Subscription) error
	UpdateSubscriptionStatus(ctx context.Context, id int, status string, canceledAt *time.Time) error
	UpdateSubscriptionTariff(ctx context.Context, id int, tariffID int, periodEnd time.Time) error
	GetExpiredSubscriptions(ctx context.Context) ([]entity.Subscription, error)
	CreateInvoice(ctx context.Context, inv *entity.Invoice) error
	GetInvoicesByOrgID(ctx context.Context, orgID int) ([]entity.Invoice, error)
	GetInvoiceByID(ctx context.Context, id int) (*entity.Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, id int, status string, paidAt *time.Time) error
	CreatePayment(ctx context.Context, p *entity.Payment) error
	GetPaymentsByOrgID(ctx context.Context, orgID int) ([]entity.Payment, error)
	GetPaymentByProviderID(ctx context.Context, providerPaymentID string) (*entity.Payment, error)
	UpdatePaymentStatus(ctx context.Context, id int, status string) error
}

type Usecase struct {
	logger *slog.Logger
	repo   billingRepo
}

func New(repo billingRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// ── Tariffs ───────────────────────────────────────────────────────────────────

func (uc *Usecase) GetTariffs(ctx context.Context) ([]entity.Tariff, error) {
	return uc.repo.GetTariffs(ctx)
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (uc *Usecase) GetCurrentSubscription(ctx context.Context, orgID int) (*entity.SubscriptionWithTariff, error) {
	sub, err := uc.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return sub, nil
}

func (uc *Usecase) Subscribe(ctx context.Context, orgID int, tariffSlug string) (*entity.Subscription, error) {
	// Check no active subscription
	existing, err := uc.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing subscription: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadySubscribed
	}

	tariff, err := uc.repo.GetTariffBySlug(ctx, tariffSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTariffNotFound
		}
		return nil, fmt.Errorf("get tariff: %w", err)
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0) // 1 month
	if tariff.Interval == "year" {
		periodEnd = now.AddDate(1, 0, 0)
	}

	status := "active"
	if tariff.Slug == "trial" {
		status = "trialing"
		periodEnd = now.AddDate(0, 0, 30)
	}

	sub := &entity.Subscription{
		OrgID:              orgID,
		TariffID:           tariff.ID,
		Status:             status,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
	}

	if err := uc.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	// Create invoice for paid plans
	if tariff.Price > 0 {
		inv := &entity.Invoice{
			OrgID:          orgID,
			SubscriptionID: &sub.ID,
			Amount:         tariff.Price,
			Currency:       tariff.Currency,
			Status:         "pending",
			DueDate:        now.AddDate(0, 0, 7),
		}
		if err := uc.repo.CreateInvoice(ctx, inv); err != nil {
			uc.logger.Error("failed to create invoice for new subscription", "error", err, "org_id", orgID)
		}
	}

	return sub, nil
}

func (uc *Usecase) ChangePlan(ctx context.Context, orgID int, tariffSlug string) (*entity.Subscription, error) {
	sub, err := uc.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	tariff, err := uc.repo.GetTariffBySlug(ctx, tariffSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTariffNotFound
		}
		return nil, fmt.Errorf("get tariff: %w", err)
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	if tariff.Interval == "year" {
		periodEnd = now.AddDate(1, 0, 0)
	}

	if err := uc.repo.UpdateSubscriptionTariff(ctx, sub.ID, tariff.ID, periodEnd); err != nil {
		return nil, fmt.Errorf("update subscription tariff: %w", err)
	}

	// Create invoice for upgrade
	if tariff.Price > 0 {
		inv := &entity.Invoice{
			OrgID:          orgID,
			SubscriptionID: &sub.ID,
			Amount:         tariff.Price,
			Currency:       tariff.Currency,
			Status:         "pending",
			DueDate:        now.AddDate(0, 0, 7),
		}
		if err := uc.repo.CreateInvoice(ctx, inv); err != nil {
			uc.logger.Error("failed to create invoice for plan change", "error", err, "org_id", orgID)
		}
	}

	sub.TariffID = tariff.ID
	sub.CurrentPeriodEnd = periodEnd
	return &sub.Subscription, nil
}

func (uc *Usecase) CancelSubscription(ctx context.Context, orgID int) error {
	sub, err := uc.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("get subscription: %w", err)
	}

	now := time.Now()
	return uc.repo.UpdateSubscriptionStatus(ctx, sub.ID, "canceled", &now)
}

// ── Feature gating ────────────────────────────────────────────────────────────

func (uc *Usecase) HasFeature(ctx context.Context, orgID int, feature string) (bool, error) {
	sub, err := uc.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check feature: %w", err)
	}

	f := sub.TariffFeatures
	switch feature {
	case "loyalty":
		return f.Loyalty, nil
	case "campaigns":
		return f.Campaigns, nil
	case "promotions":
		return f.Promotions, nil
	case "integrations":
		return f.Integrations, nil
	case "analytics":
		return f.Analytics, nil
	case "rfm":
		return f.RFM, nil
	case "advanced_campaigns":
		return f.AdvancedCampaigns, nil
	default:
		return false, nil
	}
}

func (uc *Usecase) CheckLimit(ctx context.Context, orgID int, limitKey string, currentValue int) (bool, error) {
	sub, err := uc.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check limit: %w", err)
	}

	l := sub.TariffLimits
	var limit int
	switch limitKey {
	case "max_clients":
		limit = l.MaxClients
	case "max_bots":
		limit = l.MaxBots
	case "max_campaigns_per_month":
		limit = l.MaxCampaignsPerMonth
	case "max_pos":
		limit = l.MaxPOS
	default:
		return true, nil
	}

	if limit == -1 {
		return true, nil // unlimited
	}
	return currentValue < limit, nil
}

// ── Invoices ──────────────────────────────────────────────────────────────────

func (uc *Usecase) GetInvoices(ctx context.Context, orgID int) ([]entity.Invoice, error) {
	return uc.repo.GetInvoicesByOrgID(ctx, orgID)
}

func (uc *Usecase) GetInvoice(ctx context.Context, orgID int, invoiceID int) (*entity.Invoice, error) {
	inv, err := uc.repo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if inv.OrgID != orgID {
		return nil, ErrInvoiceNotFound
	}
	return inv, nil
}

// ── Payments ──────────────────────────────────────────────────────────────────

func (uc *Usecase) ProcessPayment(ctx context.Context, orgID int, req entity.ProcessPaymentRequest) error {
	inv, err := uc.repo.GetInvoiceByID(ctx, req.InvoiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvoiceNotFound
		}
		return fmt.Errorf("get invoice: %w", err)
	}
	if inv.OrgID != orgID {
		return ErrInvoiceNotFound
	}
	if inv.Status == "paid" {
		return ErrInvoiceAlreadyPaid
	}
	if req.Amount != inv.Amount {
		return ErrAmountMismatch
	}

	payment := &entity.Payment{
		InvoiceID:         req.InvoiceID,
		OrgID:             orgID,
		Amount:            req.Amount,
		Currency:          inv.Currency,
		Provider:          req.Provider,
		ProviderPaymentID: &req.ProviderPaymentID,
		Status:            "pending",
	}

	if err := uc.repo.CreatePayment(ctx, payment); err != nil {
		return fmt.Errorf("create payment: %w", err)
	}

	uc.logger.Info("payment initiated",
		"org_id", orgID, "invoice_id", inv.ID,
		"amount", req.Amount, "provider", req.Provider)

	return nil
}

// ConfirmPayment is called by a payment provider webhook to mark a payment as succeeded.
func (uc *Usecase) ConfirmPayment(ctx context.Context, providerPaymentID string) error {
	payments, err := uc.repo.GetPaymentByProviderID(ctx, providerPaymentID)
	if err != nil {
		return fmt.Errorf("get payment by provider id: %w", err)
	}
	if payments == nil {
		return ErrPaymentNotFound
	}
	if payments.Status == "succeeded" {
		return nil // idempotent
	}

	if err := uc.repo.UpdatePaymentStatus(ctx, payments.ID, "succeeded"); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	now := time.Now()
	if err := uc.repo.UpdateInvoiceStatus(ctx, payments.InvoiceID, "paid", &now); err != nil {
		return fmt.Errorf("update invoice status: %w", err)
	}

	// Reactivate subscription if past_due
	inv, err := uc.repo.GetInvoiceByID(ctx, payments.InvoiceID)
	if err != nil {
		uc.logger.Error("failed to get invoice after payment confirmation", "error", err)
		return nil
	}
	if inv.SubscriptionID != nil {
		sub, err := uc.repo.GetSubscriptionByOrgID(ctx, inv.OrgID)
		if err == nil && (sub.Status == "past_due" || sub.Status == "trialing") {
			if err := uc.repo.UpdateSubscriptionStatus(ctx, sub.ID, "active", nil); err != nil {
				uc.logger.Error("failed to reactivate subscription", "error", err, "org_id", inv.OrgID)
			}
		}
	}

	uc.logger.Info("payment confirmed", "payment_id", payments.ID, "provider_payment_id", providerPaymentID)
	return nil
}

// ── Scheduler ─────────────────────────────────────────────────────────────────

func (uc *Usecase) HandleExpiredSubscriptions(ctx context.Context) error {
	expired, err := uc.repo.GetExpiredSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("get expired subscriptions: %w", err)
	}

	for _, sub := range expired {
		newStatus := "expired"
		if sub.Status == "active" {
			// Grace period: move to past_due first
			newStatus = "past_due"
		}

		if err := uc.repo.UpdateSubscriptionStatus(ctx, sub.ID, newStatus, nil); err != nil {
			uc.logger.Error("failed to expire subscription", "error", err, "sub_id", sub.ID)
			continue
		}
		uc.logger.Info("subscription status updated", "sub_id", sub.ID, "old_status", sub.Status, "new_status", newStatus)
	}

	return nil
}
