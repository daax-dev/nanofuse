# Sandbox Objective Validation

**Date:** 2026-05-30
**Backlog:** `TASK-47`
**Spec:** `.flowspec/features/codex-goal/spec.md`

## Scope

This validation covers the `objective.md` sandbox requirements on the current branch:

- Linux/KVM microVM execution path.
- macOS local microVM execution path through Apple `container` and Virtualization.framework.
- Windows operator path through the `nanofused` API on a reachable Linux or macOS execution environment.
- OCI/container-to-rootfs wrapping path.
- Per-VM persistent filesystem state.
- Short-running and long-running lifecycle behavior.
- SPIFFE/vsock identity posture with raw secrets kept out of VM config.
- Default-deny/proxy-only egress policy for future LLM/API/MCP interception.
- Vagrant/hypervisor closed-loop testing.

## Current Host Finding

The current host is macOS arm64 and has no `/dev/kvm`. Firecracker cannot run natively on this host. Vagrant is installed with the Parallels provider. The local Parallels Ubuntu guest does not expose `/dev/kvm`, so it is not a Firecracker boot host.

The macOS local runtime path uses Apple `container` plus Virtualization.framework instead of nested KVM. It was validated by creating a VM through `nanofused`, starting it through the API, executing `uname -a` inside the Apple-container VM, stopping it through the API, deleting it through the API, and confirming no `nf-*` runtime container remained.

The runnable API requirement is satisfied by `nanofused` on Linux/KVM and by `nanofused` with `runtime.driver=apple_container` on macOS. Windows hosts run as clients using `NANOFUSE_API_URL`, `--api-url`, curl, PowerShell, or the tray app. See `docs/API_QUICK_START.md` and `docs/MAC_WINDOWS_CLIENTS.md`.

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

macOS local runtime:

```bash
./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 5s --api-url http://127.0.0.1:18080
```

API lifecycle:

```bash
API=http://127.0.0.1:18080
VM_NAME="mac-api-alpine-$(date +%s)"
VM_ID="$(curl -fsS -X POST "$API/vms" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${VM_NAME}\",\"image\":\"alpine:3.20\",\"config\":{\"vcpus\":1,\"memory_mib\":256,\"network\":{\"mode\":\"none\"}}}" \
  | jq -r '.id')"
curl -fsS -X POST "$API/vms/$VM_ID/start"
CONTAINER_ID="$(curl -fsS "$API/vms/$VM_ID" | jq -r '.runtime.external_id')"
container exec "$CONTAINER_ID" uname -a
curl -fsS -X POST "$API/vms/$VM_ID/stop" -H "Content-Type: application/json" -d '{"timeout_seconds":10}'
curl -fsS -X DELETE "$API/vms/$VM_ID"
```

Local Vagrant result on 2026-05-30:

- Provider: `vagrant-parallels` with `bento/ubuntu-24.04` arm64.
- Guest boot and rsync succeeded.
- Updated Vagrant port forwarding was applied: guest API `8080` forwarded to host `127.0.0.1:18080`.
- Provisioning failed at the required KVM preflight: `/dev/kvm not found. Nested KVM required`.
- A second attempt enabling Parallels nested virtualization reached `prlctl start` and failed before guest boot with `Unable to start the virtual machine`.
- Guest-side `mage ci`, daemon health/capabilities, and Firecracker boot verification did not run because Firecracker cannot execute without guest KVM.
- The Parallels VM was halted after the latest KVM-unavailable run.

Local macOS runtime result on 2026-05-31:

- Host: macOS arm64 with Apple `container` 0.4.1.
- `container system start --enable-kernel-install` successfully started the Apple-container service.
- `container run --rm alpine:3.20 uname -a` returned Linux `6.12.28` on `aarch64`.
- `./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 5s` passed against `http://127.0.0.1:18080`.
- `POST /vms` and `POST /vms/{id}/start` launched `alpine:3.20` through `runtime.driver=apple_container`.
- `container exec <external_id> uname -a` returned Linux `6.12.28` on `aarch64`.
- `POST /vms/{id}/stop` returned `stopped`.
- `DELETE /vms/{id}` returned HTTP 204 and runtime cleanup left no matching `nf-*` container.

The full host run outputs are stored at `.logs/validation/vagrant-closed-loop-2026-05-30.log` and `.logs/validation/vagrant-closed-loop-2026-05-30-nested.log` in the local working tree. The committed validation record is `.logs/validation/sandbox-objective.jsonl`.

## Current Implementation Evidence

| Requirement | Evidence |
|-------------|----------|
| Persistent filesystem | Unit tests verify writable rootfs copy under `storage.data_dir/vms/<vm-id>/rootfs.ext4`, source rootfs immutability, existing VM disk preservation, and VM storage cleanup. |
| Restricted egress | Unit tests verify default-deny chain creation, DNS allow, proxy-only behavior, direct upstream suppression in proxy-only mode, and cleanup commands. |
| Container wrapping | Existing Docker/Podman extraction and layer composer paths remain the container-to-rootfs mechanism; Vagrant validation checks build prerequisites where supported. |
| Secrets/identity | SPIFFE/vsock path remains identity-only. Raw secret broker and Vault exchange are not implemented in this PR. |
| Cross-platform support | Documentation now separates Linux/KVM Firecracker, macOS Apple-container runtime, and Windows API/tray client support. |
| API-driven operation | `GET /capabilities` reports host/runtime readiness, CLI env vars support remote API configuration, and Vagrant forwards host `127.0.0.1:18080` to guest API port `8080`. |

## Known Gaps

- Jailer is not yet the default Firecracker launch path.
- Snapshot/resume methods are currently stubs.
- L7 egress proxy and credential injection are planned sidecar integration, not embedded in `nanofused`.
- Guest-side SPIFFE SVID client is still required for end-to-end identity retrieval inside the VM.
- macOS arm64 Parallels validation depends on whether the provider exposes Linux KVM to the guest; unsupported providers must be reported as KVM-unavailable, not treated as pass.
- macOS Apple-container runtime does not yet support Nanofuse egress policy enforcement, pause/resume, or snapshots. Those remain Firecracker/Linux-path capabilities or future backend work.
