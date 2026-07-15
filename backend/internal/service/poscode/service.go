// Package poscode issues and consumes ephemeral one-time guest identification
// codes for the POS loyalty plugin. A code is a single common Russian word
// shown to a guest; a cashier types it back to identify the guest at the till.
//
// Storage is Redis. The one-time consume pattern mirrors
// internal/repository/redis/master_bot_auth.go: Set with a TTL on issue, and
// an atomic GetDel on consume.
package poscode

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	codePrefix    = "poscode:code:"
	clientPrefix  = "poscode:client:"
	sessionPrefix = "poscode:session:"
	rlPrefix      = "poscode:rl:"
	seqKey        = "poscode:seq"

	defaultTTL        = 5 * time.Minute
	defaultSessionTTL = 15 * time.Minute

	issueMaxTries = 8
)

// ErrNotFound is returned when a code or session is missing or expired.
var ErrNotFound = errors.New("poscode: code not found or expired")

// Grant is the loyalty context bound to an issued code.
type Grant struct {
	ClientID      int `json:"client_id"`
	ProgramID     int `json:"program_id"`
	OrgID         int `json:"org_id"`
	IntegrationID int `json:"integration_id"`
}

// Service issues, consumes, and rate-limits POS guest codes over Redis.
//
// The redis client is resolved lazily via a getter so the service can be
// constructed before the redis module is initialised (matches eventbus.New).
type Service struct {
	redis      func() *goredis.Client
	ttl        time.Duration
	sessionTTL time.Duration
}

// Option configures a Service.
type Option func(*Service)

// WithTTL overrides the code TTL (default 5m).
func WithTTL(d time.Duration) Option {
	return func(s *Service) { s.ttl = d }
}

// WithSessionTTL overrides the session TTL (default 15m).
func WithSessionTTL(d time.Duration) Option {
	return func(s *Service) { s.sessionTTL = d }
}

// New constructs a Service. redis is a getter resolved lazily on each call.
// Defaults: ttl=5m, sessionTTL=15m.
func New(redis func() *goredis.Client, opts ...Option) *Service {
	s := &Service{
		redis:      redis,
		ttl:        defaultTTL,
		sessionTTL: defaultSessionTTL,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Issue picks an unused word, stores code->grant with the code TTL, and ensures
// only one active code per client (the client's previous code is deleted first).
// It returns the display word (original casing from the wordlist, not normalized).
//
// Word selection avoids math/rand entirely (the runtime forbids its
// auto-seeding). Instead we draw a base index from a Redis INCR counter
// ("poscode:seq") modulo len(words), then add the attempt offset on each retry.
// SetNX on the normalized code key guarantees we never overwrite a currently
// active code; on collision we advance to the next attempt index.
func (s *Service) Issue(ctx context.Context, g Grant) (string, error) {
	// Drop any previous active code for this client so only one is live.
	if err := s.clearClient(ctx, g.ClientID); err != nil {
		return "", err
	}

	value, err := json.Marshal(g)
	if err != nil {
		return "", fmt.Errorf("poscode.Issue: marshal grant: %w", err)
	}

	seq, err := s.redis().Incr(ctx, seqKey).Result()
	if err != nil {
		return "", fmt.Errorf("poscode.Issue: seq incr: %w", err)
	}
	base := int(seq % int64(len(words)))

	for attempt := 0; attempt < issueMaxTries; attempt++ {
		idx := (base + attempt) % len(words)
		display := words[idx]
		norm := Normalize(display)
		codeKey := codePrefix + norm

		ok, err := s.redis().SetNX(ctx, codeKey, value, s.ttl).Result()
		if err != nil {
			return "", fmt.Errorf("poscode.Issue: setnx: %w", err)
		}
		if !ok {
			// Code currently active (collision); try the next word.
			continue
		}

		// Reverse index so a later Issue for this client can evict this code.
		if err := s.redis().Set(ctx, clientPrefix+itoa(g.ClientID), norm, s.ttl).Err(); err != nil {
			// Best-effort cleanup of the code we just set.
			s.redis().Del(ctx, codeKey)
			return "", fmt.Errorf("poscode.Issue: set reverse: %w", err)
		}
		return display, nil
	}

	return "", fmt.Errorf("poscode.Issue: no free code after %d tries", issueMaxTries)
}

// Consume normalizes the word, atomically retrieves-and-deletes the code, and
// returns the bound Grant. It also clears the reverse client key. Returns
// ErrNotFound if the code is missing or expired.
func (s *Service) Consume(ctx context.Context, word string) (Grant, error) {
	norm := Normalize(word)
	codeKey := codePrefix + norm

	b, err := s.redis().GetDel(ctx, codeKey).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return Grant{}, ErrNotFound
		}
		return Grant{}, fmt.Errorf("poscode.Consume: getdel: %w", err)
	}

	var g Grant
	if err := json.Unmarshal(b, &g); err != nil {
		return Grant{}, fmt.Errorf("poscode.Consume: unmarshal: %w", err)
	}

	// Best-effort: clear the reverse key only if it still points at this code.
	if cur, err := s.redis().Get(ctx, clientPrefix+itoa(g.ClientID)).Result(); err == nil && cur == norm {
		s.redis().Del(ctx, clientPrefix+itoa(g.ClientID))
	}

	return g, nil
}

// CreateSession stores an opaque, random session token -> raw payload with the
// session TTL and returns the token (32 hex chars / 16 random bytes).
func (s *Service) CreateSession(ctx context.Context, payload []byte) (string, error) {
	tok, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("poscode.CreateSession: token: %w", err)
	}
	if err := s.redis().Set(ctx, sessionPrefix+tok, payload, s.sessionTTL).Err(); err != nil {
		return "", fmt.Errorf("poscode.CreateSession: set: %w", err)
	}
	return tok, nil
}

// GetSession returns the raw payload for a token without deleting it (the
// session is reusable within its TTL). Returns ErrNotFound if missing/expired.
func (s *Service) GetSession(ctx context.Context, token string) ([]byte, error) {
	b, err := s.redis().Get(ctx, sessionPrefix+token).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("poscode.GetSession: get: %w", err)
	}
	return b, nil
}

// AllowAttempt increments the counter for scope and reports whether it is still
// at or under limit. The window TTL is set on the first hit of the window.
func (s *Service) AllowAttempt(ctx context.Context, scope string, limit int, window time.Duration) (bool, error) {
	key := rlPrefix + scope
	n, err := s.redis().Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("poscode.AllowAttempt: incr: %w", err)
	}
	if n == 1 {
		if err := s.redis().Expire(ctx, key, window).Err(); err != nil {
			return false, fmt.Errorf("poscode.AllowAttempt: expire: %w", err)
		}
	}
	return n <= int64(limit), nil
}

// clearClient deletes the client's currently active code (if any) and its
// reverse key.
func (s *Service) clearClient(ctx context.Context, clientID int) error {
	revKey := clientPrefix + itoa(clientID)
	prev, err := s.redis().Get(ctx, revKey).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil
		}
		return fmt.Errorf("poscode.clearClient: get: %w", err)
	}
	if err := s.redis().Del(ctx, codePrefix+prev, revKey).Err(); err != nil {
		return fmt.Errorf("poscode.clearClient: del: %w", err)
	}
	return nil
}

// randomToken returns 16 crypto-random bytes hex-encoded (32 chars).
func randomToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// itoa converts an int to its decimal string without pulling in fmt for hot keys.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
