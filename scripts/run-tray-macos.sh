#!/usr/bin/env bash
set -euo pipefail

api_url="${NANOFUSE_TRAY_API_URL:-${NANOFUSE_API_URL:-http://127.0.0.1:18080}}"
timeout="${NANOFUSE_TRAY_TIMEOUT:-${NANOFUSE_TIMEOUT:-}}"
foreground=0
restart=0
smoke=0
debug=0
start_api=0
log_file="${NANOFUSE_TRAY_LOG:-/tmp/nanofuse-tray.log}"
api_log_file="${NANOFUSE_API_LOG:-/tmp/nanofused-macos.log}"
api_pid_file="${NANOFUSE_API_PID_FILE:-/tmp/nanofused-macos.pid}"
api_data_dir="${NANOFUSE_DATA_DIR:-/tmp/nanofuse-macos}"
api_config="${NANOFUSE_CONFIG:-${api_data_dir}/nanofused.yaml}"
api_launchd_label="${NANOFUSE_API_LAUNCHD_LABEL:-com.daax.nanofuse.macos-api}"
api_bind="${NANOFUSE_API_BIND:-}"

usage() {
  cat <<'USAGE'
Usage: scripts/run-tray-macos.sh [--api-url URL] [--timeout DURATION] [--foreground] [--restart] [--smoke] [--start-api] [--debug]

Builds and launches the Nanofuse macOS menu bar app. With --start-api, it also
starts a local nanofused daemon using the macOS Apple container microVM runtime.

Options:
  --api-url URL    nanofused API URL. Default: NANOFUSE_TRAY_API_URL, NANOFUSE_API_URL, or http://127.0.0.1:18080
  --timeout VALUE  API request timeout, for example 2s or 500ms
  --foreground     keep the tray app attached to this terminal
  --restart        stop existing nanofuse-tray processes before launching
  --smoke          run an API smoke check and exit without starting the menu bar app
  --start-api      start local nanofused on macOS using Apple container / Virtualization.framework
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
    --start-api)
      start_api=1
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
if [[ "${start_api}" -eq 1 ]]; then
  echo "Building nanofused..."
  go build -o bin/nanofused ./cmd/nanofused
fi

write_api_config() {
  mkdir -p "${api_data_dir}"
  if [[ -z "${api_bind}" ]]; then
    local without_scheme
    without_scheme="${api_url#http://}"
    without_scheme="${without_scheme#https://}"
    api_bind="${without_scheme%%/*}"
    if [[ -z "${api_bind}" || "${api_bind}" == "${api_url}" ]]; then
      api_bind="127.0.0.1:18080"
    fi
  fi
  cat > "${api_config}" <<EOF
api:
  socket: ""
  tcp_bind: "${api_bind}"

storage:
  data_dir: "${api_data_dir}"
  database: "${api_data_dir}/nanofuse.db"

runtime:
  driver: apple_container
  apple_container:
    binary_path: "/usr/local/bin/container"
    auto_start: true
    default_command: "sleep infinity"

firecracker:
  binary_path: ""

limits:
  max_vms: 25
  max_total_memory_mib: 16384
  max_vcpus_per_vm: 8
  max_memory_per_vm_mib: 8192
  max_snapshot_storage_gib: 25

registry:
  auth_config_path: "${HOME}/.docker/config.json"
  pull_timeout_secs: 600
  layer_timeout_secs: 300

logging:
  level: debug
  format: text
  file_path: "${api_log_file}"
  console_log_max_size_mb: 10
  console_log_max_backups: 3

network:
  setup: false

spire:
  enabled: false

auth:
  enabled: false
EOF
}

wait_for_api() {
  local deadline
  deadline=$((SECONDS + 45))
  while (( SECONDS < deadline )); do
    if curl -fsS "${api_url}/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  return 1
}

if [[ "${start_api}" -eq 1 ]]; then
  write_api_config
  launchd_target="gui/$(id -u)/${api_launchd_label}"
  if launchctl print "${launchd_target}" >/dev/null 2>&1; then
    if [[ "${restart}" -eq 1 ]]; then
      echo "Stopping existing nanofused launchd service ${api_launchd_label}"
      launchctl bootout "${launchd_target}" >/dev/null 2>&1 || true
      sleep 1
    else
      echo "nanofused launchd service is already loaded: ${api_launchd_label}"
    fi
  fi
  if ! curl -fsS "${api_url}/health" >/dev/null 2>&1; then
    echo "Starting local nanofused with Apple container runtime. Endpoint: ${api_url}"
    abs_daemon="$(pwd)/bin/nanofused"
    launchctl submit -l "${api_launchd_label}" -o "${api_log_file}" -e "${api_log_file}" -- "${abs_daemon}" -config "${api_config}"
    if ! wait_for_api; then
      echo "nanofused did not become healthy. Log:" >&2
      sed -n '1,160p' "${api_log_file}" >&2 || true
      exit 1
    fi
    api_pid="$(launchctl print "${launchd_target}" 2>/dev/null | awk '/pid = / {print $3; exit}' || true)"
    if [[ -n "${api_pid}" ]]; then
      echo "${api_pid}" > "${api_pid_file}"
      echo "Started nanofused PID ${api_pid}. Log: ${api_log_file}"
    else
      echo "Started nanofused. Log: ${api_log_file}"
    fi
  fi
fi

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
