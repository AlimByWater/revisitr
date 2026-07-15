package posplugin

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"revisitr/internal/entity"
	"revisitr/internal/service/poscode"
	loyaltyUC "revisitr/internal/usecase/loyalty"
)

// --- mocks (struct with function fields, established pattern) ---

type mockLoyalty struct {
	getBalance     func(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
	getProgram     func(ctx context.Context, id, orgID int) (*entity.LoyaltyProgram, error)
	getPrograms    func(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
	spendPoints    func(ctx context.Context, clientID, programID int, amount float64, desc string) (*entity.ClientLoyalty, error)
	earnFromCheck  func(ctx context.Context, clientID, programID int, checkAmount float64) (*entity.ClientLoyalty, error)
	calculateBonus func(ctx context.Context, clientID, programID int, checkAmount float64) (float64, error)
}

func (m *mockLoyalty) GetBalance(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error) {
	return m.getBalance(ctx, clientID, programID)
}
func (m *mockLoyalty) GetProgram(ctx context.Context, id, orgID int) (*entity.LoyaltyProgram, error) {
	return m.getProgram(ctx, id, orgID)
}
func (m *mockLoyalty) GetPrograms(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error) {
	return m.getPrograms(ctx, orgID)
}
func (m *mockLoyalty) SpendPoints(ctx context.Context, clientID, programID int, amount float64, desc string) (*entity.ClientLoyalty, error) {
	return m.spendPoints(ctx, clientID, programID, amount, desc)
}
func (m *mockLoyalty) EarnFromCheck(ctx context.Context, clientID, programID int, checkAmount float64) (*entity.ClientLoyalty, error) {
	return m.earnFromCheck(ctx, clientID, programID, checkAmount)
}
func (m *mockLoyalty) CalculateBonus(ctx context.Context, clientID, programID int, checkAmount float64) (float64, error) {
	return m.calculateBonus(ctx, clientID, programID, checkAmount)
}

type mockClients struct {
	getByID func(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error)
}

func (m *mockClients) GetByID(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error) {
	return m.getByID(ctx, orgID, clientID)
}

type mockKeys struct {
	create           func(ctx context.Context, k *entity.PluginKey) error
	getActiveByHash  func(ctx context.Context, hash string) (*entity.PluginKey, error)
	listByIntegration func(ctx context.Context, integrationID int) ([]entity.PluginKey, error)
	touchLastUsed    func(ctx context.Context, id int) error
	revoke           func(ctx context.Context, id, orgID int) error
}

func (m *mockKeys) Create(ctx context.Context, k *entity.PluginKey) error {
	return m.create(ctx, k)
}
func (m *mockKeys) GetActiveByHash(ctx context.Context, hash string) (*entity.PluginKey, error) {
	return m.getActiveByHash(ctx, hash)
}
func (m *mockKeys) ListByIntegration(ctx context.Context, integrationID int) ([]entity.PluginKey, error) {
	return m.listByIntegration(ctx, integrationID)
}
func (m *mockKeys) TouchLastUsed(ctx context.Context, id int) error {
	return m.touchLastUsed(ctx, id)
}
func (m *mockKeys) Revoke(ctx context.Context, id, orgID int) error {
	return m.revoke(ctx, id, orgID)
}

// mockOps keeps operations in-memory so idempotency behaves like the real repo:
// Get returns a stored op or an error, Insert records it.
type mockOps struct {
	store     map[string]*entity.PluginOperation
	insertErr error
}

func newMockOps() *mockOps { return &mockOps{store: map[string]*entity.PluginOperation{}} }

func opKey(integrationID int, extOrderID, opType string) string {
	return extOrderID + "|" + opType
}

func (m *mockOps) Get(ctx context.Context, integrationID int, extOrderID, opType string) (*entity.PluginOperation, error) {
	if op, ok := m.store[opKey(integrationID, extOrderID, opType)]; ok {
		return op, nil
	}
	return nil, errors.New("not found")
}
func (m *mockOps) Insert(ctx context.Context, op *entity.PluginOperation) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.store[opKey(op.IntegrationID, op.ExternalOrderID, op.OpType)] = op
	return nil
}

type mockIntegrations struct {
	getByID func(ctx context.Context, id int) (*entity.Integration, error)
}

func (m *mockIntegrations) GetByID(ctx context.Context, id int) (*entity.Integration, error) {
	return m.getByID(ctx, id)
}

// mockCode consumes codes via a hook and stores sessions in-memory so the
// Identify -> Redeem/Accrue flow round-trips real payloads.
type mockCode struct {
	consume  func(ctx context.Context, word string) (poscode.Grant, error)
	allow    func(ctx context.Context, scope string, limit int, window time.Duration) (bool, error)
	sessions map[string][]byte
	seq      int
}

func newMockCode() *mockCode { return &mockCode{sessions: map[string][]byte{}} }

func (m *mockCode) Consume(ctx context.Context, word string) (poscode.Grant, error) {
	return m.consume(ctx, word)
}
func (m *mockCode) AllowAttempt(ctx context.Context, scope string, limit int, window time.Duration) (bool, error) {
	if m.allow != nil {
		return m.allow(ctx, scope, limit, window)
	}
	return true, nil
}
func (m *mockCode) CreateSession(ctx context.Context, payload []byte) (string, error) {
	m.seq++
	tok := "sess-" + itoa(m.seq)
	buf := make([]byte, len(payload))
	copy(buf, payload)
	m.sessions[tok] = buf
	return tok, nil
}
func (m *mockCode) GetSession(ctx context.Context, token string) ([]byte, error) {
	if b, ok := m.sessions[token]; ok {
		return b, nil
	}
	return nil, poscode.ErrNotFound
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// --- helpers ---

func newTestUC(t *testing.T, loyalty *mockLoyalty, clients *mockClients, keys *mockKeys, ops *mockOps, integrations *mockIntegrations, code *mockCode) *Usecase {
	t.Helper()
	uc := New(loyalty, clients, keys, ops, integrations, code)
	if err := uc.Init(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return uc
}

func testKey() *entity.PluginKey {
	return &entity.PluginKey{ID: 1, OrgID: 10, IntegrationID: 5}
}

// program with a configurable max redeem cap.
func testProgram(maxPct float64) *entity.LoyaltyProgram {
	return &entity.LoyaltyProgram{
		ID:    2,
		OrgID: 10,
		Type:  "bonus",
		Config: entity.ProgramConfig{
			CurrencyName:     "баллов",
			MaxRedeemPercent: maxPct,
		},
	}
}

// --- Identify ---

func TestIdentify_HappyPath_CapsAvailableByPercent(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 800}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(50), nil // 50% of 1000 = 500 cap
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		p := &entity.ClientProfile{}
		p.FirstName = "Иван"
		return p, nil
	}}
	code := newMockCode()
	code.consume = func(_ context.Context, _ string) (poscode.Grant, error) {
		return poscode.Grant{ClientID: 100, ProgramID: 2, OrgID: 10}, nil
	}

	uc := newTestUC(t, loyalty, clients, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	res, err := uc.Identify(context.Background(), testKey(), "трасса", 1000)
	if err != nil {
		t.Fatalf("Identify: %v", err)
	}
	if res.Session == "" {
		t.Error("expected a session token")
	}
	if res.Client.Balance != 800 {
		t.Errorf("balance = %v, want 800", res.Client.Balance)
	}
	if res.Client.AvailableToRedeem != 500 {
		t.Errorf("available = %v, want 500 (capped at 50%% of 1000)", res.Client.AvailableToRedeem)
	}
	if res.Client.Name != "Иван" {
		t.Errorf("name = %q, want Иван", res.Client.Name)
	}
}

func TestIdentify_NoCap_DefaultsTo100Percent(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 300}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(0), nil // 0 => treated as 100%
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return nil, errors.New("no profile") // name falls back to "Гость"
	}}
	code := newMockCode()
	code.consume = func(_ context.Context, _ string) (poscode.Grant, error) {
		return poscode.Grant{ClientID: 100, ProgramID: 2, OrgID: 10}, nil
	}

	uc := newTestUC(t, loyalty, clients, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	res, err := uc.Identify(context.Background(), testKey(), "трасса", 1000)
	if err != nil {
		t.Fatalf("Identify: %v", err)
	}
	// balance 300 < 100% of 1000, so available = balance.
	if res.Client.AvailableToRedeem != 300 {
		t.Errorf("available = %v, want 300", res.Client.AvailableToRedeem)
	}
	if res.Client.Name != "Гость" {
		t.Errorf("name = %q, want Гость", res.Client.Name)
	}
}

func TestIdentify_RateLimited(t *testing.T) {
	code := newMockCode()
	code.allow = func(_ context.Context, _ string, _ int, _ time.Duration) (bool, error) {
		return false, nil
	}
	uc := newTestUC(t, &mockLoyalty{}, &mockClients{}, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	_, err := uc.Identify(context.Background(), testKey(), "трасса", 1000)
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("err = %v, want ErrRateLimited", err)
	}
}

func TestIdentify_CodeNotFound(t *testing.T) {
	code := newMockCode()
	code.consume = func(_ context.Context, _ string) (poscode.Grant, error) {
		return poscode.Grant{}, poscode.ErrNotFound
	}
	uc := newTestUC(t, &mockLoyalty{}, &mockClients{}, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	_, err := uc.Identify(context.Background(), testKey(), "нетслова", 1000)
	if !errors.Is(err, ErrGuestNotFound) {
		t.Errorf("err = %v, want ErrGuestNotFound", err)
	}
}

func TestIdentify_OrgMismatch(t *testing.T) {
	code := newMockCode()
	code.consume = func(_ context.Context, _ string) (poscode.Grant, error) {
		return poscode.Grant{ClientID: 100, ProgramID: 2, OrgID: 999}, nil // different org
	}
	uc := newTestUC(t, &mockLoyalty{}, &mockClients{}, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	_, err := uc.Identify(context.Background(), testKey(), "трасса", 1000)
	if !errors.Is(err, ErrGuestNotFound) {
		t.Errorf("err = %v, want ErrGuestNotFound on org mismatch", err)
	}
}

// --- flow helper: run a real Identify to get a valid session ---

func identifiedSession(t *testing.T, uc *Usecase, code *mockCode, balance, maxPct, orderTotal float64) string {
	t.Helper()
	code.consume = func(_ context.Context, _ string) (poscode.Grant, error) {
		return poscode.Grant{ClientID: 100, ProgramID: 2, OrgID: 10}, nil
	}
	res, err := uc.Identify(context.Background(), testKey(), "трасса", orderTotal)
	if err != nil {
		t.Fatalf("Identify (setup): %v", err)
	}
	return res.Session
}

// --- Redeem ---

func TestRedeem_HappyPath(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 800}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(100), nil
		},
		spendPoints: func(_ context.Context, _, _ int, amount float64, _ string) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 800 - amount}, nil
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	ops := newMockOps()
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, ops, &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 800, 100, 1000)

	res, err := uc.Redeem(context.Background(), testKey(), session, "order-1", 200)
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}
	if !res.OK || res.BalanceAfter != 600 {
		t.Errorf("res = %+v, want OK balance 600", res)
	}
	if _, err := ops.Get(context.Background(), 5, "order-1", "redeem"); err != nil {
		t.Error("expected redeem op to be recorded")
	}
}

func TestRedeem_Idempotent(t *testing.T) {
	spendCalls := 0
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 800}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(100), nil
		},
		spendPoints: func(_ context.Context, _, _ int, amount float64, _ string) (*entity.ClientLoyalty, error) {
			spendCalls++
			return &entity.ClientLoyalty{Balance: 800 - amount}, nil
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	ops := newMockOps()
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, ops, &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 800, 100, 1000)

	first, err := uc.Redeem(context.Background(), testKey(), session, "order-1", 200)
	if err != nil {
		t.Fatalf("Redeem 1: %v", err)
	}
	second, err := uc.Redeem(context.Background(), testKey(), session, "order-1", 200)
	if err != nil {
		t.Fatalf("Redeem 2: %v", err)
	}

	if spendCalls != 1 {
		t.Errorf("SpendPoints called %d times, want 1 (idempotent)", spendCalls)
	}
	if first.BalanceAfter != second.BalanceAfter {
		t.Errorf("balances differ: %v vs %v", first.BalanceAfter, second.BalanceAfter)
	}
}

func TestRedeem_InvalidAmount(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 800}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(50), nil // available capped to 500
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 800, 50, 1000)

	for _, amount := range []float64{0, -50, 600} { // 600 > 500 available
		if _, err := uc.Redeem(context.Background(), testKey(), session, "order-x", amount); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("amount %v: err = %v, want ErrInvalidAmount", amount, err)
		}
	}
}

func TestRedeem_Insufficient(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 800}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(100), nil
		},
		spendPoints: func(_ context.Context, _, _ int, _ float64, _ string) (*entity.ClientLoyalty, error) {
			return nil, loyaltyUC.ErrInsufficientPoints
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 800, 100, 1000)

	_, err := uc.Redeem(context.Background(), testKey(), session, "order-1", 200)
	if !errors.Is(err, ErrInsufficient) {
		t.Errorf("err = %v, want ErrInsufficient", err)
	}
}

func TestRedeem_SessionInvalid(t *testing.T) {
	uc := newTestUC(t, &mockLoyalty{}, &mockClients{}, &mockKeys{}, newMockOps(), &mockIntegrations{}, newMockCode())

	_, err := uc.Redeem(context.Background(), testKey(), "bogus-session", "order-1", 100)
	if !errors.Is(err, ErrSessionInvalid) {
		t.Errorf("err = %v, want ErrSessionInvalid", err)
	}
}

func TestRedeem_SessionOrgMismatch(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 800}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(100), nil
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 800, 100, 1000)

	// A different till (other integration) must not reuse this session.
	otherKey := &entity.PluginKey{ID: 2, OrgID: 10, IntegrationID: 999}
	_, err := uc.Redeem(context.Background(), otherKey, session, "order-1", 100)
	if !errors.Is(err, ErrSessionInvalid) {
		t.Errorf("err = %v, want ErrSessionInvalid on integration mismatch", err)
	}
}

// --- Accrue ---

func TestAccrue_HappyPath(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 100}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(100), nil
		},
		calculateBonus: func(_ context.Context, _, _ int, checkAmount float64) (float64, error) {
			return checkAmount * 0.1, nil // 10%
		},
		earnFromCheck: func(_ context.Context, _, _ int, checkAmount float64) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 100 + checkAmount*0.1}, nil
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	ops := newMockOps()
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, ops, &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 100, 100, 1000)

	res, err := uc.Accrue(context.Background(), testKey(), session, "order-1", 800)
	if err != nil {
		t.Fatalf("Accrue: %v", err)
	}
	if res.Accrued != 80 {
		t.Errorf("accrued = %v, want 80", res.Accrued)
	}
	if res.BalanceAfter != 180 {
		t.Errorf("balance_after = %v, want 180", res.BalanceAfter)
	}
}

func TestAccrue_Idempotent(t *testing.T) {
	earnCalls := 0
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 100}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(100), nil
		},
		calculateBonus: func(_ context.Context, _, _ int, checkAmount float64) (float64, error) {
			return checkAmount * 0.1, nil
		},
		earnFromCheck: func(_ context.Context, _, _ int, checkAmount float64) (*entity.ClientLoyalty, error) {
			earnCalls++
			return &entity.ClientLoyalty{Balance: 100 + checkAmount*0.1}, nil
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	ops := newMockOps()
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, ops, &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 100, 100, 1000)

	first, _ := uc.Accrue(context.Background(), testKey(), session, "order-1", 800)
	second, err := uc.Accrue(context.Background(), testKey(), session, "order-1", 800)
	if err != nil {
		t.Fatalf("Accrue 2: %v", err)
	}
	if earnCalls != 1 {
		t.Errorf("EarnFromCheck called %d times, want 1", earnCalls)
	}
	if first.Accrued != second.Accrued || first.BalanceAfter != second.BalanceAfter {
		t.Errorf("results differ: %+v vs %+v", first, second)
	}
}

func TestAccrue_InvalidAmount(t *testing.T) {
	loyalty := &mockLoyalty{
		getBalance: func(_ context.Context, _, _ int) (*entity.ClientLoyalty, error) {
			return &entity.ClientLoyalty{Balance: 100}, nil
		},
		getProgram: func(_ context.Context, _, _ int) (*entity.LoyaltyProgram, error) {
			return testProgram(100), nil
		},
	}
	clients := &mockClients{getByID: func(_ context.Context, _, _ int) (*entity.ClientProfile, error) {
		return &entity.ClientProfile{}, nil
	}}
	code := newMockCode()
	uc := newTestUC(t, loyalty, clients, &mockKeys{}, newMockOps(), &mockIntegrations{}, code)

	session := identifiedSession(t, uc, code, 100, 100, 1000)

	if _, err := uc.Accrue(context.Background(), testKey(), session, "order-1", 0); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("err = %v, want ErrInvalidAmount", err)
	}
}

// --- AuthenticateKey ---

func TestAuthenticateKey_OK(t *testing.T) {
	want := testKey()
	touched := false
	keys := &mockKeys{
		getActiveByHash: func(_ context.Context, hash string) (*entity.PluginKey, error) {
			if hash != hashKey("rvk_secret") {
				t.Errorf("hash = %q, want hash of rvk_secret", hash)
			}
			return want, nil
		},
		touchLastUsed: func(_ context.Context, id int) error {
			touched = true
			return nil
		},
	}
	uc := newTestUC(t, &mockLoyalty{}, &mockClients{}, keys, newMockOps(), &mockIntegrations{}, newMockCode())

	got, err := uc.AuthenticateKey(context.Background(), "rvk_secret")
	if err != nil {
		t.Fatalf("AuthenticateKey: %v", err)
	}
	if got != want {
		t.Error("returned key mismatch")
	}
	if !touched {
		t.Error("expected TouchLastUsed to be called")
	}
}

func TestAuthenticateKey_Unauthorized(t *testing.T) {
	keys := &mockKeys{
		getActiveByHash: func(_ context.Context, _ string) (*entity.PluginKey, error) {
			return nil, errors.New("no rows")
		},
	}
	uc := newTestUC(t, &mockLoyalty{}, &mockClients{}, keys, newMockOps(), &mockIntegrations{}, newMockCode())

	_, err := uc.AuthenticateKey(context.Background(), "rvk_bad")
	if !errors.Is(err, ErrUnauthorizedKey) {
		t.Errorf("err = %v, want ErrUnauthorizedKey", err)
	}
}

// --- Config ---

func TestConfig_PicksActiveProgramAndBaseAccrual(t *testing.T) {
	active := entity.LoyaltyProgram{
		ID: 2, Type: "bonus", IsActive: true,
		Config: entity.ProgramConfig{CurrencyName: "баллов", MaxRedeemPercent: 40, SumWithDiscounts: true},
		Levels: []entity.LoyaltyLevel{
			{RewardPercent: 15, SortOrder: 2},
			{RewardPercent: 5, SortOrder: 0}, // base level (lowest SortOrder)
		},
	}
	inactive := entity.LoyaltyProgram{ID: 1, IsActive: false}
	loyalty := &mockLoyalty{
		getPrograms: func(_ context.Context, _ int) ([]entity.LoyaltyProgram, error) {
			return []entity.LoyaltyProgram{inactive, active}, nil
		},
	}
	uc := newTestUC(t, loyalty, &mockClients{}, &mockKeys{}, newMockOps(), &mockIntegrations{}, newMockCode())

	res, err := uc.Config(context.Background(), testKey())
	if err != nil {
		t.Fatalf("Config: %v", err)
	}
	if res.MaxRedeemPercent != 40 {
		t.Errorf("max redeem = %v, want 40", res.MaxRedeemPercent)
	}
	if res.AccrualPercent != 5 {
		t.Errorf("accrual = %v, want 5 (base level)", res.AccrualPercent)
	}
	if !res.SumWithIikoDiscounts {
		t.Error("expected SumWithIikoDiscounts true")
	}
	if res.Currency != "баллов" {
		t.Errorf("currency = %q, want баллов", res.Currency)
	}
}

func TestConfig_NoPrograms(t *testing.T) {
	loyalty := &mockLoyalty{
		getPrograms: func(_ context.Context, _ int) ([]entity.LoyaltyProgram, error) {
			return nil, nil
		},
	}
	uc := newTestUC(t, loyalty, &mockClients{}, &mockKeys{}, newMockOps(), &mockIntegrations{}, newMockCode())

	if _, err := uc.Config(context.Background(), testKey()); !errors.Is(err, ErrGuestNotFound) {
		t.Errorf("err = %v, want ErrGuestNotFound", err)
	}
}
