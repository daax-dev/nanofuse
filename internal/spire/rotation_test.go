package spire

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jpoley/nanofuse/internal/config"
)

func testCfg() *config.SVIDRotationConfig {
	return &config.SVIDRotationConfig{
		MaxTTLSeconds:      3600,
		PreRefreshSeconds:  900,
		GracePeriodSeconds: 300,
		StaleAlertSeconds:  300,
	}
}

// TestTrackAndIsValid verifies that a freshly tracked SVID is valid.
func TestTrackAndIsValid(t *testing.T) {
	var called atomic.Int32
	m := NewRotationManager(testCfg(),
		func(_ context.Context, _ string) { called.Add(1) },
		nil,
	)
	m.Track("spiffe://example.com/vm/1", 3600)
	if !m.IsValid("spiffe://example.com/vm/1") {
		t.Error("freshly tracked SVID should be valid")
	}
}

// TestUntrack verifies that an untracked SVID is no longer valid.
func TestUntrack(t *testing.T) {
	m := NewRotationManager(testCfg(),
		func(_ context.Context, _ string) {},
		nil,
	)
	id := "spiffe://example.com/vm/untrack"
	m.Track(id, 3600)
	m.Untrack(id)
	if m.IsValid(id) {
		t.Error("untracked SVID should not be valid")
	}
}

// TestExpiredSVIDNotValid verifies that an expired SVID without a grace window
// is not considered valid.
func TestExpiredSVIDNotValid(t *testing.T) {
	m := NewRotationManager(testCfg(),
		func(_ context.Context, _ string) {},
		nil,
	)
	id := "spiffe://example.com/vm/expired"
	// Track with 1-second TTL and wait.
	m.Track(id, 1)
	time.Sleep(2 * time.Second)
	if m.IsValid(id) {
		t.Error("expired SVID should not be valid")
	}
}

// TestGraceWindowValid verifies that an expired SVID is still valid during the
// grace period after a new SVID has been issued.
func TestGraceWindowValid(t *testing.T) {
	m := NewRotationManager(testCfg(),
		func(_ context.Context, _ string) {},
		nil,
	)
	id := "spiffe://example.com/vm/grace"
	// Manually insert a state that's just expired but inside grace window.
	now := time.Now()
	m.mu.Lock()
	m.states[id] = &SVIDState{
		SpiffeID:        id,
		IssuedAt:        now.Add(-3700 * time.Second),
		ExpiresAt:       now.Add(-100 * time.Second), // already expired
		NewSVIDIssuedAt: now.Add(-60 * time.Second),  // new SVID issued 60s ago
		GraceUntil:      now.Add(240 * time.Second),  // grace ends in 4 min
	}
	m.mu.Unlock()

	if !m.IsValid(id) {
		t.Error("SVID within grace window should still be valid")
	}
}

// TestGraceWindowExpiredNotValid verifies that an SVID outside the grace window
// is invalid.
func TestGraceWindowExpiredNotValid(t *testing.T) {
	m := NewRotationManager(testCfg(),
		func(_ context.Context, _ string) {},
		nil,
	)
	id := "spiffe://example.com/vm/no-grace"
	now := time.Now()
	m.mu.Lock()
	m.states[id] = &SVIDState{
		SpiffeID:        id,
		IssuedAt:        now.Add(-4000 * time.Second),
		ExpiresAt:       now.Add(-400 * time.Second), // expired
		NewSVIDIssuedAt: now.Add(-400 * time.Second), // issued same time as expiry
		GraceUntil:      now.Add(-100 * time.Second), // grace also expired
	}
	m.mu.Unlock()

	if m.IsValid(id) {
		t.Error("SVID outside grace window should not be valid")
	}
}

// TestConfirmPickup verifies that ConfirmPickup marks the SVID as picked up.
func TestConfirmPickup(t *testing.T) {
	m := NewRotationManager(testCfg(),
		func(_ context.Context, _ string) {},
		nil,
	)
	id := "spiffe://example.com/vm/pickup"
	m.Track(id, 3600)
	m.ConfirmPickup(id)

	m.mu.RLock()
	s := m.states[id]
	m.mu.RUnlock()

	if s == nil {
		t.Fatal("state not found after ConfirmPickup")
	}
	if !s.Picked {
		t.Error("Picked should be true after ConfirmPickup")
	}
}

// TestMaxTTLClamped verifies that a TTL exceeding MaxTTLSeconds is clamped.
func TestMaxTTLClamped(t *testing.T) {
	cfg := &config.SVIDRotationConfig{
		MaxTTLSeconds:      60,
		PreRefreshSeconds:  15,
		GracePeriodSeconds: 5,
		StaleAlertSeconds:  5,
	}
	m := NewRotationManager(cfg, func(_ context.Context, _ string) {}, nil)
	id := "spiffe://example.com/vm/clamped"
	m.Track(id, 7200) // exceeds max of 60s

	m.mu.RLock()
	s := m.states[id]
	m.mu.RUnlock()

	ttl := int(s.ExpiresAt.Sub(s.IssuedAt).Seconds())
	if ttl > 60 {
		t.Errorf("TTL should be clamped to 60s, got %d", ttl)
	}
}

// TestPreRefreshTriggered verifies that the rotation callback fires when the
// SVID is within the pre-refresh window.
func TestPreRefreshTriggered(t *testing.T) {
	var triggered atomic.Int32
	cfg := &config.SVIDRotationConfig{
		MaxTTLSeconds:      3600,
		PreRefreshSeconds:  600,
		GracePeriodSeconds: 300,
		StaleAlertSeconds:  300,
	}
	m := NewRotationManager(cfg, func(_ context.Context, _ string) {
		triggered.Add(1)
	}, nil)

	id := "spiffe://example.com/vm/prerefresh"
	now := time.Now()
	// Manually insert a state that's 400s from expiry (inside the 600s pre-refresh window).
	m.mu.Lock()
	m.states[id] = &SVIDState{
		SpiffeID:  id,
		IssuedAt:  now.Add(-3200 * time.Second),
		ExpiresAt: now.Add(400 * time.Second),
	}
	m.mu.Unlock()

	m.tick(context.Background())
	time.Sleep(50 * time.Millisecond) // allow goroutine to run

	if triggered.Load() == 0 {
		t.Error("rotation callback should have been triggered in pre-refresh window")
	}
}
