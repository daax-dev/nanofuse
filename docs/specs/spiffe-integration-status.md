# SPIFFE/SPIRE Integration Status - nanofuse

**Status**: Host-side implementation complete ✅
**Date**: 2026-01-07
**Branch**: jan7

## Completed Work

### Host-Side Components ✅

1. **SPIRE Service** (`internal/spire/service.go`)
   - Workload entry registration/unregistration
   - D025 SPIFFE ID generation
   - Identity parameter validation

2. **Vsock Proxy** (`internal/firecracker/vsock_proxy.go`)
   - Bidirectional proxy: Firecracker vsock ↔ SPIRE agent socket
   - Automatic lifecycle management with VM

3. **API Integration**
   - VM create/delete handlers integrate with SPIRE service
   - SPIFFE fields exposed in OpenAPI spec

## Remaining Work

### Guest-Side SVID Client 🟡 (Medium Priority)

VM images need a client to connect to the vsock proxy and acquire SVIDs.

**Recommended Approach:**
```bash
# In VM guest, create local socket that proxies to host via vsock
socat UNIX-LISTEN:/run/spire/sockets/agent.sock,fork \
  VSOCK-CONNECT:2:8307 &

# Then use standard go-spiffe library
```

**Alternative:** Build a minimal Go binary with vsock transport support.

## Integration Testing

Once guest-side client is available, test with:
```bash
# 1. Create VM with SPIFFE identity
curl -X POST http://localhost:8080/vms \
  -d '{"owner_user_id": "jpoley", "group_id": "engineering", "auto_register_spiffe": true}'

# 2. Inside VM, verify SVID acquisition
./spiffe-client fetch-jwt --audience http://localhost:1411

# 3. Exchange for access token
curl -X POST https://auth.poley.dev/api/oidc/token \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "subject_token=${JWT_SVID}" \
  -d "subject_token_type=urn:ietf:params:oauth:token-type:jwt"
```

## Related Documents

- [auth.poley.dev/docs/specs/task-025-cross-project-requirements.md](../../../jp/auth.poley.dev/docs/specs/task-025-cross-project-requirements.md)
