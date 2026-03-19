package segments

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockSegmentsRepo struct {
	createFn        func(ctx context.Context, seg *entity.Segment) error
	getByIDFn       func(ctx context.Context, id int) (*entity.Segment, error)
	getByOrgIDFn    func(ctx context.Context, orgID int) ([]entity.Segment, error)
	updateFn        func(ctx context.Context, seg *entity.Segment) error
	deleteFn        func(ctx context.Context, id int) error
	getClientsFn    func(ctx context.Context, segmentID, limit, offset int) ([]entity.BotClient, int, error)
	syncClientsFn   func(ctx context.Context, segmentID int, clientIDs []int) error
	countByFilterFn func(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error)
}

func (m *mockSegmentsRepo) Create(ctx context.Context, seg *entity.Segment) error {
	return m.createFn(ctx, seg)
}
func (m *mockSegmentsRepo) GetByID(ctx context.Context, id int) (*entity.Segment, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockSegmentsRepo) GetByOrgID(ctx context.Context, orgID int) ([]entity.Segment, error) {
	return m.getByOrgIDFn(ctx, orgID)
}
func (m *mockSegmentsRepo) Update(ctx context.Context, seg *entity.Segment) error {
	return m.updateFn(ctx, seg)
}
func (m *mockSegmentsRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}
func (m *mockSegmentsRepo) GetClients(ctx context.Context, segmentID, limit, offset int) ([]entity.BotClient, int, error) {
	return m.getClientsFn(ctx, segmentID, limit, offset)
}
func (m *mockSegmentsRepo) SyncClients(ctx context.Context, segmentID int, clientIDs []int) error {
	return m.syncClientsFn(ctx, segmentID, clientIDs)
}
func (m *mockSegmentsRepo) CountByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) (int, error) {
	return m.countByFilterFn(ctx, orgID, f)
}

type mockClientsRepo struct {
	getIDsByFilterFn func(ctx context.Context, orgID int, f entity.SegmentFilter) ([]int, error)
}

func (m *mockClientsRepo) GetIDsByFilter(ctx context.Context, orgID int, f entity.SegmentFilter) ([]int, error) {
	return m.getIDsByFilterFn(ctx, orgID, f)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func testSegment(id, orgID int) *entity.Segment {
	return &entity.Segment{
		ID:        id,
		OrgID:     orgID,
		Name:      "test segment",
		Type:      "custom",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	repo := &mockSegmentsRepo{
		createFn: func(_ context.Context, seg *entity.Segment) error {
			seg.ID = 1
			return nil
		},
	}
	uc := New(repo, &mockClientsRepo{})

	seg, err := uc.Create(context.Background(), 10, &entity.CreateSegmentRequest{
		Name: "VIP", Type: "custom",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seg.OrgID != 10 {
		t.Errorf("expected org_id=10, got %d", seg.OrgID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, _ int) (*entity.Segment, error) {
			return nil, fmt.Errorf("segments.GetByID: %w", sql.ErrNoRows)
		},
	}
	uc := New(repo, &mockClientsRepo{})

	_, err := uc.GetByID(context.Background(), 99, 1)
	if err != ErrSegmentNotFound {
		t.Errorf("expected ErrSegmentNotFound, got: %v", err)
	}
}

func TestGetByID_WrongOrg(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 5), nil // owned by org 5
		},
	}
	uc := New(repo, &mockClientsRepo{})

	_, err := uc.GetByID(context.Background(), 1, 99) // requesting as org 99
	if err != ErrNotSegmentOwner {
		t.Errorf("expected ErrNotSegmentOwner, got: %v", err)
	}
}

func TestDelete_OwnershipCheck(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			return testSegment(id, 5), nil // owned by org 5
		},
		deleteFn: func(_ context.Context, _ int) error { return nil },
	}
	uc := New(repo, &mockClientsRepo{})

	if err := uc.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("owner should be able to delete: %v", err)
	}

	if err := uc.Delete(context.Background(), 1, 99); err != ErrNotSegmentOwner {
		t.Errorf("expected ErrNotSegmentOwner for non-owner, got: %v", err)
	}
}

func TestRecalculateCustom_NonCustomSegment(t *testing.T) {
	repo := &mockSegmentsRepo{
		getByIDFn: func(_ context.Context, id int) (*entity.Segment, error) {
			seg := testSegment(id, 1)
			seg.Type = "rfm" // not custom
			return seg, nil
		},
	}
	uc := New(repo, &mockClientsRepo{})

	err := uc.RecalculateCustom(context.Background(), 1, 1)
	if err != ErrNotCustomSegment {
		t.Errorf("expected ErrNotCustomSegment, got: %v", err)
	}
}
