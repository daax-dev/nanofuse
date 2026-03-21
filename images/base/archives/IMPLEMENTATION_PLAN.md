# Implementation Plan: CI/CD and User Documentation

## Status: Planning Phase
**Last Updated**: 2025-11-08
**Owner**: Claude Code + User

---

## Phase 1: Document Successful Build Process ✅ COMPLETE

**Status**: ✅ Done
**Artifacts**:
- `BUILD_SUCCESS.md` - Complete working documentation
- `BUILD_KERNEL_ONLY.sh` - Isolated kernel build
- `TEST_KERNEL_ONLY.sh` - Isolated test
- `BUILD_AND_TEST.sh` - Combined build+test

**Verified Working**:
- ✅ Kernel builds successfully (Linux 6.1.90)
- ✅ All boot tests pass
- ✅ VIRTIO-MMIO devices work
- ✅ Root filesystem mounts
- ✅ No kernel panics

---

## Phase 2: GitHub Actions Integration

### Status: 🔴 NOT STARTED

### Goal
Create a GitHub Actions workflow that:
1. Builds the kernel on every commit
2. Tests the kernel boots successfully
3. Publishes artifacts on release
4. Runs on Ubuntu 24.04 (matches our build environment)

### Complexity Analysis

#### Non-Trivial Issues

##### Issue 1: Firecracker in GitHub Actions
**Problem**: GitHub Actions runners don't have KVM/nested virtualization by default
- Firecracker requires `/dev/kvm`
- GitHub-hosted runners typically don't expose KVM
- Our tests require Firecracker to actually boot the VM

**Options**:
1. **Use self-hosted runner with KVM** (BEST)
   - Pros: Full KVM access, can run Firecracker normally
   - Cons: Need to set up and maintain runner infrastructure

2. **Use GitHub-hosted runner with KVM emulation**
   - Pros: No infrastructure to maintain
   - Cons: Much slower, may not work at all
   - Research: Check if GitHub Actions supports KVM

3. **Skip Firecracker boot test in CI**
   - Pros: Simple, CI runs on any runner
   - Cons: Not testing what we ship, defeats the purpose
   - Only build kernel, don't test boot

4. **Use container-based Firecracker alternative**
   - Pros: Works in containers
   - Cons: Not actually testing Firecracker, just Docker
   - Research: QEMU in Docker? Cloud Hypervisor?

**Recommendation**: Option 1 (self-hosted runner) or Option 3 (skip boot test in CI, test manually)

##### Issue 2: Docker in Docker
**Problem**: GitHub Actions runs in containers, we need Docker to build kernel
- Need Docker to build kernel Docker image
- Need to run Docker inside GitHub Actions (which is already containerized)

**Options**:
1. **Use GitHub Actions Docker service** (RECOMMENDED)
   - Actions have built-in Docker support
   - Should work out of the box

2. **Use docker-in-docker (dind)**
   - More complex
   - Security concerns

3. **Pre-build kernel image and pull from registry**
   - Chicken-and-egg problem for first build
   - Doesn't test kernel building

**Recommendation**: Option 1 (use Actions' Docker service)

##### Issue 3: Sudo in GitHub Actions
**Problem**: Our scripts require sudo, but Actions may run as root or as a user
- `build.sh` needs sudo for mounting
- `BUILD_KERNEL_ONLY.sh` needs sudo for Docker
- `TEST_KERNEL_ONLY.sh` must NOT use sudo

**Options**:
1. **Detect environment and adapt** (RECOMMENDED)
   - Check if already root
   - Check if sudo available
   - Adapt scripts accordingly

2. **Rewrite scripts to not need sudo**
   - Use Docker for mounting (no sudo needed)
   - But more complex

3. **Use Actions' built-in sudo**
   - Actions support sudo
   - Just use it as-is

**Recommendation**: Option 3 (use Actions' sudo) + Option 1 (detect and adapt)

##### Issue 4: Artifact Storage
**Problem**: Need to store and publish build artifacts
- Kernel binary is ~40MB
- Rootfs is ~2GB
- Can't store 2GB in GitHub Actions artifacts (limit 500MB per artifact, 2GB total per workflow)

**Options**:
1. **Only publish kernel binary** (RECOMMENDED)
   - Kernel is 40MB, well under limit
   - Rootfs can be rebuilt from Dockerfile

2. **Compress rootfs**
   - gzip may get it under 500MB
   - Worth trying

3. **Use external storage**
   - S3, Google Cloud Storage
   - More complex, costs money

4. **Use GitHub Releases**
   - Releases have 2GB file limit
   - Perfect for our use case

**Recommendation**: Option 1 (kernel only) + Option 4 (use Releases for full images)

##### Issue 5: Build Time
**Problem**: Kernel builds take ~4 minutes, may timeout or be expensive
- GitHub Actions has 6 hour timeout (fine)
- But we pay per minute on private repos
- Want to minimize CI time

**Options**:
1. **Cache Docker layers** (RECOMMENDED)
   - Use `actions/cache` to cache Docker build layers
   - Huge speedup on incremental builds

2. **Use pre-built kernel base image**
   - Build kernel build environment once
   - Cache that image

3. **Only build on specific events**
   - Don't build every commit
   - Only build on: PR, merge to main, tags

**Recommendation**: Options 1 + 3 (cache + selective builds)

### Implementation Plan

#### Step 2.1: Create Basic Workflow ⏳
- [ ] Create `.github/workflows/build-kernel.yml`
- [ ] Set up Ubuntu 24.04 runner
- [ ] Install dependencies (Docker, basic tools)
- [ ] Run `BUILD_KERNEL_ONLY.sh`
- [ ] Upload kernel artifact
- [ ] **Estimate**: 1 hour

#### Step 2.2: Add Docker Layer Caching ⏳
- [ ] Add `actions/cache` for Docker layers
- [ ] Configure cache key based on Dockerfile hash
- [ ] Test cache hit/miss scenarios
- [ ] **Estimate**: 30 minutes

#### Step 2.3: Add Conditional Testing ⏳
- [ ] Check if KVM available
- [ ] If yes: Run full `TEST_KERNEL_ONLY.sh`
- [ ] If no: Skip boot test, mark as warning
- [ ] Add workflow status badge
- [ ] **Estimate**: 1 hour

#### Step 2.4: Add Release Publishing ⏳
- [ ] Trigger on git tags (v*)
- [ ] Build kernel + rootfs
- [ ] Create GitHub Release
- [ ] Upload artifacts to release
- [ ] Generate release notes
- [ ] **Estimate**: 1 hour

#### Step 2.5: Add Security Scanning ⏳
- [ ] Scan kernel binary for vulns (optional)
- [ ] Scan Docker image for CVEs
- [ ] Add dependency review
- [ ] **Estimate**: 30 minutes (optional)

**Total Estimate**: 4-5 hours

### Success Criteria
- [ ] Workflow runs on every PR
- [ ] Workflow runs on every merge to main
- [ ] Workflow publishes on git tags
- [ ] Build completes in < 10 minutes
- [ ] Kernel artifact uploaded successfully
- [ ] Tests pass (or skip gracefully if no KVM)

---

## Phase 3: User Installation Guide

### Status: 🔴 NOT STARTED

### Goal
Create comprehensive documentation for users to:
1. Install NanoFuse from scratch
2. Build their own images
3. Run VMs locally
4. Troubleshoot common issues

### Complexity Analysis

#### Non-Trivial Issues

##### Issue 1: Multiple Installation Paths
**Problem**: Different users have different needs
- Dev: wants to build everything from source
- User: just wants to run pre-built images
- CI/CD: wants to integrate into their pipeline

**Solution**: Create 3 separate guides:
1. **Quick Start** - Use pre-built images (5 min)
2. **Developer Guide** - Build from source (30 min)
3. **Integration Guide** - CI/CD integration (varies)

##### Issue 2: Prerequisites Vary by OS
**Problem**: Ubuntu, Fedora, macOS, Windows all different
- Package managers differ (apt, dnf, brew, choco)
- Firecracker installation differs
- KVM access differs (or impossible on macOS/Windows)

**Solution**:
- Primary support: Ubuntu 24.04 (our build target)
- Secondary support: Fedora, Arch (Linux with KVM)
- Experimental: macOS (no Firecracker, Docker only)
- Not supported: Windows (WSL2 possible but complex)

##### Issue 3: Firecracker Setup
**Problem**: Firecracker requires specific setup
- Must have KVM access
- User must be in `kvm` group
- `/dev/kvm` must have correct permissions
- jailer (optional) requires additional setup

**Solution**: Step-by-step guide with verification at each step

##### Issue 4: Networking Setup
**Problem**: VMs need network access, not trivial
- TAP devices must be created
- Bridging or routing must be configured
- Firewall rules may block
- Complex for new users

**Solution**:
- Start with no networking (simplest)
- Then add basic networking guide
- Then add advanced networking (bridge, NAT, etc.)

##### Issue 5: Storage/Persistence
**Problem**: How do users manage VM disk images?
- Where to store rootfs images?
- How to handle multiple VMs?
- Copy-on-write vs full copies?
- Quota management?

**Solution**: Document best practices and provide helper scripts

### Implementation Plan

#### Step 3.1: Quick Start Guide ⏳
Target audience: Users who just want to run a VM

- [ ] Title: "5-Minute Quick Start"
- [ ] Prerequisites check (OS, KVM, Docker)
- [ ] Install Firecracker (one command)
- [ ] Download pre-built image from releases
- [ ] Run VM (single command)
- [ ] Verify it works
- [ ] **Estimate**: 1 hour

#### Step 3.2: Developer Build Guide ⏳
Target audience: Developers who want to build from source

- [ ] Title: "Building NanoFuse Images from Source"
- [ ] Clone repository
- [ ] Install build dependencies
- [ ] Build kernel (`BUILD_KERNEL_ONLY.sh`)
- [ ] Build rootfs (`build.sh`)
- [ ] Test locally (`TEST_KERNEL_ONLY.sh`)
- [ ] Troubleshooting common build issues
- [ ] **Estimate**: 2 hours

#### Step 3.3: Firecracker Setup Guide ⏳
Target audience: Users new to Firecracker

- [ ] Title: "Setting Up Firecracker"
- [ ] Check KVM access (`ls -l /dev/kvm`)
- [ ] Add user to kvm group
- [ ] Install Firecracker binary
- [ ] Verify installation
- [ ] Common issues (permissions, KVM not available)
- [ ] **Estimate**: 1 hour

#### Step 3.4: Networking Guide ⏳
Target audience: Users who need VM networking

- [ ] Title: "Configuring VM Networking"
- [ ] Option 1: No networking (simplest)
- [ ] Option 2: TAP device (single VM)
- [ ] Option 3: Bridge (multiple VMs)
- [ ] Option 4: NAT (internet access)
- [ ] Troubleshooting network issues
- [ ] **Estimate**: 2 hours

#### Step 3.5: CLI Usage Guide ⏳
Target audience: Users of the `nanofuse` CLI (future)

- [ ] Title: "Using the NanoFuse CLI"
- [ ] Installation
- [ ] `nanofuse image` commands
- [ ] `nanofuse vm` commands
- [ ] `nanofuse network` commands
- [ ] Configuration files
- [ ] **Estimate**: 2 hours (depends on CLI implementation)

#### Step 3.6: Troubleshooting Guide ⏳
Target audience: Everyone

- [ ] Title: "Troubleshooting Common Issues"
- [ ] Categorized by symptom
- [ ] Step-by-step diagnosis
- [ ] Known issues and workarounds
- [ ] How to report bugs
- [ ] **Estimate**: 1 hour

**Total Estimate**: 9-10 hours

### Success Criteria
- [ ] User can go from zero to running VM in 5 minutes (Quick Start)
- [ ] Developer can build from source successfully (Build Guide)
- [ ] All prerequisites are documented
- [ ] Common issues have solutions
- [ ] Guides tested on fresh Ubuntu 24.04 install

---

## Phase 4: Testing and Validation

### Status: 🔴 NOT STARTED

### Tasks
- [ ] Test GitHub Actions workflow on actual GitHub
- [ ] Test Quick Start guide on fresh Ubuntu 24.04 VM
- [ ] Test Developer Guide on fresh system
- [ ] Fix any issues found during testing
- [ ] Update documentation based on test results

**Estimate**: 3-4 hours

---

## Overall Timeline

| Phase | Estimate | Status |
|-------|----------|--------|
| Phase 1: Document Build Process | 2 hours | ✅ Done |
| Phase 2: GitHub Actions | 4-5 hours | 🔴 Not Started |
| Phase 3: User Documentation | 9-10 hours | 🔴 Not Started |
| Phase 4: Testing & Validation | 3-4 hours | 🔴 Not Started |
| **TOTAL** | **18-21 hours** | **5% Complete** |

---

## Next Actions

### Immediate (Do First)
1. Review this plan with user
2. Get approval on approach for GitHub Actions (KVM issue)
3. Get approval on user documentation structure
4. Prioritize: Do we need Phase 2 or Phase 3 first?

### Phase 2 Priority (GitHub Actions)
1. Create basic workflow (Step 2.1)
2. Test on GitHub
3. Add caching (Step 2.2)
4. Add testing if KVM available (Step 2.3)
5. Add release publishing (Step 2.4)

### Phase 3 Priority (User Docs)
1. Write Quick Start (Step 3.1)
2. Write Firecracker Setup (Step 3.3)
3. Write Developer Build Guide (Step 3.2)
4. Write Networking Guide (Step 3.4)
5. Write Troubleshooting (Step 3.6)

---

## Risks and Mitigations

### Risk: GitHub Actions doesn't have KVM
**Impact**: High - can't test kernel boots
**Mitigation**: Skip boot test in CI, test manually before releases
**Status**: Need to investigate

### Risk: Documentation gets out of sync with code
**Impact**: Medium - users get confused
**Mitigation**: Include docs in PR reviews, CI check for broken links
**Status**: Process to be defined

### Risk: Build time too long for CI
**Impact**: Medium - expensive, slow feedback
**Mitigation**: Aggressive caching, only build on important events
**Status**: Will monitor after implementation

### Risk: User guide too complex
**Impact**: Medium - high barrier to entry
**Mitigation**: Start with simplest possible quick start, progressively add detail
**Status**: Design phase

---

## Questions for User

1. **Priority**: Should we do GitHub Actions (Phase 2) or User Docs (Phase 3) first?
2. **GitHub Actions KVM**: Do you have access to a self-hosted runner with KVM? Or should we skip boot tests in CI?
3. **Release cadence**: How often do you want to publish releases? (On every tag? Manual?)
4. **User audience**: Who is the primary audience? Developers? DevOps? End users?
5. **Scope**: Should we also document the full NanoFuse platform (CLI, API) or just image building?

---

## Progress Tracking

This file will be updated as work progresses. Each checkbox represents a completable unit of work. Estimates may be revised as we learn more.

**Next Update**: After completing Phase 2.1 (Basic GitHub Actions workflow)
