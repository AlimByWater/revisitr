package rfm

import (
	"context"
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
		{entity.RFMChampions, "Чемпионы"},
		{entity.RFMLoyal, "Лояльные"},
		{"potential_loyalist", "Потенциально лояльные"},
		{"new_customers", "Новые клиенты"},
		{entity.RFMAtRisk, "В зоне риска"},
		{entity.RFMCantLose, "Нельзя потерять"},
		{entity.RFMLost, "Потерянные"},
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
