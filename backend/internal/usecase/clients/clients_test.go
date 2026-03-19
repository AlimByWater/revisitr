package clients_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"revisitr/internal/entity"
	"revisitr/internal/usecase/clients"
)

// --- mock ---

type mockRepo struct {
	getByOrgIDFn             func(ctx context.Context, orgID int, filter entity.ClientFilter) ([]entity.ClientProfile, int, error)
	getByIDFn                func(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error)
	updateFn                 func(ctx context.Context, orgID, clientID int, req *entity.UpdateClientRequest) error
	getStatsFn               func(ctx context.Context, orgID int) (*entity.ClientStats, error)
	countByFilterFn          func(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error)
	getTransactionsByClientID func(ctx context.Context, clientID int, limit, offset int) ([]entity.LoyaltyTransaction, error)
}

func (m *mockRepo) GetByOrgID(ctx context.Context, orgID int, filter entity.ClientFilter) ([]entity.ClientProfile, int, error) {
	return m.getByOrgIDFn(ctx, orgID, filter)
}
func (m *mockRepo) GetByID(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error) {
	return m.getByIDFn(ctx, orgID, clientID)
}
func (m *mockRepo) Update(ctx context.Context, orgID, clientID int, req *entity.UpdateClientRequest) error {
	return m.updateFn(ctx, orgID, clientID, req)
}
func (m *mockRepo) GetStats(ctx context.Context, orgID int) (*entity.ClientStats, error) {
	return m.getStatsFn(ctx, orgID)
}
func (m *mockRepo) CountByFilter(ctx context.Context, orgID int, filter entity.ClientFilter) (int, error) {
	return m.countByFilterFn(ctx, orgID, filter)
}
func (m *mockRepo) GetTransactionsByClientID(ctx context.Context, clientID int, limit, offset int) ([]entity.LoyaltyTransaction, error) {
	return m.getTransactionsByClientID(ctx, clientID, limit, offset)
}

// --- tests ---

func TestList_ReturnsProfiles(t *testing.T) {
	profiles := []entity.ClientProfile{{BotClient: entity.BotClient{ID: 1}}}
	repo := &mockRepo{
		getByOrgIDFn: func(_ context.Context, _ int, _ entity.ClientFilter) ([]entity.ClientProfile, int, error) {
			return profiles, 1, nil
		},
	}
	uc := clients.New(repo)

	result, err := uc.List(context.Background(), 1, entity.ClientFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total=1, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
}

func TestList_PropagatesRepoError(t *testing.T) {
	repo := &mockRepo{
		getByOrgIDFn: func(_ context.Context, _ int, _ entity.ClientFilter) ([]entity.ClientProfile, int, error) {
			return nil, 0, errors.New("db error")
		},
	}
	uc := clients.New(repo)

	_, err := uc.List(context.Background(), 1, entity.ClientFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
			return nil, sql.ErrNoRows
		},
	}
	uc := clients.New(repo)

	_, err := uc.GetProfile(context.Background(), 1, 99)
	if !errors.Is(err, clients.ErrClientNotFound) {
		t.Errorf("expected ErrClientNotFound, got %v", err)
	}
}

func TestGetProfile_ReturnsProfileWithTransactions(t *testing.T) {
	profile := &entity.ClientProfile{BotClient: entity.BotClient{ID: 5}}
	txs := []entity.LoyaltyTransaction{{ID: 10}}
	repo := &mockRepo{
		getByIDFn: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
			return profile, nil
		},
		getTransactionsByClientID: func(_ context.Context, _ int, _, _ int) ([]entity.LoyaltyTransaction, error) {
			return txs, nil
		},
	}
	uc := clients.New(repo)

	got, err := uc.GetProfile(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(got.Transactions))
	}
}

func TestUpdateTags_NotFound(t *testing.T) {
	repo := &mockRepo{
		updateFn: func(_ context.Context, _, _ int, _ *entity.UpdateClientRequest) error {
			return sql.ErrNoRows
		},
	}
	uc := clients.New(repo)

	err := uc.UpdateTags(context.Background(), 1, 99, &entity.UpdateClientRequest{})
	if !errors.Is(err, clients.ErrClientNotFound) {
		t.Errorf("expected ErrClientNotFound, got %v", err)
	}
}

func TestGetStats_ReturnsStats(t *testing.T) {
	stats := &entity.ClientStats{TotalClients: 42}
	repo := &mockRepo{
		getStatsFn: func(_ context.Context, _ int) (*entity.ClientStats, error) {
			return stats, nil
		},
	}
	uc := clients.New(repo)

	got, err := uc.GetStats(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalClients != 42 {
		t.Errorf("expected TotalClients=42, got %d", got.TotalClients)
	}
}

func TestCountByFilter_ReturnsCount(t *testing.T) {
	repo := &mockRepo{
		countByFilterFn: func(_ context.Context, _ int, _ entity.ClientFilter) (int, error) {
			return 7, nil
		},
	}
	uc := clients.New(repo)

	count, err := uc.CountByFilter(context.Background(), 1, entity.ClientFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 7 {
		t.Errorf("expected 7, got %d", count)
	}
}
