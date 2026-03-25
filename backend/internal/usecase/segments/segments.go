package segments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

type segmentsRepo interface {
	Create(ctx context.Context, seg *entity.Segment) error
	GetByID(ctx context.Context, id int) (*entity.Segment, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error)
	Update(ctx context.Context, seg *entity.Segment) error
	Delete(ctx context.Context, id int) error
	GetClients(ctx context.Context, segmentID, limit, offset int) ([]entity.BotClient, int, error)
	SyncClients(ctx context.Context, segmentID int, clientIDs []int) error
	CountByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error)
}

type clientsRepo interface {
	GetIDsByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) ([]int, error)
}

type rulesRepo interface {
	CreateRule(ctx context.Context, segmentID int, req entity.CreateSegmentRuleRequest) (*entity.SegmentRule, error)
	GetRules(ctx context.Context, segmentID int) ([]entity.SegmentRule, error)
	DeleteRule(ctx context.Context, ruleID int) error
	DeleteRulesBySegment(ctx context.Context, segmentID int) error
}

type predictionsRepo interface {
	UpsertPrediction(ctx context.Context, pred *entity.ClientPrediction) error
	GetPredictions(ctx context.Context, orgID int, limit, offset int) ([]entity.ClientPrediction, error)
	GetPredictionByClient(ctx context.Context, clientID int) (*entity.ClientPrediction, error)
	GetHighChurnClients(ctx context.Context, orgID int, threshold float32) ([]entity.ClientPrediction, error)
	GetPredictionSummary(ctx context.Context, orgID int) (*entity.PredictionSummary, error)
}

type Usecase struct {
	logger      *slog.Logger
	segments    segmentsRepo
	clients     clientsRepo
	rules       rulesRepo
	predictions predictionsRepo
}

type Option func(*Usecase)

func WithRules(r rulesRepo) Option           { return func(uc *Usecase) { uc.rules = r } }
func WithPredictions(r predictionsRepo) Option { return func(uc *Usecase) { uc.predictions = r } }

func New(segments segmentsRepo, clients clientsRepo, opts ...Option) *Usecase {
	uc := &Usecase{segments: segments, clients: clients}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) Create(ctx context.Context, orgID int, req *entity.CreateSegmentRequest) (*entity.Segment, error) {
	seg := &entity.Segment{
		OrgID:      orgID,
		Name:       req.Name,
		Type:       req.Type,
		Filter:     req.Filter,
		AutoAssign: req.AutoAssign,
	}

	if err := uc.segments.Create(ctx, seg); err != nil {
		return nil, fmt.Errorf("create segment: %w", err)
	}

	return seg, nil
}

func (uc *Usecase) GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error) {
	segs, err := uc.segments.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list segments: %w", err)
	}
	return segs, nil
}

func (uc *Usecase) GetByID(ctx context.Context, id, orgID int) (*entity.Segment, error) {
	seg, err := uc.segments.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSegmentNotFound
		}
		return nil, fmt.Errorf("get segment: %w", err)
	}

	if seg.OrgID != orgID {
		return nil, ErrNotSegmentOwner
	}

	return seg, nil
}

func (uc *Usecase) Update(ctx context.Context, id, orgID int, req *entity.UpdateSegmentRequest) (*entity.Segment, error) {
	seg, err := uc.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		seg.Name = *req.Name
	}
	if req.Filter != nil {
		seg.Filter = *req.Filter
	}
	if req.AutoAssign != nil {
		seg.AutoAssign = *req.AutoAssign
	}

	if err := uc.segments.Update(ctx, seg); err != nil {
		return nil, fmt.Errorf("update segment: %w", err)
	}

	return seg, nil
}

func (uc *Usecase) Delete(ctx context.Context, id, orgID int) error {
	if _, err := uc.GetByID(ctx, id, orgID); err != nil {
		return err
	}

	if err := uc.segments.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete segment: %w", err)
	}

	return nil
}

func (uc *Usecase) GetClients(ctx context.Context, segmentID, orgID, limit, offset int) ([]entity.BotClient, int, error) {
	if _, err := uc.GetByID(ctx, segmentID, orgID); err != nil {
		return nil, 0, err
	}

	clients, total, err := uc.segments.GetClients(ctx, segmentID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get segment clients: %w", err)
	}

	return clients, total, nil
}

func (uc *Usecase) RecalculateCustom(ctx context.Context, segmentID, orgID int) error {
	seg, err := uc.GetByID(ctx, segmentID, orgID)
	if err != nil {
		return err
	}

	if seg.Type != "custom" {
		return ErrNotCustomSegment
	}

	ids, err := uc.clients.GetIDsByFilter(ctx, orgID, seg.Filter)
	if err != nil {
		return fmt.Errorf("get client ids by filter: %w", err)
	}

	if err := uc.segments.SyncClients(ctx, segmentID, ids); err != nil {
		return fmt.Errorf("sync segment clients: %w", err)
	}

	return nil
}

func (uc *Usecase) PreviewCount(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error) {
	count, err := uc.segments.CountByFilter(ctx, orgID, f)
	if err != nil {
		return 0, fmt.Errorf("preview segment count: %w", err)
	}
	return count, nil
}

// ── Segment Rules ────────────────────────────────────────────────────────────

func (uc *Usecase) AddRule(ctx context.Context, segmentID, orgID int, req entity.CreateSegmentRuleRequest) (*entity.SegmentRule, error) {
	if uc.rules == nil {
		return nil, fmt.Errorf("rules not configured")
	}
	if _, err := uc.GetByID(ctx, segmentID, orgID); err != nil {
		return nil, err
	}
	return uc.rules.CreateRule(ctx, segmentID, req)
}

func (uc *Usecase) GetRules(ctx context.Context, segmentID, orgID int) ([]entity.SegmentRule, error) {
	if uc.rules == nil {
		return nil, nil
	}
	if _, err := uc.GetByID(ctx, segmentID, orgID); err != nil {
		return nil, err
	}
	return uc.rules.GetRules(ctx, segmentID)
}

func (uc *Usecase) DeleteRule(ctx context.Context, segmentID, orgID, ruleID int) error {
	if uc.rules == nil {
		return fmt.Errorf("rules not configured")
	}
	if _, err := uc.GetByID(ctx, segmentID, orgID); err != nil {
		return err
	}
	return uc.rules.DeleteRule(ctx, ruleID)
}

// ── Predictions ──────────────────────────────────────────────────────────────

func (uc *Usecase) GetPredictions(ctx context.Context, orgID, limit, offset int) ([]entity.ClientPrediction, error) {
	if uc.predictions == nil {
		return nil, nil
	}
	return uc.predictions.GetPredictions(ctx, orgID, limit, offset)
}

func (uc *Usecase) GetClientPrediction(ctx context.Context, clientID int) (*entity.ClientPrediction, error) {
	if uc.predictions == nil {
		return nil, ErrPredictionNotFound
	}
	pred, err := uc.predictions.GetPredictionByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if pred == nil {
		return nil, ErrPredictionNotFound
	}
	return pred, nil
}

func (uc *Usecase) GetHighChurnClients(ctx context.Context, orgID int, threshold float32) ([]entity.ClientPrediction, error) {
	if uc.predictions == nil {
		return nil, nil
	}
	if threshold <= 0 {
		threshold = 0.7
	}
	return uc.predictions.GetHighChurnClients(ctx, orgID, threshold)
}

func (uc *Usecase) GetPredictionSummary(ctx context.Context, orgID int) (*entity.PredictionSummary, error) {
	if uc.predictions == nil {
		return &entity.PredictionSummary{}, nil
	}
	return uc.predictions.GetPredictionSummary(ctx, orgID)
}

// ComputePredictions runs heuristic predictions for all clients in an org.
// This should be called by a scheduler task.
func (uc *Usecase) ComputePredictions(ctx context.Context, orgID int) error {
	if uc.predictions == nil {
		return fmt.Errorf("predictions not configured")
	}
	// Heuristic prediction will be implemented when POS/loyalty data is available.
	// For now, this is a placeholder that can be connected to the scheduler.
	uc.logger.Info("compute predictions called", "org_id", orgID)
	return nil
}
