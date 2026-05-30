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
cd dev/vagrant
vagrant status
./closed-loop.sh
```

Expected behavior:

- Linux/KVM hosts proceed through Firecracker, daemon, and VM boot validation.
- macOS/Windows hosts proceed only if their Vagrant/VM provider exposes Linux KVM to the guest.
- Unsupported providers fail before VM boot with the exact missing capability.

## API Client Path

```bash
sudo ./bin/nanofused -config config.dev.yaml -tcp 127.0.0.1:8080
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/capabilities
NANOFUSE_API_URL=http://127.0.0.1:8080 nanofuse health
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
