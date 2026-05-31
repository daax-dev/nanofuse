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
- macOS/Windows hosts proceed only if their Vagrant/VM provider exposes Linux KVM to the guest.
- Unsupported providers fail before VM boot with the exact missing capability.

`dev/vagrant` remains a secondary repo-local harness. The required harness for this objective is `daax-dev/vagrant-skill`.

## API Client Path

```bash
sudo ./bin/nanofused -config config.dev.yaml -tcp 127.0.0.1:8080
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/capabilities
NANOFUSE_API_URL=http://127.0.0.1:8080 nanofuse health
```

## Tray Client Path

macOS:

```bash
NANOFUSE_API_URL="${NANOFUSE_API_URL:-http://127.0.0.1:18080}" ./scripts/run-tray-macos.sh
```

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
