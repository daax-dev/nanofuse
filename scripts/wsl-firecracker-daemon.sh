#!/usr/bin/env bash
# wsl-firecracker-daemon.sh - Stand up a real Linux Firecracker nanofused daemon
# inside WSL2 (or any Linux host with /dev/kvm) so a Windows nanofuse.exe client
# can drive the full microVM lifecycle over TCP. This is the closed-loop backend
# for the Windows operator/client path (see docs/WINDOWS_RESUME.md).
#
# Run as root inside WSL:
#   sudo bash scripts/wsl-firecracker-daemon.sh setup   # install deps, build, fetch fixtures, register image
#   sudo bash scripts/wsl-firecracker-daemon.sh run     # setup network + run daemon (foreground)
#   sudo bash scripts/wsl-firecracker-daemon.sh status  # print endpoint, firecracker/KVM, and image rows
#
# Environment knobs:
#   NF_TCP_BIND   default 0.0.0.0:18080
#   NF_DATA_DIR   default /var/lib/nanofuse
#   NF_IMAGE_TAG  default nanofuse-base:latest
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NF_TCP_BIND="${NF_TCP_BIND:-0.0.0.0:18080}"
NF_DATA_DIR="${NF_DATA_DIR:-/var/lib/nanofuse}"
NF_DB="${NF_DATA_DIR}/nanofuse.db"
NF_IMAGE_TAG="${NF_IMAGE_TAG:-nanofuse-base:latest}"
NF_CONFIG="/etc/nanofuse/nanofused.yaml"
NF_EXEC_KEY="${NF_EXEC_KEY:-${NF_DATA_DIR}/exec_id_ed25519}"
FIXTURES="${REPO_ROOT}/test/fixtures/debug-kernel"
GO_VERSION="1.25.0"
GOROOT="/usr/local/go"
# Pin Firecracker by default for reproducible bring-up; set NF_FIRECRACKER_TAG=latest to track upstream.
NF_FIRECRACKER_TAG="${NF_FIRECRACKER_TAG:-v1.15.1}"
# Default HOME for root/non-interactive shells where it may be unset (set -u),
# and export it so child processes (curl/go/ssh-keygen) see it too.
export HOME="${HOME:-/root}"
export PATH="${GOROOT}/bin:${HOME}/go/bin:${PATH}"

# Pinned SHA256 checksums for the Go toolchain tarball (from go.dev/dl JSON).
GO_SHA256_amd64="2852af0cb20a13139b3448992e69b868e50ed0f8a1e5940ee1de9e19a123b613"
GO_SHA256_arm64="05de75d6994a2783699815ee553bd5a9327d8b79991de36e38b66862782f54ae"

log() { echo "[wsl-fc] $*"; }
die() { echo "[wsl-fc][FAIL] $*" >&2; exit 1; }

require_root() { [ "$(id -u)" -eq 0 ] || die "must run as root (e.g. 'sudo bash $0 ...' or 'wsl.exe -d Ubuntu -u root')"; }

install_deps() {
  log "installing apt dependencies"
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y
  # gcc + libc6-dev are sufficient to build go-sqlite3 (it vendors the SQLite
  # amalgamation); libsqlite3-dev is included for builds that opt into the
  # system library via the `libsqlite3` build tag.
  apt-get install -y --no-install-recommends \
    ca-certificates curl gcc libc6-dev libsqlite3-dev make openssh-client \
    sqlite3 squashfs-tools e2fsprogs iproute2 iptables jq xz-utils
}

go_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    *) die "unsupported architecture for Go install: $(uname -m)" ;;
  esac
}

install_go() {
  if "${GOROOT}/bin/go" version 2>/dev/null | grep -q "go${GO_VERSION}"; then
    log "go ${GO_VERSION} already installed"; return
  fi
  local goarch want got
  goarch="$(go_arch)"
  eval "want=\"\${GO_SHA256_${goarch}:-}\""
  [ -n "${want}" ] || die "no pinned Go checksum for arch ${goarch}"
  log "installing go ${GO_VERSION} (linux-${goarch})"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${goarch}.tar.gz" -o /tmp/go.tgz
  # Verify the tarball checksum before extracting as root (supply-chain guard).
  got="$(sha256sum /tmp/go.tgz | awk '{print $1}')"
  [ "${got}" = "${want}" ] || die "go tarball checksum mismatch: got ${got}, want ${want}"
  rm -rf "${GOROOT}"
  tar -C /usr/local -xzf /tmp/go.tgz
  rm -f /tmp/go.tgz
  "${GOROOT}/bin/go" version
}

install_firecracker() {
  local tag arch url tmp have
  tag="${NF_FIRECRACKER_TAG}"
  if command -v firecracker >/dev/null 2>&1; then
    have="$(firecracker --version 2>/dev/null | head -1)"
    # Only accept an existing install when tracking latest or when the present
    # version matches the pinned tag; otherwise reinstall to stay reproducible.
    if [ "${tag}" = "latest" ] || echo "${have}" | grep -qF "${tag}"; then
      log "firecracker present: ${have}"; return
    fi
    log "firecracker present (${have}) does not match pinned ${tag}; reinstalling"
  fi
  if [ "${tag}" = "latest" ]; then
    log "resolving latest firecracker release"
    # Keep the lookup non-fatal under set -euo pipefail: rate limits / transient
    # failures fall back to the pinned default instead of aborting the script.
    tag="$(curl -fsSL https://api.github.com/repos/firecracker-microvm/firecracker/releases/latest 2>/dev/null | jq -r .tag_name 2>/dev/null || true)"
    if [ -z "$tag" ] || [ "$tag" = "null" ]; then
      log "could not resolve latest; falling back to v1.15.1"
      tag="v1.15.1"
    fi
  else
    log "using pinned firecracker ${tag}"
  fi
  arch="$(uname -m)"
  url="https://github.com/firecracker-microvm/firecracker/releases/download/${tag}/firecracker-${tag}-${arch}.tgz"
  log "downloading firecracker ${tag}"
  tmp="$(mktemp -d)"
  curl -fsSL "$url" -o "${tmp}/fc.tgz"
  # Verify against the per-release published checksum before extracting as root.
  local want_fc got_fc
  # Non-fatal under set -euo pipefail so the actionable die message below fires
  # instead of the script aborting on a failed fetch.
  want_fc="$(curl -fsSL "${url}.sha256.txt" 2>/dev/null | awk '{print $1}' || true)"
  [ -n "${want_fc}" ] || die "could not fetch firecracker checksum (${url}.sha256.txt)"
  got_fc="$(sha256sum "${tmp}/fc.tgz" | awk '{print $1}')"
  [ "${got_fc}" = "${want_fc}" ] || die "firecracker checksum mismatch: got ${got_fc}, want ${want_fc}"
  tar -C "${tmp}" -xzf "${tmp}/fc.tgz"
  install -m 0755 "${tmp}/release-${tag}-${arch}/firecracker-${tag}-${arch}" /usr/local/bin/firecracker
  install -m 0755 "${tmp}/release-${tag}-${arch}/jailer-${tag}-${arch}" /usr/local/bin/jailer 2>/dev/null || true
  rm -rf "${tmp}"
  firecracker --version | head -1
}

build_binaries() {
  log "building nanofused + register-local-image (CGO sqlite)"
  cd "${REPO_ROOT}"
  mkdir -p "${REPO_ROOT}/bin"
  export CGO_ENABLED=1 GOCACHE=/root/.cache/go-build GOPATH=/root/go
  # -buildvcs=false means git is never consulted, so there is no need to mutate
  # root's global git config (safe.directory) just to build.
  "${GOROOT}/bin/go" build -buildvcs=false -o ./bin/nanofused ./cmd/nanofused
  "${GOROOT}/bin/go" build -buildvcs=false -o ./bin/register-local-image ./register-local-image.go
  log "built: $(ls -lh ./bin/nanofused ./bin/register-local-image | awk '{print $9, $5}' | tr '\n' ' ')"
}

fetch_fixtures() {
  if [ -f "${FIXTURES}/rootfs.ext4" ] && [ -f "${FIXTURES}/vmlinux.bin" ]; then
    log "fixtures already present"; return
  fi
  log "downloading firecracker CI fixtures"
  cd "${REPO_ROOT}"
  bash scripts/download-fixtures.sh
}

ensure_exec_key() {
  mkdir -p "${NF_DATA_DIR}"
  if [ ! -f "${NF_EXEC_KEY}" ]; then
    log "generating daemon exec SSH key ${NF_EXEC_KEY}"
    ssh-keygen -t ed25519 -N "" -C "nanofuse-exec" -f "${NF_EXEC_KEY}" >/dev/null
  fi
}

# inject_exec_key writes the daemon exec public key into the rootfs image's
# /root/.ssh/authorized_keys so `nanofuse vm exec` can SSH into guests. The
# per-VM rootfs is copied from this image, so every VM inherits the key.
inject_exec_key() {
  ensure_exec_key
  local rootfs mnt pub
  rootfs="$(readlink -f "${FIXTURES}/rootfs.ext4")"
  pub="$(cat "${NF_EXEC_KEY}.pub")"
  mnt="$(mktemp -d)"
  # Install cleanup before mounting so a mount (or later) failure under set -e
  # still unmounts and removes the temp dir.
  trap 'umount "${mnt}" 2>/dev/null || true; rmdir "${mnt}" 2>/dev/null || true' EXIT
  mount -o loop "${rootfs}" "${mnt}"
  mkdir -p "${mnt}/root/.ssh"
  chmod 700 "${mnt}/root/.ssh"
  if ! grep -qxF "${pub}" "${mnt}/root/.ssh/authorized_keys" 2>/dev/null; then
    echo "${pub}" >> "${mnt}/root/.ssh/authorized_keys"
  fi
  chmod 600 "${mnt}/root/.ssh/authorized_keys"
  # Ensure sshd permits root key auth in the guest image.
  if [ -f "${mnt}/etc/ssh/sshd_config" ]; then
    sed -i 's/^#\?PermitRootLogin.*/PermitRootLogin prohibit-password/' "${mnt}/etc/ssh/sshd_config" || true
  fi
  sync
  umount "${mnt}"
  rmdir "${mnt}"
  trap - EXIT
  log "injected exec key into ${rootfs}"
}

# resolve_kernel returns the concrete kernel file. vmlinux.bin is normally a
# symlink, but on a Windows-checked-out tree (/mnt/c, core.symlinks=false) it can
# be materialized as a small text file containing the target name; handle both.
resolve_kernel() {
  local k
  k="$(readlink -f "${FIXTURES}/vmlinux.bin")"
  if [ -f "${k}" ] && [ "$(stat -c%s "${k}" 2>/dev/null || echo 0)" -ge 1000000 ]; then
    echo "${k}"; return
  fi
  # vmlinux.bin is a text "symlink": read its target name and resolve in-dir.
  local target
  target="$(tr -d '\r\n' < "${FIXTURES}/vmlinux.bin")"
  if [ -n "${target}" ] && [ -f "${FIXTURES}/${target}" ]; then
    echo "${FIXTURES}/${target}"; return
  fi
  # Fallback: the largest vmlinux-* file in the fixtures dir.
  ls -S "${FIXTURES}"/vmlinux-* 2>/dev/null | head -1
}

register_image() {
  mkdir -p "${NF_DATA_DIR}"
  local rootfs kernel
  rootfs="$(readlink -f "${FIXTURES}/rootfs.ext4")"
  kernel="$(resolve_kernel)"
  log "registering image ${NF_IMAGE_TAG} -> ${rootfs} + ${kernel}"
  "${REPO_ROOT}/bin/register-local-image" "${NF_DB}" "${NF_IMAGE_TAG}" "${rootfs}" "${kernel}" "$(uname -m)"
}

write_config() {
  mkdir -p /etc/nanofuse "${NF_DATA_DIR}"
  cat >"${NF_CONFIG}" <<EOF
api:
  # Empty socket disables the world-writable Unix listener; this bring-up is
  # Windows-over-TCP only, so a local Unix socket is unnecessary attack surface.
  socket: ""
  tcp_bind: "${NF_TCP_BIND}"
storage:
  data_dir: "${NF_DATA_DIR}"
  database: "${NF_DB}"
firecracker:
  binary_path: /usr/local/bin/firecracker
  exec_ssh_key_path: "${NF_EXEC_KEY}"
  exec_ssh_user: root
network:
  setup: true
limits:
  max_vms: 20
  max_total_memory_mib: 8192
  max_vcpus_per_vm: 4
  max_memory_per_vm_mib: 2048
logging:
  level: info
  format: text
EOF
  log "wrote ${NF_CONFIG} (tcp_bind=${NF_TCP_BIND})"
}

cmd_setup() {
  require_root
  install_deps
  install_go
  install_firecracker
  build_binaries
  fetch_fixtures
  ensure_exec_key
  inject_exec_key
  register_image
  write_config
  log "SETUP COMPLETE"
}

cmd_run() {
  require_root
  if [ ! -r /dev/kvm ] || [ ! -w /dev/kvm ]; then
    die "/dev/kvm is missing or not read/write; enable nested virtualization / KVM for this WSL2 distro (and run as root) before starting the Firecracker daemon."
  fi
  write_config
  case "${NF_TCP_BIND}" in
    0.0.0.0:* | :::* | "[::]:"*)
      log "WARNING: binding the API to ${NF_TCP_BIND} exposes full VM control on all interfaces with no authentication. On a non-isolated host prefer NF_TCP_BIND=127.0.0.1:18080 and reach it from Windows via an SSH tunnel or WSL2 localhost forwarding (a loopback bind is NOT reachable on the WSL NIC IP)."
      ;;
  esac
  log "starting nanofused (firecracker) on ${NF_TCP_BIND}"
  exec "${REPO_ROOT}/bin/nanofused" -config "${NF_CONFIG}" -tcp "${NF_TCP_BIND}"
}

cmd_status() {
  local fcbin fcver
  fcbin="$(command -v firecracker || echo MISSING)"
  fcver=""
  # Only query the version when firecracker exists, so status works before setup
  # even under set -euo pipefail.
  if [ "${fcbin}" != "MISSING" ]; then
    fcver="$(firecracker --version 2>/dev/null | head -1 || true)"
  fi
  echo "WSL IP(s): $(hostname -I || true)"
  echo "TCP bind:  ${NF_TCP_BIND}"
  echo "Firecracker: ${fcbin} ${fcver}"
  echo "KVM: $(ls -l /dev/kvm 2>&1 || true)"
  echo "Image rows:"
  command -v sqlite3 >/dev/null 2>&1 && sqlite3 "${NF_DB}" 'select tag,architecture from image_tags;' 2>/dev/null || true
}

case "${1:-}" in
  setup) cmd_setup ;;
  run) cmd_run ;;
  status) cmd_status ;;
  *) echo "usage: $0 {setup|run|status}"; exit 2 ;;
esac
