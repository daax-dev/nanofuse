# NanoFuse CLI Integration Testing Guide

## Prerequisites

Before testing the CLI, ensure:

1. **API Daemon Running**
   ```bash
   sudo systemctl start nanofused
   sudo systemctl status nanofused
   ```

2. **Socket Permissions**
   ```bash
   ls -l /var/run/nanofused.sock
   # Should show: srw-rw---- 1 root nanofuse

   # Add your user to nanofuse group if needed
   sudo usermod -aG nanofuse $USER
   newgrp nanofuse
   ```

3. **Build CLI**
   ```bash
   cd /home/jpoley/src/_mine/nanofuse
   make build-cli
   # Binary created at: cmd/nanofuse/nanofuse
   ```

4. **Install CLI (optional)**
   ```bash
   sudo make install
   # Installs to: /usr/local/bin/nanofuse
   ```

## Test Suite

### 1. Health Check

```bash
# Test: API connectivity
./cmd/nanofuse/nanofuse health

# Expected output:
# Status:   healthy
# Version:  0.1.0
# Uptime:   ...

# Test: JSON output
./cmd/nanofuse/nanofuse health --json

# Expected: Valid JSON with status, version, uptime_seconds
```

### 2. Version Information

```bash
# Test: CLI version
./cmd/nanofuse/nanofuse version

# Expected output:
# CLI Version:  0.1.0
# API Version:  0.1.0
# Git Commit:   ...
# Built:        ...
# Go Version:   go1.22
# Platform:     linux/amd64
```

### 3. Configuration

```bash
# Test: Config init
./cmd/nanofuse/nanofuse config init

# Expected: Creates ~/.config/nanofuse/config.yaml

# Test: Config view
./cmd/nanofuse/nanofuse config view

# Expected: Shows current configuration
```

### 4. Image Operations

```bash
# Test: Pull image (requires valid image in registry)
./cmd/nanofuse/nanofuse image pull ghcr.io/owner/nanofuse-base:latest

# Expected:
# - Progress bar showing download
# - "Pull complete!" message
# - Image digest displayed

# Test: List images
./cmd/nanofuse/nanofuse image list

# Expected: Table showing pulled images

# Test: List images (JSON)
./cmd/nanofuse/nanofuse image list --json

# Expected: JSON array of images

# Test: Inspect image
./cmd/nanofuse/nanofuse image inspect sha256:abc123...

# Expected: Detailed image information

# Test: Remove image (clean up)
# ./cmd/nanofuse/nanofuse image remove sha256:abc123...
```

### 5. VM Lifecycle (Basic)

```bash
# Test: Create VM (without starting)
./cmd/nanofuse/nanofuse vm create ghcr.io/owner/nanofuse-base:latest test-vm-1 \
  --vcpus 2 \
  --memory 512

# Expected:
# - "Created VM: test-vm-1"
# - VM ID displayed
# - State: created

# Test: List VMs
./cmd/nanofuse/nanofuse vm list

# Expected: Table showing test-vm-1 in "created" state

# Test: List VMs (JSON)
./cmd/nanofuse/nanofuse vm list --json

# Expected: JSON array with VM details

# Test: VM status
./cmd/nanofuse/nanofuse vm status test-vm-1

# Expected: Detailed VM information

# Test: Start VM
./cmd/nanofuse/nanofuse vm start test-vm-1

# Expected:
# - "Starting VM: test-vm-1"
# - "VM started successfully!"
# - State: running
# - IP address displayed

# Test: Verify running
./cmd/nanofuse/nanofuse vm status test-vm-1

# Expected: State = running, runtime info populated

# Test: VM logs
./cmd/nanofuse/nanofuse vm logs test-vm-1 --tail 50

# Expected: Last 50 lines of console output

# Test: Pause VM
./cmd/nanofuse/nanofuse vm pause test-vm-1

# Expected:
# - "Pausing VM: test-vm-1"
# - "VM paused successfully!"
# - State: paused

# Test: Resume VM
./cmd/nanofuse/nanofuse vm resume test-vm-1

# Expected:
# - "Resuming VM: test-vm-1"
# - "VM resumed successfully!"
# - State: running

# Test: Restart VM
./cmd/nanofuse/nanofuse vm restart test-vm-1

# Expected:
# - "Restarting VM: test-vm-1"
# - "Stopping VM..."
# - "VM stopped!"
# - "Starting VM..."
# - "VM started successfully!"

# Test: Stop VM
./cmd/nanofuse/nanofuse vm stop test-vm-1 --timeout 30

# Expected:
# - "Stopping VM: test-vm-1"
# - "Sending shutdown signal..."
# - "VM stopped successfully!"
# - State: stopped

# Test: Delete VM
./cmd/nanofuse/nanofuse vm delete test-vm-1 --force

# Expected:
# - "Deleted VM: test-vm-1"

# Verify deletion
./cmd/nanofuse/nanofuse vm list
# test-vm-1 should not appear
```

### 6. VM Lifecycle (Run shorthand)

```bash
# Test: Create and start in one command
./cmd/nanofuse/nanofuse vm run ghcr.io/owner/nanofuse-base:latest test-vm-2 \
  --vcpus 4 \
  --memory 1024

# Expected:
# - "Created VM: test-vm-2"
# - "Starting VM..."
# - "VM started successfully!"
# - VM ID, State, IP displayed
# - SSH command suggestion

# Clean up
./cmd/nanofuse/nanofuse vm stop test-vm-2
./cmd/nanofuse/nanofuse vm delete test-vm-2 --force
```

### 7. Snapshot Operations

```bash
# Setup: Create and start a VM
./cmd/nanofuse/nanofuse vm run ghcr.io/owner/nanofuse-base:latest snap-test-vm

# Wait for VM to boot (give it 10 seconds)
sleep 10

# Test: Create snapshot
./cmd/nanofuse/nanofuse vm snapshot create snap-test-vm after-boot

# Expected:
# - "Creating snapshot for VM: snap-test-vm"
# - "Snapshot created successfully!"
# - Snapshot ID and details
# - Resume command hint

# Test: List snapshots
./cmd/nanofuse/nanofuse vm snapshot list snap-test-vm

# Expected: Table showing "after-boot" snapshot

# Test: List snapshots (JSON)
./cmd/nanofuse/nanofuse vm snapshot list snap-test-vm --json

# Expected: JSON array of snapshots

# Test: Inspect snapshot
# (use snapshot ID from create output)
./cmd/nanofuse/nanofuse vm snapshot inspect snapshot-20251030-100530

# Expected: Detailed snapshot information

# Test: Resume from snapshot
./cmd/nanofuse/nanofuse vm stop snap-test-vm
./cmd/nanofuse/nanofuse vm resume snap-test-vm --from-snapshot snapshot-20251030-100530

# Expected:
# - "Resuming VM: snap-test-vm from snapshot: ..."
# - "VM resumed successfully!"
# - Fast resume time (< 2s)

# Test: Delete snapshot
./cmd/nanofuse/nanofuse vm snapshot delete snapshot-20251030-100530 --force

# Expected: "Deleted snapshot: ..."

# Clean up VM
./cmd/nanofuse/nanofuse vm stop snap-test-vm
./cmd/nanofuse/nanofuse vm delete snap-test-vm --force
```

### 8. Error Handling

```bash
# Test: VM not found
./cmd/nanofuse/nanofuse vm status nonexistent-vm

# Expected:
# - Error: "Failed to get VM status: Virtual machine ... does not exist"
# - Hint: "Run 'nanofuse vm list' to see available VMs"
# - Exit code: 4

echo $?
# Should be 4

# Test: Image not found
./cmd/nanofuse/nanofuse image inspect sha256:invalid

# Expected:
# - Error message
# - Hint: "Pull the image first..."
# - Exit code: 4

# Test: Invalid state transition
# (try to start an already running VM)
./cmd/nanofuse/nanofuse vm run ghcr.io/owner/base:latest state-test-vm
./cmd/nanofuse/nanofuse vm start state-test-vm

# Expected:
# - Error: "Invalid operation" or "VM already running"
# - Hint: "Check VM status..."
# - Exit code: 5

# Clean up
./cmd/nanofuse/nanofuse vm stop state-test-vm
./cmd/nanofuse/nanofuse vm delete state-test-vm --force

# Test: API unreachable
sudo systemctl stop nanofused
./cmd/nanofuse/nanofuse health

# Expected:
# - Error: "Cannot connect to API"
# - Details about connection failure
# - Hint: "Ensure nanofused service is running..."
# - Exit code: 3

sudo systemctl start nanofused
```

### 9. Debug Mode

```bash
# Test: Debug output
./cmd/nanofuse/nanofuse --debug vm list

# Expected:
# - DEBUG: GET /vms
# - DEBUG: Response: 200 OK
# - DEBUG: Body: {...}
# - Normal output follows
```

### 10. Output Formats

```bash
# Test: Default table output
./cmd/nanofuse/nanofuse vm list

# Expected: Formatted ASCII table

# Test: JSON output
./cmd/nanofuse/nanofuse vm list --json

# Expected: Valid JSON that can be piped to jq
./cmd/nanofuse/nanofuse vm list --json | jq .

# Test: Color disabled
./cmd/nanofuse/nanofuse --no-color vm list

# Expected: No ANSI color codes in output

# Test: NO_COLOR environment variable
NO_COLOR=1 ./cmd/nanofuse/nanofuse vm list

# Expected: No ANSI color codes in output
```

### 11. Flags and Options

```bash
# Test: Custom socket path
./cmd/nanofuse/nanofuse --api-socket /tmp/custom.sock health

# Test: Custom timeout
./cmd/nanofuse/nanofuse --timeout 5s health

# Test: Remote API (TCP)
./cmd/nanofuse/nanofuse --api-url http://localhost:8080 health

# Test: Help
./cmd/nanofuse/nanofuse --help
./cmd/nanofuse/nanofuse vm --help
./cmd/nanofuse/nanofuse vm create --help
```

### 12. Shell Completion

```bash
# Test: Generate bash completion
./cmd/nanofuse/nanofuse completion bash

# Expected: Valid bash completion script

# Test: Generate zsh completion
./cmd/nanofuse/nanofuse completion zsh

# Expected: Valid zsh completion script

# Test: Generate fish completion
./cmd/nanofuse/nanofuse completion fish

# Expected: Valid fish completion script

# Test: Install and use (bash example)
source <(./cmd/nanofuse/nanofuse completion bash)
./cmd/nanofuse/nanofuse vm <TAB><TAB>
# Should show: create, delete, kill, list, logs, pause, restart, resume, run, snapshot, start, status, stop
```

## Automated Test Script

Save as `test-cli.sh`:

```bash
#!/bin/bash
set -e

CLI="./cmd/nanofuse/nanofuse"
TEST_IMAGE="ghcr.io/owner/nanofuse-base:latest"

echo "=== NanoFuse CLI Integration Tests ==="
echo ""

# 1. Health check
echo "1. Testing health check..."
$CLI health
echo "✓ Health check passed"
echo ""

# 2. Version
echo "2. Testing version..."
$CLI version
echo "✓ Version check passed"
echo ""

# 3. Image operations
echo "3. Testing image operations..."
echo "Skipping image pull (may take time)"
$CLI image list
echo "✓ Image list passed"
echo ""

# 4. VM lifecycle
echo "4. Testing VM lifecycle..."
VM_NAME="test-vm-$$"

echo "Creating VM..."
$CLI vm create $TEST_IMAGE $VM_NAME --vcpus 2 --memory 512

echo "Listing VMs..."
$CLI vm list

echo "Getting VM status..."
$CLI vm status $VM_NAME

echo "Starting VM..."
$CLI vm start $VM_NAME

echo "Checking running status..."
$CLI vm status $VM_NAME

echo "Stopping VM..."
$CLI vm stop $VM_NAME

echo "Deleting VM..."
$CLI vm delete $VM_NAME --force

echo "✓ VM lifecycle tests passed"
echo ""

# 5. JSON output
echo "5. Testing JSON output..."
$CLI vm list --json | jq .
echo "✓ JSON output passed"
echo ""

echo "=== All tests passed! ==="
```

Make executable and run:
```bash
chmod +x test-cli.sh
./test-cli.sh
```

## Expected Results

### Success Criteria

All commands should:
- ✅ Execute without crashing
- ✅ Return appropriate exit codes
- ✅ Display formatted output (table or JSON)
- ✅ Show clear error messages on failure
- ✅ Provide helpful hints for common errors
- ✅ Handle interrupts gracefully (Ctrl+C)

### Known Issues to Watch For

1. **Permission Denied** - User not in `nanofuse` group
2. **Connection Refused** - API daemon not running
3. **Timeout Errors** - Operations taking longer than configured timeout
4. **Invalid State** - Trying operations on VMs in wrong state

## Reporting Issues

When reporting issues, include:

1. **Command executed**:
   ```bash
   ./cmd/nanofuse/nanofuse vm start test-vm
   ```

2. **Error output**:
   ```
   Error: Failed to start VM: ...
   ```

3. **Debug output**:
   ```bash
   ./cmd/nanofuse/nanofuse --debug vm start test-vm
   ```

4. **Environment**:
   - CLI version: `./cmd/nanofuse/nanofuse version`
   - API status: `sudo systemctl status nanofused`
   - Socket permissions: `ls -l /var/run/nanofused.sock`

## Next Steps

Once integration tests pass:

1. Document any issues or deviations from spec
2. Update CLI_SPEC.md if behavior differs
3. Add CLI to CI/CD pipeline
4. Create release artifacts (binaries for x86_64, arm64)
5. Write user documentation and tutorials

## Automation Opportunities

After manual testing, consider:

- Automated integration test suite (Go tests with API running)
- Mock API server for offline testing
- Performance benchmarks (command execution time)
- Stress testing (many concurrent commands)
- Regression test suite (prevent breaking changes)
