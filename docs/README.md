# NanoFuse Documentation

Welcome to the NanoFuse documentation! This README will help you find the right documentation for your needs.

## For End Users

Start here if you want to use NanoFuse:

- **[API_QUICK_START.md](API_QUICK_START.md)** - Quick start guide for the NanoFuse API
- **[SSH_ACCESS_QUICK_START.md](SSH_ACCESS_QUICK_START.md)** - How to access VMs via SSH
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - How to contribute to NanoFuse

### Examples

The `examples/` directory contains sample configurations:
- Network debugging configurations
- Sample VM configurations (TCP and Unix socket)

## For Developers & Contributors

### Building NanoFuse

The `building/` directory contains all development and build-related documentation:

- **[building/DEVELOPMENT_GUIDE.md](building/DEVELOPMENT_GUIDE.md)** - Comprehensive guide for developers
- **[building/FIRECRACKER_IMAGE_BUILD_GUIDE.md](building/FIRECRACKER_IMAGE_BUILD_GUIDE.md)** - How to build Firecracker images
- **[building/CLEAN-INDEX.md](building/CLEAN-INDEX.md)** - Index of all cleanup scripts
- **[building/CLEAN-SCRIPTS.md](building/CLEAN-SCRIPTS.md)** - Detailed cleanup script documentation
- **[building/CLAUDE.md](building/CLAUDE.md)** - Guidance for Claude Code when working on this project

#### Implementation Details

The `building/implementation/` directory contains detailed implementation documentation:
- Architecture decisions
- API contracts
- CLI specifications
- Testing strategies
- Integration tests
- Network implementation details

#### Planning Documents

The `building/planning/` directory contains project planning documents:
- Execution plans
- Network architecture plans
- API transport architecture
- Requirements and big ideas

#### Reports & Analysis

The `building/reports/` directory contains:
- Implementation status reports
- Build and test reports
- Debug analysis and troubleshooting guides
- Technical comparisons (e.g., NanoFuse vs K7 vs SlicerVM)
- Validation reports

## Archives

The `archives/` directory contains historical documentation:
- Phase 1 completion notes
- Legacy quick start guides
- Previous implementation notes
- Historical release notes

These are kept for reference but may be outdated.

## Documentation Structure

```
docs/
├── README.md                      # This file
├── API_QUICK_START.md            # End-user: API quick start
├── SSH_ACCESS_QUICK_START.md     # End-user: SSH access
├── CONTRIBUTING.md                # End-user: Contributing guide
├── examples/                      # Sample configurations
├── building/                      # Developer documentation
│   ├── implementation/           # Implementation details
│   ├── planning/                 # Planning documents
│   └── reports/                  # Status reports & analysis
└── archives/                      # Historical documentation
```

## Quick Links

- [GitHub Repository](https://github.com/daax-dev/nanofuse)
- [Issues](https://github.com/daax-dev/nanofuse/issues)
- [Releases](https://github.com/daax-dev/nanofuse/releases)

## Getting Help

- Check the relevant documentation section above
- Look for examples in the `examples/` directory
- Search through the `building/reports/` for troubleshooting guides
- Open an issue on GitHub
