#!/usr/bin/env bash
# wsl-firecracker-daemon.sh - Stand up a real Linux Firecracker nanofused daemon
# inside WSL2 (or any Linux host with /dev/kvm) so a Windows nanofuse.exe client
# can drive the full microVM lifecycle over TCP. This is the closed-loop backend
# for the Windows operator/client path (see docs/WINDOWS_RESUME.md).
#
# Run as root inside WSL:
#   sudo bash scripts/wsl-firecracker-daemon.sh setup   # install deps, build, fetch fixtures, register image
#   sudo bash scripts/wsl-firecracker-daemon.sh run     # setup network + run daemon (foreground)
#   sudo bash scripts/wsl-firecracker-daemon.sh status  # print endpoint + image + vm state
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
FIXTURES="${REPO_ROOT}/test/fixtures/debug-kernel"
GO_VERSION="1.25.0"
GOROOT="/usr/local/go"
export PATH="${GOROOT}/bin:${HOME}/go/bin:${PATH}"

log() { echo "[wsl-fc] $*"; }
die() { echo "[wsl-fc][FAIL] $*" >&2; exit 1; }

require_root() { [ "$(id -u)" -eq 0 ] || die "must run as root (wsl.exe -u root)"; }

install_deps() {
  log "installing apt dependencies"
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y
  apt-get install -y --no-install-recommends \
    ca-certificates curl gcc libc6-dev make \
    squashfs-tools e2fsprogs iproute2 iptables jq xz-utils
}

install_go() {
  if "${GOROOT}/bin/go" version 2>/dev/null | grep -q "go${GO_VERSION}"; then
    log "go ${GO_VERSION} already installed"; return
  fi
  log "installing go ${GO_VERSION}"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
  rm -rf "${GOROOT}"
  tar -C /usr/local -xzf /tmp/go.tgz
  "${GOROOT}/bin/go" version
}

install_firecracker() {
  if command -v firecracker >/dev/null 2>&1; then
    log "firecracker present: $(firecracker --version | head -1)"; return
  fi
  log "resolving latest firecracker release"
  local tag arch url tmp
  tag="$(curl -fsSL https://api.github.com/repos/firecracker-microvm/firecracker/releases/latest | jq -r .tag_name)"
  [ -n "$tag" ] && [ "$tag" != "null" ] || tag="v1.13.1"
  arch="$(uname -m)"
  url="https://github.com/firecracker-microvm/firecracker/releases/download/${tag}/firecracker-${tag}-${arch}.tgz"
  log "downloading firecracker ${tag}"
  tmp="$(mktemp -d)"
  curl -fsSL "$url" -o "${tmp}/fc.tgz"
  tar -C "${tmp}" -xzf "${tmp}/fc.tgz"
  install -m 0755 "${tmp}/release-${tag}-${arch}/firecracker-${tag}-${arch}" /usr/local/bin/firecracker
  install -m 0755 "${tmp}/release-${tag}-${arch}/jailer-${tag}-${arch}" /usr/local/bin/jailer 2>/dev/null || true
  rm -rf "${tmp}"
  firecracker --version | head -1
}

build_binaries() {
  log "building nanofused + register-local-image (CGO sqlite)"
  cd "${REPO_ROOT}"
  export CGO_ENABLED=1 GOCACHE=/root/.cache/go-build GOPATH=/root/go
  git config --global --add safe.directory "${REPO_ROOT}" 2>/dev/null || true
  # Repo lives on /mnt/c (Windows-owned); disable VCS stamping to avoid git ownership errors.
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

register_image() {
  mkdir -p "${NF_DATA_DIR}"
  local rootfs kernel
  rootfs="$(readlink -f "${FIXTURES}/rootfs.ext4")"
  kernel="$(readlink -f "${FIXTURES}/vmlinux.bin")"
  log "registering image ${NF_IMAGE_TAG} -> ${rootfs} + ${kernel}"
  "${REPO_ROOT}/bin/register-local-image" "${NF_DB}" "${NF_IMAGE_TAG}" "${rootfs}" "${kernel}" "$(uname -m)"
}

write_config() {
  mkdir -p /etc/nanofuse "${NF_DATA_DIR}"
  cat >"${NF_CONFIG}" <<EOF
api:
  socket: /var/run/nanofused.sock
  tcp_bind: "${NF_TCP_BIND}"
storage:
  data_dir: ${NF_DATA_DIR}
  database: ${NF_DB}
firecracker:
  binary_path: /usr/local/bin/firecracker
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
  register_image
  write_config
  log "SETUP COMPLETE"
}

cmd_run() {
  require_root
  write_config
  log "starting nanofused (firecracker) on ${NF_TCP_BIND}"
  exec "${REPO_ROOT}/bin/nanofused" -config "${NF_CONFIG}" -tcp "${NF_TCP_BIND}"
}

cmd_status() {
  echo "WSL IP(s): $(hostname -I)"
  echo "TCP bind:  ${NF_TCP_BIND}"
  echo "Firecracker: $(command -v firecracker || echo MISSING) $(firecracker --version 2>/dev/null | head -1)"
  echo "KVM: $(ls -l /dev/kvm 2>&1)"
  echo "Image rows:"
  command -v sqlite3 >/dev/null 2>&1 && sqlite3 "${NF_DB}" 'select tag,architecture from image_tags;' 2>/dev/null || true
}

case "${1:-}" in
  setup) cmd_setup ;;
  run) cmd_run ;;
  status) cmd_status ;;
  *) echo "usage: $0 {setup|run|status}"; exit 2 ;;
esac
