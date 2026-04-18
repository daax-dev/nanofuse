package firecracker

import (
	"testing"
	"time"

	"github.com/jpoley/nanofuse/internal/types"
)

// poolConfigForTest returns a minimal valid PoolConfig for unit tests.
// Note: BaseImage is non-nil but Start/LoadSnapshot are never called because
// warmOne is not invoked (MinSize=0 in cold-path tests).
func poolConfigForTest() PoolConfig {
	return PoolConfig{
		MinSize:        0,
		MaxSize:        5,
		SnapshotPath:   "/tmp/test.snap",
		MemFilePath:    "/tmp/test.mem",
		BaseImage:      &types.Image{KernelPath: "/vmlinux", RootfsPath: "/rootfs.ext4"},
		RefillInterval: 10 * time.Second, // long interval so refill never fires during test
	}
}

// TestPoolConfigValidation checks that NewVMPool rejects bad configs without touching the FS.
func TestPoolConfigValidation(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", t.TempDir())

	tests := []struct {
		name    string
		mutate  func(*PoolConfig)
		wantErr bool
	}{
		{
			name:    "negative MinSize",
			mutate:  func(c *PoolConfig) { c.MinSize = -1 },
			wantErr: true,
		},
		{
			name:    "MaxSize < MinSize",
			mutate:  func(c *PoolConfig) { c.MaxSize = 0; c.MinSize = 2 },
			wantErr: true,
		},
		{
			name:    "missing SnapshotPath",
			mutate:  func(c *PoolConfig) { c.SnapshotPath = "" },
			wantErr: true,
		},
		{
			name:    "missing MemFilePath",
			mutate:  func(c *PoolConfig) { c.MemFilePath = "" },
			wantErr: true,
		},
		{
			name:    "missing BaseImage",
			mutate:  func(c *PoolConfig) { c.BaseImage = nil },
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := poolConfigForTest()
			tc.mutate(&cfg)
			_, err := NewVMPool(cfg, manager)
			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestPoolStats verifies the Stats method arithmetic is correct.
func TestPoolStats(t *testing.T) {
	// Build pool with zero MinSize so no VMs are warmed on startup.
	cfg := poolConfigForTest()
	manager := NewManager("/usr/bin/firecracker", t.TempDir())

	pool, err := NewVMPool(cfg, manager)
	if err != nil {
		t.Fatalf("NewVMPool: %v", err)
	}
	defer pool.Close()

	stats := pool.Stats()
	if stats.Ready != 0 {
		t.Errorf("expected 0 ready, got %d", stats.Ready)
	}
	if stats.InFlight != 0 {
		t.Errorf("expected 0 in-flight, got %d", stats.InFlight)
	}
	if stats.Total != 0 {
		t.Errorf("expected 0 total, got %d", stats.Total)
	}
}

// TestPoolReleaseIncreasesReady verifies that releasing a VM back to the pool
// increments the ready count (when pool is below MaxSize).
func TestPoolReleaseIncreasesReady(t *testing.T) {
	cfg := poolConfigForTest()
	manager := NewManager("/usr/bin/firecracker", t.TempDir())

	pool, err := NewVMPool(cfg, manager)
	if err != nil {
		t.Fatalf("NewVMPool: %v", err)
	}
	defer pool.Close()

	// Inject a fake VM directly into the pool (bypass warmOne which needs FC binary).
	fakeVM := &types.VM{
		ID:    "fake-vm-01",
		State: types.StateRunning,
	}
	pool.mu.Lock()
	pool.ready = append(pool.ready, &poolEntry{vm: fakeVM, readyAt: time.Now()})
	pool.mu.Unlock()

	// Acquire it.
	pool.mu.Lock()
	entry := pool.ready[0]
	pool.ready = pool.ready[1:]
	pool.inFlight++
	pool.mu.Unlock()

	stats := pool.Stats()
	if stats.InFlight != 1 {
		t.Errorf("expected 1 in-flight, got %d", stats.InFlight)
	}

	// Return it without calling manager.Stop (no FC binary in test).
	pool.mu.Lock()
	pool.inFlight--
	pool.ready = append(pool.ready, &poolEntry{vm: entry.vm, readyAt: time.Now()})
	pool.mu.Unlock()

	stats = pool.Stats()
	if stats.Ready != 1 {
		t.Errorf("expected 1 ready after release, got %d", stats.Ready)
	}
	if stats.InFlight != 0 {
		t.Errorf("expected 0 in-flight after release, got %d", stats.InFlight)
	}
}

// TestPoolDefaultRefillInterval verifies zero RefillInterval is replaced by default.
func TestPoolDefaultRefillInterval(t *testing.T) {
	cfg := poolConfigForTest()
	cfg.RefillInterval = 0 // should be replaced with 500ms default
	manager := NewManager("/usr/bin/firecracker", t.TempDir())

	pool, err := NewVMPool(cfg, manager)
	if err != nil {
		t.Fatalf("NewVMPool: %v", err)
	}
	defer pool.Close()

	if pool.cfg.RefillInterval != 500*time.Millisecond {
		t.Errorf("expected default RefillInterval 500ms, got %v", pool.cfg.RefillInterval)
	}
}

// TestWaitForVMReady_AlreadyReady ensures waitForVMReady returns immediately when PID is set.
func TestWaitForVMReady_AlreadyReady(t *testing.T) {
	vm := &types.VM{
		ID:      "test-vm",
		Runtime: &types.VMRuntime{PID: 1234},
	}
	err := waitForVMReady(t.Context(), vm, 100*time.Millisecond)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// TestWaitForVMReady_Timeout ensures waitForVMReady returns error when PID never set.
func TestWaitForVMReady_Timeout(t *testing.T) {
	vm := &types.VM{ID: "never-ready"}
	err := waitForVMReady(t.Context(), vm, 20*time.Millisecond)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}
