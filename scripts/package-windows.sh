#!/usr/bin/env bash

set -euo pipefail

GO_BIN="${GO_BIN:-go}"
OUT_DIR="${OUT_DIR:-dist}"
PACKAGE_NAME="nanofuse-windows-amd64"
STAGE_DIR="${OUT_DIR}/${PACKAGE_NAME}"
ARCHIVE_PATH="${OUT_DIR}/${PACKAGE_NAME}.zip"

if ! command -v "$GO_BIN" >/dev/null 2>&1; then
    echo "go binary not found: ${GO_BIN}" >&2
    exit 1
fi

rm -rf "$STAGE_DIR" "$ARCHIVE_PATH"
mkdir -p "$STAGE_DIR"

echo "Building Windows CLI..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 "$GO_BIN" build -buildvcs=false -o "${STAGE_DIR}/nanofuse.exe" ./cmd/nanofuse

echo "Building Windows tray..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 "$GO_BIN" build -buildvcs=false -ldflags "-H=windowsgui" -o "${STAGE_DIR}/nanofuse-tray.exe" ./cmd/nanofuse-tray

cp scripts/install-windows.ps1 "${STAGE_DIR}/install-windows.ps1"
cp docs/WINDOWS_RESUME.md "${STAGE_DIR}/WINDOWS_RESUME.md"
cp docs/QUICKSTART-WINDOWS.md "${STAGE_DIR}/QUICKSTART-WINDOWS.md"

if command -v zip >/dev/null 2>&1; then
    (
        cd "$OUT_DIR"
        zip -qr "${PACKAGE_NAME}.zip" "${PACKAGE_NAME}"
    )
elif command -v python3 >/dev/null 2>&1; then
    (
        cd "$OUT_DIR"
        python3 -m zipfile -c "${PACKAGE_NAME}.zip" "${PACKAGE_NAME}"
    )
else
    echo "neither zip nor python3 is available to create ${ARCHIVE_PATH}" >&2
    exit 1
fi

echo "Created ${ARCHIVE_PATH}"
