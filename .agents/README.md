# NanoFuse Agent Configurations

This directory contains specialized AI agent configurations for the NanoFuse project. These agents provide expert guidance for specific domains and are compatible with both Claude Code and GitHub Copilot.

## Available Agents

### MicroVM Kernel Expert (`microvm-kernel-expert.md`)

**Purpose**: Expert guidance on building, configuring, and optimizing custom Linux kernels for microVM environments.

**Core Expertise**:
- Custom kernel compilation for Firecracker and Cloud Hypervisor
- Kernel configuration optimization (minimal size, fast boot)
- VirtIO driver selection and configuration
- Cross-architecture builds (x86_64, ARM64)
- Security hardening for multi-tenant environments
- Boot process engineering and troubleshooting
- Performance tuning for microVM workloads

**When to Use**:
- Building custom kernels for Firecracker
- Debugging boot failures or device detection issues
- Optimizing kernel size (<10MB target)
- Reducing boot time (<500ms target)
- Cross-compiling for different architectures
- Security hardening for production deployments

**Example Use Cases**:
```bash
# Building minimal production kernel
"Build a minimal 6.1 kernel for Firecracker on x86_64 with <8MB target size"

# Troubleshooting
"Debug why /dev/vda is not found during boot on ARM64"

# Performance optimization
"Optimize kernel boot time from 2 seconds to <500ms"

# Cross-compilation
"Cross-compile 5.10 kernel for ARM64 from x86_64 host"
```

**Key Resources Referenced**:
- Linux Kernel Documentation (kernel.org)
- KVM API Documentation
- Firecracker GitHub and Documentation
- Slicer Documentation
- VirtIO Specification

---

## How to Use These Agents

### With Claude Code

If you're using Claude Code (claude.ai/code), simply reference the agent in your prompts:

```
@microvm-kernel-expert Help me build a minimal kernel for Firecracker
```

Or ask Claude Code to use the agent:

```
Use the microVM kernel expert agent to help me debug why my kernel won't boot
```

### With GitHub Copilot

GitHub Copilot can reference these files as context. To maximize effectiveness:

1. **Open the agent file** you want to consult
2. **Reference it in comments** in your code:
   ```go
   // Using guidance from .agents/microvm-kernel-expert.md
   // Building minimal kernel with VirtIO support
   ```

3. **Ask questions in comments**:
   ```bash
   # Following microvm-kernel-expert.md recommendations
   # What CONFIG options do I need for Firecracker on ARM64?
   ```

### With Other AI Tools

These markdown files are structured to be readable by any AI assistant:

- **ChatGPT**: Upload the file or paste relevant sections
- **Claude**: Upload as a project file or paste context
- **Other tools**: Reference as documentation

---

## Agent Development Guidelines

When creating new agents for this project:

### Structure

1. **Clear Purpose Statement**: What does this agent specialize in?
2. **Expertise Areas**: Specific domains of knowledge
3. **Knowledge Base**: Detailed technical information organized by topic
4. **Workflows**: Step-by-step procedures for common tasks
5. **Troubleshooting**: Common issues and solutions
6. **Best Practices**: DO/DON'T lists
7. **Quick Reference**: Commands and snippets
8. **Resources**: Links and references

### Format Requirements

- **Markdown**: Use GitHub-flavored markdown
- **Code Blocks**: Include language hints (\`\`\`bash, \`\`\`go, etc.)
- **Sections**: Use headers (##, ###) for clear organization
- **Examples**: Provide concrete, runnable examples
- **Cross-References**: Link to other agents where relevant

### Compatibility

Ensure agents work with:
- ✓ Claude Code
- ✓ GitHub Copilot
- ✓ Standalone reference (human-readable)
- ✓ Other AI assistants (portable context)

### Content Guidelines

**DO**:
- Provide specific, actionable guidance
- Include concrete examples and commands
- Document edge cases and gotchas
- Reference authoritative sources
- Keep information current and tested
- Use clear, concise language

**DON'T**:
- Include speculative or untested information
- Use overly broad generalizations
- Copy documentation verbatim without context
- Forget to update when tools/versions change
- Make assumptions about user's environment

---

## Contributing

When adding new agents:

1. **File Naming**: Use kebab-case (e.g., `microvm-kernel-expert.md`)
2. **Documentation**: Update this README with the new agent
3. **Testing**: Verify the agent provides useful guidance
4. **Review**: Ensure accuracy of technical content

---

## Future Agents

Potential agents to develop for NanoFuse:

- **Firecracker Runtime Expert**: Firecracker API, jailer, runtime configuration
- **RootFS Builder**: Custom rootfs creation, systemd configuration, Dockerfile optimization
- **Networking Expert**: microVM networking, vsock, tap devices, bridge configuration
- **Storage Expert**: VirtIO block devices, filesystem optimization, snapshot management
- **Security Hardening**: Seccomp, namespaces, cgroups, multi-tenancy isolation
- **Performance Tuning**: CPU pinning, memory optimization, I/O scheduling
- **CI/CD Integration**: GitHub Actions workflows, multi-arch builds, testing
- **API Service Design**: Go service development, systemd integration, API design

---

## Version History

**v1.0.0** (2025-10-30)
- Initial release
- Added MicroVM Kernel Expert agent

---

*These agent configurations are maintained as part of the NanoFuse project and should be kept in sync with project requirements and technology updates.*
