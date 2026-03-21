#!/bin/bash
# Install SSH public key from kernel command line
# The key is passed as base64-encoded sshkey= parameter

set -e

# Extract sshkey parameter from kernel cmdline
SSHKEY_B64=$(cat /proc/cmdline | grep -oP 'sshkey=\K[^ ]+' || true)

if [ -z "$SSHKEY_B64" ]; then
    echo "install-ssh-key: No SSH key provided via kernel cmdline"
    exit 0
fi

# Decode the key
SSH_KEY=$(echo "$SSHKEY_B64" | base64 -d 2>/dev/null)

if [ -z "$SSH_KEY" ]; then
    echo "install-ssh-key: Failed to decode SSH key"
    exit 1
fi

# Validate it looks like a public key
if ! echo "$SSH_KEY" | grep -qE '^(ssh-|ecdsa-)'; then
    echo "install-ssh-key: Invalid SSH key format"
    exit 1
fi

# Install the key
mkdir -p /root/.ssh
chmod 700 /root/.ssh

# Append to authorized_keys (don't overwrite existing keys)
echo "$SSH_KEY" >> /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys

echo "install-ssh-key: SSH key installed successfully"
