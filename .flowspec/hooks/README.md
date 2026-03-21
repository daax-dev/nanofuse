# Hooks Directory

This directory contains custom hooks for the Flowspec workflow.

## Files

- `hooks.yaml` - Hook configuration file
- `*.sh` - Hook scripts (must be executable)
- `audit.log` - Execution audit log (auto-created)

## Enabled by Default

When hooks are scaffolded (without `--no-hooks`), these 3 hooks are **enabled**:

| Hook | Event | Description |
|------|-------|-------------|
| `run-tests` | `implement.completed` | Runs test suite after implementation |
| `lint-code` | `task.completed` | Runs linter and formatter on task completion |
| `quality-gate` | `validate.started` | Checks code quality before validation |

**Not enabled by default**: The `update-changelog` hook is opt-in (set `enabled: true` to use it).

To disable all hooks during init, use: `flowspec init --no-hooks`

## Getting Started

1. **Review configured hooks** in `hooks.yaml`
2. **Customize scripts** for your project's tooling
3. **Test a hook**:
   ```bash
   flowspec hooks test run-tests implement.completed
   ```
4. **Validate configuration**:
   ```bash
   flowspec hooks validate
   ```

## Available Events

- `spec.created`, `spec.updated`
- `plan.created`, `plan.updated`
- `task.created`, `task.completed`
- `implement.started`, `implement.completed`
- `validate.started`, `validate.completed`
- `deploy.started`, `deploy.completed`

## CLI Commands

```bash
# List configured hooks
flowspec hooks list

# Emit an event manually
flowspec hooks emit spec.created --spec-id my-feature

# View execution history
flowspec hooks audit --tail 20

# Test a hook
flowspec hooks test <hook-name> <event-type>

# Validate configuration
flowspec hooks validate
```

## Security

- Scripts must be in `.flowspec/hooks/` directory
- Timeout enforced (default: 30s)
- All executions are audit logged
- Scripts run in sandboxed environment

## Documentation

Full documentation: https://github.com/jpoley/flowspec/blob/main/docs/guides/hooks.md
