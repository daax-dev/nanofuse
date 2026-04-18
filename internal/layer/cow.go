// Package layer provides functionality for creating and validating NanoFuse layer definitions.
// This file implements CowLayer: a Copy-on-Write filesystem layer using Linux overlayfs.
// A shared base image (lowerdir) is mounted read-only; each VM session gets its own
// writable upperdir + workdir so 10 concurrent VMs consume <2x the storage of 1 VM.
package layer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CowOptions configures a single CoW session.
type CowOptions struct {
	// BaseDir is the shared read-only lower layer (base rootfs image directory).
	BaseDir string
	// SessionDir is the root under which per-session upper/work dirs are created.
	// A sub-directory named after SessionID will be created here.
	SessionDir string
	// SessionID uniquely identifies this VM session (e.g. VM ID).
	SessionID string
	// MountDir is where the overlayfs will be mounted (the merged view).
	// If empty it defaults to <SessionDir>/<SessionID>/merged.
	MountDir string
}

// CowLayer represents a mounted overlayfs CoW layer for a single VM session.
type CowLayer struct {
	opts      CowOptions
	upperDir  string // per-session writable layer
	workDir   string // overlayfs work directory (must be on same FS as upperDir)
	mergedDir string // the final merged mount point
	mounted   bool
	createdAt time.Time
}

// NewCowLayer creates the directory structure for a CoW session and mounts the overlayfs.
// Callers must call Destroy() when the session ends to unmount and clean up.
func NewCowLayer(opts CowOptions) (*CowLayer, error) {
	if opts.BaseDir == "" {
		return nil, fmt.Errorf("cow: BaseDir is required")
	}
	if opts.SessionDir == "" {
		return nil, fmt.Errorf("cow: SessionDir is required")
	}
	if opts.SessionID == "" {
		return nil, fmt.Errorf("cow: SessionID is required")
	}

	// Verify base dir exists.
	if _, err := os.Stat(opts.BaseDir); err != nil {
		return nil, fmt.Errorf("cow: base dir %q not found: %w", opts.BaseDir, err)
	}

	sessionRoot := filepath.Join(opts.SessionDir, opts.SessionID)
	upperDir := filepath.Join(sessionRoot, "upper")
	workDir := filepath.Join(sessionRoot, "work")
	mergedDir := opts.MountDir
	if mergedDir == "" {
		mergedDir = filepath.Join(sessionRoot, "merged")
	}

	for _, dir := range []string{upperDir, workDir, mergedDir} {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("cow: create dir %s: %w", dir, err)
		}
	}

	cl := &CowLayer{
		opts:      opts,
		upperDir:  upperDir,
		workDir:   workDir,
		mergedDir: mergedDir,
		createdAt: time.Now(),
	}

	if err := cl.mount(); err != nil {
		// Clean up on failure.
		_ = os.RemoveAll(sessionRoot)
		return nil, err
	}

	return cl, nil
}

// mount performs the overlayfs mount.
// lowerdir=<base>,upperdir=<upper>,workdir=<work>  →  mergedDir
func (cl *CowLayer) mount() error {
	mountOpts := fmt.Sprintf(
		"lowerdir=%s,upperdir=%s,workdir=%s",
		cl.opts.BaseDir,
		cl.upperDir,
		cl.workDir,
	)

	// #nosec G204 — arguments are constructed from validated path variables, not user input
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", mountOpts, cl.mergedDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cow: overlayfs mount failed: %w: %s", err, string(out))
	}

	cl.mounted = true
	log.Printf("INFO [cow] overlayfs mounted for session %s at %s", cl.opts.SessionID, cl.mergedDir)
	return nil
}

// MergedDir returns the path to the merged overlayfs mount point.
// This is the path that should be used as the VM's root filesystem.
func (cl *CowLayer) MergedDir() string {
	return cl.mergedDir
}

// UpperDir returns the per-session writable layer path.
// This directory contains only the files that have been modified or created
// in this session (CoW semantics).
func (cl *CowLayer) UpperDir() string {
	return cl.upperDir
}

// Snapshot creates a point-in-time snapshot of the current session state by
// copying the upperdir (session delta) to destDir.  This must complete within 5 seconds.
// Only the changed files (upperdir) are snapshotted, so it is fast and storage-efficient.
func (cl *CowLayer) Snapshot(destDir string) error {
	if !cl.mounted {
		return fmt.Errorf("cow: layer for session %s is not mounted", cl.opts.SessionID)
	}
	if destDir == "" {
		return fmt.Errorf("cow: snapshot destDir is required")
	}

	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("cow: create snapshot dir %s: %w", destDir, err)
	}

	// Use cp -a to preserve permissions and timestamps.
	// #nosec G204 — arguments are constructed from validated path variables
	cmd := exec.Command("cp", "-a", cl.upperDir+"/.", destDir)

	done := make(chan error, 1)
	go func() {
		out, err := cmd.CombinedOutput()
		if err != nil {
			done <- fmt.Errorf("cow: snapshot copy failed: %w: %s", err, string(out))
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		return fmt.Errorf("cow: snapshot for session %s timed out after 5s", cl.opts.SessionID)
	}

	log.Printf("INFO [cow] snapshot taken for session %s → %s", cl.opts.SessionID, destDir)
	return nil
}

// Destroy unmounts the overlayfs and removes the per-session directories (upper, work, merged).
// The shared base (lowerdir) is left untouched.
func (cl *CowLayer) Destroy() error {
	if cl.mounted {
		// #nosec G204 — mergedDir is an internally constructed path
		cmd := exec.Command("umount", cl.mergedDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("cow: umount %s failed: %w: %s", cl.mergedDir, err, string(out))
		}
		cl.mounted = false
		log.Printf("INFO [cow] overlayfs unmounted for session %s", cl.opts.SessionID)
	}

	sessionRoot := filepath.Join(cl.opts.SessionDir, cl.opts.SessionID)
	if err := os.RemoveAll(sessionRoot); err != nil {
		return fmt.Errorf("cow: cleanup session dir %s: %w", sessionRoot, err)
	}

	log.Printf("INFO [cow] session %s cleaned up", cl.opts.SessionID)
	return nil
}

// IsMounted reports whether the overlayfs is currently mounted.
func (cl *CowLayer) IsMounted() bool {
	return cl.mounted
}

// SessionAge returns how long this CoW session has been alive.
func (cl *CowLayer) SessionAge() time.Duration {
	return time.Since(cl.createdAt)
}
