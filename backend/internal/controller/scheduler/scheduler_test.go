package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_RunsTaskImmediately(t *testing.T) {
	var count int32

	sched := New(nil)
	sched.StartupDelay = 0
	sched.Register(Task{
		Name:     "test",
		Interval: time.Hour, // long interval — only immediate run matters
		Fn: func(_ context.Context) error {
			atomic.AddInt32(&count, 1)
			return nil
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go sched.Run(ctx)
	<-ctx.Done()

	if atomic.LoadInt32(&count) == 0 {
		t.Error("expected task to run at least once immediately")
	}
}

func TestScheduler_LogsErrors(t *testing.T) {
	var called int32

	sched := New(nil)
	sched.StartupDelay = 0
	sched.Register(Task{
		Name:     "failing",
		Interval: time.Hour,
		Fn: func(_ context.Context) error {
			atomic.AddInt32(&called, 1)
			return errors.New("task error")
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go sched.Run(ctx)
	<-ctx.Done()

	if atomic.LoadInt32(&called) == 0 {
		t.Error("expected task to be called even if it returns an error")
	}
}

func TestScheduler_MultipleTasksRun(t *testing.T) {
	var a, b int32

	sched := New(nil)
	sched.StartupDelay = 0
	sched.Register(Task{
		Name:     "task-a",
		Interval: time.Hour,
		Fn: func(_ context.Context) error { atomic.AddInt32(&a, 1); return nil },
	})
	sched.Register(Task{
		Name:     "task-b",
		Interval: time.Hour,
		Fn: func(_ context.Context) error { atomic.AddInt32(&b, 1); return nil },
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go sched.Run(ctx)
	<-ctx.Done()

	if atomic.LoadInt32(&a) == 0 || atomic.LoadInt32(&b) == 0 {
		t.Error("expected both tasks to run")
	}
}
