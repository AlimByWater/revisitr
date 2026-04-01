package eventbus

import (
	"context"
	"encoding/json"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"
)

// BotEventHandler processes bot-related events from the event bus.
type BotEventHandler interface {
	OnBotReload(ctx context.Context, botID int) error
	OnBotStop(ctx context.Context, botID int) error
	OnBotStart(ctx context.Context, botID int) error
	OnBotSettingsChanged(ctx context.Context, botID int, field string) error
}

// Subscriber listens to Redis Pub/Sub channels and dispatches events.
type Subscriber struct {
	rds    *goredis.Client
	logger *slog.Logger
}

func NewSubscriber(rds *goredis.Client, logger *slog.Logger) *Subscriber {
	return &Subscriber{rds: rds, logger: logger}
}

// Listen subscribes to bot event channels and dispatches to the handler.
// Blocks until ctx is cancelled.
func (s *Subscriber) Listen(ctx context.Context, handler BotEventHandler) {
	pubsub := s.rds.Subscribe(ctx,
		ChannelBotReload,
		ChannelBotStop,
		ChannelBotStart,
		ChannelBotSettings,
	)
	defer pubsub.Close()

	s.logger.Info("eventbus subscriber: listening for bot events")

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("eventbus subscriber: shutting down")
			return
		case msg := <-ch:
			if msg == nil {
				return
			}
			s.dispatch(ctx, msg, handler)
		}
	}
}

func (s *Subscriber) dispatch(ctx context.Context, msg *goredis.Message, handler BotEventHandler) {
	var event BotEvent
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		s.logger.Error("eventbus: unmarshal event",
			"channel", msg.Channel,
			"payload", msg.Payload,
			"error", err,
		)
		return
	}

	s.logger.Debug("eventbus: received event",
		"channel", msg.Channel,
		"bot_id", event.BotID,
		"field", event.Field,
	)

	var err error
	switch msg.Channel {
	case ChannelBotReload:
		err = handler.OnBotReload(ctx, event.BotID)
	case ChannelBotStop:
		err = handler.OnBotStop(ctx, event.BotID)
	case ChannelBotStart:
		err = handler.OnBotStart(ctx, event.BotID)
	case ChannelBotSettings:
		err = handler.OnBotSettingsChanged(ctx, event.BotID, event.Field)
	default:
		s.logger.Warn("eventbus: unknown channel", "channel", msg.Channel)
		return
	}

	if err != nil {
		s.logger.Error("eventbus: handler error",
			"channel", msg.Channel,
			"bot_id", event.BotID,
			"error", err,
		)
	}
}
