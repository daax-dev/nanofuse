# Port Forwarding Debug Plan - PRIORITY FOR TOMORROW

**Status**: Port forwarding iptables rules are created but localhost connections FAIL
**Date**: 2025-11-03
**Priority**: 🔥 HIGH - Blocking Phase 1 completion

## Current Situation

### What Works ✅
- VM boots successfully
- VM network is configured (172.16.0.11)
- HTTP server starts in VM and responds on port 8080
- Direct connection to VM works: `curl http://172.16.0.11:8080` → SUCCESS
- iptables rules are created without errors

### What Fails ❌
- Localhost port forwarding: `curl http://localhost:8888` → TIMEOUT/NO RESPONSE
- Despite trying:
  1. Basic DNAT in OUTPUT + PREROUTING chains
  2. MASQUERADE in POSTROUTING
  3. Source-specific MASQUERADE (-s 127.0.0.1)
  4. Explicit SNAT to bridge IP (172.16.0.1)

### Current iptables Rules

```bash
# PREROUTING (external connections)
iptables -t nat -A PREROUTING -p tcp --dport 8888 -j DNAT --to-destination 172.16.0.11:8080

# OUTPUT (localhost connections)
iptables -t nat -A OUTPUT -p tcp --dport 8888 -j DNAT --to-destination 172.16.0.11:8080

# FORWARD (allow forwarded traffic)
iptables -A FORWARD -p tcp -d 172.16.0.11 --dport 8080 -j ACCEPT

# POSTROUTING (general masquerade)
iptables -t nat -A POSTROUTING -p tcp -d 172.16.0.11 --dport 8080 -j MASQUERADE

# POSTROUTING (localhost-specific SNAT)
iptables -t nat -A POSTROUTING -p tcp -s 127.0.0.1 -d 172.16.0.11 --dport 8080 -j SNAT --to-source 172.16.0.1
```

## Debug Scripts Created

### 1. Comprehensive tcpdump Debug Script

**File**: `debug-portforward-tcpdump.sh`

Run this script to capture packets at every stage of the forwarding path.

### 2. Alternative Service Tests

**File**: `test-alternative-services.sh`

Tests port forwarding with:
- SSH (port 22) - already in base image
- Netcat simple server
- Python http.server on different port
- Busybox httpd

### 3. iptables Rule Inspection

**File**: `inspect-iptables-rules.sh`

Shows rule order, counters, and potential conflicts.

## Debugging Approach for Tomorrow

### Phase 1: Packet Tracing (30 min)

Run tcpdump script to see exactly where packets go:

```bash
sudo ./debug-portforward-tcpdump.sh
```

**What to look for:**
- Do packets leave localhost (lo interface)?
- Do they hit the bridge (nanofuse0)?
- Do they reach the TAP device (tap-xxx)?
- Does the VM respond?
- Where do response packets go?

### Phase 2: Alternative Services (20 min)

If HTTP doesn't work, try simpler services:

```bash
sudo ./test-alternative-services.sh
```

**Tests:**
1. SSH port forward (22 → 22) - SSH is definitely running
2. Netcat echo server - simplest possible test
3. Different HTTP ports - rule out port conflicts

### Phase 3: Connection Tracking (15 min)

Check if conntrack is causing issues:

```bash
# Watch conntrack entries
sudo conntrack -E &
curl http://localhost:8888

# Check existing connections
sudo conntrack -L | grep 8888
```

### Phase 4: Alternative Forwarding Methods (30 min)

If iptables approach keeps failing, try:

**Option A: socat**
```bash
# Simple port forward using socat
socat TCP-LISTEN:8888,reuseaddr,fork TCP:172.16.0.11:8080 &
curl http://localhost:8888
```

**Option B: HAProxy**
```bash
# Install haproxy and configure simple forward
# More reliable than iptables for localhost
```

**Option C: SSH tunnel**
```bash
# Use SSH as port forwarder (requires SSH to be working in VM)
ssh -L 8888:localhost:8080 root@172.16.0.11 -N &
curl http://localhost:8888
```

## Possible Root Causes

### Theory 1: Routing Table Issue
After DNAT in OUTPUT chain, kernel might not know how to route to 172.16.0.0/24.

**Test:**
```bash
ip route show
# Should see route to 172.16.0.0/24 via nanofuse0
```

**Fix if missing:**
```bash
ip route add 172.16.0.0/24 dev nanofuse0
```

### Theory 2: Reverse Path Filtering
RPF might drop packets with unexpected source/dest combinations.

**Test:**
```bash
sysctl net.ipv4.conf.all.rp_filter
sysctl net.ipv4.conf.nanofuse0.rp_filter
```

**Fix:**
```bash
sysctl -w net.ipv4.conf.all.rp_filter=0
sysctl -w net.ipv4.conf.nanofuse0.rp_filter=0
```

### Theory 3: Bridge Filtering
Bridge netfilter might be interfering.

**Test:**
```bash
sysctl net.bridge.bridge-nf-call-iptables
```

**Fix:**
```bash
sysctl -w net.bridge.bridge-nf-call-iptables=0
```

### Theory 4: Conntrack State
Connection might be in wrong conntrack state.

**Test:**
```bash
sudo conntrack -L | grep -E "(8888|8080)"
```

**Fix:**
```bash
# Flush conntrack table
sudo conntrack -F
```

### Theory 5: OUTPUT Chain Routing Decision
Packets modified in OUTPUT chain might not trigger routing recalculation.

**Fix: Use REDIRECT instead of DNAT**
```bash
# REDIRECT is specifically for locally-destined traffic
iptables -t nat -A OUTPUT -p tcp --dport 8888 -j REDIRECT --to-port 8080
# But this only works if service is on localhost, not remote VM
```

## Known Working Port Forward Pattern

From Docker and other tools that DO work:

```bash
# Docker-style port forwarding (known to work)
iptables -t nat -A PREROUTING -p tcp --dport 8888 -j DNAT --to-destination 172.16.0.11:8080
iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 8888 -j DNAT --to-destination 172.16.0.11:8080
iptables -t filter -A FORWARD -d 172.16.0.11 -p tcp --dport 8080 -j ACCEPT
iptables -t nat -A POSTROUTING -s 172.16.0.0/24 ! -d 172.16.0.0/24 -j MASQUERADE
```

**Key difference**: Docker matches on `-d 127.0.0.1` in OUTPUT chain, not just any port.

## Implementation Options if iptables Fails

### Option 1: Use socat for localhost forwarding
- Simple, reliable, well-tested
- One socat process per port forward
- Easy to implement and maintain

### Option 2: Disable localhost forwarding, document limitation
- Port forwards only work from external hosts
- Document: "Access port forwards from another machine or use VM IP directly"
- Simplest solution

### Option 3: Use combination approach
- iptables for external → VM
- socat/haproxy for localhost → VM
- Best of both worlds

## Files to Check Tomorrow

1. `debug-portforward-tcpdump.sh` - Main debugging script
2. `test-alternative-services.sh` - Alternative service tests
3. `inspect-iptables-rules.sh` - Rule inspection
4. `internal/network/portforward.go` - Implementation to modify

## Success Criteria

✅ Localhost port forward works: `curl http://localhost:8888` returns response
✅ External port forward works: `curl http://HOST_IP:8888` returns response (from another machine)
✅ Direct VM access still works: `curl http://172.16.0.11:8080` returns response
✅ Port forward cleanup works: No leaked iptables rules
✅ Multiple port forwards work simultaneously

## Fallback Plan

If after 2 hours iptables approach still doesn't work:

1. **Switch to socat** for localhost forwarding
2. Keep iptables for external forwarding
3. Document limitation in Phase 1
4. Move forward with snapshot/resume (Phase 2)
5. Revisit port forwarding optimization later

## References

- [Docker iptables rules](https://docs.docker.com/network/iptables/)
- [Kubernetes kube-proxy iptables mode](https://kubernetes.io/docs/concepts/services-networking/service/#proxy-mode-iptables)
- [Linux NAT HOWTO](https://www.netfilter.org/documentation/HOWTO/NAT-HOWTO.html)
- [Conntrack zones](https://people.netfilter.org/pablo/docs/login.pdf)

---

**Bottom Line**: We have everything working EXCEPT localhost port forwarding through iptables. Need systematic packet tracing to find where packets are being dropped/misrouted. If iptables proves too problematic, fall back to socat which is known to work reliably for this use case.
