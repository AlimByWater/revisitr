package masterbot

import (
	"context"

	"revisitr/internal/entity"
)

// Legacy admin bot links (admin_bot_links table — backward compat)
type adminBotLinksRepo interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.AdminBotLink, error)
	GetByLinkCode(ctx context.Context, code string) (*entity.AdminBotLink, error)
	ActivateLink(ctx context.Context, id int, telegramID int64) error
}

type dashboardRepository interface {
	GetWidgets(ctx context.Context, orgID int, filter entity.DashboardFilter) (*entity.DashboardWidgets, error)
}

type campaignsRepository interface {
	GetByOrgID(ctx context.Context, orgID, limit, offset int) ([]entity.Campaign, int, error)
}

type promotionsRepository interface {
	CreatePromoCode(ctx context.Context, pc *entity.PromoCode) error
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Promotion, error)
}

// New repos for master bot features

type masterBotLinksRepo interface {
	CreateLink(ctx context.Context, link *entity.MasterBotLink) error
	GetLinkByTelegramID(ctx context.Context, telegramUserID int64) (*entity.MasterBotLink, error)
	GetLinkByOrgID(ctx context.Context, orgID int) ([]entity.MasterBotLink, error)
	DeactivateLink(ctx context.Context, id int) error
}

type masterBotAuthRepo interface {
	StoreToken(ctx context.Context, token string, data entity.MasterBotAuthToken) error
	ValidateAndConsume(ctx context.Context, token string) (*entity.MasterBotAuthToken, error)
}

type botsRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.Bot, error)
	Create(ctx context.Context, bot *entity.Bot) error
	Update(ctx context.Context, bot *entity.Bot) error
}

type postCodesRepo interface {
	Create(ctx context.Context, pc *entity.PostCode) error
	GetByCode(ctx context.Context, orgID int, code string) (*entity.PostCode, error)
	GetByOrgID(ctx context.Context, orgID int) ([]entity.PostCode, error)
	UpdateContent(ctx context.Context, id int, content entity.PostCodeContent) error
}
