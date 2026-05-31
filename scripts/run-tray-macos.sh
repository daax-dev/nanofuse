#!/usr/bin/env bash
set -euo pipefail

api_url="${NANOFUSE_API_URL:-http://127.0.0.1:18080}"
mkdir -p bin
go build -o bin/nanofuse-tray ./cmd/nanofuse-tray
exec ./bin/nanofuse-tray --api-url "${api_url}"
