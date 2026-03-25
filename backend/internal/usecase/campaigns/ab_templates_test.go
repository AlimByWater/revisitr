package campaigns_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"revisitr/internal/entity"
	"revisitr/internal/usecase/campaigns"
)

// ── Variant mock ─────────────────────────────────────────────────────────────

type mockVariantsRepo struct {
	createVariantFn             func(ctx context.Context, v *entity.CampaignVariant) error
	getVariantsByCampaignIDFn   func(ctx context.Context, campaignID int) ([]entity.CampaignVariant, error)
	updateVariantStatsFn        func(ctx context.Context, id int, stats *entity.CampaignStats) error
	setVariantWinnerFn          func(ctx context.Context, campaignID, variantID int) error
	deleteVariantsByCampaignFn  func(ctx context.Context, campaignID int) error
	getVariantAnalyticsFn       func(ctx context.Context, variantID int) (*entity.VariantResult, error)
}

func (m *mockVariantsRepo) CreateVariant(ctx context.Context, v *entity.CampaignVariant) error {
	return m.createVariantFn(ctx, v)
}
func (m *mockVariantsRepo) GetVariantsByCampaignID(ctx context.Context, campaignID int) ([]entity.CampaignVariant, error) {
	return m.getVariantsByCampaignIDFn(ctx, campaignID)
}
func (m *mockVariantsRepo) UpdateVariantStats(ctx context.Context, id int, stats *entity.CampaignStats) error {
	return m.updateVariantStatsFn(ctx, id, stats)
}
func (m *mockVariantsRepo) SetVariantWinner(ctx context.Context, campaignID, variantID int) error {
	return m.setVariantWinnerFn(ctx, campaignID, variantID)
}
func (m *mockVariantsRepo) DeleteVariantsByCampaignID(ctx context.Context, campaignID int) error {
	return m.deleteVariantsByCampaignFn(ctx, campaignID)
}
func (m *mockVariantsRepo) GetVariantAnalytics(ctx context.Context, variantID int) (*entity.VariantResult, error) {
	return m.getVariantAnalyticsFn(ctx, variantID)
}

// ── Templates mock ───────────────────────────────────────────────────────────

type mockTemplatesRepo struct {
	createTemplateFn   func(ctx context.Context, t *entity.CampaignTemplate) error
	getTemplatesFn     func(ctx context.Context, orgID int) ([]entity.CampaignTemplate, error)
	getTemplateByIDFn  func(ctx context.Context, id int) (*entity.CampaignTemplate, error)
	updateTemplateFn   func(ctx context.Context, t *entity.CampaignTemplate) error
	deleteTemplateFn   func(ctx context.Context, id int) error
}

func (m *mockTemplatesRepo) CreateTemplate(ctx context.Context, t *entity.CampaignTemplate) error {
	return m.createTemplateFn(ctx, t)
}
func (m *mockTemplatesRepo) GetTemplates(ctx context.Context, orgID int) ([]entity.CampaignTemplate, error) {
	return m.getTemplatesFn(ctx, orgID)
}
func (m *mockTemplatesRepo) GetTemplateByID(ctx context.Context, id int) (*entity.CampaignTemplate, error) {
	return m.getTemplateByIDFn(ctx, id)
}
func (m *mockTemplatesRepo) UpdateTemplate(ctx context.Context, t *entity.CampaignTemplate) error {
	return m.updateTemplateFn(ctx, t)
}
func (m *mockTemplatesRepo) DeleteTemplate(ctx context.Context, id int) error {
	return m.deleteTemplateFn(ctx, id)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func draftCampaignRepo(orgID int) *mockCampaignsRepo {
	return &mockCampaignsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Campaign, error) {
			return &entity.Campaign{ID: 1, OrgID: orgID, Status: "draft"}, nil
		},
		createFn: func(_ context.Context, c *entity.Campaign) error {
			c.ID = 99
			return nil
		},
	}
}

func newABUC(cr *mockCampaignsRepo, vr *mockVariantsRepo) *campaigns.Usecase {
	return campaigns.New(cr, &mockScenariosRepo{}, &mockClientsRepo{},
		campaigns.WithVariants(vr),
	)
}

func newTemplateUC(cr *mockCampaignsRepo, tr *mockTemplatesRepo) *campaigns.Usecase {
	return campaigns.New(cr, &mockScenariosRepo{}, &mockClientsRepo{},
		campaigns.WithTemplates(tr),
	)
}

// ── A/B Tests ────────────────────────────────────────────────────────────────

func TestCreateABTest_Success(t *testing.T) {
	cr := draftCampaignRepo(1)
	nextID := 0
	vr := &mockVariantsRepo{
		deleteVariantsByCampaignFn: func(_ context.Context, _ int) error { return nil },
		createVariantFn: func(_ context.Context, v *entity.CampaignVariant) error {
			nextID++
			v.ID = nextID
			return nil
		},
	}
	uc := newABUC(cr, vr)

	req := &entity.CreateABTestRequest{
		Variants: []entity.CreateVariantRequest{
			{Name: "A", AudiencePct: 50, Message: "Hello A"},
			{Name: "B", AudiencePct: 50, Message: "Hello B"},
		},
	}

	variants, err := uc.CreateABTest(context.Background(), 1, 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(variants))
	}
	if variants[0].Name != "A" || variants[1].Name != "B" {
		t.Errorf("unexpected variant names: %q, %q", variants[0].Name, variants[1].Name)
	}
}

func TestCreateABTest_InvalidPercentages(t *testing.T) {
	cr := draftCampaignRepo(1)
	vr := &mockVariantsRepo{}
	uc := newABUC(cr, vr)

	req := &entity.CreateABTestRequest{
		Variants: []entity.CreateVariantRequest{
			{Name: "A", AudiencePct: 30, Message: "Hello A"},
			{Name: "B", AudiencePct: 30, Message: "Hello B"},
		},
	}

	_, err := uc.CreateABTest(context.Background(), 1, 1, req)
	if !errors.Is(err, campaigns.ErrInvalidVariantPct) {
		t.Errorf("expected ErrInvalidVariantPct, got %v", err)
	}
}

func TestCreateABTest_NotDraft(t *testing.T) {
	cr := &mockCampaignsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Campaign, error) {
			return &entity.Campaign{ID: 1, OrgID: 1, Status: "sent"}, nil
		},
	}
	vr := &mockVariantsRepo{}
	uc := newABUC(cr, vr)

	req := &entity.CreateABTestRequest{
		Variants: []entity.CreateVariantRequest{
			{Name: "A", AudiencePct: 50, Message: "A"},
			{Name: "B", AudiencePct: 50, Message: "B"},
		},
	}

	_, err := uc.CreateABTest(context.Background(), 1, 1, req)
	if !errors.Is(err, campaigns.ErrCampaignNotDraft) {
		t.Errorf("expected ErrCampaignNotDraft, got %v", err)
	}
}

func TestCreateABTest_WrongOrg(t *testing.T) {
	cr := draftCampaignRepo(5) // org 5
	vr := &mockVariantsRepo{}
	uc := newABUC(cr, vr)

	req := &entity.CreateABTestRequest{
		Variants: []entity.CreateVariantRequest{
			{Name: "A", AudiencePct: 50, Message: "A"},
			{Name: "B", AudiencePct: 50, Message: "B"},
		},
	}

	_, err := uc.CreateABTest(context.Background(), 99, 1, req) // org 99 ≠ 5
	if !errors.Is(err, campaigns.ErrNotCampaignOwner) {
		t.Errorf("expected ErrNotCampaignOwner, got %v", err)
	}
}

func TestGetABResults_Success(t *testing.T) {
	cr := draftCampaignRepo(1)
	vr := &mockVariantsRepo{
		getVariantsByCampaignIDFn: func(_ context.Context, _ int) ([]entity.CampaignVariant, error) {
			return []entity.CampaignVariant{
				{ID: 1, Name: "A", IsWinner: true, Stats: entity.CampaignStats{Total: 10, Sent: 8}},
				{ID: 2, Name: "B", IsWinner: false, Stats: entity.CampaignStats{Total: 10, Sent: 9}},
			}, nil
		},
		getVariantAnalyticsFn: func(_ context.Context, variantID int) (*entity.VariantResult, error) {
			return &entity.VariantResult{
				ID:       variantID,
				Name:     fmt.Sprintf("Variant %d", variantID),
				Clicked:  3,
				ClickRate: 0.3,
			}, nil
		},
	}
	uc := newABUC(cr, vr)

	results, err := uc.GetABResults(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(results.Variants))
	}
	if results.WinnerID == nil || *results.WinnerID != 1 {
		t.Errorf("expected winner ID 1, got %v", results.WinnerID)
	}
}

func TestPickWinner_Success(t *testing.T) {
	cr := draftCampaignRepo(1)
	setWinnerCalled := false
	vr := &mockVariantsRepo{
		setVariantWinnerFn: func(_ context.Context, campaignID, variantID int) error {
			if campaignID != 1 || variantID != 2 {
				t.Errorf("expected (1, 2), got (%d, %d)", campaignID, variantID)
			}
			setWinnerCalled = true
			return nil
		},
	}
	uc := newABUC(cr, vr)

	err := uc.PickWinner(context.Background(), 1, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !setWinnerCalled {
		t.Error("SetVariantWinner was not called")
	}
}

func TestPickWinner_CampaignNotFound(t *testing.T) {
	cr := &mockCampaignsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Campaign, error) {
			return nil, errors.New("not found")
		},
	}
	vr := &mockVariantsRepo{}
	uc := newABUC(cr, vr)

	err := uc.PickWinner(context.Background(), 1, 99, 1)
	if !errors.Is(err, campaigns.ErrCampaignNotFound) {
		t.Errorf("expected ErrCampaignNotFound, got %v", err)
	}
}

// ── Template Tests ───────────────────────────────────────────────────────────

func TestCreateCampaignTemplate_Success(t *testing.T) {
	cr := &mockCampaignsRepo{}
	tr := &mockTemplatesRepo{
		createTemplateFn: func(_ context.Context, tmpl *entity.CampaignTemplate) error {
			tmpl.ID = 1
			return nil
		},
	}
	uc := newTemplateUC(cr, tr)

	req := &entity.CreateCampaignTemplateRequest{
		Name:    "My Template",
		Message: "Hello {name}",
	}
	tmpl, err := uc.CreateCampaignTemplate(context.Background(), 10, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.ID != 1 {
		t.Errorf("expected ID=1, got %d", tmpl.ID)
	}
	if tmpl.OrgID == nil || *tmpl.OrgID != 10 {
		t.Errorf("expected OrgID=10, got %v", tmpl.OrgID)
	}
	if tmpl.Category != "general" {
		t.Errorf("expected default category=general, got %s", tmpl.Category)
	}
}

func TestListCampaignTemplates_Success(t *testing.T) {
	cr := &mockCampaignsRepo{}
	tr := &mockTemplatesRepo{
		getTemplatesFn: func(_ context.Context, orgID int) ([]entity.CampaignTemplate, error) {
			if orgID != 10 {
				t.Errorf("expected orgID=10, got %d", orgID)
			}
			return []entity.CampaignTemplate{
				{ID: 1, Name: "System", IsSystem: true},
				{ID: 2, Name: "Custom"},
			}, nil
		},
	}
	uc := newTemplateUC(cr, tr)

	templates, err := uc.ListCampaignTemplates(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}
}

func TestUpdateCampaignTemplate_SystemDenied(t *testing.T) {
	cr := &mockCampaignsRepo{}
	tr := &mockTemplatesRepo{
		getTemplateByIDFn: func(_ context.Context, _ int) (*entity.CampaignTemplate, error) {
			return &entity.CampaignTemplate{ID: 1, IsSystem: true}, nil
		},
	}
	uc := newTemplateUC(cr, tr)

	err := uc.UpdateCampaignTemplate(context.Background(), 10, 1, &entity.UpdateCampaignTemplateRequest{})
	if !errors.Is(err, campaigns.ErrTemplateIsSystem) {
		t.Errorf("expected ErrTemplateIsSystem, got %v", err)
	}
}

func TestUpdateCampaignTemplate_WrongOrg(t *testing.T) {
	orgID := 5
	cr := &mockCampaignsRepo{}
	tr := &mockTemplatesRepo{
		getTemplateByIDFn: func(_ context.Context, _ int) (*entity.CampaignTemplate, error) {
			return &entity.CampaignTemplate{ID: 1, OrgID: &orgID}, nil
		},
	}
	uc := newTemplateUC(cr, tr)

	err := uc.UpdateCampaignTemplate(context.Background(), 99, 1, &entity.UpdateCampaignTemplateRequest{})
	if !errors.Is(err, campaigns.ErrNotTemplateOwner) {
		t.Errorf("expected ErrNotTemplateOwner, got %v", err)
	}
}

func TestDeleteCampaignTemplate_Success(t *testing.T) {
	orgID := 10
	deleted := false
	cr := &mockCampaignsRepo{}
	tr := &mockTemplatesRepo{
		getTemplateByIDFn: func(_ context.Context, _ int) (*entity.CampaignTemplate, error) {
			return &entity.CampaignTemplate{ID: 1, OrgID: &orgID}, nil
		},
		deleteTemplateFn: func(_ context.Context, id int) error {
			if id != 1 {
				t.Errorf("expected id=1, got %d", id)
			}
			deleted = true
			return nil
		},
	}
	uc := newTemplateUC(cr, tr)

	err := uc.DeleteCampaignTemplate(context.Background(), 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("DeleteTemplate was not called")
	}
}

func TestDeleteCampaignTemplate_NotFound(t *testing.T) {
	cr := &mockCampaignsRepo{}
	tr := &mockTemplatesRepo{
		getTemplateByIDFn: func(_ context.Context, _ int) (*entity.CampaignTemplate, error) {
			return nil, fmt.Errorf("not found: %w", sql.ErrNoRows)
		},
	}
	uc := newTemplateUC(cr, tr)

	err := uc.DeleteCampaignTemplate(context.Background(), 10, 99)
	if !errors.Is(err, campaigns.ErrTemplateNotFound) {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestCreateCampaignFromTemplate_Success(t *testing.T) {
	orgID := 10
	cr := draftCampaignRepo(orgID) // cr.createFn assigns ID=99
	tr := &mockTemplatesRepo{
		getTemplateByIDFn: func(_ context.Context, _ int) (*entity.CampaignTemplate, error) {
			return &entity.CampaignTemplate{
				ID:       1,
				OrgID:    &orgID,
				Name:     "Welcome",
				Message:  "Hello!",
				Category: "welcome",
			}, nil
		},
	}
	uc := newTemplateUC(cr, tr)

	campaign, err := uc.CreateCampaignFromTemplate(context.Background(), 10, 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if campaign.ID != 99 {
		t.Errorf("expected ID=99, got %d", campaign.ID)
	}
	if campaign.Name != "Welcome" {
		t.Errorf("expected name=Welcome, got %s", campaign.Name)
	}
	if campaign.BotID != 5 {
		t.Errorf("expected botID=5, got %d", campaign.BotID)
	}
	if campaign.Status != "draft" {
		t.Errorf("expected status=draft, got %s", campaign.Status)
	}
}

func TestCreateCampaignFromTemplate_SystemTemplate(t *testing.T) {
	cr := draftCampaignRepo(10)
	tr := &mockTemplatesRepo{
		getTemplateByIDFn: func(_ context.Context, _ int) (*entity.CampaignTemplate, error) {
			return &entity.CampaignTemplate{
				ID:       1,
				OrgID:    nil, // system template
				Name:     "System Welcome",
				Message:  "Hello from system!",
				IsSystem: true,
			}, nil
		},
	}
	uc := newTemplateUC(cr, tr)

	campaign, err := uc.CreateCampaignFromTemplate(context.Background(), 10, 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if campaign.Name != "System Welcome" {
		t.Errorf("expected name from system template, got %s", campaign.Name)
	}
}

func TestCreateCampaignFromTemplate_WrongOrg(t *testing.T) {
	otherOrg := 5
	cr := &mockCampaignsRepo{}
	tr := &mockTemplatesRepo{
		getTemplateByIDFn: func(_ context.Context, _ int) (*entity.CampaignTemplate, error) {
			return &entity.CampaignTemplate{ID: 1, OrgID: &otherOrg}, nil
		},
	}
	uc := newTemplateUC(cr, tr)

	_, err := uc.CreateCampaignFromTemplate(context.Background(), 99, 1, 3)
	if !errors.Is(err, campaigns.ErrNotTemplateOwner) {
		t.Errorf("expected ErrNotTemplateOwner, got %v", err)
	}
}
