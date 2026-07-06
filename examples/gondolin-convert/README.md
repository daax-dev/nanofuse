# gondolin -> nanofuse conversion example

This example shows `nanofuse convert gondolin` in action and the migration path
from a gondolin-style agent sandbox to a nanofuse spec.

## Reframe: gondolin has no config file

The gondolin project (`earendil-works/gondolin`) has **no declarative sandbox
config file**. Its sandbox is defined imperatively through `gondolin bash|exec`
CLI flags (see gondolin `docs/cli.md`) and a TypeScript `VM.create()` API. There
is therefore nothing to "parse" from gondolin itself.

`nanofuse convert gondolin` instead reads a **nanofuse-authored mirror YAML**
([`sandbox.yaml`](./sandbox.yaml)) whose keys map 1:1 to gondolin's flag surface,
and emits:

1. a nanofuse spec ([`nanofuse-spec.yaml`](./nanofuse-spec.yaml)), and
2. an explicit **divergence report** naming every gondolin feature that has no
   faithful nanofuse equivalent.

The honest value is the divergence report, not a false claim of equivalence.

## Run it

```bash
# Fails closed: unrepresentable gondolin features are reported and the command
# exits non-zero rather than silently dropping them.
nanofuse convert gondolin sandbox.yaml

# Drop unrepresentable features (loudly) and write the spec.
nanofuse convert gondolin sandbox.yaml --allow-lossy -o nanofuse-spec.yaml
```

## Mapping summary

| gondolin flag        | nanofuse handling                                                        |
|----------------------|--------------------------------------------------------------------------|
| `--image`            | Clean: `image`.                                                          |
| `resources` (hint)   | `vcpus` / `memory_mib`. Gondolin has no CPU/memory model; if absent, nanofuse defaults are assumed and disclosed. |
| `--allow-host` (L7)  | Safe degrade: locked-down default-drop egress + warning. An HTTP host allowlist cannot be an L3/L4 CIDR policy. Opt in to `--resolve-egress` for point-in-time hostname->/32 resolution of literal hosts. |
| `--vmm`              | No equivalent: gondolin runs qemu/krun; nanofuse runs only firecracker.  |
| `--cwd`              | No equivalent: no guest working-directory field.                         |
| `--env`              | No equivalent: no guest environment injection.                           |
| `--host-secret`      | No equivalent: no host-secret injection.                                 |
| `--mount-hostfs`     | No equivalent: no host-directory bind mounts.                            |
| `--mount-memfs`      | No equivalent: no in-memory mount primitive.                             |
| `--ssh-allow-host`   | No equivalent: no SSH egress broker.                                     |
| `--tcp-map`          | No equivalent: CIDR-based egress, no guest->upstream TCP remapping.       |
| `--dns`              | Coarse only: gondolin resolver modes collapse to a single AllowDNS toggle.|
| `--rootfs-size`      | No equivalent: no rootfs-size control in the create request.             |

Features marked "No equivalent" are **blocking**: the conversion fails closed
unless `--allow-lossy` is passed, which drops them loudly and proceeds. This is
deliberate for a security sandbox: refusing is safer than silently dropping an
isolation-relevant control.
