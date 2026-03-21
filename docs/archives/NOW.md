# What We're Working On NOW

**Last Updated**: 2025-11-04 AM
**Current Phase**: Phase 1 - Port Forwarding Debug

---

## ACTIVE WORK 🔥

### Port Forwarding - FIX IMPLEMENTED ✅

**Root Causes Found:**
1. Docker rules in OUTPUT/PREROUTING run first (intercepted traffic)
2. curl tries IPv6 (::1) first, our rules are IPv4 only (127.0.0.1)

**Fixes Applied:**
- iptables: Changed `-A` (APPEND) to `-I chain 1` (INSERT at top)
- iptables: Added `-d 127.0.0.1` to OUTPUT rule
- Test script: Added `curl -4` to force IPv4

**Files Changed:**
- `internal/network/portforward.go` - Added `-d 127.0.0.1` to all OUTPUT chain rules
- `internal/network/bridge.go` - Added explicit route verification

**Testing:**
Terminal 1: `sudo ./scripts/test-network-e2e.sh`
Terminal 2 (when test is waiting): `./debug-while-test-running.sh`

**Debug Scripts Ready:**
1. `debug-portforward-tcpdump.sh` - Packet tracing
2. `test-alternative-services.sh` - Test SSH/netcat/other services
3. `inspect-iptables-rules.sh` - Rule inspection
4. `scripts/vm-port-forward.sh` - Port forward management

**Debug Plan:**
1. Run tcpdump packet tracing (30 min)
2. Test alternative services (20 min)
3. Check conntrack/routing/filtering (15 min)
4. Implement fallback if needed (30 min)
   - Option A: socat for localhost forwarding
   - Option B: Document limitation
   - Option C: Hybrid approach

**Possible Root Causes:**
- Routing table missing route to 172.16.0.0/24
- Reverse path filtering blocking packets
- Bridge netfilter interference
- Conntrack state issues
- OUTPUT chain not triggering routing recalculation

---

## Action Plan for Today

1. **YOU:** Test Phase 1 features (5-10 min) ✅
2. **ME:** Debug port forwarding (1-2 hours) ⏳
3. **BOTH:** Verify fix works
4. **DONE:** Phase 1 fully complete

---

## Next Up (After Port Forwarding)

**Phase 2 Priority 1:** Snapshot/Resume

---

## Quick Reference

### Test Phase 1 Features
```bash
sudo ./scripts/test-network-e2e.sh
```

### Debug Files

- `DEBUG_PORT_FORWARD_TOMORROW.md` - Full debug plan
- `debug-portforward-tcpdump.sh` - Main debug script
- `internal/network/portforward.go` - Implementation
- `scripts/vm-port-forward.sh` - CLI tool

### Completed Work
See `DONE.md` for all Phase 1 completed features
