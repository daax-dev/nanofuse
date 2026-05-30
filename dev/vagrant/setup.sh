#!/usr/bin/env bash
# setup.sh — nanofuse dev environment provisioner
#
# Installs everything needed to build nanofuse from source, build base
# microVM images, and run Firecracker VMs inside a Vagrant VM with nested KVM.
#
# Idempotent: safe to run multiple times.
# Run as root (Vagrant provisioner handles this).

set -euo pipefail

# ─── Colors ─────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

# ─── Versions ────────────────────────────────────────────────────────────────
GO_VERSION="1.24.3"
FIRECRACKER_VERSION="1.7.0"

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        GO_ARCH="amd64"
        FIRECRACKER_ARCH="x86_64"
        NANOFUSE_IMAGE_ARCH="x86_64"
        ;;
    aarch64|arm64)
        GO_ARCH="arm64"
        FIRECRACKER_ARCH="aarch64"
        NANOFUSE_IMAGE_ARCH="aarch64"
        ;;
    *)
        error "Unsupported guest architecture: $ARCH"
        ;;
esac

# ─── Preflight ───────────────────────────────────────────────────────────────
info "Checking prerequisites..."
[[ $EUID -eq 0 ]] || error "Must run as root"

if [[ ! -e /dev/kvm ]]; then
    error "/dev/kvm not found. Nested KVM required — ensure host has host-passthrough CPU mode."
fi

if [[ ! -d /nanofuse ]]; then
    error "/nanofuse not found — Vagrant synced_folder failed"
fi

info "Prerequisites OK: KVM available, root, ${ARCH}, /nanofuse present"

# ─── 1. System packages ────────────────────────────────────────────────────
install_system_deps() {
    info "Installing system packages (apt-get install is idempotent)..."
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y -qq \
        build-essential \
        gcc \
        curl \
        git \
        jq \
        iptables \
        dnsmasq \
        dnsutils \
        net-tools \
        iproute2 \
        ca-certificates \
        gnupg \
        sqlite3 \
        libsqlite3-dev \
        procps \
        strace \
        util-linux \
        > /dev/null

    # dnsmasq conflicts with systemd-resolved on port 53; disable it for now
    # (setup scripts for security layers will configure dnsmasq properly)
    systemctl stop dnsmasq 2>/dev/null || true
    systemctl disable dnsmasq 2>/dev/null || true

    info "System packages installed"
}

# ─── 2. Docker ──────────────────────────────────────────────────────────────
install_docker() {
    if command -v docker &>/dev/null; then
        info "Docker already installed: $(docker --version | awk '{print $3}' | tr -d ',')"
        return 0
    fi

    info "Installing Docker..."
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc

    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io > /dev/null

    systemctl enable --now docker
    usermod -aG docker vagrant

    info "Docker installed"
}

# ─── 3. Go ──────────────────────────────────────────────────────────────────
install_go() {
    if [[ -x /usr/local/go/bin/go ]]; then
        local installed_ver
        installed_ver=$(/usr/local/go/bin/go version | awk '{print $3}' | tr -d 'go')
        if [[ "$installed_ver" == "$GO_VERSION" ]]; then
            info "Go already installed: v${installed_ver}"
            return 0
        fi
        info "Upgrading Go from v${installed_ver} to v${GO_VERSION}..."
        rm -rf /usr/local/go
    fi

    info "Installing Go v${GO_VERSION}..."
    local go_tarball="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    local go_base_url="https://go.dev/dl"
    local tmpdir
    tmpdir="$(mktemp -d)"

    curl -fsSL "${go_base_url}/${go_tarball}" -o "${tmpdir}/${go_tarball}"
    curl -fsSL "${go_base_url}/${go_tarball}.sha256" -o "${tmpdir}/${go_tarball}.sha256"

    info "Verifying Go tarball checksum..."
    (cd "${tmpdir}" && sha256sum -c "${go_tarball}.sha256")

    info "Extracting Go to /usr/local..."
    tar -C /usr/local -xzf "${tmpdir}/${go_tarball}"
    rm -rf "${tmpdir}"

    # Make Go available system-wide (login + non-login shells)
    cat > /etc/profile.d/go.sh << 'GOEOF'
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
export GOPATH="$HOME/go"
GOEOF

    info "Go v${GO_VERSION} installed"
}

# ─── 4. Mage ───────────────────────────────────────────────────────────────
install_mage() {
    export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
    export GOPATH="$HOME/go"

    if command -v mage &>/dev/null; then
        info "Mage already installed"
        return 0
    fi

    info "Installing mage..."
    go install github.com/magefile/mage@latest
    cp "$HOME/go/bin/mage" /usr/local/bin/mage

    info "Mage installed"
}

# ─── 5. Firecracker ────────────────────────────────────────────────────────
install_firecracker() {
    if [[ -x /usr/local/bin/firecracker ]]; then
        info "Firecracker already installed: $(firecracker --version 2>&1 | head -1)"
        return 0
    fi

    info "Installing Firecracker v${FIRECRACKER_VERSION}..."
    local tmpdir
    tmpdir=$(mktemp -d)

    local fc_base_url="https://github.com/firecracker-microvm/firecracker/releases/download/v${FIRECRACKER_VERSION}"
    curl -fsSL "${fc_base_url}/firecracker-v${FIRECRACKER_VERSION}-${FIRECRACKER_ARCH}.tgz" \
        -o "$tmpdir/firecracker.tgz"
    curl -fsSL "${fc_base_url}/SHA256SUMS" -o "$tmpdir/SHA256SUMS"

    info "Verifying Firecracker tarball checksum..."
    (cd "$tmpdir" && grep "firecracker-v${FIRECRACKER_VERSION}-${FIRECRACKER_ARCH}.tgz" SHA256SUMS | sha256sum -c -)

    tar -C "$tmpdir" -xzf "$tmpdir/firecracker.tgz"

    local release_dir="$tmpdir/release-v${FIRECRACKER_VERSION}-${FIRECRACKER_ARCH}"
    cp "$release_dir/firecracker-v${FIRECRACKER_VERSION}-${FIRECRACKER_ARCH}" /usr/local/bin/firecracker
    cp "$release_dir/jailer-v${FIRECRACKER_VERSION}-${FIRECRACKER_ARCH}" /usr/local/bin/jailer
    chmod +x /usr/local/bin/firecracker /usr/local/bin/jailer

    rm -rf "$tmpdir"
    info "Firecracker v${FIRECRACKER_VERSION} installed"
}

# ─── 6. Build nanofuse from source ─────────────────────────────────────────
build_nanofuse() {
    info "Building nanofuse from source..."
    export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
    export GOPATH="$HOME/go"

    cd /nanofuse

    # Download Go modules first (better error messages)
    go mod download 2>&1 | tail -5

    # Build all binaries (CLI + daemon + register-local-image)
    # Redirect to log — mage output is verbose and .git warnings are expected (rsync excludes .git)
    mage all > /tmp/nanofuse-build.log 2>&1 || {
        error "nanofuse build failed. Last 20 lines:"
        tail -20 /tmp/nanofuse-build.log
        exit 1
    }

    # Install to /usr/local/bin so they're on PATH for all users
    cp bin/nanofuse /usr/local/bin/nanofuse
    cp bin/nanofused /usr/local/bin/nanofused
    if [[ -f bin/register-local-image ]]; then
        cp bin/register-local-image /usr/local/bin/register-local-image
    fi
    chmod +x /usr/local/bin/nanofuse /usr/local/bin/nanofused

    info "nanofuse built and installed:"
    info "  $(nanofuse version 2>&1 || echo 'version check skipped')"
}

# ─── 7. Build base microVM image ───────────────────────────────────────────
build_base_image() {
    local build_dir="/nanofuse/images/base/build"
    local img_dir="/var/lib/nanofuse/images"
    mkdir -p "$build_dir" "$img_dir"

    if [[ "$NANOFUSE_IMAGE_ARCH" != "x86_64" ]]; then
        warn "Base image build is currently x86_64-only; guest architecture is $ARCH."
        warn "Closed-loop VM boot requires a matching ${NANOFUSE_IMAGE_ARCH} kernel and rootfs."
        warn "Skipping base image build on this guest architecture."
        return 0
    fi

    # ── Use pre-built kernel from vagrant-scripts/ if available (cache across rebuilds) ──
    if [[ -f "$build_dir/vmlinux" ]]; then
        info "Kernel already present: $(ls -lh $build_dir/vmlinux | awk '{print $5}')"
    elif [[ -f /vagrant-scripts/vmlinux ]]; then
        info "Copying pre-built kernel from vagrant-scripts/vmlinux..."
        cp /vagrant-scripts/vmlinux "$build_dir/vmlinux"
    fi

    # ── Build via nanofuse's build.sh (builds both kernel 6.1.90 + rootfs) ──
    if [[ -f "$build_dir/vmlinux" ]] && [[ -f "$build_dir/rootfs.ext4" ]]; then
        info "Base image already built — kernel + rootfs present"
    else
        info "Building base image via build.sh (kernel 6.1.90 + rootfs — log at /tmp/image-build.log)..."
        info "  This builds the kernel from source via Docker — may take 10-15 min on first run."
        cd /nanofuse/images/base

        # Redirect verbose Docker output to a log file to avoid
        # overwhelming Vagrant's SSH output buffer
        ./build.sh > /tmp/image-build.log 2>&1 || {
            warn "Build script returned non-zero. Last 20 lines of log:"
            tail -20 /tmp/image-build.log
        }
    fi

    # Verify both artifacts
    if [[ -f "$build_dir/vmlinux" ]] && [[ -f "$build_dir/rootfs.ext4" ]]; then
        info "Base image complete:"
        info "  kernel: $(ls -lh $build_dir/vmlinux | awk '{print $5}')"
        info "  rootfs: $(ls -lh $build_dir/rootfs.ext4 | awk '{print $5}')"
    else
        warn "Base image incomplete — some artifacts missing"
        warn "  vmlinux:     $(test -f $build_dir/vmlinux && echo 'OK' || echo 'MISSING')"
        warn "  rootfs.ext4: $(test -f $build_dir/rootfs.ext4 && echo 'OK' || echo 'MISSING')"
        warn "  Retry: cd /nanofuse/images/base && sudo ./build.sh"
    fi
}

# ─── 8. Register base image with nanofused ──────────────────────────────────
register_base_image() {
    local build_dir="/nanofuse/images/base/build"

    if [[ ! -f "$build_dir/vmlinux" ]] || [[ ! -f "$build_dir/rootfs.ext4" ]]; then
        warn "Cannot register base image — build artifacts missing"
        return 0
    fi

    info "Setting up nanofuse data directories..."
    mkdir -p /var/lib/nanofuse/images
    mkdir -p /tmp/nanofuse

    # Copy artifacts to where nanofused expects them
    cp "$build_dir/vmlinux" /var/lib/nanofuse/images/vmlinux
    cp "$build_dir/rootfs.ext4" /var/lib/nanofuse/images/rootfs.ext4
    if [[ -f "$build_dir/manifest.json" ]]; then
        cp "$build_dir/manifest.json" /var/lib/nanofuse/images/manifest.json
    fi

    info "Base image registered at /var/lib/nanofuse/images/"
}

# ─── 9. nanofuse config + systemd service ──────────────────────────────────
setup_nanofuse_service() {
    info "Setting up nanofuse configuration..."

    mkdir -p /etc/nanofuse /var/lib/nanofuse /tmp/nanofuse

    cat > /etc/nanofuse/nanofused.yaml << 'EOF'
# nanofuse config (Vagrant dev VM)
api:
  socket: /var/run/nanofused.sock
  socket_mode: "0660"
  tcp_bind: "0.0.0.0:8080"

storage:
  data_dir: /var/lib/nanofuse
  database: /var/lib/nanofuse/nanofuse.db

firecracker:
  binary_path: /usr/local/bin/firecracker

limits:
  max_vms: 10
  max_total_memory_mib: 3072
  max_vcpus_per_vm: 2
  max_memory_per_vm_mib: 1024
  max_snapshot_storage_gib: 10

logging:
  level: debug
  format: text
  console_log_max_size_mb: 10
  console_log_max_backups: 3
EOF

    # Install systemd service if the unit file exists in the repo
    if [[ -f /nanofuse/nanofused.service ]]; then
        cp /nanofuse/nanofused.service /etc/systemd/system/nanofused.service
    else
        cat > /etc/systemd/system/nanofused.service << 'EOF'
[Unit]
Description=NanoFuse API Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nanofused -config /etc/nanofuse/nanofused.yaml
Restart=on-failure
RestartSec=5s
StateDirectory=nanofuse
User=root
Group=root
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW CAP_SYS_ADMIN
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF
    fi

    systemctl daemon-reload
    systemctl enable nanofused

    info "nanofuse config: /etc/nanofuse/nanofused.yaml"
    info "nanofused service installed (not started — start manually or via verify)"
}

# ─── Run ─────────────────────────────────────────────────────────────────────
info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
info "nanofuse dev environment setup"
info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

install_system_deps
install_docker
install_go
install_mage
install_firecracker
build_nanofuse
build_base_image
register_base_image
setup_nanofuse_service

info ""
info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
info "Setup complete!"
info ""
info "  nanofuse source:  /nanofuse"
info "  binaries:         /usr/local/bin/nanofuse, nanofused"
info "  config:           /etc/nanofuse/nanofused.yaml"
info "  base image:       /var/lib/nanofuse/images/"
info "  data:             /var/lib/nanofuse/"
info "  guest API:        http://0.0.0.0:8080"
info "  host API:         http://127.0.0.1:18080 (Vagrant forwarded port; override with NANOFUSE_API_HOST_PORT)"
info ""
info "Quick start:"
info "  sudo systemctl start nanofused"
info "  nanofuse health"
info "  curl http://127.0.0.1:18080/health   # from host when Vagrant port forwarding is active"
info "  nanofuse vm list"
info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
