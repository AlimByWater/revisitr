package autoaction

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"revisitr/internal/entity"
)

// --- mocks ---

type mockBotClientsRepo struct {
	getByBotIDFn func(ctx context.Context, botID, limit, offset int) ([]entity.BotClient, int, error)
}

func (m *mockBotClientsRepo) GetByBotID(ctx context.Context, botID, limit, offset int) ([]entity.BotClient, int, error) {
	return m.getByBotIDFn(ctx, botID, limit, offset)
}

// --- helpers ---

func schedulerLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func makeExecutor(repo scenariosRepo) *ActionExecutor {
	return NewActionExecutor(repo, schedulerLogger())
}

func birthdayClient(id, botID int, month time.Month, day int) entity.BotClient {
	bd := time.Date(1990, month, day, 0, 0, 0, 0, time.Local)
	return entity.BotClient{
		ID:           id,
		BotID:        botID,
		BirthDate:    &bd,
		RegisteredAt: time.Now().AddDate(-1, 0, 0),
	}
}

func clientNoBirthday(id, botID int) entity.BotClient {
	return entity.BotClient{
		ID:           id,
		BotID:        botID,
		RegisteredAt: time.Now().AddDate(-1, 0, 0),
	}
}

// --- Evaluate ---

func TestEvaluate_Success(t *testing.T) {
	scenariosProcessed := 0
	scenRepo := &mockScenariosRepo{
		getActiveDateBasedFn: func(_ context.Context) ([]entity.AutoScenario, error) {
			return []entity.AutoScenario{
				{ID: 1, BotID: 1, TriggerType: "birthday"},
				{ID: 2, BotID: 1, TriggerType: "holiday"},
			}, nil
		},
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			scenariosProcessed++
			return true, nil // dedup = skip execution, but still counts as processed
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	err := scheduler.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluate_Empty(t *testing.T) {
	scenRepo := &mockScenariosRepo{
		getActiveDateBasedFn: func(_ context.Context) ([]entity.AutoScenario, error) {
			return nil, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, nil
		},
	}, schedulerLogger())

	err := scheduler.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluate_RepoError(t *testing.T) {
	scenRepo := &mockScenariosRepo{
		getActiveDateBasedFn: func(_ context.Context) ([]entity.AutoScenario, error) {
			return nil, errors.New("db error")
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, nil
		},
	}, schedulerLogger())

	err := scheduler.Evaluate(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- evaluateBirthday ---

func TestEvaluateBirthday_MatchToday(t *testing.T) {
	now := time.Now()
	executed := false

	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			executed = true
			return nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				birthdayClient(1, 1, now.Month(), now.Day()),
			}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{
		ID:          1,
		BotID:       1,
		TriggerType: "birthday",
		Timing:      entity.ActionTiming{},
		Message:     "Happy Birthday!",
	}
	scheduler.evaluateBirthday(context.Background(), scenario)

	if !executed {
		t.Error("birthday client matching today should be executed")
	}
}

func TestEvaluateBirthday_WithWindow(t *testing.T) {
	now := time.Now()
	// Client's birthday is 2 days from now
	futureDate := now.AddDate(0, 0, 2)
	executedCount := 0

	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			executedCount++
			return nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				birthdayClient(1, 1, futureDate.Month(), futureDate.Day()),
			}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	// daysBefore = 3 means we should trigger (birthday in 2 days, window starts 3 days before)
	scenario := entity.AutoScenario{
		ID:          1,
		BotID:       1,
		TriggerType: "birthday",
		Timing:      entity.ActionTiming{DaysBefore: intPtr(3)},
		Message:     "Happy Birthday!",
	}
	scheduler.evaluateBirthday(context.Background(), scenario)

	if executedCount != 1 {
		t.Errorf("got %d executions, want 1 (birthday within window)", executedCount)
	}
}

func TestEvaluateBirthday_NoBirthdate(t *testing.T) {
	executed := false
	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			executed = true
			return false, nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{clientNoBirthday(1, 1)}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{ID: 1, BotID: 1, TriggerType: "birthday"}
	scheduler.evaluateBirthday(context.Background(), scenario)

	if executed {
		t.Error("client without birthdate should be skipped")
	}
}

func TestEvaluateBirthday_NoMatch(t *testing.T) {
	now := time.Now()
	// Birthday 30 days from now, no window
	futureDate := now.AddDate(0, 0, 30)
	executed := false

	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			executed = true
			return false, nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				birthdayClient(1, 1, futureDate.Month(), futureDate.Day()),
			}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{ID: 1, BotID: 1, TriggerType: "birthday"}
	scheduler.evaluateBirthday(context.Background(), scenario)

	if executed {
		t.Error("birthday 30 days away should not trigger without window")
	}
}

// --- evaluateHoliday ---

func TestEvaluateHoliday_Match(t *testing.T) {
	now := time.Now()
	executed := false

	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			executed = true
			return nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{testClient(1, 1)}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{
		ID:          1,
		BotID:       1,
		TriggerType: "holiday",
		Timing: entity.ActionTiming{
			Month: intPtr(int(now.Month())),
			Day:   intPtr(now.Day()),
		},
		Message: "Happy holiday!",
	}
	scheduler.evaluateHoliday(context.Background(), scenario)

	if !executed {
		t.Error("holiday matching today should trigger")
	}
}

func TestEvaluateHoliday_NoMatch(t *testing.T) {
	executed := false
	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			executed = true
			return false, nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{testClient(1, 1)}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{
		ID:          1,
		BotID:       1,
		TriggerType: "holiday",
		Timing: entity.ActionTiming{
			Month: intPtr(1),
			Day:   intPtr(1), // Jan 1 — only matches on that day
		},
	}
	// Only skip if today is not Jan 1
	now := time.Now()
	if int(now.Month()) == 1 && now.Day() == 1 {
		t.Skip("today is Jan 1, test not applicable")
	}

	scheduler.evaluateHoliday(context.Background(), scenario)

	if executed {
		t.Error("holiday not matching today should not trigger")
	}
}

func TestEvaluateHoliday_MissingConfig(t *testing.T) {
	executed := false
	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			executed = true
			return false, nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{testClient(1, 1)}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	// Missing month and day
	scenario := entity.AutoScenario{
		ID:          1,
		BotID:       1,
		TriggerType: "holiday",
		Timing:      entity.ActionTiming{},
	}
	scheduler.evaluateHoliday(context.Background(), scenario)

	if executed {
		t.Error("holiday with missing config should not trigger")
	}
}

// --- evaluateInactivity ---

func TestEvaluateInactivity_InactiveClient(t *testing.T) {
	executed := false
	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			return false, nil
		},
		createDedupFn: func(_ context.Context, _, _ int, _ string) error {
			executed = true
			return nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				{ID: 1, BotID: 1, RegisteredAt: time.Now().AddDate(0, 0, -60)},
			}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{
		ID:            1,
		BotID:         1,
		TriggerType:   "inactive_days",
		TriggerConfig: entity.TriggerConfig{Days: intPtr(30)},
		Message:       "We miss you!",
	}
	scheduler.evaluateInactivity(context.Background(), scenario)

	if !executed {
		t.Error("client registered 60 days ago should trigger for 30-day inactivity")
	}
}

func TestEvaluateInactivity_ActiveClient(t *testing.T) {
	executed := false
	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			executed = true
			return false, nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				{ID: 1, BotID: 1, RegisteredAt: time.Now().AddDate(0, 0, -5)},
			}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{
		ID:            1,
		BotID:         1,
		TriggerType:   "inactive_days",
		TriggerConfig: entity.TriggerConfig{Days: intPtr(30)},
	}
	scheduler.evaluateInactivity(context.Background(), scenario)

	if executed {
		t.Error("recently registered client should not trigger inactivity")
	}
}

func TestEvaluateInactivity_MissingDays(t *testing.T) {
	executed := false
	scenRepo := &mockScenariosRepo{
		checkDedupFn: func(_ context.Context, _, _ int, _ string) (bool, error) {
			executed = true
			return false, nil
		},
	}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return []entity.BotClient{
				{ID: 1, BotID: 1, RegisteredAt: time.Now().AddDate(-1, 0, 0)},
			}, 1, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{
		ID:            1,
		BotID:         1,
		TriggerType:   "inactive_days",
		TriggerConfig: entity.TriggerConfig{},
	}
	scheduler.evaluateInactivity(context.Background(), scenario)

	if executed {
		t.Error("missing days config should return early")
	}
}

// --- getAllClients ---

func TestGetAllClients_Pagination(t *testing.T) {
	callCount := 0
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, limit, offset int) ([]entity.BotClient, int, error) {
			callCount++
			total := 250
			remaining := total - offset
			if remaining <= 0 {
				return nil, total, nil
			}
			count := limit
			if count > remaining {
				count = remaining
			}
			batch := make([]entity.BotClient, count)
			for i := range batch {
				batch[i] = entity.BotClient{ID: offset + i + 1, BotID: 1}
			}
			return batch, total, nil
		},
	}

	scenRepo := &mockScenariosRepo{}
	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	clients := scheduler.getAllClients(context.Background(), 1)
	if len(clients) != 250 {
		t.Errorf("got %d clients, want 250", len(clients))
	}
	if callCount != 3 {
		t.Errorf("got %d batches, want 3 (100+100+50)", callCount)
	}
}

func TestGetAllClients_Error(t *testing.T) {
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, errors.New("db error")
		},
	}

	scenRepo := &mockScenariosRepo{}
	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	clients := scheduler.getAllClients(context.Background(), 1)
	if len(clients) != 0 {
		t.Errorf("got %d clients, want 0 on error", len(clients))
	}
}

// --- evaluateScenario routing ---

func TestEvaluateScenario_UnknownType(t *testing.T) {
	// Should not panic on unknown trigger type
	scenRepo := &mockScenariosRepo{}
	clientsRepo := &mockBotClientsRepo{
		getByBotIDFn: func(_ context.Context, _, _, _ int) ([]entity.BotClient, int, error) {
			return nil, 0, nil
		},
	}

	scheduler := NewAutoActionScheduler(scenRepo, makeExecutor(scenRepo), clientsRepo, schedulerLogger())

	scenario := entity.AutoScenario{
		ID:          1,
		BotID:       1,
		TriggerType: "completely_unknown",
	}

	// Should not panic
	scheduler.evaluateScenario(context.Background(), scenario)
}
