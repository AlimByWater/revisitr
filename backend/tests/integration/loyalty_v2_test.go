//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
)

type loyaltyProgramResp struct {
	ID       int    `json:"id"`
	OrgID    int    `json:"org_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsActive bool   `json:"is_active"`
}

type loyaltyLevelResp struct {
	ID            int     `json:"id"`
	ProgramID     int     `json:"program_id"`
	Name          string  `json:"name"`
	Threshold     int     `json:"threshold"`
	RewardPercent float64 `json:"reward_percent"`
	SortOrder     int     `json:"sort_order"`
}

type clientLoyaltyResp struct {
	ID          int     `json:"id"`
	ClientID    int     `json:"client_id"`
	ProgramID   int     `json:"program_id"`
	Balance     float64 `json:"balance"`
	TotalEarned float64 `json:"total_earned"`
}

func mustCreateProgram(t *testing.T, token, name, pType string) loyaltyProgramResp {
	t.Helper()

	body := map[string]interface{}{
		"name": name,
		"type": pType,
		"config": map[string]interface{}{
			"welcome_bonus": 100,
			"currency_name": "points",
		},
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/loyalty/programs", body, token)
	assertStatus(t, resp, http.StatusCreated)

	var p loyaltyProgramResp
	decodeJSON(t, resp, &p)

	t.Cleanup(func() {
		db := pgMod.DB()
		db.ExecContext(t.Context(), "DELETE FROM loyalty_levels WHERE program_id = $1", p.ID)
		db.ExecContext(t.Context(), "DELETE FROM client_loyalty WHERE program_id = $1", p.ID)
		db.ExecContext(t.Context(), "DELETE FROM balance_reserves WHERE program_id = $1", p.ID)
		db.ExecContext(t.Context(), "DELETE FROM loyalty_programs WHERE id = $1", p.ID)
	})

	return p
}

func mustCreateLevel(t *testing.T, token string, programID int, name string, threshold int, rewardPercent float64) loyaltyLevelResp {
	t.Helper()

	body := map[string]interface{}{
		"name":           name,
		"threshold":      threshold,
		"reward_percent": rewardPercent,
		"reward_type":    "percent",
		"sort_order":     threshold, // use threshold as sort order for simplicity
	}
	resp := doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/loyalty/programs/%d/levels", programID),
		body, token)
	assertStatus(t, resp, http.StatusCreated)

	var l loyaltyLevelResp
	decodeJSON(t, resp, &l)

	return l
}

// mustSeedBotClient inserts a bot_client directly into the DB and returns its ID.
// This is needed for loyalty operations that require a client_id.
func mustSeedBotClient(t *testing.T, orgID, botID int) int {
	t.Helper()

	tgID := uniqueTelegramID()
	var clientID int
	err := pgMod.DB().QueryRowContext(t.Context(),
		`INSERT INTO bot_clients (bot_id, telegram_id, first_name, last_name, username, phone)
		 VALUES ($1, $2, 'TestClient', '', 'testuser', $3)
		 RETURNING id`,
		botID, tgID, fmt.Sprintf("+7900%07d", tgID%10000000)).Scan(&clientID)
	if err != nil {
		t.Fatalf("seed bot client: %v", err)
	}

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(), "DELETE FROM bot_clients WHERE id = $1", clientID)
	})

	return clientID
}

var tgIDCounter int64

func uniqueTelegramID() int64 {
	tgIDCounter++
	return 900000000 + tgIDCounter
}

// --- Program CRUD ---

func TestLoyalty_Program_CRUD(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "LoyalOwner", "Loyal Org")

	// Create
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Test Program", "bonus")
	if p.ID == 0 {
		t.Fatal("expected non-zero program ID")
	}
	if p.Type != "bonus" {
		t.Errorf("expected type bonus, got %s", p.Type)
	}
	if !p.IsActive {
		t.Error("expected program to be active on creation")
	}

	// Get
	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/loyalty/programs/%d", p.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	var fetched loyaltyProgramResp
	decodeJSON(t, resp, &fetched)
	if fetched.Name != "Test Program" {
		t.Errorf("expected name %q, got %q", "Test Program", fetched.Name)
	}

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/loyalty/programs", nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	var programs []loyaltyProgramResp
	decodeJSON(t, resp, &programs)
	if len(programs) < 1 {
		t.Error("expected at least 1 program")
	}

	// Update
	updateBody := map[string]interface{}{"name": "Updated Program"}
	resp = doRequest(t, http.MethodPatch,
		fmt.Sprintf("/api/v1/loyalty/programs/%d", p.ID),
		updateBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

// --- Levels ---

func TestLoyalty_Level_Create(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "LvlOwner", "Lvl Org")
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Level Program", "bonus")

	l := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "Bronze", 0, 5.0)
	if l.ID == 0 {
		t.Fatal("expected non-zero level ID")
	}
	if l.ProgramID != p.ID {
		t.Errorf("expected program_id %d, got %d", p.ID, l.ProgramID)
	}
}

func TestLoyalty_Level_BatchUpdate(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "BatchOwner", "Batch Org")
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Batch Program", "bonus")

	l1 := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "Bronze", 0, 5.0)
	l2 := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "Silver", 100, 10.0)

	body := map[string]interface{}{
		"levels": []map[string]interface{}{
			{"id": l1.ID, "program_id": p.ID, "name": "Bronze Updated", "threshold": 0, "reward_percent": 7.0, "reward_type": "percent", "sort_order": 1},
			{"id": l2.ID, "program_id": p.ID, "name": "Silver Updated", "threshold": 200, "reward_percent": 12.0, "reward_type": "percent", "sort_order": 2},
		},
	}
	resp := doRequest(t, http.MethodPut,
		fmt.Sprintf("/api/v1/loyalty/programs/%d/levels", p.ID),
		body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var levels []loyaltyLevelResp
	decodeJSON(t, resp, &levels)
	if len(levels) < 2 {
		t.Errorf("expected at least 2 levels, got %d", len(levels))
	}
}

func TestLoyalty_Level_Delete(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "DelLvlOwner", "DelLvl Org")
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Delete Level Program", "bonus")
	l := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "ToDelete", 0, 5.0)

	resp := doRequest(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/loyalty/programs/%d/levels/%d", p.ID, l.ID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

// mustSeedClientLoyalty creates a client_loyalty record with a level pre-assigned
// so that CalculateBonus works (it requires LevelID to be set).
func mustSeedClientLoyalty(t *testing.T, clientID, programID, levelID int) {
	t.Helper()
	_, err := pgMod.DB().ExecContext(t.Context(),
		`INSERT INTO client_loyalty (client_id, program_id, level_id, balance, total_earned, total_spent)
		 VALUES ($1, $2, $3, 0, 0, 0)
		 ON CONFLICT (client_id, program_id) DO UPDATE SET level_id = $3`,
		clientID, programID, levelID)
	if err != nil {
		t.Fatalf("seed client loyalty: %v", err)
	}
	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(),
			"DELETE FROM client_loyalty WHERE client_id = $1 AND program_id = $2", clientID, programID)
	})
}

// --- EarnFromCheck ---

func TestLoyalty_EarnFromCheck(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "EarnOwner", "Earn Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Earn Bot")
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Earn Program", "bonus")
	level := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "Base", 0, 10.0)

	clientID := mustSeedBotClient(t, ar.User.OrgID, bot.ID)
	mustSeedClientLoyalty(t, clientID, p.ID, level.ID)

	body := map[string]interface{}{
		"client_id":    clientID,
		"program_id":   p.ID,
		"check_amount": 1000.0,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/loyalty/earn-from-check", body, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)

	var cl clientLoyaltyResp
	decodeJSON(t, resp, &cl)
	if cl.Balance <= 0 {
		t.Error("expected positive balance after earning")
	}
}

// --- Reserve / Confirm / Cancel ---

func TestLoyalty_Reserve_Confirm(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "ResOwner", "Res Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Reserve Bot")
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Reserve Program", "bonus")
	level := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "Base", 0, 10.0)

	clientID := mustSeedBotClient(t, ar.User.OrgID, bot.ID)
	mustSeedClientLoyalty(t, clientID, p.ID, level.ID)

	// Earn some points first
	earnBody := map[string]interface{}{
		"client_id":    clientID,
		"program_id":   p.ID,
		"check_amount": 5000.0,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/loyalty/earn-from-check", earnBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Reserve
	reserveBody := map[string]interface{}{
		"client_id":  clientID,
		"program_id": p.ID,
		"amount":     100.0,
	}
	resp = doRequest(t, http.MethodPost, "/api/v1/loyalty/reserve", reserveBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)

	var reserveResp struct {
		ReserveID int `json:"reserve_id"`
	}
	decodeJSON(t, resp, &reserveResp)
	if reserveResp.ReserveID == 0 {
		t.Fatal("expected non-zero reserve ID")
	}

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(),
			"DELETE FROM balance_reserves WHERE id = $1", reserveResp.ReserveID)
	})

	// Confirm
	resp = doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/loyalty/reserve/%d/confirm", reserveResp.ReserveID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestLoyalty_Reserve_Cancel(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "CancelOwner", "Cancel Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Cancel Bot")
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Cancel Program", "bonus")
	level := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "Base", 0, 10.0)

	clientID := mustSeedBotClient(t, ar.User.OrgID, bot.ID)
	mustSeedClientLoyalty(t, clientID, p.ID, level.ID)

	// Earn points
	earnBody := map[string]interface{}{
		"client_id":    clientID,
		"program_id":   p.ID,
		"check_amount": 5000.0,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/loyalty/earn-from-check", earnBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Reserve
	reserveBody := map[string]interface{}{
		"client_id":  clientID,
		"program_id": p.ID,
		"amount":     50.0,
	}
	resp = doRequest(t, http.MethodPost, "/api/v1/loyalty/reserve", reserveBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusCreated)

	var reserveResp struct {
		ReserveID int `json:"reserve_id"`
	}
	decodeJSON(t, resp, &reserveResp)

	t.Cleanup(func() {
		pgMod.DB().ExecContext(t.Context(),
			"DELETE FROM balance_reserves WHERE id = $1", reserveResp.ReserveID)
	})

	// Cancel
	resp = doRequest(t, http.MethodPost,
		fmt.Sprintf("/api/v1/loyalty/reserve/%d/cancel", reserveResp.ReserveID),
		nil, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestLoyalty_Reserve_InsufficientPoints(t *testing.T) {
	ar := mustRegister(t, uniqueEmail(t), "password123", "InsfOwner", "Insf Org")
	bot := mustCreateBot(t, ar.Tokens.AccessToken, "Insf Bot")
	p := mustCreateProgram(t, ar.Tokens.AccessToken, "Insf Program", "bonus")
	level := mustCreateLevel(t, ar.Tokens.AccessToken, p.ID, "Base", 0, 10.0)

	clientID := mustSeedBotClient(t, ar.User.OrgID, bot.ID)
	mustSeedClientLoyalty(t, clientID, p.ID, level.ID)

	// Try to reserve without any points
	reserveBody := map[string]interface{}{
		"client_id":  clientID,
		"program_id": p.ID,
		"amount":     99999.0,
	}
	resp := doRequest(t, http.MethodPost, "/api/v1/loyalty/reserve", reserveBody, ar.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

// --- Cross-org isolation ---

func TestLoyalty_WrongOrg(t *testing.T) {
	ar1 := mustRegister(t, uniqueEmail(t), "password123", "LoyalOrg1", "Loyal Org1")
	ar2 := mustRegister(t, uniqueEmail(t), "password123", "LoyalOrg2", "Loyal Org2")
	p := mustCreateProgram(t, ar1.Tokens.AccessToken, "Private Program", "bonus")

	// Get
	resp := doRequest(t, http.MethodGet,
		fmt.Sprintf("/api/v1/loyalty/programs/%d", p.ID),
		nil, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()

	// Update
	resp = doRequest(t, http.MethodPatch,
		fmt.Sprintf("/api/v1/loyalty/programs/%d", p.ID),
		map[string]string{"name": "hack"}, ar2.Tokens.AccessToken)
	assertStatus(t, resp, http.StatusForbidden)
	resp.Body.Close()
}
