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

type Usecase struct {
	logger   *slog.Logger
	segments segmentsRepo
	clients  clientsRepo
}

func New(segments segmentsRepo, clients clientsRepo) *Usecase {
	return &Usecase{segments: segments, clients: clients}
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
