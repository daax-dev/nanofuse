package firecracker

import (
	"context"
	"errors"
	"testing"

	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

func TestExecUnsupportedWhenKeyUnset(t *testing.T) {
	m := NewManager("/usr/bin/firecracker", t.TempDir())
	vm := &types.VM{ID: "x", Runtime: &types.VMRuntime{NetworkInfo: &types.NetworkRuntimeInfo{GuestIP: "172.16.0.10"}}}
	_, err := m.Exec(context.Background(), vm, []string{"echo", "hi"})
	if !errors.Is(err, vmm.ErrUnsupportedOperation) {
		t.Fatalf("want ErrUnsupportedOperation, got %v", err)
	}
}

func TestExecRequiresGuestIP(t *testing.T) {
	m := NewManager("/usr/bin/firecracker", t.TempDir())
	m.SetExecSSH("/tmp/key", "root")
	vm := &types.VM{ID: "x"} // no runtime/IP
	_, err := m.Exec(context.Background(), vm, []string{"echo", "hi"})
	if err == nil || errors.Is(err, vmm.ErrUnsupportedOperation) {
		t.Fatalf("want network-not-ready error, got %v", err)
	}
}

func TestGuestIP(t *testing.T) {
	t.Run("runtime preferred", func(t *testing.T) {
		vm := &types.VM{
			Runtime: &types.VMRuntime{NetworkInfo: &types.NetworkRuntimeInfo{GuestIP: "10.0.0.5"}},
			Config:  types.VMConfig{Network: types.NetworkConfig{IPAddress: "10.0.0.9"}},
		}
		if got := guestIP(vm); got != "10.0.0.5" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("config fallback", func(t *testing.T) {
		vm := &types.VM{Config: types.VMConfig{Network: types.NetworkConfig{IPAddress: "10.0.0.9"}}}
		if got := guestIP(vm); got != "10.0.0.9" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("none", func(t *testing.T) {
		if got := guestIP(&types.VM{}); got != "" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestFirecrackerRuntimeID(t *testing.T) {
	if got := firecrackerRuntimeID(&types.VM{ID: "vm-1", Runtime: &types.VMRuntime{ExternalID: "ext-9"}}); got != "ext-9" {
		t.Fatalf("want external id, got %q", got)
	}
	if got := firecrackerRuntimeID(&types.VM{ID: "vm-1"}); got != "vm-1" {
		t.Fatalf("want vm id fallback, got %q", got)
	}
	if got := firecrackerRuntimeID(nil); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestShellQuoteJoin(t *testing.T) {
	// shellQuote always single-quotes for safety, even simple words.
	cases := map[string]string{
		"plain":     "'plain'",
		"a b":       `'a b'`,
		"":          "''",
		"it's":      `'it'\''s'`,
		"$(rm -rf)": `'$(rm -rf)'`,
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Fatalf("shellQuote(%q)=%q want %q", in, got, want)
		}
	}
	if got := shellJoin([]string{"sh", "-lc", "echo hi"}); got != `'sh' '-lc' 'echo hi'` {
		t.Fatalf("shellJoin=%q", got)
	}
}
