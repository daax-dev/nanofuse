#!/bin/bash
# Generate SSH host keys if they don't exist
# This runs before sshd on first boot to ensure unique host keys per VM

set -euo pipefail

echo "generate-ssh-keys: Ensuring SSH host keys exist..."
ssh-keygen -A
echo "generate-ssh-keys: SSH host keys ensured"
