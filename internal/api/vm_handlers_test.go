package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daax-dev/nanofuse/internal/types"
)

func TestMaterializeWritableRootDisksCopiesRootfsPerVM(t *testing.T) {
	dataDir := t.TempDir()
	imageDir := t.TempDir()
	sourceRootfs := filepath.Join(imageDir, "source-rootfs.ext4")
	if err := os.WriteFile(sourceRootfs, []byte("source-rootfs"), 0600); err != nil {
		t.Fatalf("write source rootfs: %v", err)
	}

	cfg := types.VMConfig{
		Disks: []types.DiskConfig{
			{
				DriveID:      "rootfs",
				PathOnHost:   sourceRootfs,
				IsReadOnly:   false,
				IsRootDevice: true,
			},
		},
	}

	if err := materializeWritableRootDisks(dataDir, "vm-123", &cfg); err != nil {
		t.Fatalf("materialize rootfs: %v", err)
	}

	wantPath := filepath.Join(dataDir, "vms", "vm-123", "rootfs.ext4")
	if cfg.Disks[0].PathOnHost != wantPath {
		t.Fatalf("root disk path = %q, want %q", cfg.Disks[0].PathOnHost, wantPath)
	}

	got, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read VM rootfs: %v", err)
	}
	if string(got) != "source-rootfs" {
		t.Fatalf("VM rootfs contents = %q, want source copy", got)
	}

	source, err := os.ReadFile(sourceRootfs)
	if err != nil {
		t.Fatalf("read source rootfs: %v", err)
	}
	if string(source) != "source-rootfs" {
		t.Fatalf("source rootfs mutated: %q", source)
	}
}

func TestMaterializeWritableRootDisksPreservesExistingVMDisk(t *testing.T) {
	dataDir := t.TempDir()
	imageDir := t.TempDir()
	sourceRootfs := filepath.Join(imageDir, "source-rootfs.ext4")
	if err := os.WriteFile(sourceRootfs, []byte("source-rootfs"), 0600); err != nil {
		t.Fatalf("write source rootfs: %v", err)
	}

	existingRootfs := vmRootfsPath(dataDir, "vm-123")
	if err := os.MkdirAll(filepath.Dir(existingRootfs), 0700); err != nil {
		t.Fatalf("create VM storage: %v", err)
	}
	if err := os.WriteFile(existingRootfs, []byte("persisted-state"), 0600); err != nil {
		t.Fatalf("write existing rootfs: %v", err)
	}

	cfg := types.VMConfig{
		Disks: []types.DiskConfig{
			{
				DriveID:      "rootfs",
				PathOnHost:   sourceRootfs,
				IsReadOnly:   false,
				IsRootDevice: true,
			},
		},
	}

	if err := materializeWritableRootDisks(dataDir, "vm-123", &cfg); err != nil {
		t.Fatalf("materialize rootfs: %v", err)
	}

	got, err := os.ReadFile(existingRootfs)
	if err != nil {
		t.Fatalf("read existing rootfs: %v", err)
	}
	if string(got) != "persisted-state" {
		t.Fatalf("existing VM rootfs overwritten: %q", got)
	}
}

func TestMaterializeWritableRootDisksSkipsReadOnlyRootfs(t *testing.T) {
	dataDir := t.TempDir()
	sourceRootfs := filepath.Join(t.TempDir(), "source-rootfs.ext4")
	if err := os.WriteFile(sourceRootfs, []byte("source-rootfs"), 0600); err != nil {
		t.Fatalf("write source rootfs: %v", err)
	}

	cfg := types.VMConfig{
		Disks: []types.DiskConfig{
			{
				DriveID:      "rootfs",
				PathOnHost:   sourceRootfs,
				IsReadOnly:   true,
				IsRootDevice: true,
			},
		},
	}

	if err := materializeWritableRootDisks(dataDir, "vm-123", &cfg); err != nil {
		t.Fatalf("materialize rootfs: %v", err)
	}
	if cfg.Disks[0].PathOnHost != sourceRootfs {
		t.Fatalf("read-only rootfs path changed to %q", cfg.Disks[0].PathOnHost)
	}
	if _, err := os.Stat(vmRootfsPath(dataDir, "vm-123")); !os.IsNotExist(err) {
		t.Fatalf("read-only rootfs copy exists or stat failed: %v", err)
	}
}

func TestCleanupVMStorageRemovesVMDirectory(t *testing.T) {
	dataDir := t.TempDir()
	vmDir := vmStorageDir(dataDir, "vm-123")
	if err := os.MkdirAll(vmDir, 0700); err != nil {
		t.Fatalf("create VM dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "rootfs.ext4"), []byte("state"), 0600); err != nil {
		t.Fatalf("write VM state: %v", err)
	}

	if err := cleanupVMStorage(dataDir, "vm-123"); err != nil {
		t.Fatalf("cleanup VM storage: %v", err)
	}
	if _, err := os.Stat(vmDir); !os.IsNotExist(err) {
		t.Fatalf("VM storage still exists or stat failed: %v", err)
	}
}
