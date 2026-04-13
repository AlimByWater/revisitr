package bots

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"revisitr/internal/entity"
	botsUC "revisitr/internal/usecase/bots"
)

type managedDepsStub struct {
	storeAuthTokenFn func(ctx context.Context, token string, data entity.MasterBotAuthToken) error
	createPendingFn  func(ctx context.Context, orgID int, req *entity.CreateManagedBotRequest) (*entity.Bot, error)
	getStatusFn      func(ctx context.Context, botID, orgID int) (string, error)
}

func (m *managedDepsStub) StoreAuthToken(ctx context.Context, token string, data entity.MasterBotAuthToken) error {
	if m.storeAuthTokenFn != nil {
		return m.storeAuthTokenFn(ctx, token, data)
	}
	return nil
}

func (m *managedDepsStub) CreatePendingBot(ctx context.Context, orgID int, req *entity.CreateManagedBotRequest) (*entity.Bot, error) {
	if m.createPendingFn != nil {
		return m.createPendingFn(ctx, orgID, req)
	}
	return nil, nil
}

func (m *managedDepsStub) GetBotStatus(ctx context.Context, botID, orgID int) (string, error) {
	if m.getStatusFn != nil {
		return m.getStatusFn(ctx, botID, orgID)
	}
	return "", nil
}

func performManagedRequest(
	t *testing.T,
	handler gin.HandlerFunc,
	method, path string,
	body any,
	params gin.Params,
	keys map[string]any,
) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)

	var bodyReader *bytes.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(payload)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = params
	for key, value := range keys {
		c.Set(key, value)
	}

	handler(c)
	return rec
}

func decodeRecorderJSON[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response body: %v; body=%s", err, rec.Body.String())
	}
	return out
}

func TestManagedHandler_ActivationLink_StoresTokenAndBuildsDeepLink(t *testing.T) {
	t.Parallel()

	var storedToken string
	var storedData entity.MasterBotAuthToken

	group := New(nil, "jwt-secret", WithManagedBots(&managedDepsStub{
		storeAuthTokenFn: func(_ context.Context, token string, data entity.MasterBotAuthToken) error {
			storedToken = token
			storedData = data
			return nil
		},
	}, "masterbossbot"))

	_, _, handler := group.handleActivationLink()
	rec := performManagedRequest(
		t,
		handler,
		http.MethodPost,
		"/api/v1/bots/activation-link",
		nil,
		nil,
		map[string]any{"org_id": 21, "user_id": 34},
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(storedToken) != 32 {
		t.Fatalf("stored token len = %d, want 32", len(storedToken))
	}
	if storedData.OrgID != 21 || storedData.UserID != 34 {
		t.Fatalf("stored data = %+v, want org_id=21 user_id=34", storedData)
	}

	resp := decodeRecorderJSON[entity.ActivationLinkResponse](t, rec)
	if !strings.HasPrefix(resp.DeepLink, "https://t.me/masterbossbot?start="+storedToken) {
		t.Fatalf("deep_link = %q, want prefix %q", resp.DeepLink, "https://t.me/masterbossbot?start="+storedToken)
	}
	if diff := time.Until(resp.ExpiresAt); diff < 14*time.Minute || diff > 16*time.Minute {
		t.Fatalf("expires_at delta = %s, want about 15m", diff)
	}
}

func TestManagedHandler_CreateManaged_ValidatesAndBuildsDeepLink(t *testing.T) {
	t.Parallel()

	var capturedOrgID int
	var capturedReq *entity.CreateManagedBotRequest

	group := New(nil, "jwt-secret", WithManagedBots(&managedDepsStub{
		createPendingFn: func(_ context.Context, orgID int, req *entity.CreateManagedBotRequest) (*entity.Bot, error) {
			capturedOrgID = orgID
			reqCopy := *req
			capturedReq = &reqCopy
			return &entity.Bot{ID: 55, Status: "pending"}, nil
		},
	}, "masterbossbot"))

	_, _, handler := group.handleCreateManaged()
	rec := performManagedRequest(
		t,
		handler,
		http.MethodPost,
		"/api/v1/bots/create-managed",
		map[string]any{
			"name":     "ManagedLaunchBot",
			"username": "@managedlaunchbot",
			"modules":  []string{"loyalty", "campaigns"},
		},
		nil,
		map[string]any{"org_id": 87},
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if capturedOrgID != 87 {
		t.Fatalf("captured org_id = %d, want 87", capturedOrgID)
	}
	if capturedReq == nil {
		t.Fatal("expected request to be forwarded to managed deps")
	}
	if capturedReq.Username != "managedlaunchbot" {
		t.Fatalf("username = %q, want %q", capturedReq.Username, "managedlaunchbot")
	}

	resp := decodeRecorderJSON[entity.CreateManagedBotResponse](t, rec)
	if resp.BotID != 55 {
		t.Fatalf("bot_id = %d, want 55", resp.BotID)
	}
	if resp.Status != "pending" {
		t.Fatalf("status = %q, want %q", resp.Status, "pending")
	}
	wantDeepLink := "https://t.me/newbot/masterbossbot/managedlaunchbot?name=ManagedLaunchBot"
	if resp.DeepLink != wantDeepLink {
		t.Fatalf("deep_link = %q, want %q", resp.DeepLink, wantDeepLink)
	}
}

func TestManagedHandler_CreateManaged_RejectsInvalidUsername(t *testing.T) {
	t.Parallel()

	called := false
	group := New(nil, "jwt-secret", WithManagedBots(&managedDepsStub{
		createPendingFn: func(_ context.Context, _ int, _ *entity.CreateManagedBotRequest) (*entity.Bot, error) {
			called = true
			return &entity.Bot{}, nil
		},
	}, "masterbossbot"))

	_, _, handler := group.handleCreateManaged()
	rec := performManagedRequest(
		t,
		handler,
		http.MethodPost,
		"/api/v1/bots/create-managed",
		map[string]any{"name": "Managed", "username": "@managed-launch"},
		nil,
		map[string]any{"org_id": 87},
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if called {
		t.Fatal("expected managed create deps to not be called")
	}
	if !strings.Contains(rec.Body.String(), "username must end with 'bot'") {
		t.Fatalf("body = %q, want validation error", rec.Body.String())
	}
}

func TestManagedHandler_GetBotStatus_MapsOwnershipError(t *testing.T) {
	t.Parallel()

	group := New(nil, "jwt-secret", WithManagedBots(&managedDepsStub{
		getStatusFn: func(_ context.Context, botID, orgID int) (string, error) {
			if botID != 99 || orgID != 4 {
				t.Fatalf("got botID=%d orgID=%d, want botID=99 orgID=4", botID, orgID)
			}
			return "", botsUC.ErrNotBotOwner
		},
	}, "masterbossbot"))

	_, _, handler := group.handleGetBotStatus()
	rec := performManagedRequest(
		t,
		handler,
		http.MethodGet,
		"/api/v1/bots/99/status",
		nil,
		gin.Params{{Key: "id", Value: "99"}},
		map[string]any{"org_id": 4},
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestManagedHandler_GetBotStatus_SuccessAndInvalidID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		group := New(nil, "jwt-secret", WithManagedBots(&managedDepsStub{
			getStatusFn: func(_ context.Context, botID, orgID int) (string, error) {
				if botID != 7 || orgID != 3 {
					t.Fatalf("got botID=%d orgID=%d, want botID=7 orgID=3", botID, orgID)
				}
				return "pending", nil
			},
		}, "masterbossbot"))

		_, _, handler := group.handleGetBotStatus()
		rec := performManagedRequest(
			t,
			handler,
			http.MethodGet,
			"/api/v1/bots/7/status",
			nil,
			gin.Params{{Key: "id", Value: "7"}},
			map[string]any{"org_id": 3},
		)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var body struct {
			Status string `json:"status"`
		}
		body = decodeRecorderJSON[struct {
			Status string `json:"status"`
		}](t, rec)
		if body.Status != "pending" {
			t.Fatalf("status body = %q, want %q", body.Status, "pending")
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		t.Parallel()

		group := New(nil, "jwt-secret", WithManagedBots(&managedDepsStub{
			getStatusFn: func(_ context.Context, _, _ int) (string, error) {
				return "", errors.New("should not be called")
			},
		}, "masterbossbot"))

		_, _, handler := group.handleGetBotStatus()
		rec := performManagedRequest(
			t,
			handler,
			http.MethodGet,
			"/api/v1/bots/not-an-id/status",
			nil,
			gin.Params{{Key: "id", Value: "not-an-id"}},
			map[string]any{"org_id": 3},
		)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}
