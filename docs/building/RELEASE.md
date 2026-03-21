# Container Image Release Strategy Guide

## Overview

This document provides a **comprehensive, step-by-step guide** for implementing the container image release strategy used in this repository. This strategy is designed for **SLSA v1.1 Build Level 3 compliance** and includes automatic version bumping, semantic versioning, multi-architecture builds, security scanning, attestation generation, and artifact signing.

**Target Audience**: Other repositories that want to copy this release strategy **exactly**.

---

## Table of Contents

1. [Core Principles](#core-principles)
2. [Release Strategy Overview](#release-strategy-overview)
3. [Prerequisites](#prerequisites)
4. [Step-by-Step Implementation Guide](#step-by-step-implementation-guide)
5. [Version Bumping Process](#version-bumping-process)
6. [Git Tagging and Release Workflow](#git-tagging-and-release-workflow)
7. [GitHub Actions Workflow Configuration](#github-actions-workflow-configuration)
8. [Dockerfile Configuration](#dockerfile-configuration)
9. [Testing Your Release Process](#testing-your-release-process)
10. [Troubleshooting](#troubleshooting)
11. [Complete Example](#complete-example)

---

## Core Principles

This release strategy adheres to the following constitutional principles:

### ✅ **Principle III: SLSA v1.1 Compliance**
- Build once, deploy everywhere
- Content-addressable storage (SHA256 digests only)
- Full provenance attestation and SBOM
- **No mutable tags** (`:latest` is only for convenience, always use digest in production)

### ✅ **Principle V: Automation-Driven Repeatability**
- Version bumping is **manual but explicit**
- Tagging is **manual via git tags**
- Building, signing, and attestation are **fully automated**
- No human intervention after pushing a git tag

### ✅ **Principle VI: Security by Architectural Design**
- All images signed with Cosign (keyless OIDC)
- SBOM generation with Syft
- Security scanning with Trivy
- SLSA provenance attestation

---

## Release Strategy Overview

### High-Level Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Developer Updates Dockerfile Version                         │
│    LABEL org.opencontainers.image.version="1.2.3"               │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Developer Commits with [release] Comment                     │
│    git commit -m "release: bump version to 1.2.3 [release]"     │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Developer Creates Git Tag                                    │
│    git tag -a v1.2.3 -m "Release v1.2.3"                        │
│    git push origin v1.2.3                                       │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. GitHub Actions Workflow Triggers Automatically               │
│    - Detects git tag matching 'v*' pattern                      │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Automated Build Process                                      │
│    ├── Multi-arch build (linux/amd64, linux/arm64)              │
│    ├── Generate tags:                                           │
│    │   ├── 1.2.3 (semver full)                                  │
│    │   ├── 1.2 (semver major.minor)                             │
│    │   ├── sha-<commit-sha> (commit digest)                     │
│    │   └── latest (only on main branch)                         │
│    ├── Push to GitHub Container Registry (ghcr.io)              │
│    ├── Sign with Cosign (keyless OIDC)                          │
│    ├── Generate SBOM with Syft                                  │
│    ├── Attest SBOM                                              │
│    ├── Generate SLSA provenance                                 │
│    └── Scan with Trivy                                          │
└─────────────────────────────────────────────────────────────────┘
```

### Key Features

1. **Manual Version Bumping**: Developers explicitly update the version in the Dockerfile
2. **Explicit Release Commits**: Commits include `[release]` marker for clarity
3. **Git Tag-Driven**: Git tags trigger the automated release workflow
4. **Semantic Versioning**: Full support for semver (major.minor.patch)
5. **Multi-Architecture**: Builds for both `linux/amd64` and `linux/arm64`
6. **SLSA L3 Compliance**: Full provenance, SBOM, and signing
7. **Immutable References**: Production deployments use SHA256 digests

---

## Prerequisites

Before implementing this strategy, ensure you have:

### Repository Setup

- [ ] GitHub repository with GitHub Actions enabled
- [ ] GitHub Container Registry (ghcr.io) access
- [ ] Repository secrets configured (automatically available via `GITHUB_TOKEN`)

### Permissions Required

The GitHub Actions workflow needs the following permissions (automatically configured in workflow):

```yaml
permissions:
  contents: read          # Read repository contents
  packages: write         # Push to GitHub Container Registry
  id-token: write         # OIDC token for Cosign keyless signing
  attestations: write     # Create SLSA provenance attestations
```

### Tools Used (Pre-installed in GitHub Actions)

- Docker Buildx (multi-arch builds)
- Cosign (image signing)
- Syft (SBOM generation)
- Trivy (security scanning)

---

## Step-by-Step Implementation Guide

### Step 1: Create the Dockerfile

Your Dockerfile **must** include a version label. This is the **single source of truth** for your image version.

**File Location**: `infrastructure/docker/<image-name>/Dockerfile`

**Required Labels** (add these at the end of your Dockerfile):

```dockerfile
# Labels for metadata
LABEL org.opencontainers.image.title="<Your Image Title>"
LABEL org.opencontainers.image.description="<Your Image Description>"
LABEL org.opencontainers.image.vendor="<Your Organization>"
LABEL org.opencontainers.image.source="https://github.com/<org>/<repo>"
LABEL org.opencontainers.image.documentation="https://github.com/<org>/<repo>/blob/main/docs/<your-docs>.md"
LABEL org.opencontainers.image.version="1.0.0"

# Constitutional compliance labels (optional but recommended)
LABEL netcon.constitutional.principle.iii="SLSA v1.1 Compliance"
LABEL netcon.constitutional.principle.v="Automation-Driven Repeatability"
LABEL netcon.constitutional.principle.vi="Security by Architectural Design"
```

**CRITICAL**: The `org.opencontainers.image.version` label is the version that will be bumped for each release.

**Example**:

```dockerfile
FROM ubuntu:24.04@sha256:66460d557b25769b102175144d538d88219c077c678a49af4afca6fbfc1b5252

# ... your build steps ...

EXPOSE 8080
CMD ["./myapp"]

# Labels for metadata (REQUIRED)
LABEL org.opencontainers.image.title="My Application"
LABEL org.opencontainers.image.description="A sample application with CI/CD"
LABEL org.opencontainers.image.vendor="ACME Corp"
LABEL org.opencontainers.image.source="https://github.com/acme/myapp"
LABEL org.opencontainers.image.documentation="https://github.com/acme/myapp/blob/main/README.md"
LABEL org.opencontainers.image.version="1.0.0"

# Constitutional compliance labels (optional)
LABEL netcon.constitutional.principle.iii="SLSA v1.1 Compliance"
LABEL netcon.constitutional.principle.v="Automation-Driven Repeatability"
LABEL netcon.constitutional.principle.vi="Security by Architectural Design"
```

---

### Step 2: Create the GitHub Actions Workflow

**File Location**: `.github/workflows/build-<image-name>.yml`

Copy this **complete** workflow file (adjust the comments and paths as needed):

```yaml
# GitHub Actions Workflow: Build Container Image
# Builds multi-arch container image with SLSA provenance and SBOM
# Constitutional Compliance: Principles III (SLSA), V (Automation), VI (Security)

name: Build <Image Name>

on:
  push:
    branches:
      - main
    paths:
      - 'infrastructure/docker/<image-name>/**'
      - '.github/workflows/build-<image-name>.yml'
    tags:
      - 'v*'  # Trigger on version tags (v1.0.0, v1.2.3, etc.)
  pull_request:
    paths:
      - 'infrastructure/docker/<image-name>/**'
      - '.github/workflows/build-<image-name>.yml'
  schedule:
    # Weekly rebuild for security updates (every Monday at 02:00 UTC)
    - cron: '0 2 * * 1'
  workflow_dispatch:  # Manual trigger

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/<image-name>

permissions:
  contents: read
  packages: write
  id-token: write  # OIDC for Cosign signing
  attestations: write  # SLSA provenance

jobs:
  build-and-push:
    name: Build and Push Container Image
    runs-on: ubuntu-latest
    steps:
      # Step 1: Checkout code
      - name: Checkout repository
        uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

      # Step 2: Set up QEMU for multi-arch builds
      - name: Set up QEMU
        uses: docker/setup-qemu-action@49b3bc8e6bdd4a60e6116a5414239cba5943d3cf  # v3.2.0

      # Step 3: Set up Docker Buildx
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@c47758b77c9736f4b2ef4073d4d51994fabfe349  # v3.7.1

      # Step 4: Log in to GitHub Container Registry
      - name: Log in to GitHub Container Registry
        uses: docker/login-action@9780b0c442fbb1117ed29e0efdff1e18412f7567  # v3.3.0
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      # Step 5: Extract metadata (tags, labels)
      - name: Extract Docker metadata
        id: meta
        uses: docker/metadata-action@8e5442c4ef9f78752691e2d8f8d19755c6f78e81  # v5.5.1
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix=sha-
            type=raw,value=latest,enable={{is_default_branch}}

      # Step 6: Build and push Docker image
      - name: Build and push Docker image
        id: build
        uses: docker/build-push-action@4f58ea79222b3b9dc2c8bbdd6debcef730109a75  # v6.9.0
        with:
          context: infrastructure/docker/<image-name>
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      # Step 7: Install Cosign
      - name: Install Cosign
        uses: sigstore/cosign-installer@dc72c7d5c4d10cd6bcb8cf6e3fd625a9e5e537da  # v3.7.0

      # Step 8: Sign image with Cosign (keyless, OIDC)
      - name: Sign image with Cosign
        if: github.event_name != 'pull_request'
        env:
          COSIGN_EXPERIMENTAL: "true"
          IMAGE_DIGEST: ${{ steps.build.outputs.digest }}
          REPO: ${{ github.repository }}
          WORKFLOW: ${{ github.workflow }}
          REF: ${{ github.sha }}
          ACTOR: ${{ github.actor }}
        run: |
          cosign sign --yes \
            -a "repo=${REPO}" \
            -a "workflow=${WORKFLOW}" \
            -a "ref=${REF}" \
            -a "actor=${ACTOR}" \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${IMAGE_DIGEST}

      # Step 9: Install Syft
      - name: Install Syft
        uses: anchore/sbom-action/download-syft@fd74a6fb98a204a1ad35bbfae0122c1a302ff88b  # v0.15.0

      # Step 10: Generate SBOM
      - name: Generate SBOM with Syft
        if: github.event_name != 'pull_request'
        env:
          IMAGE_DIGEST: ${{ steps.build.outputs.digest }}
        run: |
          syft ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${IMAGE_DIGEST} \
            -o cyclonedx-json=sbom.json

      # Step 11: Attest SBOM (SLSA)
      - name: Attest SBOM
        if: github.event_name != 'pull_request'
        continue-on-error: true  # Optional: requires GitHub Enterprise for private repos
        uses: actions/attest-sbom@10926c72720ffc3f7b666661c8e55b1344e2a365  # v2
        with:
          subject-name: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          subject-digest: ${{ steps.build.outputs.digest }}
          sbom-path: sbom.json
          push-to-registry: true

      # Step 12: Generate SLSA provenance
      - name: Generate SLSA provenance
        if: github.event_name != 'pull_request'
        continue-on-error: true  # Optional: requires GitHub Enterprise for private repos
        uses: actions/attest-build-provenance@96b4a1ef7235a096b17240c259729fdd70c83d45  # v2
        with:
          subject-name: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          subject-digest: ${{ steps.build.outputs.digest }}
          push-to-registry: true

      # Step 13: Install Trivy
      - name: Install Trivy
        run: |
          wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
          echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee /etc/apt/sources.list.d/trivy.list
          sudo apt-get update
          sudo apt-get install -y trivy

      # Step 14: Scan image with Trivy
      - name: Scan image with Trivy
        env:
          IMAGE_DIGEST: ${{ steps.build.outputs.digest }}
        run: |
          trivy image \
            --severity HIGH,CRITICAL \
            --exit-code 0 \
            --no-progress \
            --format sarif \
            --output trivy-results.sarif \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${IMAGE_DIGEST}

      # Step 15: Upload Trivy results
      - name: Upload Trivy results
        if: always()
        continue-on-error: true  # Optional: requires Advanced Security
        uses: github/codeql-action/upload-sarif@662472033e021d55d94146f66f6058822b0b39fd  # v3.27.0
        with:
          sarif_file: trivy-results.sarif

      # Step 16: Verify SLSA provenance (Build L3 requirement)
      - name: Verify SLSA provenance
        if: github.event_name != 'pull_request'
        continue-on-error: true  # Optional: only works if attestation succeeded
        env:
          IMAGE_DIGEST: ${{ steps.build.outputs.digest }}
        run: |
          echo "Verifying SLSA provenance attestation..."
          cosign verify-attestation \
            --type slsaprovenance \
            --certificate-identity-regexp "https://github.com/${GITHUB_REPOSITORY}/.+" \
            --certificate-oidc-issuer https://token.actions.githubusercontent.com \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${IMAGE_DIGEST}
          echo "✅ Provenance verification successful"

      # Step 17: Output image digest
      - name: Output image digest
        env:
          IMAGE_DIGEST: ${{ steps.build.outputs.digest }}
        run: |
          echo "Image digest: ${IMAGE_DIGEST}"
          echo "Full image reference: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${IMAGE_DIGEST}"
          echo ""
          echo "Update your Kubernetes deployment with:"
          echo "image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${IMAGE_DIGEST}"
```

**CRITICAL MODIFICATIONS YOU MUST MAKE**:

1. Replace `<image-name>` with your actual image name (e.g., `github-runner`, `api-server`)
2. Replace `<Image Name>` in the workflow name with a human-readable name
3. Adjust the `paths` to match your Dockerfile location
4. Update `IMAGE_NAME` environment variable to match your desired image path

---

### Step 3: Configure Repository Settings (Optional but Recommended)

#### Branch Protection Rules

Configure branch protection for `main`:

1. Go to **Settings** → **Branches** → **Add rule**
2. Branch name pattern: `main`
3. Enable:
   - ✅ Require a pull request before merging
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Select your build workflow as a required check

#### Tag Protection Rules

Configure tag protection to prevent accidental tag deletion:

1. Go to **Settings** → **Tags** → **Add rule**
2. Tag name pattern: `v*`
3. Enable tag protection

---

## Version Bumping Process

### When to Bump Versions

Follow **Semantic Versioning (semver)** principles:

- **MAJOR** version (`1.0.0` → `2.0.0`): Breaking changes, incompatible API changes
- **MINOR** version (`1.0.0` → `1.1.0`): New features, backward-compatible
- **PATCH** version (`1.0.0` → `1.0.1`): Bug fixes, backward-compatible

### How to Bump the Version

#### Step-by-Step Process

1. **Identify the version to bump to** based on your changes (e.g., `1.2.3`)

2. **Update the Dockerfile version label**:

   ```bash
   # Edit the Dockerfile
   vim infrastructure/docker/<image-name>/Dockerfile
   ```

   Change:
   ```dockerfile
   LABEL org.opencontainers.image.version="1.0.0"
   ```

   To:
   ```dockerfile
   LABEL org.opencontainers.image.version="1.2.3"
   ```

3. **Commit the version bump with a clear message**:

   ```bash
   git add infrastructure/docker/<image-name>/Dockerfile
   git commit -m "release: bump <image-name> version to 1.2.3 [release]"
   ```

   **CRITICAL**: Include `[release]` in your commit message for clarity and traceability.

4. **Push to the main branch** (or create a PR):

   ```bash
   git push origin main
   ```

   OR if using PRs:

   ```bash
   git push origin your-feature-branch
   # Create PR, get approval, merge to main
   ```

---

## Git Tagging and Release Workflow

### Creating a Release Tag

After your version bump commit is merged to `main`:

1. **Pull the latest main branch**:

   ```bash
   git checkout main
   git pull origin main
   ```

2. **Create an annotated git tag** (MUST start with `v`):

   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3: <brief description of changes>"
   ```

   **Examples**:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3: Add support for multi-tenancy"
   git tag -a v2.0.0 -m "Release v2.0.0: Breaking API changes for v2"
   git tag -a v1.0.1 -m "Release v1.0.1: Fix memory leak in worker pool"
   ```

3. **Push the tag to trigger the workflow**:

   ```bash
   git push origin v1.2.3
   ```

   **CRITICAL**: This push triggers the automated build, sign, and attest workflow.

### Tag Naming Convention

**MUST follow**: `v<MAJOR>.<MINOR>.<PATCH>`

✅ **Valid Tags**:
- `v1.0.0`
- `v1.2.3`
- `v2.0.0-rc.1` (pre-release)
- `v1.0.1-beta.2` (beta)

❌ **Invalid Tags**:
- `1.0.0` (missing `v` prefix)
- `release-1.0.0` (wrong prefix)
- `v1.0` (incomplete semver)

---

## GitHub Actions Workflow Configuration

### Understanding the Workflow Triggers

The workflow triggers on:

```yaml
on:
  push:
    branches:
      - main
    paths:
      - 'infrastructure/docker/<image-name>/**'
      - '.github/workflows/build-<image-name>.yml'
    tags:
      - 'v*'  # THIS IS THE RELEASE TRIGGER
  pull_request:
    paths:
      - 'infrastructure/docker/<image-name>/**'
      - '.github/workflows/build-<image-name>.yml'
  schedule:
    - cron: '0 2 * * 1'  # Weekly security rebuild
  workflow_dispatch:  # Manual trigger
```

**Trigger Breakdown**:

1. **Push to `main`** with Dockerfile changes:
   - Builds and pushes with branch tags
   - Tags: `main`, `sha-<commit>`

2. **Push git tag `v*`** (e.g., `v1.2.3`):
   - **THIS IS YOUR RELEASE TRIGGER**
   - Builds and pushes with semver tags
   - Tags: `1.2.3`, `1.2`, `sha-<commit>`, `latest`

3. **Pull Request**:
   - Builds but does NOT push
   - Validates build succeeds

4. **Schedule** (weekly):
   - Rebuilds for security updates
   - Pushes with branch tags

5. **Manual** (`workflow_dispatch`):
   - On-demand builds

### Understanding the Tagging Strategy

The `docker/metadata-action` generates the following tags:

```yaml
tags: |
  type=ref,event=branch          # Branch name (e.g., main)
  type=ref,event=pr              # PR number (e.g., pr-123)
  type=semver,pattern={{version}} # Full semver (e.g., 1.2.3)
  type=semver,pattern={{major}}.{{minor}} # Major.minor (e.g., 1.2)
  type=sha,prefix=sha-           # Commit SHA (e.g., sha-abc123)
  type=raw,value=latest,enable={{is_default_branch}} # latest on main
```

**Example Tag Output for `git tag v1.2.3` on main**:

```
ghcr.io/org/repo/image-name:1.2.3
ghcr.io/org/repo/image-name:1.2
ghcr.io/org/repo/image-name:sha-abc123def
ghcr.io/org/repo/image-name:latest
```

**CRITICAL**: For production deployments, **ALWAYS use the SHA256 digest**, not tags:

```yaml
# ✅ CORRECT (immutable, SLSA-compliant)
image: ghcr.io/org/repo/image-name@sha256:abc123...

# ❌ INCORRECT (mutable, not SLSA-compliant)
image: ghcr.io/org/repo/image-name:1.2.3
image: ghcr.io/org/repo/image-name:latest
```

---

## Dockerfile Configuration

### Required Dockerfile Structure

Your Dockerfile must include specific base image references and labels.

#### Base Image Requirements

**ALWAYS** use digest-pinned base images for SLSA compliance:

```dockerfile
# ✅ CORRECT: Digest-pinned base image
FROM ubuntu:24.04@sha256:66460d557b25769b102175144d538d88219c077c678a49af4afca6fbfc1b5252

# ❌ INCORRECT: Mutable tag
FROM ubuntu:24.04
FROM ubuntu:latest
```

#### Required Labels

At minimum, include these labels:

```dockerfile
LABEL org.opencontainers.image.title="<Your Image Title>"
LABEL org.opencontainers.image.description="<Description>"
LABEL org.opencontainers.image.vendor="<Organization>"
LABEL org.opencontainers.image.source="https://github.com/<org>/<repo>"
LABEL org.opencontainers.image.version="1.0.0"  # THIS IS CRITICAL
```

#### Complete Dockerfile Example

```dockerfile
# Multi-stage build example
FROM golang:1.23@sha256:abc123... AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19@sha256:def456...

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/root/server", "healthcheck"] || exit 1

# Required labels
LABEL org.opencontainers.image.title="My API Server"
LABEL org.opencontainers.image.description="REST API server with authentication"
LABEL org.opencontainers.image.vendor="ACME Corp"
LABEL org.opencontainers.image.source="https://github.com/acme/api-server"
LABEL org.opencontainers.image.documentation="https://github.com/acme/api-server/blob/main/README.md"
LABEL org.opencontainers.image.version="1.0.0"

# Constitutional compliance labels
LABEL netcon.constitutional.principle.iii="SLSA v1.1 Compliance"
LABEL netcon.constitutional.principle.v="Automation-Driven Repeatability"
LABEL netcon.constitutional.principle.vi="Security by Architectural Design"

CMD ["./server"]
```

---

## Testing Your Release Process

### Pre-Release Checklist

Before creating your first release:

- [ ] Dockerfile exists with correct version label
- [ ] GitHub Actions workflow file exists
- [ ] Workflow file paths match Dockerfile location
- [ ] Base images use digest pins
- [ ] Required labels are present

### Test the Workflow Locally (Optional)

Use [act](https://github.com/nektos/act) to test workflows locally:

```bash
# Install act
brew install act  # macOS
# OR
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash  # Linux

# Test the workflow
act push -j build-and-push --secret GITHUB_TOKEN=<your-token>
```

### Test with a Pre-Release Tag

Create a pre-release to test:

```bash
# Bump to pre-release version
vim infrastructure/docker/<image-name>/Dockerfile
# Change: LABEL org.opencontainers.image.version="1.0.0-rc.1"

git add infrastructure/docker/<image-name>/Dockerfile
git commit -m "release: bump to 1.0.0-rc.1 [release]"
git push origin main

git tag -a v1.0.0-rc.1 -m "Release Candidate 1"
git push origin v1.0.0-rc.1
```

Monitor the workflow in GitHub Actions:
1. Go to **Actions** tab
2. Find your workflow run
3. Verify all steps complete successfully
4. Check the "Output image digest" step for the final digest

### Verify the Release

1. **Check GitHub Container Registry**:

   ```bash
   # List tags
   docker pull ghcr.io/<org>/<repo>/<image-name>:1.0.0-rc.1

   # Verify digest
   docker images --digests ghcr.io/<org>/<repo>/<image-name>
   ```

2. **Verify Cosign signature**:

   ```bash
   cosign verify \
     --certificate-identity-regexp "https://github.com/<org>/<repo>/.+" \
     --certificate-oidc-issuer https://token.actions.githubusercontent.com \
     ghcr.io/<org>/<repo>/<image-name>@sha256:<digest>
   ```

3. **Verify SBOM**:

   ```bash
   cosign verify-attestation \
     --type cyclonedx \
     --certificate-identity-regexp "https://github.com/<org>/<repo>/.+" \
     --certificate-oidc-issuer https://token.actions.githubusercontent.com \
     ghcr.io/<org>/<repo>/<image-name>@sha256:<digest>
   ```

4. **Verify SLSA provenance**:

   ```bash
   cosign verify-attestation \
     --type slsaprovenance \
     --certificate-identity-regexp "https://github.com/<org>/<repo>/.+" \
     --certificate-oidc-issuer https://token.actions.githubusercontent.com \
     ghcr.io/<org>/<repo>/<image-name>@sha256:<digest>
   ```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: Workflow doesn't trigger on tag push

**Symptoms**: No workflow run appears in GitHub Actions after pushing a tag.

**Solution**:
1. Verify tag format: Must start with `v` (e.g., `v1.0.0`)
2. Check workflow trigger configuration:
   ```yaml
   on:
     push:
       tags:
         - 'v*'  # Matches v1.0.0, v2.1.3, etc.
   ```
3. Ensure workflow file is on the `main` branch
4. Push the tag from the correct branch:
   ```bash
   git checkout main
   git pull origin main
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

#### Issue 2: Docker build fails with permission denied

**Symptoms**: `permission denied while trying to connect to the Docker daemon socket`

**Solution**: This should not occur in GitHub Actions. If it does, verify:
1. You're using the latest `docker/setup-buildx-action`
2. The runner has Docker installed (should be automatic)

#### Issue 3: Cosign signing fails

**Symptoms**: `failed to sign image: unauthorized: authentication required`

**Solution**:
1. Verify `permissions` in workflow:
   ```yaml
   permissions:
     id-token: write  # Required for OIDC
     packages: write  # Required for GHCR
   ```
2. Ensure you're using keyless signing with OIDC:
   ```yaml
   env:
     COSIGN_EXPERIMENTAL: "true"
   ```

#### Issue 4: SLSA attestation fails on private repos

**Symptoms**: `attestation creation failed: requires GitHub Enterprise`

**Solution**: SLSA attestations are only available on:
- Public repositories (free)
- GitHub Enterprise repositories

For private repos without Enterprise, set `continue-on-error: true`:
```yaml
- name: Generate SLSA provenance
  continue-on-error: true  # Skip on private repos without Enterprise
```

#### Issue 5: Version in Dockerfile doesn't match git tag

**Symptoms**: Image is tagged with `1.2.3` but Dockerfile says `1.0.0`

**Solution**: The Dockerfile version label is **metadata only**. The actual image tags come from the git tag via `docker/metadata-action`. However, for consistency:

1. **ALWAYS** update the Dockerfile version to match the git tag
2. Create a pre-push hook to verify:

```bash
#!/bin/bash
# .git/hooks/pre-push

# Extract tag being pushed
while read local_ref local_sha remote_ref remote_sha; do
  if [[ "$local_ref" == refs/tags/v* ]]; then
    TAG_VERSION="${local_ref#refs/tags/v}"

    # Extract Dockerfile version
    DOCKERFILE_VERSION=$(grep 'org.opencontainers.image.version=' infrastructure/docker/*/Dockerfile | grep -oP '"\K[^"]+')

    if [[ "$TAG_VERSION" != "$DOCKERFILE_VERSION" ]]; then
      echo "ERROR: Git tag version ($TAG_VERSION) does not match Dockerfile version ($DOCKERFILE_VERSION)"
      exit 1
    fi
  fi
done

exit 0
```

#### Issue 6: Multi-arch build fails

**Symptoms**: `error: failed to solve: no match for platform in manifest`

**Solution**:
1. Ensure QEMU is set up:
   ```yaml
   - name: Set up QEMU
     uses: docker/setup-qemu-action@v3
   ```
2. Ensure Buildx is configured:
   ```yaml
   - name: Set up Docker Buildx
     uses: docker/setup-buildx-action@v3
   ```
3. Verify platforms in build step:
   ```yaml
   platforms: linux/amd64,linux/arm64
   ```

#### Issue 7: Trivy scan blocks the workflow

**Symptoms**: Trivy finds HIGH/CRITICAL vulnerabilities and fails

**Solution**: Trivy is configured with `--exit-code 0` to NOT block:

```yaml
trivy image \
  --severity HIGH,CRITICAL \
  --exit-code 0 \  # <-- Does not fail the workflow
  --no-progress \
  --format sarif \
  --output trivy-results.sarif \
  ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${IMAGE_DIGEST}
```

To fail on vulnerabilities, change to `--exit-code 1`.

---

## Complete Example

### Scenario: Creating v1.2.3 Release

#### Context
- Current version: `1.2.2`
- New features added, backward-compatible
- Bumping to: `1.2.3`

#### Step 1: Update Dockerfile

```bash
cd ~/myrepo
git checkout main
git pull origin main
git checkout -b release-v1.2.3

# Edit Dockerfile
vim infrastructure/docker/api-server/Dockerfile
```

Change:
```dockerfile
LABEL org.opencontainers.image.version="1.2.2"
```

To:
```dockerfile
LABEL org.opencontainers.image.version="1.2.3"
```

#### Step 2: Commit and Push

```bash
git add infrastructure/docker/api-server/Dockerfile
git commit -m "release: bump api-server version to 1.2.3 [release]"
git push origin release-v1.2.3
```

#### Step 3: Create PR and Merge

1. Create PR: `release-v1.2.3` → `main`
2. Get approval
3. Merge to `main`

#### Step 4: Create and Push Tag

```bash
git checkout main
git pull origin main

git tag -a v1.2.3 -m "Release v1.2.3: Add user profile API endpoint"
git push origin v1.2.3
```

#### Step 5: Monitor Workflow

1. Go to GitHub Actions
2. Find workflow run for `v1.2.3`
3. Monitor all steps
4. Wait for completion (typically 5-10 minutes)

#### Step 6: Verify Release

```bash
# Pull the new image by tag
docker pull ghcr.io/myorg/myrepo/api-server:1.2.3

# Get the digest
docker images --digests ghcr.io/myorg/myrepo/api-server

# Output:
# REPOSITORY                              TAG    DIGEST                                                                  ...
# ghcr.io/myorg/myrepo/api-server        1.2.3  sha256:abc123def456...                                                 ...

# Verify signature
cosign verify \
  --certificate-identity-regexp "https://github.com/myorg/myrepo/.+" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/myorg/myrepo/api-server@sha256:abc123def456...
```

#### Step 7: Update Deployment

Update Kubernetes deployment with **digest reference**:

```yaml
# k8s/api-server-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  template:
    spec:
      containers:
      - name: api-server
        # ✅ CORRECT: Use digest from Step 6
        image: ghcr.io/myorg/myrepo/api-server@sha256:abc123def456...

        # ❌ INCORRECT: Mutable tag
        # image: ghcr.io/myorg/myrepo/api-server:1.2.3
        # image: ghcr.io/myorg/myrepo/api-server:latest
```

Deploy:

```bash
kubectl apply -f k8s/api-server-deployment.yaml
kubectl rollout status deployment/api-server
```

---

## Summary Checklist

Use this checklist for every release:

### Pre-Release
- [ ] Changes committed and merged to `main`
- [ ] Version number determined (semver)
- [ ] Dockerfile version label updated
- [ ] Commit message includes `[release]`
- [ ] Changes pushed to `main` and merged

### Release
- [ ] `main` branch pulled locally
- [ ] Git tag created with `v` prefix
- [ ] Tag pushed to GitHub
- [ ] GitHub Actions workflow triggered

### Post-Release
- [ ] Workflow completed successfully
- [ ] Image digest obtained from workflow output
- [ ] Cosign signature verified
- [ ] SBOM verified
- [ ] SLSA provenance verified
- [ ] Kubernetes deployment updated with digest
- [ ] Deployment tested in staging
- [ ] Deployment promoted to production

---

## Advanced Topics

### Pre-Release and Release Candidates

For pre-releases:

```bash
# Release candidate
git tag -a v2.0.0-rc.1 -m "Release Candidate 1 for v2.0.0"

# Beta release
git tag -a v1.5.0-beta.1 -m "Beta 1 for v1.5.0"

# Alpha release
git tag -a v1.4.0-alpha.1 -m "Alpha 1 for v1.4.0"
```

These tags will generate:
- `2.0.0-rc.1` (full version)
- `sha-<commit>` (commit digest)

### Hotfix Releases

For critical bug fixes on older versions:

```bash
# Create hotfix branch from tag
git checkout v1.2.3
git checkout -b hotfix-1.2.4

# Make fixes
vim infrastructure/docker/api-server/Dockerfile
# Update version to 1.2.4

git add .
git commit -m "fix: critical security fix [hotfix]"

# Create hotfix tag
git tag -a v1.2.4 -m "Hotfix v1.2.4: Fix CVE-2024-12345"
git push origin v1.2.4

# Optionally merge back to main
git checkout main
git merge hotfix-1.2.4
git push origin main
```

### Multi-Image Repositories

For repos with multiple images:

```
.github/workflows/
  build-api-server.yml
  build-worker.yml
  build-frontend.yml

infrastructure/docker/
  api-server/Dockerfile
  worker/Dockerfile
  frontend/Dockerfile
```

Use separate workflows and tags:

```bash
# API Server release
git tag -a api-server/v1.2.3 -m "API Server v1.2.3"

# Worker release
git tag -a worker/v2.0.0 -m "Worker v2.0.0"

# Frontend release
git tag -a frontend/v3.1.0 -m "Frontend v3.1.0"
```

Adjust workflow triggers:

```yaml
on:
  push:
    tags:
      - 'api-server/v*'  # Only trigger for api-server tags
```

---

## References

### Documentation
- [SLSA v1.1 Specification](https://slsa.dev/spec/v1.1/)
- [Docker Metadata Action](https://github.com/docker/metadata-action)
- [Cosign Keyless Signing](https://docs.sigstore.dev/cosign/keyless/)
- [GitHub OIDC for Actions](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
- [Semantic Versioning](https://semver.org/)

### Tools
- [Cosign](https://github.com/sigstore/cosign)
- [Syft](https://github.com/anchore/syft)
- [Trivy](https://github.com/aquasecurity/trivy)
- [act - Local GitHub Actions](https://github.com/nektos/act)

---

**Last Updated**: 2025-11-08
**Version**: 1.0.0
**Maintained By**: Platform Engineering Team
