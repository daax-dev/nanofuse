#!/bin/bash
# Inspect iptables rules for port forwarding
# Shows rule order, packet counters, and potential issues

echo "=== iptables Port Forward Rule Inspection ==="
echo ""

echo "1. NAT Table - PREROUTING Chain"
echo "================================"
echo "Purpose: DNAT for external connections"
echo ""
sudo iptables -t nat -L PREROUTING -n -v --line-numbers

echo ""
echo "2. NAT Table - OUTPUT Chain"
echo "==========================="
echo "Purpose: DNAT for localhost connections"
echo ""
sudo iptables -t nat -L OUTPUT -n -v --line-numbers

echo ""
echo "3. NAT Table - POSTROUTING Chain"
echo "================================"
echo "Purpose: SNAT/MASQUERADE for return traffic"
echo ""
sudo iptables -t nat -L POSTROUTING -n -v --line-numbers

echo ""
echo "4. FILTER Table - FORWARD Chain"
echo "==============================="
echo "Purpose: Allow forwarded packets"
echo ""
sudo iptables -L FORWARD -n -v --line-numbers

echo ""
echo "5. Rule Counters Analysis"
echo "========================"
echo ""

# Check for port forward rules with zero packet counts
echo "Checking for unused rules (zero packet count):"
ZERO_RULES=$(sudo iptables -t nat -L -n -v | grep " 0 " | grep -E "(8888|DNAT|SNAT|MASQUERADE)" | wc -l)
if [ "$ZERO_RULES" -gt 0 ]; then
    echo "  ⚠️  Found $ZERO_RULES port forward rules with zero packets"
    echo "  These rules exist but no traffic is matching them:"
    sudo iptables -t nat -L -n -v | grep " 0 " | grep -E "(8888|DNAT|SNAT|MASQUERADE)"
else
    echo "  ✓ All port forward rules have matched traffic"
fi

echo ""
echo "6. Connection Tracking"
echo "====================="
echo ""

if command -v conntrack &> /dev/null; then
    echo "Current connection tracking entries:"
    sudo conntrack -L 2>/dev/null | grep -E "(8888|8080)" || echo "  (no tracked connections for ports 8888/8080)"

    echo ""
    echo "Connection tracking statistics:"
    sudo conntrack -S 2>/dev/null | head -5
else
    echo "  conntrack tool not installed"
    echo "  Install: sudo apt-get install conntrack"
fi

echo ""
echo "7. Potential Issues"
echo "=================="
echo ""

# Check for rule order problems
OUTPUT_LINE=$(sudo iptables -t nat -L OUTPUT -n --line-numbers | grep "DNAT.*8888" | awk '{print $1}' | head -1)
if [ -n "$OUTPUT_LINE" ]; then
    echo "  OUTPUT DNAT rule is at line $OUTPUT_LINE"
    if [ "$OUTPUT_LINE" -gt 5 ]; then
        echo "  ⚠️  Rule is far down the chain, earlier rules might be catching traffic"
    fi
fi

# Check for conflicting rules
REDIRECT_COUNT=$(sudo iptables -t nat -L OUTPUT -n | grep "REDIRECT" | wc -l)
if [ "$REDIRECT_COUNT" -gt 0 ]; then
    echo "  ⚠️  Found REDIRECT rules in OUTPUT chain that might conflict"
    sudo iptables -t nat -L OUTPUT -n | grep "REDIRECT"
fi

# Check for SNAT before MASQUERADE
POSTROUTING_RULES=$(sudo iptables -t nat -L POSTROUTING -n --line-numbers)
SNAT_LINE=$(echo "$POSTROUTING_RULES" | grep "SNAT" | head -1 | awk '{print $1}')
MASQ_LINE=$(echo "$POSTROUTING_RULES" | grep "MASQUERADE" | head -1 | awk '{print $1}')

if [ -n "$SNAT_LINE" ] && [ -n "$MASQ_LINE" ]; then
    echo "  SNAT rule at line $SNAT_LINE"
    echo "  MASQUERADE rule at line $MASQ_LINE"
    if [ "$SNAT_LINE" -lt "$MASQ_LINE" ]; then
        echo "  ✓ SNAT before MASQUERADE (correct order)"
    else
        echo "  ⚠️  MASQUERADE before SNAT (might cause issues)"
    fi
fi

echo ""
echo "8. System Parameters"
echo "==================="
echo ""
echo "  IP forwarding: $(sysctl -n net.ipv4.ip_forward)"
[ "$(sysctl -n net.ipv4.ip_forward)" = "0" ] && echo "    ❌ DISABLED - This will break forwarding!"

echo "  RP filter (all): $(sysctl -n net.ipv4.conf.all.rp_filter)"
echo "  RP filter (default): $(sysctl -n net.ipv4.conf.default.rp_filter)"
echo "  RP filter (lo): $(sysctl -n net.ipv4.conf.lo.rp_filter)"
echo "  RP filter (nanofuse0): $(sysctl -n net.ipv4.conf.nanofuse0.rp_filter 2>/dev/null || echo 'N/A')"

if [ "$(sysctl -n net.ipv4.conf.all.rp_filter)" = "1" ]; then
    echo "    ⚠️  Strict RP filter might drop packets with modified source/dest"
fi

echo ""
echo "=== Recommendations ==="
echo ""

if [ "$(sysctl -n net.ipv4.ip_forward)" = "0" ]; then
    echo "❌ CRITICAL: Enable IP forwarding:"
    echo "   sudo sysctl -w net.ipv4.ip_forward=1"
    echo ""
fi

if [ "$ZERO_RULES" -gt 2 ]; then
    echo "⚠️  Many rules with zero packets - traffic not matching"
    echo "   Check rule order and matching criteria"
    echo ""
fi

if [ "$(sysctl -n net.ipv4.conf.all.rp_filter)" = "1" ]; then
    echo "⚠️  Consider disabling RP filter for testing:"
    echo "   sudo sysctl -w net.ipv4.conf.all.rp_filter=0"
    echo "   sudo sysctl -w net.ipv4.conf.nanofuse0.rp_filter=0"
    echo ""
fi

echo "To reset all iptables rules and start fresh:"
echo "  sudo iptables -t nat -F"
echo "  sudo iptables -F"
echo "  sudo conntrack -F"
echo ""
