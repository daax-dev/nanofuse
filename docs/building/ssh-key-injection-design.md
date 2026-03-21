# SSH Key Injection Design

## Overview

Enable SSH access to VMs by injecting user's public key at VM creation time via kernel command line parameters.

## Problem Statement

Currently there's no way to:
1. SSH into a running VM for debugging
2. Inject user-specific SSH keys without rebuilding images
3. Debug service failures inside VMs (no exec command)

## Solution

Pass SSH public key via kernel cmdline, have in-VM service install it on boot.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        VM Creation Flow                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. User runs:                                                   │
│     nanofuse vm create image:tag myvm --ssh-key ~/.ssh/id_rsa.pub│
│                                                                  │
│  2. CLI reads public key, base64 encodes it                      │
│                                                                  │
│  3. Key appended to kernel args:                                 │
│     "console=ttyS0 ... sshkey=c3NoLXJzYSBBQUFBQjN..."           │
│                                                                  │
│  4. VM boots with kernel args                                    │
│                                                                  │
│  5. install-ssh-key.service runs (oneshot, early boot):          │
│     - Reads /proc/cmdline                                        │
│     - Extracts sshkey= parameter                                 │
│     - Base64 decodes                                             │
│     - Writes to /root/.ssh/authorized_keys                       │
│                                                                  │
│  6. SSH service starts, user can connect                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. CLI Changes (`cmd/nanofuse/vm_create.go`)

Add `--ssh-key` flag:

```go
createCmd.Flags().StringVar(&sshKeyPath, "ssh-key", "", "Path to SSH public key for VM access")
```

Read and encode key:

```go
func readAndEncodeSSHKey(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", fmt.Errorf("failed to read SSH key: %w", err)
    }
    // Trim whitespace, validate it looks like a public key
    key := strings.TrimSpace(string(data))
    if !strings.HasPrefix(key, "ssh-") {
        return "", fmt.Errorf("invalid SSH public key format")
    }
    return base64.StdEncoding.EncodeToString([]byte(key)), nil
}
```

### 2. API Changes (`internal/types/vm.go`)

Add SSH key to VM config:

```go
type VMConfig struct {
    // ... existing fields
    SSHPublicKey string `json:"ssh_public_key,omitempty"` // Base64 encoded
}
```

### 3. Kernel Args Assembly (`internal/firecracker/vm.go`)

Append sshkey to kernel args when present:

```go
func buildKernelArgs(vm *types.VM) string {
    args := vm.Config.KernelArgs
    if vm.Config.SSHPublicKey != "" {
        args += " sshkey=" + vm.Config.SSHPublicKey
    }
    return args
}
```

### 4. In-VM Service (`images/base/units/install-ssh-key.service`)

Systemd service:

```ini
[Unit]
Description=Install SSH key from kernel cmdline
Before=ssh.service
After=local-fs.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/install-ssh-key.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
```

### 5. In-VM Script (`images/base/scripts/install-ssh-key.sh`)

```bash
#!/bin/bash
set -e

# Extract sshkey parameter from kernel cmdline
SSHKEY_B64=$(cat /proc/cmdline | grep -oP 'sshkey=\K[^ ]+' || true)

if [ -z "$SSHKEY_B64" ]; then
    echo "No SSH key provided via kernel cmdline"
    exit 0
fi

# Decode and install
SSH_KEY=$(echo "$SSHKEY_B64" | base64 -d)

mkdir -p /root/.ssh
chmod 700 /root/.ssh

# Append to authorized_keys (don't overwrite existing)
echo "$SSH_KEY" >> /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys

echo "SSH key installed successfully"
```

## Size Considerations

- Typical SSH public key: 400-800 bytes
- Base64 overhead: ~33%
- Encoded key: ~530-1060 bytes
- Kernel cmdline limit: ~4KB (kernel default) to 2MB (configurable)
- Plenty of room for one key

## Security Considerations

1. **Key visible in /proc/cmdline** - Anyone with root in VM can see it. Acceptable for debug use case.
2. **Key in VM metadata** - Stored in nanofuse database. Same security as other VM config.
3. **No private key exposure** - Only public key is passed.

## Acceptance Criteria

### AC1: CLI Accepts SSH Key Flag
```bash
# Flag exists and reads key file
nanofuse vm create myimage:latest test-ssh --ssh-key ~/.ssh/id_rsa.pub
echo $?
# Expected: 0

# Invalid key path fails gracefully
nanofuse vm create myimage:latest test-ssh --ssh-key /nonexistent
echo $?
# Expected: non-zero, error message
```

### AC2: Key Appears in Kernel Args
```bash
# After VM created, check stored config
sudo nanofuse vm inspect test-ssh --json | jq -r '.config.kernel_args' | grep -q 'sshkey='
# Expected: exit code 0
```

### AC3: Key Installed in VM
```bash
# SSH key is in authorized_keys after boot
ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 root@${VM_IP} \
    'cat /root/.ssh/authorized_keys' | grep -q 'ssh-'
# Expected: exit code 0
```

### AC4: SSH Connection Works
```bash
# Can execute command via SSH
ssh -o StrictHostKeyChecking=no root@${VM_IP} 'hostname'
# Expected: exit code 0, prints hostname
```

### AC5: No Key = No Failure
```bash
# VM without --ssh-key still boots fine
nanofuse vm create myimage:latest test-nokey
nanofuse vm start test-nokey
# Expected: VM boots, no errors about missing key
```

### AC6: Invalid Key Rejected
```bash
# Non-public-key file rejected
echo "not a key" > /tmp/badkey
nanofuse vm create myimage:latest test-bad --ssh-key /tmp/badkey
# Expected: error "invalid SSH public key format"
```

## Testing Plan

1. Unit test: Key encoding/validation
2. Integration test: Create VM with --ssh-key, verify SSH works
3. Manual test: SSH into running VM

## Files Changed

| File | Change |
|------|--------|
| `cmd/nanofuse/vm_create.go` | Add --ssh-key flag |
| `internal/types/vm.go` | Add SSHPublicKey field |
| `internal/firecracker/vm.go` | Append sshkey to kernel args |
| `internal/api/handlers.go` | Pass through SSH key |
| `images/base/Dockerfile` | Add install-ssh-key script/service |
| `images/base/units/install-ssh-key.service` | New file |
| `images/base/scripts/install-ssh-key.sh` | New file |
| `examples/todo-app/docker/Dockerfile` | Add openssh-server |

## Rollout

1. Update base image with install-ssh-key service
2. Update todo-app to include openssh-server
3. Add CLI flag
4. Document usage
