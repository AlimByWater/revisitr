package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"revisitr/internal/entity"
)

const campaignQueueKey = "revisitr:campaign_queue"

// MessageQueue abstracts message queue operations for future Kafka swap.
type MessageQueue interface {
	Enqueue(ctx context.Context, msg *QueueMessage) error
	Dequeue(ctx context.Context, timeout time.Duration) (*QueueMessage, error)
	Ack(ctx context.Context, msgID string) error
	Len(ctx context.Context) (int64, error)
}

// QueueMessage represents a single campaign message to be sent.
type QueueMessage struct {
	ID         string                  `json:"id"`
	CampaignID int                     `json:"campaign_id"`
	ClientID   int                     `json:"client_id"`
	TelegramID int64                   `json:"telegram_id"`
	Text       string                  `json:"text"`
	MediaURL   string                  `json:"media_url,omitempty"`
	MediaType  string                  `json:"media_type,omitempty"`
	Buttons    []entity.CampaignButton `json:"buttons,omitempty"`
}

// CampaignQueue implements MessageQueue via Redis lists.
type CampaignQueue struct {
	rds *Module
}

// NewCampaignQueue creates a new CampaignQueue.
func NewCampaignQueue(rds *Module) *CampaignQueue {
	return &CampaignQueue{rds: rds}
}

// Enqueue marshals the message to JSON and pushes it to the queue.
func (q *CampaignQueue) Enqueue(ctx context.Context, msg *QueueMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("campaignQueue.Enqueue marshal: %w", err)
	}

	if err := q.rds.Client().RPush(ctx, campaignQueueKey, data).Err(); err != nil {
		return fmt.Errorf("campaignQueue.Enqueue: %w", err)
	}

	return nil
}

// Dequeue blocks until a message is available or timeout expires.
func (q *CampaignQueue) Dequeue(ctx context.Context, timeout time.Duration) (*QueueMessage, error) {
	result, err := q.rds.Client().BLPop(ctx, timeout, campaignQueueKey).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("campaignQueue.Dequeue: %w", err)
	}

	if len(result) < 2 {
		return nil, nil
	}

	var msg QueueMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("campaignQueue.Dequeue unmarshal: %w", err)
	}

	return &msg, nil
}

// Ack is a no-op for Redis simple queue — message already consumed by BLPop.
func (q *CampaignQueue) Ack(_ context.Context, _ string) error {
	return nil
}

// Len returns the number of messages in the queue.
func (q *CampaignQueue) Len(ctx context.Context) (int64, error) {
	length, err := q.rds.Client().LLen(ctx, campaignQueueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("campaignQueue.Len: %w", err)
	}

	return length, nil
}
