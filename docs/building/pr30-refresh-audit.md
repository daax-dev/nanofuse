# PR30 Replacement Audit

Closed PR #30 retained Copilot inline comments, but the comments inspected before this replacement pass were attached to stale commit `f21a35369709026bf9bb21d9ee6aeebaa61db680` and reported as outdated review threads. No current-head Copilot comments were found against branch head `19ed505e71ae6692fc46ae1a666a0d972170c12c`.

The stale comments target the removed trusted-header policy engine, Aembit-style policy claims, and SVID rotation code. Those implementations remain removed rather than patched in place.

## Implemented Auth Slice

The replacement branch keeps the narrowed TCP mTLS scope:

- TCP listeners use `tls.RequireAndVerifyClientCert` when `auth.enabled` and `api.tcp_bind` are configured.
- Request identity is extracted only from a verified client certificate SPIFFE URI SAN.
- `X-SPIFFE-ID` and other client-controlled identity headers are ignored.
- Unix socket listeners remain local/plain and rely on filesystem permissions.
- Aembit-style policy enforcement and SVID rotation are not implemented by this PR.

## Replacement Hardening

This pass adds explicit tests that reject a spoofed `X-SPIFFE-ID` header without TLS and prove a spoofed header does not override a verified certificate identity. It also adds config validation tests proving TCP auth fails closed when server cert, server key, or client CA paths are missing.
