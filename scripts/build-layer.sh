#!/bin/bash
# build-layer.sh - Build and extract NanoFuse runtime layers
#
# This script builds Docker images from layer Dockerfiles and extracts
# the filesystem content into layer rootfs directories.
#
# Usage:
#   ./scripts/build-layer.sh <layer-name>
#   ./scripts/build-layer.sh all
#   ./scripts/build-layer.sh --list
#
# Examples:
#   ./scripts/build-layer.sh python-runtime
#   ./scripts/build-layer.sh node-runtime
#   ./scripts/build-layer.sh go-runtime
#   ./scripts/build-layer.sh all

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
LAYERS_DIR="${PROJECT_ROOT}/layers"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# Available layers (must have Dockerfile)
get_available_layers() {
    local layers=()
    for dir in "${LAYERS_DIR}"/*/; do
        if [[ -f "${dir}Dockerfile" ]]; then
            layers+=("$(basename "${dir}")")
        fi
    done
    echo "${layers[@]}"
}

# Build a single layer
build_layer() {
    local layer_name="$1"
    local layer_dir="${LAYERS_DIR}/${layer_name}"
    local dockerfile="${layer_dir}/Dockerfile"
    local rootfs_dir="${layer_dir}/rootfs"
    local image_name="nanofuse-layer-${layer_name}:latest"

    if [[ ! -f "${dockerfile}" ]]; then
        log_error "Dockerfile not found: ${dockerfile}"
        return 1
    fi

    log_info "Building layer: ${layer_name}"
    log_info "  Dockerfile: ${dockerfile}"
    log_info "  Image: ${image_name}"

    # Build Docker image
    log_info "Building Docker image..."
    if ! docker build -t "${image_name}" -f "${dockerfile}" "${layer_dir}"; then
        log_error "Docker build failed for ${layer_name}"
        return 1
    fi

    # Create temporary container
    log_info "Creating container for extraction..."
    local container_id
    container_id=$(docker create "${image_name}")

    # Clean and recreate rootfs directory
    log_info "Extracting rootfs to ${rootfs_dir}..."
    rm -rf "${rootfs_dir}"
    mkdir -p "${rootfs_dir}"

    # Export container filesystem
    if ! docker export "${container_id}" | tar -xf - -C "${rootfs_dir}"; then
        docker rm "${container_id}" >/dev/null 2>&1 || true
        log_error "Failed to extract rootfs for ${layer_name}"
        return 1
    fi

    # Cleanup container
    docker rm "${container_id}" >/dev/null 2>&1 || true

    # Remove unnecessary files to reduce size
    log_info "Cleaning up extracted rootfs..."
    cleanup_rootfs "${rootfs_dir}"

    # Calculate size
    local size_mb
    size_mb=$(du -sm "${rootfs_dir}" | cut -f1)
    log_info "Layer size: ${size_mb}MB"

    # Generate SHA256 of the layer tarball
    log_info "Generating layer tarball and digest..."
    local tarball="${layer_dir}/${layer_name}.tar.gz"
    tar -czf "${tarball}" -C "${rootfs_dir}" .
    local digest
    digest=$(sha256sum "${tarball}" | cut -d' ' -f1)

    log_success "Layer ${layer_name} built successfully"
    log_info "  Rootfs: ${rootfs_dir}"
    log_info "  Tarball: ${tarball}"
    log_info "  Size: ${size_mb}MB"
    log_info "  SHA256: ${digest}"

    # Update layer.yaml with digest
    update_layer_yaml "${layer_name}" "${digest}" "${size_mb}"

    return 0
}

# Cleanup unnecessary files from rootfs
cleanup_rootfs() {
    local rootfs_dir="$1"

    # Remove apt cache and lists
    rm -rf "${rootfs_dir}/var/lib/apt/lists"/* 2>/dev/null || true
    rm -rf "${rootfs_dir}/var/cache/apt"/* 2>/dev/null || true

    # Remove logs
    rm -rf "${rootfs_dir}/var/log"/* 2>/dev/null || true

    # Remove temporary files
    rm -rf "${rootfs_dir}/tmp"/* 2>/dev/null || true

    # Remove documentation (optional, saves space)
    rm -rf "${rootfs_dir}/usr/share/doc"/* 2>/dev/null || true
    rm -rf "${rootfs_dir}/usr/share/man"/* 2>/dev/null || true

    # Remove locales except en_US
    if [[ -d "${rootfs_dir}/usr/share/locale" ]]; then
        find "${rootfs_dir}/usr/share/locale" -mindepth 1 -maxdepth 1 \
            ! -name 'en*' -type d -exec rm -rf {} + 2>/dev/null || true
    fi
}

# Update layer.yaml with generated digest
update_layer_yaml() {
    local layer_name="$1"
    local digest="$2"
    local size_mb="$3"
    local layer_yaml="${LAYERS_DIR}/${layer_name}/layer.yaml"

    if [[ ! -f "${layer_yaml}" ]]; then
        log_warn "layer.yaml not found: ${layer_yaml}"
        return 0
    fi

    # Check if sha256 field exists, add or update it
    if grep -q "^sha256:" "${layer_yaml}"; then
        sed -i "s/^sha256:.*/sha256: \"${digest}\"/" "${layer_yaml}"
    else
        # Add sha256 after version line
        sed -i "/^version:/a sha256: \"${digest}\"" "${layer_yaml}"
    fi

    # Update or add size field
    if grep -q "^size_mb:" "${layer_yaml}"; then
        sed -i "s/^size_mb:.*/size_mb: ${size_mb}/" "${layer_yaml}"
    else
        sed -i "/^sha256:/a size_mb: ${size_mb}" "${layer_yaml}"
    fi

    log_info "Updated ${layer_yaml} with digest and size"
}

# Show usage
usage() {
    echo "Usage: $0 <layer-name|all|--list>"
    echo ""
    echo "Commands:"
    echo "  <layer-name>  Build a specific layer"
    echo "  all           Build all available layers"
    echo "  --list        List available layers"
    echo ""
    echo "Available layers:"
    for layer in $(get_available_layers); do
        echo "  - ${layer}"
    done
}

# Main entry point
main() {
    if [[ $# -lt 1 ]]; then
        usage
        exit 1
    fi

    local cmd="$1"

    case "${cmd}" in
        --list|-l)
            echo "Available layers:"
            for layer in $(get_available_layers); do
                echo "  - ${layer}"
            done
            ;;
        all)
            log_info "Building all layers..."
            local failed=0
            for layer in $(get_available_layers); do
                if ! build_layer "${layer}"; then
                    log_error "Failed to build ${layer}"
                    ((failed++))
                fi
                echo ""
            done
            if [[ ${failed} -gt 0 ]]; then
                log_error "${failed} layer(s) failed to build"
                exit 1
            fi
            log_success "All layers built successfully"
            ;;
        --help|-h)
            usage
            ;;
        *)
            if [[ ! -d "${LAYERS_DIR}/${cmd}" ]]; then
                log_error "Layer not found: ${cmd}"
                usage
                exit 1
            fi
            build_layer "${cmd}"
            ;;
    esac
}

main "$@"
