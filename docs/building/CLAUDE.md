# CRITICAL: Non-Negotiable Development Requirements

**READ THIS SECTION FIRST** - These requirements are mandatory and supersede all other guidance.

## 1. Specification-Driven Development (MANDATORY)

**❌ PROHIBITED: Writing code without spec.md first**

**✅ REQUIRED:**

### jp-spec-kit Workflow

This project uses **jp-spec-kit** for specification-driven development. All features MUST follow the spec → plan → implement workflow.

#### Installation Verification
```bash
# Verify jp-spec-kit is installed
specify --help

# If not installed, run:
uv tool install specify-cli --from git+https://github.com/jpoley/jp-spec-kit.git
```

#### Workflow Commands

**JPSpec Commands** (High-level multi-agent workflows):
```bash
/jpspec.specify   # PM Planner agent creates specifications
/jpspec.plan      # Architect & Platform Engineer design artifacts
/jpspec.research  # Market and technical research + business validation
/jpspec.implement # Frontend/Backend engineers with code review
/jpspec.validate  # QA, security, documentation, release management
/jpspec.operate   # SRE agents for CI/CD, DevSecOps, resilience
```

**SpecKit Commands** (Detailed workflow steps):
```bash
/speckit.specify      # Create feature specification from natural language
/speckit.clarify      # Resolve ambiguities (max 3 questions)
/speckit.plan         # Generate technical implementation plan
/speckit.tasks        # Break down into dependency-ordered tasks
/speckit.implement    # Execute implementation from tasks.md
/speckit.analyze      # Cross-artifact consistency validation
/speckit.constitution # Create/update project constitution
/speckit.taskstoissues # Convert tasks to GitHub issues
```

#### Mandatory Principles

1. **Specs are technology-agnostic**: spec.md contains WHAT and WHY only (no HOW)
2. **Maximum 3 clarifications**: Make informed guesses rather than over-marking `[NEEDS CLARIFICATION]`
3. **Success criteria must be measurable**: Every criterion must be verifiable
4. **No implementation leakage**: Never mention languages, frameworks, databases in spec.md
5. **Follow templates**: All specs follow `.specify/templates/spec-template.md`
6. **Human approval at phase boundaries**: AI executes, humans govern

#### Directory Structure
```
.specify/
├── features/{branch}/    # Feature-specific artifacts
│   ├── spec.md          # Technology-agnostic specification
│   ├── plan.md          # Technical implementation plan
│   ├── tasks.md         # Dependency-ordered tasks
│   ├── research.md      # Research findings
│   ├── data-model.md    # Entity models
│   └── contracts/       # API contracts (OpenAPI, etc.)
├── templates/           # Standardized templates
├── scripts/bash/        # Workflow automation
└── memory/
    └── constitution.md  # Project principles (SUPREME AUTHORITY)
```

## 2. Backlog.md Task Management (MANDATORY)

**✅ REQUIRED: ALL tasks tracked in backlog/**

### Backlog MCP Server

This project uses **Backlog.md** for task management, integrated via MCP server.

#### Initialization Verification
```bash
# Verify backlog is initialized (should show backlog/ directory)
ls -la backlog/

# If not initialized:
backlog init "<Project Name>" --defaults --integration-mode mcp
```

#### Backlog Structure
```
backlog/
├── tasks/        # Active tasks
├── completed/    # Completed tasks
├── drafts/       # Draft tasks
├── archive/      # Archived tasks
├── decisions/    # Decision records
├── docs/         # Task documentation
└── config.yml    # Backlog configuration
```

#### Using Backlog

**Via MCP Server** (recommended):
- Access backlog resources via `backlog://` URI scheme
- MCP server provides task management integration
- Tasks automatically synced across AI agent sessions

**Via CLI**:
```bash
backlog task create "Task description"           # Add task
backlog task list                             # List tasks
backlog task archive <task-id>               # Mark complete
backlog browser                              # Open web UI
```

#### Mandatory Practices

1. **Create tasks BEFORE implementation**: No coding without task tracking
2. **Link tasks to specifications**: Reference spec.md in task descriptions
3. **Update task status**: Keep tasks current (Backlog → In Progress → Done)
4. **Use labels and milestones**: Organize work with metadata
5. **Track decisions**: Document key decisions in backlog/decisions/

## 3. Sequential Thinking for Complex Decisions (MANDATORY)

**✅ REQUIRED: Use mcp__sequential__sequentialthinking for multi-step analysis**

### When to Use Sequential Thinking

Use the `mcp__sequential__sequentialthinking` tool when:
- Making architectural decisions
- Analyzing complex problems
- Planning multi-step implementations
- Evaluating trade-offs
- Designing system architecture
- Investigating bugs or issues

### Example Usage

```
# In Claude Code, use the tool:
mcp__sequential__sequentialthinking with:
- thought: "Current analysis step..."
- thoughtNumber: 1
- totalThoughts: 10
- nextThoughtNeeded: true
```

### Mandatory Practices

1. **Show your work**: Break down complex problems into steps
2. **Express uncertainty**: Use is_revision and branch_from_thought when reconsidering
3. **Adjust estimates**: Update totalThoughts as understanding evolves
4. **Generate hypotheses**: Propose solutions and verify them
5. **Verify before concluding**: Only set nextThoughtNeeded=false when truly done

## 4. Constitution as Supreme Authority (MANDATORY)

**✅ REQUIRED: .specify/memory/constitution.md is the authoritative source**

### Constitutional Hierarchy

1. **Constitution** (`.specify/memory/constitution.md`) - SUPREME AUTHORITY
2. **CLAUDE.md** (this file) - Quick reference and workflows
3. **README.md** - User-facing documentation
4. **Other docs** - Supporting guidance

### When Conflicts Arise

- Constitution ALWAYS takes precedence
- If guidance conflicts with constitution, follow constitution
- If constitution is unclear, escalate to human authority
- Document decisions in ADRs (Architecture Decision Records)

### Key Constitutional Principles

Most projects follow these core principles:
1. **Production-Grade Quality**: No shortcuts, quality gates are immutable
2. **Test-Driven Development**: Tests before implementation (Red-Green-Refactor)
3. **Security by Design**: Threat modeling in specification phase
4. **Observability First**: Metrics, logs, traces from day one
5. **Immutable Artifacts**: Build once, deploy everywhere
6. **Documentation as Code**: ADRs for decisions, docs version-controlled

**Read your project's constitution before starting any significant work.**

## 5. Quality Gates Checklist

Before committing code:
- [ ] Specification exists in `.specify/features/{branch}/spec.md`
- [ ] Plan exists in `.specify/features/{branch}/plan.md`
- [ ] Tasks tracked in `backlog/tasks/`
- [ ] Tests written BEFORE implementation (TDD)
- [ ] All tests passing
- [ ] Code follows project conventions
- [ ] No secrets in code
- [ ] Documentation updated
- [ ] Constitution compliance verified

## 6. Common Workflows

### Starting a New Feature

1. **Create specification**:
   ```
   /jpspec.specify <feature description>
   ```
   Creates branch, generates spec.md, validates quality

2. **Clarify ambiguities** (if needed, max 3):
   ```
   /speckit.clarify
   ```

3. **Generate implementation plan**:
   ```
   /jpspec.plan
   ```
   Creates research.md, data-model.md, contracts/, quickstart.md

4. **Create tasks**:
   ```
   /speckit.tasks
   ```
   Generates dependency-ordered tasks.md

5. **Add to backlog**:
   ```bash
   backlog task create "Implement feature: <name>" --label "Feature" --milestone "Sprint X"
   ```

6. **Implement**:
   ```
   /jpspec.implement
   ```
   Dispatches specialized engineer agents

7. **Validate**:
   ```
   /jpspec.validate
   ```
   QA, security review, documentation check

### Daily Development

```bash
# Morning: Review backlog
backlog task list --status "In Progress"

# Work on tasks using spec-driven approach
/speckit.implement

# Track progress
backlog task edit <task-id> --status "In Progress"

# Complete tasks
backlog task archive <task-id>

# End of day: Check status
backlog browser
```

## 7. File Organization Rules

**CRITICAL: Files MUST be in correct locations**

| File Type | Correct Location | WRONG Location |
|-----------|-----------------|----------------|
| Specifications | `.specify/features/{branch}/spec.md` | Root, docs/ |
| Plans | `.specify/features/{branch}/plan.md` | Root, docs/ |
| Tasks | `backlog/tasks/` | Root, docs/ |
| Constitution | `.specify/memory/constitution.md` | Root |
| Templates | `.specify/templates/` | Root, docs/ |
| Scripts | `.specify/scripts/bash/` | Root, scripts/ |
| Documentation | `docs/` | Root, .specify/ |
| WIP docs | `docs/building/` | Root, docs/ |

**No exceptions. Maintain repository cleanliness.**

## 8. Error Handling Requirements

**All code MUST handle errors explicitly:**

- ❌ No silent failures
- ❌ No ignored error returns
- ❌ No generic error messages
- ✅ Log all errors with context
- ✅ Return actionable error messages
- ✅ Fail fast on critical errors
- ✅ Validate all inputs

## 9. Testing Requirements

**TDD cycle strictly enforced:**

1. Write test (RED)
2. Get approval
3. Run test (FAILS)
4. Implement (GREEN)
5. Tests pass
6. Refactor
7. Repeat

**No code merges without:**
- Passing unit tests
- Passing integration tests
- Passing contract tests (for APIs)
- Test coverage ≥ 80% for new code

## 10. When in Doubt

1. **Consult the constitution**: `.specify/memory/constitution.md`
2. **Use sequential thinking**: Break down the problem
3. **Ask for clarification**: Use `/speckit.clarify`
4. **Escalate to human**: Don't guess on critical decisions
5. **Document decisions**: Create ADR for architectural choices

---

**Remember**:
- 🚫 NO coding without specs
- 📋 ALL tasks in backlog
- 🧠 Complex decisions require sequential thinking
- 📜 Constitution is supreme authority
- ✅ Quality gates are non-negotiable
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NanoFuse is an open-source Firecracker-based microVM platform for running untrusted code in secure, isolated sandboxes. Similar to E2B but self-hosted.

**Primary Use Cases** (in priority order):
1. **AI Code Execution Sandbox** - Run LLM-generated code securely with sub-200ms boot times
2. **General Isolated Workloads** - Ephemeral compute for any untrusted or multi-tenant workload
3. **Dev Environment VMs** - Fast-spinning development sandboxes

**Current Status**: Core infrastructure ~60% complete. Pivoting focus from Trigger.dev to AI sandbox use case.

## Architecture

The system consists of four main components:

1. **CLI Tool (`nanofuse`)**: Command-line interface for VM and template management
   - Subcommands: `pull`, `up`, `stop`, `resume`, `status`
   - Go-based, static binary
   - GHCR authentication support
2. **API Daemon (`nanofused`)**: Systemd service managing Firecracker processes
   - Full Firecracker VM lifecycle management
   - **Snapshot/resume support** (required for <200ms boot times)
   - REST API + gRPC for real-time operations
   - Go-based service
3. **In-VM Daemon (`nanofuse-envd`)**: Runs inside each sandbox (like E2B's envd)
   - Provides consistent SDK interaction interface
   - Handles filesystem, process, and terminal operations
   - HTTP/gRPC on well-known port
4. **MicroVM Templates**: Snapshotted VM images for instant boot
   - Base template: Ubuntu 24.04 + systemd + nanofuse-envd
   - Custom templates built from Dockerfiles
   - OverlayFS for efficient multi-instance storage

## Key Design Decisions

- **Base OS**: Ubuntu 24.04 for a minimal base image (FROM ubuntu:24.04 in Dockerfile)
- **Build System**: GitHub Actions (with Make where needed)
- **Image Distribution**: Private GitHub artifact registry (GHCR, similar to Slicer's model)
- **Image Strategy**: Build FROM ubuntu:24.04 with systemd, following Slicer's Dockerfile patterns for learning purposes
  - Start FROM ubuntu:24.04 base container image
  - Install systemd, openssh-server, networking packages
  - Bundle Slicer's proven 5.10.240 kernel in the image
  - Add packages and systemd units in Dockerfile
  - Enable (not start) systemd services during build (critical: use `systemctl enable` not `systemctl start`)
  - Use one-shot systemd units for first-boot initialization
  - This approach provides learning value (building from scratch) while being practical (using proven kernel + Dockerfile pattern)

## Implementation Strategy

Based on E2B architecture analysis (see `docs/building/planning/e2b-learnings.md`):

**Phase 1: Core Infrastructure** (current)
- CLI and API daemon for VM lifecycle
- Basic Firecracker integration
- TAP networking with IPAM

**Phase 2: Fast Boot** (next priority)
- OverlayFS for shared rootfs across instances
- Snapshot/resume for <200ms boot times
- Template system (Dockerfile → snapshot conversion)

**Phase 3: SDK & Daemon**
- In-VM daemon (nanofuse-envd) for SDK interaction
- Python SDK for AI agent integration
- JavaScript SDK
- Code execution API

**Phase 4: Advanced Features**
- Jupyter kernel support for code interpreter
- Multi-language execution contexts
- Pause/resume with state preservation

**Important Build Constraints**:
- Do NOT use `systemctl start` during Docker build
- Use `systemctl enable` for services
- Add one-shot systemd units for first-boot tasks
- Keep CMD unchanged (expects systemd)
- Set `DEBIAN_FRONTEND=noninteractive` for apt operations

## Development Notes

### When Building Images

- Images must be arch-specific: tag appropriately for x86_64 vs arm64
- Kernel cmdline for Firecracker (if building custom kernel): `console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k`
- Console must be on ttyS0 for Firecracker compatibility
- Test kernels thoroughly - newer kernels can have `/dev/vda not found` errors
- For GPU/PCI passthrough, Cloud Hypervisor is required (not Firecracker)

### CI/CD Pipeline

- GitHub Actions workflow should build on push
- Publish to GHCR with tags like `:latest` and `:sha-<short>` for reproducibility
- Support both x86_64 and arm64 architectures

## Future Development Commands

Once implemented, common commands will likely include:

- Build commands (TBD - likely `make build` or Docker build commands)
- Test commands (TBD)
- CLI commands for VM management (TBD)
- API service management via systemd (TBD)

## References

- E2B (primary inspiration): https://github.com/e2b-dev - Architecture analysis in `docs/building/planning/e2b-learnings.md`
- Firecracker: https://github.com/firecracker-microvm/firecracker
- Slicer (image patterns): https://docs.slicervm.com
