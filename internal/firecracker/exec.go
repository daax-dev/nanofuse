package firecracker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

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
	// terminator is used (ssh would forward it into the remote command).
	remoteCommand := shellJoin(command)
	args := []string{
		"-i", m.execSSHKey,
		"-o", "IdentitiesOnly=yes",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		"-o", "LogLevel=ERROR",
	}
	args = append(args, m.hostKeyOptions()...)
	args = append(args, fmt.Sprintf("%s@%s", user, ip), remoteCommand)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	result := &types.VMExecResult{
		Command:   append([]string(nil), command...),
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		RuntimeID: firecrackerRuntimeID(vm),
	}

	if runErr == nil {
		result.ExitCode = 0
		return result, nil
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return result, ctxErr
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		code := exitErr.ExitCode()
		// ssh uses 255 for its own connection/transport failures. Surface those
		// as an error rather than a guest command exit code, but still return the
		// populated result so callers can read captured stdout/stderr diagnostics.
		result.ExitCode = code
		if code == 255 {
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				msg = "ssh connection failed (no stderr); check guest sshd, network reachability, and the exec key"
			}
			return result, fmt.Errorf("firecracker exec ssh transport error: %s", msg)
		}
		return result, nil
	}

	// ssh binary missing or could not start is a host misconfiguration, not a
	// backend capability gap, so return a regular error (with the populated
	// result). ErrUnsupportedOperation stays reserved for true gaps such as a
	// missing exec key.
	return result, fmt.Errorf("firecracker exec could not run ssh client: %w", runErr)
}

// hostKeyOptions returns the ssh host-key verification options. When enabled,
// it uses accept-new TOFU with a known_hosts file under the data dir. The
// default disables host-key checks because guest host keys are ephemeral and
// the bridge recycles guest IPs across VMs (a persistent known_hosts would then
// fail with "host key changed"); the exec bridge is daemon-controlled.
func (m *Manager) hostKeyOptions() []string {
	if m.execVerifyHostK {
		knownHosts := filepath.Join(m.dataDir, "exec_known_hosts")
		return []string{"-o", "StrictHostKeyChecking=accept-new", "-o", "UserKnownHostsFile=" + knownHosts}
	}
	return []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}
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
