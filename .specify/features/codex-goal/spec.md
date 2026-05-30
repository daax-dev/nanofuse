# Feature Specification: Sandbox Objective Closed-Loop Validation

**Feature Branch**: `codex-goal`
**Created**: 2026-05-30
**Status**: Approved
**Input**: User description: "provide a microvm with a small security surface area - capable of running on Linux, Windows, Mac. Containers, persistent filesystem, fast and long lifetimes, secrets/identity away from LLM, egress/API/MCP interception/restriction, Vagrant closed-loop testing, and GOALS.md fixes."

## User Scenarios & Testing

### User Story 1 - Isolated Workload Execution (Priority: P1)

An operator can start an untrusted workload in a microVM boundary with a dedicated guest kernel and a minimized host-facing surface, then verify from local tests that the workload cannot share the host kernel.

**Why this priority**: Hardware isolation is the core security promise and all other controls depend on a VM boundary.

**Independent Test**: Create and start a sandbox in the closed-loop Linux/KVM environment, then verify the sandbox runs from a Firecracker-backed microVM with a dedicated root disk and network interface.

**Acceptance Scenarios**:

1. **Given** a host that exposes Linux KVM, **When** the operator creates and starts a sandbox, **Then** the workload runs behind a microVM boundary and does not reuse a shared container kernel.
2. **Given** a host that cannot expose Linux KVM, **When** the operator runs closed-loop validation, **Then** validation fails before launch with a concrete unsupported-host reason.

---

### User Story 2 - Persistent and Disposable Filesystems (Priority: P1)

An operator can choose a sandbox lifetime model where a VM's writable filesystem persists across stop/start, while the source image remains immutable and reusable for other VMs.

**Why this priority**: Coding agents need state during a session, but image mutation across tenants breaks isolation and repeatability.

**Independent Test**: Create a VM and verify its writable root disk path is VM-specific, persists while the VM exists, and is removed when the VM is deleted.

**Acceptance Scenarios**:

1. **Given** a registered image, **When** a VM is created, **Then** the VM uses its own writable root disk copy.
2. **Given** a stopped VM, **When** it is started again, **Then** the same VM root disk is reused.
3. **Given** a deleted VM, **When** cleanup completes, **Then** VM-specific disk state and host network policy are removed.

---

### User Story 3 - Restricted Egress and Credential Separation (Priority: P1)

An operator can declare a network posture that denies arbitrary egress and only permits explicit destinations such as a host-controlled proxy, DNS, or bootstrapping endpoints. Raw secrets are not required in guest-visible VM configuration.

**Why this priority**: LLM-generated code must not be able to exfiltrate arbitrary data or read ambient credentials.

**Independent Test**: Apply a default-deny egress policy for a VM and verify generated host firewall rules only allow declared traffic and fail closed otherwise.

**Acceptance Scenarios**:

1. **Given** a sandbox with default-deny egress, **When** network rules are installed, **Then** outbound traffic is denied except for declared allow rules.
2. **Given** a proxy-only sandbox, **When** network rules are installed, **Then** guest traffic can reach only the configured proxy/DNS/bootstrapping endpoints.
3. **Given** a sandbox with identity metadata, **When** VM configuration is inspected, **Then** only identity references or public material are present, not raw secret values.

---

### User Story 4 - Container Workload Wrapping (Priority: P2)

An operator can build a microVM root filesystem from OCI/container inputs and run the resulting workload inside the microVM isolation boundary.

**Why this priority**: Existing developer workflows and agent workloads are distributed as container images.

**Independent Test**: Validate the container-to-rootfs build path with dry-run/unit tests and include it in the Vagrant closed-loop validation where host capabilities permit.

**Acceptance Scenarios**:

1. **Given** a supported container image or layer manifest, **When** the build path runs, **Then** it produces a bootable rootfs artifact for the microVM workflow.
2. **Given** a missing kernel/rootfs prerequisite, **When** validation runs, **Then** the failure identifies the missing artifact.

---

### User Story 5 - Cross-Platform Operator Path (Priority: P2)

An operator on Linux, macOS, or Windows can understand the supported execution model and run validation only when the required Linux/KVM substrate is available.

**Why this priority**: The product goal names all three host families, but the security boundary cannot be misrepresented.

**Independent Test**: Run the closed-loop preflight on the current host and capture whether the host/provider exposes Linux KVM.

**Acceptance Scenarios**:

1. **Given** Linux with KVM, **When** validation runs, **Then** Firecracker validation proceeds.
2. **Given** macOS or Windows without exposed nested KVM, **When** validation runs, **Then** validation reports that a Linux/KVM runner is required.
3. **Given** a remote or Vagrant Linux VM exposing KVM, **When** validation runs, **Then** kernel-level tests can execute inside that VM.

### Edge Cases

- KVM unavailable, unreadable, or not writable.
- Firecracker binary absent or for the wrong architecture.
- Vagrant provider creates a guest architecture that cannot run the configured kernel/rootfs.
- Rootfs copy fails midway and leaves partial state.
- Egress rules fail to install after a TAP device is created.
- Cleanup runs for partially-created VMs.
- A policy asks for proxy-only egress without a proxy endpoint.

## Requirements

### Functional Requirements

- **FR-001**: System MUST run untrusted workload sessions behind a hardware-virtualized microVM boundary when Linux KVM is available.
- **FR-002**: System MUST fail closed with actionable diagnostics when Linux KVM or a compatible hypervisor path is unavailable.
- **FR-003**: System MUST support writable per-VM filesystem state without mutating the registered source image.
- **FR-004**: System MUST support deletion cleanup for VM-specific disk and network policy state.
- **FR-005**: System MUST support a default-deny egress mode with explicit allow rules.
- **FR-006**: System MUST support a proxy-only egress mode suitable for LLM/API/MCP interception by a host-controlled proxy.
- **FR-007**: System MUST keep raw secret values out of guest-visible VM configuration and document any remaining credential-broker gaps.
- **FR-008**: System MUST retain a container-to-rootfs path for wrapping container-distributed workloads in the microVM boundary.
- **FR-009**: System MUST provide closed-loop validation that exercises local build/test gates and Vagrant/hypervisor capability checks.
- **FR-010**: System MUST document supported and unsupported platform paths for Linux, macOS, and Windows without overstating security guarantees.

### Key Entities

- **Sandbox VM**: A microVM instance with identity, lifecycle state, filesystem state, and network policy.
- **Root Disk**: A VM-specific writable filesystem image derived from an immutable source image.
- **Egress Policy**: The outbound network policy attached to a sandbox VM.
- **Identity Reference**: A SPIFFE/SPIRE identity or other non-secret reference used to fetch scoped credentials outside LLM-visible config.
- **Validation Run**: A recorded local or Vagrant execution with command, result, and blocker details.

## Success Criteria

### Measurable Outcomes

- **SC-001**: VM creation stores a writable root disk under VM-specific storage for every writable rootfs-backed VM.
- **SC-002**: Unit tests verify per-VM root disk creation, cleanup, and egress rule generation without requiring root.
- **SC-003**: Closed-loop Vagrant validation records pass/fail evidence for KVM, Firecracker, image artifacts, daemon gates, and VM boot where the host supports it.
- **SC-004**: `docs/GOALS.md` states current support, target support, and constraints for every objective listed in `objective.md`.
- **SC-005**: `mage ci` passes, or any blocker includes the exact command, environment, and failure cause.
