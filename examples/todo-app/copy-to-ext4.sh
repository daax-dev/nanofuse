#!/bin/bash
set -e

ROOTFS_EXT4="$1"
SOURCE_DIR="$2"

echo "Copying files from ${SOURCE_DIR} to ${ROOTFS_EXT4}..."

# Create a tar of the source directory
TMPTAR=$(mktemp)
tar -C "${SOURCE_DIR}" -cf "${TMPTAR}" .

# Use debugfs to extract tar into ext4
{
    echo "cd /"
    echo "write ${TMPTAR} /.tmp.tar"
    echo "quit"
} | debugfs -w "${ROOTFS_EXT4}"

# Extract the tar inside the ext4 using debugfs
# Unfortunately debugfs can't extract tars, we need e2tools or mount
echo "debugfs alone won't work - need mount or e2tools"
rm -f "${TMPTAR}"
