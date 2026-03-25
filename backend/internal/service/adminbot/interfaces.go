package adminbot

import (
	"context"

	"revisitr/internal/entity"
)

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
