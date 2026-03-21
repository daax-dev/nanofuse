#!/bin/bash
# compose-image.sh - Compose NanoFuse microVM image from layers
#
# This script takes an image manifest and composes it into a bootable ext4 image
# by stacking layers in dependency order.
#
# Usage:
#   ./scripts/compose-image.sh <manifest-path> [output-path]
#
# Examples:
#   ./scripts/compose-image.sh images/falcondev-agents/image.manifest.yaml
#   ./scripts/compose-image.sh images/falcondev-agents/image.manifest.yaml build/falcondev.ext4

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

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

# Default values
DEFAULT_OUTPUT_DIR="${PROJECT_ROOT}/build"
DEFAULT_IMAGE_SIZE_MB=2048

# Check for required tools
check_requirements() {
    local missing=()

    for cmd in yq mkfs.ext4 rsync fakeroot; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required tools: ${missing[*]}"
        log_info "Install with: sudo apt-get install yq e2fsprogs rsync fakeroot"
        exit 1
    fi
}

# Parse manifest and extract layer info
parse_manifest() {
    local manifest="$1"

    if [[ ! -f "$manifest" ]]; then
        log_error "Manifest not found: $manifest"
        exit 1
    fi

    # Extract image name
    IMAGE_NAME=$(yq -r '.name' "$manifest")

    # Extract kernel info
    KERNEL_SOURCE=$(yq -r '.kernel.source' "$manifest")
    KERNEL_CMDLINE=$(yq -r '.kernel.cmdline' "$manifest")

    # Extract output config
    OUTPUT_SIZE_MB=$(yq -r '.output.size_mb // 2048' "$manifest")
    OUTPUT_FORMAT=$(yq -r '.output.format // "ext4"' "$manifest")

    log_info "Image: $IMAGE_NAME"
    log_info "Kernel: $KERNEL_SOURCE"
    log_info "Size: ${OUTPUT_SIZE_MB}MB"
}

# Get ordered list of layers (respecting dependencies)
get_layer_order() {
    local manifest="$1"

    # For now, use the order in the manifest (should be dependency-ordered)
    # TODO: Implement proper topological sort based on dependencies
    yq -r '.layers[].name' "$manifest"
}

# Resolve layer source to actual path
resolve_layer_source() {
    local source="$1"

    # Handle local:// sources
    if [[ "$source" == local://* ]]; then
        local path="${source#local://}"
        echo "${PROJECT_ROOT}/${path}"
    else
        log_error "Unsupported source type: $source"
        exit 1
    fi
}

# Check if layer should be included (evaluate conditions)
should_include_layer() {
    local manifest="$1"
    local layer_name="$2"

    local required=$(yq -r ".layers[] | select(.name == \"$layer_name\") | .required // false" "$manifest")
    local condition=$(yq -r ".layers[] | select(.name == \"$layer_name\") | .condition // \"\"" "$manifest")

    # Required layers always included
    if [[ "$required" == "true" ]]; then
        echo "true"
        return
    fi

    # No condition means include
    if [[ -z "$condition" || "$condition" == "null" ]]; then
        echo "true"
        return
    fi

    # Evaluate condition (simple ${VAR:-default} syntax)
    # Extract variable name and default
    if [[ "$condition" =~ \$\{([A-Z_]+):-([a-z]+)\} ]]; then
        local var_name="${BASH_REMATCH[1]}"
        local default_val="${BASH_REMATCH[2]}"
        local value="${!var_name:-$default_val}"

        if [[ "$value" == "true" ]]; then
            echo "true"
        else
            echo "false"
        fi
    else
        # Unknown condition format, include by default
        echo "true"
    fi
}

# Create ext4 image from layers
compose_image() {
    local manifest="$1"
    local output_path="$2"
    local size_mb="${3:-$DEFAULT_IMAGE_SIZE_MB}"

    local temp_dir=$(mktemp -d)
    local rootfs_dir="${temp_dir}/rootfs"
    local fakeroot_state="${temp_dir}/fakeroot.state"

    trap "cleanup_compose '$temp_dir'" EXIT

    mkdir -p "$rootfs_dir"

    log_info "Composing layers into rootfs (with fakeroot for proper ownership)..."

    # Get layer order
    local layers=$(get_layer_order "$manifest")

    # Stack each layer using fakeroot to preserve root ownership
    for layer_name in $layers; do
        # Check if layer should be included
        local include=$(should_include_layer "$manifest" "$layer_name")
        if [[ "$include" != "true" ]]; then
            log_info "Skipping layer (condition not met): $layer_name"
            continue
        fi

        local source=$(yq -r ".layers[] | select(.name == \"$layer_name\") | .source" "$manifest")
        local layer_path=$(resolve_layer_source "$source")
        local layer_rootfs="${layer_path}/rootfs"

        if [[ ! -d "$layer_rootfs" ]]; then
            log_warn "Layer rootfs not found: $layer_rootfs"
            log_warn "  Try building it: ./scripts/build-layer.sh $layer_name"
            continue
        fi

        log_info "  Adding layer: $layer_name ($(du -sh "$layer_rootfs" | cut -f1))"

        # Copy layer content using fakeroot to track root ownership
        # The -s option saves fakeroot state between invocations
        fakeroot -s "$fakeroot_state" -i "$fakeroot_state" -- \
            rsync -a --quiet "$layer_rootfs/" "$rootfs_dir/" 2>/dev/null || \
        fakeroot -s "$fakeroot_state" -- \
            rsync -a --quiet "$layer_rootfs/" "$rootfs_dir/"

        # Run post-install hook if exists (also under fakeroot)
        local hook="${layer_path}/hooks/post-install.sh"
        if [[ -f "$hook" ]]; then
            log_info "    Running post-install hook..."
            LAYER_NAME="$layer_name" \
            ROOTFS_PATH="$rootfs_dir" \
            fakeroot -s "$fakeroot_state" -i "$fakeroot_state" -- \
                bash "$hook" || log_warn "    Hook returned non-zero"
        fi
    done

    # Calculate actual rootfs size and adjust if needed
    local rootfs_size_mb=$(du -sm "$rootfs_dir" | cut -f1)
    local min_size_mb=$((rootfs_size_mb + 100)) # Add 100MB buffer

    if [[ $size_mb -lt $min_size_mb ]]; then
        log_warn "Requested size ${size_mb}MB too small for rootfs (${rootfs_size_mb}MB)"
        size_mb=$min_size_mb
        log_info "Adjusting image size to ${size_mb}MB"
    fi

    log_info "Creating ${size_mb}MB ext4 image from rootfs..."

    # Create output directory
    mkdir -p "$(dirname "$output_path")"

    # Use mke2fs with -d option under fakeroot to preserve ownership
    # The fakeroot state contains the mapping of fake UID/GID -> real UID/GID
    fakeroot -i "$fakeroot_state" -- \
        mke2fs -t ext4 -d "$rootfs_dir" -L nanofuse -F -q \
            "$output_path" "${size_mb}M"

    # Calculate SHA256
    local sha256=$(sha256sum "$output_path" | cut -d' ' -f1)
    local final_size=$(du -h "$output_path" | cut -f1)

    log_success "Image composed successfully"
    log_info "  Output: $output_path"
    log_info "  Size: $final_size"
    log_info "  SHA256: $sha256"
}

cleanup_compose() {
    local temp_dir="$1"

    # Remove temp dir
    rm -rf "$temp_dir" 2>/dev/null || true
}

# Show usage
usage() {
    echo "Usage: $0 <manifest-path> [output-path]"
    echo ""
    echo "Compose a NanoFuse microVM image from layers."
    echo ""
    echo "Arguments:"
    echo "  manifest-path   Path to image.manifest.yaml"
    echo "  output-path     Output image path (default: build/<image-name>.ext4)"
    echo ""
    echo "Environment variables:"
    echo "  INCLUDE_RECORDING=true/false   Include recording-agent layer"
    echo ""
    echo "Examples:"
    echo "  $0 images/falcondev-agents/image.manifest.yaml"
    echo "  $0 images/falcondev-agents/image.manifest.yaml build/custom.ext4"
    echo "  INCLUDE_RECORDING=false $0 images/falcondev-agents/image.manifest.yaml"
}

# Main entry point
main() {
    if [[ $# -lt 1 ]]; then
        usage
        exit 1
    fi

    local manifest="$1"

    # Check requirements
    check_requirements

    # Parse manifest
    parse_manifest "$manifest"

    # Determine output path
    local output_path="${2:-${DEFAULT_OUTPUT_DIR}/${IMAGE_NAME}.ext4}"

    # Get image size from manifest
    local size_mb="$OUTPUT_SIZE_MB"

    # Compose the image
    compose_image "$manifest" "$output_path" "$size_mb"
}

main "$@"
