package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"
)

const (
	ChannelBotReload         = "revisitr:bot:reload"
	ChannelBotStop           = "revisitr:bot:stop"
	ChannelBotStart          = "revisitr:bot:start"
	ChannelBotSettings       = "revisitr:bot:settings"
	ChannelNotifyClient      = "revisitr:notify:client"
	ChannelLunchOrderCreated = "revisitr:lunch:order.created"
)

// BotEvent is the payload for bot-related events.
type BotEvent struct {
	BotID int    `json:"bot_id"`
	Field string `json:"field,omitempty"` // "welcome", "buttons", "modules", ""
}

// NotifyClientEvent is the payload for a one-off message to a client's chat.
type NotifyClientEvent struct {
	BotID  int    `json:"bot_id"`
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

// LunchOrderEvent is the payload for a lunch order placed by a guest.
type LunchOrderEvent struct {
	OrderID  int     `json:"order_id"`
	BotID    int     `json:"bot_id"`
	TableNum string  `json:"table_num"`
	Total    float64 `json:"total"`
}

// EventBus publishes events to Redis Pub/Sub channels.
type EventBus struct {
	rdsClient func() *goredis.Client
	logger    *slog.Logger
}

func New(rdsClient func() *goredis.Client, logger *slog.Logger) *EventBus {
	return &EventBus{rdsClient: rdsClient, logger: logger}
}

func (eb *EventBus) PublishBotReload(ctx context.Context, botID int) error {
	return eb.publish(ctx, ChannelBotReload, BotEvent{BotID: botID})
}

func (eb *EventBus) PublishBotStop(ctx context.Context, botID int) error {
	return eb.publish(ctx, ChannelBotStop, BotEvent{BotID: botID})
}

func (eb *EventBus) PublishBotStart(ctx context.Context, botID int) error {
	return eb.publish(ctx, ChannelBotStart, BotEvent{BotID: botID})
}

func (eb *EventBus) PublishBotSettings(ctx context.Context, botID int, field string) error {
	return eb.publish(ctx, ChannelBotSettings, BotEvent{BotID: botID, Field: field})
}

// PublishNotifyClient asks the bot process to send a plain-text message to a
// client's chat via the given bot. Best-effort: callers treat errors as non-fatal.
func (eb *EventBus) PublishNotifyClient(ctx context.Context, botID int, chatID int64, text string) error {
	if eb == nil || eb.rdsClient == nil {
		return fmt.Errorf("eventbus: redis client getter is not configured")
	}

	rds := eb.rdsClient()
	if rds == nil {
		return fmt.Errorf("eventbus: redis client is not initialized")
	}

	data, err := json.Marshal(NotifyClientEvent{BotID: botID, ChatID: chatID, Text: text})
	if err != nil {
		return fmt.Errorf("eventbus: marshal notify event: %w", err)
	}

	if err := rds.Publish(ctx, ChannelNotifyClient, data).Err(); err != nil {
		eb.logger.Error("eventbus: publish failed",
			"channel", ChannelNotifyClient,
			"bot_id", botID,
			"error", err,
		)
		return fmt.Errorf("eventbus: publish to %s: %w", ChannelNotifyClient, err)
	}

	eb.logger.Debug("eventbus: published notify",
		"bot_id", botID,
		"chat_id", chatID,
	)
	return nil
}

func (eb *EventBus) PublishLunchOrderCreated(ctx context.Context, event LunchOrderEvent) error {
	return eb.publish(ctx, ChannelLunchOrderCreated, event)
}

func (eb *EventBus) publish(ctx context.Context, channel string, event any) error {
	if eb == nil || eb.rdsClient == nil {
		return fmt.Errorf("eventbus: redis client getter is not configured")
	}

	rds := eb.rdsClient()
	if rds == nil {
		return fmt.Errorf("eventbus: redis client is not initialized")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("eventbus: marshal event: %w", err)
	}

	if err := rds.Publish(ctx, channel, data).Err(); err != nil {
		eb.logger.Error("eventbus: publish failed",
			"channel", channel,
			"error", err,
		)
		return fmt.Errorf("eventbus: publish to %s: %w", channel, err)
	}

	eb.logger.Debug("eventbus: published", "channel", channel)
	return nil
}
