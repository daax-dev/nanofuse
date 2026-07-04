# Hemingway Bridge — Vagrant Closed-Loop / "everything working" goal

Last updated: 2026-07-04 (loop iteration 1). Cron heartbeat job `5f33cd98` (every 10m, session-only).

## Objective (from docs/goal.md + docs/loop.md + /loop prompt)
Get nanofuse **fully working end-to-end**: startup → shutdown → running reliable
containers via the API, for **any container we build**. Validate everything for
real inside a **vagrant nested-KVM sandbox** (only way to get ephemeral root +
/dev/kvm to boot Firecracker). Then work through every open GitHub issue.

## Harness / how to run the closed loop
- Repo harness: `dev/vagrant/` (Vagrantfile = libvirt + nested KVM host-passthrough).
- Host: has /dev/kvm, VT/SVM, vagrant-libvirt 0.12.2, bento/ubuntu-24.04 libvirt box cached.
- User `jpoley` is in the `libvirt` group → `vagrant` reaches qemu:///system **without host sudo**.
  Host `sudo -n` FAILS (needs password) — never rely on host sudo; guest provisioning runs as root.
- Bring up: `cd dev/vagrant && VAGRANT_DEFAULT_PROVIDER=libvirt vagrant up` (VM already created & provisioned).
- Iterate after host edits: `vagrant rsync && vagrant provision` (setup.sh + verify.sh; idempotent).
- Run a command in guest: `vagrant ssh -c "..."`. Repo synced to `/nanofuse` inside guest.
- Full closed loop script: `dev/vagrant/closed-loop.sh`.

## DONE this iteration (validated empirically in guest)
1. **FIXED** `dev/vagrant/setup.sh` Firecracker install — was fetching nonexistent
   `SHA256SUMS`; now uses per-asset `<tarball>.sha256.txt` and saves tarball under
   real asset name. (FIRECRACKER_VERSION still 1.7.0.)
2. **FIXED** `images/base/build.sh` kernel path — sub-script writes vmlinux to
   `scripts/archives/build/vmlinux`; build.sh now searches that path post-build,
   so `images/base/build/vmlinux` gets populated → setup.sh registers it →
   `/var/lib/nanofuse/images/{vmlinux,rootfs.ext4}` populated.
3. `verify.sh` now **PASS 19 / FAIL 0 / SKIP 0**, incl. real Firecracker boot
   (InstanceStart accepted, process alive, VM state Running).
4. Full API lifecycle proven: register-local-image → `vm run <digest>` → running +
   IP 172.16.0.11 + ping 0% loss → `vm stop` (0.13s, process gone) → `vm delete`.

## DONE iteration 2 (validated in sandbox; decisions 034-035; on branch fix/issue-140)
- **FIXED image-ref shorthand:** `setup.sh register_base_image` now registers the base
  image in the daemon DB under canonical tag `ghcr.io/daax-dev/nanofuse/base:latest`
  (via register-local-image). `nanofuse vm run base|default <name>` works out of the box.
  Validated warm + cold-DB + full `vagrant provision`.
- **FIXED ETXTBSY on re-provision:** `setup.sh` installs binaries via copy-to-temp + `mv`
  (atomic rename) so `vagrant provision` is idempotent while nanofused runs.
- `vagrant provision` now completes clean: PASS 19/0/0, "fully verified!".

## DONE iteration 3 (validated; issue #141; branch fix/issue-141-graceful-shutdown-exit; commit 890cbca)
- **FIXED nanofused status=1 on graceful SIGTERM.** server.go treats
  `http.ErrServerClosed` as a clean exit at both serve sites. Validated in guest:
  `systemctl stop nanofused` → rc=0, unit inactive (not failed), Result=success,
  ExecMainStatus=0. Real bind errors still exit non-zero. api tests pass.
  NOTE: this fix is on the #141 branch (off main), NOT the #140 branch.

## DONE iteration 4 (validated; issue #142; branch fix/issue-142-arbitrary-container-kernel; commit b0bb12e)
- **FIXED "run any container we build".** `nanofuse image pull docker.io/library/alpine:latest`
  → "Pull complete!"; `nanofuse vm run <alpine>` boots as a microVM. Three fixes:
  (1) shared fallback kernel wired through registry → builder, default
  `<data_dir>/images/vmlinux`, overridable via new `firecracker.kernel_path`;
  (2) tar `--numeric-owner` + surface tar stderr in errors;
  (3) `nanofused.service` granted CAP_CHOWN/FOWNER/DAC_OVERRIDE/MKNOD/SETFCAP
  (bounding set previously dropped CAP_CHOWN → tar EPERM on etc/shadow even as root).
  Base-image regression: run → running (ping 0% loss) → stop → delete. go test config/builder/api pass.
  NOTE: on #142 branch (off main), NOT #140.

## CORE GOAL STATUS: startup ✅  shutdown ✅ (exit 0, #141)  reliable container via API ✅ (base, ping 0%)  any container ✅ (pull+run alpine, #142)

## OPEN DEFECTS (fix next)
- **[#142 follow-up] non-systemd container networking:** running an arbitrary container
  whose init != /lib/systemd/systemd needs `--kernel-args` override, which clobbers the
  daemon-generated `ip=` autoconfig arg → guest loses networking. Daemon should
  merge/append `ip=`/`root=`/`console=` even when kernel args are user-overridden.
- **[#142 follow-up] two divergent systemd units:** repo-root `nanofused.service`
  (installed by setup.sh) vs `systemd/nanofused.service` (different ExecStart/caps). Reconcile.
- **[HARNESS, #140] daemon left stopped after provision:** verify.sh cleanup trap
  (verify.sh:29 `systemctl stop nanofused`) stops it; env should end with nanofused
  running+enabled so `vm run base` works immediately after `vagrant up`.
- **vm exec unsupported** on firecracker driver but CLI exposes it unconditionally.
- **ttyS0 boot delay** ~90s (`dev-ttyS0.device/start`) though net comes up early.
- Stale `scripts/e2e-test.sh` uses OLD CLI (`vm create <name> --image base`,
  `vm show --format json`). Current CLI: `vm create <image-ref> [name]`, `vm status`.

## TODO (next iterations, in order)
1. Decide git structure: harness fixes (setup.sh, build.sh) → create GH issue →
   branch `fix/issue-NNN-...` (branch name MUST contain issue #; never rename/reuse-with-open-PR).
2. Fix image-ref shorthand so `nanofuse vm run base my-vm` works with a locally built image.
3. "Any container we build": build a custom Docker image → microVM via `nanofuse image build`
   / layer pipeline → run via API. Validate in sandbox.
4. Work open issues: #17 SPIFFE (PR #138 CI ALL FAILING — needs fix), #18 gondolin config,
   #19 cred isolation (PR #139 CI GREEN), #130 snapshot tiering.
5. PR flow per docs/loop.md: local `mage ci` green → 3 adversarial codex review rounds →
   premortem → open PR → Copilot review → close PR, reuse branch, fix, NEW PR → repeat
   until Copilot "generated no issues". NEVER merge to main (human only). NEVER update an open PR.

## Gotchas
- `nanofuse vm delete` prompts interactively — use `--force`.
- Bash tool default timeout 120s; pass `timeout` up to 560000 for long guest builds.
- Kernel build via Docker ~10-15 min first run; cached after. rootfs rebuild ~1-2 min each build.sh run.
