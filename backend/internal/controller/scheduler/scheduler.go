package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Task represents a scheduled background task.
type Task struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
}

// Scheduler runs registered tasks on fixed intervals.
// It lives in the controller layer as an input adapter (like HTTP handlers):
// it receives timer ticks and calls usecases/services with no business logic.
type Scheduler struct {
	tasks  []Task
	logger *slog.Logger
}

func New(logger *slog.Logger) *Scheduler {
	return &Scheduler{logger: logger}
}

func (s *Scheduler) Register(task Task) {
	s.tasks = append(s.tasks, task)
}

// Run starts all registered tasks in separate goroutines and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	for _, task := range s.tasks {
		go s.runTask(ctx, task)
	}
	<-ctx.Done()
}

func (s *Scheduler) runTask(ctx context.Context, task Task) {
	if s.logger != nil {
		s.logger.Info("scheduler: task started", "task", task.Name, "interval", task.Interval)
	}

	// Run immediately once on startup
	if err := task.Fn(ctx); err != nil {
		if s.logger != nil {
			s.logger.Error("scheduler: task failed", "task", task.Name, "error", err)
		}
	}

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if s.logger != nil {
				s.logger.Info("scheduler: task stopped", "task", task.Name)
			}
			return
		case <-ticker.C:
			if err := task.Fn(ctx); err != nil {
				if s.logger != nil {
					s.logger.Error("scheduler: task failed", "task", task.Name, "error", err)
				}
			}
		}
	}
}
