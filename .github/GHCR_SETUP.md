# GitHub Container Registry Setup

This document explains how to configure GHCR authentication for nanofuse images.

## Overview

The CI/CD pipeline automatically builds and pushes Docker images to GitHub Container Registry (GHCR) at:
- `ghcr.io/jpoley/nanofuse/base`

**Images are PRIVATE** and require authentication to pull.

## Authentication (Required)

### Pulling Images

All images are private and require authentication:

```bash
# Option 1: Interactive login
docker login ghcr.io

# Option 2: Using GitHub token
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

Create a token with `read:packages` scope: https://github.com/settings/tokens/new?scopes=read:packages

### CI/CD Authentication

The GitHub Actions workflow automatically authenticates using `GITHUB_TOKEN`:

```yaml
- name: Log in to GitHub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

This token has `packages:write` permission by default for workflows.

## Image Tags

The CI pipeline creates the following tags:

- `latest` - Latest build from main branch
- `main` - Also points to latest main
- `sha-<commit>` - Specific commit (e.g., `sha-abc1234`)
- `v1.0.0` - Version tags (on git tags)

## Troubleshooting

### "Package not found" or "authentication required" when pulling

**Cause**: Not authenticated or package hasn't been published yet

**Solutions**:
1. Authenticate with `docker login ghcr.io` (see above)
2. Ensure you have access to the repository
3. Check if CI has run: https://github.com/jpoley/nanofuse/actions

### CI fails to push images

**Cause**: Permissions issue

**Solutions**:
1. Ensure workflow has `packages: write` permission (already configured in ci.yaml)
2. Check repository Settings → Actions → General → Workflow permissions
3. Should be set to "Read and write permissions"

## References

- [GitHub Packages Documentation](https://docs.github.com/en/packages)
- [GHCR Authentication](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
