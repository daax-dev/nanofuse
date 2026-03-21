#!/bin/bash
# NanoFuse Mage Installation and PATH Check Script
# This script ensures mage is installed and available in PATH

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if mage is in PATH
if command -v mage >/dev/null 2>&1; then
    MAGE_VERSION=$(mage -version 2>&1 | head -n 1 || echo "unknown")
    pass "Mage is installed and in PATH"
    info "Version: $MAGE_VERSION"
    info "Location: $(command -v mage)"
    exit 0
fi

# Check if mage is in ~/go/bin but not in PATH
if [ -f "$HOME/go/bin/mage" ]; then
    warn "Mage is installed at $HOME/go/bin/mage but not in PATH"
    echo
    info "Adding $HOME/go/bin to PATH for this session..."
    export PATH="$PATH:$HOME/go/bin"

    if command -v mage >/dev/null 2>&1; then
        pass "Mage is now available in PATH"
        echo
        warn "To make this permanent, add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo
        echo "  export PATH=\"\$PATH:\$HOME/go/bin\""
        echo
        exit 0
    fi
fi

# Mage is not installed
fail "Mage is not installed"
echo
info "Installing mage..."

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    fail "Go is not installed. Please install Go first: https://go.dev/doc/install"
    exit 1
fi

# Install mage
if go install github.com/magefile/mage@latest; then
    pass "Mage installed successfully"

    # Check if it's now in PATH
    if command -v mage >/dev/null 2>&1; then
        pass "Mage is available in PATH"
        MAGE_VERSION=$(mage -version 2>&1 | head -n 1 || echo "unknown")
        info "Version: $MAGE_VERSION"
        exit 0
    else
        # Installed but not in PATH
        if [ -f "$HOME/go/bin/mage" ]; then
            warn "Mage installed at $HOME/go/bin/mage but not in PATH"
            echo
            info "Adding $HOME/go/bin to PATH for this session..."
            export PATH="$PATH:$HOME/go/bin"

            if command -v mage >/dev/null 2>&1; then
                pass "Mage is now available in PATH"
                echo
                warn "To make this permanent, add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
                echo
                echo "  export PATH=\"\$PATH:\$HOME/go/bin\""
                echo
                exit 0
            fi
        fi

        fail "Mage installed but could not be found in PATH"
        echo
        warn "You may need to add $HOME/go/bin to your PATH:"
        echo
        echo "  export PATH=\"\$PATH:\$HOME/go/bin\""
        echo
        exit 1
    fi
else
    fail "Failed to install mage"
    echo
    info "You can manually install mage with:"
    echo "  go install github.com/magefile/mage@latest"
    exit 1
fi
