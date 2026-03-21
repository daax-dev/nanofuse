# Quick Start: Todo-App Diagnostic

## TL;DR - Run This Now

```bash
cd /home/jpoley/ps/nanofuse
chmod +x scripts/diagnose-and-fix-todo-app.sh
sudo ./scripts/diagnose-and-fix-todo-app.sh
```

**Time**: ~2 minutes
**Output**: `/home/jpoley/ps/nanofuse/examples/todo-app/diagnostic-output/`

## What It Does

The diagnostic script performs:

1. ✅ **Environment Check** - Validates daemon, API, image files
2. 📋 **Console Log Analysis** - Examines boot sequence
3. 🔍 **Rootfs Inspection** - Deep filesystem analysis
4. ⚙️ **Kernel Args Validation** - Checks boot configuration
5. 🔧 **Fix Generation** - Creates automated repair scripts
6. 📊 **Report Creation** - Comprehensive findings document

## After Running

### Option 1: Apply Systemd Fixes (Recommended)

```bash
cd examples/todo-app/diagnostic-output

# Apply rootfs fixes
sudo ./apply-fixes.sh

# Create VM with enhanced configuration
./create-vm-enhanced.sh

# Monitor results
nanofuse vm logs my-todo-app --tail 100
```

### Option 2: Use Simple Init (If systemd fails)

```bash
cd examples/todo-app/diagnostic-output

# Install simple init
sudo ./install-simple-init.sh

# Create VM with simple init
nanofuse vm create sha256:0c854... my-todo-app \
  --kernel-args "console=ttyS0 root=/dev/vda1 rw init=/init"

nanofuse vm start my-todo-app
```

## Expected Results

After fix application:

```bash
# Should work:
curl http://172.16.0.11/health
# {"status":"ok","timestamp":"..."}

curl http://172.16.0.11/api/todos
# {"todos":[]}

# Ports should be open:
nmap -p 80,8080 172.16.0.11
# 80/tcp   open  http
# 8080/tcp open  http-proxy
```

## Troubleshooting

### If diagnostic fails:
1. Ensure `nanofused` is running: `systemctl status nanofused`
2. Check API: `curl http://localhost:8080/health`
3. Verify image exists: `ls -lh /var/lib/nanofuse/images/sha256:*/rootfs.ext4`

### If fix doesn't work:
1. Check console logs: `cat /var/lib/nanofuse/vms/*/console.log`
2. Try simple init (Option 2)
3. Review diagnostic report: `less diagnostic-output/diagnostic-report.md`

## Files Generated

```
diagnostic-output/
├── diagnostic-report.md          # Main report
├── console-*.log                  # Boot sequence
├── systemd-messages.txt           # Systemd output
├── boot-errors.txt               # Error messages
├── todo-backend.service          # Service definition
├── firecracker-config.json       # VM configuration
├── apply-fixes.sh                # Automated fix script
├── create-vm-enhanced.sh         # Enhanced VM creation
├── simple-init.sh                # Alternative init
└── install-simple-init.sh        # Install alternative
```

## More Information

- **Detailed Analysis**: See `PLATFORM_ENGINEERING_ANALYSIS.md`
- **Original Issue**: See `PRIORITY_TODO.md`
- **Architecture**: See `README.md`

## Quick Reference: Key Commands

```bash
# Run diagnostic
sudo ./scripts/diagnose-and-fix-todo-app.sh

# View console logs
cat /var/lib/nanofuse/vms/*/console.log | tail -100

# Check VM status
nanofuse vm list

# Get VM IP
nanofuse vm list --json | jq -r '.vms[] | select(.name=="my-todo-app") | .config.network.ip_address'

# Test connectivity
ping -c 3 172.16.0.11

# Check ports
nmap -p 80,8080 172.16.0.11

# View diagnostic report
less examples/todo-app/diagnostic-output/diagnostic-report.md
```

## Support

If issues persist after following this guide:
1. Review the full diagnostic report
2. Check the Platform Engineering Analysis
3. Examine all generated artifacts
4. Consider manual rootfs inspection

---

**Remember**: The diagnostic is non-destructive. It only reads and analyzes. Apply fixes only after reviewing the findings.
