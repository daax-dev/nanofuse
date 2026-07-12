package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/types"
)

// TestUpdateVMPreservesLock is a regression guard: UpdateVM must not clear the
// lock columns. AcquireLock writes locked_by/locked_at in the DB but not into
// the in-memory VM struct, so if UpdateVM persisted the struct's (nil) lock
// fields it would silently release a lock the caller is still holding — re-opening
// the races the lock exists to prevent (e.g. snapshot delete vs. resume).
func TestUpdateVMPreservesLock(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	vm := &types.VM{
		ID:        "vm-lock-test",
		Name:      "lock-test",
		State:     types.StateStopped,
		Image:     "img",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateVM(vm); err != nil {
		t.Fatalf("CreateVM: %v", err)
	}

	if err := db.AcquireLock(vm.ID, "resume"); err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}

	// Simulate a state transition under the lock (as resume/start/stop/etc. do).
	vm.State = types.StateResuming
	if err := db.UpdateVM(vm); err != nil {
		t.Fatalf("UpdateVM: %v", err)
	}

	// The lock must still be held: a second AcquireLock must fail.
	if err := db.AcquireLock(vm.ID, "delete"); err == nil {
		t.Fatal("AcquireLock succeeded after UpdateVM; the state update cleared the lock")
	}

	// And the persisted row must still report the lock owner.
	got, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if got.LockedBy == nil || *got.LockedBy != "resume" {
		t.Fatalf("locked_by = %v, want \"resume\" (UpdateVM must not clear the lock)", got.LockedBy)
	}

	// ReleaseLock still clears it, and the state update persisted.
	if err := db.ReleaseLock(vm.ID); err != nil {
		t.Fatalf("ReleaseLock: %v", err)
	}
	got, err = db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM after release: %v", err)
	}
	if got.LockedBy != nil {
		t.Fatalf("locked_by = %v after ReleaseLock, want nil", got.LockedBy)
	}
	if got.State != types.StateResuming {
		t.Fatalf("state = %q, want %q (UpdateVM's state change must persist)", got.State, types.StateResuming)
	}
}
