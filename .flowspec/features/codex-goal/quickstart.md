# Sandbox Objective Validation Quickstart

Run from the repository root unless noted.

## Local Gates

```bash
go test ./internal/api ./internal/network ./internal/types
go test ./cmd/nanofuse ./internal/client
go fmt ./...
mage ci
```

## Vagrant Closed Loop

```bash
gh repo clone daax-dev/vagrant-skill /tmp/daax-vagrant-skill-readonly
cd /tmp/daax-vagrant-skill-readonly
VM_NAME=nanofuse-vagrant-skill-test PROJECT_SRC=/path/to/nanofuse VM_CPUS=4 VM_MEMORY=8192 vagrant up --provider=parallels
VM_NAME=nanofuse-vagrant-skill-test PROJECT_SRC=/path/to/nanofuse vagrant ssh -c 'cd /project && ./scripts/ensure-mage.sh && mage ci'
```

Expected behavior:

- Linux/KVM hosts proceed through Firecracker, daemon, and VM boot validation.
- macOS Vagrant/Linux guest validation proceeds only if the VM provider exposes Linux KVM to the guest.
- Unsupported providers fail before VM boot with the exact missing capability.

`dev/vagrant` remains a secondary repo-local harness. The required harness for this objective is `daax-dev/vagrant-skill`.

## Linux/KVM API Path

```bash
sudo ./bin/nanofused -config config.dev.yaml -tcp 127.0.0.1:8080
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/capabilities
NANOFUSE_API_URL=http://127.0.0.1:8080 nanofuse health
```

## macOS Local Runtime Path

macOS:

```bash
./scripts/run-tray-macos.sh --start-api --restart
./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 5s
curl http://127.0.0.1:18080/capabilities
```

Create and start a local Apple-container VM:

```bash
API=http://127.0.0.1:18080
VM_ID="$(curl -fsS -X POST "$API/vms" \
  -H "Content-Type: application/json" \
  -d '{"name":"mac-api-alpine","image":"alpine:3.20","config":{"vcpus":1,"memory_mib":256,"network":{"mode":"none"}}}' \
  | jq -r '.vm.id')"
curl -fsS -X POST "$API/vms/$VM_ID/start"
CONTAINER_ID="$(curl -fsS "$API/vms/$VM_ID" | jq -r '.vm.runtime.external_id')"
container exec "$CONTAINER_ID" uname -a
curl -fsS -X POST "$API/vms/$VM_ID/stop" -H "Content-Type: application/json" -d '{"timeout_seconds":10}'
curl -fsS -X DELETE "$API/vms/$VM_ID"
```

Expected runtime driver: `apple_container`.

## Windows Client Path

Windows:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-tray-windows.ps1 -ApiUrl "$env:NANOFUSE_API_URL"
```

Smoke mode:

```bash
go build -o bin/nanofuse-tray ./cmd/nanofuse-tray
./bin/nanofuse-tray --smoke --api-url "${NANOFUSE_API_URL:-http://127.0.0.1:18080}"
```

For Vagrant:

```bash
cd dev/vagrant
NANOFUSE_API_HOST_PORT=18080 vagrant up
vagrant ssh -c "sudo systemctl start nanofused"
curl http://127.0.0.1:18080/health
```

## Evidence

Record command output in:

- Backlog task `TASK-47` notes.
- `.logs/decisions/sandbox-objective.jsonl` for non-trivial decisions.
- `.logs/references/sandbox-objective.jsonl` for citable external claims.
