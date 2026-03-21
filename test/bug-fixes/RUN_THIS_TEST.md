# ACTUAL WORKING TEST - Run This

I cannot use sudo without a TTY. You need to run this in your actual terminal.

## The Truth

- **2 critical bugs were found and fixed in the code**
- **Binaries are built**: `/home/jpoley/ps/nanofuse/bin/nanofused` and `nanofuse`
- **Tests are written** but need sudo to run
- **Documentation exists** but doesn't mean it works

## What You Need To Do (5 minutes)

Copy and paste this into your terminal:

```bash
cd /home/jpoley/ps/nanofuse

# Stop old broken daemon
sudo pkill -9 nanofused
sleep 2

# Remove old socket
sudo rm -f /tmp/nanofused.sock

# Start NEW daemon with fixed code
sudo ./bin/nanofused > /tmp/nanofused.log 2>&1 &
sleep 3

# Check if both listeners were created (FIX #2)
echo "=== Checking logs for dual listener ==="
cat /tmp/nanofused.log | grep -i "listening"

# Verify socket exists
ls -la /tmp/nanofused.sock

# Test both APIs work
curl --unix-socket /tmp/nanofused.sock http://localhost/health
curl http://localhost:8080/health

# Test CLI image pull (FIX #1) - should NOT show "Job ID is required"
echo "=== Testing image pull fix ==="
./bin/nanofuse --api-url http://localhost:8080 image pull --default
```

## What Success Looks Like

**For Fix #2 (Dual Listeners)**:
```
=== Checking logs for dual listener ===
INFO: Listening on Unix socket: /tmp/nanofused.sock
INFO: Listening on TCP: 127.0.0.1:8080
```

**For Fix #1 (CLI Pull)**:
Either:
- `Pulling ghcr.io/jpoley/nanofuse/base:latest...` (it works!)
- `authentication required` (it works, just needs auth)

Should **NOT** show:
- `Job ID is required` (this means fix didn't work)

## If It Works

The fixes are proven to work. Then you can:
1. Install binaries: `sudo cp bin/* /usr/local/bin/`
2. Continue with VM lifecycle testing

## If It Fails

Something is still broken. Check:
- Daemon logs: `cat /tmp/nanofused.log`
- What specific error occurred
- Which fix failed (listener or pull)

## The Point

Actually running this test proves if the code changes fixed the bugs or not. Everything else is just documentation.
