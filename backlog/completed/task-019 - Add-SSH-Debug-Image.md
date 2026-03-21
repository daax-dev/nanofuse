---
id: task-019
title: Add SSH support with user key injection
status: Done
assignee: []
created_date: '2025-11-26'
labels:
  - Feature
  - Debug
  - High
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Objective: Enable SSH access to VMs for debugging, with user-provided SSH keys injected at VM creation time.

Currently:
- Base image has SSH installed but no keys
- Todo-app image doesn't have SSH at all
- No way to inject user's SSH key
- No way to debug inside VM without exec command

## Acceptance Criteria

### AC1: VM Create Accepts SSH Key
**Given** a user wants to create a VM with SSH access
**When** running `nanofuse vm create` with `--ssh-key` flag
**Then** the public key is stored for injection

**Verification:**
```bash
nanofuse vm create myimage:latest test-vm --ssh-key ~/.ssh/id_rsa.pub
# Expected: exit code 0, key stored
```

### AC2: SSH Key Injected into VM
**Given** a VM is created with an SSH key
**When** the VM boots
**Then** the key is in /root/.ssh/authorized_keys

**Verification:**
```bash
ssh root@${VM_IP} 'cat ~/.ssh/authorized_keys' | grep -q "ssh-"
# Expected: exit code 0
```

### AC3: SSH Connection Works
**Given** a running VM with injected SSH key
**When** connecting via SSH
**Then** connection succeeds without password

**Verification:**
```bash
ssh -o StrictHostKeyChecking=no root@${VM_IP} 'hostname'
# Expected: exit code 0
```

### AC4: Todo-App Image Has SSH
**Given** the todo-app Dockerfile is updated
**When** building the image
**Then** openssh-server is installed and enabled

**Verification:**
```bash
grep -q openssh-server examples/todo-app/docker/Dockerfile
# Expected: exit code 0
```

## Implementation Approach

**Kernel cmdline injection** - cleanest approach:

1. `nanofuse vm create --ssh-key <path>` reads user's public key
2. Key is base64-encoded and added to kernel args: `sshkey=<base64>`
3. A systemd service in the image reads `/proc/cmdline`, decodes, installs key

**Why this approach:**
- No rootfs modification required
- Works with read-only images
- Key can vary per VM without rebuilding image
- SSH public key (~400 bytes) fits easily in cmdline (~4KB limit)
- Standard cloud-init-like pattern

**Design doc:** `docs/building/ssh-key-injection-design.md`

## Definition of Done
- [ ] All 4 acceptance criteria pass
- [ ] `--ssh-key` flag added to `vm create`
- [ ] Key injection implemented (Option A)
- [ ] Todo-app Dockerfile has openssh-server
- [ ] Documentation updated

Priority: High
<!-- SECTION:DESCRIPTION:END -->
