package campaigns_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"revisitr/internal/entity"
	"revisitr/internal/usecase/campaigns"
)

// --- mocks ---

type mockCampaignsRepo struct {
	createFn               func(ctx context.Context, c *entity.Campaign) error
	getByIDFn              func(ctx context.Context, id int) (*entity.Campaign, error)
	getByOrgIDFn           func(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error)
	updateFn               func(ctx context.Context, c *entity.Campaign) error
	updateStatusFn         func(ctx context.Context, id int, status string) error
	updateStatsFn          func(ctx context.Context, id int, stats *entity.CampaignStats) error
	deleteFn               func(ctx context.Context, id int) error
	createMessagesBatchFn  func(ctx context.Context, msgs []entity.CampaignMessage) error
	updateMessageStatusFn  func(ctx context.Context, id int, status string, errMsg *string) error
	getMessagesByCampaign  func(ctx context.Context, campaignID, limit, offset int) ([]entity.CampaignMessage, int, error)
	getMessageStatsFn      func(ctx context.Context, campaignID int) (*entity.CampaignStats, error)
	createClickFn          func(ctx context.Context, click *entity.CampaignClick) error
	getClicksByCampaignFn  func(ctx context.Context, campaignID int) ([]entity.CampaignClick, error)
	getCampaignAnalyticsFn func(ctx context.Context, campaignID int) (*entity.CampaignAnalyticsDetail, error)
	getScheduledCampaignsFn func(ctx context.Context, before time.Time) ([]entity.Campaign, error)
}

func (m *mockCampaignsRepo) Create(ctx context.Context, c *entity.Campaign) error {
	return m.createFn(ctx, c)
}
func (m *mockCampaignsRepo) GetByID(ctx context.Context, id int) (*entity.Campaign, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockCampaignsRepo) GetByOrgID(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error) {
	return m.getByOrgIDFn(ctx, orgID, limit, offset)
}
func (m *mockCampaignsRepo) Update(ctx context.Context, c *entity.Campaign) error {
	return m.updateFn(ctx, c)
}
func (m *mockCampaignsRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	return m.updateStatusFn(ctx, id, status)
}
func (m *mockCampaignsRepo) UpdateStats(ctx context.Context, id int, stats *entity.CampaignStats) error {
	return m.updateStatsFn(ctx, id, stats)
}
func (m *mockCampaignsRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}
func (m *mockCampaignsRepo) CreateMessagesBatch(ctx context.Context, msgs []entity.CampaignMessage) error {
	return m.createMessagesBatchFn(ctx, msgs)
}
func (m *mockCampaignsRepo) UpdateMessageStatus(ctx context.Context, id int, status string, errMsg *string) error {
	return m.updateMessageStatusFn(ctx, id, status, errMsg)
}
func (m *mockCampaignsRepo) GetMessagesByCampaignID(ctx context.Context, campaignID, limit, offset int) ([]entity.CampaignMessage, int, error) {
	return m.getMessagesByCampaign(ctx, campaignID, limit, offset)
}
func (m *mockCampaignsRepo) GetMessageStats(ctx context.Context, campaignID int) (*entity.CampaignStats, error) {
	return m.getMessageStatsFn(ctx, campaignID)
}
func (m *mockCampaignsRepo) CreateClick(ctx context.Context, click *entity.CampaignClick) error {
	return m.createClickFn(ctx, click)
}
func (m *mockCampaignsRepo) GetClicksByCampaign(ctx context.Context, campaignID int) ([]entity.CampaignClick, error) {
	return m.getClicksByCampaignFn(ctx, campaignID)
}
func (m *mockCampaignsRepo) GetCampaignAnalytics(ctx context.Context, campaignID int) (*entity.CampaignAnalyticsDetail, error) {
	return m.getCampaignAnalyticsFn(ctx, campaignID)
}
func (m *mockCampaignsRepo) GetScheduledCampaigns(ctx context.Context, before time.Time) ([]entity.Campaign, error) {
	return m.getScheduledCampaignsFn(ctx, before)
}

type mockScenariosRepo struct {
	createFn       func(ctx context.Context, s *entity.AutoScenario) error
	getByOrgIDFn   func(ctx context.Context, orgID int) ([]entity.AutoScenario, error)
	getByIDFn      func(ctx context.Context, id int) (*entity.AutoScenario, error)
	updateFn       func(ctx context.Context, s *entity.AutoScenario) error
	deleteFn       func(ctx context.Context, id int) error
	getTemplatesFn func(ctx context.Context) ([]entity.AutoScenario, error)
	getActionLogFn func(ctx context.Context, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error)
}

func (m *mockScenariosRepo) Create(ctx context.Context, s *entity.AutoScenario) error {
	return m.createFn(ctx, s)
}
func (m *mockScenariosRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.AutoScenario, error) {
	return m.getByOrgIDFn(ctx, orgID)
}
func (m *mockScenariosRepo) GetByID(ctx context.Context, id int) (*entity.AutoScenario, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockScenariosRepo) Update(ctx context.Context, s *entity.AutoScenario) error {
	return m.updateFn(ctx, s)
}
func (m *mockScenariosRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}
func (m *mockScenariosRepo) GetTemplates(ctx context.Context) ([]entity.AutoScenario, error) {
	if m.getTemplatesFn != nil {
		return m.getTemplatesFn(ctx)
	}
	return nil, nil
}
func (m *mockScenariosRepo) GetActionLog(ctx context.Context, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error) {
	if m.getActionLogFn != nil {
		return m.getActionLogFn(ctx, scenarioID, limit, offset)
	}
	return nil, 0, nil
}

type mockClientsRepo struct {
	countByFilterFn func(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error)
}

func (m *mockClientsRepo) CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error) {
	return m.countByFilterFn(ctx, orgID, filter)
}

func newUC(cr *mockCampaignsRepo, sr *mockScenariosRepo, clr *mockClientsRepo) *campaigns.Usecase {
	return campaigns.New(cr, sr, clr)
}

// --- tests ---

func TestCreate_ReturnsCampaign(t *testing.T) {
	cr := &mockCampaignsRepo{
		createFn: func(_ context.Context, c *entity.Campaign) error {
			c.ID = 1
			return nil
		},
	}
	uc := newUC(cr, &mockScenariosRepo{}, &mockClientsRepo{})

	req := &entity.CreateCampaignRequest{
		BotID:   1,
		Name:    "Test",
		Message: "Hello",
	}
	campaign, err := uc.Create(context.Background(), 10, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if campaign.OrgID != 10 {
		t.Errorf("expected OrgID=10, got %d", campaign.OrgID)
	}
	if campaign.Status != "draft" {
		t.Errorf("expected status=draft, got %s", campaign.Status)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	cr := &mockCampaignsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Campaign, error) {
			return nil, errors.New("not found")
		},
	}
	uc := newUC(cr, &mockScenariosRepo{}, &mockClientsRepo{})

	_, err := uc.GetByID(context.Background(), 1, 99)
	if !errors.Is(err, campaigns.ErrCampaignNotFound) {
		t.Errorf("expected ErrCampaignNotFound, got %v", err)
	}
}

func TestGetByID_WrongOrg(t *testing.T) {
	cr := &mockCampaignsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Campaign, error) {
			return &entity.Campaign{ID: 1, OrgID: 5, Status: "draft"}, nil
		},
	}
	uc := newUC(cr, &mockScenariosRepo{}, &mockClientsRepo{})

	_, err := uc.GetByID(context.Background(), 99, 1)
	if !errors.Is(err, campaigns.ErrNotCampaignOwner) {
		t.Errorf("expected ErrNotCampaignOwner, got %v", err)
	}
}

func TestUpdate_AlreadySent(t *testing.T) {
	cr := &mockCampaignsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Campaign, error) {
			return &entity.Campaign{ID: 1, OrgID: 1, Status: "sent"}, nil
		},
	}
	uc := newUC(cr, &mockScenariosRepo{}, &mockClientsRepo{})

	err := uc.Update(context.Background(), 1, 1, &entity.UpdateCampaignRequest{})
	if !errors.Is(err, campaigns.ErrCampaignAlreadySent) {
		t.Errorf("expected ErrCampaignAlreadySent, got %v", err)
	}
}

func TestDelete_Success(t *testing.T) {
	cr := &mockCampaignsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Campaign, error) {
			return &entity.Campaign{ID: 1, OrgID: 1, Status: "draft"}, nil
		},
		deleteFn: func(_ context.Context, _ int) error {
			return nil
		},
	}
	uc := newUC(cr, &mockScenariosRepo{}, &mockClientsRepo{})

	if err := uc.Delete(context.Background(), 1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPreviewAudience_ReturnsCount(t *testing.T) {
	clr := &mockClientsRepo{
		countByFilterFn: func(_ context.Context, _ int, _ entity.ClientFilter) (int, error) {
			return 25, nil
		},
	}
	uc := newUC(&mockCampaignsRepo{}, &mockScenariosRepo{}, clr)

	count, err := uc.PreviewAudience(context.Background(), 1, entity.AudienceFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 25 {
		t.Errorf("expected 25, got %d", count)
	}
}

func TestCreateScenario_SetsInactive(t *testing.T) {
	sr := &mockScenariosRepo{
		createFn: func(_ context.Context, s *entity.AutoScenario) error {
			s.ID = 1
			return nil
		},
	}
	uc := newUC(&mockCampaignsRepo{}, sr, &mockClientsRepo{})

	req := &entity.CreateScenarioRequest{
		BotID:       1,
		Name:        "Birthday",
		TriggerType: "birthday",
		Message:     "Happy birthday, {name}!",
	}
	scenario, err := uc.CreateScenario(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scenario.IsActive {
		t.Error("new scenario should be inactive by default")
	}
}

func TestUpdateScenario_NotFound(t *testing.T) {
	sr := &mockScenariosRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.AutoScenario, error) {
			return nil, errors.New("not found")
		},
	}
	uc := newUC(&mockCampaignsRepo{}, sr, &mockClientsRepo{})

	err := uc.UpdateScenario(context.Background(), 1, 99, &entity.UpdateScenarioRequest{})
	if !errors.Is(err, campaigns.ErrScenarioNotFound) {
		t.Errorf("expected ErrScenarioNotFound, got %v", err)
	}
}

func TestDeleteScenario_WrongOrg(t *testing.T) {
	sr := &mockScenariosRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.AutoScenario, error) {
			return &entity.AutoScenario{ID: 1, OrgID: 5}, nil
		},
	}
	uc := newUC(&mockCampaignsRepo{}, sr, &mockClientsRepo{})

	err := uc.DeleteScenario(context.Background(), 99, 1)
	if !errors.Is(err, campaigns.ErrNotScenarioOwner) {
		t.Errorf("expected ErrNotScenarioOwner, got %v", err)
	}
}
