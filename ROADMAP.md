# NanoFuse Roadmap and Task Priority

**Last Updated**: 2025-11-14
**Current Status**: Phase 1 Bug Fixes 🔧 | Testing In Progress ⏳

---

## Overview

This document outlines the incomplete tasks and their priorities for the NanoFuse project. Phase 1 is complete with all core features implemented. Phase 2 focuses on advanced features like snapshot/resume, security hardening, and production readiness.

---

## Current State Summary

### 🔧 Phase 1: Testing & Bug Fixes (November 2025)

**Implementation Status**: Features implemented, critical bugs found during testing

**Critical Bugs Fixed (2025-11-14)**:
- ✅ CLI image pull failing with "Job ID is required" error
  - **Root cause**: Type mismatch between API response and client expectations
  - **Fix**: Added PullImageResponse type with proper field mapping
  - **Files**: internal/client/types.go, internal/client/client.go

- ✅ Unix socket not created when TCP also configured
  - **Root cause**: if-else logic only created one listener
  - **Fix**: Support both Unix socket AND TCP simultaneously
  - **Files**: internal/api/server.go

**Testing Status**:
- ✅ Test scripts created for image pull and dual listeners
- ⏳ End-to-end VM lifecycle testing pending
- ⏳ Demo script verification pending
- ⏳ Resource cleanup verification pending

**Implemented Features**:
- Core infrastructure (CLI, API daemon, base images)
- Networking (bridge, NAT, TAP devices, IPAM)
- VM lifecycle management (untested end-to-end)
- Image management with OCI registry support (partially tested)
- CI/CD pipeline with automated releases
- Documentation (being updated)

**Status**: Alpha - NOT production-ready. Core functionality exists but requires comprehensive testing and additional bug fixes.

---

## Phase 2: Snapshot/Resume & Advanced Features

### Priority 1: Snapshot/Resume (CRITICAL) 🔴

**Goal**: Enable fast VM snapshot and resume capabilities for sub-2-second cold starts

**Implementation Tasks**:

1. **Snapshot Creation** (3-5 days)
   - Implement `CreateSnapshot(vmID, snapshotID)` via Firecracker API
   - Store snapshot state files (memory + vmstate)
   - Add snapshot metadata to database
   - CLI command: `nanofuse vm snapshot create <vm-id> <snapshot-name>`
   - API endpoint: `POST /vms/{id}/snapshots`
   - **Files to modify**:
     - `internal/firecracker/vm.go:308` (TODO marker exists)
     - `internal/api/vm_handlers.go` (add snapshot handlers)
     - `internal/storage/db.go` (snapshot table)
   - **Testing**: Unit tests + integration tests with real Firecracker VMs

2. **Snapshot Listing** (1 day)
   - List all snapshots for a VM
   - Show snapshot metadata (size, creation time, state)
   - CLI command: `nanofuse vm snapshot list <vm-id>`
   - API endpoint: `GET /vms/{id}/snapshots`
   - **Files to modify**:
     - `internal/api/vm_handlers.go:315` (TODO marker exists)
     - `internal/storage/db.go` (query snapshots)

3. **Snapshot Deletion** (1 day)
   - Delete snapshot files and metadata
   - Cleanup storage
   - CLI command: `nanofuse vm snapshot delete <vm-id> <snapshot-name>`
   - API endpoint: `DELETE /vms/{id}/snapshots/{snapshot-id}`
   - **Files to modify**:
     - `internal/api/vm_handlers.go:322` (TODO marker exists)
     - `internal/storage/db.go` (delete snapshot records)

4. **Resume from Snapshot** (3-5 days)
   - Load snapshot and resume VM execution
   - Restore network state
   - Verify all resources are properly restored
   - CLI command: `nanofuse vm resume <vm-id> <snapshot-name>`
   - API endpoint: `POST /vms/{id}/resume`
   - **Files to modify**:
     - `internal/api/vm_handlers.go:841` (TODO marker exists)
     - `internal/firecracker/vm.go:329` (TODO marker exists)
   - **Testing**: End-to-end tests with snapshot → stop → resume cycle

5. **Snapshot Storage Management** (2-3 days)
   - Implement configurable snapshot storage location
   - Add snapshot size limits
   - Implement snapshot cleanup policies (retention, max count)
   - Optional: S3 backend for snapshot storage

**Expected Timeline**: 2-3 weeks
**Estimated Effort**: 10-15 days
**Priority**: CRITICAL - This is the core Phase 2 feature

**Success Criteria**:
- ✅ Sub-2-second cold starts with snapshot resume
- ✅ Snapshots persist across daemon restarts
- ✅ Network state properly restored
- ✅ All integration tests pass

---

### Priority 2: S3 Backup/Restore (MEDIUM) 🟡

**Goal**: Enable backup of VMs to S3-compatible storage

**Status**: Stubbed handlers exist but not implemented

**Implementation Tasks**:

1. **S3 Client Integration** (2 days)
   - Add S3 client library (AWS SDK)
   - Configuration for S3 endpoint, credentials, bucket
   - **Files to modify**:
     - `internal/storage/s3.go` (new file)
     - `config.dev.yaml` (S3 configuration)

2. **Backup Creation** (2-3 days)
   - Package VM state + rootfs + snapshot into tarball
   - Upload to S3 with metadata
   - **Files to modify**:
     - `internal/api/backup_handlers.go:80` (TODO marker exists)
     - `internal/storage/s3.go` (upload implementation)

3. **Backup Listing** (1 day)
   - List backups from S3
   - Show metadata (size, date, vm info)
   - **Files to modify**:
     - `internal/api/backup_handlers.go:50` (TODO marker exists)

4. **Backup Retrieval** (1 day)
   - Download backup from S3
   - **Files to modify**:
     - `internal/api/backup_handlers.go:119` (TODO marker exists)

5. **Backup Deletion** (1 day)
   - Delete backup from S3
   - **Files to modify**:
     - `internal/api/backup_handlers.go:126` (TODO marker exists)

6. **Restore from Backup** (2-3 days)
   - Download and extract backup
   - Create VM from restored state
   - **Files to modify**:
     - `internal/api/backup_handlers.go:133` (TODO marker exists)

**Expected Timeline**: 2 weeks
**Estimated Effort**: 9-11 days
**Priority**: MEDIUM - Useful for disaster recovery but not critical

**Success Criteria**:
- ✅ Can backup running VMs to S3
- ✅ Can restore VMs from S3 backups
- ✅ Supports S3-compatible storage (MinIO, etc.)
- ✅ Proper error handling and retry logic

---

### Priority 3: Security Hardening (HIGH) 🟠

**Goal**: Production-grade security for deployments

**Implementation Tasks**:

1. **TLS/SSL for TCP Transport** (2-3 days)
   - Add TLS configuration for TCP API
   - Certificate management
   - Mutual TLS (mTLS) support
   - **Files to modify**:
     - `cmd/nanofused/main.go` (TLS listener)
     - `internal/client/client.go` (TLS client)

2. **Authentication & Authorization** (5-7 days)
   - API key authentication
   - Role-based access control (RBAC)
   - JWT token support
   - **Files to modify**:
     - `internal/api/middleware.go` (new file)
     - `internal/auth/*` (new package)
     - `internal/storage/db.go` (auth tables)

3. **Security Audit** (3-5 days)
   - Code review for security vulnerabilities
   - Penetration testing
   - Fix identified issues
   - Document security posture

4. **Secrets Management** (2 days)
   - Secure storage for sensitive config
   - Integration with vault/secrets manager
   - **Files to modify**:
     - `internal/secrets/*` (new package)
     - `config.dev.yaml` (vault configuration)

**Expected Timeline**: 3 weeks
**Estimated Effort**: 12-17 days
**Priority**: HIGH - Required before production use with sensitive data

---

### Priority 4: Multi-Architecture Support (MEDIUM) 🟡

**Goal**: Full ARM64 support alongside x86_64

**Implementation Tasks**:

1. **ARM64 Base Image** (3-4 days)
   - Build ARM64 version of base image
   - Test with Firecracker on ARM64
   - **Files to modify**:
     - `images/base/Dockerfile` (multi-arch build)
     - `.github/workflows/ci.yaml` (ARM64 builds)

2. **Architecture Detection** (DONE ✅)
   - Already implemented in Phase 1
   - Dynamic architecture from OCI images

3. **Testing Infrastructure** (2-3 days)
   - ARM64 test runners in CI
   - Integration tests on ARM64
   - **Files to modify**:
     - `.github/workflows/ci.yaml` (ARM64 runners)

**Expected Timeline**: 1 week
**Estimated Effort**: 5-7 days
**Priority**: MEDIUM - Nice to have for broader platform support

---

### Priority 5: Observability & Monitoring (MEDIUM) 🟡

**Goal**: Production-grade monitoring and metrics

**Implementation Tasks**:

1. **Metrics Endpoint** (2-3 days)
   - Prometheus metrics
   - VM count, network stats, API latency
   - **Files to modify**:
     - `internal/api/metrics.go` (new file)
     - `cmd/nanofused/main.go` (metrics endpoint)

2. **Structured Logging** (1-2 days)
   - JSON logging format
   - Log aggregation support
   - **Files to modify**:
     - `internal/logging/*` (new package)

3. **Health Checks** (1 day)
   - Liveness and readiness endpoints
   - **Files to modify**:
     - `internal/api/health.go` (new file)

4. **Distributed Tracing** (2-3 days)
   - OpenTelemetry integration
   - Trace VM lifecycle operations
   - **Files to modify**:
     - `internal/tracing/*` (new package)

**Expected Timeline**: 1-2 weeks
**Estimated Effort**: 6-9 days
**Priority**: MEDIUM - Useful for production operations

---

### Priority 6: Performance & Scale (LOW) 🟢

**Goal**: Optimize for larger deployments

**Implementation Tasks**:

1. **Load Testing** (3-5 days)
   - Benchmark VM creation/deletion
   - Test with 245 concurrent VMs
   - Identify bottlenecks
   - **Files**: Test scripts in `test/performance/`

2. **Database Optimization** (2-3 days)
   - Add indexes for common queries
   - Optimize IPAM allocation
   - **Files to modify**:
     - `internal/storage/db.go` (indexes)
     - `internal/network/ipam.go` (optimization)

3. **Connection Pooling** (1-2 days)
   - Optimize HTTP client connections
   - Database connection pooling
   - **Files to modify**:
     - `internal/client/client.go` (connection pool)

4. **Caching** (2-3 days)
   - Cache image metadata
   - Cache VM states
   - **Files to modify**:
     - `internal/cache/*` (new package)

**Expected Timeline**: 2 weeks
**Estimated Effort**: 8-13 days
**Priority**: LOW - Current performance is acceptable

---

## Minor Improvements & Tech Debt

### Code Quality

1. **Better Random Number Generation** (30 minutes)
   - Replace `math/rand` with `crypto/rand` for MAC address generation
   - **File**: `internal/firecracker/vm.go:430` (TODO marker exists)

2. **Test Coverage Improvement** (ongoing)
   - Target: 80%+ code coverage
   - Add unit tests for edge cases

3. **Documentation Updates** (ongoing)
   - Keep docs in sync with features
   - Add more examples and tutorials

---

## Timeline Summary

### Phase 2 Implementation Plan (Q1 2026)

**Month 1 (Weeks 1-4)**:
- ✅ Week 1-3: Snapshot/Resume implementation
- ✅ Week 4: Testing and bug fixes

**Month 2 (Weeks 5-8)**:
- Week 5-6: Security hardening (TLS, auth)
- Week 7-8: S3 backup/restore OR multi-arch support

**Month 3 (Weeks 9-12)**:
- Week 9-10: Observability & monitoring
- Week 11-12: Performance testing & optimization

---

## Success Metrics

### Phase 2 Goals

- **Performance**:
  - Sub-2-second cold starts with snapshot resume
  - Support 245 concurrent VMs
  - <10ms API latency (p95)

- **Reliability**:
  - 99.9% uptime for daemon
  - Graceful degradation on failures
  - Automatic recovery from crashes

- **Security**:
  - TLS for all TCP connections
  - Authentication & authorization
  - No critical vulnerabilities in security scans

- **Usability**:
  - Comprehensive documentation
  - Clear error messages
  - Easy deployment and configuration

---

## Dependencies & Risks

### External Dependencies

- Firecracker API: Snapshot/resume requires Firecracker 1.0+
- AWS SDK: For S3 backup functionality
- OpenTelemetry: For distributed tracing

### Risks

1. **Firecracker API Changes**: Mitigation: Pin to stable Firecracker version
2. **Storage Limits**: Mitigation: Implement snapshot cleanup policies
3. **Network Complexity**: Mitigation: Comprehensive integration tests

---

## Next Steps

### Immediate Actions (This Week)

1. ✅ Complete documentation cleanup (DONE)
2. ⏳ Review and prioritize Phase 2 tasks with team
3. ⏳ Set up Phase 2 project board
4. ⏳ Begin snapshot/resume design document

### Next Month

1. Start snapshot/resume implementation
2. Add comprehensive tests for snapshot feature
3. Update user documentation with snapshot examples

---

## Conclusion

NanoFuse Phase 1 provides a solid foundation for Firecracker-based VM management. Phase 2 will add the advanced features needed for production use at scale, with snapshot/resume being the highest priority feature.

The roadmap is flexible and priorities may shift based on user feedback and business needs.

**For questions or feedback, see**: [CONTRIBUTING.md](docs/CONTRIBUTING.md)
