package spire

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jpoley/nanofuse/internal/config"
)

// SVIDState tracks the lifecycle of a single SVID.
type SVIDState struct {
	// SpiffeID is the SPIFFE URI this SVID represents.
	SpiffeID string

	// IssuedAt is when the current SVID was issued.
	IssuedAt time.Time

	// ExpiresAt is when the current SVID expires.
	ExpiresAt time.Time

	// GraceUntil is ExpiresAt + GracePeriod; the old SVID stays valid until this
	// time even after a new one has been issued.
	GraceUntil time.Time

	// NewSVIDIssuedAt records when the replacement SVID was issued (zero if not
	// yet issued).  Used to detect stale agent pickups.
	NewSVIDIssuedAt time.Time

	// Picked indicates whether the agent has acknowledged the new SVID.
	Picked bool
}

// RotationCallback is called when a pre-refresh is triggered.  Implementations
// should fetch a new SVID from SPIRE and call ConfirmPickup on the manager when
// the agent has accepted it.
type RotationCallback func(ctx context.Context, spiffeID string)

// StaleAlertCallback is called when an agent has not picked up the new SVID
// within StaleAlertSeconds.  Implementations should page on-call / emit metrics.
type StaleAlertCallback func(spiffeID string, issuedAt time.Time)

// RotationManager supervises SVID TTLs, triggers pre-refresh, enforces grace
// windows, and emits stale-pickup alerts (issue #4).
//
// Behaviour summary:
//   - MaxTTL:          60 minutes (configurable)
//   - Pre-refresh:     triggered PreRefreshSeconds before expiry (default 15 min)
//   - Grace period:    old SVID valid for GracePeriodSeconds after new issuance (default 5 min)
//   - Stale alert:     warning + optional callback if agent hasn't picked up within StaleAlertSeconds (default 5 min)
type RotationManager struct {
	cfg          *config.SVIDRotationConfig
	mu           sync.RWMutex
	states       map[string]*SVIDState // keyed by SPIFFE ID
	onRotate     RotationCallback
	onStaleAlert StaleAlertCallback
}

// NewRotationManager creates a RotationManager with the given configuration.
// onRotate is called when pre-refresh is triggered (required).
// onStaleAlert is called on stale-pickup detection (optional, may be nil).
func NewRotationManager(
	cfg *config.SVIDRotationConfig,
	onRotate RotationCallback,
	onStaleAlert StaleAlertCallback,
) *RotationManager {
	return &RotationManager{
		cfg:          cfg,
		states:       make(map[string]*SVIDState),
		onRotate:     onRotate,
		onStaleAlert: onStaleAlert,
	}
}

// Track starts tracking a newly issued SVID.  ttlSeconds must not exceed
// MaxTTLSeconds; if it does it is clamped silently.
func (m *RotationManager) Track(spiffeID string, ttlSeconds int) {
	maxTTL := m.cfg.MaxTTLSeconds
	if maxTTL <= 0 {
		maxTTL = 3600
	}
	if ttlSeconds <= 0 || ttlSeconds > maxTTL {
		ttlSeconds = maxTTL
	}

	now := time.Now()
	state := &SVIDState{
		SpiffeID:  spiffeID,
		IssuedAt:  now,
		ExpiresAt: now.Add(time.Duration(ttlSeconds) * time.Second),
	}

	m.mu.Lock()
	m.states[spiffeID] = state
	m.mu.Unlock()

	slog.Info("svid tracked",
		slog.String("spiffe_id", spiffeID),
		slog.Time("expires_at", state.ExpiresAt),
		slog.Int("ttl_seconds", ttlSeconds),
	)
}

// Untrack removes an SVID from the manager (e.g. when the VM is deleted).
func (m *RotationManager) Untrack(spiffeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, spiffeID)
}

// ConfirmPickup marks the agent as having picked up the new SVID.
// Call this from the RotationCallback after the agent acknowledges.
func (m *RotationManager) ConfirmPickup(spiffeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.states[spiffeID]; ok {
		s.Picked = true
		slog.Info("svid pickup confirmed",
			slog.String("spiffe_id", spiffeID),
			slog.Duration("latency", time.Since(s.NewSVIDIssuedAt)),
		)
	}
}

// IsValid reports whether the SVID for spiffeID is currently acceptable.
// An SVID is acceptable if:
//  1. It has not yet expired, OR
//  2. A new SVID has been issued and we are still within the grace window.
func (m *RotationManager) IsValid(spiffeID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.states[spiffeID]
	if !ok {
		return false
	}
	now := time.Now()
	if now.Before(s.ExpiresAt) {
		return true
	}
	// Expired — allow if still within grace window.
	if !s.NewSVIDIssuedAt.IsZero() && now.Before(s.GraceUntil) {
		return true
	}
	return false
}

// Run starts the background tick loop.  It blocks until ctx is cancelled.
// Call it in a goroutine: go manager.Run(ctx).
func (m *RotationManager) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	slog.Info("svid rotation manager started",
		slog.Int("max_ttl_seconds", m.cfg.MaxTTLSeconds),
		slog.Int("pre_refresh_seconds", m.cfg.PreRefreshSeconds),
		slog.Int("grace_period_seconds", m.cfg.GracePeriodSeconds),
		slog.Int("stale_alert_seconds", m.cfg.StaleAlertSeconds),
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("svid rotation manager stopped")
			return
		case <-ticker.C:
			m.tick(ctx)
		}
	}
}

// tick is called every 30 seconds to check SVID state.
func (m *RotationManager) tick(ctx context.Context) {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	preRefresh := time.Duration(m.cfg.PreRefreshSeconds) * time.Second
	grace := time.Duration(m.cfg.GracePeriodSeconds) * time.Second
	staleAlert := time.Duration(m.cfg.StaleAlertSeconds) * time.Second

	for id, s := range m.states {
		timeToExpiry := s.ExpiresAt.Sub(now)

		// Pre-refresh: trigger rotation when we're within PreRefreshSeconds of expiry
		// and we haven't already issued a replacement.
		if timeToExpiry <= preRefresh && s.NewSVIDIssuedAt.IsZero() {
			slog.Info("svid pre-refresh triggered",
				slog.String("spiffe_id", id),
				slog.Duration("time_to_expiry", timeToExpiry),
			)
			s.NewSVIDIssuedAt = now
			s.GraceUntil = s.ExpiresAt.Add(grace)

			// Call the rotation callback outside the lock to avoid deadlock.
			// We need to copy the ID because the closure captures the variable.
			spiffeID := id
			go m.onRotate(ctx, spiffeID)
			continue
		}

		// Stale alert: we issued a new SVID but the agent hasn't confirmed pickup.
		if !s.NewSVIDIssuedAt.IsZero() && !s.Picked {
			sincIssued := now.Sub(s.NewSVIDIssuedAt)
			if sincIssued >= staleAlert {
				slog.Warn("svid stale pickup alert",
					slog.String("event", "svid_stale_pickup"),
					slog.String("spiffe_id", id),
					slog.Duration("since_issued", sincIssued),
					slog.Time("new_svid_issued_at", s.NewSVIDIssuedAt),
				)
				if m.onStaleAlert != nil {
					issuedAt := s.NewSVIDIssuedAt
					go m.onStaleAlert(id, issuedAt)
				}
				// Reset stale timer to avoid spamming every tick.
				s.NewSVIDIssuedAt = now
			}
		}

		// Fully expired and outside grace window — remove.
		if now.After(s.ExpiresAt) && (s.NewSVIDIssuedAt.IsZero() || now.After(s.GraceUntil)) {
			slog.Warn("svid expired and outside grace window, removing",
				slog.String("event", "svid_expired"),
				slog.String("spiffe_id", id),
				slog.Time("expired_at", s.ExpiresAt),
			)
			delete(m.states, id)
		}
	}
}
