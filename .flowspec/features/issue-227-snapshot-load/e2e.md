# E2E Validation: Snapshot Load / Resume (issue #227, AC1)

## Status in this environment
Real Firecracker snapshot resume requires KVM. It was **NOT run** here because:
- The agent user is not in the `kvm` group and `/dev/kvm` is not readable/writable.
- No passwordless `sudo`, so the user cannot be added to `kvm`, nor can
  vagrant/libvirt be brought up (both need root for `/dev/kvm` and networking).
- The host `firecracker` binary is **v1.7.0**, while this repo targets the
  Firecracker CI **v1.15** channel (see `scripts/download-fixtures.sh`). An e2e
  on v1.7 would not validate the intended target version.

No results were fabricated. The steps below are the exact procedure to validate
AC1 on a root/KVM sandbox (e.g. the vagrant-skill box with nested KVM).

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
