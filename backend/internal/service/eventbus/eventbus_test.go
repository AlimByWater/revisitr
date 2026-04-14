package eventbus

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestPublishBotSettings_UninitializedRedisClient_ReturnsError(t *testing.T) {
	t.Parallel()

	eb := New(func() *goredis.Client { return nil }, slog.New(slog.NewTextHandler(io.Discard, nil)))

	err := eb.PublishBotSettings(context.Background(), 6, "welcome")
	if err == nil {
		t.Fatal("expected error when redis client is nil")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("unexpected error: %v", err)
	}
}
