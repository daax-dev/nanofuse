#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output-fixed"
IMAGE_TAG="ghcr.io/daax-dev/nanofuse/todo-app:latest"

echo "Building fixed NanoFuse Todo-App Image"
echo ""

# Clean output directory
rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

# Step 1: Export Docker image
echo "1. Exporting Docker container rootfs..."
CONTAINER_ID=$(docker create "${IMAGE_TAG}")
docker export "${CONTAINER_ID}" -o "${OUTPUT_DIR}/rootfs.tar"
docker rm "${CONTAINER_ID}"

# Step 2: Extract rootfs
echo "2. Extracting rootfs..."
mkdir -p "${OUTPUT_DIR}/rootfs"
cd "${OUTPUT_DIR}/rootfs"
tar -xf ../rootfs.tar

# Step 3: Create ext4 filesystem
echo "3. Creating ext4 filesystem (requires sudo)..."
ROOTFS_FILE="${OUTPUT_DIR}/rootfs.ext4"
dd if=/dev/zero of="${ROOTFS_FILE}" bs=1 count=0 seek=2G 2>/dev/null
mkfs.ext4 -F "${ROOTFS_FILE}" > /dev/null 2>&1

# Step 4: Mount and copy (requires sudo)
echo "4. Copying files to ext4 image (requires sudo)..."
MOUNT_POINT="${OUTPUT_DIR}/mnt"
mkdir -p "${MOUNT_POINT}"
sudo mount -o loop "${ROOTFS_FILE}" "${MOUNT_POINT}"
sudo rsync -a "${OUTPUT_DIR}/rootfs/" "${MOUNT_POINT}/"
sudo umount "${MOUNT_POINT}"
rmdir "${MOUNT_POINT}"

# Step 5: Copy kernel from old build
echo "5. Using existing kernel..."
KERNEL_FILE="${OUTPUT_DIR}/vmlinux"
cp "${SCRIPT_DIR}/output/vmlinux" "${KERNEL_FILE}"

# Step 6: Register with NanoFuse
echo "6. Registering with NanoFuse (requires sudo)..."
DB_PATH="/var/lib/nanofuse/nanofuse.db"
sudo /home/jpoley/ps/nanofuse/bin/register-local-image \
    "${DB_PATH}" \
    "${IMAGE_TAG}" \
    "${ROOTFS_FILE}" \
    "${KERNEL_FILE}" \
    "x86_64"

# Cleanup
echo "7. Cleaning up..."
rm -rf "${OUTPUT_DIR}/rootfs"
rm -f "${OUTPUT_DIR}/rootfs.tar"

echo ""
echo "Build Complete!"
echo "Image: ${IMAGE_TAG}"
echo ""
