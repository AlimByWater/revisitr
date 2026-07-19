package poscode

import (
	"testing"
	"time"
)

// Redis-dependent behavior (Issue/Consume/sessions/rate-limit) requires a live
// Redis server and is exercised by integration tests, not here. These unit
// tests cover construction defaults and pure helpers only.

func TestNewDefaults(t *testing.T) {
	s := New(nil)
	if s.ttl != defaultTTL {
		t.Errorf("default ttl = %v, want %v", s.ttl, defaultTTL)
	}
	if s.sessionTTL != defaultSessionTTL {
		t.Errorf("default sessionTTL = %v, want %v", s.sessionTTL, defaultSessionTTL)
	}
}

func TestNewOptions(t *testing.T) {
	s := New(nil, WithTTL(time.Minute), WithSessionTTL(2*time.Hour))
	if s.ttl != time.Minute {
		t.Errorf("ttl = %v, want %v", s.ttl, time.Minute)
	}
	if s.sessionTTL != 2*time.Hour {
		t.Errorf("sessionTTL = %v, want %v", s.sessionTTL, 2*time.Hour)
	}
}

func TestRandomTokenUniqueHex(t *testing.T) {
	a, err := randomToken()
	if err != nil {
		t.Fatalf("randomToken: %v", err)
	}
	b, err := randomToken()
	if err != nil {
		t.Fatalf("randomToken: %v", err)
	}
	if len(a) != 32 {
		t.Errorf("token length = %d, want 32", len(a))
	}
	if a == b {
		t.Errorf("two tokens collided: %q", a)
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 7: "7", 42: "42", 1000: "1000", -5: "-5"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Errorf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}
