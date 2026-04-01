package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"
)

const (
	ChannelBotReload   = "revisitr:bot:reload"
	ChannelBotStop     = "revisitr:bot:stop"
	ChannelBotStart    = "revisitr:bot:start"
	ChannelBotSettings = "revisitr:bot:settings"
)

// BotEvent is the payload for bot-related events.
type BotEvent struct {
	BotID int    `json:"bot_id"`
	Field string `json:"field,omitempty"` // "welcome", "buttons", "modules", ""
}

// EventBus publishes events to Redis Pub/Sub channels.
type EventBus struct {
	rds    *goredis.Client
	logger *slog.Logger
}

func New(rds *goredis.Client, logger *slog.Logger) *EventBus {
	return &EventBus{rds: rds, logger: logger}
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

func (eb *EventBus) publish(ctx context.Context, channel string, event BotEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("eventbus: marshal event: %w", err)
	}

	if err := eb.rds.Publish(ctx, channel, data).Err(); err != nil {
		eb.logger.Error("eventbus: publish failed",
			"channel", channel,
			"bot_id", event.BotID,
			"error", err,
		)
		return fmt.Errorf("eventbus: publish to %s: %w", channel, err)
	}

	eb.logger.Debug("eventbus: published",
		"channel", channel,
		"bot_id", event.BotID,
		"field", event.Field,
	)
	return nil
}
