// Package spire provides SPIRE workload registration integration for nanofuse.
// This file adds per-VM SVID issuance with 1-hour TTL and auto-rotation.
package spire

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jpoley/nanofuse/internal/types"
)

const (
	// DefaultSVIDTTL is the SVID time-to-live: 1 hour expressed in seconds.
	DefaultSVIDTTL = 3600

	// rotationBuffer is the fraction of the TTL remaining that triggers rotation.
	// At 10 minutes before expiry the background rotator will renew the SVID.
	rotationBuffer = 10 * time.Minute
)

// SVID represents a workload SVID issued for a single microVM.
type SVID struct {
	// SpiffeID is the full SPIFFE ID, e.g. spiffe://poley.dev/g/eng/u/alice/w/microvm/abc123
	SpiffeID string
	// VMID is the microVM this SVID was issued for.
	VMID string
	// IssuedAt is when the SVID was first created.
	IssuedAt time.Time
	// ExpiresAt is the absolute expiry time (IssuedAt + TTL).
	ExpiresAt time.Time
	// TTLSeconds is the requested TTL in seconds.
	TTLSeconds int
}

// IsExpired reports whether the SVID has expired.
func (s *SVID) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// NeedsRotation reports whether the SVID is within the rotation buffer window.
func (s *SVID) NeedsRotation() bool {
	return time.Now().After(s.ExpiresAt.Add(-rotationBuffer))
}

// SVIDStore tracks active SVIDs per VM and drives auto-rotation.
type SVIDStore struct {
	svc *Service

	mu    sync.RWMutex
	svids map[string]*SVID // vmID -> SVID

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSVIDStore creates a store and starts the background rotation ticker.
func NewSVIDStore(svc *Service) *SVIDStore {
	ctx, cancel := context.WithCancel(context.Background())
	st := &SVIDStore{
		svc:    svc,
		svids:  make(map[string]*SVID),
		ctx:    ctx,
		cancel: cancel,
	}
	st.wg.Add(1)
	go st.rotationLoop()
	return st
}

// IssueForVM issues a fresh SVID for a VM and embeds it in vm.SpiffeID.
// It also registers a SPIRE workload entry so the VM can actually obtain an X.509 SVID.
// The SVID is stored in the SVIDStore for auto-rotation tracking.
func (st *SVIDStore) IssueForVM(ctx context.Context, vm *types.VM) (*SVID, error) {
	if !st.svc.IsEnabled() {
		// SPIRE disabled – set a placeholder so downstream code can detect this.
		vm.SpiffeID = ""
		return nil, nil
	}

	ttl := st.svc.cfg.DefaultTTL
	if ttl <= 0 {
		ttl = DefaultSVIDTTL
	}

	spiffeID, err := st.svc.CreateVMWorkloadEntry(ctx, vm.ID, vm.GroupID, vm.OwnerUserID)
	if err != nil {
		return nil, fmt.Errorf("issue SVID for VM %s: %w", vm.ID, err)
	}

	now := time.Now()
	svid := &SVID{
		SpiffeID:   spiffeID,
		VMID:       vm.ID,
		IssuedAt:   now,
		ExpiresAt:  now.Add(time.Duration(ttl) * time.Second),
		TTLSeconds: ttl,
	}

	// Embed SVID in VM metadata.
	vm.SpiffeID = spiffeID

	st.mu.Lock()
	st.svids[vm.ID] = svid
	st.mu.Unlock()

	slog.Info("SVID issued for VM",
		slog.String("vm_id", vm.ID),
		slog.String("spiffe_id", spiffeID),
		slog.Time("expires_at", svid.ExpiresAt),
	)

	return svid, nil
}

// RevokeForVM deletes the SPIRE workload entry and removes the SVID from the store.
func (st *SVIDStore) RevokeForVM(ctx context.Context, vm *types.VM) error {
	st.mu.Lock()
	svid, ok := st.svids[vm.ID]
	if ok {
		delete(st.svids, vm.ID)
	}
	st.mu.Unlock()

	if !ok || svid == nil {
		return nil
	}

	if err := st.svc.DeleteVMWorkloadEntry(ctx, svid.SpiffeID); err != nil {
		return fmt.Errorf("revoke SVID for VM %s: %w", vm.ID, err)
	}

	slog.Info("SVID revoked for VM",
		slog.String("vm_id", vm.ID),
		slog.String("spiffe_id", svid.SpiffeID),
	)
	return nil
}

// GetSVID returns the current SVID for a VM, or nil if none exists.
func (st *SVIDStore) GetSVID(vmID string) *SVID {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.svids[vmID]
}

// Close stops the rotation goroutine gracefully.
func (st *SVIDStore) Close() {
	st.cancel()
	st.wg.Wait()
}

// rotationLoop runs every minute and renews SVIDs that are approaching expiry.
func (st *SVIDStore) rotationLoop() {
	defer st.wg.Done()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-st.ctx.Done():
			return
		case <-ticker.C:
			st.rotateExpiring()
		}
	}
}

// rotateExpiring iterates the SVID map and renews any that need rotation.
func (st *SVIDStore) rotateExpiring() {
	st.mu.RLock()
	candidates := make([]*SVID, 0, len(st.svids))
	for _, s := range st.svids {
		if s.NeedsRotation() {
			candidates = append(candidates, s)
		}
	}
	st.mu.RUnlock()

	for _, svid := range candidates {
		if err := st.renewSVID(svid); err != nil {
			slog.Error("SVID auto-rotation failed",
				slog.String("vm_id", svid.VMID),
				slog.String("spiffe_id", svid.SpiffeID),
				slog.Any("error", err),
			)
		}
	}
}

// renewSVID issues a fresh SPIRE entry for an existing SVID and updates the store.
func (st *SVIDStore) renewSVID(old *SVID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ttl := old.TTLSeconds
	if ttl <= 0 {
		ttl = DefaultSVIDTTL
	}

	// Re-register the workload entry to reset the TTL on the SPIRE server side.
	// The SPIFFE ID stays the same, so clients see no disruption.
	if err := st.svc.RegisterWorkload(ctx, &WorkloadEntry{
		SpiffeID:   old.SpiffeID,
		ParentID:   st.svc.cfg.ParentID,
		Selectors:  []string{fmt.Sprintf("docker:label:vm_id:%s", old.VMID)},
		TTL:        ttl,
		WorkloadID: old.VMID,
	}); err != nil {
		return fmt.Errorf("re-register workload entry: %w", err)
	}

	now := time.Now()
	updated := &SVID{
		SpiffeID:   old.SpiffeID,
		VMID:       old.VMID,
		IssuedAt:   now,
		ExpiresAt:  now.Add(time.Duration(ttl) * time.Second),
		TTLSeconds: ttl,
	}

	st.mu.Lock()
	st.svids[old.VMID] = updated
	st.mu.Unlock()

	slog.Info("SVID auto-rotated",
		slog.String("vm_id", old.VMID),
		slog.Time("new_expiry", updated.ExpiresAt),
	)
	return nil
}
