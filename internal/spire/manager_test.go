package spire

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func newTestManager(t *testing.T, src Source, clk Clock) (*Manager, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "secrets", "spiffe", "svid.json")
	mgr, err := NewManager(ManagerConfig{
		Source:        src,
		Path:          path,
		RefreshBefore: DefaultRefreshBefore,
		Clock:         clk,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return mgr, path
}

func TestManager_Start_FailSafeOnUnreachableSPIRE(t *testing.T) {
	mgr, path := newTestManager(t, failingSource{}, newFakeClock(time.Now()))

	err := mgr.Start(context.Background())
	if err == nil {
		t.Fatal("Start must fail when SPIRE is unreachable (fail-safe)")
	}
	if !errors.Is(err, ErrSPIREUnavailable) {
		t.Fatalf("error must wrap ErrSPIREUnavailable, got %v", err)
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "spire") || !strings.Contains(msg, "unreachable") {
		t.Fatalf("fail-safe error must name SPIRE unreachability, got %q", err.Error())
	}
	// No credential must be written on failure.
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatal("no SVID document must be written when startup fails")
	}
	if mgr.Current() != nil {
		t.Fatal("Current() must be nil after failed Start")
	}
}

func TestManager_Current_ReturnsDefensiveCopy(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-copy")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	mgr, _ := newTestManager(t, src, clk)
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop()

	got := mgr.Current()
	if got == nil {
		t.Fatal("Current must be set after Start")
	}
	// Two calls must not alias the same struct: a caller mutating one copy must
	// not affect Manager state or another caller.
	if other := mgr.Current(); got == other {
		t.Fatal("Current() must return a fresh copy, not the shared internal pointer")
	}
	origID := got.ID
	origLen := len(got.Certificates)
	origLeaf := got.Certificates[0]

	// Mutate the returned value and its slices.
	got.ID = "spiffe://attacker.example/evil"
	got.ExpiresAt = got.ExpiresAt.Add(-100 * time.Hour)
	got.Certificates[0] = nil
	got.Bundle[0] = nil

	after := mgr.Current()
	if after.ID != origID {
		t.Fatalf("mutating returned SVID corrupted Manager ID: got %q, want %q", after.ID, origID)
	}
	if len(after.Certificates) != origLen || after.Certificates[0] != origLeaf {
		t.Fatal("mutating returned Certificates slice corrupted Manager state")
	}
	if after.Bundle[0] == nil {
		t.Fatal("mutating returned Bundle slice corrupted Manager state")
	}
	if err := after.Verify(clk.Now()); err != nil {
		t.Fatalf("Manager SVID must remain valid after caller mutation: %v", err)
	}
}

func TestManager_Start_WritesMode0400(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-mount")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	mgr, path := newTestManager(t, src, clk)
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat SVID document: %v", err)
	}
	if runtime.GOOS != "windows" {
		if perm := info.Mode().Perm(); perm != svidFileMode {
			t.Fatalf("SVID document mode = %o, want %o", perm, svidFileMode)
		}
	}
	// The persisted document must parse back to the same identity and verify.
	data, err := os.ReadFile(path) //nolint:gosec // test-controlled temp path
	if err != nil {
		t.Fatalf("read SVID document: %v", err)
	}
	parsed, err := ParseDocument(data)
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	if parsed.ID != id {
		t.Fatalf("persisted identity = %q, want %q", parsed.ID, id)
	}
	if err := parsed.Verify(clk.Now()); err != nil {
		t.Fatalf("persisted SVID must verify: %v", err)
	}
}

func TestManager_RotatesBeforeExpiry(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-rotate")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	base, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	src := &countingSource{inner: base}
	mgr, path := newTestManager(t, src, clk)

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop()

	first := mgr.Current()
	if first == nil {
		t.Fatal("Current must be set after Start")
	}
	if src.count() != 1 {
		t.Fatalf("expected 1 issuance after Start, got %d", src.count())
	}

	// The rotation loop registers one waiter for (TTL - RefreshBefore).
	if err := clk.blockUntilWaiters(1, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	// Advance to the refresh point: 15 min before expiry.
	clk.Advance(DefaultSVIDTTL - DefaultRefreshBefore)

	// Wait for the background loop to publish the rotated SVID.
	waitForRotation(t, mgr, first, 2*time.Second)

	rotated := mgr.Current()
	if rotated.Certificates[0].SerialNumber.Cmp(first.Certificates[0].SerialNumber) == 0 {
		t.Fatal("rotation must replace the certificate")
	}
	// New SVID issued before the old expires (no identity gap).
	if !rotated.IssuedAt.Before(first.ExpiresAt) {
		t.Fatal("rotated SVID must be issued before the previous one expires")
	}
	// Disk reflects the rotated identity and still verifies.
	data, err := os.ReadFile(path) //nolint:gosec // test-controlled temp path
	if err != nil {
		t.Fatalf("read SVID document: %v", err)
	}
	parsed, err := ParseDocument(data)
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	if parsed.Certificates[0].SerialNumber.Cmp(rotated.Certificates[0].SerialNumber) != 0 {
		t.Fatal("disk must reflect the rotated SVID")
	}
	if err := parsed.Verify(clk.Now()); err != nil {
		t.Fatalf("rotated on-disk SVID must verify: %v", err)
	}
}

func TestManager_RotationRetriesAndRetainsCurrent(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-retry")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	base, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	// First rotation attempt (call #2) fails; the retry (call #3) succeeds.
	flaky := &flakySource{failCount: 0, inner: base}
	counter := &countingSource{inner: flaky}
	mgr, _ := newTestManager(t, counter, clk)

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer mgr.Stop()
	first := mgr.Current()

	// Make the next fetch fail to exercise the retry path.
	flaky.mu.Lock()
	flaky.failCount = 1 // the next (rotation) call fails once
	flaky.calls = 0     // reset so failCount applies to the upcoming call
	flaky.mu.Unlock()

	if err := clk.blockUntilWaiters(1, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	clk.Advance(DefaultSVIDTTL - DefaultRefreshBefore)

	// After the failed rotation, the loop schedules a retry waiter.
	if err := clk.blockUntilWaiters(1, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	// During the failure, the previous valid SVID must be retained.
	if mgr.Current().Certificates[0].SerialNumber.Cmp(first.Certificates[0].SerialNumber) != 0 {
		t.Fatal("current SVID must be retained while rotation is failing")
	}
	// Advance the retry interval; the retry should succeed and rotate.
	clk.Advance(mgr.retryInterval)
	waitForRotation(t, mgr, first, 2*time.Second)
}

func waitForRotation(t *testing.T, mgr *Manager, prev *SVID, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cur := mgr.Current()
		if cur != nil && cur.Certificates[0].SerialNumber.Cmp(prev.Certificates[0].SerialNumber) != 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for rotation to replace the SVID")
}

func TestManager_Start_RejectsSecondCallAfterFailure(t *testing.T) {
	// A first Start that fails (SPIRE unreachable) must still consume the
	// single-call guard: a second Start is rejected and never re-attempts
	// issuance, so a failed startup cannot be silently retried on the same
	// Manager.
	src := &countingSource{inner: failingSource{}}
	mgr, _ := newTestManager(t, src, newFakeClock(time.Now()))

	if err := mgr.Start(context.Background()); err == nil {
		t.Fatal("first Start must fail when SPIRE is unreachable")
	}

	err := mgr.Start(context.Background())
	if err == nil {
		t.Fatal("second Start must be rejected even after a failed first Start")
	}
	if !strings.Contains(err.Error(), "already started") {
		t.Fatalf("second Start must report already-started, got %q", err.Error())
	}
	if calls := src.count(); calls != 1 {
		t.Fatalf("second Start must not attempt issuance again; FetchSVID calls = %d, want 1", calls)
	}
}

func TestManager_IssueAndPersist_NoWriteOnCanceledContext(t *testing.T) {
	// Cancellation that lands after the post-fetch check but before the write
	// (driven via the Now() call during verification) must abort without
	// persisting a credential.
	id := testSPIFFEID("engineering", "jpoley", "vm-cancel")
	base := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	// The source uses its own clock so fetch does not trip the cancellation; only
	// the manager's verification Now() does.
	src, err := NewLocalCASource(id, DefaultSVIDTTL, base)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	clk := &cancelOnNowClock{inner: base, cancel: cancel}
	mgr, path := newTestManager(t, src, clk)

	err = mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start must fail when the context is canceled mid-flow")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error must wrap context.Canceled, got %v", err)
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatal("no SVID document must be written when the context is canceled before the write")
	}
	if mgr.Current() != nil {
		t.Fatal("Current() must be nil when issuance was canceled before persist")
	}
}

func TestManager_Start_NilContext_RejectedWithoutConsumingGuard(t *testing.T) {
	// A nil context must be rejected before the single-call guard is consumed:
	// issueAndPersist dereferences ctx, so a nil ctx would panic, and invalid
	// input must not burn the one-shot guard. A subsequent Start with a valid
	// context must still succeed.
	id := testSPIFFEID("engineering", "jpoley", "vm-nilctx")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	mgr, _ := newTestManager(t, src, clk)

	//nolint:staticcheck // SA1012: passing nil is the behavior under test.
	if err := mgr.Start(nil); err == nil {
		t.Fatal("Start(nil) must return an error, not panic")
	} else if !strings.Contains(err.Error(), "non-nil context") {
		t.Fatalf("Start(nil) error must name the non-nil context requirement, got %q", err.Error())
	}
	if mgr.started.Load() {
		t.Fatal("single-call guard must not be consumed by a rejected nil context")
	}

	// A valid retry on the same Manager must still work.
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start with a valid context after a rejected nil ctx must succeed: %v", err)
	}
	defer mgr.Stop()
	if mgr.Current() == nil {
		t.Fatal("Current must be set after the valid Start")
	}
}

func TestRemoveCredential_RemovesAndIsIdempotent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "spiffe")
	if err := os.MkdirAll(dir, svidDirMode); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	name := "svid.json"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("cred"), svidFileMode); err != nil {
		t.Fatalf("write cred: %v", err)
	}

	if err := removeCredential(dir, name); err != nil {
		t.Fatalf("removeCredential: %v", err)
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatal("credential must be removed")
	}
	// The directory itself must survive the removal.
	if _, statErr := os.Stat(dir); statErr != nil {
		t.Fatalf("directory must survive credential removal: %v", statErr)
	}
	// Removing an already-gone credential is a no-op success (goal state met).
	if err := removeCredential(dir, name); err != nil {
		t.Fatalf("removeCredential must be idempotent, got %v", err)
	}
	// A missing directory is also success.
	if err := removeCredential(filepath.Join(dir, "nope"), name); err != nil {
		t.Fatalf("removeCredential on missing dir must be a no-op, got %v", err)
	}
}

func TestManager_Invalidate_RemovesCredentialAndClearsState(t *testing.T) {
	id := testSPIFFEID("engineering", "jpoley", "vm-invalidate")
	clk := newFakeClock(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	src, err := NewLocalCASource(id, DefaultSVIDTTL, clk)
	if err != nil {
		t.Fatalf("NewLocalCASource: %v", err)
	}
	mgr, path := newTestManager(t, src, clk)
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	mgr.Stop()
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("credential must exist before invalidate: %v", statErr)
	}

	if err := mgr.invalidate(); err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatal("invalidate must remove the credential from disk")
	}
	if mgr.Current() != nil {
		t.Fatal("invalidate must clear in-memory state on success")
	}
}

func TestNewManager_Validation(t *testing.T) {
	if _, err := NewManager(ManagerConfig{}); err == nil {
		t.Fatal("expected error when source is nil")
	}
	if _, err := NewManager(ManagerConfig{Source: failingSource{}, Path: "relative/path.json"}); err == nil {
		t.Fatal("expected error for non-absolute path")
	}
}
