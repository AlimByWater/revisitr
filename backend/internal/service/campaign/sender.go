package campaign

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type campaignsRepo interface {
	GetByID(ctx context.Context, id int) (*entity.Campaign, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateStats(ctx context.Context, id int, stats *entity.CampaignStats) error
	CreateMessagesBatch(ctx context.Context, messages []entity.CampaignMessage) error
	UpdateMessageStatus(ctx context.Context, id int, status string, errorMsg *string) error
}

type botsRepo interface {
	GetByID(ctx context.Context, id int) (*entity.Bot, error)
}

type botClientsRepo interface {
	GetByBotID(ctx context.Context, botID int, limit, offset int) ([]entity.BotClient, int, error)
}

type Sender struct {
	campaigns  campaignsRepo
	bots       botsRepo
	botClients botClientsRepo
	logger     *slog.Logger
}

func NewSender(
	campaigns campaignsRepo,
	bots botsRepo,
	botClients botClientsRepo,
	logger *slog.Logger,
) *Sender {
	return &Sender{
		campaigns:  campaigns,
		bots:       bots,
		botClients: botClients,
		logger:     logger,
	}
}

func (s *Sender) SendCampaign(ctx context.Context, campaignID int) error {
	campaign, err := s.campaigns.GetByID(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("campaign sender: get campaign: %w", err)
	}

	if campaign.Status != "draft" && campaign.Status != "scheduled" {
		return fmt.Errorf("campaign sender: campaign %d has status %s, expected draft or scheduled", campaignID, campaign.Status)
	}

	// Update status to sending
	if err := s.campaigns.UpdateStatus(ctx, campaignID, "sending"); err != nil {
		return fmt.Errorf("campaign sender: update status to sending: %w", err)
	}

	// Get bot token for sending
	bot, err := s.bots.GetByID(ctx, campaign.BotID)
	if err != nil {
		s.campaigns.UpdateStatus(ctx, campaignID, "failed")
		return fmt.Errorf("campaign sender: get bot: %w", err)
	}

	tBot, err := telego.NewBot(bot.Token)
	if err != nil {
		s.campaigns.UpdateStatus(ctx, campaignID, "failed")
		return fmt.Errorf("campaign sender: create telego bot: %w", err)
	}

	// Get all bot clients (paginated)
	var allClients []entity.BotClient
	limit := 100
	offset := 0

	for {
		clients, total, err := s.botClients.GetByBotID(ctx, campaign.BotID, limit, offset)
		if err != nil {
			s.campaigns.UpdateStatus(ctx, campaignID, "failed")
			return fmt.Errorf("campaign sender: get clients: %w", err)
		}

		allClients = append(allClients, clients...)
		offset += limit
		if offset >= total {
			break
		}
	}

	if len(allClients) == 0 {
		s.campaigns.UpdateStatus(ctx, campaignID, "sent")
		return nil
	}

	// Create message records
	messages := make([]entity.CampaignMessage, len(allClients))
	for i, client := range allClients {
		messages[i] = entity.CampaignMessage{
			CampaignID: campaignID,
			ClientID:   client.ID,
			TelegramID: client.TelegramID,
			Status:     "pending",
		}
	}

	if err := s.campaigns.CreateMessagesBatch(ctx, messages); err != nil {
		s.campaigns.UpdateStatus(ctx, campaignID, "failed")
		return fmt.Errorf("campaign sender: create messages batch: %w", err)
	}

	// Send messages
	stats := entity.CampaignStats{Total: len(messages)}
	for i := range messages {
		if ctx.Err() != nil {
			break
		}

		tgMsg := tu.Message(tu.ID(messages[i].TelegramID), campaign.Message)
		_, err := tBot.SendMessage(tgMsg)

		if err != nil {
			stats.Failed++
			errMsg := err.Error()
			s.campaigns.UpdateMessageStatus(ctx, messages[i].ID, "failed", &errMsg)
			s.logger.Warn("campaign message failed",
				"campaign_id", campaignID,
				"client_id", messages[i].ClientID,
				"error", err,
			)
		} else {
			stats.Sent++
			s.campaigns.UpdateMessageStatus(ctx, messages[i].ID, "sent", nil)
		}

		// Rate limit: ~30 msgs/sec (Telegram limit)
		time.Sleep(35 * time.Millisecond)
	}

	// Update campaign stats and status
	s.campaigns.UpdateStats(ctx, campaignID, &stats)

	finalStatus := "sent"
	if stats.Sent == 0 && stats.Failed > 0 {
		finalStatus = "failed"
	}
	s.campaigns.UpdateStatus(ctx, campaignID, finalStatus)

	s.logger.Info("campaign sent",
		"campaign_id", campaignID,
		"total", stats.Total,
		"sent", stats.Sent,
		"failed", stats.Failed,
	)

	return nil
}
