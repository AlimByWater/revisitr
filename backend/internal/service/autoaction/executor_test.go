package autoaction

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockScenariosRepo struct {
	createActionLogFn    func(ctx context.Context, log *entity.AutoActionLog) error
	checkDedupFn         func(ctx context.Context, scenarioID, clientID int, key string) (bool, error)
	createDedupFn        func(ctx context.Context, scenarioID, clientID int, key string) error
	getActiveDateBasedFn func(ctx context.Context) ([]entity.AutoScenario, error)
	getActiveByTriggerFn func(ctx context.Context, triggerType string) ([]entity.AutoScenario, error)
}

func (m *mockScenariosRepo) CreateActionLog(ctx context.Context, log *entity.AutoActionLog) error {
	if m.createActionLogFn != nil {
		return m.createActionLogFn(ctx, log)
	}
	return nil
}
func (m *mockScenariosRepo) CheckDedup(ctx context.Context, scenarioID, clientID int, key string) (bool, error) {
	return m.checkDedupFn(ctx, scenarioID, clientID, key)
}
func (m *mockScenariosRepo) CreateDedup(ctx context.Context, scenarioID, clientID int, key string) error {
	if m.createDedupFn != nil {
		return m.createDedupFn(ctx, scenarioID, clientID, key)
	}
	return nil
}
func (m *mockScenariosRepo) GetActiveDateBased(ctx context.Context) ([]entity.AutoScenario, error) {
	if m.getActiveDateBasedFn != nil {
		return m.getActiveDateBasedFn(ctx)
	}
	return nil, nil
}
func (m *mockScenariosRepo) GetActiveByTriggerType(ctx context.Context, triggerType string) ([]entity.AutoScenario, error) {
	if m.getActiveByTriggerFn != nil {
		return m.getActiveByTriggerFn(ctx, triggerType)
	}
	return nil, nil
}

// --- helpers ---

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testScenario(id, botID int) entity.AutoScenario {
	return entity.AutoScenario{
		ID:    id,
		BotID: botID,
		Name:  "Test Scenario",
	}
}

func testClient(id, botID int) entity.BotClient {
	return entity.BotClient{
		ID:    id,
		BotID: botID,
	}
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

// --- Execute ---

func TestExecute_MultipleActions(t *testing.T) {
	logCount := 0
	repo := &mockScenariosRepo{
		createActionLogFn: func(_ context.Context, _ *entity.AutoActionLog) error {
			logCount++
			return nil
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Actions = entity.ActionDefs{
		{Type: "bonus", Amount: intPtr(100)},
		{Type: "campaign", Template: strPtr("Hello!")},
	}

	err := executor.Execute(context.Background(), scenario, testClient(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logCount != 2 {
		t.Errorf("got %d log entries, want 2", logCount)
	}
}

func TestExecute_MessageFallback(t *testing.T) {
	var loggedType string
	repo := &mockScenariosRepo{
		createActionLogFn: func(_ context.Context, log *entity.AutoActionLog) error {
			loggedType = log.ActionType
			return nil
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Message = "Birthday greeting!"
	scenario.Actions = nil

	err := executor.Execute(context.Background(), scenario, testClient(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loggedType != "campaign" {
		t.Errorf("got action type %q, want %q (message fallback)", loggedType, "campaign")
	}
}

func TestExecute_EmptyActionsNoMessage(t *testing.T) {
	logCount := 0
	repo := &mockScenariosRepo{
		createActionLogFn: func(_ context.Context, _ *entity.AutoActionLog) error {
			logCount++
			return nil
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Actions = nil
	scenario.Message = ""

	err := executor.Execute(context.Background(), scenario, testClient(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logCount != 0 {
		t.Errorf("got %d log entries, want 0 for empty scenario", logCount)
	}
}

func TestExecute_UnknownActionType(t *testing.T) {
	var loggedResult string
	var loggedError *string
	repo := &mockScenariosRepo{
		createActionLogFn: func(_ context.Context, log *entity.AutoActionLog) error {
			loggedResult = log.Result
			loggedError = log.ErrorMsg
			return nil
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Actions = entity.ActionDefs{{Type: "unknown_type"}}

	err := executor.Execute(context.Background(), scenario, testClient(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loggedResult != "skipped" {
		t.Errorf("got result %q, want %q", loggedResult, "skipped")
	}
	if loggedError == nil || *loggedError == "" {
		t.Error("expected error message for unknown action type")
	}
}

func TestExecute_AllActionTypes(t *testing.T) {
	types := []string{"bonus", "campaign", "promo_code", "level_change"}
	var loggedTypes []string
	repo := &mockScenariosRepo{
		createActionLogFn: func(_ context.Context, log *entity.AutoActionLog) error {
			loggedTypes = append(loggedTypes, log.ActionType)
			return nil
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	actions := make(entity.ActionDefs, len(types))
	for i, t := range types {
		actions[i] = entity.ActionDef{Type: t}
	}

	scenario := testScenario(1, 1)
	scenario.Actions = actions

	err := executor.Execute(context.Background(), scenario, testClient(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loggedTypes) != 4 {
		t.Fatalf("got %d logged types, want 4", len(loggedTypes))
	}
	for i, want := range types {
		if loggedTypes[i] != want {
			t.Errorf("loggedTypes[%d] = %q, want %q", i, loggedTypes[i], want)
		}
	}
}

func TestExecute_LogCreationError(t *testing.T) {
	repo := &mockScenariosRepo{
		createActionLogFn: func(_ context.Context, _ *entity.AutoActionLog) error {
			return errors.New("db error")
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Actions = entity.ActionDefs{{Type: "bonus", Amount: intPtr(100)}}

	// Should not return error — log errors are only logged, not propagated
	err := executor.Execute(context.Background(), scenario, testClient(1, 1))
	if err != nil {
		t.Fatalf("log error should not propagate: %v", err)
	}
}

// --- ExecuteWithDedup ---

func TestExecuteWithDedup_FirstExecution(t *testing.T) {
	executed := false
	dedupCreated := false
	repo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createActionLogFn: func(_ context.Context, _ *entity.AutoActionLog) error {
			executed = true
			return nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			dedupCreated = true
			return nil
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Actions = entity.ActionDefs{{Type: "bonus", Amount: intPtr(50)}}

	err := executor.ExecuteWithDedup(context.Background(), scenario, testClient(1, 1), "birthday:2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("actions should be executed on first run")
	}
	if !dedupCreated {
		t.Error("dedup record should be created")
	}
}

func TestExecuteWithDedup_Duplicate(t *testing.T) {
	executed := false
	repo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return true, nil
		},
		createActionLogFn: func(_ context.Context, _ *entity.AutoActionLog) error {
			executed = true
			return nil
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Actions = entity.ActionDefs{{Type: "bonus", Amount: intPtr(50)}}

	err := executor.ExecuteWithDedup(context.Background(), scenario, testClient(1, 1), "birthday:2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if executed {
		t.Error("actions should NOT be executed for duplicate")
	}
}

func TestExecuteWithDedup_CheckDedupError(t *testing.T) {
	repo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, errors.New("redis error")
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	err := executor.ExecuteWithDedup(context.Background(), testScenario(1, 1), testClient(1, 1), "key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecuteWithDedup_CreateDedupError(t *testing.T) {
	// CreateDedup error should be logged but not fail the overall execution
	repo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			return errors.New("db error")
		},
	}
	executor := NewActionExecutor(repo, testLogger())

	scenario := testScenario(1, 1)
	scenario.Actions = entity.ActionDefs{{Type: "bonus"}}

	err := executor.ExecuteWithDedup(context.Background(), scenario, testClient(1, 1), "key")
	if err != nil {
		t.Fatalf("createDedup error should not propagate: %v", err)
	}
}
