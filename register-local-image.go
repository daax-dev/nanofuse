package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/jpoley/nanofuse/internal/storage"
	"github.com/jpoley/nanofuse/internal/types"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: register-local-image <db-path> <tag> <rootfs-path> <kernel-path> [architecture]")
		fmt.Println("Example: register-local-image /tmp/nanofuse/nanofuse.db nanofuse-base:latest /tmp/nanofuse/images/nanofuse-base/latest/rootfs.ext4 /tmp/nanofuse/images/nanofuse-base/latest/vmlinux x86_64")
		os.Exit(1)
	}

	dbPath := os.Args[1]
	tag := os.Args[2]
	rootfsPath := os.Args[3]
	kernelPath := os.Args[4]

	// Default architecture to x86_64 if not specified
	architecture := "x86_64"
	if len(os.Args) > 5 {
		architecture = os.Args[5]
	}

	// Resolve symlinks to get actual file paths
	resolvedRootfs, err := filepath.EvalSymlinks(rootfsPath)
	if err != nil {
		resolvedRootfs = rootfsPath // Fall back to original if symlink resolution fails
	}
	resolvedKernel, err := filepath.EvalSymlinks(kernelPath)
	if err != nil {
		resolvedKernel = kernelPath
	}

	// Open database
	db, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if rootfs and kernel exist
	if _, err := os.Stat(resolvedRootfs); os.IsNotExist(err) {
		log.Fatalf("Rootfs not found: %s", resolvedRootfs)
	}
	if _, err := os.Stat(resolvedKernel); os.IsNotExist(err) {
		log.Fatalf("Kernel not found: %s", resolvedKernel)
	}

	// Get file sizes
	rootfsInfo, _ := os.Stat(resolvedRootfs)
	kernelInfo, _ := os.Stat(resolvedKernel)
	totalSize := rootfsInfo.Size() + kernelInfo.Size()

	// Generate digest from rootfs file
	digest, err := generateDigest(resolvedRootfs)
	if err != nil {
		log.Fatalf("Failed to generate digest: %v", err)
	}

	// Extract kernel version from filename (e.g., vmlinux-5.10.245-no-acpi -> 5.10.245)
	kernelVersion := extractKernelVersion(filepath.Base(resolvedKernel))

	// Create image record
	image := &types.Image{
		Digest:        digest,
		Tags:          []string{tag},
		Architecture:  architecture,
		SizeBytes:     totalSize,
		KernelVersion: kernelVersion,
		RootfsPath:    resolvedRootfs,
		KernelPath:    resolvedKernel,
		PulledAt:      time.Now(),
	}

	// Upsert into database (idempotent - can run multiple times)
	if err := db.UpsertImage(image); err != nil {
		log.Fatalf("Failed to register image: %v", err)
	}

	fmt.Printf("✓ Registered image: %s\n", tag)
	fmt.Printf("  Digest: %s\n", digest)
	fmt.Printf("  Rootfs: %s (%d bytes)\n", rootfsPath, rootfsInfo.Size())
	fmt.Printf("  Kernel: %s (%d bytes)\n", kernelPath, kernelInfo.Size())
}

func generateDigest(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// extractKernelVersion extracts version from kernel filename
// e.g., "vmlinux-5.10.245-no-acpi" -> "5.10.245"
// e.g., "vmlinux-6.1.155" -> "6.1.155"
// e.g., "vmlinux.bin" -> "unknown"
func extractKernelVersion(filename string) string {
	// Match version pattern like 5.10.245 or 6.1.155
	re := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}
