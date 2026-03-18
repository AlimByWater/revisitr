package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Sessions struct {
	rds *Module
}

func NewSessions(rds *Module) *Sessions {
	return &Sessions{rds: rds}
}

func refreshTokenKey(tokenHash string) string {
	return "refresh:" + tokenHash
}

func userSessionsKey(userID int) string {
	return "user_sessions:" + strconv.Itoa(userID)
}

func (r *Sessions) StoreRefreshToken(ctx context.Context, userID int, tokenHash string, ttl time.Duration) error {
	pipe := r.rds.Client().Pipeline()
	pipe.Set(ctx, refreshTokenKey(tokenHash), strconv.Itoa(userID), ttl)
	pipe.SAdd(ctx, userSessionsKey(userID), tokenHash)
	pipe.Expire(ctx, userSessionsKey(userID), ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("sessions.StoreRefreshToken: %w", err)
	}
	return nil
}

func (r *Sessions) GetUserIDByToken(ctx context.Context, tokenHash string) (int, error) {
	val, err := r.rds.Client().Get(ctx, refreshTokenKey(tokenHash)).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, fmt.Errorf("sessions.GetUserIDByToken: token not found")
		}
		return 0, fmt.Errorf("sessions.GetUserIDByToken: %w", err)
	}

	userID, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("sessions.GetUserIDByToken parse user_id: %w", err)
	}

	return userID, nil
}

func (r *Sessions) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	val, err := r.rds.Client().Get(ctx, refreshTokenKey(tokenHash)).Result()
	if err != nil && err != goredis.Nil {
		return fmt.Errorf("sessions.DeleteRefreshToken get: %w", err)
	}

	pipe := r.rds.Client().Pipeline()
	pipe.Del(ctx, refreshTokenKey(tokenHash))

	if val != "" {
		userID, _ := strconv.Atoi(val)
		pipe.SRem(ctx, userSessionsKey(userID), tokenHash)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("sessions.DeleteRefreshToken: %w", err)
	}
	return nil
}

func (r *Sessions) DeleteUserSessions(ctx context.Context, userID int) error {
	setKey := userSessionsKey(userID)

	tokenHashes, err := r.rds.Client().SMembers(ctx, setKey).Result()
	if err != nil {
		return fmt.Errorf("sessions.DeleteUserSessions smembers: %w", err)
	}

	if len(tokenHashes) == 0 {
		return nil
	}

	keys := make([]string, 0, len(tokenHashes)+1)
	for _, hash := range tokenHashes {
		keys = append(keys, refreshTokenKey(hash))
	}
	keys = append(keys, setKey)

	if err := r.rds.Client().Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("sessions.DeleteUserSessions del: %w", err)
	}

	return nil
}
