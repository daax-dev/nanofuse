# Authentication Setup for Testing

## The Problem

Docker login requires interactive TTY which we don't have in this environment.

## Solutions

### Option 1: Use existing Docker credentials (if available)

Check if you already have credentials:
```bash
cat ~/.docker/config.json
```

If you see `ghcr.io` in there, you're already authenticated and can skip this.

### Option 2: Manual Docker login (you must do this)

On your actual terminal (not here):
```bash
# Create token: https://github.com/settings/tokens/new?scopes=read:packages
docker login ghcr.io
# Enter your GitHub username
# Paste the token as password
```

### Option 3: Non-interactive auth with token file

If you have a GitHub token:
```bash
# Save token to file
echo "YOUR_GITHUB_TOKEN" > /tmp/gh_token

# Login non-interactively
cat /tmp/gh_token | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# Clean up
rm /tmp/gh_token
```

### Option 4: Skip authentication for now

Test with a public image instead, or test the fixes without actually pulling images (just verify the command doesn't error with "Job ID required").

## What Actually Matters

The critical fixes we made were:
1. **CLI pull command type mismatch** - This fix makes the command work correctly
2. **Dual listener support** - Unix socket now gets created

These fixes can be verified even without pulling real images. The "Job ID required" error happened BEFORE any authentication, so we can verify the fix works.
