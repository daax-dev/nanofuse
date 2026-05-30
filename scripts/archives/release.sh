#!/bin/bash
# release.sh - Create a release tag on the latest CI-validated commit
# Creates tags like: v0.0.1, v0.0.2, etc.
#
# Workflow:
# 1. Ensures you're on main and synced with origin
# 2. Verifies CI passed on the latest commit
# 3. Creates and pushes tag to trigger release build

set -euo pipefail

echo "=== NanoFuse Release Script ==="
echo ""

# Check for uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
  echo "ERROR: You have uncommitted changes. Commit and push them first."
  git status --short
  exit 1
fi

# Get current branch
current_branch=$(git rev-parse --abbrev-ref HEAD)

# Ensure we're on main
if [[ "$current_branch" != "main" ]]; then
  echo "ERROR: You must be on the 'main' branch to create a release."
  echo "Current branch: $current_branch"
  exit 1
fi

echo "✓ On main branch with no uncommitted changes"

# Fetch latest from origin (force update all refs)
echo ""
echo "Fetching latest from origin..."
git fetch origin main --prune
git fetch --tags --force

# Verify we're actually on the main branch (not detached HEAD)
if ! git symbolic-ref HEAD &>/dev/null; then
  echo "ERROR: You are in detached HEAD state!"
  echo "Checkout main first: git checkout main"
  exit 1
fi

# Get commit SHAs
LOCAL_MAIN=$(git rev-parse main)
REMOTE_MAIN=$(git rev-parse origin/main)
HEAD_COMMIT=$(git rev-parse HEAD)

echo ""
echo "=== Commit Status ==="
echo "HEAD:        $HEAD_COMMIT"
echo "main:        $LOCAL_MAIN"
echo "origin/main: $REMOTE_MAIN"
echo ""

# Verify HEAD == main (no weird state)
if [ "$HEAD_COMMIT" != "$LOCAL_MAIN" ]; then
  echo "ERROR: HEAD is not pointing to main!"
  echo "This should not happen if you're on the main branch."
  git status
  exit 1
fi

# Check if local main matches origin/main
if [ "$LOCAL_MAIN" != "$REMOTE_MAIN" ]; then
  echo "ERROR: Your local main is not in sync with origin/main"
  echo ""

  # Check if we're behind or ahead
  AHEAD_BEHIND=$(git rev-list --left-right --count main...origin/main)
  AHEAD=$(echo "$AHEAD_BEHIND" | cut -f1)
  BEHIND=$(echo "$AHEAD_BEHIND" | cut -f2)

  if [ "$BEHIND" -gt 0 ]; then
    echo "❌ You are $BEHIND commit(s) BEHIND origin/main"
    echo ""
    echo "These commits are on origin but not local:"
    git log --oneline HEAD..origin/main
    echo ""
    echo "Pull first: git pull origin main"
  fi

  if [ "$AHEAD" -gt 0 ]; then
    echo "❌ You have $AHEAD UNPUSHED commit(s)"
    echo ""
    echo "These commits are local but not on origin:"
    git log --oneline origin/main..HEAD
    echo ""
    echo "Push first: git push origin main"
    echo ""
    echo "IMPORTANT: Wait for CI to pass after pushing!"
  fi

  exit 1
fi

echo "✓ Local main == origin/main == HEAD"
echo "✓ All three point to the same commit: $HEAD_COMMIT"

# Show the commit we'll tag
echo ""
echo "Commit to be tagged:"
git log -1 --format="%h - %s (%an, %ar)" HEAD
echo ""
git log -1 --format="Full SHA: %H" HEAD

# Check CI status using gh CLI if available
if command -v gh &> /dev/null; then
  echo ""
  echo "Checking CI status..."

  # Get the status of the latest CI run for this commit
  CI_STATUS=$(gh run list --commit "$COMMIT" --limit 1 --json conclusion --jq '.[0].conclusion' 2>/dev/null || echo "unknown")

  if [ "$CI_STATUS" = "success" ]; then
    echo "✓ CI passed for this commit"
  elif [ "$CI_STATUS" = "failure" ]; then
    echo "ERROR: CI failed for this commit!"
    echo "View runs: gh run list --commit $COMMIT_SHORT"
    exit 1
  elif [ "$CI_STATUS" = "in_progress" ] || [ "$CI_STATUS" = "queued" ]; then
    echo "WARNING: CI is still running for this commit"
    echo "View runs: gh run list --commit $COMMIT_SHORT"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
    fi
  else
    echo "WARNING: Could not determine CI status (might not have run yet)"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      exit 1
    fi
  fi
else
  echo ""
  echo "WARNING: 'gh' CLI not found, cannot verify CI status"
  echo "Install: https://cli.github.com/"
  read -p "Continue without CI verification? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# Determine next version
echo ""
echo "Determining next version..."

latest_tag=$(git tag -l --sort=-v:refname 'v[0-9]*' | grep -v '^image-v' | head -n1 || true)

if [ -z "$latest_tag" ]; then
  echo "No existing tags found, creating first release: v0.0.1"
  new_version="0.0.1"
  new_tag="v$new_version"
else
  # Strip leading 'v'
  version=${latest_tag#v}
  IFS='.' read -r major minor patch <<< "$version"

  # Bump patch
  patch=$((patch + 1))
  new_version="$major.$minor.$patch"
  new_tag="v$new_version"

  echo "Latest tag: $latest_tag"
  echo "Next tag:   $new_tag"
fi

# Check if tag already exists
if git rev-parse "$new_tag" >/dev/null 2>&1; then
  echo ""
  echo "ERROR: Tag $new_tag already exists!"
  exit 1
fi

# Final confirmation
echo ""
echo "=== Release Summary ==="
echo "  Tag:    $new_tag"
echo "  Commit: $COMMIT_SHORT"
echo "  Branch: main (synced with origin)"
echo ""
echo "This will:"
echo "  1. Create annotated tag $new_tag on commit $COMMIT_SHORT"
echo "  2. Push tag to origin"
echo "  3. Trigger GitHub Actions to build and release binaries"
echo ""
read -p "Create release? (y/N) " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

# Create tag locally
echo ""
echo "Creating tag $new_tag..."
git tag -a "$new_tag" -m "Release $new_tag"

# Verify tag points to correct commit
TAG_COMMIT=$(git rev-parse "$new_tag")
if [ "$TAG_COMMIT" != "$HEAD_COMMIT" ]; then
  echo "ERROR: Tag was created but points to wrong commit!"
  echo "Expected: $HEAD_COMMIT"
  echo "Got:      $TAG_COMMIT"
  git tag -d "$new_tag"
  exit 1
fi

echo "✓ Tag created and verified locally"
echo "  Tag:    $new_tag"
echo "  Points: $TAG_COMMIT"

# Push tag to origin
echo ""
echo "Pushing tag to origin..."
if ! git push origin "$new_tag"; then
  echo ""
  echo "ERROR: Failed to push tag!"
  echo "The tag exists locally but failed to push to origin."
  echo "To clean up, run: git tag -d $new_tag"
  exit 1
fi

echo ""
echo "✓ Release $new_tag created successfully!"
echo ""
echo "=== Summary ==="
echo "  Tag:        $new_tag"
echo "  Commit:     $HEAD_COMMIT"
echo "  Pushed to:  origin"
echo ""
echo "GitHub Actions is now building the release."
echo "Monitor: https://github.com/daax-dev/nanofuse/actions"
echo "Release: https://github.com/daax-dev/nanofuse/releases/tag/$new_tag"
echo ""
echo "To verify the tag on GitHub:"
echo "  git ls-remote --tags origin | grep $new_tag"
