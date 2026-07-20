package posplugin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"revisitr/internal/entity"
	"revisitr/internal/service/poscode"
	loyaltyUC "revisitr/internal/usecase/loyalty"
)

var (
	ErrUnauthorizedKey = errors.New("unauthorized api key")
	ErrGuestNotFound   = errors.New("guest not found")
	ErrSessionInvalid  = errors.New("session invalid or expired")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrInsufficient    = errors.New("insufficient points")
	ErrRateLimited     = errors.New("rate limited")
)

type loyaltyService interface {
	GetBalance(ctx context.Context, clientID, programID int) (*entity.ClientLoyalty, error)
	GetProgram(ctx context.Context, id, orgID int) (*entity.LoyaltyProgram, error)
	GetPrograms(ctx context.Context, orgID int) ([]entity.LoyaltyProgram, error)
	SpendPoints(ctx context.Context, clientID, programID int, amount float64, description string) (*entity.ClientLoyalty, error)
	EarnFromCheck(ctx context.Context, clientID, programID int, checkAmount float64) (*entity.ClientLoyalty, error)
	CalculateBonus(ctx context.Context, clientID, programID int, checkAmount float64) (float64, error)
}

type clientsRepo interface {
	GetByID(ctx context.Context, orgID, clientID int) (*entity.ClientProfile, error)
}

type keysRepo interface {
	Create(ctx context.Context, k *entity.PluginKey) error
	GetActiveByHash(ctx context.Context, hash string) (*entity.PluginKey, error)
	ListByIntegration(ctx context.Context, integrationID int) ([]entity.PluginKey, error)
	TouchLastUsed(ctx context.Context, id int) error
	Revoke(ctx context.Context, id, orgID int) error
}

type opsRepo interface {
	Get(ctx context.Context, integrationID int, extOrderID, opType string) (*entity.PluginOperation, error)
	Insert(ctx context.Context, op *entity.PluginOperation) error
}

type integrationsRepo interface {
	GetByID(ctx context.Context, id int) (*entity.Integration, error)
	UpsertOrder(ctx context.Context, order *entity.ExternalOrder) error
}

type codeService interface {
	Consume(ctx context.Context, word string) (poscode.Grant, error)
	CreateSession(ctx context.Context, payload []byte) (string, error)
	GetSession(ctx context.Context, token string) ([]byte, error)
	AllowAttempt(ctx context.Context, scope string, limit int, window time.Duration) (bool, error)
}

// notifier delivers a one-off message to a client's chat (via the event bus).
// Optional: when nil, POS operations skip guest notifications.
type notifier interface {
	PublishNotifyClient(ctx context.Context, botID int, chatID int64, text string) error
}

type Usecase struct {
	logger       *slog.Logger
	loyalty      loyaltyService
	clients      clientsRepo
	keys         keysRepo
	ops          opsRepo
	integrations integrationsRepo
	code         codeService
	notify       notifier
}

func New(loyalty loyaltyService, clients clientsRepo, keys keysRepo, ops opsRepo, integrations integrationsRepo, code codeService) *Usecase {
	return &Usecase{
		loyalty:      loyalty,
		clients:      clients,
		keys:         keys,
		ops:          ops,
		integrations: integrations,
		code:         code,
	}
}

func (uc *Usecase) Init(_ context.Context, logger *slog.Logger) error {
	uc.logger = logger
	return nil
}

// SetEventBus wires the event bus used to notify guests after redeem/accrue.
func (uc *Usecase) SetEventBus(n notifier) { uc.notify = n }

// --- result structs (HTTP contract) ---

type ClientInfo struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Balance           float64 `json:"balance"`
	Currency          string  `json:"currency"`
	AvailableToRedeem float64 `json:"available_to_redeem"`
}

type IdentifyResult struct {
	Session string     `json:"session"`
	Client  ClientInfo `json:"client"`
}

type OpResult struct {
	OK           bool    `json:"ok"`
	BalanceAfter float64 `json:"balance_after"`
	Accrued      float64 `json:"accrued,omitempty"`
}

type ConfigResult struct {
	ProgramType          string  `json:"program_type"`
	Currency             string  `json:"currency"`
	MaxRedeemPercent     float64 `json:"max_redeem_percent"`
	AccrualPercent       float64 `json:"accrual_percent"`
	SumWithIikoDiscounts bool    `json:"sum_with_iiko_discounts"`
}

type SubmitOrderRequest struct {
	OrderID       string            `json:"order_id"`
	Source        string            `json:"source"`
	OrderedAt     time.Time         `json:"ordered_at"`
	Total         float64           `json:"total"`
	TableNum      string            `json:"table_num"`
	WaiterName    string            `json:"waiter_name"`
	Items         entity.OrderItems `json:"items"`
	CustomerPhone string            `json:"customer_phone"`
	CustomerName  string            `json:"customer_name"`
}

type sessionData struct {
	ClientID      int     `json:"client_id"`
	ProgramID     int     `json:"program_id"`
	OrgID         int     `json:"org_id"`
	IntegrationID int     `json:"integration_id"`
	OrderTotal    float64 `json:"order_total"`
	Available     float64 `json:"available"`
}

func (uc *Usecase) SubmitOrder(ctx context.Context, key *entity.PluginKey, req SubmitOrderRequest) error {
	if req.OrderID == "" || req.OrderedAt.IsZero() || req.Total < 0 || (req.Source != "hall" && req.Source != "delivery") {
		return fmt.Errorf("invalid order")
	}
	order := &entity.ExternalOrder{
		IntegrationID: key.IntegrationID,
		ExternalID:    req.OrderID,
		Source:        req.Source,
		Items:         req.Items,
		Total:         req.Total,
		OrderedAt:     &req.OrderedAt,
	}
	if req.TableNum != "" {
		order.TableNum = &req.TableNum
	}
	if req.WaiterName != "" {
		order.WaiterName = &req.WaiterName
	}
	if req.CustomerPhone != "" {
		order.CustomerPhone = &req.CustomerPhone
	}
	if req.CustomerName != "" {
		order.CustomerName = &req.CustomerName
	}
	if err := uc.integrations.UpsertOrder(ctx, order); err != nil {
		return fmt.Errorf("upsert order: %w", err)
	}
	return nil
}

// AuthenticateKey hashes the raw key and looks it up. On success it best-effort
// touches last_used_at and returns the key.
func (uc *Usecase) AuthenticateKey(ctx context.Context, rawKey string) (*entity.PluginKey, error) {
	hash := hashKey(rawKey)
	key, err := uc.keys.GetActiveByHash(ctx, hash)
	if err != nil {
		return nil, ErrUnauthorizedKey
	}
	if err := uc.keys.TouchLastUsed(ctx, key.ID); err != nil {
		uc.logger.Warn("touch plugin key last_used", "error", err, "key_id", key.ID)
	}
	return key, nil
}

func (uc *Usecase) Identify(ctx context.Context, key *entity.PluginKey, code string, orderTotal float64) (*IdentifyResult, error) {
	allowed, err := uc.code.AllowAttempt(ctx, fmt.Sprintf("identify:%d", key.IntegrationID), 30, time.Minute)
	if err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}
	if !allowed {
		return nil, ErrRateLimited
	}

	grant, err := uc.code.Consume(ctx, code)
	if err != nil {
		if errors.Is(err, poscode.ErrNotFound) {
			return nil, ErrGuestNotFound
		}
		return nil, fmt.Errorf("consume code: %w", err)
	}
	if grant.OrgID != key.OrgID {
		return nil, ErrGuestNotFound
	}

	bal, err := uc.loyalty.GetBalance(ctx, grant.ClientID, grant.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	prog, err := uc.loyalty.GetProgram(ctx, grant.ProgramID, key.OrgID)
	if err != nil {
		return nil, fmt.Errorf("get program: %w", err)
	}

	name := "Гость"
	if profile, err := uc.clients.GetByID(ctx, key.OrgID, grant.ClientID); err == nil {
		n := strings.TrimSpace(profile.FirstName + " " + profile.LastName)
		if n != "" {
			name = n
		}
	}

	maxPct := prog.Config.MaxRedeemPercent
	if maxPct <= 0 {
		maxPct = 100
	}

	available := bal.Balance
	capAmount := orderTotal * maxPct / 100
	if available > capAmount {
		available = capAmount
	}
	if available < 0 {
		available = 0
	}

	sd := sessionData{
		ClientID:      grant.ClientID,
		ProgramID:     grant.ProgramID,
		OrgID:         key.OrgID,
		IntegrationID: key.IntegrationID,
		OrderTotal:    orderTotal,
		Available:     available,
	}
	payload, err := json.Marshal(sd)
	if err != nil {
		return nil, fmt.Errorf("marshal session: %w", err)
	}
	token, err := uc.code.CreateSession(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	currency := prog.Config.CurrencyName
	if currency == "" {
		currency = "бонусов"
	}

	return &IdentifyResult{
		Session: token,
		Client: ClientInfo{
			ID:                grant.ClientID,
			Name:              name,
			Balance:           bal.Balance,
			Currency:          currency,
			AvailableToRedeem: available,
		},
	}, nil
}

func (uc *Usecase) Redeem(ctx context.Context, key *entity.PluginKey, session, orderID string, amount float64) (*OpResult, error) {
	if existing, err := uc.ops.Get(ctx, key.IntegrationID, orderID, "redeem"); err == nil {
		return &OpResult{OK: true, BalanceAfter: existing.BalanceAfter}, nil
	}

	sd, err := uc.loadSession(ctx, key, session)
	if err != nil {
		return nil, err
	}

	if amount <= 0 || amount > sd.Available {
		return nil, ErrInvalidAmount
	}

	cl, err := uc.loyalty.SpendPoints(ctx, sd.ClientID, sd.ProgramID, amount, "POS: списание бонусов")
	if err != nil {
		if errors.Is(err, loyaltyUC.ErrInsufficientPoints) {
			return nil, ErrInsufficient
		}
		return nil, fmt.Errorf("spend points: %w", err)
	}

	op := &entity.PluginOperation{
		IntegrationID:   key.IntegrationID,
		ExternalOrderID: orderID,
		OpType:          "redeem",
		ClientID:        sd.ClientID,
		ProgramID:       sd.ProgramID,
		Amount:          amount,
		BalanceAfter:    cl.Balance,
	}
	if err := uc.ops.Insert(ctx, op); err != nil {
		// Possible race: another request inserted the same (integration, order, type).
		if existing, getErr := uc.ops.Get(ctx, key.IntegrationID, orderID, "redeem"); getErr == nil {
			return &OpResult{OK: true, BalanceAfter: existing.BalanceAfter}, nil
		}
		return nil, fmt.Errorf("insert operation: %w", err)
	}

	if uc.notify != nil {
		cur := uc.programCurrency(ctx, sd.ProgramID, key.OrgID)
		uc.notifyClient(ctx, key.OrgID, sd.ClientID,
			fmt.Sprintf("➖ Списано %.0f %s. Остаток: %.0f %s.", amount, cur, cl.Balance, cur))
	}

	return &OpResult{OK: true, BalanceAfter: cl.Balance}, nil
}

func (uc *Usecase) Accrue(ctx context.Context, key *entity.PluginKey, session, orderID string, amount float64) (*OpResult, error) {
	if existing, err := uc.ops.Get(ctx, key.IntegrationID, orderID, "accrue"); err == nil {
		return &OpResult{OK: true, Accrued: existing.Amount, BalanceAfter: existing.BalanceAfter}, nil
	}

	sd, err := uc.loadSession(ctx, key, session)
	if err != nil {
		return nil, err
	}

	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	accrued, _ := uc.loyalty.CalculateBonus(ctx, sd.ClientID, sd.ProgramID, amount)

	cl, err := uc.loyalty.EarnFromCheck(ctx, sd.ClientID, sd.ProgramID, amount)
	if err != nil {
		return nil, fmt.Errorf("earn from check: %w", err)
	}

	op := &entity.PluginOperation{
		IntegrationID:   key.IntegrationID,
		ExternalOrderID: orderID,
		OpType:          "accrue",
		ClientID:        sd.ClientID,
		ProgramID:       sd.ProgramID,
		Amount:          accrued,
		BalanceAfter:    cl.Balance,
	}
	if err := uc.ops.Insert(ctx, op); err != nil {
		if existing, getErr := uc.ops.Get(ctx, key.IntegrationID, orderID, "accrue"); getErr == nil {
			return &OpResult{OK: true, Accrued: existing.Amount, BalanceAfter: existing.BalanceAfter}, nil
		}
		return nil, fmt.Errorf("insert operation: %w", err)
	}

	if uc.notify != nil && accrued > 0 {
		cur := uc.programCurrency(ctx, sd.ProgramID, key.OrgID)
		uc.notifyClient(ctx, key.OrgID, sd.ClientID,
			fmt.Sprintf("➕ Начислено %.0f %s. Баланс: %.0f %s.", accrued, cur, cl.Balance, cur))
	}

	return &OpResult{OK: true, Accrued: accrued, BalanceAfter: cl.Balance}, nil
}

// programCurrency returns the program's currency label, defaulting to "бонусов".
func (uc *Usecase) programCurrency(ctx context.Context, programID, orgID int) string {
	prog, err := uc.loyalty.GetProgram(ctx, programID, orgID)
	if err != nil || prog == nil || prog.Config.CurrencyName == "" {
		return "бонусов"
	}
	return prog.Config.CurrencyName
}

// notifyClient resolves the guest's chat and publishes a best-effort Telegram
// message. Failures are logged and swallowed — the POS operation already succeeded.
func (uc *Usecase) notifyClient(ctx context.Context, orgID, clientID int, text string) {
	profile, err := uc.clients.GetByID(ctx, orgID, clientID)
	if err != nil {
		uc.logger.Warn("notify client: get profile", "error", err, "client_id", clientID)
		return
	}
	if err := uc.notify.PublishNotifyClient(ctx, profile.BotID, profile.TelegramID, text); err != nil {
		uc.logger.Warn("notify client: publish", "error", err, "client_id", clientID)
	}
}

func (uc *Usecase) Config(ctx context.Context, key *entity.PluginKey) (*ConfigResult, error) {
	progs, err := uc.loyalty.GetPrograms(ctx, key.OrgID)
	if err != nil {
		return nil, fmt.Errorf("get programs: %w", err)
	}
	if len(progs) == 0 {
		return nil, ErrGuestNotFound
	}

	prog := progs[0]
	for i := range progs {
		if progs[i].IsActive {
			prog = progs[i]
			break
		}
	}

	maxPct := prog.Config.MaxRedeemPercent
	if maxPct <= 0 {
		maxPct = 100
	}

	accrualPct := 0.0
	if len(prog.Levels) > 0 {
		base := prog.Levels[0]
		for i := range prog.Levels {
			if prog.Levels[i].SortOrder < base.SortOrder {
				base = prog.Levels[i]
			}
		}
		accrualPct = base.RewardPercent
	}

	currency := prog.Config.CurrencyName
	if currency == "" {
		currency = "бонусов"
	}

	return &ConfigResult{
		ProgramType:          prog.Type,
		Currency:             currency,
		MaxRedeemPercent:     maxPct,
		AccrualPercent:       accrualPct,
		SumWithIikoDiscounts: prog.Config.SumWithDiscounts,
	}, nil
}

func (uc *Usecase) loadSession(ctx context.Context, key *entity.PluginKey, session string) (*sessionData, error) {
	raw, err := uc.code.GetSession(ctx, session)
	if err != nil {
		if errors.Is(err, poscode.ErrNotFound) {
			return nil, ErrSessionInvalid
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	var sd sessionData
	if err := json.Unmarshal(raw, &sd); err != nil {
		return nil, ErrSessionInvalid
	}
	if sd.OrgID != key.OrgID || sd.IntegrationID != key.IntegrationID {
		return nil, ErrSessionInvalid
	}
	return &sd, nil
}

// --- admin (JWT-scoped) ---

func (uc *Usecase) CreateKey(ctx context.Context, orgID, integrationID int, label string) (string, *entity.PluginKey, error) {
	intg, err := uc.integrations.GetByID(ctx, integrationID)
	if err != nil {
		return "", nil, fmt.Errorf("get integration: %w", err)
	}
	if intg.OrgID != orgID {
		return "", nil, ErrUnauthorizedKey
	}

	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("generate key: %w", err)
	}
	rawKey := "rvk_" + hex.EncodeToString(buf)

	k := &entity.PluginKey{
		OrgID:         orgID,
		IntegrationID: integrationID,
		KeyHash:       hashKey(rawKey),
		Label:         label,
	}
	if err := uc.keys.Create(ctx, k); err != nil {
		return "", nil, fmt.Errorf("create key: %w", err)
	}

	return rawKey, k, nil
}

func (uc *Usecase) ListKeys(ctx context.Context, orgID, integrationID int) ([]entity.PluginKey, error) {
	intg, err := uc.integrations.GetByID(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("get integration: %w", err)
	}
	if intg.OrgID != orgID {
		return nil, ErrUnauthorizedKey
	}
	return uc.keys.ListByIntegration(ctx, integrationID)
}

func (uc *Usecase) RevokeKey(ctx context.Context, orgID, keyID int) error {
	return uc.keys.Revoke(ctx, keyID, orgID)
}

func hashKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}
