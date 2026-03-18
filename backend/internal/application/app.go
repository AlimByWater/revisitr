package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

type Repository interface {
	Init(ctx context.Context, logger *slog.Logger) error
	Close() error
}

type Usecase interface {
	Init(ctx context.Context, logger *slog.Logger) error
}

type Controller interface {
	Init(ctx context.Context, stop context.CancelFunc, logger *slog.Logger) error
	Run()
	Shutdown() error
}

type Application struct {
	repositories []Repository
	usecases     []Usecase
	controllers  []Controller
	logger       *slog.Logger
}

func New(logger *slog.Logger, repos []Repository, ucs []Usecase, ctrls []Controller) *Application {
	return &Application{
		repositories: repos,
		usecases:     ucs,
		controllers:  ctrls,
		logger:       logger,
	}
}

func (a *Application) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	for _, repo := range a.repositories {
		if err := repo.Init(ctx, a.logger); err != nil {
			return fmt.Errorf("repository init: %w", err)
		}
	}

	for _, uc := range a.usecases {
		if err := uc.Init(ctx, a.logger); err != nil {
			return fmt.Errorf("usecase init: %w", err)
		}
	}

	for _, ctrl := range a.controllers {
		if err := ctrl.Init(ctx, stop, a.logger); err != nil {
			return fmt.Errorf("controller init: %w", err)
		}
	}

	for _, ctrl := range a.controllers {
		go ctrl.Run()
	}

	a.logger.Info("application started")

	<-ctx.Done()
	a.logger.Info("shutting down")

	for _, ctrl := range a.controllers {
		if err := ctrl.Shutdown(); err != nil {
			a.logger.Error("controller shutdown error", "error", err)
		}
	}

	for _, repo := range a.repositories {
		if err := repo.Close(); err != nil {
			a.logger.Error("repository close error", "error", err)
		}
	}

	a.logger.Info("application stopped")
	return nil
}
