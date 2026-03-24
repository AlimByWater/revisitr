package autoaction

import (
	"context"
	"errors"
	"testing"

	"revisitr/internal/entity"
)

// Note: mockScenariosRepo and mockBotClientsRepo are defined in executor_test.go and scheduler_test.go

// --- OnEvent ---

func TestOnEvent_MatchingScenario(t *testing.T) {
	dedupCreated := false
	scenRepo := &mockScenariosRepo{
		getActiveByTriggerFn: func(_ context.Context, _ string) ([]entity.AutoScenario, error) {
			return []entity.AutoScenario{
				{ID: 1, BotID: 1, TriggerType: "registration", Message: "Welcome!"},
			}, nil
		},
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			dedupCreated = true
			return nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				{ID: 42, BotID: 1},
			}, 1, nil
		},
	}

	hook := NewEventHook(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	err := hook.OnEvent(context.Background(), "registration", 42, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dedupCreated {
		t.Error("action should have been executed for matching client")
	}
}

func TestOnEvent_NoScenarios(t *testing.T) {
	scenRepo := &mockScenariosRepo{
		getActiveByTriggerFn: func(_ context.Context, _ string) ([]entity.AutoScenario, error) {
			return nil, nil
		},
	}

	hook := NewEventHook(scenRepo, makeExecutor(scenRepo), &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, nil
		},
	}, schedulerLogger())

	err := hook.OnEvent(context.Background(), "registration", 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOnEvent_ClientNotFound(t *testing.T) {
	dedupChecked := false
	scenRepo := &mockScenariosRepo{
		getActiveByTriggerFn: func(_ context.Context, _ string) ([]entity.AutoScenario, error) {
			return []entity.AutoScenario{
				{ID: 1, BotID: 1, TriggerType: "registration"},
			}, nil
		},
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			dedupChecked = true
			return false, nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			// Return clients that don't match the requested ID
			return []entity.BotClient{{ID: 99, BotID: 1}}, 1, nil
		},
	}

	hook := NewEventHook(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	err := hook.OnEvent(context.Background(), "registration", 42, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dedupChecked {
		t.Error("should not attempt execution when client not found")
	}
}

func TestOnEvent_MultipleScenarios(t *testing.T) {
	execCount := 0
	scenRepo := &mockScenariosRepo{
		getActiveByTriggerFn: func(_ context.Context, _ string) ([]entity.AutoScenario, error) {
			return []entity.AutoScenario{
				{ID: 1, BotID: 1, TriggerType: "level_change", Message: "Congrats!"},
				{ID: 2, BotID: 1, TriggerType: "level_change", Message: "Bonus!"},
			}, nil
		},
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			execCount++
			return nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{{ID: 1, BotID: 1}}, 1, nil
		},
	}

	hook := NewEventHook(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	err := hook.OnEvent(context.Background(), "level_change", 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if execCount != 2 {
		t.Errorf("got %d executions, want 2 (one per scenario)", execCount)
	}
}

func TestOnEvent_RepoError(t *testing.T) {
	scenRepo := &mockScenariosRepo{
		getActiveByTriggerFn: func(_ context.Context, _ string) ([]entity.AutoScenario, error) {
			return nil, errors.New("db error")
		},
	}

	hook := NewEventHook(scenRepo, makeExecutor(scenRepo), &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, nil
		},
	}, schedulerLogger())

	err := hook.OnEvent(context.Background(), "registration", 1, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOnEvent_ExecuteError(t *testing.T) {
	scenRepo := &mockScenariosRepo{
		getActiveByTriggerFn: func(_ context.Context, _ string) ([]entity.AutoScenario, error) {
			return []entity.AutoScenario{
				{ID: 1, BotID: 1, TriggerType: "registration", Message: "Hi"},
			}, nil
		},
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, errors.New("dedup check failed")
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{{ID: 1, BotID: 1}}, 1, nil
		},
	}

	hook := NewEventHook(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	// Should not return error — execution errors are logged, not propagated
	err := hook.OnEvent(context.Background(), "registration", 1, nil)
	if err != nil {
		t.Fatalf("execution error should be logged, not propagated: %v", err)
	}
}

// --- findClient ---

func TestFindClient_FoundFirstBatch(t *testing.T) {
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				{ID: 1, BotID: 1},
				{ID: 2, BotID: 1},
				{ID: 42, BotID: 1},
			}, 3, nil
		},
	}

	hook := NewEventHook(&mockScenariosRepo{}, makeExecutor(&mockScenariosRepo{}), clientsRepo, schedulerLogger())

	client := hook.findClient(context.Background(), 1, 42)
	if client == nil {
		t.Fatal("expected to find client 42")
	}
	if client.ID != 42 {
		t.Errorf("got client ID %d, want 42", client.ID)
	}
}

func TestFindClient_NotFound(t *testing.T) {
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{{ID: 1, BotID: 1}}, 1, nil
		},
	}

	hook := NewEventHook(&mockScenariosRepo{}, makeExecutor(&mockScenariosRepo{}), clientsRepo, schedulerLogger())

	client := hook.findClient(context.Background(), 1, 999)
	if client != nil {
		t.Errorf("expected nil, got client ID %d", client.ID)
	}
}

func TestFindClient_PaginatedSearch(t *testing.T) {
	callCount := 0
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, limit, offset int) ([]entity.BotClient, int, error) {
			callCount++
			total := 150
			if offset == 0 {
				batch := make([]entity.BotClient, limit)
				for i := range batch {
					batch[i] = entity.BotClient{ID: i + 1, BotID: 1}
				}
				return batch, total, nil
			}
			// Second batch: client 101-150, target is 120
			remaining := total - offset
			batch := make([]entity.BotClient, remaining)
			for i := range batch {
				batch[i] = entity.BotClient{ID: offset + i + 1, BotID: 1}
			}
			return batch, total, nil
		},
	}

	hook := NewEventHook(&mockScenariosRepo{}, makeExecutor(&mockScenariosRepo{}), clientsRepo, schedulerLogger())

	client := hook.findClient(context.Background(), 1, 120)
	if client == nil {
		t.Fatal("expected to find client 120 in second batch")
	}
	if client.ID != 120 {
		t.Errorf("got client ID %d, want 120", client.ID)
	}
	if callCount != 2 {
		t.Errorf("got %d batch calls, want 2", callCount)
	}
}

func TestFindClient_RepoError(t *testing.T) {
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, errors.New("db error")
		},
	}

	hook := NewEventHook(&mockScenariosRepo{}, makeExecutor(&mockScenariosRepo{}), clientsRepo, schedulerLogger())

	client := hook.findClient(context.Background(), 1, 42)
	if client != nil {
		t.Errorf("expected nil on error, got client ID %d", client.ID)
	}
}
