package adminbot

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"testing"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
)

// ── Mocks ────────────────────────────────────────────────────────────────────

type mockLinksRepo struct {
	getByTelegramIDFn func(ctx context.Context, telegramID int64) (*entity.AdminBotLink, error)
	getByLinkCodeFn   func(ctx context.Context, code string) (*entity.AdminBotLink, error)
	activateLinkFn    func(ctx context.Context, id int, telegramID int64) error
}

func (m *mockLinksRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.AdminBotLink, error) {
	return m.getByTelegramIDFn(ctx, telegramID)
}
func (m *mockLinksRepo) GetByLinkCode(ctx context.Context, code string) (*entity.AdminBotLink, error) {
	return m.getByLinkCodeFn(ctx, code)
}
func (m *mockLinksRepo) ActivateLink(ctx context.Context, id int, telegramID int64) error {
	return m.activateLinkFn(ctx, id, telegramID)
}

type mockDashboardRepo struct {
	getWidgetsFn func(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error)
}

func (m *mockDashboardRepo) GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error) {
	return m.getWidgetsFn(ctx, orgID, filter)
}

type mockCampaignsRepo struct {
	getByOrgIDFn func(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error)
}

func (m *mockCampaignsRepo) GetByOrgID(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error) {
	return m.getByOrgIDFn(ctx, orgID, limit, offset)
}

type mockPromotionsRepo struct {
	createPromoCodeFn func(ctx context.Context, pc *entity.PromoCode) error
	getByOrgIDFn      func(ctx context.Context, orgID int) ([]entity.Promotion, error)
}

func (m *mockPromotionsRepo) CreatePromoCode(ctx context.Context, pc *entity.PromoCode) error {
	return m.createPromoCodeFn(ctx, pc)
}
func (m *mockPromotionsRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Promotion, error) {
	return m.getByOrgIDFn(ctx, orgID)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestHandler(
	links *mockLinksRepo,
	dashboard *mockDashboardRepo,
	campaigns *mockCampaignsRepo,
	promotions *mockPromotionsRepo,
) *handler {
	return &handler{
		bot:            nil, // bot is nil; sendText/sendWithKeyboard will panic but we recover
		linksRepo:      links,
		dashboardRepo:  dashboard,
		campaignsRepo:  campaigns,
		promotionsRepo: promotions,
		logger:         testLogger(),
	}
}

func linkedUser() *mockLinksRepo {
	return &mockLinksRepo{
		getByTelegramIDFn: func(_ context.Context, _ int64) (*entity.AdminBotLink, error) {
			return &entity.AdminBotLink{ID: 1, UserID: 10, OrgID: 100, Role: "owner"}, nil
		},
	}
}

func unlinkedUser() *mockLinksRepo {
	return &mockLinksRepo{
		getByTelegramIDFn: func(_ context.Context, _ int64) (*entity.AdminBotLink, error) {
			return nil, sql.ErrNoRows
		},
	}
}

func makeTelegoMsg(fromID, chatID int64) *telego.Message {
	return &telego.Message{
		From: &telego.User{ID: fromID},
		Chat: telego.Chat{ID: chatID},
	}
}

// ── Tests: getLink ───────────────────────────────────────────────────────────

func TestGetLink_Found(t *testing.T) {
	links := linkedUser()
	h := newTestHandler(links, nil, nil, nil)

	link := h.getLink(context.Background(), 12345)
	if link == nil {
		t.Fatal("expected link, got nil")
	}
	if link.OrgID != 100 {
		t.Errorf("expected OrgID 100, got %d", link.OrgID)
	}
}

func TestGetLink_NotFound(t *testing.T) {
	links := unlinkedUser()
	h := newTestHandler(links, nil, nil, nil)

	link := h.getLink(context.Background(), 99999)
	if link != nil {
		t.Fatalf("expected nil, got %+v", link)
	}
}

func TestGetLink_Error(t *testing.T) {
	links := &mockLinksRepo{
		getByTelegramIDFn: func(_ context.Context, _ int64) (*entity.AdminBotLink, error) {
			return nil, errors.New("db connection error")
		},
	}
	h := newTestHandler(links, nil, nil, nil)

	link := h.getLink(context.Background(), 12345)
	if link != nil {
		t.Fatalf("expected nil on error, got %+v", link)
	}
}

// ── Tests: requireAuth ───────────────────────────────────────────────────────

func TestRequireAuth_Linked(t *testing.T) {
	links := linkedUser()
	h := newTestHandler(links, nil, nil, nil)

	// For linked user, requireAuth returns the link without sending a message.
	msg := makeTelegoMsg(12345, 100)
	link := h.requireAuth(context.Background(), msg)
	if link == nil {
		t.Fatal("expected link for linked user")
	}
	if link.Role != "owner" {
		t.Errorf("expected role owner, got %s", link.Role)
	}
}

// ── Tests: handleLink ────────────────────────────────────────────────────────

func TestHandleLink_EmptyCode(t *testing.T) {
	links := unlinkedUser()
	h := newTestHandler(links, nil, nil, nil)

	// Empty code — handler tries to send error text; bot is nil so it panics in sendText.
	// We don't test sendText behavior, just that the code path is correct.
	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }() // recover from nil bot panic
		h.handleLink(context.Background(), msg, "")
	}()
}

func TestHandleLink_AlreadyLinked(t *testing.T) {
	links := linkedUser()
	h := newTestHandler(links, nil, nil, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleLink(context.Background(), msg, "ABCDEF")
	}()
}

func TestHandleLink_InvalidCode(t *testing.T) {
	links := unlinkedUser()
	links.getByLinkCodeFn = func(_ context.Context, _ string) (*entity.AdminBotLink, error) {
		return nil, sql.ErrNoRows
	}
	h := newTestHandler(links, nil, nil, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleLink(context.Background(), msg, "BADCODE")
	}()
}

func TestHandleLink_ActivateSuccess(t *testing.T) {
	activated := false
	links := unlinkedUser()
	links.getByLinkCodeFn = func(_ context.Context, code string) (*entity.AdminBotLink, error) {
		if code != "GOODCODE" {
			t.Errorf("expected code GOODCODE, got %s", code)
		}
		return &entity.AdminBotLink{ID: 5, UserID: 10, OrgID: 100, Role: "manager"}, nil
	}
	links.activateLinkFn = func(_ context.Context, id int, telegramID int64) error {
		if id != 5 {
			t.Errorf("expected link ID 5, got %d", id)
		}
		if telegramID != 12345 {
			t.Errorf("expected telegram ID 12345, got %d", telegramID)
		}
		activated = true
		return nil
	}
	h := newTestHandler(links, nil, nil, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }() // sendWithKeyboard will panic with nil bot
		h.handleLink(context.Background(), msg, "GOODCODE")
	}()

	if !activated {
		t.Error("ActivateLink was not called")
	}
}

func TestHandleLink_ActivateError(t *testing.T) {
	links := unlinkedUser()
	links.getByLinkCodeFn = func(_ context.Context, _ string) (*entity.AdminBotLink, error) {
		return &entity.AdminBotLink{ID: 5, UserID: 10, OrgID: 100}, nil
	}
	links.activateLinkFn = func(_ context.Context, _ int, _ int64) error {
		return errors.New("db error")
	}
	h := newTestHandler(links, nil, nil, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleLink(context.Background(), msg, "CODE123")
	}()
}

// ── Tests: handleStats ───────────────────────────────────────────────────────

func TestHandleStats_Success(t *testing.T) {
	calledWidgets := false
	links := linkedUser()
	dashboard := &mockDashboardRepo{
		getWidgetsFn: func(_ context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error) {
			if orgID != 100 {
				t.Errorf("expected orgID 100, got %d", orgID)
			}
			if filter.Period != "7d" {
				t.Errorf("expected period 7d, got %s", filter.Period)
			}
			calledWidgets = true
			return &entity.DashboardWidgets{
				Revenue:       entity.DashboardMetric{Value: 50000, Trend: 5.2},
				AvgCheck:      entity.DashboardMetric{Value: 1200, Trend: -1.3},
				NewClients:    entity.DashboardMetric{Value: 15, Trend: 0},
				ActiveClients: entity.DashboardMetric{Value: 42, Trend: 3.1},
			}, nil
		},
	}
	h := newTestHandler(links, dashboard, nil, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleStats(context.Background(), msg)
	}()

	if !calledWidgets {
		t.Error("GetWidgets was not called")
	}
}

func TestHandleStats_Error(t *testing.T) {
	links := linkedUser()
	dashboard := &mockDashboardRepo{
		getWidgetsFn: func(_ context.Context, _ int, _ entity.DashboardFilter) (*entity.DashboardWidgets, error) {
			return nil, errors.New("db error")
		},
	}
	h := newTestHandler(links, dashboard, nil, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleStats(context.Background(), msg)
	}()
}

// ── Tests: handleCampaigns ───────────────────────────────────────────────────

func TestHandleCampaigns_WithResults(t *testing.T) {
	calledCampaigns := false
	links := linkedUser()
	campaigns := &mockCampaignsRepo{
		getByOrgIDFn: func(_ context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error) {
			if orgID != 100 {
				t.Errorf("expected orgID 100, got %d", orgID)
			}
			if limit != 5 {
				t.Errorf("expected limit 5, got %d", limit)
			}
			calledCampaigns = true
			return []entity.Campaign{
				{Name: "Welcome", Status: "sent"},
				{Name: "Promo", Status: "draft"},
			}, 2, nil
		},
	}
	h := newTestHandler(links, nil, campaigns, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleCampaigns(context.Background(), msg)
	}()

	if !calledCampaigns {
		t.Error("GetByOrgID was not called")
	}
}

func TestHandleCampaigns_Empty(t *testing.T) {
	links := linkedUser()
	campaigns := &mockCampaignsRepo{
		getByOrgIDFn: func(_ context.Context, _, _, _ int) ([]entity.Campaign, int, error) {
			return []entity.Campaign{}, 0, nil
		},
	}
	h := newTestHandler(links, nil, campaigns, nil)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleCampaigns(context.Background(), msg)
	}()
}

// ── Tests: handlePromotions ──────────────────────────────────────────────────

func TestHandlePromotions_ActiveExists(t *testing.T) {
	calledPromos := false
	links := linkedUser()
	promos := &mockPromotionsRepo{
		getByOrgIDFn: func(_ context.Context, orgID int) ([]entity.Promotion, error) {
			calledPromos = true
			return []entity.Promotion{
				{Name: "Happy Hour", Type: "discount", Active: true},
				{Name: "Old Promo", Type: "bonus", Active: false},
			}, nil
		},
	}
	h := newTestHandler(links, nil, nil, promos)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handlePromotions(context.Background(), msg)
	}()

	if !calledPromos {
		t.Error("GetByOrgID was not called")
	}
}

func TestHandlePromotions_NoneActive(t *testing.T) {
	links := linkedUser()
	promos := &mockPromotionsRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.Promotion, error) {
			return []entity.Promotion{
				{Name: "Expired", Type: "discount", Active: false},
			}, nil
		},
	}
	h := newTestHandler(links, nil, nil, promos)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handlePromotions(context.Background(), msg)
	}()
}

func TestHandlePromotions_Empty(t *testing.T) {
	links := linkedUser()
	promos := &mockPromotionsRepo{
		getByOrgIDFn: func(_ context.Context, _ int) ([]entity.Promotion, error) {
			return []entity.Promotion{}, nil
		},
	}
	h := newTestHandler(links, nil, nil, promos)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handlePromotions(context.Background(), msg)
	}()
}

// ── Tests: handleCreatePromo ─────────────────────────────────────────────────

func TestHandleCreatePromo_OwnerSuccess(t *testing.T) {
	created := false
	links := linkedUser() // role: owner
	promos := &mockPromotionsRepo{
		createPromoCodeFn: func(_ context.Context, pc *entity.PromoCode) error {
			if pc.OrgID != 100 {
				t.Errorf("expected orgID 100, got %d", pc.OrgID)
			}
			if pc.Code != "TESTCODE" {
				t.Errorf("expected code TESTCODE, got %s", pc.Code)
			}
			if !pc.Active {
				t.Error("expected active promo code")
			}
			created = true
			return nil
		},
	}
	h := newTestHandler(links, nil, nil, promos)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleCreatePromo(context.Background(), msg, "TESTCODE")
	}()

	if !created {
		t.Error("CreatePromoCode was not called")
	}
}

func TestHandleCreatePromo_GeneratesCode(t *testing.T) {
	var capturedCode string
	links := linkedUser()
	promos := &mockPromotionsRepo{
		createPromoCodeFn: func(_ context.Context, pc *entity.PromoCode) error {
			capturedCode = pc.Code
			return nil
		},
	}
	h := newTestHandler(links, nil, nil, promos)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleCreatePromo(context.Background(), msg, "")
	}()

	if capturedCode == "" {
		t.Error("expected auto-generated code, got empty")
	}
	if len(capturedCode) != 8 {
		t.Errorf("expected 8-char hex code, got %q (len %d)", capturedCode, len(capturedCode))
	}
}

func TestHandleCreatePromo_NonOwnerDenied(t *testing.T) {
	links := &mockLinksRepo{
		getByTelegramIDFn: func(_ context.Context, _ int64) (*entity.AdminBotLink, error) {
			return &entity.AdminBotLink{ID: 1, UserID: 10, OrgID: 100, Role: "manager"}, nil
		},
	}
	promos := &mockPromotionsRepo{
		createPromoCodeFn: func(_ context.Context, _ *entity.PromoCode) error {
			t.Error("CreatePromoCode should not be called for non-owner")
			return nil
		},
	}
	h := newTestHandler(links, nil, nil, promos)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleCreatePromo(context.Background(), msg, "CODE")
	}()
}

func TestHandleCreatePromo_RepoError(t *testing.T) {
	links := linkedUser()
	promos := &mockPromotionsRepo{
		createPromoCodeFn: func(_ context.Context, _ *entity.PromoCode) error {
			return errors.New("unique violation")
		},
	}
	h := newTestHandler(links, nil, nil, promos)

	msg := makeTelegoMsg(12345, 100)
	func() {
		defer func() { recover() }()
		h.handleCreatePromo(context.Background(), msg, "DUP")
	}()
}

// ── Tests: keyboard ──────────────────────────────────────────────────────────

func TestBuildAdminMenu(t *testing.T) {
	menu := buildAdminMenu()
	if menu == nil {
		t.Fatal("expected non-nil menu")
	}
	if len(menu.Keyboard) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(menu.Keyboard))
	}
	if len(menu.Keyboard[0]) != 2 {
		t.Fatalf("expected 2 buttons in row 1, got %d", len(menu.Keyboard[0]))
	}
	if menu.Keyboard[0][0].Text != btnStats {
		t.Errorf("expected first button %q, got %q", btnStats, menu.Keyboard[0][0].Text)
	}
	if !menu.ResizeKeyboard {
		t.Error("expected ResizeKeyboard true")
	}
}
