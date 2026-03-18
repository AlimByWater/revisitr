package campaigns

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

type campaignsRepo interface {
	Create(ctx context.Context, campaign *entity.Campaign) error
	GetByID(ctx context.Context, id int) (*entity.Campaign, error)
	GetByOrgID(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error)
	Update(ctx context.Context, campaign *entity.Campaign) error
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateStats(ctx context.Context, id int, stats *entity.CampaignStats) error
	Delete(ctx context.Context, id int) error
	CreateMessagesBatch(ctx context.Context, messages []entity.CampaignMessage) error
	UpdateMessageStatus(ctx context.Context, id int, status string, errorMsg *string) error
	GetMessagesByCampaignID(ctx context.Context, campaignID, limit, offset int) ([]entity.CampaignMessage, int, error)
	GetMessageStats(ctx context.Context, campaignID int) (*entity.CampaignStats, error)
}

type scenariosRepo interface {
	Create(ctx context.Context, scenario *entity.AutoScenario) error
	GetByOrgID(ctx context.Context, orgID int) ([]entity.AutoScenario, error)
	GetByID(ctx context.Context, id int) (*entity.AutoScenario, error)
	Update(ctx context.Context, scenario *entity.AutoScenario) error
	Delete(ctx context.Context, id int) error
}

type clientsRepo interface {
	CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error)
}

type Usecase struct {
	logger    *slog.Logger
	campaigns campaignsRepo
	scenarios scenariosRepo
	clients   clientsRepo
}

func New(campaigns campaignsRepo, scenarios scenariosRepo, clients clientsRepo) *Usecase {
	return &Usecase{
		campaigns: campaigns,
		scenarios: scenarios,
		clients:   clients,
	}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// Campaign methods

func (uc *Usecase) Create(ctx context.Context, orgID int, req *entity.CreateCampaignRequest) (*entity.Campaign, error) {
	campaign := &entity.Campaign{
		OrgID:          orgID,
		BotID:          req.BotID,
		Name:           req.Name,
		Type:           "manual",
		Status:         "draft",
		AudienceFilter: req.AudienceFilter,
		Message:        req.Message,
		MediaURL:       req.MediaURL,
		ScheduledAt:    req.ScheduledAt,
		Stats:          entity.CampaignStats{},
	}

	if err := uc.campaigns.Create(ctx, campaign); err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}

	return campaign, nil
}

func (uc *Usecase) List(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error) {
	campaigns, total, err := uc.campaigns.GetByOrgID(ctx, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list campaigns: %w", err)
	}

	return campaigns, total, nil
}

func (uc *Usecase) GetByID(ctx context.Context, orgID, id int) (*entity.Campaign, error) {
	campaign, err := uc.campaigns.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return nil, ErrNotCampaignOwner
	}

	return campaign, nil
}

func (uc *Usecase) Update(ctx context.Context, orgID, id int, req *entity.UpdateCampaignRequest) error {
	campaign, err := uc.campaigns.GetByID(ctx, id)
	if err != nil {
		return ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return ErrNotCampaignOwner
	}

	if campaign.Status != "draft" {
		return ErrCampaignAlreadySent
	}

	if req.Name != nil {
		campaign.Name = *req.Name
	}
	if req.Message != nil {
		campaign.Message = *req.Message
	}
	if req.AudienceFilter != nil {
		campaign.AudienceFilter = *req.AudienceFilter
	}
	if req.MediaURL != nil {
		campaign.MediaURL = req.MediaURL
	}
	if req.ScheduledAt != nil {
		campaign.ScheduledAt = req.ScheduledAt
	}

	if err := uc.campaigns.Update(ctx, campaign); err != nil {
		return fmt.Errorf("update campaign: %w", err)
	}

	return nil
}

func (uc *Usecase) Delete(ctx context.Context, orgID, id int) error {
	campaign, err := uc.campaigns.GetByID(ctx, id)
	if err != nil {
		return ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return ErrNotCampaignOwner
	}

	if campaign.Status != "draft" {
		return ErrCampaignAlreadySent
	}

	if err := uc.campaigns.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete campaign: %w", err)
	}

	return nil
}

func (uc *Usecase) Send(ctx context.Context, orgID, id int) error {
	campaign, err := uc.campaigns.GetByID(ctx, id)
	if err != nil {
		return ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return ErrNotCampaignOwner
	}

	if campaign.Status != "draft" {
		return ErrCampaignAlreadySent
	}

	// MVP: just update status to "sent" and set sent_at
	now := time.Now()
	if err := uc.campaigns.UpdateStatus(ctx, id, "sent"); err != nil {
		return fmt.Errorf("send campaign update status: %w", err)
	}

	campaign.SentAt = &now
	if err := uc.campaigns.Update(ctx, campaign); err != nil {
		uc.logger.Error("send campaign update sent_at", "error", err, "campaign_id", id)
	}

	return nil
}

func (uc *Usecase) PreviewAudience(ctx context.Context, orgID int, filter entity.AudienceFilter) (int, error) {
	clientFilter := entity.ClientFilter{
		BotID: filter.BotID,
	}

	count, err := uc.clients.CountByFilter(ctx, orgID, clientFilter)
	if err != nil {
		return 0, fmt.Errorf("preview audience: %w", err)
	}

	return count, nil
}

// Scenario methods

func (uc *Usecase) CreateScenario(ctx context.Context, orgID int, req *entity.CreateScenarioRequest) (*entity.AutoScenario, error) {
	scenario := &entity.AutoScenario{
		OrgID:         orgID,
		BotID:         req.BotID,
		Name:          req.Name,
		TriggerType:   req.TriggerType,
		TriggerConfig: req.TriggerConfig,
		Message:       req.Message,
		IsActive:      false,
	}

	if err := uc.scenarios.Create(ctx, scenario); err != nil {
		return nil, fmt.Errorf("create scenario: %w", err)
	}

	return scenario, nil
}

func (uc *Usecase) ListScenarios(ctx context.Context, orgID int) ([]entity.AutoScenario, error) {
	scenarios, err := uc.scenarios.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list scenarios: %w", err)
	}

	return scenarios, nil
}

func (uc *Usecase) UpdateScenario(ctx context.Context, orgID, id int, req *entity.UpdateScenarioRequest) error {
	scenario, err := uc.scenarios.GetByID(ctx, id)
	if err != nil {
		return ErrScenarioNotFound
	}

	if scenario.OrgID != orgID {
		return ErrNotScenarioOwner
	}

	if req.Name != nil {
		scenario.Name = *req.Name
	}
	if req.TriggerConfig != nil {
		scenario.TriggerConfig = *req.TriggerConfig
	}
	if req.Message != nil {
		scenario.Message = *req.Message
	}
	if req.IsActive != nil {
		scenario.IsActive = *req.IsActive
	}

	if err := uc.scenarios.Update(ctx, scenario); err != nil {
		return fmt.Errorf("update scenario: %w", err)
	}

	return nil
}

func (uc *Usecase) DeleteScenario(ctx context.Context, orgID, id int) error {
	scenario, err := uc.scenarios.GetByID(ctx, id)
	if err != nil {
		return ErrScenarioNotFound
	}

	if scenario.OrgID != orgID {
		return ErrNotScenarioOwner
	}

	if err := uc.scenarios.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete scenario: %w", err)
	}

	return nil
}

func (uc *Usecase) ToggleScenario(ctx context.Context, orgID, id int, active bool) error {
	scenario, err := uc.scenarios.GetByID(ctx, id)
	if err != nil {
		return ErrScenarioNotFound
	}

	if scenario.OrgID != orgID {
		return ErrNotScenarioOwner
	}

	scenario.IsActive = active

	if err := uc.scenarios.Update(ctx, scenario); err != nil {
		return fmt.Errorf("toggle scenario: %w", err)
	}

	return nil
}
