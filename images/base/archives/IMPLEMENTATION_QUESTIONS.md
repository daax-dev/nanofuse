# Implementation Questions - Awaiting Answers

**Status**: 🟡 Waiting for User Input
**Date**: 2025-11-08
**Context**: Planning GitHub Actions CI/CD and User Documentation

---

## Critical Questions

### Q1: Repository Migration
**Question**: You mentioned moving to `peregrinesummit` org with self-hosted runners.

**Need to Know**:
- [ ] When will the repo be migrated to peregrinesummit org?
- [ ] Do the self-hosted runners have KVM enabled?
- [ ] What OS do the runners use? (Ubuntu 24.04 preferred)
- [ ] Are they already set up or do we need to configure them?
- [ ] Can we test on them before migration?

**Impact**: Determines whether we can run Firecracker boot tests in CI
**Blocks**: Phase 2 implementation

---

### Q2: Build Strategy - Local vs CI
**Question**: Where should builds happen?

**Options**:
A. **Build locally, test in CI** (safest)
   - Dev builds kernel locally
   - Pushes to git
   - CI just tests the committed artifacts

B. **Build in CI, publish artifacts** (standard)
   - CI builds kernel from scratch
   - CI runs tests
   - CI publishes to GHCR/releases

C. **Hybrid** (flexible)
   - CI builds and tests on every PR/commit
   - Manual releases use local builds + CI validation

**Your Answer**: ________________

**Impact**: Determines CI workflow structure
**Blocks**: Phase 2.1 (Basic workflow)

---

### Q3: Container Registry Strategy
**Question**: You mentioned GHCR "until migration to private runners"

**Need to Clarify**:
- [ ] Use GHCR (GitHub Container Registry) for now?
- [ ] Will eventually move to private registry on peregrinesummit infrastructure?
- [ ] What credentials/permissions needed for GHCR push?
- [ ] Should we set up both registries now, or GHCR only?

**Your Statement**: "use ghcr where needed (until migration to private runners)"

**Impact**: Determines where we publish container images
**Blocks**: Phase 2.4 (Release publishing)

---

### Q4: Release Process (RELEASE.md)
**Question**: What level of release automation?

**You Want**:
- ✅ Semver (semantic versioning)
- ✅ `[release]` tag/keyword trigger
- ⏸️ Easy on SLSA levels for now
- 📝 Document in `./docs/RELEASE.md`

**Need to Know**:
- [ ] Manual tag creation (dev tags `v1.2.3` manually)?
- [ ] Automatic from commit message (e.g., commit with `[release]`)?
- [ ] What goes in release:
  - [ ] Kernel binary (vmlinux)
  - [ ] Rootfs image (rootfs.ext4)
  - [ ] Both as container image?
  - [ ] Manifest.json
  - [ ] Release notes (auto-generated?)

**Impact**: Determines release workflow
**Blocks**: Phase 2.4 (Release publishing) + docs/RELEASE.md

---

### Q5: Documentation Scope
**Question**: How much documentation to write now?

**Options**:
A. **Just image building** (minimal, focused)
   - How to build kernel
   - How to test locally
   - How to use in CI

B. **Full platform docs** (comprehensive)
   - Above + CLI usage
   - Above + API integration
   - Above + networking setup
   - Above + production deployment

**Your Answer**: ________________

**Impact**: Determines Phase 3 scope
**Blocks**: Phase 3 (User documentation)

---

### Q6: Target Audience Priority
**Question**: Who should we write for first?

**Audiences**:
1. **Developers** (contribute to NanoFuse)
2. **DevOps Engineers** (integrate into their pipelines)
3. **End Users** (just want to run VMs)

**Your Answer**: Priority order: _____, _____, _____

**Impact**: Determines documentation structure and detail level
**Blocks**: Phase 3 (User documentation)

---

### Q7: Testing in CI - What to Validate?
**Question**: Given self-hosted runners with KVM, what should CI test?

**Options** (can select multiple):
- [ ] Kernel builds successfully
- [ ] Firecracker boots kernel
- [ ] All VIRTIO devices work
- [ ] Filesystem mounts
- [ ] Network connectivity works
- [ ] SSH into VM works
- [ ] systemd reaches multi-user.target
- [ ] HTTP test server responds

**Your Answer**: Check all that apply above

**Impact**: Determines test coverage in CI
**Blocks**: Phase 2.3 (Conditional testing)

---

## Answers Summary (Fill This In)

```yaml
q1_migration:
  when: "TBD"
  runners_have_kvm: true/false
  runners_os: "Ubuntu 24.04"
  setup_status: "ready" or "need setup"

q2_build_strategy: "A" or "B" or "C"

q3_registry:
  use_ghcr_now: true/false
  future_private_registry: "URL or TBD"
  ghcr_credentials: "who sets up?"

q4_release:
  trigger: "manual tag" or "[release] in commit"
  artifacts:
    - "vmlinux"
    - "rootfs.ext4"
    - "container image"
    - "manifest.json"
  release_notes: "auto" or "manual"

q5_docs_scope: "A" or "B"

q6_audience_priority:
  - "Developers"
  - "DevOps"
  - "End Users"

q7_ci_tests:
  - "kernel_builds"
  - "firecracker_boots"
  - "virtio_works"
  - "filesystem_mounts"
  - "networking"
  - "ssh"
  - "systemd"
  - "http_server"
```

---

## Next Steps After Answers

1. **If repo migrating soon**: Wait for migration, then implement on peregrinesummit org
2. **If repo staying**: Implement now, migrate later
3. **Start with**: Phase 2.1 (Basic CI workflow) using answers above
4. **Then**: Write docs/RELEASE.md based on Q4 answers
5. **Then**: Phase 3 (User docs) based on Q5-Q6 answers

---

## Notes

- **SLSA**: Starting easy as requested - basic semver + signed releases, not full SLSA attestation
- **Self-hosted runners**: Big advantage! Means we can actually test Firecracker boots in CI
- **GHCR**: Good interim solution, easy to migrate to private registry later
- **docs/RELEASE.md**: Will document the release process once we decide on automation level

---

## Status Tracking

- [ ] Questions answered by user
- [ ] Implementation plan updated with answers
- [ ] Proceed with Phase 2 (GitHub Actions)
- [ ] Proceed with docs/RELEASE.md
- [ ] Proceed with Phase 3 (User docs)
