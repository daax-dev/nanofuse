# E2E Testing

## Overview

End-to-end (E2E) testing validates the complete nanofuse lifecycle, from daemon startup through VM creation, SSH connectivity, and cleanup.

## E2E Test Lifecycle

```
1. Stop daemon (clean slate)
2. Build/reinstall nanofused
3. Start daemon
4. Register image
5. Create VM
6. Start VM
7. Test SSH connectivity
8. Test HTTP connectivity (curl)
9. Cleanup (stop VM, delete VM)
```

## Requirements

### Local Requirements

| Requirement | Check Command | Notes |
|-------------|---------------|-------|
| KVM access | `test -e /dev/kvm` | Required for Firecracker |
| Root/sudo | `sudo -n true` | Daemon operations need sudo |
| Firecracker | `which firecracker` | Must be installed |
| 512MB+ RAM | `free -m` | For VM allocation |

### CI Requirements

See [GitHub Actions E2E Feasibility](#github-actions-e2e-feasibility) below.

## Running E2E Tests

```bash
# Full E2E test (requires sudo and KVM)
sudo mage TestE2E

# Run standalone script
sudo scripts/e2e-test.sh

# Run with gdt
sudo go test -v ./test/gdt/e2e/...

# Skip cleanup for debugging
E2E_SKIP_CLEANUP=1 sudo mage TestE2E
```

## Test Location

```
test/
├── e2e/
│   ├── e2e_test.go           # Go test wrapper
│   ├── lifecycle_test.go     # Daemon lifecycle tests
│   └── connectivity_test.go  # SSH/HTTP tests
├── gdt/
│   └── e2e/
│       ├── e2e_test.go       # gdt test wrapper
│       └── full_lifecycle.yaml
scripts/
└── e2e-test.sh               # Standalone E2E script
```

## Test Phases

### Phase 1: Stop Daemon

```bash
sudo systemctl stop nanofused || true
# Wait for clean shutdown
sleep 2
```

### Phase 2: Build and Install

```bash
mage All
sudo cp bin/nanofused /usr/local/bin/
sudo cp bin/nanofuse /usr/local/bin/
```

### Phase 3: Start Daemon

```bash
sudo systemctl start nanofused
# Wait for daemon to be ready
sleep 2
```

### Phase 4: Register Image

```bash
sudo bin/register-local-image images/base/rootfs.ext4 base
```

### Phase 5: Create and Start VM

```bash
sudo nanofuse vm create e2e-test --image base
sudo nanofuse vm start e2e-test

# Wait for VM to boot (up to 30 seconds)
for i in {1..30}; do
    if nanofuse vm status e2e-test | grep -q "running"; then
        break
    fi
    sleep 1
done
```

### Phase 6: Test SSH

```bash
VM_IP=$(nanofuse vm show e2e-test --format json | jq -r '.ip_address')
ssh -o ConnectTimeout=30 -o StrictHostKeyChecking=no root@${VM_IP} "echo 'SSH OK'"
```

### Phase 7: Test HTTP

```bash
# Test outbound connectivity from VM
ssh -o StrictHostKeyChecking=no root@${VM_IP} \
    "wget -q -O- http://example.com | head -1"
```

### Phase 8: Cleanup

```bash
sudo nanofuse vm stop e2e-test
sudo nanofuse vm delete e2e-test
```

## GitHub Actions E2E Feasibility

### Summary

| Runner Type | E2E Support | Notes |
|-------------|-------------|-------|
| GitHub-hosted | NO | No KVM access |
| Self-hosted (bare metal) | YES | Full KVM support |
| Self-hosted (nested virt) | MAYBE | Depends on cloud provider |
| Fireactions | YES | Purpose-built for Firecracker |
| Actuated | YES | Commercial, free for OSS |

### GitHub-Hosted Runners

**Cannot run Firecracker E2E tests** because:
- `/dev/kvm` is not exposed
- Nested virtualization not available
- Standard runners use containers without KVM passthrough

**Source**: [GitHub Actions Runner Images](https://github.com/actions/runner-images)

### Recommended CI Strategy

```yaml
# .github/workflows/ci.yaml

jobs:
  # Unit tests run on GitHub-hosted
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: mage Test

  # E2E tests run on self-hosted (optional)
  e2e-tests:
    runs-on: self-hosted
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    needs: [unit-tests]
    steps:
      - uses: actions/checkout@v4
      - run: sudo mage TestE2E
```

### Self-Hosted Runner Options

1. **Fireactions** (Recommended for Firecracker projects)
   - https://github.com/hostinger/fireactions
   - Runs jobs inside Firecracker VMs
   - Native KVM support

2. **Actuated**
   - https://actuated.dev/
   - ARM64 and x86 with KVM
   - Free tier for open source

3. **AWS Metal Instances**
   - `m5.metal`, `c5.metal`, etc.
   - Full KVM support
   - Higher cost

4. **GCP Nested Virtualization**
   - Enable on N1/N2 instances
   - `--enable-nested-virtualization`

### Hybrid Approach

For most projects, a hybrid approach works best:

1. **GitHub-hosted**: Unit tests, lint, build
2. **Self-hosted** (optional): E2E tests on main branch
3. **Manual**: Developers run E2E locally before merge

## Troubleshooting

### SSH Connection Refused

```bash
# Check VM is running
nanofuse vm status e2e-test

# Check IP assignment
nanofuse vm show e2e-test --format json | jq '.ip_address'

# Check network namespace
sudo ip netns exec nanofuse-net ip addr

# Check firewall
sudo iptables -L -n
```

### VM Won't Start

```bash
# Check KVM access
ls -la /dev/kvm

# Check Firecracker logs
sudo journalctl -u nanofused -f

# Check VM console logs
nanofuse vm logs e2e-test
```

### Timeout Waiting for Boot

```bash
# Increase timeout
E2E_BOOT_TIMEOUT=60 sudo scripts/e2e-test.sh

# Check for kernel panic
nanofuse vm logs e2e-test | grep -i panic

# Verify image is valid
file images/base/rootfs.ext4
```

## References

- [Firecracker Getting Started](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md)
- [GitHub Actions Self-Hosted Runners](https://docs.github.com/en/actions/hosting-your-own-runners)
- [Fireactions](https://github.com/hostinger/fireactions)
- [Actuated](https://actuated.dev/)
