# E2E Validation: Snapshot Load / Resume (issue #227, AC1)

## Prerequisites
Real Firecracker snapshot resume requires KVM: a readable/writable `/dev/kvm`
(nested KVM is fine), a `firecracker` binary matching the target version, and
root for tap/bridge networking. The `dev/vagrant` closed-loop box (nested KVM,
pinned Firecracker) satisfies all three; a CI runner without hardware
virtualization does not.

## Validation performed
AC1 was validated on the `dev/vagrant` closed-loop harness (libvirt + nested
KVM, Ubuntu 24.04, Firecracker 1.16.1): a real base microVM was created, a guest
marker written over SSH, then paused → snapshot → stopped → resumed via
`POST /vms/{id}/resume` with `snapshot_id`. The resume booted a **fresh**
Firecracker process (new PID), whose API reported `state:"Running"`, and the
pre-snapshot marker survived the resume (guest memory + disk state restored).
The procedure below is the durable, reproducible recipe for that validation.

## What WAS verified without KVM
- Firecracker `PUT /snapshot/load` request schema (`snapshot_path`,
  `mem_backend{backend_type:"File",backend_path}`, `resume_vm:false` (loads
  paused; LoadSnapshot resumes via a subsequent `PATCH /vm {Resumed}` after the
  vsock proxy is up); deprecated
  `mem_file_path` absent) — unit test `TestSendSnapshotLoadSendsFirecrackerRequest`
  drives a real unix-socket HTTP server; 100% coverage of `sendSnapshotLoad`.
- Schema confirmed against the primary source:
  `firecracker/v1.15.0/src/firecracker/swagger/firecracker.yaml`
  (SnapshotLoadParams + MemoryBackend). The same fields are valid on v1.7.
- `waitForSocketReady` polling (success + timeout) — 100% coverage.
- `LoadSnapshot` input/backing-file validation (nil VM, empty/missing paths).
- Full HTTP handler state machine and every error branch: happy path
  (state -> Running, correct load args), running-VM conflict (409), snapshot
  not found (404), wrong owner (400), missing backing files (404),
  path-outside-root (500), unsupported runtime (501, state restored).

## What REMAINS to validate on KVM
The fresh-process spawn + real `/snapshot/load` round-trip that resumes a live
guest (the part gated on `/dev/kvm`).

## Procedure (run as root on a KVM host with firecracker v1.15)

```bash
# 1. Fixtures (kernel + rootfs), matching the repo target channel.
./scripts/download-fixtures.sh          # test/fixtures/debug-kernel/{vmlinux*,ubuntu-24.04.ext4}

# 2. Build and start the daemon.
mage daemon && sudo ./bin/nanofused &   # needs root for tap/bridge + /dev/kvm

# 3. Register a local image, then run the lifecycle.
VM=$(nanofuse vm create --image <local-image> --name snaptest -o json | jq -r .id)
nanofuse vm start  "$VM"
nanofuse vm exec   "$VM" -- sh -c 'echo marker-$(date +%s) > /root/marker'   # write in-guest state
nanofuse vm pause  "$VM"
SNAP=$(nanofuse vm snapshot create "$VM" -o json | jq -r .id)
nanofuse vm stop   "$VM"                 # process dies; a fresh one loads the snapshot

# 4. THE FEATURE UNDER TEST: resume from the local snapshot.
nanofuse vm resume "$VM" --from-snapshot "$SNAP"

# 5. Assertions (AC1):
#    a) VM state is running:
nanofuse vm get "$VM" -o json | jq -r .state            # expect: running
#    b) A fresh Firecracker process exists and its instance is Running:
curl --unix-socket /var/lib/nanofuse/vms/$VM/firecracker.sock http://localhost/ \
     | jq -r .state                                     # expect: Running
#    c) Guest memory/state was restored (marker survives across the resume):
nanofuse vm exec "$VM" -- cat /root/marker              # expect: the marker written pre-snapshot
```

## Direct runtime-level check (no daemon, mode "none")
For a lighter KVM check of just `firecracker.Manager.{CreateSnapshot,LoadSnapshot}`,
boot a microVM with a minimal initrd and `Network.Mode="none"`, pause, snapshot,
kill, then `LoadSnapshot`, and assert `GET /` returns `state:"Running"`. This
isolates the exact code path changed by this issue from networking and the API
layer. (Not run here: `/dev/kvm` inaccessible to the agent user.)
