package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type config interface {
	GetHost() string
	GetPort() string
	GetUser() string
	GetPassword() string
	GetDatabase() string
	GetSSLMode() string
}

type Module struct {
	cfg config
	db  *sqlx.DB
}

func New(cfg config) *Module {
	return &Module{cfg: cfg}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		m.cfg.GetHost(),
		m.cfg.GetPort(),
		m.cfg.GetUser(),
		m.cfg.GetPassword(),
		m.cfg.GetDatabase(),
		m.cfg.GetSSLMode(),
	)

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}

	m.db = db
	logger.Info("postgres connected",
		"host", m.cfg.GetHost(),
		"port", m.cfg.GetPort(),
		"database", m.cfg.GetDatabase(),
	)
	return nil
}

func (m *Module) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *Module) DB() *sqlx.DB {
	return m.db
}
