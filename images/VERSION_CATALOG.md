# NanoFuse Image Version Catalog

This document describes the versioning and tagging strategy for NanoFuse base images published to GHCR.

## Image Registry

All NanoFuse images are published to GitHub Container Registry (GHCR) under:
```
ghcr.io/daax-dev/nanofuse/base
```

**Access**: Private repository - requires authentication with GitHub token (read:packages scope)

## Authentication

Before pulling images, authenticate with GHCR:

```bash
# Option 1: Interactive login
docker login ghcr.io

# Option 2: Using GitHub token (recommended for automation)
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

**Create token**: https://github.com/settings/tokens/new?scopes=read:packages

## Available Tags

### Latest (Rolling)
- `latest` - Always points to the most recent build from main branch
- **Use case**: Development, testing, getting latest features
- **Update frequency**: Every push to main
- **Stability**: May include breaking changes

```bash
# Pull latest
nanofuse image pull --default
# or
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:latest
# or shorthand
nanofuse vm run default my-vm
```

### Versioned Releases (Stable)
- `1.0.0`, `1.0.1`, etc. - Semantic versioned releases
- **Use case**: Production deployments, reproducible builds
- **Update frequency**: Manual releases via `./image-release.sh`
- **Stability**: Immutable, tested, documented changes

```bash
# Pull specific version
nanofuse image pull --default --tag 1.0.0
# or
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:1.0.0
# or shorthand
nanofuse vm run default:1.0.0 my-vm
```

### Commit-based (Reproducible)
- `sha-abc1234` - Build from specific commit
- **Use case**: Reproducible builds, debugging, rollback
- **Update frequency**: Every commit to main
- **Stability**: Immutable, exact commit state

```bash
# Pull by commit SHA
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:sha-abc1234
```

### Branch-based
- `main` - Latest build from main branch (alias for `latest`)
- `feature-xyz` - Builds from feature branches (if enabled)
- **Use case**: Development, CI/CD testing
- **Update frequency**: Every push to branch
- **Stability**: Development quality

```bash
# Pull from main branch
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:main
```

## CLI Shortcuts

The `nanofuse` CLI provides convenient shortcuts:

| Shortcut | Expands to | Description |
|----------|------------|-------------|
| `default` | `ghcr.io/daax-dev/nanofuse/base:latest` | Default base image (latest) |
| `default:1.0.0` | `ghcr.io/daax-dev/nanofuse/base:1.0.0` | Specific version |
| `base` | `ghcr.io/daax-dev/nanofuse/base:latest` | Alias for default |
| `base:1.0.0` | `ghcr.io/daax-dev/nanofuse/base:1.0.0` | Base with version |

## Image Labels

All images include OCI-compliant labels for metadata:

### Standard OCI Labels
- `org.opencontainers.image.version` - Image version (e.g., "1.0.0", "latest")
- `org.opencontainers.image.created` - Build timestamp (RFC 3339)
- `org.opencontainers.image.revision` - Git commit SHA
- `org.opencontainers.image.source` - Source repository URL
- `org.opencontainers.image.title` - "NanoFuse Base Image"
- `org.opencontainers.image.description` - Image description
- `org.opencontainers.image.licenses` - "MIT"

### NanoFuse-specific Labels
- `com.nanofuse.base-os` - "ubuntu:24.04"
- `com.nanofuse.kernel.version` - Kernel version (e.g., "5.10.204")
- `com.nanofuse.kernel.source` - "Firecracker CI"
- `com.nanofuse.architecture` - "x86_64"
- `com.nanofuse.type` - "base"
- `com.nanofuse.firecracker.compatible` - "true"
- `com.nanofuse.services.ssh` - "enabled"
- `com.nanofuse.services.systemd-networkd` - "enabled"

### Inspect Image Labels

```bash
# View all labels
docker pull ghcr.io/daax-dev/nanofuse/base:latest
docker inspect ghcr.io/daax-dev/nanofuse/base:latest | jq '.[0].Config.Labels'

# Or via nanofuse CLI
nanofuse image inspect default
```

## Version History

### Latest Releases

| Version | Date | Changes | Notes |
|---------|------|---------|-------|
| `latest` | Rolling | Latest from main | Development builds |
| `1.0.0` | TBD | Initial stable release | Ubuntu 24.04, systemd, SSH |

> **Note**: Versioned releases will be added here as they are created via `./image-release.sh`

## Creating New Releases

### For Maintainers

**Create a new versioned image release:**
```bash
./image-release.sh
```

This will:
1. Determine next version (auto-increment patch)
2. Create and push image tag (e.g., `image-v1.0.1`)
3. Trigger CI to build and push with version labels
4. Update GHCR with new versioned tag

**Manual version specification:**
```bash
git tag image-v1.1.0 -m "Release 1.1.0: Add new features"
git push origin image-v1.1.0
```

## Best Practices

### For Users

✅ **DO:**
- Use versioned tags (`1.0.0`) for production
- Use `latest` for development and testing
- Pin specific versions in automation/scripts
- Review changelog before upgrading

❌ **DON'T:**
- Use `latest` in production (unpredictable updates)
- Assume tags are mutable (they are immutable once pushed)
- Skip authentication (all images are private)

### For CI/CD

```yaml
# Good: Pinned version
image: ghcr.io/daax-dev/nanofuse/base:1.0.0

# Acceptable: Use variable for flexibility
image: ghcr.io/daax-dev/nanofuse/base:${NANOFUSE_VERSION:-1.0.0}

# Risky: Latest tag (unpredictable)
image: ghcr.io/daax-dev/nanofuse/base:latest
```

## Catalog API

Future enhancement: Add catalog endpoint to query available versions

```bash
# Planned feature
nanofuse image catalog
# Would output:
# Available versions for ghcr.io/daax-dev/nanofuse/base:
#   latest (updated: 2025-11-03)
#   1.0.1 (released: 2025-11-01)
#   1.0.0 (released: 2025-10-15)
```

## Migration Guide

### Upgrading Between Versions

When upgrading to a new base image version:

1. **Check release notes** for breaking changes
2. **Test in development** environment first
3. **Pull new version**:
   ```bash
   nanofuse image pull --default --tag 1.1.0
   ```
4. **Update VM configurations** if needed
5. **Create new VMs** with new image:
   ```bash
   nanofuse vm create my-vm --image default:1.1.0
   ```
6. **Verify functionality** before production rollout

### Rollback

If issues occur after upgrade:
```bash
# Revert to previous version
nanofuse vm create my-vm --image default:1.0.0
```

## Support

- 📚 Documentation: https://github.com/daax-dev/nanofuse
- 🐛 Issues: https://github.com/daax-dev/nanofuse/issues
- 💬 Discussions: https://github.com/daax-dev/nanofuse/discussions
