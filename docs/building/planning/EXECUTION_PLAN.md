# NanoFuse Execution Plan (Corrected)

Date: 2025-10-30 (Corrected after validation)

This is a pragmatic, iterative plan to ship a working Firecracker-based microVM foundation built on Ubuntu 24.04 that can "pull and run like Slicer," serving as a complete rebuild for learning purposes while delivering a production-ready foundation for Trigger.dev dual-environment workloads.

## Goals

- Build and publish a minimal Ubuntu 24.04 systemd-based microVM image to GHCR, consumable by a Slicer-like workflow.
- Provide a full-featured CLI and API to pull images and start/stop/resume VMs with snapshot support.
- Automate builds with GitHub Actions and tag images for x86_64 (arm64 to follow).
- Extend base image for Trigger.dev web and worker components.

## Scope

- VM Image: Build FROM ubuntu:24.04 with systemd, following Slicer's Dockerfile patterns (systemctl enable not start).
- CLI: Go-based, static-friendly, no external deps. Subcommands: image pull, vm up, vm stop, vm resume, vm status.
- API: Go HTTP service managed by systemd, exposing start/stop/resume/status with full Firecracker snapshot support.
- CI: Build Go binaries and Docker image, publish to GHCR when secrets exist.

## Assumptions

- Host: Linux with KVM (/dev/kvm), cgroups v2, root or suitable caps to configure TAP.
- Registry: GHCR private available; PATs with read:packages (hosts) and write:packages (CI).
- Architecture: Start x86_64; add arm64 when end-to-end stable.

## One-way door decisions

- Base image path: Build FROM ubuntu:24.04 with systemd (learning goal) following Slicer's Dockerfile best practices, NOT extending Slicer's base.
- Language: Go for CLI/API (static binaries, systemd friendly).
- Orchestrator contract: Slicer-style OCI image contract (bundle kernel in image, use Dockerfile).
- Networking: Default NAT + TAP per-VM; plan for bridged mode for Trigger.dev inter-VM communication.
- Registry: GHCR private with PAT-based pulls; later consider alternative registries if needed.
- Kernel: Use proven Slicer 5.10.240 kernel bundled in image; defer custom kernel compilation to future learning phase.

## Deliverables

### Phase 0: Architecture & Contracts (1-2 days) ✅ COMPLETE
- ✅ docs/EXECUTION_PLAN.md (this file, corrected)
- ✅ docs/API_CONTRACT.md (REST API specification)
- ✅ docs/CLI_SPEC.md (CLI interface specification)
- ✅ docs/ARCHITECTURE_DECISIONS.md (comprehensive architectural decisions with justifications)
- ✅ Updated CLAUDE.md with corrected strategy

### Phase 1: Core Components (5-7 days, parallel execution) ✅ COMPLETE
- ✅ images/base/Dockerfile (FROM ubuntu:24.04, systemd, ttyS0 console, first-boot units)
- ✅ cmd/nanofuse/main.go (CLI with all subcommands: image, vm, snapshot, health, version, config)
- ✅ internal/api/server.go (HTTP server with Firecracker integration and snapshot support)
- ✅ systemd/nanofused.service (unit template)
- ✅ .github/workflows/ci.yaml (Go build + Docker build + GHCR publish + security scanning)

### Phase 2: Integration & Testing (2-3 days)
- E2E test suite
- Performance benchmarks

### Phase 3: Trigger.dev Extensions (3-4 days)
- images/trigger-web/Dockerfile.web
- images/trigger-worker/Dockerfile.worker
- Networking configuration for inter-VM communication

### Phase 4: Security & Documentation (2-3 days, parallel)
- Security threat model and hardening guide
- Complete user documentation (README, CLI guide, API reference, deployment guide)

## Milestones and acceptance criteria

### Milestone 1: Base Image Complete
- Ubuntu 24.04 image builds without interactive prompts
- Systemd services enabled (not started) during build
- Image boots in Firecracker with console on ttyS0
- SSH accessible (openssh-server running)
- Network up via systemd-networkd with DHCP
- Boot time < 2 seconds
- Image pushed to GHCR with proper tagging

### Milestone 2: CLI/API Functional
- CLI binary builds successfully
- All subcommands (pull/up/stop/resume/status) execute correctly
- GHCR authentication works
- API service starts via systemd
- API manages full VM lifecycle including snapshot/resume
- CLI can communicate with API

### Milestone 3: Integration Validated
- E2E workflow complete: pull image → boot VM → manage lifecycle → snapshot → resume
- Performance benchmarks meet targets
- All integration tests pass
- Can SSH into running VM

### Milestone 4: Trigger.dev Extensions Complete
- Web and worker images build successfully
- Both images extend base image properly
- Can deploy isolated web + worker VMs
- Inter-VM communication works (web ↔ worker)
- Services start automatically on boot

### Milestone 5: Production Ready
- CI pipeline fully green (builds + publishes to GHCR)
- Security review complete, no critical findings
- Documentation complete and validated
- Deployment guide successfully tested

## Implementation details

- Serial console: console=ttyS0; enable agetty on ttyS0 inside guest.
- Rootfs: OCI image used as VM root filesystem (Slicer-compatible path); persistent disk /dev/vdb optional.
- Auth: docker login ghcr.io with PAT (read:packages) on hosts; CI uses GITHUB_TOKEN or PAT to push.
- Config: CLI uses ~/.config/nanofuse/config.yaml for registry creds and defaults.

## Risks and mitigations

- Private pulls fail: validate PAT scope (read:packages); pre-login step in CLI/API.
- No serial output: ensure ttyS0 getty enabled in image.
- Kernel mismatch: prefer bases known to boot with Firecracker; upgrade to 6.1 guest later.
- Networking flake: begin with NAT+TAP example; document bridged mode as follow-up.

## Sub-Agent Assignments for Parallel Execution

### Phase 0 (Sequential - Must Complete First)
**Agent**: software-architect
**Task**: Define API contracts, CLI specifications, architecture decisions
**Duration**: 1-2 days

### Phase 1 (Parallel Execution - 4 Streams)

**Stream 1A - Base Image**
- **Agent**: platform-engineer
- **Task**: Build Ubuntu 24.04 Dockerfile with systemd
- **Testable Boundary**: Image boots in Firecracker, SSH works

**Stream 1B - CLI Tool**
- **Agent**: go-expert-developer
- **Task**: Implement Go CLI with all subcommands
- **Testable Boundary**: All commands execute, can call API

**Stream 1C - API Service**
- **Agent**: backend-engineer
- **Task**: Build API with Firecracker integration and snapshot support
- **Testable Boundary**: API manages VM lifecycle including resume

**Stream 1D - CI/CD Pipeline**
- **Agent**: sre-agent
- **Task**: GitHub Actions workflow for builds and publishing
- **Testable Boundary**: Pipeline green, artifacts in GHCR

### Phase 2 (Sequential - After Phase 1)
**Agent**: platform-engineer
**Task**: Integration testing and validation
**Duration**: 2-3 days

### Phase 3 (Sequential - After Phase 2)
**Agent**: backend-engineer
**Task**: Build Trigger.dev web and worker images
**Duration**: 3-4 days

### Phase 4 (Parallel - With Phase 3)

**Stream 4A - Security**
- **Agent**: secure-by-design-engineer
- **Task**: Security review and hardening

**Stream 4B - Documentation**
- **Agent**: tech-writer
- **Task**: Complete user and API documentation

## Future Enhancements (Post-Production)

- Custom kernel compilation (learning exercise).
- Multi-arch images (amd64 + arm64) with manifest list.
- Integration tests in CI on KVM-enabled runner.
- Advanced networking: custom bridge configurations.
- GPU support via Cloud Hypervisor variant.

## References

- Firecracker docs: kernel/rootfs/virtio/console
- GHCR: private registry auth and multi-arch manifests
- Slicer model: OCI image rootfs, systemd units enabled during build (no start)
