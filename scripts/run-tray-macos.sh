#!/usr/bin/env bash
set -euo pipefail

api_url="${NANOFUSE_TRAY_API_URL:-${NANOFUSE_API_URL:-http://127.0.0.1:18080}}"
timeout="${NANOFUSE_TRAY_TIMEOUT:-${NANOFUSE_TIMEOUT:-}}"
foreground=0
restart=0
smoke=0
debug=0
log_file="${NANOFUSE_TRAY_LOG:-/tmp/nanofuse-tray.log}"

usage() {
  cat <<'USAGE'
Usage: scripts/run-tray-macos.sh [--api-url URL] [--timeout DURATION] [--foreground] [--restart] [--smoke] [--debug]

Builds and launches the Nanofuse macOS menu bar app.

Options:
  --api-url URL    nanofused API URL. Default: NANOFUSE_TRAY_API_URL, NANOFUSE_API_URL, or http://127.0.0.1:18080
  --timeout VALUE  API request timeout, for example 2s or 500ms
  --foreground     keep the tray app attached to this terminal
  --restart        stop existing nanofuse-tray processes before launching
  --smoke          run an API smoke check and exit without starting the menu bar app
  --debug          enable API client debug logging
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api-url)
      if [[ $# -lt 2 ]]; then
        echo "missing value for --api-url" >&2
        exit 2
      fi
      api_url="$2"
      shift 2
      ;;
    --timeout)
      if [[ $# -lt 2 ]]; then
        echo "missing value for --timeout" >&2
        exit 2
      fi
      timeout="$2"
      shift 2
      ;;
    --foreground)
      foreground=1
      shift
      ;;
    --restart)
      restart=1
      shift
      ;;
    --smoke)
      smoke=1
      shift
      ;;
    --debug)
      debug=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

mkdir -p bin
echo "Building nanofuse-tray..."
go build -o bin/nanofuse-tray ./cmd/nanofuse-tray

args=(--api-url "${api_url}")
if [[ -n "${timeout}" ]]; then
  args+=(--timeout "${timeout}")
fi
if [[ "${debug}" -eq 1 ]]; then
  args+=(--debug)
fi

if [[ "${smoke}" -eq 1 ]]; then
  echo "Running tray API smoke check against ${api_url}..."
  exec ./bin/nanofuse-tray --smoke "${args[@]}"
fi

existing_pids="$(pgrep -f '(^|/)nanofuse-tray( |$)' || true)"
if [[ -n "${existing_pids}" && "${restart}" -eq 1 ]]; then
  echo "Stopping existing nanofuse-tray process(es): ${existing_pids//$'\n'/ }"
  while IFS= read -r pid; do
    [[ -n "${pid}" ]] && kill "${pid}" 2>/dev/null || true
  done <<< "${existing_pids}"
  sleep 1
  existing_pids="$(pgrep -f '(^|/)nanofuse-tray( |$)' || true)"
fi

if [[ -n "${existing_pids}" ]]; then
  echo "nanofuse-tray is already running: ${existing_pids//$'\n'/ }"
  echo "Endpoint requested: ${api_url}"
  echo "Use --restart to stop the existing tray process and launch a new one."
  exit 0
fi

if [[ "${foreground}" -eq 1 ]]; then
  echo "Starting nanofuse-tray in foreground. Endpoint: ${api_url}"
  exec ./bin/nanofuse-tray "${args[@]}"
fi

echo "Starting nanofuse-tray in the background. Endpoint: ${api_url}"
echo "Log: ${log_file}"
./bin/nanofuse-tray "${args[@]}" >"${log_file}" 2>&1 &
pid=$!
sleep 1

if ! kill -0 "${pid}" 2>/dev/null; then
  echo "nanofuse-tray exited during startup. Log:" >&2
  sed -n '1,120p' "${log_file}" >&2 || true
  exit 1
fi

echo "Started nanofuse-tray PID ${pid}."
echo "Look for 'NF' in the macOS menu bar."
echo "Stop it with: kill ${pid}"
