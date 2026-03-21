#!/bin/bash
# VM Port Forwarding Script for NanoFuse
# Creates iptables DNAT rules to expose VM ports to the host

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Constants
BRIDGE_NAME="nanofuse0"
HOST_IP="0.0.0.0" # Listen on all interfaces by default

# Usage information
usage() {
    cat <<EOF
Usage: $0 <command> [options]

Commands:
    add <vm-ip> <vm-port> <host-port>    Add port forward rule
    remove <vm-ip> <vm-port> <host-port> Remove port forward rule
    list                                  List all active port forwards
    flush                                 Remove all NanoFuse port forwards
    help                                  Show this help message

Examples:
    # Forward host:8080 to VM's port 8080
    sudo $0 add 172.16.0.10 8080 8080

    # Forward host:8443 to VM's port 443
    sudo $0 add 172.16.0.10 443 8443

    # Remove a port forward
    sudo $0 remove 172.16.0.10 8080 8080

    # List all active forwards
    sudo $0 list

    # Remove all port forwards
    sudo $0 flush

Notes:
    - This script requires root privileges
    - VM must be running and have the specified port listening
    - Host port must not be in use by another service
EOF
    exit 1
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}ERROR: This script must be run as root${NC}"
        echo "Please run: sudo $0 $@"
        exit 1
    fi
}

# Validate IP address
validate_ip() {
    local ip=$1
    if [[ ! $ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        echo -e "${RED}ERROR: Invalid IP address: $ip${NC}"
        exit 1
    fi
}

# Validate port number
validate_port() {
    local port=$1
    if [[ ! $port =~ ^[0-9]+$ ]] || [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
        echo -e "${RED}ERROR: Invalid port number: $port${NC}"
        echo "Port must be between 1 and 65535"
        exit 1
    fi
}

# Check if port is already in use on host
check_host_port() {
    local port=$1
    if lsof -i ":$port" >/dev/null 2>&1; then
        echo -e "${YELLOW}WARNING: Port $port is already in use on host${NC}"
        lsof -i ":$port"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Add DNAT rule for port forwarding
add_port_forward() {
    local vm_ip=$1
    local vm_port=$2
    local host_port=$3

    echo -e "${BLUE}Adding port forward: host:$host_port -> $vm_ip:$vm_port${NC}"

    # Validate inputs
    validate_ip "$vm_ip"
    validate_port "$vm_port"
    validate_port "$host_port"
    check_host_port "$host_port"

    # Check if rule already exists
    if iptables -t nat -C PREROUTING -p tcp --dport "$host_port" \
        -j DNAT --to-destination "$vm_ip:$vm_port" 2>/dev/null; then
        echo -e "${YELLOW}Port forward already exists${NC}"
        return 0
    fi

    # Add DNAT rule (PREROUTING chain)
    # This redirects incoming packets to host_port to vm_ip:vm_port
    iptables -t nat -A PREROUTING -p tcp --dport "$host_port" \
        -j DNAT --to-destination "$vm_ip:$vm_port"

    # Add DNAT rule for localhost connections (OUTPUT chain)
    # This allows connections from the host itself to work
    iptables -t nat -A OUTPUT -p tcp --dport "$host_port" \
        -j DNAT --to-destination "$vm_ip:$vm_port"

    # Ensure FORWARD chain allows the traffic
    # (This might already exist from NAT setup, but adding it is safe)
    if ! iptables -C FORWARD -p tcp -d "$vm_ip" --dport "$vm_port" \
        -j ACCEPT 2>/dev/null; then
        iptables -A FORWARD -p tcp -d "$vm_ip" --dport "$vm_port" \
            -j ACCEPT
    fi

    echo -e "${GREEN}✓ Port forward added successfully${NC}"
    echo
    echo "Test with:"
    echo "  curl http://localhost:$host_port"
    echo "  or from another machine: curl http://<host-ip>:$host_port"
}

# Remove DNAT rule
remove_port_forward() {
    local vm_ip=$1
    local vm_port=$2
    local host_port=$3

    echo -e "${BLUE}Removing port forward: host:$host_port -> $vm_ip:$vm_port${NC}"

    # Validate inputs
    validate_ip "$vm_ip"
    validate_port "$vm_port"
    validate_port "$host_port"

    # Remove PREROUTING rule
    if iptables -t nat -C PREROUTING -p tcp --dport "$host_port" \
        -j DNAT --to-destination "$vm_ip:$vm_port" 2>/dev/null; then
        iptables -t nat -D PREROUTING -p tcp --dport "$host_port" \
            -j DNAT --to-destination "$vm_ip:$vm_port"
        echo -e "${GREEN}✓ Removed PREROUTING rule${NC}"
    else
        echo -e "${YELLOW}PREROUTING rule not found${NC}"
    fi

    # Remove OUTPUT rule
    if iptables -t nat -C OUTPUT -p tcp --dport "$host_port" \
        -j DNAT --to-destination "$vm_ip:$vm_port" 2>/dev/null; then
        iptables -t nat -D OUTPUT -p tcp --dport "$host_port" \
            -j DNAT --to-destination "$vm_ip:$vm_port"
        echo -e "${GREEN}✓ Removed OUTPUT rule${NC}"
    else
        echo -e "${YELLOW}OUTPUT rule not found${NC}"
    fi

    # Note: We don't remove FORWARD rules as they might be used by other port forwards
    # or by the general NAT setup

    echo -e "${GREEN}✓ Port forward removed${NC}"
}

# List all active port forwards
list_port_forwards() {
    echo -e "${BLUE}Active Port Forwards (PREROUTING):${NC}"
    echo
    iptables -t nat -L PREROUTING -n -v --line-numbers | grep DNAT || echo "No port forwards found"

    echo
    echo -e "${BLUE}Active Port Forwards (OUTPUT for localhost):${NC}"
    echo
    iptables -t nat -L OUTPUT -n -v --line-numbers | grep DNAT || echo "No port forwards found"
}

# Flush all DNAT rules
flush_port_forwards() {
    echo -e "${YELLOW}WARNING: This will remove ALL DNAT rules in PREROUTING and OUTPUT chains${NC}"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled"
        exit 0
    fi

    echo -e "${BLUE}Flushing DNAT rules...${NC}"

    # Remove all DNAT rules from PREROUTING
    iptables -t nat -S PREROUTING | grep DNAT | cut -d " " -f 2- | while read -r rule; do
        iptables -t nat -D $rule
        echo "  Removed: $rule"
    done

    # Remove all DNAT rules from OUTPUT
    iptables -t nat -S OUTPUT | grep DNAT | cut -d " " -f 2- | while read -r rule; do
        iptables -t nat -D $rule
        echo "  Removed: $rule"
    done

    echo -e "${GREEN}✓ All DNAT rules flushed${NC}"
}

# Main script logic
main() {
    local command=$1

    case "$command" in
        add)
            check_root
            if [ $# -ne 4 ]; then
                echo -e "${RED}ERROR: Invalid arguments${NC}"
                echo "Usage: $0 add <vm-ip> <vm-port> <host-port>"
                exit 1
            fi
            add_port_forward "$2" "$3" "$4"
            ;;
        remove)
            check_root
            if [ $# -ne 4 ]; then
                echo -e "${RED}ERROR: Invalid arguments${NC}"
                echo "Usage: $0 remove <vm-ip> <vm-port> <host-port>"
                exit 1
            fi
            remove_port_forward "$2" "$3" "$4"
            ;;
        list)
            check_root
            list_port_forwards
            ;;
        flush)
            check_root
            flush_port_forwards
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            echo -e "${RED}ERROR: Unknown command: $command${NC}"
            echo
            usage
            ;;
    esac
}

# Run main function
main "$@"
