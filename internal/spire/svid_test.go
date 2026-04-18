package spire

import (
	"testing"
	"time"

	"github.com/jpoley/nanofuse/internal/config"
	"github.com/jpoley/nanofuse/internal/types"
)

// makeTestVM creates a minimal types.VM for unit tests.
func makeTestVM(id string) *types.VM {
	return &types.VM{
		ID:          id,
		OwnerUserID: "testuser",
		GroupID:     "testgroup",
	}
}

// newDisabledService creates a SPIRE service with integration disabled.
func newDisabledService() *Service {
	return NewService(&config.SPIREConfig{
		Enabled:     false,
		TrustDomain: "test.example",
		ParentID:    "spiffe://test.example/agent",
		DefaultTTL:  3600,
	})
}

// TestSVIDIsExpired verifies that IsExpired reflects the expiry time correctly.
func TestSVIDIsExpired(t *testing.T) {
	past := &SVID{
		SpiffeID:  "spiffe://test.example/vm/old",
		VMID:      "old-vm",
		IssuedAt:  time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if !past.IsExpired() {
		t.Error("expected past SVID to be expired")
	}

	future := &SVID{
		SpiffeID:  "spiffe://test.example/vm/new",
		VMID:      "new-vm",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if future.IsExpired() {
		t.Error("expected future SVID to not be expired")
	}
}

// TestSVIDNeedsRotation verifies the rotation buffer logic.
func TestSVIDNeedsRotation(t *testing.T) {
	// SVID that expires in 5 minutes — within the 10-minute rotation buffer.
	soonExpiring := &SVID{
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if !soonExpiring.NeedsRotation() {
		t.Error("expected SVID expiring in 5m to need rotation (buffer is 10m)")
	}

	// SVID that expires in 30 minutes — outside the rotation buffer.
	safeExpiry := &SVID{
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if safeExpiry.NeedsRotation() {
		t.Error("expected SVID expiring in 30m to not need rotation yet")
	}
}

// TestNewSVIDStore creates a store and checks it initialises cleanly.
func TestNewSVIDStore(t *testing.T) {
	svc := newDisabledService()
	store := NewSVIDStore(svc)
	defer store.Close()

	if store.svids == nil {
		t.Error("expected svids map to be initialised")
	}
}

// TestSVIDStoreGetMissing returns nil for an unknown vmID.
func TestSVIDStoreGetMissing(t *testing.T) {
	svc := newDisabledService()
	store := NewSVIDStore(svc)
	defer store.Close()

	if got := store.GetSVID("does-not-exist"); got != nil {
		t.Errorf("expected nil for unknown vmID, got %+v", got)
	}
}

// TestSVIDStoreIssueDisabled verifies that when SPIRE is disabled, IssueForVM is a no-op.
func TestSVIDStoreIssueDisabled(t *testing.T) {
	svc := newDisabledService()
	store := NewSVIDStore(svc)
	defer store.Close()

	vm := makeTestVM("vm-disabled")
	svid, err := store.IssueForVM(t.Context(), vm)
	if err != nil {
		t.Fatalf("unexpected error when SPIRE disabled: %v", err)
	}
	if svid != nil {
		t.Errorf("expected nil SVID when SPIRE disabled, got %+v", svid)
	}
	if vm.SpiffeID != "" {
		t.Errorf("expected empty SpiffeID when SPIRE disabled, got %q", vm.SpiffeID)
	}
}

// TestSVIDStoreRevokeNoOp verifies revoke is safe when no SVID exists.
func TestSVIDStoreRevokeNoOp(t *testing.T) {
	svc := newDisabledService()
	store := NewSVIDStore(svc)
	defer store.Close()

	vm := makeTestVM("vm-no-svid")
	if err := store.RevokeForVM(t.Context(), vm); err != nil {
		t.Errorf("RevokeForVM with no SVID should be no-op, got: %v", err)
	}
}

// TestSVIDStoreManualInsertAndGet exercises the store's Get path with a manually inserted SVID.
func TestSVIDStoreManualInsertAndGet(t *testing.T) {
	svc := newDisabledService()
	store := NewSVIDStore(svc)
	defer store.Close()

	svid := &SVID{
		SpiffeID:   "spiffe://test.example/vm/abc",
		VMID:       "abc",
		IssuedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
		TTLSeconds: 3600,
	}

	store.mu.Lock()
	store.svids["abc"] = svid
	store.mu.Unlock()

	got := store.GetSVID("abc")
	if got == nil {
		t.Fatal("expected to retrieve inserted SVID, got nil")
	}
	if got.SpiffeID != svid.SpiffeID {
		t.Errorf("SpiffeID mismatch: got %q, want %q", got.SpiffeID, svid.SpiffeID)
	}
}

// TestDefaultSVIDTTL checks the constant value matches the 1-hour requirement.
func TestDefaultSVIDTTL(t *testing.T) {
	const expected = 3600
	if DefaultSVIDTTL != expected {
		t.Errorf("DefaultSVIDTTL = %d, want %d (1 hour)", DefaultSVIDTTL, expected)
	}
}
