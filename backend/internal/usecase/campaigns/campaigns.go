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
	CreateClick(ctx context.Context, click *entity.CampaignClick) error
	GetClicksByCampaign(ctx context.Context, campaignID int) ([]entity.CampaignClick, error)
	GetCampaignAnalytics(ctx context.Context, campaignID int) (*entity.CampaignAnalyticsDetail, error)
	GetScheduledCampaigns(ctx context.Context, before time.Time) ([]entity.Campaign, error)
}

type scenariosRepo interface {
	Create(ctx context.Context, scenario *entity.AutoScenario) error
	GetByOrgID(ctx context.Context, orgID int) ([]entity.AutoScenario, error)
	GetByID(ctx context.Context, id int) (*entity.AutoScenario, error)
	Update(ctx context.Context, scenario *entity.AutoScenario) error
	Delete(ctx context.Context, id int) error
	GetTemplates(ctx context.Context) ([]entity.AutoScenario, error)
	GetActionLog(ctx context.Context, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error)
}

type variantsRepo interface {
	CreateVariant(ctx context.Context, v *entity.CampaignVariant) error
	GetVariantsByCampaignID(ctx context.Context, campaignID int) ([]entity.CampaignVariant, error)
	UpdateVariantStats(ctx context.Context, id int, stats *entity.CampaignStats) error
	SetVariantWinner(ctx context.Context, campaignID, variantID int) error
	DeleteVariantsByCampaignID(ctx context.Context, campaignID int) error
	GetVariantAnalytics(ctx context.Context, variantID int) (*entity.VariantResult, error)
}

type templatesRepo interface {
	CreateTemplate(ctx context.Context, t *entity.CampaignTemplate) error
	GetTemplates(ctx context.Context, orgID int) ([]entity.CampaignTemplate, error)
	GetTemplateByID(ctx context.Context, id int) (*entity.CampaignTemplate, error)
	UpdateTemplate(ctx context.Context, t *entity.CampaignTemplate) error
	DeleteTemplate(ctx context.Context, id int) error
}

type clientsRepo interface {
	CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error)
}

type Usecase struct {
	logger    *slog.Logger
	campaigns campaignsRepo
	scenarios scenariosRepo
	clients   clientsRepo
	variants  variantsRepo
	templates templatesRepo
}

func New(campaigns campaignsRepo, scenarios scenariosRepo, clients clientsRepo, opts ...Option) *Usecase {
	uc := &Usecase{
		campaigns: campaigns,
		scenarios: scenarios,
		clients:   clients,
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

type Option func(*Usecase)

func WithVariants(r variantsRepo) Option {
	return func(uc *Usecase) { uc.variants = r }
}

func WithTemplates(r templatesRepo) Option {
	return func(uc *Usecase) { uc.templates = r }
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

	if campaign.Status != "draft" && campaign.Status != "scheduled" {
		return ErrCampaignNotSendable
	}

	if err := uc.campaigns.UpdateStatus(ctx, id, "sending"); err != nil {
		return fmt.Errorf("send campaign update status: %w", err)
	}

	return nil
}

func (uc *Usecase) Schedule(ctx context.Context, orgID, id int, at time.Time) error {
	campaign, err := uc.campaigns.GetByID(ctx, id)
	if err != nil {
		return ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return ErrNotCampaignOwner
	}

	if campaign.Status != "draft" {
		return ErrCampaignNotDraft
	}

	campaign.Status = "scheduled"
	campaign.ScheduledAt = &at

	if err := uc.campaigns.UpdateStatus(ctx, id, "scheduled"); err != nil {
		return fmt.Errorf("schedule campaign update status: %w", err)
	}

	campaign.ScheduledAt = &at
	if err := uc.campaigns.Update(ctx, campaign); err != nil {
		return fmt.Errorf("schedule campaign update scheduled_at: %w", err)
	}

	return nil
}

func (uc *Usecase) CancelScheduled(ctx context.Context, orgID, id int) error {
	campaign, err := uc.campaigns.GetByID(ctx, id)
	if err != nil {
		return ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return ErrNotCampaignOwner
	}

	if campaign.Status != "scheduled" {
		return ErrCampaignNotScheduled
	}

	if err := uc.campaigns.UpdateStatus(ctx, id, "draft"); err != nil {
		return fmt.Errorf("cancel scheduled campaign update status: %w", err)
	}

	campaign.ScheduledAt = nil
	if err := uc.campaigns.Update(ctx, campaign); err != nil {
		return fmt.Errorf("cancel scheduled campaign clear scheduled_at: %w", err)
	}

	return nil
}

func (uc *Usecase) RecordClick(ctx context.Context, campaignID, clientID int, buttonIdx *int, url *string) error {
	click := &entity.CampaignClick{
		CampaignID: campaignID,
		ClientID:   clientID,
		ButtonIdx:  buttonIdx,
		URL:        url,
	}

	if err := uc.campaigns.CreateClick(ctx, click); err != nil {
		return fmt.Errorf("record click: %w", err)
	}

	return nil
}

func (uc *Usecase) GetAnalytics(ctx context.Context, orgID, id int) (*entity.CampaignAnalyticsDetail, error) {
	campaign, err := uc.campaigns.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return nil, ErrNotCampaignOwner
	}

	analytics, err := uc.campaigns.GetCampaignAnalytics(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get analytics: %w", err)
	}

	return analytics, nil
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
		Actions:       req.Actions,
		Timing:        req.Timing,
		Conditions:    req.Conditions,
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
	if req.Actions != nil {
		scenario.Actions = *req.Actions
	}
	if req.Timing != nil {
		scenario.Timing = *req.Timing
	}
	if req.Conditions != nil {
		scenario.Conditions = *req.Conditions
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

func (uc *Usecase) GetTemplates(ctx context.Context) ([]entity.AutoScenario, error) {
	templates, err := uc.scenarios.GetTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("get templates: %w", err)
	}
	return templates, nil
}

func (uc *Usecase) CloneTemplate(ctx context.Context, orgID int, templateKey string, botID int) (*entity.AutoScenario, error) {
	templates, err := uc.scenarios.GetTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("clone template get templates: %w", err)
	}

	var template *entity.AutoScenario
	for i := range templates {
		if templates[i].TemplateKey != nil && *templates[i].TemplateKey == templateKey {
			template = &templates[i]
			break
		}
	}
	if template == nil {
		return nil, ErrScenarioNotFound
	}

	scenario := &entity.AutoScenario{
		OrgID:         orgID,
		BotID:         botID,
		Name:          template.Name,
		TriggerType:   template.TriggerType,
		TriggerConfig: template.TriggerConfig,
		Message:       template.Message,
		Actions:       template.Actions,
		Timing:        template.Timing,
		Conditions:    template.Conditions,
		IsTemplate:    false,
		IsActive:      false,
	}

	if err := uc.scenarios.Create(ctx, scenario); err != nil {
		return nil, fmt.Errorf("clone template create: %w", err)
	}

	return scenario, nil
}

// ── A/B Testing ──────────────────────────────────────────────────────────────

func (uc *Usecase) CreateABTest(ctx context.Context, orgID, campaignID int, req *entity.CreateABTestRequest) ([]entity.CampaignVariant, error) {
	campaign, err := uc.campaigns.GetByID(ctx, campaignID)
	if err != nil {
		return nil, ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return nil, ErrNotCampaignOwner
	}

	if campaign.Status != "draft" {
		return nil, ErrCampaignNotDraft
	}

	// Validate percentages sum to 100
	totalPct := 0
	for _, v := range req.Variants {
		totalPct += v.AudiencePct
	}
	if totalPct != 100 {
		return nil, ErrInvalidVariantPct
	}

	// Delete existing variants
	if err := uc.variants.DeleteVariantsByCampaignID(ctx, campaignID); err != nil {
		return nil, fmt.Errorf("create ab test delete existing: %w", err)
	}

	var variants []entity.CampaignVariant
	for _, vReq := range req.Variants {
		v := &entity.CampaignVariant{
			CampaignID:  campaignID,
			Name:        vReq.Name,
			AudiencePct: vReq.AudiencePct,
			Message:     vReq.Message,
			MediaURL:    vReq.MediaURL,
			Buttons:     vReq.Buttons,
			Stats:       entity.CampaignStats{},
		}
		if err := uc.variants.CreateVariant(ctx, v); err != nil {
			return nil, fmt.Errorf("create ab test variant: %w", err)
		}
		variants = append(variants, *v)
	}

	return variants, nil
}

func (uc *Usecase) GetABResults(ctx context.Context, orgID, campaignID int) (*entity.ABTestResults, error) {
	campaign, err := uc.campaigns.GetByID(ctx, campaignID)
	if err != nil {
		return nil, ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return nil, ErrNotCampaignOwner
	}

	variants, err := uc.variants.GetVariantsByCampaignID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("get ab results variants: %w", err)
	}

	results := &entity.ABTestResults{
		CampaignID: campaignID,
	}

	for _, v := range variants {
		vr, err := uc.variants.GetVariantAnalytics(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("get ab results analytics: %w", err)
		}
		results.Variants = append(results.Variants, *vr)
		if v.IsWinner {
			id := v.ID
			results.WinnerID = &id
		}
	}

	return results, nil
}

func (uc *Usecase) PickWinner(ctx context.Context, orgID, campaignID, variantID int) error {
	campaign, err := uc.campaigns.GetByID(ctx, campaignID)
	if err != nil {
		return ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return ErrNotCampaignOwner
	}

	if err := uc.variants.SetVariantWinner(ctx, campaignID, variantID); err != nil {
		return fmt.Errorf("pick winner: %w", err)
	}

	return nil
}

func (uc *Usecase) GetVariants(ctx context.Context, orgID, campaignID int) ([]entity.CampaignVariant, error) {
	campaign, err := uc.campaigns.GetByID(ctx, campaignID)
	if err != nil {
		return nil, ErrCampaignNotFound
	}

	if campaign.OrgID != orgID {
		return nil, ErrNotCampaignOwner
	}

	variants, err := uc.variants.GetVariantsByCampaignID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("get variants: %w", err)
	}

	return variants, nil
}

// ── Campaign Templates ───────────────────────────────────────────────────────

func (uc *Usecase) CreateCampaignTemplate(ctx context.Context, orgID int, req *entity.CreateCampaignTemplateRequest) (*entity.CampaignTemplate, error) {
	t := &entity.CampaignTemplate{
		OrgID:          &orgID,
		Name:           req.Name,
		Category:       req.Category,
		Description:    req.Description,
		Message:        req.Message,
		MediaURL:       req.MediaURL,
		Buttons:        req.Buttons,
		AudienceFilter: req.AudienceFilter,
		TrackingMode:   req.TrackingMode,
	}

	if t.Category == "" {
		t.Category = "general"
	}
	if t.TrackingMode == "" {
		t.TrackingMode = "none"
	}

	if err := uc.templates.CreateTemplate(ctx, t); err != nil {
		return nil, fmt.Errorf("create campaign template: %w", err)
	}

	return t, nil
}

func (uc *Usecase) ListCampaignTemplates(ctx context.Context, orgID int) ([]entity.CampaignTemplate, error) {
	templates, err := uc.templates.GetTemplates(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list campaign templates: %w", err)
	}

	return templates, nil
}

func (uc *Usecase) GetCampaignTemplate(ctx context.Context, orgID, id int) (*entity.CampaignTemplate, error) {
	t, err := uc.templates.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	// Allow access to system templates or own templates
	if t.OrgID != nil && *t.OrgID != orgID {
		return nil, ErrNotTemplateOwner
	}

	return t, nil
}

func (uc *Usecase) UpdateCampaignTemplate(ctx context.Context, orgID, id int, req *entity.UpdateCampaignTemplateRequest) error {
	t, err := uc.templates.GetTemplateByID(ctx, id)
	if err != nil {
		return ErrTemplateNotFound
	}

	if t.IsSystem {
		return ErrTemplateIsSystem
	}

	if t.OrgID != nil && *t.OrgID != orgID {
		return ErrNotTemplateOwner
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Category != nil {
		t.Category = *req.Category
	}
	if req.Description != nil {
		t.Description = req.Description
	}
	if req.Message != nil {
		t.Message = *req.Message
	}
	if req.MediaURL != nil {
		t.MediaURL = req.MediaURL
	}
	if req.Buttons != nil {
		t.Buttons = *req.Buttons
	}
	if req.AudienceFilter != nil {
		t.AudienceFilter = *req.AudienceFilter
	}
	if req.TrackingMode != nil {
		t.TrackingMode = *req.TrackingMode
	}

	if err := uc.templates.UpdateTemplate(ctx, t); err != nil {
		return fmt.Errorf("update campaign template: %w", err)
	}

	return nil
}

func (uc *Usecase) DeleteCampaignTemplate(ctx context.Context, orgID, id int) error {
	t, err := uc.templates.GetTemplateByID(ctx, id)
	if err != nil {
		return ErrTemplateNotFound
	}

	if t.IsSystem {
		return ErrTemplateIsSystem
	}

	if t.OrgID != nil && *t.OrgID != orgID {
		return ErrNotTemplateOwner
	}

	if err := uc.templates.DeleteTemplate(ctx, id); err != nil {
		return fmt.Errorf("delete campaign template: %w", err)
	}

	return nil
}

func (uc *Usecase) CreateCampaignFromTemplate(ctx context.Context, orgID, templateID, botID int) (*entity.Campaign, error) {
	t, err := uc.templates.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	// Allow system templates or own templates
	if t.OrgID != nil && *t.OrgID != orgID {
		return nil, ErrNotTemplateOwner
	}

	campaign := &entity.Campaign{
		OrgID:          orgID,
		BotID:          botID,
		Name:           t.Name,
		Type:           "manual",
		Status:         "draft",
		AudienceFilter: t.AudienceFilter,
		Message:        t.Message,
		MediaURL:       t.MediaURL,
		Buttons:        t.Buttons,
		TrackingMode:   t.TrackingMode,
		Stats:          entity.CampaignStats{},
	}

	if err := uc.campaigns.Create(ctx, campaign); err != nil {
		return nil, fmt.Errorf("create campaign from template: %w", err)
	}

	return campaign, nil
}

// Scenario methods (continued)

func (uc *Usecase) GetActionLog(ctx context.Context, orgID, scenarioID, limit, offset int) ([]entity.AutoActionLog, int, error) {
	scenario, err := uc.scenarios.GetByID(ctx, scenarioID)
	if err != nil {
		return nil, 0, ErrScenarioNotFound
	}

	if scenario.OrgID != orgID {
		return nil, 0, ErrNotScenarioOwner
	}

	logs, total, err := uc.scenarios.GetActionLog(ctx, scenarioID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get action log: %w", err)
	}

	return logs, total, nil
}
