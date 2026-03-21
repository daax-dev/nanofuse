# NanoFuse: Next Steps - START HERE

**Date**: 2025-11-23
**Status**: Actionable Plan
**Time Commitment**: 1-2 weeks to Phase 1 complete

---

## 🎯 What This Document Is

This is your **actionable playbook** for completing Phase 1 of NanoFuse using SVPG Product Operating Model principles.

**Read this if**:
- You want to know what to do RIGHT NOW
- You're overwhelmed by planning docs
- You need a clear path forward
- You want outcome-focused work

**Skip the theory, get to work** ⬇️

---

## 📊 Current Situation (Honest Assessment)

### What Works ✅
- VM boots successfully
- Network configured (172.16.0.10, pingable)
- GHCR authentication
- Image pull and extraction
- CLI and daemon build

### What's Broken ❌
- Services (nginx, todo-backend) not running
- Ports 80 and 8080 closed
- No end-to-end test has ever passed

### What This Means
**We have 60% of infrastructure, but 0% of user value.**

A VM that can't run services is useless. Fix this first.

---

## 🎯 Phase 1 Outcome (What Success Looks Like)

**Outcome**: "Developer can deploy a working VM"

**Measurable Success**:
```bash
# This sequence works 10/10 times:
nanofuse image pull --default
nanofuse vm run default my-vm
curl http://172.16.0.10:80           # Returns HTML
curl http://172.16.0.10:8080/health  # Returns {"status":"healthy"}
```

**NOT Success**:
- "I wrote 1000 lines of code"
- "I documented 10 features"
- "I planned 5 more features"

---

## 📅 Week 1 Plan (2025-11-23 to 2025-11-30)

### Monday: Diagnose (4 hours)

**Goal**: Understand why services fail

**Tasks**:
1. Read full console log systematically
2. Run comprehensive test script
3. Identify root cause with evidence
4. Document hypothesis

**Commands**:
```bash
# Get current VM ID
nanofuse vm list

# Read console log
sudo tail -200 /var/lib/nanofuse/vms/<VM_ID>/console.log

# Look for specific patterns
sudo grep "systemd\[1\]" /var/lib/nanofuse/vms/<VM_ID>/console.log | head -20
sudo grep "nginx.service" /var/lib/nanofuse/vms/<VM_ID>/console.log
sudo grep "\[FAILED\]" /var/lib/nanofuse/vms/<VM_ID>/console.log
```

**Output**: Decision record in `backlog/decisions/service-startup-root-cause.md`

**Go/No-Go**: If no root cause in 4 hours, ask for help

---

### Tuesday: Fix (4 hours)

**Goal**: Get services running

**Likely Fixes**:

**Option A**: Missing init parameter
```go
// In internal/api/vm_handlers.go
KernelArgs: "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd ip=..."
```

**Option B**: Service dependencies
```bash
# Mount rootfs, check service files
sudo mkdir -p /tmp/nanofuse-mount
sudo mount -o loop /path/to/rootfs.ext4 /tmp/nanofuse-mount
cat /tmp/nanofuse-mount/etc/systemd/system/nginx.service
# Fix After= dependencies if needed
sudo umount /tmp/nanofuse-mount
```

**Commands**:
```bash
# After fix: rebuild and test
cd /home/jpoley/ps/nanofuse
mage daemon
sudo systemctl stop nanofused
sudo cp bin/nanofused /usr/local/bin/
sudo systemctl start nanofused

# Create fresh VM
nanofuse vm delete my-vm --force
nanofuse vm run default my-vm

# Test services
curl http://172.16.0.10:80
curl http://172.16.0.10:8080/health
```

**Output**: Services accessible from host

**Go/No-Go**: If fix doesn't work, try alternative approach

---

### Wednesday: Automate (4 hours)

**Goal**: Never lose working config again

**Task 1**: Health check script
```bash
#!/bin/bash
# scripts/building/health-check.sh

VM_ID="$1"
VM_IP=$(nanofuse vm list | grep $VM_ID | awk '{print $3}')

echo "Testing VM: $VM_ID at $VM_IP"

# Check ping
if ! ping -c 1 -W 2 $VM_IP > /dev/null; then
    echo "❌ FAIL: VM not pingable"
    exit 1
fi
echo "✅ PASS: VM pingable"

# Check nginx
if ! curl -s -f http://$VM_IP:80 > /dev/null; then
    echo "❌ FAIL: Nginx not responding"
    exit 1
fi
echo "✅ PASS: Nginx responding"

# Check backend
HEALTH=$(curl -s http://$VM_IP:8080/health | jq -r .status 2>/dev/null)
if [ "$HEALTH" != "healthy" ]; then
    echo "❌ FAIL: Backend unhealthy"
    exit 1
fi
echo "✅ PASS: Backend healthy"

echo ""
echo "✅ ALL CHECKS PASSED"
exit 0
```

**Task 2**: E2E test script
```bash
#!/bin/bash
# test/e2e/full-workflow-test.sh

set -e
trap 'echo "❌ Test failed"; exit 1' ERR

echo "=== E2E Test: Pull to Running VM ==="

# Cleanup any previous state
nanofuse vm delete test-vm --force 2>/dev/null || true

# Pull image
echo "Step 1: Pull image"
nanofuse image pull --default

# Run VM
echo "Step 2: Create VM"
nanofuse vm run default test-vm

# Wait for boot
echo "Step 3: Wait for boot (10s)"
sleep 10

# Run health checks
echo "Step 4: Health checks"
./scripts/building/health-check.sh test-vm

# Cleanup
echo "Step 5: Cleanup"
nanofuse vm delete test-vm --force

echo ""
echo "✅ E2E TEST PASSED"
```

**Output**: Automated validation that catches regressions

---

### Thursday: Document (2 hours)

**Goal**: Never debug the same issue twice

**Task**: Create troubleshooting guide
```markdown
# TROUBLESHOOTING.md

## Problem: Services Not Starting

### Symptoms
- VM boots but ports closed
- curl connection refused
- Console shows [FAILED] messages

### Diagnosis
1. Read console log:
   sudo tail -200 /var/lib/nanofuse/vms/<VM_ID>/console.log

2. Check for systemd startup:
   sudo grep "systemd\[1\]" /var/lib/nanofuse/vms/<VM_ID>/console.log

3. Check for service failures:
   sudo grep "\[FAILED\]" /var/lib/nanofuse/vms/<VM_ID>/console.log

### Common Causes

#### Cause 1: Missing init parameter
**Evidence**: No systemd messages in console
**Fix**: Add init=/lib/systemd/systemd to kernel args

#### Cause 2: Service dependencies
**Evidence**: Services start too early
**Fix**: Add After=network.target to service files

#### Cause 3: Missing binaries
**Evidence**: "No such file" errors
**Fix**: Rebuild image, verify binaries present
```

**Output**: Self-service troubleshooting for future issues

---

### Friday: Review (1 hour)

**Goal**: Assess progress, make go/no-go decision

**Questions to Answer**:
1. ✅ Are services running reliably?
2. ✅ Does E2E test pass 10/10 times?
3. ✅ Is Phase 1 outcome achieved?
4. 🤔 Should we proceed to Phase 2?

**If YES to all**:
- ✅ Phase 1 COMPLETE
- 🎉 Celebrate working system
- 📋 Plan Phase 2 (usability + snapshot research)

**If NO to any**:
- 🔍 Identify blockers
- 🤔 Re-assess approach
- 💬 Seek help if stuck

---

## 🗓️ Week 2 Plan (2025-12-01 to 2025-12-07)

### If Phase 1 Complete

**Monday-Tuesday**: EPIC 3 - Usability
- Add `nanofuse vm logs <name>` command
- Improve error messages with next steps
- Add `--debug` mode

**Wednesday-Friday**: EPIC 4 - Research Snapshot/Resume
- Read Firecracker snapshot docs
- Research systemd compatibility
- Build proof of concept
- Make go/no-go decision

### If Phase 1 Incomplete

**Continue debugging**:
- Maximum 2 more days
- Consider alternative approaches
- Seek external help if needed

**Pivot criteria**:
- If stuck > 2 days with no progress
- If fundamental incompatibility discovered
- If time investment exceeds value

---

## 🚨 Warning Signs (Stop and Re-Assess)

### Feature Factory Indicators
🚨 Adding features before services work
🚨 Writing code without testing
🚨 Planning Phase 3 before Phase 1 done
🚨 Ignoring usability issues
🚨 No actual usage (dogfooding)

### Healthy Behaviors
✅ Fixing broken things before adding features
✅ Testing every change
✅ Completing phases sequentially
✅ Prioritizing usability
✅ Using the tool yourself daily

---

## 📋 Quick Reference Commands

### Check VM Status
```bash
nanofuse vm list
nanofuse vm status <vm-name>
```

### Read Console Logs
```bash
sudo tail -100 /var/lib/nanofuse/vms/<VM_ID>/console.log
```

### Test Services
```bash
# From host
curl http://172.16.0.10:80
curl http://172.16.0.10:8080/health

# Or with VM IP
VM_IP=$(nanofuse vm list | grep my-vm | awk '{print $3}')
curl http://$VM_IP:80
```

### Rebuild and Deploy
```bash
# Rebuild daemon
cd /home/jpoley/ps/nanofuse
mage daemon

# Stop service
sudo systemctl stop nanofused

# Deploy new binary
sudo cp bin/nanofused /usr/local/bin/

# Start service
sudo systemctl start nanofused

# Check status
sudo systemctl status nanofused
```

### Fresh VM
```bash
nanofuse vm delete my-vm --force
nanofuse vm run default my-vm
sleep 10  # Wait for boot
curl http://172.16.0.10:8080/health
```

---

## 🎯 Success Metrics (How You Know It's Working)

### Daily Metrics
- [ ] Health check script passes
- [ ] E2E test passes
- [ ] No manual intervention needed

### Weekly Metrics
- [ ] Used NanoFuse for actual work
- [ ] 10+ VMs created successfully
- [ ] Zero service failures

### Phase 1 Complete When
- [ ] E2E test passes 10/10 times
- [ ] Services start 100% reliably
- [ ] Troubleshooting guide complete
- [ ] Dogfooding successfully

---

## 📚 Additional Reading (If You Want Context)

**Strategic Context**:
- Product Requirements Analysis: `docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md`
- Decision Record: `backlog/decisions/001-product-strategy-outcome-driven-development.md`

**Technical Context**:
- Execution Plan: `docs/building/planning/EXECUTION_PLAN.md`
- Testing Plan: `docs/building/COMPREHENSIVE_TESTING_PLAN.md`
- Phase 1 Investigation: `docs/building/PHASE1CD_COMPREHENSIVE_PLAN.md`

**But honestly**: Skip the reading and start doing. Learn by debugging.

---

## 💡 Key Principles (Remember These)

### 1. Outcomes Over Outputs
❌ "I wrote code"
✅ "Developer can deploy working VM"

### 2. Validate Before Scaling
❌ "Let's add snapshots and Trigger.dev"
✅ "Let's make services work first"

### 3. Evidence Over Assumptions
❌ "It probably works now"
✅ "E2E test passed 10 times"

### 4. Simplicity Over Completeness
❌ "34 CLI commands implemented"
✅ "5 essential commands work perfectly"

### 5. Dogfood Immediately
❌ "I'll use it when it's done"
✅ "I'm using it for real work today"

---

## 🆘 When to Ask for Help

**Ask immediately if**:
- Stuck on root cause for > 4 hours
- Fix doesn't work after 2 attempts
- Fundamental incompatibility discovered
- Not sure how to proceed

**Don't suffer in silence.** Debugging is faster with fresh eyes.

---

## ✅ Checklist for This Week

### Monday (Diagnose)
- [ ] Read console log completely
- [ ] Run systematic testing (Layer 0-6)
- [ ] Identify root cause with evidence
- [ ] Document hypothesis in backlog/decisions/

### Tuesday (Fix)
- [ ] Implement fix based on diagnosis
- [ ] Rebuild and deploy daemon
- [ ] Create fresh VM
- [ ] Verify services accessible

### Wednesday (Automate)
- [ ] Create health check script
- [ ] Create E2E test script
- [ ] Run E2E test 10 times
- [ ] All tests passing

### Thursday (Document)
- [ ] Write troubleshooting guide
- [ ] Document common failures
- [ ] Update README with status
- [ ] Create runbook for deployment

### Friday (Review)
- [ ] Assess progress against outcome
- [ ] Make go/no-go decision for Phase 2
- [ ] Plan next week
- [ ] Celebrate if Phase 1 complete!

---

## 🎉 When You're Done

**You'll know Phase 1 is complete when**:

A developer (even a fresh one) can run these commands and succeed:
```bash
nanofuse image pull --default
nanofuse vm run default demo-vm
curl http://172.16.0.10:80           # ✅ HTML response
curl http://172.16.0.10:8080/health  # ✅ {"status":"healthy"}
```

**10 times in a row. With zero failures.**

That's the outcome. That's success.

---

**Now stop reading and start doing.** 🚀

**First action**: Read that console log and find out why services are failing.

**Location**: `/var/lib/nanofuse/vms/<VM_ID>/console.log`

**Time limit**: 4 hours to root cause

**Go!** 🏃
