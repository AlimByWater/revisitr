package emojipacks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"revisitr/internal/entity"
	"revisitr/internal/service/emojisync"

	"github.com/mymmrac/telego"
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
	UpdateItemTg(ctx context.Context, itemID int, stickerSet, customEmojiID string) error
	DeleteItem(ctx context.Context, id int) error
	ReorderItems(ctx context.Context, packID int, itemIDs []int) error
}

type botsRepo interface {
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
}

type Usecase struct {
	logger    *slog.Logger
	repo      emojiPacksRepo
	bots      botsRepo
	syncSvc   *emojisync.Service
}

func New(repo emojiPacksRepo, opts ...Option) *Usecase {
	uc := &Usecase{repo: repo}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

type Option func(*Usecase)

func WithSync(bots botsRepo, syncSvc *emojisync.Service) Option {
	return func(uc *Usecase) {
		uc.bots = bots
		uc.syncSvc = syncSvc
	}
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

var (
	ErrSyncNotConfigured = errors.New("emoji sync not configured")
	ErrBotNoOwner        = errors.New("bot has no owner telegram ID")
)

func (uc *Usecase) SyncToTelegram(ctx context.Context, orgID, packID, botID int) (*entity.EmojiPack, error) {
	if uc.syncSvc == nil || uc.bots == nil {
		return nil, ErrSyncNotConfigured
	}

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

	bot, err := uc.bots.GetByID(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("get bot: %w", err)
	}
	if bot.OrgID != orgID {
		return nil, ErrNotOwner
	}
	if bot.CreatedByTelegramID == nil {
		return nil, ErrBotNoOwner
	}

	// Create telego bot from token
	tgBot, err := telego.NewBot(bot.Token)
	if err != nil {
		return nil, fmt.Errorf("create telego bot: %w", err)
	}

	syncedItems, err := uc.syncSvc.SyncPack(ctx, tgBot, bot.Username, *bot.CreatedByTelegramID, pack)
	if err != nil {
		return nil, fmt.Errorf("sync pack: %w", err)
	}

	// Persist custom_emoji_ids
	for _, item := range syncedItems {
		if item.TgCustomEmojiID != nil && *item.TgCustomEmojiID != "" {
			if err := uc.repo.UpdateItemTg(ctx, item.ID, *item.TgStickerSet, *item.TgCustomEmojiID); err != nil {
				uc.logger.Error("persist custom_emoji_id", "error", err, "item_id", item.ID)
			}
		}
	}

	// Return updated pack
	return uc.repo.GetByID(ctx, packID)
}
