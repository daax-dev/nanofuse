#!/bin/bash
# Create backlog tasks from Product Requirements Analysis
# Based on SVPG Product Operating Model

set -e

echo "Creating NanoFuse backlog tasks from Product Requirements Analysis..."
echo ""

# Check if backlog command exists
if ! command -v backlog &> /dev/null; then
    echo "❌ Error: backlog command not found"
    echo "Install with: npm install -g @taskboard/backlog-cli"
    exit 1
fi

# Epic 1: Core Functionality Validation (MUST HAVE)
echo "Creating EPIC 1: Core Functionality Validation..."
backlog task create \
    "EPIC 1: Core Functionality Validation" \
    --status "To Do" \
    --description "Outcome: Developer can deploy a working VM reliably

Success Criteria:
- Services (nginx, todo-backend) start automatically on boot
- Services accessible from host (ports 80, 8080)
- 5/5 fresh VM deployments succeed without manual intervention

Time Box: 3 days maximum

Priority: P0 (MUST HAVE)

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6" \
    --labels "Epic,P0,Phase1"

# Task 1.1: Diagnose Service Startup Failure
echo "Creating Task 1.1: Diagnose Service Startup Failure..."
backlog task create \
    "Task 1.1: Diagnose Service Startup Failure" \
    --status "To Do" \
    --description "Epic: EPIC 1 - Core Functionality Validation

Objective: Understand why nginx and todo-backend fail to start in VM

Acceptance Criteria:
- Console log analyzed systematically (Layer 0-6 from testing plan)
- Root cause identified with evidence
- Hypothesis documented in backlog/decisions/
- Fix approach defined

Success Metrics:
- Root cause identified within 4 hours
- Fix approach has > 80% confidence of success

Estimated Effort: 4 hours
Priority: P0
Prerequisites: None

Commands to run:
sudo tail -200 /var/lib/nanofuse/vms/<VM_ID>/console.log
sudo grep 'systemd\[1\]' /var/lib/nanofuse/vms/<VM_ID>/console.log | head -20
sudo grep 'nginx.service' /var/lib/nanofuse/vms/<VM_ID>/console.log
sudo grep '\[FAILED\]' /var/lib/nanofuse/vms/<VM_ID>/console.log" \
    --labels "Task,P0,Phase1,Diagnosis"

# Task 1.2: Fix Service Startup
echo "Creating Task 1.2: Fix Service Startup..."
backlog task create \
    "Task 1.2: Fix Service Startup" \
    --status "To Do" \
    --description "Epic: EPIC 1 - Core Functionality Validation

Objective: Implement fix to get nginx and todo-backend running

Acceptance Criteria:
- Fix implemented (init parameter or service config)
- VM rebuilt with fix
- Services start successfully on boot
- Services accessible from host (ports 80, 8080)
- Fix documented in code and docs

Success Metrics:
- curl http://172.16.0.10:80 returns HTML (nginx)
- curl http://172.16.0.10:8080/health returns {\"status\":\"healthy\"}
- 5/5 fresh VM deployments succeed

Estimated Effort: 2-4 hours (after root cause known)
Priority: P0
Prerequisites: Task 1.1 complete" \
    --labels "Task,P0,Phase1,Implementation"

# Task 1.3: Create VM Health Check Script
echo "Creating Task 1.3: Create VM Health Check Script..."
backlog task create \
    "Task 1.3: Create VM Health Check Script" \
    --status "To Do" \
    --description "Epic: EPIC 1 - Core Functionality Validation

Objective: Automated validation that VMs are fully functional

Acceptance Criteria:
- Script checks VM boot status
- Script validates network connectivity
- Script checks service health endpoints
- Script validates performance (boot time, latency)
- Script outputs pass/fail with details

Success Metrics:
- Script runs in < 10 seconds
- Detects all known failure modes
- Exit code 0 = success, 1 = failure

Output: scripts/building/health-check.sh

Estimated Effort: 2 hours
Priority: P0
Prerequisites: Task 1.2 complete" \
    --labels "Task,P0,Phase1,Automation"

# Epic 2: End-to-End Workflow Validation (MUST HAVE)
echo "Creating EPIC 2: E2E Workflow Validation..."
backlog task create \
    "EPIC 2: End-to-End Workflow Validation" \
    --status "To Do" \
    --description "Outcome: Complete pull-to-running cycle works reliably

Success Criteria:
- E2E test script runs pull → run → verify → cleanup
- Test passes with default base image
- Test passes with todo-app example
- Test passes 10/10 times on clean system

Time Box: 2 days

Priority: P0 (MUST HAVE)

Prerequisite: EPIC 1 complete

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6" \
    --labels "Epic,P0,Phase1"

# Task 2.1: Create End-to-End Test Script
echo "Creating Task 2.1: Create E2E Test Script..."
backlog task create \
    "Task 2.1: Create End-to-End Test Script" \
    --status "To Do" \
    --description "Epic: EPIC 2 - E2E Workflow Validation

Objective: Automated test for complete pull-to-running workflow

Acceptance Criteria:
- Script authenticates to GHCR
- Script pulls default image
- Script creates and starts VM
- Script validates services running
- Script cleans up VM and image
- Script is idempotent (can run repeatedly)

Success Metrics:
- Script completes in < 90 seconds
- Passes 10/10 runs on clean system
- Catches regressions automatically

Output: test/e2e/full-workflow-test.sh

Estimated Effort: 4 hours
Priority: P0
Prerequisites: EPIC 1 complete" \
    --labels "Task,P0,Phase1,Testing"

# Task 2.2: Document Common Failure Modes
echo "Creating Task 2.2: Document Common Failure Modes..."
backlog task create \
    "Task 2.2: Document Common Failure Modes" \
    --status "To Do" \
    --description "Epic: EPIC 2 - E2E Workflow Validation

Objective: Comprehensive troubleshooting guide for users

Acceptance Criteria:
- Document all failures encountered during testing
- Provide diagnostic commands for each failure
- Provide fix steps for each failure
- Include prevention strategies
- Link to relevant logs/configs

Success Metrics:
- Covers 90% of issues encountered
- Users can self-diagnose without asking

Output: docs/TROUBLESHOOTING.md

Estimated Effort: 3 hours
Priority: P0
Prerequisites: EPIC 2 testing complete" \
    --labels "Task,P0,Phase1,Documentation"

# Epic 3: Usability Improvements (SHOULD HAVE)
echo "Creating EPIC 3: Usability Improvements..."
backlog task create \
    "EPIC 3: Usability Improvements" \
    --status "To Do" \
    --description "Outcome: Error messages are actionable, logs are accessible

Success Criteria:
- nanofuse vm logs command implemented
- Error messages include next steps
- Fresh user can diagnose failures without help

Time Box: 2 days

Priority: P1 (SHOULD HAVE)

Prerequisite: EPICs 1 & 2 complete

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6" \
    --labels "Epic,P1,Phase2"

# Task 3.1: Add VM Logs Command
echo "Creating Task 3.1: Add VM Logs Command..."
backlog task create \
    "Task 3.1: Add VM Logs Command" \
    --status "To Do" \
    --description "Epic: EPIC 3 - Usability Improvements

Objective: Make console logs accessible without sudo

Acceptance Criteria:
- CLI command 'nanofuse vm logs <name>' implemented
- Supports --follow for tail -f behavior
- Supports --lines N for limiting output
- Handles permission errors gracefully
- Works with both running and stopped VMs

Success Metrics:
- User can read logs without sudo access
- Command feels familiar to Docker users

Estimated Effort: 3 hours
Priority: P1
Prerequisites: EPICs 1 & 2 complete" \
    --labels "Task,P1,Phase2,Feature"

# Task 3.2: Improve Error Messages
echo "Creating Task 3.2: Improve Error Messages..."
backlog task create \
    "Task 3.2: Improve Error Messages" \
    --status "To Do" \
    --description "Epic: EPIC 3 - Usability Improvements

Objective: Actionable error messages with next steps

Acceptance Criteria:
- All errors include suggested fix
- Errors reference documentation
- Errors include relevant context (VM ID, image, etc.)
- Errors formatted for readability
- Errors logged with correlation IDs

Success Metrics:
- 80% of errors self-diagnosable
- Average resolution time < 10 minutes

Estimated Effort: 4 hours
Priority: P1
Prerequisites: EPICs 1 & 2 complete" \
    --labels "Task,P1,Phase2,UX"

# Epic 4: Snapshot/Resume Validation (SHOULD HAVE)
echo "Creating EPIC 4: Snapshot/Resume Validation..."
backlog task create \
    "EPIC 4: Snapshot/Resume Validation" \
    --status "To Do" \
    --description "Outcome: VMs can be paused and resumed reliably

Success Criteria:
- Snapshot/resume works 9/10 times
- Resume time < 500ms
- State preservation validated

Time Box: 5 days (includes research)

Priority: P1 (SHOULD HAVE)

Prerequisite: EPICs 1 & 2 complete

IMPORTANT: Must research feasibility BEFORE implementation

See: docs/building/PRODUCT_REQUIREMENTS_ANALYSIS.md Section 6" \
    --labels "Epic,P1,Phase2"

# Task 4.1: Research Snapshot/Resume Feasibility
echo "Creating Task 4.1: Research Snapshot/Resume Feasibility..."
backlog task create \
    "Task 4.1: Research Snapshot/Resume Feasibility" \
    --status "To Do" \
    --description "Epic: EPIC 4 - Snapshot/Resume Validation

Objective: Validate technical feasibility before implementation

Acceptance Criteria:
- Firecracker snapshot docs reviewed
- Systemd compatibility researched
- Limitations documented
- Proof of concept approach defined
- Go/no-go decision made

Success Metrics:
- Clear understanding of constraints
- Confidence level > 70% in approach

Output: docs/building/snapshot-resume-research.md

Estimated Effort: 1 day
Priority: P1
Prerequisites: EPICs 1 & 2 complete and stable" \
    --labels "Task,P1,Phase2,Research"

# Task 4.2: Implement Basic Snapshot
echo "Creating Task 4.2: Implement Basic Snapshot..."
backlog task create \
    "Task 4.2: Implement Basic Snapshot" \
    --status "To Do" \
    --description "Epic: EPIC 4 - Snapshot/Resume Validation

Objective: Minimal viable snapshot/resume functionality

Acceptance Criteria:
- 'nanofuse vm pause <name>' creates snapshot
- 'nanofuse vm resume <name>' restores from snapshot
- VM state preserved (processes, network, files)
- Resume time < 500ms
- Snapshot/resume succeeds 9/10 times

Success Metrics:
- Resume time < 500ms
- State preservation validated
- Reliability > 90%

Estimated Effort: 3 days
Priority: P1
Prerequisites: Task 4.1 complete with GO decision" \
    --labels "Task,P1,Phase2,Feature"

echo ""
echo "✅ All backlog tasks created successfully!"
echo ""
echo "View tasks with: backlog task list"
echo "Or open browser with: backlog browser"
echo ""
echo "Tasks organized by:"
echo "  - Priority: P0 (MUST HAVE), P1 (SHOULD HAVE)"
echo "  - Phase: Phase1, Phase2"
echo "  - Type: Epic, Task, Feature, Documentation, etc."
echo ""
echo "Next steps:"
echo "1. Review tasks in backlog browser"
echo "2. Start with Task 1.1 (Diagnose Service Startup)"
echo "3. See: docs/building/NEXT_STEPS_START_HERE.md"
