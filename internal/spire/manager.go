package spire

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Default rotation parameters per issue #17: 60-minute SVIDs refreshed 15
// minutes before expiry.
const (
	DefaultSVIDTTL       = 60 * time.Minute
	DefaultRefreshBefore = 15 * time.Minute
	svidFileMode         = 0o400
	svidDirMode          = 0o700
)

// Clock abstracts time so rotation scheduling is deterministically testable.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// RealClock is the production Clock backed by the time package.
type RealClock struct{}

// Now returns the current time.
func (RealClock) Now() time.Time { return time.Now() }

// After returns a channel that fires after d.
func (RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// ManagerConfig configures an SVID Manager.
type ManagerConfig struct {
	// Source issues SVIDs (required).
	Source Source
	// Path is where the SVID document is written, mode 0400 (defaults to
	// SVIDDocPath()).
	Path string
	// RefreshBefore is how long before expiry to rotate (defaults to
	// DefaultRefreshBefore).
	RefreshBefore time.Duration
	// RetryInterval is how long to wait before retrying a failed rotation while
	// the current SVID is still valid (defaults to RefreshBefore/3).
	RetryInterval time.Duration
	// Clock defaults to RealClock.
	Clock Clock
	// Logger defaults to slog.Default().
	Logger *slog.Logger
}

// Manager acquires an SVID, persists it as a 0400 document, and rotates it
// before expiry. It is the portable issuance lifecycle for issue #17. In
// production this runs guest-side against a SPIRE Workload API Source over the
// vsock proxy; the mount-inside-guest and abort-VM-on-failure wiring is handled
// by the runtime and is out of scope for this in-process core.
type Manager struct {
	src           Source
	path          string
	refreshBefore time.Duration
	retryInterval time.Duration
	clk           Clock
	log           *slog.Logger

	mu      sync.RWMutex
	current *SVID

	// started guards Start against re-entry. It is set atomically at entry so a
	// second call is rejected even if the first returned an error; the manager
	// starts at most once over its lifetime.
	started atomic.Bool
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewManager validates configuration and constructs a Manager. It does not
// perform any I/O or network calls.
func NewManager(cfg ManagerConfig) (*Manager, error) {
	if cfg.Source == nil {
		return nil, fmt.Errorf("svid manager: source is required")
	}
	path := cfg.Path
	if path == "" {
		path = svidDocPath
	}
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("svid manager: path must be absolute, got %q", path)
	}
	// The path must name a file, not a directory. A trailing separator (e.g.
	// "/var/run/secrets/spiffe/") or a final element of "/", "." or ".." (e.g.
	// "/", "/var/run/.") names a directory; filepath.Base/Dir would then split it
	// into a surprising location, so reject it up front. filepath.Base cleans "."
	// and ".." away, so the final element is checked on the raw path.
	if os.IsPathSeparator(path[len(path)-1]) {
		return nil, fmt.Errorf("svid manager: path must include a filename, got a trailing separator: %q", path)
	}
	last := path
	if i := strings.LastIndexByte(path, byte(os.PathSeparator)); i >= 0 {
		last = path[i+1:]
	}
	if last == "" || last == "." || last == ".." {
		return nil, fmt.Errorf("svid manager: path must end in a filename, got %q", path)
	}
	refreshBefore := cfg.RefreshBefore
	if refreshBefore <= 0 {
		refreshBefore = DefaultRefreshBefore
	}
	retryInterval := cfg.RetryInterval
	if retryInterval <= 0 {
		retryInterval = refreshBefore / 3
	}
	if retryInterval <= 0 {
		retryInterval = time.Minute
	}
	clk := cfg.Clock
	if clk == nil {
		clk = RealClock{}
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		src:           cfg.Source,
		path:          path,
		refreshBefore: refreshBefore,
		retryInterval: retryInterval,
		clk:           clk,
		log:           logger,
	}, nil
}

// Start acquires the initial SVID, persists it, and launches the rotation loop.
//
// ctx is the lifecycle context: it governs the initial acquisition AND the
// background rotation loop. Cancel ctx (or call Stop) to stop rotation; the
// goroutine never outlives the owner's context.
//
// Fail-safe: if the initial SVID cannot be obtained, Start returns an error that
// names SPIRE unreachability and starts no background work. Callers MUST treat
// this as fatal (refuse to launch the workload) rather than continuing without
// an identity. Start must be called at most once: the single-call guard is
// consumed atomically once inputs are valid, so any subsequent call is rejected
// with an error even if that first call failed — construct a new Manager to
// retry. Invalid input is the one exception: a nil ctx is rejected before the
// guard is consumed, so it does not burn the one-shot and Start may be retried
// with a valid context.
func (m *Manager) Start(ctx context.Context) error {
	// Reject a nil context before consuming the single-call guard. issueAndPersist
	// dereferences ctx (ctx.Err()), so a nil ctx would panic; rejecting it here —
	// ahead of the CompareAndSwap — means invalid input does not burn the one-shot
	// guard, so a caller can still retry Start with a valid context.
	if ctx == nil {
		return fmt.Errorf("svid manager: Start requires a non-nil context")
	}
	if !m.started.CompareAndSwap(false, true) {
		return fmt.Errorf("svid manager: already started")
	}
	if err := m.issueAndPersist(ctx); err != nil {
		// Fail-safe either way, but name the cause accurately: only blame SPIRE
		// reachability when that is genuinely the failure. Validation and mount
		// failures are distinct security signals and must not be masked.
		if errors.Is(err, ErrSPIREUnavailable) {
			return fmt.Errorf("svid issuance failed at startup, SPIRE agent unreachable, refusing to start workload without a cryptographic identity: %w", err)
		}
		return fmt.Errorf("svid issuance failed at startup, refusing to start workload without a valid cryptographic identity: %w", err)
	}

	loopCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.done = make(chan struct{})
	go m.run(loopCtx)
	return nil
}

// Current returns a defensive copy of the most recently issued SVID, or nil
// before the first issue. The copy has independent Certificates and Bundle
// slices so a caller mutating the returned value (reassigning fields or slice
// elements) cannot corrupt Manager state or race the rotation loop.
func (m *Manager) Current() *SVID {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current.clone()
}

// Stop cancels the rotation loop and waits for it to exit. Safe to call once
// after Start; a no-op if Start never succeeded.
func (m *Manager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.done != nil {
		<-m.done
	}
}

// issueAndPersist fetches a fresh SVID, validates and verifies it, and atomically
// writes the 0400 document, then records it as current.
func (m *Manager) issueAndPersist(ctx context.Context) error {
	// Honor cancellation before reaching out to SPIRE and again after, so a
	// canceled lifecycle never results in a freshly written credential.
	if err := ctx.Err(); err != nil {
		return err
	}
	svid, err := m.src.FetchSVID(ctx)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if svid == nil {
		return fmt.Errorf("source returned a nil SVID without an error")
	}
	// Verify is the single verification path: it runs all structural/SPIFFE
	// checks (Validate) plus chain verification and the advertised
	// IssuedAt/ExpiresAt window at the current time, so a source cannot hand us a
	// structurally invalid, already-expired, or not-yet-valid SVID.
	now := m.clk.Now()
	if err := svid.Verify(now); err != nil {
		return fmt.Errorf("issued SVID failed verification: %w", err)
	}
	// Reject an SVID whose remaining lifetime does not exceed the refresh lead
	// time. Accepting it would compute a zero/negative rotation delay and spin
	// the loop, hammering SPIRE/CPU; it also indicates a misconfiguration
	// (issued TTL must be larger than RefreshBefore) or severe clock skew.
	if remaining := svid.ExpiresAt.Sub(now); remaining <= m.refreshBefore {
		return fmt.Errorf("issued SVID remaining lifetime %s does not exceed refresh lead time %s; refusing", remaining, m.refreshBefore)
	}
	data, err := svid.MarshalDocument()
	if err != nil {
		return fmt.Errorf("render SVID document: %w", err)
	}
	// Final cancellation check immediately before the write: a lifecycle canceled
	// during verification/marshal must not result in a freshly persisted
	// credential landing on disk.
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := m.writeDocument(data); err != nil {
		return err
	}
	m.mu.Lock()
	m.current = svid
	m.mu.Unlock()
	m.log.Info("SVID issued",
		slog.String("spiffe_id", svid.ID),
		slog.Time("expires_at", svid.ExpiresAt),
		slog.String("path", m.path),
	)
	return nil
}

// writeDocument creates the SVID directory if needed, then writes the document
// atomically with mode 0400. The actual write is delegated to a platform
// implementation: on unix it is directory-fd-anchored (openat/renameat with
// O_NOFOLLOW), so the directory cannot be redirected via symlink or swapped
// between validation and write.
func (m *Manager) writeDocument(data []byte) error {
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, svidDirMode); err != nil {
		return fmt.Errorf("create SVID directory %q: %w", dir, err)
	}
	return writeCredentialAtomic(dir, filepath.Base(m.path), data)
}

// run is the rotation loop. It rotates refreshBefore ahead of expiry; on
// failure it retains the still-valid current SVID and retries, but never past
// the current SVID's expiry — once expired it removes the credential (fail
// closed). It exits when ctx is cancelled.
func (m *Manager) run(ctx context.Context) {
	defer close(m.done)
	failing := false
	for {
		delay := m.refreshDelay()
		if failing {
			delay = m.failureDelay()
		}
		select {
		case <-ctx.Done():
			return
		case <-m.clk.After(delay):
		}
		if err := m.rotateOnce(ctx); err != nil {
			failing = true
			// Retain the current SVID only while it is still valid. Once it has
			// expired, a stale credential must not remain mounted: remove it and
			// drop it from state so consumers fail closed rather than presenting
			// an expired identity. Rotation keeps retrying so it can recover.
			if m.currentExpired() {
				if rmErr := m.invalidate(); rmErr != nil {
					m.log.Error("SVID expired but removal failed; credential still on disk, will keep retrying",
						slog.String("rotation_error", err.Error()),
						slog.String("removal_error", rmErr.Error()),
						slog.String("path", m.path),
					)
				} else {
					m.log.Error("SVID expired while rotation is failing; removed credential (fail-safe)",
						slog.String("error", err.Error()),
						slog.String("path", m.path),
					)
				}
			} else {
				m.log.Error("SVID rotation failed; retaining still-valid SVID and retrying",
					slog.String("error", err.Error()),
					slog.Duration("retry_in", m.failureDelay()),
				)
			}
			continue
		}
		failing = false
	}
}

// rotateOnce performs a single rotation attempt. The fetch is bounded by the
// current SVID's remaining lifetime so a hung acquisition cannot keep an
// expired credential mounted: if the current SVID has already expired it
// refuses to attempt a (possibly blocking) fetch and signals failure so the
// loop removes the stale credential.
func (m *Manager) rotateOnce(ctx context.Context) error {
	if cur := m.Current(); cur != nil {
		untilExpiry := cur.ExpiresAt.Sub(m.clk.Now())
		if untilExpiry <= 0 {
			return fmt.Errorf("current SVID expired before it could be rotated (expires_at %s)",
				cur.ExpiresAt.Format(time.RFC3339))
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, untilExpiry)
		defer cancel()
	}
	return m.issueAndPersist(ctx)
}

// currentExpired reports whether the current SVID has reached its advertised
// expiry (or there is no current SVID).
func (m *Manager) currentExpired() bool {
	cur := m.Current()
	if cur == nil {
		return true
	}
	return !m.clk.Now().Before(cur.ExpiresAt)
}

// invalidate removes the mounted credential and, only if removal succeeds,
// clears in-memory state. If removal fails it returns the error and KEEPS
// m.current set so the rotation loop keeps retrying removal — the manager must
// not report a credential as gone while an expired private-key document is still
// on disk for consumers to read.
func (m *Manager) invalidate() error {
	// Anchor the removal the same way the write path anchors the write: on unix
	// the parent directory is opened O_NOFOLLOW and the credential is removed via
	// unlinkat relative to that fd, so a parent swapped to a symlink/mount between
	// write and removal cannot redirect the unlink. See removeCredential in
	// credwrite_{unix,other}.go.
	if err := removeCredential(filepath.Dir(m.path), filepath.Base(m.path)); err != nil {
		return fmt.Errorf("remove expired SVID document %q: %w", m.path, err)
	}
	m.mu.Lock()
	m.current = nil
	m.mu.Unlock()
	return nil
}

// refreshDelay returns how long to wait before the next rotation: the time from
// now until refreshBefore ahead of the current SVID's expiry, clamped to >= 0.
func (m *Manager) refreshDelay() time.Duration {
	cur := m.Current()
	if cur == nil {
		return 0
	}
	target := cur.ExpiresAt.Add(-m.refreshBefore)
	d := target.Sub(m.clk.Now())
	if d < 0 {
		return 0
	}
	return d
}

// failureDelay returns how long to wait before the next retry while rotation is
// failing. While the current SVID is still valid, it caps the wait at the time
// remaining until expiry so the loop wakes exactly at expiry to remove a stale
// credential. Once the SVID has already expired (untilExpiry <= 0) it uses the
// full retryInterval to avoid busy-looping on repeated refresh/removal retries.
// Clamped to >= 0.
func (m *Manager) failureDelay() time.Duration {
	d := m.retryInterval
	if cur := m.Current(); cur != nil {
		if untilExpiry := cur.ExpiresAt.Sub(m.clk.Now()); untilExpiry > 0 && untilExpiry < d {
			d = untilExpiry
		}
	}
	if d < 0 {
		return 0
	}
	return d
}
