package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"revisitr/internal/entity"
)

const (
	masterBotAuthPrefix = "masterbot:auth:"
	masterBotAuthTTL    = 15 * time.Minute
)

type MasterBotAuth struct {
	redis *Module
}

func NewMasterBotAuth(redis *Module) *MasterBotAuth {
	return &MasterBotAuth{redis: redis}
}

// StoreToken saves a one-time auth token with TTL. Returns the token string.
func (r *MasterBotAuth) StoreToken(ctx context.Context, token string, data entity.MasterBotAuthToken) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("master_bot_auth.StoreToken: marshal: %w", err)
	}
	return r.redis.Client().Set(ctx, masterBotAuthPrefix+token, b, masterBotAuthTTL).Err()
}

// ValidateAndConsume retrieves and deletes the token (one-time use).
func (r *MasterBotAuth) ValidateAndConsume(ctx context.Context, token string) (*entity.MasterBotAuthToken, error) {
	key := masterBotAuthPrefix + token

	b, err := r.redis.Client().GetDel(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("master_bot_auth.ValidateAndConsume: %w", err)
	}

	var data entity.MasterBotAuthToken
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("master_bot_auth.ValidateAndConsume: unmarshal: %w", err)
	}
	return &data, nil
}
