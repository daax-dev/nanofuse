#!/bin/bash
# Build todo-app as a NanoFuse microVM image
# Based on images/base/build.sh pattern

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
IMAGE_TAG="ghcr.io/daax-dev/nanofuse/todo-app:latest"

# Parse flags
NO_CACHE=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-cache)
            NO_CACHE="--no-cache"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--no-cache]"
            exit 1
            ;;
    esac
done

echo "=============================================="
echo "Building NanoFuse Todo-App Image"
echo "=============================================="
echo ""

# Step 1: Build Docker image
echo "1. Building Docker image${NO_CACHE:+ (no cache)}..."
cd "$SCRIPT_DIR"
docker build ${NO_CACHE} -f docker/Dockerfile -t "${IMAGE_TAG}" .

# Step 2: Export container filesystem (NOT docker save - that gives OCI metadata)
echo ""
echo "2. Exporting container filesystem..."
mkdir -p "${OUTPUT_DIR}"

# Create a temporary container and export its filesystem
# docker export gives us the actual Linux filesystem
# docker save would give us OCI image format with blobs/manifests (wrong!)
CONTAINER_ID=$(docker create "${IMAGE_TAG}")
echo "   Created container: ${CONTAINER_ID:0:12}"
docker export "${CONTAINER_ID}" -o "${OUTPUT_DIR}/rootfs.tar"
docker rm "${CONTAINER_ID}" > /dev/null
echo "   ✓ Container filesystem exported"

# Step 3: Extract rootfs from tarball
echo ""
echo "3. Extracting rootfs..."
mkdir -p "${OUTPUT_DIR}/extract"
cd "${OUTPUT_DIR}/extract"
tar -xf ../rootfs.tar

# Step 4: Create ext4 filesystem image
echo ""
echo "4. Creating ext4 filesystem..."
ROOTFS_SIZE="2G"  # 2GB should be enough for todo-app
ROOTFS_FILE="${OUTPUT_DIR}/rootfs.ext4"

# Create empty file
dd if=/dev/zero of="${ROOTFS_FILE}" bs=1 count=0 seek="${ROOTFS_SIZE}" 2>/dev/null

# Create ext4 filesystem
mkfs.ext4 -F "${ROOTFS_FILE}"

# Mount and copy files
MOUNT_POINT="${OUTPUT_DIR}/mnt"
mkdir -p "${MOUNT_POINT}"
sudo mount -o loop "${ROOTFS_FILE}" "${MOUNT_POINT}"

echo "   Copying rootfs contents..."
sudo rsync -a "${OUTPUT_DIR}/extract/" "${MOUNT_POINT}/"

sudo umount "${MOUNT_POINT}"
rmdir "${MOUNT_POINT}"

# Step 5: Get Firecracker kernel from base image build
echo ""
echo "5. Getting kernel from base image (6.1.90)..."
KERNEL_FILE="${OUTPUT_DIR}/vmlinux"
BASE_KERNEL="${SCRIPT_DIR}/../../images/base/build/vmlinux"

if [ -f "${BASE_KERNEL}" ]; then
    echo "   Copying kernel from base image build..."
    cp "${BASE_KERNEL}" "${KERNEL_FILE}"

    # Verify kernel version
    KERNEL_VERSION=$(strings "${KERNEL_FILE}" | grep "Linux version" | head -1)
    echo "   ✓ Kernel: $KERNEL_VERSION"
    echo "   ✓ Size: $(ls -lh ${KERNEL_FILE} | awk '{print $5}')"

    if ! echo "$KERNEL_VERSION" | grep -q "6\.1\."; then
        echo "   ERROR: Wrong kernel version! Expected 6.1.x"
        echo "   Rebuild base image: cd ../../images/base && sudo ./build.sh"
        exit 1
    fi
else
    echo "   ERROR: Base kernel not found at: ${BASE_KERNEL}"
    echo "   Build base image first: cd ../../images/base && sudo ./build.sh"
    exit 1
fi

# Step 6: Register with NanoFuse
echo ""
echo "6. Registering image with NanoFuse..."
DB_PATH="/var/lib/nanofuse/nanofuse.db"

sudo /home/jpoley/ps/nanofuse/bin/register-local-image \
    "${DB_PATH}" \
    "${IMAGE_TAG}" \
    "${ROOTFS_FILE}" \
    "${KERNEL_FILE}" \
    "x86_64"

# Cleanup
echo ""
echo "7. Cleaning up..."
rm -rf "${OUTPUT_DIR}/extract"
rm -f "${OUTPUT_DIR}/rootfs.tar"

echo ""
echo "=============================================="
echo "Build Complete!"
echo "=============================================="
echo ""
echo "Image registered: ${IMAGE_TAG}"
echo "Rootfs: ${ROOTFS_FILE}"
echo "Kernel: ${KERNEL_FILE}"
echo ""
echo "To use:"
echo "  nanofuse vm create ${IMAGE_TAG} my-todo-app --vcpus 2 --memory 1024"
echo ""
