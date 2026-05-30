# Sandbox Objective Validation Quickstart

Run from the repository root unless noted.

## Local Gates

```bash
go test ./internal/api ./internal/network ./internal/types
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

## Evidence

Record command output in:

- Backlog task `TASK-47` notes.
- `.logs/decisions/sandbox-objective.jsonl` for non-trivial decisions.
- `.logs/references/sandbox-objective.jsonl` for citable external claims.
