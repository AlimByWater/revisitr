package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

type Billing struct {
	pg *Module
}

func NewBilling(pg *Module) *Billing {
	return &Billing{pg: pg}
}

// ── Tariffs ───────────────────────────────────────────────────────────────────

func (r *Billing) GetTariffs(ctx context.Context) ([]entity.Tariff, error) {
	var tariffs []entity.Tariff
	err := r.pg.DB().SelectContext(ctx, &tariffs,
		"SELECT * FROM tariffs WHERE active = true ORDER BY sort_order")
	if err != nil {
		return nil, fmt.Errorf("billing.GetTariffs: %w", err)
	}
	return tariffs, nil
}

func (r *Billing) GetTariffByID(ctx context.Context, id int) (*entity.Tariff, error) {
	var t entity.Tariff
	err := r.pg.DB().GetContext(ctx, &t, "SELECT * FROM tariffs WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("billing.GetTariffByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("billing.GetTariffByID: %w", err)
	}
	return &t, nil
}

func (r *Billing) GetTariffBySlug(ctx context.Context, slug string) (*entity.Tariff, error) {
	var t entity.Tariff
	err := r.pg.DB().GetContext(ctx, &t, "SELECT * FROM tariffs WHERE slug = $1", slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("billing.GetTariffBySlug: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("billing.GetTariffBySlug: %w", err)
	}
	return &t, nil
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (r *Billing) GetSubscriptionByOrgID(ctx context.Context, orgID int) (*entity.SubscriptionWithTariff, error) {
	var sub entity.SubscriptionWithTariff
	err := r.pg.DB().GetContext(ctx, &sub, `
		SELECT s.*,
		       t.name     AS tariff_name,
		       t.slug     AS tariff_slug,
		       t.price    AS tariff_price,
		       t.features AS tariff_features,
		       t.limits   AS tariff_limits
		FROM subscriptions s
		JOIN tariffs t ON t.id = s.tariff_id
		WHERE s.org_id = $1 AND s.status IN ('trialing', 'active', 'past_due')
		ORDER BY s.created_at DESC
		LIMIT 1`, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("billing.GetSubscriptionByOrgID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("billing.GetSubscriptionByOrgID: %w", err)
	}
	return &sub, nil
}

func (r *Billing) CreateSubscription(ctx context.Context, sub *entity.Subscription) error {
	err := r.pg.DB().QueryRowContext(ctx, `
		INSERT INTO subscriptions (org_id, tariff_id, status, current_period_start, current_period_end)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		sub.OrgID, sub.TariffID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("billing.CreateSubscription: %w", err)
	}
	return nil
}

func (r *Billing) UpdateSubscriptionStatus(ctx context.Context, id int, status string, canceledAt *time.Time) error {
	result, err := r.pg.DB().ExecContext(ctx, `
		UPDATE subscriptions SET status = $1, canceled_at = $2, updated_at = now()
		WHERE id = $3`, status, canceledAt, id)
	if err != nil {
		return fmt.Errorf("billing.UpdateSubscriptionStatus: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("billing.UpdateSubscriptionStatus: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Billing) UpdateSubscriptionTariff(ctx context.Context, id int, tariffID int, periodEnd time.Time) error {
	result, err := r.pg.DB().ExecContext(ctx, `
		UPDATE subscriptions SET tariff_id = $1, current_period_end = $2, updated_at = now()
		WHERE id = $3`, tariffID, periodEnd, id)
	if err != nil {
		return fmt.Errorf("billing.UpdateSubscriptionTariff: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("billing.UpdateSubscriptionTariff: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *Billing) GetExpiredSubscriptions(ctx context.Context) ([]entity.Subscription, error) {
	var subs []entity.Subscription
	err := r.pg.DB().SelectContext(ctx, &subs, `
		SELECT * FROM subscriptions
		WHERE status IN ('trialing', 'active', 'past_due')
		  AND current_period_end < now()`)
	if err != nil {
		return nil, fmt.Errorf("billing.GetExpiredSubscriptions: %w", err)
	}
	return subs, nil
}

// ── Invoices ──────────────────────────────────────────────────────────────────

func (r *Billing) CreateInvoice(ctx context.Context, inv *entity.Invoice) error {
	err := r.pg.DB().QueryRowContext(ctx, `
		INSERT INTO invoices (org_id, subscription_id, amount, currency, status, due_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		inv.OrgID, inv.SubscriptionID, inv.Amount, inv.Currency, inv.Status, inv.DueDate,
	).Scan(&inv.ID, &inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("billing.CreateInvoice: %w", err)
	}
	return nil
}

func (r *Billing) GetInvoicesByOrgID(ctx context.Context, orgID int) ([]entity.Invoice, error) {
	var invoices []entity.Invoice
	err := r.pg.DB().SelectContext(ctx, &invoices,
		"SELECT * FROM invoices WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("billing.GetInvoicesByOrgID: %w", err)
	}
	return invoices, nil
}

func (r *Billing) GetInvoiceByID(ctx context.Context, id int) (*entity.Invoice, error) {
	var inv entity.Invoice
	err := r.pg.DB().GetContext(ctx, &inv, "SELECT * FROM invoices WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("billing.GetInvoiceByID: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("billing.GetInvoiceByID: %w", err)
	}
	return &inv, nil
}

func (r *Billing) UpdateInvoiceStatus(ctx context.Context, id int, status string, paidAt *time.Time) error {
	result, err := r.pg.DB().ExecContext(ctx, `
		UPDATE invoices SET status = $1, paid_at = $2 WHERE id = $3`,
		status, paidAt, id)
	if err != nil {
		return fmt.Errorf("billing.UpdateInvoiceStatus: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("billing.UpdateInvoiceStatus: %w", sql.ErrNoRows)
	}
	return nil
}

// ── Payments ──────────────────────────────────────────────────────────────────

func (r *Billing) CreatePayment(ctx context.Context, p *entity.Payment) error {
	err := r.pg.DB().QueryRowContext(ctx, `
		INSERT INTO payments (invoice_id, org_id, amount, currency, provider, provider_payment_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`,
		p.InvoiceID, p.OrgID, p.Amount, p.Currency, p.Provider, p.ProviderPaymentID, p.Status,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("billing.CreatePayment: %w", err)
	}
	return nil
}

func (r *Billing) GetPaymentsByOrgID(ctx context.Context, orgID int) ([]entity.Payment, error) {
	var payments []entity.Payment
	err := r.pg.DB().SelectContext(ctx, &payments,
		"SELECT * FROM payments WHERE org_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("billing.GetPaymentsByOrgID: %w", err)
	}
	return payments, nil
}

func (r *Billing) GetPaymentByProviderID(ctx context.Context, providerPaymentID string) (*entity.Payment, error) {
	var p entity.Payment
	err := r.pg.DB().GetContext(ctx, &p,
		"SELECT * FROM payments WHERE provider_payment_id = $1", providerPaymentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("billing.GetPaymentByProviderID: %w", err)
	}
	return &p, nil
}

func (r *Billing) UpdatePaymentStatus(ctx context.Context, id int, status string) error {
	_, err := r.pg.DB().ExecContext(ctx,
		"UPDATE payments SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("billing.UpdatePaymentStatus: %w", err)
	}
	return nil
}
