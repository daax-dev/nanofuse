# CRITICAL: Non-Negotiable Development Requirements

**READ THIS SECTION FIRST** - These requirements are mandatory and supersede all other guidance.

## 1. Specification-Driven Development (MANDATORY)

**❌ PROHIBITED: Writing code without spec.md first**

**✅ REQUIRED:**

### Flowspec Workflow

This project uses **Flowspec** for specification-driven development. All features MUST follow the spec -> plan -> implement workflow.

#### Installation Verification
```bash
# Verify Flowspec is installed
flowspec --help

# If not installed, run:
uv tool install flowspec-cli --from git+https://github.com/jpoley/flowspec.git
```

#### Workflow Commands

**Flow Commands** (High-level workflow):
```bash
/flow:assess     # Classify work and validate readiness
/flow:specify    # Create feature specification from natural language
/flow:research   # Capture research and alternatives when required
/flow:plan       # Generate technical implementation plan
/flow:implement  # Execute implementation from tasks.md
/flow:validate   # QA, security, documentation, release management
/flow:operate    # Operational follow-through
```

**Spec Commands** (Detailed workflow steps):
```bash
/spec.flowspec      # Create or select a feature specification
/spec.clarify       # Resolve ambiguities (max 3 questions)
/spec.plan          # Generate technical implementation plan
/spec.tasks         # Break down into dependency-ordered tasks
/spec.implement     # Execute implementation from tasks.md
/spec.checklist     # Cross-artifact consistency validation
/spec.constitution  # Create/update project constitution
```

#### Mandatory Principles

1. **Specs are technology-agnostic**: spec.md contains WHAT and WHY only (no HOW)
2. **Maximum 3 clarifications**: Make informed guesses rather than over-marking `[NEEDS CLARIFICATION]`
3. **Success criteria must be measurable**: Every criterion must be verifiable
4. **No implementation leakage**: Never mention languages, frameworks, databases in spec.md
5. **Follow templates**: All specs follow `.flowspec/templates/spec-template.md`
6. **Human approval at phase boundaries**: AI executes, humans govern

#### Directory Structure
```
.flowspec/
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

**✅ REQUIRED: .flowspec/memory/constitution.md is the authoritative source**

### Constitutional Hierarchy

1. **Constitution** (`.flowspec/memory/constitution.md`) - SUPREME AUTHORITY
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
- [ ] Specification exists in `.flowspec/features/{branch}/spec.md`
- [ ] Plan exists in `.flowspec/features/{branch}/plan.md`
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
   /flow:specify <feature description>
   ```
   Creates branch, generates spec.md, validates quality

2. **Clarify ambiguities** (if needed, max 3):
   ```
   /spec.clarify
   ```

3. **Generate implementation plan**:
   ```
   /flow:plan
   ```
   Creates research.md, data-model.md, contracts/, quickstart.md

4. **Create tasks**:
   ```
   /spec.tasks
   ```
   Generates dependency-ordered tasks.md

5. **Add to backlog**:
   ```bash
   backlog task create "Implement feature: <name>" --label "Feature" --milestone "Sprint X"
   ```

6. **Implement**:
   ```
   /flow:implement
   ```
   Dispatches specialized engineer agents

7. **Validate**:
   ```
   /flow:validate
   ```
   QA, security review, documentation check

### Daily Development

```bash
# Morning: Review backlog
backlog task list --status "In Progress"

# Work on tasks using spec-driven approach
/flow:implement

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
| Specifications | `.flowspec/features/{branch}/spec.md` | Root, docs/ |
| Plans | `.flowspec/features/{branch}/plan.md` | Root, docs/ |
| Tasks | `backlog/tasks/` | Root, docs/ |
| Constitution | `.flowspec/memory/constitution.md` | Root |
| Templates | `.flowspec/templates/` | Root, docs/ |
| Scripts | `.flowspec/scripts/bash/` | Root, scripts/ |
| Documentation | `docs/` | Root, .flowspec/ |
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

1. **Consult the constitution**: `.flowspec/memory/constitution.md`
2. **Use sequential thinking**: Break down the problem
3. **Ask for clarification**: Use `/spec.clarify`
4. **Escalate to human**: Don't guess on critical decisions
5. **Document decisions**: Create ADR for architectural choices

## 11. NO STALE IMAGES OR ARTIFACTS (MANDATORY)

**❌ PROHIBITED: Using EOL, outdated, or unsupported base images**

**✅ REQUIRED: Always use current, supported versions**

### Image Freshness Rules

| Artifact Type | Requirement | Check Method |
|---------------|-------------|--------------|
| Base OS images | Must be **actively supported** (not EOL) | Check distro EOL dates |
| Kernels | Use **LTS versions** with active support | Check kernel.org LTS list |
| Container images | No images older than **6 months** without justification | Check image creation date |
| Test fixtures | Must use **same versions as production targets** | Compare versions |

### Firecracker Images Specifically

- **NEVER use** quickstart_guide images (stale since 2021)
- **ALWAYS use** official Firecracker CI images from `s3://spec.ccfc.min/firecracker-ci/`
- **Target OS**: Ubuntu 24.04 (not 18.04, 20.04, or 22.04)
- **Target Kernel**: 6.1.x LTS (support until 2028)

### Before Using Any Base Image

1. **Check EOL date**: Is the OS/kernel still supported?
2. **Check age**: When was this image built?
3. **Check source**: Is this from an official/trusted source?
4. **Verify versions**: Does this match our target stack?

### Why This Matters

- EOL images have **unpatched security vulnerabilities**
- Old images cause **mysterious failures** (dead repos, missing packages)
- Stale fixtures create **false confidence** in testing
- Version drift causes **production surprises**

### Current Approved Sources

```bash
# Firecracker CI images (updated regularly)
s3://spec.ccfc.min/firecracker-ci/{latest-version}/x86_64/

# Download script (checks for freshness)
./scripts/download-fixtures.sh
```

### Incident Record

| Date | Issue | Root Cause | Resolution |
|------|-------|------------|------------|
| 2025-12-28 | apt-daily failures, stuck VM | Ubuntu 18.04.5 (EOL 2023) from 2021 quickstart | Upgraded to Ubuntu 24.04 CI images |

---

**Remember**:
- 🚫 NO coding without specs
- 📋 ALL tasks in backlog
- 🧠 Complex decisions require sequential thinking
- 📜 Constitution is supreme authority
- ✅ Quality gates are non-negotiable
- 🗓️ NO stale/EOL images - always use current supported versions
# Claude Code Repository Rules

## CRITICAL: Repository Organization

### File Organization by Type

**NEVER mix file types in the wrong directories. This is non-negotiable.**

#### Scripts (`scripts/`)
- **ONLY executable files**: `.sh`, `.py`, etc.
- **NO markdown files**
- **NO documentation**
- **NO test data**
- Purpose: Things you RUN

#### Documentation (`docs/`)
- **User-facing documentation only**
- Installation guides
- Usage instructions
- API references
- **NO internal/work-in-progress docs**

#### Development Documentation (`docs/building/`)
- **Work-in-progress documentation**
- Implementation guides
- Architecture decisions
- Debugging guides
- Platform engineering analysis
- **This is where markdown files about development go**

#### Test Scripts and Data (`test/`)
- Test scripts
- Test fixtures
- Test data
- **NOT in scripts/**
- **NOT in examples/**
- **NOT scattered around root**

#### Examples (`examples/`)
- **Complete working examples**
- Each example in its own subdirectory
- Example-specific scripts stay in example directory
- **NOT a dumping ground for random test files**

### Clean Repository Discipline

1. **Before creating any file, ask: "Where does this type of file belong?"**
2. **Script = scripts/**
3. **User docs = docs/**
4. **Dev docs = docs/building/**
5. **Tests = test/**
6. **Examples = examples/{example-name}/**

### What NOT to Do

❌ **NEVER** put markdown files in `scripts/`
❌ **NEVER** scatter test scripts in root directory
❌ **NEVER** put random `.sh` files in examples unless example-specific
❌ **NEVER** put work-in-progress docs in `docs/` (use `docs/building/`)
❌ **NEVER** create temporary files without cleanup plan

### Cleanup Protocol

When you notice files in wrong locations:
1. **STOP** whatever you're doing
2. **MOVE** files to correct locations
3. **UPDATE** any references
4. **VERIFY** repo structure is clean
5. **THEN** continue with original task

### Directory Structure Reference

```
nanofuse/
├── scripts/           # Executable scripts ONLY
│   ├── *.sh          # Build, deploy, utility scripts
│   └── NO .md files
├── docs/             # User-facing documentation
│   ├── *.md          # User guides, API docs
│   └── building/     # Development documentation
│       └── *.md      # Architecture, debugging, analysis
├── test/             # All test-related files
│   ├── *.go          # Test code
│   ├── *.sh          # Test scripts
│   └── fixtures/     # Test data
├── examples/         # Complete examples
│   └── {name}/       # Each example self-contained
└── internal/         # Implementation code
    └── */            # Package structure
```

## Principle

**"A place for everything, and everything in its place."**

If you're not sure where something goes, **ASK** before creating it.
