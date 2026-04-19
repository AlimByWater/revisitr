package emojipacks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
)

var (
	ErrNotFound = errors.New("emoji pack not found")
	ErrNotOwner = errors.New("not the owner of this emoji pack")
)

type emojiPacksRepo interface {
	Create(ctx context.Context, pack *entity.EmojiPack) error
	GetByID(ctx context.Context, id int) (*entity.EmojiPack, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.EmojiPack, error)
	Update(ctx context.Context, pack *entity.EmojiPack) error
	Delete(ctx context.Context, id int) error
	CreateItem(ctx context.Context, item *entity.EmojiItem) error
	GetItemByID(ctx context.Context, id int) (*entity.EmojiItem, error)
	UpdateItem(ctx context.Context, item *entity.EmojiItem) error
	DeleteItem(ctx context.Context, id int) error
	ReorderItems(ctx context.Context, packID int, itemIDs []int) error
}

type Usecase struct {
	logger *slog.Logger
	repo   emojiPacksRepo
}

func New(repo emojiPacksRepo) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

func (uc *Usecase) Create(ctx context.Context, orgID int, req entity.CreateEmojiPackRequest) (*entity.EmojiPack, error) {
	pack := &entity.EmojiPack{
		OrgID: orgID,
		Name:  req.Name,
	}
	if err := uc.repo.Create(ctx, pack); err != nil {
		return nil, fmt.Errorf("create emoji pack: %w", err)
	}
	return pack, nil
}

func (uc *Usecase) Get(ctx context.Context, orgID, packID int) (*entity.EmojiPack, error) {
	pack, err := uc.repo.GetByID(ctx, packID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if pack.OrgID != orgID {
		return nil, ErrNotOwner
	}
	return pack, nil
}

func (uc *Usecase) List(ctx context.Context, orgID int) ([]entity.EmojiPack, error) {
	return uc.repo.GetByOrgID(ctx, orgID)
}

func (uc *Usecase) Update(ctx context.Context, orgID, packID int, req entity.UpdateEmojiPackRequest) (*entity.EmojiPack, error) {
	pack, err := uc.repo.GetByID(ctx, packID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if pack.OrgID != orgID {
		return nil, ErrNotOwner
	}

	if req.Name != nil {
		pack.Name = *req.Name
	}
	if req.SortOrder != nil {
		pack.SortOrder = *req.SortOrder
	}

	if err := uc.repo.Update(ctx, pack); err != nil {
		return nil, err
	}
	return pack, nil
}

func (uc *Usecase) Delete(ctx context.Context, orgID, packID int) error {
	pack, err := uc.repo.GetByID(ctx, packID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if pack.OrgID != orgID {
		return ErrNotOwner
	}
	return uc.repo.Delete(ctx, packID)
}

func (uc *Usecase) AddItem(ctx context.Context, orgID, packID int, req entity.CreateEmojiItemRequest) (*entity.EmojiItem, error) {
	pack, err := uc.repo.GetByID(ctx, packID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if pack.OrgID != orgID {
		return nil, ErrNotOwner
	}

	item := &entity.EmojiItem{
		PackID:   packID,
		Name:     req.Name,
		ImageURL: req.ImageURL,
	}
	if err := uc.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("add emoji item: %w", err)
	}
	return item, nil
}

func (uc *Usecase) UpdateItem(ctx context.Context, orgID, itemID int, req entity.UpdateEmojiItemRequest) (*entity.EmojiItem, error) {
	item, err := uc.repo.GetItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	pack, err := uc.repo.GetByID(ctx, item.PackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if pack.OrgID != orgID {
		return nil, ErrNotOwner
	}

	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}

	if err := uc.repo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *Usecase) DeleteItem(ctx context.Context, orgID, itemID int) error {
	item, err := uc.repo.GetItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	pack, err := uc.repo.GetByID(ctx, item.PackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if pack.OrgID != orgID {
		return ErrNotOwner
	}
	return uc.repo.DeleteItem(ctx, itemID)
}

func (uc *Usecase) ReorderItems(ctx context.Context, orgID, packID int, req entity.ReorderEmojiItemsRequest) error {
	pack, err := uc.repo.GetByID(ctx, packID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if pack.OrgID != orgID {
		return ErrNotOwner
	}
	return uc.repo.ReorderItems(ctx, packID, req.ItemIDs)
}
