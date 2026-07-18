package botmanager

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"revisitr/internal/entity"
)

type fakeOrgsRepo struct {
	tz  string
	err error
}

func (f fakeOrgsRepo) GetTimezone(_ context.Context, _ int) (string, error) {
	return f.tz, f.err
}

func orgTimeHandler(orgs organizationsRepository) *handler {
	return &handler{
		mgr:    &Manager{orgsRepo: orgs},
		info:   &entity.Bot{ID: 1, OrgID: 1},
		logger: slog.Default(),
	}
}

func TestOrgNowUsesOrgTimezone(t *testing.T) {
	h := orgTimeHandler(fakeOrgsRepo{tz: "Asia/Kamchatka"})
	got := h.orgNow(context.Background())

	want, _ := time.LoadLocation("Asia/Kamchatka")
	if got.Location().String() != want.String() {
		t.Errorf("orgNow location = %s, want Asia/Kamchatka", got.Location())
	}
}

func TestOrgNowFallsBackToServerTime(t *testing.T) {
	serverLoc := time.Now().Location().String()
	cases := map[string]organizationsRepository{
		"nil repo":    nil,
		"empty tz":    fakeOrgsRepo{tz: ""},
		"repo error":  fakeOrgsRepo{err: errors.New("db down")},
		"bad tz name": fakeOrgsRepo{tz: "Mars/Olympus"},
	}
	for name, repo := range cases {
		h := orgTimeHandler(repo)
		if got := h.orgNow(context.Background()); got.Location().String() != serverLoc {
			t.Errorf("%s: orgNow location = %s, want server-local %s", name, got.Location(), serverLoc)
		}
	}
}
