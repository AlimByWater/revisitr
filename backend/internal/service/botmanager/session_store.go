package botmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const flowStateTTL = 24 * time.Hour

type FlowState struct {
	SelectedPOSID    int    `json:"selected_pos_id,omitempty"`
	Flow             string `json:"flow,omitempty"`
	FlowMessageID    int    `json:"flow_message_id,omitempty"`
	MenuCategoryID   int    `json:"menu_category_id,omitempty"`
	MenuPage         int    `json:"menu_page,omitempty"`
	BookingStage     string `json:"booking_stage,omitempty"`
	BookingPage      int    `json:"booking_page,omitempty"`
	BookingDate      string `json:"booking_date,omitempty"`
	BookingTime      string `json:"booking_time,omitempty"`
	BookingPartySize string `json:"booking_party_size,omitempty"`
	AwaitingFeedback bool   `json:"awaiting_feedback,omitempty"`
}

func (s FlowState) resetFlow() FlowState {
	return FlowState{
		SelectedPOSID: s.SelectedPOSID,
	}
}

type RedisSessionStore struct {
	client *goredis.Client
}

func NewRedisSessionStore(client *goredis.Client) *RedisSessionStore {
	return &RedisSessionStore{client: client}
}

func flowStateKey(botID int, chatID int64) string {
	return fmt.Sprintf("botflow:%d:%d", botID, chatID)
}

func (s *RedisSessionStore) Load(ctx context.Context, botID int, chatID int64) (*FlowState, error) {
	if s == nil || s.client == nil {
		return &FlowState{}, nil
	}

	raw, err := s.client.Get(ctx, flowStateKey(botID, chatID)).Result()
	if err != nil {
		if err == goredis.Nil {
			return &FlowState{}, nil
		}
		return nil, fmt.Errorf("botflow load: %w", err)
	}

	var state FlowState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, fmt.Errorf("botflow decode: %w", err)
	}
	return &state, nil
}

func (s *RedisSessionStore) Save(ctx context.Context, botID int, chatID int64, state FlowState) error {
	if s == nil || s.client == nil {
		return nil
	}

	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("botflow encode: %w", err)
	}

	if err := s.client.Set(ctx, flowStateKey(botID, chatID), payload, flowStateTTL).Err(); err != nil {
		return fmt.Errorf("botflow save: %w", err)
	}
	return nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, botID int, chatID int64) error {
	if s == nil || s.client == nil {
		return nil
	}
	if err := s.client.Del(ctx, flowStateKey(botID, chatID)).Err(); err != nil {
		return fmt.Errorf("botflow delete: %w", err)
	}
	return nil
}
