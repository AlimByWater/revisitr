package rfm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"
)

type rfmRepo interface {
	GetConfig(ctx context.Context, orgID int) (*entity.RFMConfig, error)
	UpsertConfig(ctx context.Context, cfg *entity.RFMConfig) error
	UpdateCalcStats(ctx context.Context, orgID, clientsProcessed int) error
	InsertHistory(ctx context.Context, h *entity.RFMHistory) error
	GetHistory(ctx context.Context, orgID int, from, to time.Time) ([]entity.RFMHistory, error)
	GetSegmentSummary(ctx context.Context, orgID int) ([]entity.RFMSegmentSummary, error)
	GetSegmentClients(ctx context.Context, orgID int, segment string, sortCol, order string, limit, offset int) ([]entity.SegmentClientRow, int, error)
}

type segmentsRepo interface {
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error)
	Create(ctx context.Context, seg *entity.Segment) error
}

type rfmService interface {
	RecalculateAll(ctx context.Context, orgID int) error
}

type Usecase struct {
	logger   *slog.Logger
	rfmRepo  rfmRepo
	segments segmentsRepo
	rfmSvc   rfmService
}

func New(rfmRepo rfmRepo, segments segmentsRepo, rfmSvc rfmService) *Usecase {
	return &Usecase{
		rfmRepo:  rfmRepo,
		segments: segments,
		rfmSvc:   rfmSvc,
	}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) GetDashboard(ctx context.Context, orgID int) (*entity.RFMDashboard, error) {
	segments, err := uc.rfmRepo.GetSegmentSummary(ctx, orgID)
	if err != nil {
		return nil, err
	}

	cfg, _ := uc.rfmRepo.GetConfig(ctx, orgID)

	to := time.Now()
	from := to.AddDate(0, 0, -90)
	trends, err := uc.rfmRepo.GetHistory(ctx, orgID, from, to)
	if err != nil {
		return nil, err
	}

	return &entity.RFMDashboard{
		Segments: segments,
		Trends:   trends,
		Config:   cfg,
	}, nil
}

func (uc *Usecase) Recalculate(ctx context.Context, orgID int) error {
	// Ensure standard RFM segments exist
	if err := uc.ensureRFMSegments(ctx, orgID); err != nil {
		uc.logger.Error("rfm: ensure segments", "org_id", orgID, "error", err)
	}

	if err := uc.rfmSvc.RecalculateAll(ctx, orgID); err != nil {
		return err
	}

	// Record history snapshot
	segments, err := uc.rfmRepo.GetSegmentSummary(ctx, orgID)
	if err != nil {
		uc.logger.Error("rfm: get summary for history", "error", err)
		return nil
	}

	now := time.Now()
	var total int
	for _, s := range segments {
		total += s.ClientCount
		h := &entity.RFMHistory{
			OrgID:        orgID,
			Segment:      s.Segment,
			ClientCount:  s.ClientCount,
			CalculatedAt: now,
		}
		if err := uc.rfmRepo.InsertHistory(ctx, h); err != nil {
			uc.logger.Error("rfm: insert history", "segment", s.Segment, "error", err)
		}
	}

	if err := uc.rfmRepo.UpdateCalcStats(ctx, orgID, total); err != nil {
		uc.logger.Error("rfm: update calc stats", "error", err)
	}

	return nil
}

func (uc *Usecase) GetConfig(ctx context.Context, orgID int) (*entity.RFMConfig, error) {
	cfg, err := uc.rfmRepo.GetConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		// Return defaults
		return &entity.RFMConfig{
			OrgID:          orgID,
			PeriodDays:     365,
			RecalcInterval: "24h",
		}, nil
	}
	return cfg, nil
}

func (uc *Usecase) UpdateConfig(ctx context.Context, orgID int, req entity.UpdateRFMConfigRequest) (*entity.RFMConfig, error) {
	cfg, _ := uc.rfmRepo.GetConfig(ctx, orgID)
	if cfg == nil {
		cfg = &entity.RFMConfig{
			OrgID:          orgID,
			PeriodDays:     365,
			RecalcInterval: "24h",
		}
	}

	if req.PeriodDays != nil {
		cfg.PeriodDays = *req.PeriodDays
	}
	if req.RecalcInterval != nil {
		cfg.RecalcInterval = *req.RecalcInterval
	}

	if err := uc.rfmRepo.UpsertConfig(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ListTemplates returns all standard RFM templates.
func (uc *Usecase) ListTemplates() []entity.RFMTemplate {
	keys := entity.StandardTemplateKeys()
	templates := make([]entity.RFMTemplate, 0, len(keys))
	for _, k := range keys {
		templates = append(templates, entity.StandardTemplates[k])
	}
	return templates
}

// GetActiveTemplate returns the currently active template for the org.
func (uc *Usecase) GetActiveTemplate(ctx context.Context, orgID int) (*entity.RFMConfig, *entity.RFMTemplate, error) {
	cfg, err := uc.rfmRepo.GetConfig(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}
	if cfg == nil {
		// Default: standard tsr
		defaultTpl := entity.StandardTemplates["tsr"]
		defaultCfg := &entity.RFMConfig{
			OrgID:              orgID,
			PeriodDays:         365,
			RecalcInterval:     "24h",
			ActiveTemplateType: entity.TemplateTypeStandard,
			ActiveTemplateKey:  "tsr",
		}
		return defaultCfg, &defaultTpl, nil
	}

	tpl, ok := cfg.ActiveTemplate()
	if !ok {
		defaultTpl := entity.StandardTemplates["tsr"]
		return cfg, &defaultTpl, nil
	}
	return cfg, &tpl, nil
}

// SetTemplate validates and saves the active template, then triggers recalculation.
func (uc *Usecase) SetTemplate(ctx context.Context, orgID int, req entity.SetTemplateRequest) (*entity.RFMTemplate, error) {
	// Validate
	if req.TemplateType == entity.TemplateTypeStandard {
		if _, ok := entity.StandardTemplates[req.TemplateKey]; !ok {
			return nil, fmt.Errorf("unknown template key: %s", req.TemplateKey)
		}
	} else {
		if err := req.ValidateCustomThresholds(); err != nil {
			return nil, err
		}
	}

	// Load or create config
	cfg, _ := uc.rfmRepo.GetConfig(ctx, orgID)
	if cfg == nil {
		cfg = &entity.RFMConfig{
			OrgID:          orgID,
			PeriodDays:     365,
			RecalcInterval: "24h",
		}
	}

	// Apply template
	cfg.ActiveTemplateType = req.TemplateType
	if req.TemplateType == entity.TemplateTypeStandard {
		cfg.ActiveTemplateKey = req.TemplateKey
		cfg.CustomTemplateName = nil
		cfg.CustomRThresholds = nil
		cfg.CustomFThresholds = nil
	} else {
		cfg.ActiveTemplateKey = "custom"
		cfg.CustomTemplateName = req.CustomName
		rJSON, _ := json.Marshal(req.RThresholds)
		fJSON, _ := json.Marshal(req.FThresholds)
		cfg.CustomRThresholds = rJSON
		cfg.CustomFThresholds = fJSON
	}

	if err := uc.rfmRepo.UpsertConfig(ctx, cfg); err != nil {
		return nil, err
	}

	// Trigger recalculation in background
	go func() {
		bgCtx := context.Background()
		if err := uc.Recalculate(bgCtx, orgID); err != nil {
			uc.logger.Error("rfm: auto-recalculate after template change", "org_id", orgID, "error", err)
		}
	}()

	tpl, _ := cfg.ActiveTemplate()
	return &tpl, nil
}

// GetOnboardingQuestions returns the quiz questions for template recommendation.
func (uc *Usecase) GetOnboardingQuestions() []entity.OnboardingQuestion {
	return entity.GetOnboardingQuestions()
}

// RecommendTemplate processes quiz answers and returns a template recommendation.
func (uc *Usecase) RecommendTemplate(answers []int) (*entity.TemplateRecommendation, error) {
	return entity.RecommendTemplate(answers)
}

// GetSegmentClients returns paginated clients for a specific RFM segment.
func (uc *Usecase) GetSegmentClients(ctx context.Context, orgID int, segment string, page, perPage int, sortCol, order string) (*entity.SegmentClientsResponse, error) {
	// Validate segment
	name, ok := entity.SegmentNames[segment]
	if !ok {
		return nil, fmt.Errorf("unknown segment: %s", segment)
	}

	// Clamp pagination
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	clients, total, err := uc.rfmRepo.GetSegmentClients(ctx, orgID, segment, sortCol, order, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &entity.SegmentClientsResponse{
		Segment:     segment,
		SegmentName: name,
		Total:       total,
		Page:        page,
		PerPage:     perPage,
		Clients:     clients,
	}, nil
}

// ensureRFMSegments creates the 7 standard RFM segments if they don't exist for the org.
func (uc *Usecase) ensureRFMSegments(ctx context.Context, orgID int) error {
	existing, err := uc.segments.GetByOrgID(ctx, orgID)
	if err != nil {
		return err
	}

	rfmExists := make(map[string]bool)
	for _, seg := range existing {
		if seg.Type == "rfm" && seg.Filter.RFMCategory != nil {
			rfmExists[*seg.Filter.RFMCategory] = true
		}
	}

	standardSegments := []struct {
		category string
		name     string
	}{
		{entity.RFMSegmentNew, entity.SegmentNames[entity.RFMSegmentNew]},
		{entity.RFMSegmentPromising, entity.SegmentNames[entity.RFMSegmentPromising]},
		{entity.RFMSegmentRegular, entity.SegmentNames[entity.RFMSegmentRegular]},
		{entity.RFMSegmentVIP, entity.SegmentNames[entity.RFMSegmentVIP]},
		{entity.RFMSegmentRareValue, entity.SegmentNames[entity.RFMSegmentRareValue]},
		{entity.RFMSegmentChurnRisk, entity.SegmentNames[entity.RFMSegmentChurnRisk]},
		{entity.RFMSegmentLost, entity.SegmentNames[entity.RFMSegmentLost]},
	}

	for _, ss := range standardSegments {
		if rfmExists[ss.category] {
			continue
		}
		cat := ss.category
		seg := &entity.Segment{
			OrgID:      orgID,
			Name:       ss.name,
			Type:       "rfm",
			Filter:     entity.SegmentFilter{RFMCategory: &cat},
			AutoAssign: true,
		}
		if err := uc.segments.Create(ctx, seg); err != nil {
			uc.logger.Error("rfm: create segment", "category", ss.category, "error", err)
		}
	}

	return nil
}
