# Building an eBPF Load Balancer from Scratch

> A demonstration by Liz Rice showing how to build a simple load balancer using eBPF and XDP in just a few lines of C code

## Overview

This talk demonstrates building a functional HTTP load balancer using eBPF (extended Berkeley Packet Filter) and XDP (eXpress Data Path) to manipulate network packets at the kernel level, achieving high-performance packet processing.

## Architecture

### Components
- **Client**: Docker container sending HTTP requests
- **Load Balancer**: Container with eBPF program attached
- **Backend A & B**: Two backend servers responding to requests
- **Network**: All containers on same host with Docker networking

### Network Setup
- Docker assigns sequential IP addresses (172.17.0.x)
- MAC addresses mirror IP addresses in last bytes
- Virtual ethernet (veth) interfaces for each container
- Simplifies address manipulation in demo

## XDP (eXpress Data Path)

### Key Features
- Processes packets at earliest possible point
- Before any network stack processing
- Can inspect, modify, and redirect packets
- Return codes control packet fate

### XDP Return Codes
- `XDP_PASS`: Continue normal processing up the stack
- `XDP_TX`: Send packet back out the same interface
- `XDP_DROP`: Discard the packet
- `XDP_REDIRECT`: Send to different interface

## Network Packet Structure

```
[Ethernet Header] → [IP Header] → [TCP Header] → [Payload]
```

### Important Considerations
- Ethernet header contains MAC addresses and protocol type
- IP header contains source/destination IPs and checksum
- Checksums must be recalculated when headers modified
- Packet boundaries must be validated

## Implementation Steps

### 1. Basic XDP Program Structure

```c
SEC("xdp")
int xdp_load_balancer(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    // Packet processing logic here

    return XDP_PASS;
}
```

### 2. Parse Ethernet Header

```c
struct ethhdr *eth = data;

// Bounds checking - critical for verifier
if ((void*)(eth + 1) > data_end)
    return XDP_PASS;

// Check if IP packet
if (bpf_ntohs(eth->h_proto) != ETH_P_IP)
    return XDP_PASS;
```

### 3. Parse IP Header

```c
struct iphdr *ip = (struct iphdr *)(eth + 1);

// Bounds checking
if ((void*)(ip + 1) > data_end)
    return XDP_PASS;

// Check if TCP
if (ip->protocol != IPPROTO_TCP)
    return XDP_PASS;
```

### 4. Implement Load Balancing Logic

```c
// Determine direction based on source
if (ip->saddr == CLIENT_IP) {
    // Client → Backend (forward)

    // Choose backend (simple round-robin using time)
    __u64 now = bpf_ktime_get_ns();
    __u32 backend = (now % 2) ? BACKEND_A : BACKEND_B;

    // Rewrite destination
    ip->daddr = backend;
    eth->h_dest[5] = backend & 0xFF;

    // Rewrite source to be load balancer
    ip->saddr = LB_IP;
    eth->h_source[5] = LB_IP & 0xFF;

} else {
    // Backend → Client (return)

    // Rewrite destination to client
    ip->daddr = CLIENT_IP;
    eth->h_dest[5] = CLIENT_IP & 0xFF;

    // Rewrite source to be load balancer
    ip->saddr = LB_IP;
    eth->h_source[5] = LB_IP & 0xFF;
}
```

### 5. Update Checksums

```c
// Recalculate IP header checksum
update_ip_checksum(ip);

// Return TX to send packet back out
return XDP_TX;
```

## Key Implementation Details

### Address Translation

The load balancer performs double NAT:
1. **Client → LB → Backend**:
   - Source: Client → LB
   - Destination: LB → Backend

2. **Backend → LB → Client**:
   - Source: Backend → LB
   - Destination: LB → Client

### Why Double NAT?

- Backend must see request coming from LB (not client)
- Client must see response coming from LB (not backend)
- Ensures proper TCP connection tracking
- Maintains session affinity requirements

### Load Balancing Algorithm

Simple time-based selection:
```c
__u64 now = bpf_ktime_get_ns();
if (now % 2 == 0) {
    backend = BACKEND_A;
} else {
    backend = BACKEND_B;
}
```

## Building and Loading

### Compilation
```bash
clang -O2 -target bpf -c lb.c -o lb.o
```

### Loading with bpftool
```bash
# Detach any existing program
bpftool net detach xdp dev eth0

# Load and attach new program
bpftool prog load lb.o /sys/fs/bpf/lb
bpftool net attach xdp name lb dev eth0
```

## Production Considerations

### Current Limitations
- Hard-coded addresses
- No connection tracking
- Random backend selection
- Single client support
- No health checking

### Production Improvements Needed

#### 1. Dynamic Configuration
- Store backend IPs in eBPF maps
- Update from userspace control plane
- No recompilation needed

#### 2. Connection Tracking
- Map client IP:port to assigned backend
- Ensure session persistence
- Handle connection state

#### 3. Multiple Client Support
- Port translation/multiplexing
- Connection table management
- NAT state tracking

#### 4. Advanced Load Balancing
- Weighted round-robin
- Least connections
- Response time based
- Geographic/topology aware

#### 5. Health Checking
- Backend availability monitoring
- Automatic failover
- Graceful degradation

#### 6. Observability
- Connection metrics
- Latency tracking
- Error rates
- Packet drops

## Performance Benefits

### XDP Advantages
- Process packets before allocation of sk_buff
- Bypass entire network stack when forwarding
- Zero-copy packet modification
- Line-rate packet processing
- Minimal CPU overhead

### Benchmark Results
- Near wire-speed forwarding
- Millions of packets per second per core
- Microsecond-level latency
- Linear scaling with cores

## Use Cases

### When to Use XDP Load Balancing
- High-performance requirements
- Low latency critical
- DDoS mitigation
- Simple L3/L4 load balancing
- Edge/CDN applications

### When NOT to Use
- Complex L7 logic needed
- TLS termination required
- Advanced session persistence
- Limited kernel support

## Security Considerations

- XDP programs run with high privileges
- Careful bounds checking required
- Potential for packet loops
- Resource exhaustion risks
- Need proper access controls

## Debugging and Troubleshooting

### Tools
- `bpftool`: Program management
- `tcpdump`: Packet capture
- `bpf_trace_printk()`: Kernel tracing
- `/sys/kernel/debug/tracing/trace_pipe`: Trace output

### Common Issues
- Verifier rejections (bounds checks)
- Incorrect byte ordering
- Checksum calculation errors
- Packet structure assumptions

## Advanced Topics

### eBPF Maps for State
```c
struct bpf_map_def SEC("maps") backends = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32),
    .value_size = sizeof(struct backend),
    .max_entries = 10,
};
```

### Integration with Container Orchestration
- Kubernetes service discovery
- Automatic backend updates
- Health check integration
- Metrics exposition

## Conclusion

This demonstration shows that with just ~100 lines of C code, we can build a functional L3/L4 load balancer using eBPF/XDP that:
- Forwards packets between clients and backends
- Performs address translation
- Implements basic load distribution
- Operates at kernel speed

While not production-ready, it illustrates the power of eBPF for building high-performance networking tools and the fundamental concepts behind modern load balancers like Cilium and Katran.

## Resources

- GitHub: Code examples and full implementation
- "Learning eBPF" book for deeper understanding
- XDP documentation and tutorials
- Cilium project for production eBPF networking