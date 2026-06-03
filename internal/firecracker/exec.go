package firecracker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

// exitSentinel prefixes the trailing line the remote shell prints so the daemon
// can recover the guest command's true exit code independent of ssh's status.
const exitSentinel = "__NANOFUSE_EXIT__="

// parseExitSentinel extracts the trailing exit-code sentinel from captured
// stdout. ok is true only when the sentinel is present, meaning the remote shell
// actually executed the command. The sentinel line is stripped from cleanStdout.
func parseExitSentinel(s string) (code int, cleanStdout string, ok bool) {
	idx := strings.LastIndex(s, exitSentinel)
	if idx < 0 {
		return 0, s, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(s[idx+len(exitSentinel):]))
	if err != nil {
		return 0, s, false
	}
	// Drop the sentinel and the single newline we injected before it.
	clean := strings.TrimSuffix(s[:idx], "\n")
	return n, clean, true
}

// Exec runs a command inside a running Firecracker guest over SSH and returns
// its stdout, stderr, and exit code. This gives the Firecracker backend the
// same `vm exec` capability as the apple_container backend.
//
// It requires a daemon-side private key (SetExecSSH) whose matching public key
// is present in the guest image's authorized_keys, and a reachable guest IP.
// When exec is not configured, it returns vmm.ErrUnsupportedOperation so the
// API reports the runtime as not supporting exec.
func (m *Manager) Exec(ctx context.Context, vm *types.VM, command []string) (*types.VMExecResult, error) {
	if m.execSSHKey == "" {
		return nil, fmt.Errorf("firecracker exec requires firecracker.exec_ssh_key_path: %w", vmm.ErrUnsupportedOperation)
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("command is required")
	}

	ip := guestIP(vm)
	if ip == "" {
		return nil, fmt.Errorf("firecracker exec requires a guest IP address; VM is not network-ready")
	}

	user := m.execSSHUser
	if user == "" {
		user = "root"
	}

	// remoteCommand is a single, fully shell-quoted string. ssh sends everything
	// after the destination to the guest's login shell verbatim, so no "--"
	// terminator is used (ssh would forward it into the remote command). A
	// trailing sentinel reports the command's true exit code: ssh collapses both
	// its own transport failures and a remote "exit 255" onto status 255, so the
	// sentinel is the only reliable way to recover the real guest exit code.
	remoteCommand := shellJoin(command) + "; printf '\\n" + exitSentinel + "%d\\n' \"$?\""
	hostKeyOpts := m.hostKeyOptions(vm.ID)
	args := make([]string, 0, 10+len(hostKeyOpts)+2)
	args = append(args,
		"-i", m.execSSHKey,
		"-o", "IdentitiesOnly=yes",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		"-o", "LogLevel=ERROR",
	)
	args = append(args, hostKeyOpts...)
	args = append(args, fmt.Sprintf("%s@%s", user, ip), remoteCommand)

	// Cap captured output: the guest is untrusted and could emit unbounded data.
	stdout := &cappedBuffer{limit: execOutputCap}
	stderr := &cappedBuffer{limit: execOutputCap}
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	runErr := cmd.Run()

	// Recover the true guest exit code from the sentinel. Its presence means the
	// remote shell actually ran, so ssh's own 255 (transport) is distinguishable
	// from a guest "exit 255".
	guestCode, cleanStdout, ranInGuest := parseExitSentinel(stdout.String())
	result := &types.VMExecResult{
		Command:   append([]string(nil), command...),
		Stdout:    cleanStdout,
		Stderr:    stderr.String(),
		RuntimeID: firecrackerRuntimeID(vm),
	}

	if ranInGuest {
		result.ExitCode = guestCode
		return result, nil
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return result, ctxErr
	}

	// No sentinel. It may have been truncated away by the output cap (>1 MiB of
	// stdout) even though the command ran, so fall back to ssh's own exit code:
	// a non-255 ExitError is the guest's code; only 255 (or a failure to start
	// ssh) is the ambiguous transport case the sentinel exists to disambiguate.
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		if code := exitErr.ExitCode(); code != 255 {
			result.ExitCode = code
			return result, nil
		}
	}

	// ssh transport failure (connection/auth) or the ssh client could not start —
	// a host-side problem, not a guest command result.
	result.ExitCode = 255
	msg := strings.TrimSpace(stderr.String())
	if msg == "" {
		if runErr != nil {
			msg = runErr.Error()
		} else {
			msg = "ssh connection failed (no stderr); check guest sshd, network reachability, and the exec key"
		}
	}
	// Keep the returned error small (full stderr remains in result.Stderr) so it
	// does not flood logs or oversize API error details.
	const maxErrLen = 256
	if len(msg) > maxErrLen {
		msg = msg[:maxErrLen] + "… (truncated; see stderr)"
	}
	return result, fmt.Errorf("firecracker exec ssh transport error: %s", msg)
}

// hostKeyOptions returns the ssh host-key verification options. When enabled,
// it uses accept-new TOFU with a known_hosts file under the data dir, keyed by a
// stable per-VM HostKeyAlias (with CheckHostIP disabled) so a recycled guest IP
// does not trip "host key changed". The default disables host-key checks because
// guest host keys are ephemeral and the exec bridge is daemon-controlled.
func (m *Manager) hostKeyOptions(vmID string) []string {
	if m.execVerifyHostK {
		knownHosts := filepath.Join(m.dataDir, "exec_known_hosts")
		return []string{
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "UserKnownHostsFile=" + knownHosts,
			"-o", "CheckHostIP=no",
			"-o", "HostKeyAlias=nf-" + vmID,
		}
	}
	return []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}
}

// execOutputCap bounds captured stdout/stderr per stream so a hostile guest
// cannot exhaust daemon memory by emitting unbounded output.
const execOutputCap = 1 << 20 // 1 MiB

// cappedBuffer accumulates up to limit bytes, then silently discards the rest
// and flags truncation. Write always reports the full length so the ssh process
// is never blocked or errored by a short write.
type cappedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	if remaining := c.limit - c.buf.Len(); remaining > 0 {
		if len(p) > remaining {
			c.buf.Write(p[:remaining])
			c.truncated = true
		} else {
			c.buf.Write(p)
		}
	} else if len(p) > 0 {
		c.truncated = true
	}
	return len(p), nil
}

const truncationMarker = "\n[output truncated]"

func (c *cappedBuffer) String() string {
	if !c.truncated {
		return c.buf.String()
	}
	// Keep the final string within limit, reserving room for the marker.
	keep := c.limit - len(truncationMarker)
	if keep < 0 {
		keep = 0
	}
	s := c.buf.String()
	if len(s) > keep {
		s = s[:keep]
	}
	out := s + truncationMarker
	// Final guard so the result never exceeds limit, even when limit is smaller
	// than the marker itself.
	if len(out) > c.limit {
		out = out[:c.limit]
	}
	return out
}

// firecrackerRuntimeID returns the runtime-owned identifier for a VM, matching
// the VMExecResult.runtime_id contract (a backend handle, not a network address).
func firecrackerRuntimeID(vm *types.VM) string {
	if vm == nil {
		return ""
	}
	if vm.Runtime != nil && strings.TrimSpace(vm.Runtime.ExternalID) != "" {
		return vm.Runtime.ExternalID
	}
	return vm.ID
}

// guestIP resolves the guest IP from runtime info, falling back to configured IP.
func guestIP(vm *types.VM) string {
	if vm == nil {
		return ""
	}
	if vm.Runtime != nil && vm.Runtime.NetworkInfo != nil && vm.Runtime.NetworkInfo.GuestIP != "" {
		return vm.Runtime.NetworkInfo.GuestIP
	}
	return vm.Config.Network.IPAddress
}

// shellJoin builds a single POSIX-shell command string from an argv slice,
// single-quoting each argument so the remote shell preserves word boundaries.
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = shellQuote(a)
	}
	return strings.Join(quoted, " ")
}

// shellQuote single-quotes a string for safe use in a POSIX shell.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	// Replace each single quote with the '\'' sequence.
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
