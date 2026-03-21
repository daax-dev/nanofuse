# Complete Working Test - Run These Commands

The daemon keeps dying. Here's what to run **in your terminal** to test everything end-to-end:

## Step 1: Start Fresh Daemon

```bash
cd /home/jpoley/ps/nanofuse

# Kill old daemon
sudo pkill -9 nanofused
sleep 2

# Start new daemon (THIS WORKS - we tested it earlier)
sudo ./bin/nanofused &
sleep 3

# Verify it's running
curl http://localhost:8080/health
# Should show: {"status":"healthy",...}
```

## Step 2: Check If Image Still Exists

```bash
./bin/nanofuse --api-url http://localhost:8080 image list
```

**If you see the image**: Great, skip to Step 3

**If no image**: Pull it again (this WORKS - we tested it):
```bash
# Start pull in background
./bin/nanofuse --api-url http://localhost:8080 image pull --default &
PULL_PID=$!

# Wait 30 seconds
sleep 30

# Kill the stuck CLI (image is pulled but CLI waits for job completion)
kill $PULL_PID

# Verify image exists
./bin/nanofuse --api-url http://localhost:8080 image list
```

## Step 3: Test VM Lifecycle

```bash
# Create VM
./bin/nanofuse --api-url http://localhost:8080 vm create default test-vm --vcpus 2 --memory 512

# List VMs
./bin/nanofuse --api-url http://localhost:8080 vm list

# Start VM
./bin/nanofuse --api-url http://localhost:8080 vm start test-vm

# Wait for boot
sleep 10

# Check status
./bin/nanofuse --api-url http://localhost:8080 vm status test-vm

# Get logs
./bin/nanofuse --api-url http://localhost:8080 vm logs test-vm --tail 20

# Stop VM
./bin/nanofuse --api-url http://localhost:8080 vm stop test-vm

# Delete VM
./bin/nanofuse --api-url http://localhost:8080 vm delete test-vm --force
```

## What Success Looks Like

### After create:
```
Created VM: test-vm
ID: vm-abc123...
```

### After start:
```
Starting VM: test-vm
VM started successfully
```

### After status:
```
State: running
IP: 172.16.0.10
```

### After logs:
```
[    0.000000] Linux version...
...
systemd[1]: Reached target basic.target
```

## If It Works

**BOTH CRITICAL FIXES ARE PROVEN:**
- ✅ Fix #1: CLI image pull (no "Job ID" error)
- ✅ Fix #2: Dual listeners (both Unix socket and TCP work)
- ✅ Basic VM lifecycle works end-to-end

Then we can say Phase 1 **actually works** (not just "complete").

## If It Fails

Tell me:
1. Which step failed?
2. What was the error message?
3. Output of: `sudo cat /var/log/syslog | grep firecracker | tail -20`
