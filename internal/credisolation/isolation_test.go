package credisolation

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateVMID(t *testing.T) {
	t.Parallel()
	good := []string{"vm1", "vm-2", "vm_3", "abc123", strings.Repeat("a", maxVMIDLen)}
	for _, id := range good {
		if err := ValidateVMID(id); err != nil {
			t.Errorf("ValidateVMID(%q) = %v, want nil", id, err)
		}
	}
	bad := []string{
		"",                                // empty
		"../vm2",                          // traversal
		"vm/2",                            // separator
		"vm 2",                            // space
		"vm;rm -rf",                       // shell metacharacters
		"vm$2",                            // expansion
		strings.Repeat("a", maxVMIDLen+1), // too long
	}
	for _, id := range bad {
		if err := ValidateVMID(id); !errors.Is(err, ErrInvalidVMID) {
			t.Errorf("ValidateVMID(%q) = %v, want ErrInvalidVMID", id, err)
		}
	}
}

func TestAuthorize(t *testing.T) {
	t.Parallel()
	if err := Authorize("vm1", "vm1"); err != nil {
		t.Errorf("Authorize(vm1, vm1) = %v, want nil (a VM may access its own store)", err)
	}
	if err := Authorize("vm1", "vm2"); !errors.Is(err, ErrCrossVMAccess) {
		t.Errorf("Authorize(vm1, vm2) = %v, want ErrCrossVMAccess", err)
	}
	// Invalid identifiers must not be silently authorized, and the underlying
	// ErrInvalidVMID must stay inspectable alongside ErrCrossVMAccess.
	err := Authorize("../vm2", "../vm2")
	if !errors.Is(err, ErrCrossVMAccess) {
		t.Errorf("Authorize with invalid id = %v, want ErrCrossVMAccess", err)
	}
	if !errors.Is(err, ErrInvalidVMID) {
		t.Errorf("Authorize with invalid id = %v, want errors.Is(err, ErrInvalidVMID) (chain preserved)", err)
	}
}

func TestGuardMounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		spec    MountSpec
		wantErr bool
	}{
		{"host bind over store", MountSpec{Target: GuestSecretsDir, Source: "/host/leak", Type: "bind"}, true},
		{"volume over store", MountSpec{Target: GuestSecretsDir, Source: "shared-vol", Type: "volume"}, true},
		{"descendant bind", MountSpec{Target: GuestSecretsDir + "/svid.json", Source: "/host/svid", Type: "bind"}, true},
		{"ancestor bind shadows store", MountSpec{Target: "/var/run", Source: "/host/run", Type: "bind"}, true},
		{"ancestor bind (/var)", MountSpec{Target: "/var", Source: "/host/var", Type: "bind"}, true},
		{"root bind shadows store", MountSpec{Target: "/", Source: "/host/root", Type: "bind"}, true},
		{"root volume shadows store", MountSpec{Target: "/", Source: "rootvol", Type: "volume"}, true},
		{"root tmpfs allowed (no source)", MountSpec{Target: "/", Type: "tmpfs"}, false},
		{"non-normalized descendant", MountSpec{Target: "/var/run/../run/secrets/daax", Source: "/h", Type: "bind"}, true},
		{"private tmpfs over store", MountSpec{Target: GuestSecretsDir, Type: "tmpfs"}, false},
		{"unrelated bind", MountSpec{Target: "/data", Source: "/host/data", Type: "bind"}, false},
		{"sibling path", MountSpec{Target: "/var/run/secrets/other", Source: "/h", Type: "bind"}, false},
		{"prefix-but-not-ancestor", MountSpec{Target: "/var/run/secrets/daaxxy", Source: "/h", Type: "bind"}, false},
		{"empty target", MountSpec{}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := GuardMounts([]MountSpec{tc.spec})
			if tc.wantErr {
				if !errors.Is(err, ErrSharedSecretsMount) {
					t.Errorf("GuardMounts(%+v) = %v, want ErrSharedSecretsMount", tc.spec, err)
				}
			} else if err != nil {
				t.Errorf("GuardMounts(%+v) = %v, want nil", tc.spec, err)
			}
		})
	}
}

func TestGuardMountsFailsClosedOnRelativeTarget(t *testing.T) {
	t.Parallel()
	for _, target := range []string{"relative/path", "secrets/daax", "../escape"} {
		err := GuardMounts([]MountSpec{{Target: target, Source: "/h", Type: "bind"}})
		if !errors.Is(err, ErrInvalidMountTarget) {
			t.Errorf("GuardMounts(target=%q) = %v, want ErrInvalidMountTarget (fail closed)", target, err)
		}
	}
}

func TestVerifyDistinctIdentities(t *testing.T) {
	t.Parallel()
	t.Run("distinct", func(t *testing.T) {
		res, err := VerifyDistinctIdentities([]VMIdentity{
			{VMID: "vm1", SpiffeID: "spiffe://poley.dev/w/microvm/vm1"},
			{VMID: "vm2", SpiffeID: "spiffe://poley.dev/w/microvm/vm2"},
			{VMID: "vm3"}, // no SPIFFE id; still valid
		})
		if err != nil || !res.Pass {
			t.Fatalf("distinct identities: pass=%v err=%v detail=%q", res.Pass, err, res.Detail)
		}
	})
	t.Run("shared spiffe id", func(t *testing.T) {
		res, err := VerifyDistinctIdentities([]VMIdentity{
			{VMID: "vm1", SpiffeID: "spiffe://poley.dev/shared"},
			{VMID: "vm2", SpiffeID: "spiffe://poley.dev/shared"},
		})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if res.Pass {
			t.Error("shared SPIFFE ID must fail distinctness")
		}
	})
	t.Run("duplicate vm id", func(t *testing.T) {
		res, _ := VerifyDistinctIdentities([]VMIdentity{{VMID: "vm1"}, {VMID: "vm1"}})
		if res.Pass {
			t.Error("duplicate VM id must fail")
		}
	})
	t.Run("invalid vm id", func(t *testing.T) {
		_, err := VerifyDistinctIdentities([]VMIdentity{{VMID: "../evil"}})
		if !errors.Is(err, ErrInvalidVMID) {
			t.Errorf("invalid id err = %v, want ErrInvalidVMID", err)
		}
	})
}

func TestVerifyDirPerms(t *testing.T) {
	t.Parallel()
	t.Run("0700 passes", func(t *testing.T) {
		dir := mkStore(t, 0o700)
		res, err := VerifyDirPerms(dir, false)
		if err != nil || !res.Pass {
			t.Fatalf("0700 dir: pass=%v err=%v detail=%q", res.Pass, err, res.Detail)
		}
	})
	t.Run("group/world bits fail", func(t *testing.T) {
		for _, mode := range []os.FileMode{0o750, 0o705, 0o755, 0o777, 0o701} {
			dir := mkStore(t, mode)
			res, err := VerifyDirPerms(dir, false)
			if err != nil {
				t.Fatalf("mode %o: unexpected err %v", mode, err)
			}
			if res.Pass {
				t.Errorf("mode %o must fail (grants group/world access)", mode)
			}
		}
	})
	t.Run("missing dir errors", func(t *testing.T) {
		_, err := VerifyDirPerms(filepath.Join(t.TempDir(), "absent"), false)
		if err == nil {
			t.Error("missing dir must return an error")
		}
	})
	t.Run("file not dir fails", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		res, err := VerifyDirPerms(f, false)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if res.Pass {
			t.Error("a regular file must not pass the store-perms check")
		}
	})
	t.Run("symlink is rejected (no following)", func(t *testing.T) {
		real := mkStore(t, 0o700)
		link := filepath.Join(t.TempDir(), "secrets-link")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		res, err := VerifyDirPerms(link, false)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if res.Pass {
			t.Error("a symlinked store path must fail (verifier must not follow symlinks)")
		}
		if !strings.Contains(res.Detail, "symlink") {
			t.Errorf("detail = %q, want it to mention symlink", res.Detail)
		}
	})
	t.Run("require-root rejects non-root owner", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("running as root; cannot exercise the non-root-owner rejection path")
		}
		dir := mkStore(t, 0o700)
		res, err := VerifyDirPerms(dir, true)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if res.Pass {
			t.Error("require-root must fail when the store is not owned by root:root")
		}
	})
}

// TestCrossVMLeakagePrevented is the core invariant test. It builds two
// independent per-VM credential stores, plants a secret in VM2's store, and
// proves the host-side controls prevent VM1 from reaching it:
//
//  1. Authorization denies any cross-VM access predicate.
//  2. The mount guard rejects every configuration that would bridge VM1's view
//     to VM2's secret (host bind, shared volume, or an ancestor mount).
//  3. Each store excludes group/world access (mode 0700), so even co-located
//     stores are not readable across ownership boundaries.
//  4. The two stores are physically distinct paths.
//
// The kernel-LSM denial of an in-guest cross-read and live-VM termination are
// runtime-enforced and out of scope for this unit harness; see package doc.
func TestCrossVMLeakagePrevented(t *testing.T) {
	t.Parallel()

	vm1Store := mkStore(t, 0o700)
	vm2Store := mkStore(t, 0o700)

	vm2SVID := filepath.Join(vm2Store, "svid.json")
	if err := os.WriteFile(vm2SVID, []byte(`{"svid":"vm2-secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// 1. Authorization: VM1 may never access VM2's store; VM2 may access its own.
	if err := Authorize("vm1", "vm2"); !errors.Is(err, ErrCrossVMAccess) {
		t.Fatalf("VM1 access to VM2 store: %v, want ErrCrossVMAccess", err)
	}
	if err := Authorize("vm2", "vm2"); err != nil {
		t.Fatalf("VM2 access to its own store denied: %v", err)
	}

	// 2. The guard rejects every attempt to bridge VM1's secrets view to VM2's
	//    backing storage.
	bridges := []MountSpec{
		{Target: GuestSecretsDir, Source: vm2Store, Type: "bind"},
		{Target: GuestSecretsDir, Source: vm2SVID, Type: "bind"},
		{Target: GuestSecretsDir + "/svid.json", Source: vm2SVID, Type: "bind"},
		{Target: "/var/run", Source: vm2Store, Type: "bind"},
		{Target: "/", Source: vm2Store, Type: "bind"}, // root mount is an ancestor of the store
		{Target: GuestSecretsDir, Source: "cross-vm-vol", Type: "volume"},
	}
	for _, b := range bridges {
		if err := GuardMounts([]MountSpec{b}); !errors.Is(err, ErrSharedSecretsMount) {
			t.Errorf("guard admitted a cross-VM bridge %+v: err=%v", b, err)
		}
	}

	// 3. Neither store grants group/world access.
	for _, store := range []string{vm1Store, vm2Store} {
		res, err := VerifyDirPerms(store, false)
		if err != nil || !res.Pass {
			t.Errorf("store %s perms: pass=%v err=%v detail=%q", store, res.Pass, err, res.Detail)
		}
		info, err := os.Stat(store)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o077 != 0 {
			t.Errorf("store %s mode %#o leaks access to group/world", store, info.Mode().Perm())
		}
	}

	// 4. The stores are physically distinct.
	if vm1Store == vm2Store {
		t.Fatal("two VMs resolved to the same credential store path")
	}
}

func TestMonitorHandleCrossVMAttempt(t *testing.T) {
	t.Parallel()
	t.Run("audits and terminates offending vm", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
		var killed string
		mon := NewMonitor(logger, func(vmID string) error { killed = vmID; return nil })

		err := mon.HandleCrossVMAttempt(AccessAttempt{
			RequestingVMID: "vm1",
			TargetVMID:     "vm2",
			Path:           GuestSecretsDir + "/svid.json",
			When:           time.Unix(1700000000, 0).UTC(),
		})
		if !errors.Is(err, ErrCrossVMAccess) {
			t.Fatalf("HandleCrossVMAttempt err = %v, want ErrCrossVMAccess", err)
		}
		if killed != "vm1" {
			t.Errorf("terminated %q, want offending VM vm1", killed)
		}
		logs := buf.String()
		for _, want := range []string{
			"cred_isolation.cross_vm_access_attempt",
			"cred_isolation.failsafe_response",
			"requesting_vm_valid", "target_vm_valid",
			"vm1", "vm2", "svid.json",
		} {
			if !strings.Contains(logs, want) {
				t.Errorf("audit log missing %q; got: %s", want, logs)
			}
		}
	})
	t.Run("auto-populated timestamp is UTC", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, nil))
		mon := NewMonitor(logger, func(string) error { return nil })
		// Omit When so the monitor auto-populates it; it must be UTC, which slog
		// renders with a trailing "Z" rather than a numeric offset.
		_ = mon.HandleCrossVMAttempt(AccessAttempt{RequestingVMID: "vm1", TargetVMID: "vm2"})
		var rec struct {
			Timestamp string `json:"timestamp"`
		}
		dec := json.NewDecoder(bytes.NewReader(buf.Bytes()))
		if err := dec.Decode(&rec); err != nil {
			t.Fatalf("decode audit log: %v", err)
		}
		if rec.Timestamp == "" || !strings.HasSuffix(rec.Timestamp, "Z") {
			t.Errorf("auto timestamp = %q, want a UTC value ending in Z", rec.Timestamp)
		}
	})
	t.Run("propagates terminator failure (inspectable chain)", func(t *testing.T) {
		errKill := errors.New("kill failed")
		mon := NewMonitor(nil, func(string) error { return errKill })
		err := mon.HandleCrossVMAttempt(AccessAttempt{RequestingVMID: "vm1", TargetVMID: "vm2"})
		if !errors.Is(err, ErrCrossVMAccess) {
			t.Errorf("err = %v, want it to wrap ErrCrossVMAccess", err)
		}
		// The underlying terminator error must remain inspectable via errors.Is,
		// not merely be stringified.
		if !errors.Is(err, errKill) {
			t.Errorf("err = %v, want errors.Is(err, errKill) true (chain preserved)", err)
		}
	})
	t.Run("nil terminator still audits and reports", func(t *testing.T) {
		mon := NewMonitor(nil, nil)
		err := mon.HandleCrossVMAttempt(AccessAttempt{RequestingVMID: "vm1", TargetVMID: "vm2"})
		if !errors.Is(err, ErrCrossVMAccess) {
			t.Errorf("err = %v, want ErrCrossVMAccess", err)
		}
	})
}

func TestReportAndStatusLine(t *testing.T) {
	t.Parallel()
	if (Report{}).Pass() {
		t.Error("an empty report must not pass (it proved nothing)")
	}
	if got := (Report{}).StatusLine(); got != statusFailLine {
		t.Errorf("empty report StatusLine = %q, want %q", got, statusFailLine)
	}
	// All checks pass but no real subject was inspected: must NOT claim PASS.
	selfOnly := Report{Results: []VerifyResult{{Name: "policy-contract", Pass: true}}, Subjects: 0}
	if selfOnly.Pass() {
		t.Error("a report with zero subjects must not pass")
	}
	if got := selfOnly.StatusLine(); got != "credential isolation: NOT VERIFIED" {
		t.Errorf("self-only StatusLine = %q, want NOT VERIFIED", got)
	}
	// At least one subject checked and all pass: the literal PASS line.
	passing := Report{Results: []VerifyResult{{Name: "store-perms", Pass: true}}, Subjects: 1}
	if !passing.Pass() {
		t.Error("a report with a passing subject must pass")
	}
	if got := passing.StatusLine(); got != "credential isolation: PASS" {
		t.Errorf("passing StatusLine = %q, want the literal PASS line", got)
	}
	mixed := Report{Results: []VerifyResult{{Pass: true}, {Pass: false}}, Subjects: 1}
	if mixed.Pass() {
		t.Error("a report with any failing check must not pass")
	}
	if got := mixed.StatusLine(); got != statusFailLine {
		t.Errorf("mixed StatusLine = %q, want FAIL", got)
	}
}

func TestVerifyHost(t *testing.T) {
	t.Parallel()
	t.Run("policy self-check present; no subject means NOT VERIFIED", func(t *testing.T) {
		r := VerifyHost(HostCheckOptions{})
		if len(r.Results) == 0 {
			t.Fatal("VerifyHost produced no results")
		}
		if r.Results[0].Name != "policy-contract" || !r.Results[0].Pass {
			t.Errorf("first result = %+v, want passing policy-contract", r.Results[0])
		}
		if r.Pass() {
			t.Error("a sweep with no real subject must not report PASS")
		}
		if got := r.StatusLine(); got != "credential isolation: NOT VERIFIED" {
			t.Errorf("StatusLine = %q, want NOT VERIFIED", got)
		}
	})
	t.Run("good store + clean mounts + distinct ids pass", func(t *testing.T) {
		dir := mkStore(t, 0o700)
		r := VerifyHost(HostCheckOptions{
			SecretsDir: dir,
			CheckDir:   true,
			Mounts:     []MountSpec{{Target: "/data", Source: "/host/data", Type: "bind"}},
			Identities: []VMIdentity{{VMID: "vm1", SpiffeID: "id1"}, {VMID: "vm2", SpiffeID: "id2"}},
		})
		if !r.Pass() || r.StatusLine() != "credential isolation: PASS" {
			t.Errorf("expected PASS, got %q with results %+v", r.StatusLine(), r.Results)
		}
	})
	t.Run("bad mount fails the sweep", func(t *testing.T) {
		r := VerifyHost(HostCheckOptions{
			Mounts: []MountSpec{{Target: GuestSecretsDir, Source: "/host/leak", Type: "bind"}},
		})
		if r.Pass() {
			t.Error("a host-backed store mount must fail the sweep")
		}
	})
	t.Run("world-readable store fails the sweep", func(t *testing.T) {
		dir := mkStore(t, 0o755)
		r := VerifyHost(HostCheckOptions{SecretsDir: dir, CheckDir: true})
		if r.Pass() {
			t.Error("a 0755 store must fail the sweep")
		}
	})
}

// mkStore creates a temp directory and forces its mode (defeating umask) so
// permission assertions are deterministic.
func mkStore(t *testing.T, mode os.FileMode) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "secrets")
	if err := os.Mkdir(dir, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, mode); err != nil {
		t.Fatal(err)
	}
	return dir
}
