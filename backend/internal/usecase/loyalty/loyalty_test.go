package loyalty

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"revisitr/internal/entity"
)

// mockRepo implements the repository interface for testing.
type mockRepo struct {
	CreateProgramFn        func(ctx context.Context, program *entity.LoyaltyProgram) error
	GetProgramByIDFn       func(ctx context.Context, id int) (*entity.LoyaltyProgram, error)
	GetProgramsByOrgIDFn   func(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
	UpdateProgramFn        func(ctx context.Context, program *entity.LoyaltyProgram) error
	CreateLevelFn          func(ctx context.Context, level *entity.LoyaltyLevel) error
	GetLevelsByProgramIDFn func(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error)
	UpdateLevelFn          func(ctx context.Context, level *entity.LoyaltyLevel) error
	DeleteLevelFn          func(ctx context.Context, id int) error
	GetClientLoyaltyFn     func(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
	UpsertClientLoyaltyFn  func(ctx context.Context, cl *entity.ClientLoyalty) error
	CreateTransactionFn    func(ctx context.Context, tx *entity.LoyaltyTransaction) error
	GetClientsWithLevelsFn func(ctx context.Context) ([]entity.ClientLoyalty, error)
	CreateReserveFn        func(ctx context.Context, reserve *entity.BalanceReserve) error
	GetReserveFn           func(ctx context.Context, id int) (*entity.BalanceReserve, error)
	UpdateReserveFn        func(ctx context.Context, reserve *entity.BalanceReserve) error
	GetPendingReservesFn   func(ctx context.Context, clientID, programID int) ([]entity.BalanceReserve, error)
	ExpireOldReservesFn    func(ctx context.Context) (int, error)
}

func (m *mockRepo) CreateProgram(ctx context.Context, program *entity.LoyaltyProgram) error {
	return m.CreateProgramFn(ctx, program)
}
func (m *mockRepo) GetProgramByID(ctx context.Context, id int) (*entity.LoyaltyProgram, error) {
	return m.GetProgramByIDFn(ctx, id)
}
func (m *mockRepo) GetProgramsByOrgID(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error) {
	return m.GetProgramsByOrgIDFn(ctx, orgID)
}
func (m *mockRepo) UpdateProgram(ctx context.Context, program *entity.LoyaltyProgram) error {
	return m.UpdateProgramFn(ctx, program)
}
func (m *mockRepo) CreateLevel(ctx context.Context, level *entity.LoyaltyLevel) error {
	return m.CreateLevelFn(ctx, level)
}
func (m *mockRepo) GetLevelsByProgramID(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error) {
	return m.GetLevelsByProgramIDFn(ctx, programID)
}
func (m *mockRepo) UpdateLevel(ctx context.Context, level *entity.LoyaltyLevel) error {
	return m.UpdateLevelFn(ctx, level)
}
func (m *mockRepo) DeleteLevel(ctx context.Context, id int) error {
	return m.DeleteLevelFn(ctx, id)
}
func (m *mockRepo) GetClientLoyalty(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
	return m.GetClientLoyaltyFn(ctx, clientID, programID)
}
func (m *mockRepo) UpsertClientLoyalty(ctx context.Context, cl *entity.ClientLoyalty) error {
	return m.UpsertClientLoyaltyFn(ctx, cl)
}
func (m *mockRepo) CreateTransaction(ctx context.Context, tx *entity.LoyaltyTransaction) error {
	return m.CreateTransactionFn(ctx, tx)
}
func (m *mockRepo) GetClientsWithLevels(ctx context.Context) ([]entity.ClientLoyalty, error) {
	return m.GetClientsWithLevelsFn(ctx)
}
func (m *mockRepo) CreateReserve(ctx context.Context, reserve *entity.BalanceReserve) error {
	return m.CreateReserveFn(ctx, reserve)
}
func (m *mockRepo) GetReserve(ctx context.Context, id int) (*entity.BalanceReserve, error) {
	return m.GetReserveFn(ctx, id)
}
func (m *mockRepo) UpdateReserve(ctx context.Context, reserve *entity.BalanceReserve) error {
	return m.UpdateReserveFn(ctx, reserve)
}
func (m *mockRepo) GetPendingReserves(ctx context.Context, clientID, programID int) ([]entity.BalanceReserve, error) {
	return m.GetPendingReservesFn(ctx, clientID, programID)
}
func (m *mockRepo) ExpireOldReserves(ctx context.Context) (int, error) {
	return m.ExpireOldReservesFn(ctx)
}

func newTestUsecase(repo *mockRepo) *Usecase {
	uc := New(repo)
	_ = uc.Init(context.Background(), slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	return uc
}

func intPtr(v int) *int { return &v }

func TestCalculateBonus(t *testing.T) {
	tests := []struct {
		name        string
		clientLevel *int
		levels      []entity.LoyaltyLevel
		checkAmount float64
		wantBonus   float64
	}{
		{
			name:        "percent type 5% of 1000",
			clientLevel: intPtr(1),
			levels: []entity.LoyaltyLevel{
				{ID: 1, ProgramID: 10, Name: "Silver", RewardType: "percent", RewardPercent: 5, RewardAmount: 0},
			},
			checkAmount: 1000,
			wantBonus:   50,
		},
		{
			name:        "fixed type always returns fixed amount",
			clientLevel: intPtr(2),
			levels: []entity.LoyaltyLevel{
				{ID: 2, ProgramID: 10, Name: "Gold", RewardType: "fixed", RewardPercent: 0, RewardAmount: 30},
			},
			checkAmount: 1000,
			wantBonus:   30,
		},
		{
			name:        "no level returns 0",
			clientLevel: nil,
			levels:      nil,
			checkAmount: 1000,
			wantBonus:   0,
		},
		{
			name:        "percent type 10% of 500",
			clientLevel: intPtr(3),
			levels: []entity.LoyaltyLevel{
				{ID: 3, ProgramID: 10, Name: "Platinum", RewardType: "percent", RewardPercent: 10, RewardAmount: 0},
			},
			checkAmount: 500,
			wantBonus:   50,
		},
		{
			name:        "level not found in levels list returns 0",
			clientLevel: intPtr(99),
			levels: []entity.LoyaltyLevel{
				{ID: 1, ProgramID: 10, Name: "Silver", RewardType: "percent", RewardPercent: 5},
			},
			checkAmount: 1000,
			wantBonus:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				GetClientLoyaltyFn: func(_ context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
					if tt.clientLevel == nil {
						return &entity.ClientLoyalty{
							ClientID:  clientID,
							ProgramID: programID,
							LevelID:   nil,
						}, nil
					}
					return &entity.ClientLoyalty{
						ClientID:  clientID,
						ProgramID: programID,
						LevelID:   tt.clientLevel,
					}, nil
				},
				GetLevelsByProgramIDFn: func(_ context.Context, _ int) ([]entity.LoyaltyLevel, error) {
					return tt.levels, nil
				},
			}

			uc := newTestUsecase(repo)
			got, err := uc.CalculateBonus(context.Background(), 1, 10, tt.checkAmount)
			if err != nil {
				t.Fatalf("CalculateBonus() error: %v", err)
			}
			if got != tt.wantBonus {
				t.Errorf("CalculateBonus() = %v, want %v", got, tt.wantBonus)
			}
		})
	}
}

func TestEarnFromCheck(t *testing.T) {
	t.Run("earns bonus from check", func(t *testing.T) {
		levelID := intPtr(1)
		var earnedAmount float64
		var earnedDesc string

		repo := &mockRepo{
			// CalculateBonus path
			GetClientLoyaltyFn: func(_ context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
				return &entity.ClientLoyalty{
					ClientID:    clientID,
					ProgramID:   programID,
					LevelID:     levelID,
					Balance:     100,
					TotalEarned: 500,
				}, nil
			},
			GetLevelsByProgramIDFn: func(_ context.Context, _ int) ([]entity.LoyaltyLevel, error) {
				return []entity.LoyaltyLevel{
					{ID: 1, ProgramID: 10, Name: "Silver", RewardType: "percent", RewardPercent: 5, Threshold: 0},
				}, nil
			},
			// EarnPoints path
			CreateTransactionFn: func(_ context.Context, tx *entity.LoyaltyTransaction) error {
				earnedAmount = tx.Amount
				earnedDesc = tx.Description
				return nil
			},
			UpsertClientLoyaltyFn: func(_ context.Context, _ *entity.ClientLoyalty) error {
				return nil
			},
		}

		uc := newTestUsecase(repo)
		cl, err := uc.EarnFromCheck(context.Background(), 1, 10, 1000)
		if err != nil {
			t.Fatalf("EarnFromCheck() error: %v", err)
		}

		// 5% of 1000 = 50
		if earnedAmount != 50 {
			t.Errorf("earned amount = %v, want 50", earnedAmount)
		}
		if earnedDesc == "" {
			t.Error("expected non-empty description")
		}
		// Balance should be 100 (existing) + 50 (earned) = 150
		if cl.Balance != 150 {
			t.Errorf("balance = %v, want 150", cl.Balance)
		}
	})

	t.Run("no bonus returns balance without earning", func(t *testing.T) {
		repo := &mockRepo{
			GetClientLoyaltyFn: func(_ context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
				// No level assigned -> CalculateBonus returns 0
				return &entity.ClientLoyalty{
					ClientID:  clientID,
					ProgramID: programID,
					LevelID:   nil,
					Balance:   200,
				}, nil
			},
		}

		uc := newTestUsecase(repo)
		cl, err := uc.EarnFromCheck(context.Background(), 1, 10, 1000)
		if err != nil {
			t.Fatalf("EarnFromCheck() error: %v", err)
		}
		// Should return existing balance without changes
		if cl.Balance != 200 {
			t.Errorf("balance = %v, want 200", cl.Balance)
		}
	})

	t.Run("new client with no loyalty record", func(t *testing.T) {
		repo := &mockRepo{
			GetClientLoyaltyFn: func(_ context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
				return nil, sql.ErrNoRows
			},
		}

		uc := newTestUsecase(repo)
		cl, err := uc.EarnFromCheck(context.Background(), 1, 10, 500)
		if err != nil {
			t.Fatalf("EarnFromCheck() error: %v", err)
		}
		// New client with no level -> 0 bonus -> returns zero balance
		if cl.Balance != 0 {
			t.Errorf("balance = %v, want 0", cl.Balance)
		}
	})
}
