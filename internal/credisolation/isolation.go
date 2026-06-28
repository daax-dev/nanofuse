// Package credisolation enforces and verifies per-microVM credential isolation
// for nanofuse.
//
// Threat model: the LiteLLM credential-vault exploit (2026-04-07) succeeded
// partly because a compromised tenant could reach credentials belonging to
// other tenants that shared a storage location. Nanofuse runs many concurrent
// microVMs; if VM1 is compromised it must expose only VM1's SVID/credentials,
// never VM2's.
//
// Nanofuse delivers credentials guest-side over a per-VM vsock proxy to the
// SPIRE agent (see internal/firecracker/vsock_proxy.go); each guest writes its
// own SVID into a guest-local store at GuestSecretsDir. There is no host-path
// shared mount into guests. This package codifies and verifies the invariants
// that keep that isolation intact:
//
//  1. The credential store has mode 0700 owned by root:root (no group/world).
//  2. No host or shared-volume mount may target the credential store path.
//  3. Each VM has a distinct SPIFFE identity (no cross-VM impersonation).
//  4. Any detected cross-VM credential access is audited and fails safe by
//     terminating the offending VM (see monitor.go).
//
// The host-side invariants (1-3, plus the audit/terminate policy) are unit
// testable on any host. The kernel-LSM denial of an in-guest cross-read and the
// termination of a live Firecracker VM are runtime-enforced and are not
// exercised by the unit harness.
package credisolation

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	// GuestSecretsDir is the canonical in-guest path that holds a microVM's own
	// credentials (SVID, tokens). It is local to each guest filesystem and must
	// never be backed by a host or shared mount.
	GuestSecretsDir = "/var/run/secrets/daax" //nolint:gosec // G101: a filesystem path, not a credential

	// RequiredDirMode is the only permitted mode for the credential store:
	// owner read/write/execute, no group or world access.
	RequiredDirMode fs.FileMode = 0o700

	// RequiredOwnerUID and RequiredOwnerGID are the required owner/group of the
	// credential store (root:root).
	RequiredOwnerUID = 0
	RequiredOwnerGID = 0

	// maxVMIDLen bounds VM identifier length to keep derived paths sane.
	maxVMIDLen = 128

	statusPassLine        = "credential isolation: PASS" //nolint:gosec // G101: a status string, not a credential
	statusFailLine        = "credential isolation: FAIL"
	statusNotVerifiedLine = "credential isolation: NOT VERIFIED"
)

// Sentinel errors. Callers should test with errors.Is.
var (
	// ErrCrossVMAccess indicates one VM attempted to reach another VM's
	// credential store.
	ErrCrossVMAccess = errors.New("cross-VM credential access denied")

	// ErrSharedSecretsMount indicates a proposed mount would back the credential
	// store with a host path or shared volume.
	ErrSharedSecretsMount = errors.New("shared/host mount targeting credential store is forbidden")

	// ErrInvalidMountTarget indicates a mount target the guard cannot reason
	// about safely (a non-empty, non-absolute target). The guard fails closed on
	// these rather than assume a caller validated them.
	ErrInvalidMountTarget = errors.New("mount target must be an absolute path")

	// ErrInvalidVMID indicates a VM identifier failed validation.
	ErrInvalidVMID = errors.New("invalid VM identifier")
)

// safeIDPattern matches identifiers containing only alphanumerics, hyphen, and
// underscore — the same character class internal/spire accepts, so SPIFFE IDs
// and isolation checks agree on which characters are legal. ValidateVMID is
// stricter than internal/spire in one respect: it also bounds the length
// (maxVMIDLen) since a VM identifier here flows into paths and audit records.
var safeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateVMID validates a VM identifier at the trust boundary. A VM identifier
// flows into paths and audit records, so it must be non-empty, bounded, and free
// of separators or shell-significant characters.
func ValidateVMID(vmID string) error {
	if vmID == "" {
		return fmt.Errorf("%w: empty", ErrInvalidVMID)
	}
	if len(vmID) > maxVMIDLen {
		return fmt.Errorf("%w: length %d exceeds %d", ErrInvalidVMID, len(vmID), maxVMIDLen)
	}
	if !safeIDPattern.MatchString(vmID) {
		return fmt.Errorf("%w: %q contains characters outside [a-zA-Z0-9_-]", ErrInvalidVMID, vmID)
	}
	return nil
}

// VerifyResult is the outcome of a single isolation check.
type VerifyResult struct {
	Name   string
	Pass   bool
	Detail string
}

// Report aggregates the results of a host-side isolation verification sweep.
type Report struct {
	Results []VerifyResult

	// Subjects counts the real-world checks that ran (store permissions,
	// proposed mounts, VM identities). The always-present policy self-check is
	// not a subject: passing it alone proves the guard's own logic, not the
	// isolation of any actual VM. A report with zero subjects has verified
	// nothing concrete and must not report PASS.
	Subjects int
}

// allPass reports whether every recorded check passed.
func (r Report) allPass() bool {
	if len(r.Results) == 0 {
		return false
	}
	for _, res := range r.Results {
		if !res.Pass {
			return false
		}
	}
	return true
}

// Pass reports whether at least one real subject was checked and every check
// passed. It is deliberately false when nothing concrete was verified.
func (r Report) Pass() bool {
	return r.Subjects > 0 && r.allPass()
}

// HasFailure reports whether any recorded check actively failed. It is distinct
// from !Pass(): a report that verified nothing concrete (Subjects == 0) has no
// failure but does not Pass. Callers that must treat "nothing verified" as a
// non-error (e.g. a lenient health check) use HasFailure for the exit decision.
func (r Report) HasFailure() bool {
	for _, res := range r.Results {
		if !res.Pass {
			return true
		}
	}
	return false
}

// StatusLine returns the canonical status line consumed by status tooling:
// "credential isolation: PASS" when at least one subject was verified and all
// checks passed, "credential isolation: NOT VERIFIED" when checks passed but no
// real subject was inspected, otherwise the FAIL form.
func (r Report) StatusLine() string {
	if !r.allPass() {
		return statusFailLine
	}
	if r.Subjects == 0 {
		return statusNotVerifiedLine
	}
	return statusPassLine
}

// Authorize reports whether the VM identified by requestingVMID may access the
// credential store of targetVMID. Access is permitted only when a VM accesses
// its own store; every cross-VM access is denied. This is the authorization
// predicate the runtime consults; it is intentionally deny-by-default.
func Authorize(requestingVMID, targetVMID string) error {
	if err := ValidateVMID(requestingVMID); err != nil {
		// Wrap both sentinels so callers can distinguish an invalid ID
		// (ErrInvalidVMID) from a plain cross-VM denial via errors.Is.
		return fmt.Errorf("%w: requesting vm: %w", ErrCrossVMAccess, err)
	}
	if err := ValidateVMID(targetVMID); err != nil {
		return fmt.Errorf("%w: target vm: %w", ErrCrossVMAccess, err)
	}
	if requestingVMID != targetVMID {
		return fmt.Errorf("%w: VM %q may not access VM %q credentials",
			ErrCrossVMAccess, requestingVMID, targetVMID)
	}
	return nil
}

// MountSpec is a dependency-free description of a mount or secret binding that
// another subsystem proposes to apply to a microVM. The guard inspects only the
// fields relevant to credential isolation, so callers can adapt their own mount
// representation without coupling to this package.
type MountSpec struct {
	Target string // in-guest absolute path
	Source string // host path or volume name; empty for tmpfs
	Type   string // "bind" | "volume" | "tmpfs" | ...
}

// GuardMounts rejects any proposed mount that would shadow, contain, or be
// contained by the credential store with a host-backed or shared source. A
// private tmpfs (no source) over the store is permitted: it is per-VM and
// in-memory, which strengthens isolation rather than weakening it.
//
// This is the preventative control behind issue requirement "secrets cannot be
// mounted from host or from a shared volume". Mount-injection subsystems must
// call it before applying operator-supplied mounts. It fails closed: a
// non-empty target that is not an absolute path is rejected with
// ErrInvalidMountTarget rather than waved through, so a caller that forgets to
// validate paths cannot create a bypass.
func GuardMounts(specs []MountSpec) error {
	for i := range specs {
		if err := guardOne(specs[i]); err != nil {
			return fmt.Errorf("mount[%d]: %w", i, err)
		}
	}
	return nil
}

func guardOne(m MountSpec) error {
	target := strings.TrimSpace(m.Target)
	if target == "" {
		// An empty target cannot mount anything; the mount layer rejects it on
		// its own and it definitionally cannot reach the credential store.
		return nil
	}
	if !strings.HasPrefix(target, "/") {
		// Fail closed: a relative target cannot be reasoned about by absolute
		// path comparison, so the guard refuses it rather than waving it through.
		return fmt.Errorf("%w: %q", ErrInvalidMountTarget, target)
	}
	clean := path.Clean(target)
	if !pathOverlapsSecrets(clean) {
		return nil
	}
	mtype := strings.ToLower(strings.TrimSpace(m.Type))
	source := strings.TrimSpace(m.Source)
	if mtype == "tmpfs" && source == "" {
		return nil
	}
	return fmt.Errorf("%w: target %q (type=%q, source=%q) overlaps %s",
		ErrSharedSecretsMount, clean, mtype, source, GuestSecretsDir)
}

// pathOverlapsSecrets reports whether p is the credential store, an ancestor of
// it, or a descendant of it. Any of those would let a mount influence what the
// store resolves to.
func pathOverlapsSecrets(p string) bool {
	secrets := GuestSecretsDir
	// p is the store, p is an ancestor of the store, or p is a descendant.
	return isAncestorOrSame(p, secrets) || isAncestorOrSame(secrets, p)
}

// isAncestorOrSame reports whether ancestor equals or contains descendant,
// treating both as cleaned absolute paths. The root path "/" is an ancestor of
// every absolute path; handling it explicitly avoids the "//" prefix bug that a
// naive ancestor+"/" check would introduce.
func isAncestorOrSame(ancestor, descendant string) bool {
	if ancestor == descendant {
		return true
	}
	if ancestor == "/" {
		return strings.HasPrefix(descendant, "/")
	}
	return strings.HasPrefix(descendant, ancestor+"/")
}

// VMIdentity is the minimal per-VM identity used to verify cross-VM
// distinctness.
type VMIdentity struct {
	VMID     string
	SpiffeID string
}

// VerifyDistinctIdentities verifies that every VM identifier is valid and unique
// and that no two VMs share a SPIFFE ID. Shared identities would let one VM
// present another's credentials and defeat the per-VM boundary.
//
// A SPIFFE identity is optional in nanofuse (see types.CreateVMRequest: owner,
// group, and auto-registration are all optional), so VMs without a SPIFFE ID are
// not failed; they are excluded from the SPIFFE-uniqueness check but still
// validated for identifier uniqueness. The result detail reports how many VMs
// carry a SPIFFE ID so a PASS is not misread as "every VM is SPIFFE-identified".
func VerifyDistinctIdentities(ids []VMIdentity) (VerifyResult, error) {
	res := VerifyResult{Name: "identity-distinctness"}
	seenSpiffe := make(map[string]string, len(ids))
	seenVM := make(map[string]struct{}, len(ids))
	withSpiffe := 0
	for _, id := range ids {
		if err := ValidateVMID(id.VMID); err != nil {
			return res, err
		}
		if _, dup := seenVM[id.VMID]; dup {
			res.Detail = fmt.Sprintf("duplicate VM identifier %q", id.VMID)
			return res, nil
		}
		seenVM[id.VMID] = struct{}{}
		if id.SpiffeID == "" {
			continue
		}
		if other, dup := seenSpiffe[id.SpiffeID]; dup {
			res.Detail = fmt.Sprintf("VMs %q and %q share SPIFFE ID %q", other, id.VMID, id.SpiffeID)
			return res, nil
		}
		seenSpiffe[id.SpiffeID] = id.VMID
		withSpiffe++
	}
	res.Pass = true
	res.Detail = fmt.Sprintf("%d VMs, identifiers unique; %d carry a distinct SPIFFE ID (%d without)",
		len(ids), withSpiffe, len(ids)-withSpiffe)
	return res, nil
}

// VerifyDirPerms verifies that the directory at p has mode exactly 0700 with no
// group or world bits. When requireRoot is true it additionally asserts the
// directory is owned by root:root. That ownership assertion is independent of
// who runs the check — any caller can confirm root:root ownership on platforms
// that expose POSIX stat ownership; it is a no-op only on platforms that do not
// (see statOwner). requireRoot is named for the production invariant it
// enforces (the daemon runs as root, so the store must be root-owned), not for a
// requirement that the caller be root.
func VerifyDirPerms(p string, requireRoot bool) (VerifyResult, error) {
	res := VerifyResult{Name: "store-perms"}
	// Lstat, not Stat: a security verifier must not follow a symlink. A symlinked
	// store path could point at a 0700 directory elsewhere and yield a
	// false-positive PASS while the real store is attacker-controlled.
	info, err := os.Lstat(p)
	if err != nil {
		return res, fmt.Errorf("stat credential store %s: %w", p, err)
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		res.Detail = fmt.Sprintf("%s is a symlink; the credential store must be a real directory", p)
		return res, nil
	}
	if !info.IsDir() {
		res.Detail = fmt.Sprintf("%s is not a directory", p)
		return res, nil
	}
	// The contract is exactly 0700, so reject special mode bits that Perm()
	// does not surface (setuid/setgid/sticky) — a setgid store directory in
	// particular changes the group semantics of files created inside it.
	if special := info.Mode() & (fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky); special != 0 {
		res.Detail = fmt.Sprintf("%s has special mode bits %v set (require plain 0700)", p, special)
		return res, nil
	}
	mode := info.Mode().Perm()
	if mode&0o077 != 0 {
		res.Detail = fmt.Sprintf("%s mode %#o grants group/world access (require 0700)", p, mode)
		return res, nil
	}
	if mode != RequiredDirMode {
		res.Detail = fmt.Sprintf("%s mode %#o != required 0700", p, mode)
		return res, nil
	}
	ownerDetail := ""
	if requireRoot {
		if uid, gid, ok := statOwner(info); ok {
			if uid != RequiredOwnerUID || gid != RequiredOwnerGID {
				res.Detail = fmt.Sprintf("%s owner %d:%d != required 0:0", p, uid, gid)
				return res, nil
			}
			ownerDetail = " owner 0:0"
		}
	}
	res.Pass = true
	res.Detail = fmt.Sprintf("%s mode 0700%s", p, ownerDetail)
	return res, nil
}

// HostCheckOptions configures the host-side verification sweep performed by
// VerifyHost.
type HostCheckOptions struct {
	// SecretsDir is the credential store path to stat. Defaults to
	// GuestSecretsDir. Overridable for in-guest verification and tests.
	SecretsDir string

	// CheckDir enables the on-disk permission check. When false (e.g. the host
	// has no guest store to inspect), the permission check is skipped and
	// reported, not failed.
	CheckDir bool

	// RequireRoot enforces root:root ownership of the store (set when running as
	// root).
	RequireRoot bool

	// Mounts are proposed mounts to validate against the guard.
	Mounts []MountSpec

	// Identities are per-VM identities to check for cross-VM distinctness.
	Identities []VMIdentity
}

// VerifyHost runs the host-side isolation checks and returns an aggregate
// report. It always includes a policy self-check that asserts the mount guard
// denies a host-backed mount over the store and admits a private tmpfs, so the
// report is never empty and the guard's wiring is proven on every run.
func VerifyHost(opts HostCheckOptions) Report {
	var r Report
	r.Results = append(r.Results, policySelfCheck())

	if opts.CheckDir {
		dir := opts.SecretsDir
		if dir == "" {
			dir = GuestSecretsDir
		}
		res, err := VerifyDirPerms(dir, opts.RequireRoot)
		if err != nil {
			res = VerifyResult{Name: "store-perms", Pass: false, Detail: err.Error()}
		}
		r.Results = append(r.Results, res)
		r.Subjects++
	}

	if len(opts.Mounts) > 0 {
		res := VerifyResult{Name: "mount-guard", Pass: true,
			Detail: fmt.Sprintf("%d proposed mounts respect the credential store", len(opts.Mounts))}
		if err := GuardMounts(opts.Mounts); err != nil {
			res.Pass = false
			res.Detail = err.Error()
		}
		r.Results = append(r.Results, res)
		r.Subjects++
	}

	if len(opts.Identities) > 0 {
		res, err := VerifyDistinctIdentities(opts.Identities)
		if err != nil {
			res = VerifyResult{Name: "identity-distinctness", Pass: false, Detail: err.Error()}
		}
		r.Results = append(r.Results, res)
		r.Subjects++
	}

	return r
}

// policySelfCheck asserts the guard's core invariant: a host-backed bind over
// the credential store is rejected while a private tmpfs over it is admitted.
// This is defense in depth — it fails the report loudly if the guard logic is
// ever weakened.
func policySelfCheck() VerifyResult {
	res := VerifyResult{Name: "policy-contract"}
	hostBind := MountSpec{Target: GuestSecretsDir, Source: "/host/leak", Type: "bind"}
	privateTmpfs := MountSpec{Target: GuestSecretsDir, Type: "tmpfs"}
	if err := GuardMounts([]MountSpec{hostBind}); err == nil {
		res.Detail = "guard admitted a host-backed mount over the credential store"
		return res
	}
	if err := GuardMounts([]MountSpec{privateTmpfs}); err != nil {
		res.Detail = fmt.Sprintf("guard rejected a private tmpfs store: %v", err)
		return res
	}
	res.Pass = true
	res.Detail = "guard denies host-backed store mounts and admits private tmpfs"
	return res
}
