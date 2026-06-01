---
id: TASK-51
title: 'Expose VM ports, exec access, and image launch workflows'
status: In Progress
assignee: []
created_date: '2026-05-31 23:06'
updated_date: '2026-05-31 23:18'
labels:
  - codex-goal
  - operator-gap
  - api
  - tray
dependencies: []
references:
  - .flowspec/features/codex-goal/spec.md
priority: high
---

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Operators can see configured VM port forwards without reading raw JSON.
- [x] #2 Operators can run a command inside an Apple-container-backed running VM through the Nanofuse API/CLI boundary.
- [x] #3 The tray VM list shows multiple VMs with state and port context.
- [x] #4 Documentation explains the easy process for enabling more launchable images.
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add first-party exec API/client/CLI support through nanofused for runtime backends that can execute commands.
2. Surface configured/published ports in VM list/status and tray VM rows.
3. Document exact local commands for ports, exec/SSH, multiple VMs, and image enablement.
4. Validate with focused unit tests, mage ci, and macOS API/tray smoke.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Validation completed on 2026-05-31:
- go test ./internal/client ./internal/api ./internal/applecontainer ./cmd/nanofuse ./cmd/nanofuse-tray
- git diff --check
- ./scripts/ensure-mage.sh && mage ci
- ./scripts/run-tray-macos.sh --start-api --restart --smoke --timeout 10s --api-url http://127.0.0.1:18080
- macOS Apple-container closed loop with two concurrent VMs, distinct port forwards 19191/19192, nanofuse vm ports, nanofuse vm exec uname/os-release, delete cleanup to remaining=0
- tray launchd process running against http://127.0.0.1:18080 with successful health/capabilities/vms/images refresh logs
<!-- SECTION:NOTES:END -->
