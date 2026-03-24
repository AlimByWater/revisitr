package loyalty

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrProgramNotFound    = fmt.Errorf("loyalty program not found")
	ErrNotProgramOwner    = fmt.Errorf("not authorized to manage this program")
	ErrInsufficientPoints = fmt.Errorf("insufficient points")
	ErrReserveExpired     = fmt.Errorf("reserve expired")
	ErrReserveNotPending  = fmt.Errorf("reserve is not pending")
)

type repository interface {
	CreateProgram(ctx context.Context, program *entity.LoyaltyProgram) error
	GetProgramByID(ctx context.Context, id int) (*entity.LoyaltyProgram, error)
	GetProgramsByOrgID(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
	UpdateProgram(ctx context.Context, program *entity.LoyaltyProgram) error
	CreateLevel(ctx context.Context, level *entity.LoyaltyLevel) error
	GetLevelsByProgramID(ctx context.Context, programID int) ([]entity.LoyaltyLevel, error)
	UpdateLevel(ctx context.Context, level *entity.LoyaltyLevel) error
	DeleteLevel(ctx context.Context, id int) error
	GetClientLoyalty(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
	UpsertClientLoyalty(ctx context.Context, cl *entity.ClientLoyalty) error
	CreateTransaction(ctx context.Context, tx *entity.LoyaltyTransaction) error
	GetClientsWithLevels(ctx context.Context) ([]entity.ClientLoyalty, error)
	CreateReserve(ctx context.Context, reserve *entity.BalanceReserve) error
	GetReserve(ctx context.Context, id int) (*entity.BalanceReserve, error)
	UpdateReserve(ctx context.Context, reserve *entity.BalanceReserve) error
	GetPendingReserves(ctx context.Context, clientID, programID int) ([]entity.BalanceReserve, error)
	ExpireOldReserves(ctx context.Context) (int, error)
}

type Usecase struct {
	repo   repository
	logger *slog.Logger
}

func New(repo repository) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) CreateProgram(ctx context.Context, orgID int, req *entity.CreateProgramRequest) (*entity.LoyaltyProgram, error) {
	program := &entity.LoyaltyProgram{
		OrgID:    orgID,
		Name:     req.Name,
		Type:     req.Type,
		Config:   req.Config,
		IsActive: true,
	}

	if err := uc.repo.CreateProgram(ctx, program); err != nil {
		return nil, fmt.Errorf("usecase.CreateProgram: %w", err)
	}

	return program, nil
}

func (uc *Usecase) GetPrograms(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error) {
	programs, err := uc.repo.GetProgramsByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetPrograms: %w", err)
	}
	return programs, nil
}

func (uc *Usecase) GetProgram(ctx context.Context, id, orgID int) (*entity.LoyaltyProgram, error) {
	program, err := uc.repo.GetProgramByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProgramNotFound
		}
		return nil, fmt.Errorf("usecase.GetProgram: %w", err)
	}

	if program.OrgID != orgID {
		return nil, ErrNotProgramOwner
	}

	return program, nil
}

func (uc *Usecase) UpdateProgram(ctx context.Context, id, orgID int, req *entity.UpdateProgramRequest) (*entity.LoyaltyProgram, error) {
	program, err := uc.GetProgram(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		program.Name = *req.Name
	}
	if req.IsActive != nil {
		program.IsActive = *req.IsActive
	}
	if req.Config != nil {
		program.Config = *req.Config
	}

	if err := uc.repo.UpdateProgram(ctx, program); err != nil {
		return nil, fmt.Errorf("usecase.UpdateProgram: %w", err)
	}

	return uc.repo.GetProgramByID(ctx, id)
}

func (uc *Usecase) CreateLevel(ctx context.Context, programID, orgID int, req *entity.CreateLevelRequest) (*entity.LoyaltyLevel, error) {
	_, err := uc.GetProgram(ctx, programID, orgID)
	if err != nil {
		return nil, err
	}

	level := &entity.LoyaltyLevel{
		ProgramID:     programID,
		Name:          req.Name,
		Threshold:     req.Threshold,
		RewardPercent: req.RewardPercent,
		RewardType:    req.RewardType,
		RewardAmount:  req.RewardAmount,
		SortOrder:     req.SortOrder,
	}

	if err := uc.repo.CreateLevel(ctx, level); err != nil {
		return nil, fmt.Errorf("usecase.CreateLevel: %w", err)
	}

	return level, nil
}

func (uc *Usecase) UpdateLevels(ctx context.Context, programID, orgID int, req *entity.BatchUpdateLevelsRequest) ([]entity.LoyaltyLevel, error) {
	_, err := uc.GetProgram(ctx, programID, orgID)
	if err != nil {
		return nil, err
	}

	for i := range req.Levels {
		req.Levels[i].ProgramID = programID
		if err := uc.repo.UpdateLevel(ctx, &req.Levels[i]); err != nil {
			return nil, fmt.Errorf("usecase.UpdateLevels level %d: %w", req.Levels[i].ID, err)
		}
	}

	return uc.repo.GetLevelsByProgramID(ctx, programID)
}

func (uc *Usecase) DeleteLevel(ctx context.Context, programID, orgID, levelID int) error {
	_, err := uc.GetProgram(ctx, programID, orgID)
	if err != nil {
		return err
	}

	if err := uc.repo.DeleteLevel(ctx, levelID); err != nil {
		return fmt.Errorf("usecase.DeleteLevel: %w", err)
	}

	return nil
}

func (uc *Usecase) EarnPoints(ctx context.Context, clientID, programID int, amount float64, description string) (*entity.ClientLoyalty, error) {
	cl, err := uc.getOrCreateClientLoyalty(ctx, clientID, programID)
	if err != nil {
		return nil, err
	}

	cl.Balance += amount
	cl.TotalEarned += amount

	cl.LevelID = uc.determineLevelID(ctx, programID, cl.TotalEarned)

	tx := &entity.LoyaltyTransaction{
		ClientID:     clientID,
		ProgramID:    programID,
		Type:         "earn",
		Amount:       amount,
		BalanceAfter: cl.Balance,
		Description:  description,
	}

	if err := uc.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("usecase.EarnPoints transaction: %w", err)
	}

	if err := uc.repo.UpsertClientLoyalty(ctx, cl); err != nil {
		return nil, fmt.Errorf("usecase.EarnPoints upsert: %w", err)
	}

	return cl, nil
}

func (uc *Usecase) SpendPoints(ctx context.Context, clientID, programID int, amount float64, description string) (*entity.ClientLoyalty, error) {
	cl, err := uc.getOrCreateClientLoyalty(ctx, clientID, programID)
	if err != nil {
		return nil, err
	}

	if cl.Balance < amount {
		return nil, ErrInsufficientPoints
	}

	cl.Balance -= amount
	cl.TotalSpent += amount

	tx := &entity.LoyaltyTransaction{
		ClientID:     clientID,
		ProgramID:    programID,
		Type:         "spend",
		Amount:       amount,
		BalanceAfter: cl.Balance,
		Description:  description,
	}

	if err := uc.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("usecase.SpendPoints transaction: %w", err)
	}

	if err := uc.repo.UpsertClientLoyalty(ctx, cl); err != nil {
		return nil, fmt.Errorf("usecase.SpendPoints upsert: %w", err)
	}

	return cl, nil
}

func (uc *Usecase) GetBalance(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
	cl, err := uc.repo.GetClientLoyalty(ctx, clientID, programID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &entity.ClientLoyalty{
				ClientID:  clientID,
				ProgramID: programID,
			}, nil
		}
		return nil, fmt.Errorf("usecase.GetBalance: %w", err)
	}
	return cl, nil
}

func (uc *Usecase) getOrCreateClientLoyalty(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
	cl, err := uc.repo.GetClientLoyalty(ctx, clientID, programID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &entity.ClientLoyalty{
				ClientID:  clientID,
				ProgramID: programID,
			}, nil
		}
		return nil, fmt.Errorf("usecase.getOrCreateClientLoyalty: %w", err)
	}
	return cl, nil
}

func (uc *Usecase) CalculateBonus(ctx context.Context, clientID, programID int, checkAmount float64) (float64, error) {
	cl, err := uc.getOrCreateClientLoyalty(ctx, clientID, programID)
	if err != nil {
		return 0, err
	}
	if cl.LevelID == nil {
		return 0, nil
	}

	levels, err := uc.repo.GetLevelsByProgramID(ctx, programID)
	if err != nil {
		return 0, fmt.Errorf("CalculateBonus: %w", err)
	}

	for _, level := range levels {
		if level.ID == *cl.LevelID {
			switch level.RewardType {
			case "fixed":
				return level.RewardAmount, nil
			default:
				return checkAmount * level.RewardPercent / 100, nil
			}
		}
	}
	return 0, nil
}

func (uc *Usecase) EarnFromCheck(ctx context.Context, clientID, programID int, checkAmount float64) (*entity.ClientLoyalty, error) {
	bonus, err := uc.CalculateBonus(ctx, clientID, programID, checkAmount)
	if err != nil {
		return nil, err
	}
	if bonus <= 0 {
		return uc.GetBalance(ctx, clientID, programID)
	}

	desc := fmt.Sprintf("Бонус %.2f с чека %.2f", bonus, checkAmount)
	return uc.EarnPoints(ctx, clientID, programID, bonus, desc)
}

func (uc *Usecase) determineLevelID(ctx context.Context, programID int, totalEarned float64) *int {
	levels, err := uc.repo.GetLevelsByProgramID(ctx, programID)
	if err != nil || len(levels) == 0 {
		return nil
	}

	var bestLevel *entity.LoyaltyLevel
	for i := range levels {
		if float64(levels[i].Threshold) <= totalEarned {
			if bestLevel == nil || levels[i].Threshold > bestLevel.Threshold {
				bestLevel = &levels[i]
			}
		}
	}

	if bestLevel != nil {
		return &bestLevel.ID
	}
	return nil
}
