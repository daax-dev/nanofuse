# SPIFFE SVID Issuance for microVMs

nanofuse gives every microVM a short-lived, cryptographically-verifiable
identity (a SPIFFE X.509-SVID) instead of static, at-rest credentials. SVIDs
auto-rotate with sub-hour TTLs, so a leaked credential is useless within
minutes. This closes the static-credential window demonstrated by the LiteLLM
credential-vault exploit (2026-04-07).

This document covers the issuance lifecycle, how to deploy a SPIRE agent
alongside a daax cluster, and how to bind workload identities to policies.

## Architecture

```
                         host                         │      guest microVM
                                                       │
  SPIRE server ──registers──> SPIRE agent ──Workload──>│ vsock proxy ─> SVID Manager
   (entry: spiffe ID,          (issues SVIDs           │ (CID:port)      (this package)
    selectors, TTL)             over Workload API)     │                    │
                                                       │                    ▼
                                                       │   /var/run/secrets/spiffe/svid.json (0400)
```

Two layers, two repositories of responsibility:

1. **Workload registration** (`internal/spire/service.go`, pre-existing): the
   host registers a SPIRE *entry* for each VM — the D025 SPIFFE ID
   (`spiffe://<trust-domain>/g/<group>/u/<user>/w/microvm/<vm-id>`), parent ID,
   selectors, and TTL.
2. **SVID issuance** (`internal/spire/{svid,issuer,manager}.go`, this work): the
   workload acquires the actual SVID for its registered identity, persists it as
   a `0400` document, and rotates it before expiry.

### The SVID document

The Manager writes the issued SVID to `/var/run/secrets/spiffe/svid.json`, mode
`0400` (owner read-only). The schema mirrors a SPIFFE Workload API X509-SVID
response so a `go-spiffe`-based consumer can read it without translation:

```json
{
  "spiffe_id": "spiffe://poley.dev/g/engineering/u/jpoley/w/microvm/vm-abc123",
  "x509_svid": "<PEM leaf + intermediates>",
  "x509_svid_key": "<PEM PKCS#8 private key>",
  "bundle": "<PEM trust bundle (CA roots)>",
  "issued_at": "2026-06-28T12:00:00Z",
  "expires_at": "2026-06-28T13:00:00Z"
}
```

### Lifecycle and guarantees

- **Acquire before use.** `Manager.Start` obtains the initial SVID before any
  consumer reads the mount.
- **Fail-safe, never fail-open.** If the SPIRE agent is unreachable at startup,
  `Start` returns an error naming SPIRE unreachability and starts no background
  work. The workload must treat this as fatal — there is no fallback to
  plaintext or static credentials.
- **Rotate before expiry.** SVIDs default to a 60-minute TTL and are refreshed
  15 minutes before expiry, so a valid SVID is always available with overlap. A
  failed rotation retains the still-valid current SVID and retries; it never
  deletes a usable identity.
- **Atomic, restrictive writes.** Each issuance writes a temp file
  (`chmod 0400`) and renames it over the destination, so readers see an atomic
  swap and the credential is never world/group-readable. On unix the write is
  directory-fd-anchored (`openat`/`renameat` with `O_NOFOLLOW`, plus an `fstat`
  ownership/mode check on the directory fd), closing the TOCTOU window where the
  parent directory could be swapped or redirected via symlink between validation
  and write.

### The `Source` seam (production vs. dev)

`Source` is the injection point for SVID acquisition:

- **Production:** a `go-spiffe/v2` `workloadapi.X509Source` dialed over the
  existing Firecracker vsock proxy (`internal/firecracker/vsock_proxy.go`) from
  inside the guest. This requires a live SPIRE agent and is wired at runtime
  (see *Deferred to runtime* below).
- **Development / tests:** `LocalCASource`, an in-process CA that mints real,
  independently-verifiable X.509-SVIDs. Used to validate the lifecycle and
  cryptography without a SPIRE deployment.

## Deploying a SPIRE agent alongside a daax cluster

> Versions: use a currently-supported SPIRE release and a supported base image
> per the repo's no-stale-images policy.

### 1. Run the SPIRE server (trust domain authority)

```bash
docker run -d --name spire-server \
  -v "$PWD/spire-server.conf:/opt/spire/conf/server/server.conf:ro" \
  ghcr.io/spiffe/spire-server:<supported-tag> \
  -config /opt/spire/conf/server/server.conf
```

`server.conf` sets the trust domain (e.g. `poley.dev`) — it must match
`spire.trust_domain` in nanofuse config.

### 2. Run the SPIRE agent on each daax host

The agent attests to the server and exposes the Workload API socket that the
vsock proxy bridges into guests:

```bash
docker run -d --name spire-agent \
  -v /run/spire/sockets:/run/spire/sockets \
  ghcr.io/spiffe/spire-agent:<supported-tag> \
  -config /opt/spire/conf/agent/agent.conf
```

Point nanofuse at it:

```yaml
# config.yaml
spire:
  enabled: true
  trust_domain: poley.dev
  parent_id: spiffe://poley.dev/spire/agent/...   # agent's SPIFFE ID
  workload_type: microvm
  default_ttl: 3600          # 60 minutes
  container_name: spire-server
  vsock_cid: 3               # >=3 enables the guest<->agent vsock proxy
  vsock_port: 8307
  agent_socket: /run/spire/sockets/agent.sock
```

### 3. Guest-side acquisition

Inside the microVM, bridge the vsock to a local Workload API socket and run the
SVID Manager (or a `go-spiffe` client) against it:

```bash
# Guest: expose the host SPIRE agent socket locally via vsock.
socat UNIX-LISTEN:/run/spire/sockets/agent.sock,fork \
  VSOCK-CONNECT:2:8307 &
```

## Binding daax workload identity to policies

The SPIFFE ID is the subject; authorization happens at the relying party.

1. **Register the entry** (nanofuse does this automatically on VM create when
   `auto_register_spiffe` is set):

   ```bash
   spire-server entry create \
     -spiffeID spiffe://poley.dev/g/engineering/u/jpoley/w/microvm/vm-abc123 \
     -parentID spiffe://poley.dev/spire/agent/... \
     -selector docker:label:vm_id:vm-abc123 \
     -ttl 3600
   ```

2. **Bind to access policy.** Exchange the SVID for a scoped access token, then
   authorize on the SPIFFE ID:

   ```bash
   curl -X POST https://auth.poley.dev/api/oidc/token \
     -d grant_type=urn:ietf:params:oauth:grant-type:token-exchange \
     -d subject_token_type=urn:ietf:params:oauth:token-type:jwt \
     -d subject_token="${JWT_SVID}"
   ```

   Policies match on the SPIFFE ID path segments (`/g/<group>`, `/u/<user>`),
   so a policy can grant a group, a user, or a single VM. Because the SVID is
   verified via mTLS / signature against the trust bundle, the identity cannot
   be spoofed by a header.

## Deferred to runtime (honest scope)

This change ships the portable, testable issuance core. The following require a
live SPIRE agent / Firecracker runtime and are **not** exercised on a dev box:

- The production `go-spiffe` Workload API `Source` over the vsock proxy.
- Mounting `svid.json` into the guest filesystem and aborting VM start when
  issuance fails (the in-process Manager surfaces the fatal error; the VM
  lifecycle integration enforces the abort).
- The literal 45-minute rotation soak test. Rotation correctness
  (rotate-before-expiry, overlap, fresh verifiable cert, fail-safe, retry) is
  validated deterministically with an injectable clock in
  `internal/spire/manager_test.go`.

## See also

- `docs/specs/spiffe-integration-status.md` — host-side registration + vsock
  proxy status.
- `internal/spire/service.go` — workload entry registration.
