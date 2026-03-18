package redis

import (
	"context"
	"fmt"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"
)

type config interface {
	GetHost() string
	GetPort() string
	GetPassword() string
}

type Module struct {
	cfg    config
	client *goredis.Client
}

func New(cfg config) *Module {
	return &Module{cfg: cfg}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.client = goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", m.cfg.GetHost(), m.cfg.GetPort()),
		Password: m.cfg.GetPassword(),
		DB:       0,
	})

	if err := m.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	logger.Info("redis connected",
		"host", m.cfg.GetHost(),
		"port", m.cfg.GetPort(),
	)
	return nil
}

func (m *Module) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

func (m *Module) Client() *goredis.Client {
	return m.client
}
