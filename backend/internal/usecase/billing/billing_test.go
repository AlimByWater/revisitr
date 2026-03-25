package billing

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"revisitr/internal/entity"
)

type mockBillingRepo struct {
	getTariffsFn               func(ctx context.Context) ([]entity.Tariff, error)
	getTariffByIDFn            func(ctx context.Context, id int) (*entity.Tariff, error)
	getTariffBySlugFn          func(ctx context.Context, slug string) (*entity.Tariff, error)
	getSubscriptionByOrgIDFn   func(ctx context.Context, orgID int) (*entity.SubscriptionWithTariff, error)
	createSubscriptionFn       func(ctx context.Context, sub *entity.Subscription) error
	updateSubscriptionStatusFn func(ctx context.Context, id int, status string, canceledAt *time.Time) error
	updateSubscriptionTariffFn func(ctx context.Context, id int, tariffID int, periodEnd time.Time) error
	getExpiredSubscriptionsFn  func(ctx context.Context) ([]entity.Subscription, error)
	createInvoiceFn            func(ctx context.Context, inv *entity.Invoice) error
	getInvoicesByOrgIDFn       func(ctx context.Context, orgID int) ([]entity.Invoice, error)
	getInvoiceByIDFn           func(ctx context.Context, id int) (*entity.Invoice, error)
	updateInvoiceStatusFn      func(ctx context.Context, id int, status string, paidAt *time.Time) error
	createPaymentFn            func(ctx context.Context, p *entity.Payment) error
	getPaymentsByOrgIDFn       func(ctx context.Context, orgID int) ([]entity.Payment, error)
	getPaymentByProviderIDFn   func(ctx context.Context, providerPaymentID string) (*entity.Payment, error)
	updatePaymentStatusFn      func(ctx context.Context, id int, status string) error
}

func (m *mockBillingRepo) GetTariffs(ctx context.Context) ([]entity.Tariff, error) {
	return m.getTariffsFn(ctx)
}
func (m *mockBillingRepo) GetTariffByID(ctx context.Context, id int) (*entity.Tariff, error) {
	return m.getTariffByIDFn(ctx, id)
}
func (m *mockBillingRepo) GetTariffBySlug(ctx context.Context, slug string) (*entity.Tariff, error) {
	return m.getTariffBySlugFn(ctx, slug)
}
func (m *mockBillingRepo) GetSubscriptionByOrgID(ctx context.Context, orgID int) (*entity.SubscriptionWithTariff, error) {
	return m.getSubscriptionByOrgIDFn(ctx, orgID)
}
func (m *mockBillingRepo) CreateSubscription(ctx context.Context, sub *entity.Subscription) error {
	return m.createSubscriptionFn(ctx, sub)
}
func (m *mockBillingRepo) UpdateSubscriptionStatus(ctx context.Context, id int, status string, canceledAt *time.Time) error {
	return m.updateSubscriptionStatusFn(ctx, id, status, canceledAt)
}
func (m *mockBillingRepo) UpdateSubscriptionTariff(ctx context.Context, id int, tariffID int, periodEnd time.Time) error {
	return m.updateSubscriptionTariffFn(ctx, id, tariffID, periodEnd)
}
func (m *mockBillingRepo) GetExpiredSubscriptions(ctx context.Context) ([]entity.Subscription, error) {
	return m.getExpiredSubscriptionsFn(ctx)
}
func (m *mockBillingRepo) CreateInvoice(ctx context.Context, inv *entity.Invoice) error {
	return m.createInvoiceFn(ctx, inv)
}
func (m *mockBillingRepo) GetInvoicesByOrgID(ctx context.Context, orgID int) ([]entity.Invoice, error) {
	return m.getInvoicesByOrgIDFn(ctx, orgID)
}
func (m *mockBillingRepo) GetInvoiceByID(ctx context.Context, id int) (*entity.Invoice, error) {
	return m.getInvoiceByIDFn(ctx, id)
}
func (m *mockBillingRepo) UpdateInvoiceStatus(ctx context.Context, id int, status string, paidAt *time.Time) error {
	return m.updateInvoiceStatusFn(ctx, id, status, paidAt)
}
func (m *mockBillingRepo) CreatePayment(ctx context.Context, p *entity.Payment) error {
	return m.createPaymentFn(ctx, p)
}
func (m *mockBillingRepo) GetPaymentsByOrgID(ctx context.Context, orgID int) ([]entity.Payment, error) {
	return m.getPaymentsByOrgIDFn(ctx, orgID)
}
func (m *mockBillingRepo) GetPaymentByProviderID(ctx context.Context, providerPaymentID string) (*entity.Payment, error) {
	return m.getPaymentByProviderIDFn(ctx, providerPaymentID)
}
func (m *mockBillingRepo) UpdatePaymentStatus(ctx context.Context, id int, status string) error {
	return m.updatePaymentStatusFn(ctx, id, status)
}

func newTestUsecase(repo *mockBillingRepo) *Usecase {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	uc := New(repo)
	_ = uc.Init(context.Background(), logger)
	return uc
}

func TestGetTariffs(t *testing.T) {
	repo := &mockBillingRepo{
		getTariffsFn: func(_ context.Context) ([]entity.Tariff, error) {
			return []entity.Tariff{
				{ID: 1, Name: "Trial", Slug: "trial"},
				{ID: 2, Name: "Basic", Slug: "basic"},
			}, nil
		},
	}
	uc := newTestUsecase(repo)

	tariffs, err := uc.GetTariffs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tariffs) != 2 {
		t.Errorf("expected 2 tariffs, got %d", len(tariffs))
	}
}

func TestSubscribe_Success(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return nil, fmt.Errorf("billing.GetSubscriptionByOrgID: %w", sql.ErrNoRows)
		},
		getTariffBySlugFn: func(_ context.Context, slug string) (*entity.Tariff, error) {
			return &entity.Tariff{ID: 2, Name: "Basic", Slug: "basic", Price: 290000, Interval: "month"}, nil
		},
		createSubscriptionFn: func(_ context.Context, sub *entity.Subscription) error {
			sub.ID = 1
			sub.CreatedAt = time.Now()
			sub.UpdatedAt = time.Now()
			return nil
		},
		createInvoiceFn: func(_ context.Context, inv *entity.Invoice) error {
			inv.ID = 1
			return nil
		},
	}
	uc := newTestUsecase(repo)

	sub, err := uc.Subscribe(context.Background(), 1, "basic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.TariffID != 2 {
		t.Errorf("expected tariff_id=2, got %d", sub.TariffID)
	}
	if sub.Status != "active" {
		t.Errorf("expected status=active, got %s", sub.Status)
	}
}

func TestSubscribe_Trial(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return nil, fmt.Errorf("billing.GetSubscriptionByOrgID: %w", sql.ErrNoRows)
		},
		getTariffBySlugFn: func(_ context.Context, _ string) (*entity.Tariff, error) {
			return &entity.Tariff{ID: 1, Name: "Trial", Slug: "trial", Price: 0, Interval: "month"}, nil
		},
		createSubscriptionFn: func(_ context.Context, sub *entity.Subscription) error {
			sub.ID = 1
			return nil
		},
	}
	uc := newTestUsecase(repo)

	sub, err := uc.Subscribe(context.Background(), 1, "trial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != "trialing" {
		t.Errorf("expected status=trialing, got %s", sub.Status)
	}
}

func TestSubscribe_AlreadyActive(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return &entity.SubscriptionWithTariff{
				Subscription: entity.Subscription{ID: 1, Status: "active"},
			}, nil
		},
	}
	uc := newTestUsecase(repo)

	_, err := uc.Subscribe(context.Background(), 1, "basic")
	if err != ErrAlreadySubscribed {
		t.Errorf("expected ErrAlreadySubscribed, got %v", err)
	}
}

func TestSubscribe_TariffNotFound(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return nil, fmt.Errorf("billing.GetSubscriptionByOrgID: %w", sql.ErrNoRows)
		},
		getTariffBySlugFn: func(_ context.Context, _ string) (*entity.Tariff, error) {
			return nil, fmt.Errorf("billing.GetTariffBySlug: %w", sql.ErrNoRows)
		},
	}
	uc := newTestUsecase(repo)

	_, err := uc.Subscribe(context.Background(), 1, "nonexistent")
	if err != ErrTariffNotFound {
		t.Errorf("expected ErrTariffNotFound, got %v", err)
	}
}

func TestChangePlan(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return &entity.SubscriptionWithTariff{
				Subscription: entity.Subscription{ID: 1, TariffID: 1, Status: "active"},
			}, nil
		},
		getTariffBySlugFn: func(_ context.Context, _ string) (*entity.Tariff, error) {
			return &entity.Tariff{ID: 3, Name: "Pro", Slug: "pro", Price: 790000, Interval: "month"}, nil
		},
		updateSubscriptionTariffFn: func(_ context.Context, _ int, _ int, _ time.Time) error {
			return nil
		},
		createInvoiceFn: func(_ context.Context, inv *entity.Invoice) error {
			inv.ID = 2
			return nil
		},
	}
	uc := newTestUsecase(repo)

	sub, err := uc.ChangePlan(context.Background(), 1, "pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.TariffID != 3 {
		t.Errorf("expected tariff_id=3, got %d", sub.TariffID)
	}
}

func TestCancelSubscription(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return &entity.SubscriptionWithTariff{
				Subscription: entity.Subscription{ID: 1, Status: "active"},
			}, nil
		},
		updateSubscriptionStatusFn: func(_ context.Context, _ int, status string, canceledAt *time.Time) error {
			if status != "canceled" {
				t.Errorf("expected canceled, got %s", status)
			}
			if canceledAt == nil {
				t.Error("expected canceledAt to be set")
			}
			return nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.CancelSubscription(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCancelSubscription_NotFound(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return nil, fmt.Errorf("billing.GetSubscriptionByOrgID: %w", sql.ErrNoRows)
		},
	}
	uc := newTestUsecase(repo)

	err := uc.CancelSubscription(context.Background(), 1)
	if err != ErrSubscriptionNotFound {
		t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
	}
}

func TestHasFeature(t *testing.T) {
	tests := []struct {
		name     string
		feature  string
		features entity.TariffFeatures
		want     bool
	}{
		{"rfm enabled", "rfm", entity.TariffFeatures{RFM: true}, true},
		{"rfm disabled", "rfm", entity.TariffFeatures{RFM: false}, false},
		{"advanced_campaigns enabled", "advanced_campaigns", entity.TariffFeatures{AdvancedCampaigns: true}, true},
		{"unknown feature", "nonexistent", entity.TariffFeatures{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingRepo{
				getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
					return &entity.SubscriptionWithTariff{
						TariffFeatures: tt.features,
					}, nil
				},
			}
			uc := newTestUsecase(repo)

			got, err := uc.HasFeature(context.Background(), 1, tt.feature)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("HasFeature(%q) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}

func TestHasFeature_NoSubscription(t *testing.T) {
	repo := &mockBillingRepo{
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return nil, fmt.Errorf("billing.GetSubscriptionByOrgID: %w", sql.ErrNoRows)
		},
	}
	uc := newTestUsecase(repo)

	got, err := uc.HasFeature(context.Background(), 1, "rfm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != false {
		t.Error("expected false for no subscription")
	}
}

func TestCheckLimit(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		current int
		limits  entity.TariffLimits
		want    bool
	}{
		{"under limit", "max_bots", 1, entity.TariffLimits{MaxBots: 5}, true},
		{"at limit", "max_bots", 5, entity.TariffLimits{MaxBots: 5}, false},
		{"unlimited", "max_bots", 100, entity.TariffLimits{MaxBots: -1}, true},
		{"unknown key", "unknown", 0, entity.TariffLimits{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingRepo{
				getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
					return &entity.SubscriptionWithTariff{
						TariffLimits: tt.limits,
					}, nil
				},
			}
			uc := newTestUsecase(repo)

			got, err := uc.CheckLimit(context.Background(), 1, tt.key, tt.current)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("CheckLimit(%q, %d) = %v, want %v", tt.key, tt.current, got, tt.want)
			}
		})
	}
}

func TestProcessPayment_CreatesAsPending(t *testing.T) {
	var createdStatus string
	repo := &mockBillingRepo{
		getInvoiceByIDFn: func(_ context.Context, _ int) (*entity.Invoice, error) {
			return &entity.Invoice{ID: 1, OrgID: 1, Amount: 290000, Currency: "RUB", Status: "pending"}, nil
		},
		createPaymentFn: func(_ context.Context, p *entity.Payment) error {
			createdStatus = p.Status
			p.ID = 1
			return nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.ProcessPayment(context.Background(), 1, entity.ProcessPaymentRequest{
		InvoiceID: 1, Provider: "yukassa", Amount: 290000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdStatus != "pending" {
		t.Errorf("expected payment status=pending, got %s", createdStatus)
	}
}

func TestProcessPayment_AmountMismatch(t *testing.T) {
	repo := &mockBillingRepo{
		getInvoiceByIDFn: func(_ context.Context, _ int) (*entity.Invoice, error) {
			return &entity.Invoice{ID: 1, OrgID: 1, Amount: 290000, Status: "pending"}, nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.ProcessPayment(context.Background(), 1, entity.ProcessPaymentRequest{
		InvoiceID: 1, Provider: "yukassa", Amount: 100000,
	})
	if err != ErrAmountMismatch {
		t.Errorf("expected ErrAmountMismatch, got %v", err)
	}
}

func TestProcessPayment_AlreadyPaid(t *testing.T) {
	repo := &mockBillingRepo{
		getInvoiceByIDFn: func(_ context.Context, _ int) (*entity.Invoice, error) {
			return &entity.Invoice{ID: 1, OrgID: 1, Amount: 290000, Status: "paid"}, nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.ProcessPayment(context.Background(), 1, entity.ProcessPaymentRequest{
		InvoiceID: 1, Provider: "yukassa", Amount: 290000,
	})
	if err != ErrInvoiceAlreadyPaid {
		t.Errorf("expected ErrInvoiceAlreadyPaid, got %v", err)
	}
}

func TestProcessPayment_WrongOrg(t *testing.T) {
	repo := &mockBillingRepo{
		getInvoiceByIDFn: func(_ context.Context, _ int) (*entity.Invoice, error) {
			return &entity.Invoice{ID: 1, OrgID: 2, Status: "pending"}, nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.ProcessPayment(context.Background(), 1, entity.ProcessPaymentRequest{InvoiceID: 1, Provider: "yukassa", Amount: 290000})
	if err != ErrInvoiceNotFound {
		t.Errorf("expected ErrInvoiceNotFound, got %v", err)
	}
}

func TestConfirmPayment_Success(t *testing.T) {
	subID := 1
	repo := &mockBillingRepo{
		getPaymentByProviderIDFn: func(_ context.Context, _ string) (*entity.Payment, error) {
			return &entity.Payment{ID: 1, InvoiceID: 10, OrgID: 1, Status: "pending"}, nil
		},
		updatePaymentStatusFn: func(_ context.Context, _ int, status string) error {
			if status != "succeeded" {
				t.Errorf("expected succeeded, got %s", status)
			}
			return nil
		},
		updateInvoiceStatusFn: func(_ context.Context, _ int, status string, _ *time.Time) error {
			if status != "paid" {
				t.Errorf("expected paid, got %s", status)
			}
			return nil
		},
		getInvoiceByIDFn: func(_ context.Context, _ int) (*entity.Invoice, error) {
			return &entity.Invoice{ID: 10, OrgID: 1, SubscriptionID: &subID}, nil
		},
		getSubscriptionByOrgIDFn: func(_ context.Context, _ int) (*entity.SubscriptionWithTariff, error) {
			return &entity.SubscriptionWithTariff{
				Subscription: entity.Subscription{ID: 1, Status: "past_due"},
			}, nil
		},
		updateSubscriptionStatusFn: func(_ context.Context, _ int, status string, _ *time.Time) error {
			if status != "active" {
				t.Errorf("expected active, got %s", status)
			}
			return nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.ConfirmPayment(context.Background(), "provider_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfirmPayment_NotFound(t *testing.T) {
	repo := &mockBillingRepo{
		getPaymentByProviderIDFn: func(_ context.Context, _ string) (*entity.Payment, error) {
			return nil, nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.ConfirmPayment(context.Background(), "nonexistent")
	if err != ErrPaymentNotFound {
		t.Errorf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestConfirmPayment_Idempotent(t *testing.T) {
	repo := &mockBillingRepo{
		getPaymentByProviderIDFn: func(_ context.Context, _ string) (*entity.Payment, error) {
			return &entity.Payment{ID: 1, Status: "succeeded"}, nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.ConfirmPayment(context.Background(), "already_done")
	if err != nil {
		t.Fatalf("expected nil (idempotent), got %v", err)
	}
}

func TestHandleExpiredSubscriptions(t *testing.T) {
	var updatedStatuses []string
	repo := &mockBillingRepo{
		getExpiredSubscriptionsFn: func(_ context.Context) ([]entity.Subscription, error) {
			return []entity.Subscription{
				{ID: 1, Status: "active"},
				{ID: 2, Status: "past_due"},
			}, nil
		},
		updateSubscriptionStatusFn: func(_ context.Context, _ int, status string, _ *time.Time) error {
			updatedStatuses = append(updatedStatuses, status)
			return nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.HandleExpiredSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updatedStatuses) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(updatedStatuses))
	}
	if updatedStatuses[0] != "past_due" {
		t.Errorf("active should become past_due, got %s", updatedStatuses[0])
	}
	if updatedStatuses[1] != "expired" {
		t.Errorf("past_due should become expired, got %s", updatedStatuses[1])
	}
}
