# Sandbox Objective Validation

**Date:** 2026-05-30
**Backlog:** `TASK-47`
**Spec:** `.flowspec/features/codex-goal/spec.md`

## Scope

This validation covers the `objective.md` sandbox requirements on the current branch:

- Linux/KVM microVM execution path.
- macOS and Windows operator paths through the `nanofused` API on a Linux/KVM execution environment.
- OCI/container-to-rootfs wrapping path.
- Per-VM persistent filesystem state.
- Short-running and long-running lifecycle behavior.
- SPIFFE/vsock identity posture with raw secrets kept out of VM config.
- Default-deny/proxy-only egress policy for future LLM/API/MCP interception.
- Vagrant/hypervisor closed-loop testing.

## Current Host Finding

The current host is macOS arm64 and has no `/dev/kvm`. Firecracker cannot run natively on this host. Vagrant is installed with the Parallels provider. Closed-loop validation must therefore either:

1. Run in a Vagrant guest/provider that exposes Linux KVM, or
2. Fail before VM boot with an explicit `/dev/kvm` capability error.

The runnable API requirement is satisfied by `nanofused` on the Linux/KVM runtime host. macOS and Windows hosts run as clients using `NANOFUSE_API_URL`, `--api-url`, curl, PowerShell, or the planned tray app. See `docs/API_QUICK_START.md` and `docs/MAC_WINDOWS_CLIENTS.md`.

## Validation Commands

Local code gates:

```bash
go test ./internal/network ./internal/api ./internal/client
go fmt ./...
mage ci
```

Local result on 2026-05-30:

- `mage ci` passed on the macOS arm64 host.
- `go test ./...` passed on the macOS arm64 host.
- YAML/OpenAPI parsing and JSONL parsing passed.
- Shell syntax, Ruby syntax, and `vagrant validate` passed.
- `gosec` was not installed; the existing `mage ci` target reports that condition and continues.

Vagrant closed-loop:

```bash
cd dev/vagrant
./closed-loop.sh
```

The Vagrant script performs:

- `vagrant up` or `vagrant provision`.
- Guest `/dev/kvm` read/write preflight.
- `sudo mage ci` inside `/nanofuse`.
- `nanofused` restart and health check.
- Firecracker/image verification through `verify.sh`.

Local Vagrant result on 2026-05-30:

- Provider: `vagrant-parallels` with `bento/ubuntu-24.04` arm64.
- Guest boot and rsync succeeded.
- Updated Vagrant port forwarding was applied: guest API `8080` forwarded to host `127.0.0.1:18080`.
- Provisioning failed at the required KVM preflight: `/dev/kvm not found. Nested KVM required`.
- A second attempt enabling Parallels nested virtualization reached `prlctl start` and failed before guest boot with `Unable to start the virtual machine`.
- Guest-side `mage ci`, daemon health/capabilities, and Firecracker boot verification did not run because Firecracker cannot execute without guest KVM.
- The Parallels VM was halted after the latest KVM-unavailable run.

The full host run outputs are stored at `.logs/validation/vagrant-closed-loop-2026-05-30.log` and `.logs/validation/vagrant-closed-loop-2026-05-30-nested.log` in the local working tree. The committed validation record is `.logs/validation/sandbox-objective.jsonl`.

## Current Implementation Evidence

| Requirement | Evidence |
|-------------|----------|
| Persistent filesystem | Unit tests verify writable rootfs copy under `storage.data_dir/vms/<vm-id>/rootfs.ext4`, source rootfs immutability, existing VM disk preservation, and VM storage cleanup. |
| Restricted egress | Unit tests verify default-deny chain creation, DNS allow, proxy-only behavior, direct upstream suppression in proxy-only mode, and cleanup commands. |
| Container wrapping | Existing Docker/Podman extraction and layer composer paths remain the container-to-rootfs mechanism; Vagrant validation checks build prerequisites where supported. |
| Secrets/identity | SPIFFE/vsock path remains identity-only. Raw secret broker and Vault exchange are not implemented in this PR. |
| Cross-platform support | Documentation now separates Linux/KVM runtime support from macOS/Windows operator paths. |
| API-driven operation | `GET /capabilities` reports host/runtime readiness, CLI env vars support remote API configuration, and Vagrant forwards host `127.0.0.1:18080` to guest API port `8080`. |

## Known Gaps

- Jailer is not yet the default Firecracker launch path.
- Snapshot/resume methods are currently stubs.
- L7 egress proxy and credential injection are planned sidecar integration, not embedded in `nanofused`.
- Guest-side SPIFFE SVID client is still required for end-to-end identity retrieval inside the VM.
- macOS arm64 Parallels validation depends on whether the provider exposes Linux KVM to the guest; unsupported providers must be reported as KVM-unavailable, not treated as pass.
