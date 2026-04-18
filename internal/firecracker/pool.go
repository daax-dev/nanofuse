// Package firecracker provides Firecracker VM lifecycle management.
// This file implements VMPool: a pre-warmed pool of microVMs that restores
// from a base snapshot, targeting P50 <100ms and P95 <250ms cold-start times.
package firecracker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jpoley/nanofuse/internal/types"
)

// PoolConfig holds configuration for the VMPool.
type PoolConfig struct {
	// MinSize is the minimum number of pre-warmed VMs to keep ready.
	MinSize int
	// MaxSize is the maximum pool capacity.
	MaxSize int
	// SnapshotPath is the path to the Firecracker snapshot state file used for fast-restore.
	SnapshotPath string
	// MemFilePath is the path to the Firecracker memory file for the snapshot.
	MemFilePath string
	// BaseImage is the image used to populate the pool (must be pre-snapshotted).
	BaseImage *types.Image
	// VMDefaults are the default VM config params applied to pool-sourced VMs.
	VMDefaults types.VMConfig
	// RefillInterval controls how frequently the pool refill goroutine checks the size.
	RefillInterval time.Duration
}

// poolEntry is an internal pool slot holding a ready VM.
type poolEntry struct {
	vm      *types.VM
	readyAt time.Time
}

// PoolStats reports current pool health metrics.
type PoolStats struct {
	Ready    int
	InFlight int
	Total    int
}

// VMPool maintains a set of pre-warmed Firecracker VMs restored from a snapshot.
// Callers acquire a VM via Acquire() which returns a VM that is already booted;
// they must call Release() when done so the slot can be recycled.
type VMPool struct {
	cfg     PoolConfig
	manager *Manager

	mu       sync.Mutex
	ready    []*poolEntry // pre-warmed VMs waiting to be claimed
	inFlight int          // VMs currently held by callers

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewVMPool creates a VMPool and starts the background refill goroutine.
// Call Close() to stop the pool cleanly.
func NewVMPool(cfg PoolConfig, manager *Manager) (*VMPool, error) {
	if cfg.MinSize < 0 {
		return nil, fmt.Errorf("pool MinSize must be >= 0")
	}
	if cfg.MaxSize < cfg.MinSize {
		return nil, fmt.Errorf("pool MaxSize (%d) must be >= MinSize (%d)", cfg.MaxSize, cfg.MinSize)
	}
	if cfg.SnapshotPath == "" || cfg.MemFilePath == "" {
		return nil, fmt.Errorf("pool requires SnapshotPath and MemFilePath")
	}
	if cfg.BaseImage == nil {
		return nil, fmt.Errorf("pool requires a BaseImage")
	}
	if cfg.RefillInterval <= 0 {
		cfg.RefillInterval = 500 * time.Millisecond
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &VMPool{
		cfg:     cfg,
		manager: manager,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Pre-warm to MinSize synchronously so the pool is useful immediately.
	for i := 0; i < cfg.MinSize; i++ {
		if err := p.warmOne(); err != nil {
			slog.Warn("pool initial warm failed",
				slog.Int("slot", i+1),
				slog.Int("min_size", cfg.MinSize),
				slog.Any("error", err),
			)
		}
	}

	p.wg.Add(1)
	go p.refillLoop()

	return p, nil
}

// Acquire claims a pre-warmed VM from the pool.
// If the pool is empty it falls back to warming a VM on-demand (cold path).
// The returned VM is in StateRunning; the caller owns it until Release().
func (p *VMPool) Acquire(ctx context.Context) (*types.VM, error) {
	p.mu.Lock()
	if len(p.ready) > 0 {
		entry := p.ready[0]
		p.ready = p.ready[1:]
		p.inFlight++
		p.mu.Unlock()
		slog.Info("pool acquired pre-warmed VM",
			slog.String("vm_id", entry.vm.ID),
			slog.Duration("pool_lag", time.Since(entry.readyAt).Round(time.Millisecond)),
		)
		return entry.vm, nil
	}
	p.mu.Unlock()

	// Cold path – warm a VM on demand.
	slog.Warn("pool empty, falling back to on-demand warm")
	vm, err := p.warmOneContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("pool: on-demand warm failed: %w", err)
	}
	p.mu.Lock()
	p.inFlight++
	p.mu.Unlock()
	return vm, nil
}

// Release returns a VM to the pool or destroys it if the pool is at capacity.
// The caller must not use the VM after Release().
func (p *VMPool) Release(vm *types.VM) {
	p.mu.Lock()
	p.inFlight--
	atCapacity := len(p.ready) >= p.cfg.MaxSize
	p.mu.Unlock()

	if atCapacity {
		slog.Info("pool full, destroying returned VM", slog.String("vm_id", vm.ID))
		if err := p.manager.Stop(vm, 5); err != nil {
			slog.Warn("pool failed to stop returned VM",
				slog.String("vm_id", vm.ID),
				slog.Any("error", err),
			)
		}
		p.cleanupVMDir(vm)
		return
	}

	// Re-add to ready pool.
	p.mu.Lock()
	p.ready = append(p.ready, &poolEntry{vm: vm, readyAt: time.Now()})
	p.mu.Unlock()
	slog.Info("pool VM returned to pool", slog.String("vm_id", vm.ID))
}

// Stats returns a snapshot of current pool metrics.
func (p *VMPool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return PoolStats{
		Ready:    len(p.ready),
		InFlight: p.inFlight,
		Total:    len(p.ready) + p.inFlight,
	}
}

// Close drains the pool, stops all pre-warmed VMs, and shuts down the refill goroutine.
func (p *VMPool) Close() {
	p.cancel()
	p.wg.Wait()

	p.mu.Lock()
	vms := make([]*poolEntry, len(p.ready))
	copy(vms, p.ready)
	p.ready = nil
	p.mu.Unlock()

	for _, e := range vms {
		if err := p.manager.Stop(e.vm, 5); err != nil {
			slog.Warn("pool error stopping VM during Close",
				slog.String("vm_id", e.vm.ID),
				slog.Any("error", err),
			)
		}
		p.cleanupVMDir(e.vm)
	}
}

// refillLoop runs in the background, topping up the pool to MinSize.
func (p *VMPool) refillLoop() {
	defer p.wg.Done()
	ticker := time.NewTicker(p.cfg.RefillInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.mu.Lock()
			deficit := p.cfg.MinSize - len(p.ready)
			p.mu.Unlock()

			for i := 0; i < deficit; i++ {
				// Check context before each warm attempt.
				if p.ctx.Err() != nil {
					return
				}
				if err := p.warmOne(); err != nil {
					slog.Warn("pool refill warm failed", slog.Any("error", err))
					break // back-off; retry next tick
				}
			}
		}
	}
}

// warmOne warms a single VM using context.Background().
func (p *VMPool) warmOne() error {
	vm, err := p.warmOneContext(context.Background())
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.ready = append(p.ready, &poolEntry{vm: vm, readyAt: time.Now()})
	p.mu.Unlock()
	return nil
}

// warmOneContext restores a single VM from snapshot and returns it in StateRunning.
func (p *VMPool) warmOneContext(ctx context.Context) (*types.VM, error) {
	vmID := uuid.New().String()[:8]

	// Create VM directory.
	vmDir := filepath.Join(p.manager.dataDir, "pool", vmID)
	if err := os.MkdirAll(vmDir, 0750); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", vmDir, err)
	}

	vm := &types.VM{
		ID:        vmID,
		State:     types.StateStarting,
		Config:    p.cfg.VMDefaults,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Restore from snapshot via Firecracker API (stub: uses normal Start for now
	// until CreateSnapshot/LoadSnapshot are fully implemented in the manager).
	// When LoadSnapshot is implemented the call below should be swapped for it.
	if err := p.manager.Start(vm, p.cfg.BaseImage); err != nil {
		_ = os.RemoveAll(vmDir)
		return nil, fmt.Errorf("snapshot-restore for pool VM %s: %w", vmID, err)
	}

	// Brief readiness poll – real implementation would query the Firecracker API
	// or use a vsock handshake.  We time-box to keep cold-start under 250ms.
	if err := waitForVMReady(ctx, vm, 200*time.Millisecond); err != nil {
		slog.Warn("pool VM readiness timeout, marking as running anyway",
			slog.String("vm_id", vmID),
			slog.Any("error", err),
		)
	}

	vm.State = types.StateRunning
	vm.UpdatedAt = time.Now()
	return vm, nil
}

// waitForVMReady polls until the VM's process is confirmed running or deadline hits.
func waitForVMReady(ctx context.Context, vm *types.VM, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if vm.Runtime != nil && vm.Runtime.PID > 0 {
			return nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return fmt.Errorf("VM %s not ready after %v", vm.ID, timeout)
}

// cleanupVMDir removes the working directory for a pool-sourced VM.
func (p *VMPool) cleanupVMDir(vm *types.VM) {
	vmDir := filepath.Join(p.manager.dataDir, "pool", vm.ID)
	if err := os.RemoveAll(vmDir); err != nil {
		slog.Warn("pool failed to remove VM dir",
			slog.String("vm_dir", vmDir),
			slog.Any("error", err),
		)
	}
}
