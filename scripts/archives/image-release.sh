#!/bin/bash
# image-release.sh - Auto-bump Docker image version, tag, and push
# Creates tags like: image-v0.0.1, image-v0.0.2, etc.
# These trigger CI to build and push Docker images to GHCR

set -euo pipefail

# Ensure we're up to date
git fetch --tags

# Get latest image tag (image-v* pattern)
latest_tag=$(git tag -l --sort=-v:refname 'image-v[0-9]*' | head -n1)

if [ -z "$latest_tag" ]; then
  echo "No Docker image tags found, creating first tag image-v0.0.1"
  new_version="0.0.1"
  new_tag="image-v$new_version"
else
  # Strip 'image-v' prefix
  version=${latest_tag#image-v}
  IFS='.' read -r major minor patch <<< "$version"

  # Bump patch
  patch=$((patch + 1))
  new_version="$major.$minor.$patch"
  new_tag="image-v$new_version"

  echo "Bumping Docker image version: $latest_tag → $new_tag"
fi

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
  echo "Error: You have uncommitted changes. Please commit or stash them first."
  exit 1
fi

# Create tag (no file changes needed for image releases)
git tag "$new_tag"
echo "Created tag: $new_tag"
echo "Pushing to origin..."
git push origin "$new_tag"

echo ""
echo "✓ Image release complete: $new_tag"
echo "  GitHub Actions will build and push to ghcr.io/jpoley/nanofuse/base:$new_version"
echo "  View workflow at: https://github.com/jpoley/nanofuse/actions"
echo ""
echo "Users can pull with:"
echo "  nanofuse image pull ghcr.io/jpoley/nanofuse/base:$new_version"
echo "  # Or use the --default flag for latest"
