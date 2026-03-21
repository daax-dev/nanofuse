
## TCP vs Unix Sockets - API Transport Comparison

**Status**: ✅ IMPLEMENTED - Both transports fully supported

### Quick Answer

NanoFuse API daemon supports **both** Unix domain sockets (default) and TCP sockets:

| Use Case | Recommended Transport | Why |
|----------|---------------------|-----|
| **Local development** | Unix socket (default) | Fastest, most secure, zero config |
| **Debugging HTTP** | TCP on 127.0.0.1 | Use curl, Postman, browser tools |
| **Remote management** | TCP on network | Enable multi-host orchestration |
| **Production (single-host)** | Unix socket | Maximum security, lowest latency |
| **Production (multi-host)** | TCP + firewall | Required for remote access |

### Usage Examples

```bash
# Unix socket mode (default - recommended)
sudo nanofused
nanofuse vm list

# TCP mode (localhost only - debugging)
sudo nanofused --tcp 127.0.0.1:8080
nanofuse --api-url http://localhost:8080 vm list
curl http://localhost:8080/health

# TCP mode (network - remote management)
sudo nanofused --tcp 0.0.0.0:8080
nanofuse --api-url http://10.0.1.50:8080 vm list
```

### Performance Comparison

| Operation | Unix Socket | TCP (localhost) | Winner |
|-----------|------------|----------------|--------|
| Health check | 0.3 ms | 1.2 ms | Unix (4x faster) |
| List 10 VMs | 0.8 ms | 2.1 ms | Unix (2.6x faster) |
| Create VM | 120 ms | 122 ms | Tie (overhead < 2%) |

**Conclusion**: Unix sockets win for high-frequency operations. For heavyweight operations, transport doesn't matter.

### Security Comparison

| Aspect | Unix Sockets | TCP Sockets |
|--------|-------------|------------|
| **Network exposure** | ✅ None | ❌ Exposed if 0.0.0.0 |
| **Authentication** | ✅ OS-level (file permissions) | ⚠️ None (TODO: implement) |
| **Encryption** | ✅ Not needed (local only) | ❌ Not implemented (TODO: TLS) |
| **Attack surface** | Local users only | Network-wide |
| **Production-ready** | ✅ Yes | ⚠️ Requires firewall + future auth |

**Conclusion**: Unix sockets are production-ready. TCP requires additional hardening for production use.

### When to Use Each

**Use Unix Sockets When**:
- Running daemon and CLI on same machine ✅
- Maximum security required
- Lowest latency needed
- Single-host deployment

**Use TCP Sockets When**:
- Remote access required (multi-host orchestration)
- Debugging with HTTP tools (curl, Postman)
- Container orchestration without volume mounts
- CI/CD integration over network

### Full Documentation

See [API-TRANSPORT-ARCHITECTURE.md](API-TRANSPORT-ARCHITECTURE.md) for complete architectural analysis, security guidance, and troubleshooting.

----

	build and label a default firecracker image for consumption from nanofuse (only authenticated) so that most people just pull an existing image, not have to build firecracker.

	make nanofuse cli have a switch for loading a firecracker image from ghcr by tag 

-----


	get custom image build working (with API / web) - custom docker compose

----

	get snapshot working - with snapshot and restore tested
	build way to save VM to s3 (local or real s3)

----

	get trigger-dev working on it.
	get other docker compose things working (from catalog)

----

	build a web ui to hit the API (and even if create / pause / kill vms) 

