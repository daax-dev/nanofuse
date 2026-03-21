#!/bin/bash
# Install SSH public key from kernel command line
# The key is passed as base64-encoded sshkey= parameter

set -euo pipefail

# Extract sshkey parameter from kernel cmdline
SSHKEY_B64=$(sed -n 's/.*sshkey=\([^ ]*\).*/\1/p' /proc/cmdline)

if [ -z "$SSHKEY_B64" ]; then
    echo "install-ssh-key: No SSH key provided via kernel cmdline"
    exit 0
fi

# Decode the key (use conditional to avoid set -e termination on decode failure)
if ! SSH_KEY=$(printf '%s' "$SSHKEY_B64" | base64 -d 2>/dev/null) || [ -z "$SSH_KEY" ]; then
    echo "install-ssh-key: Failed to decode SSH key"
    exit 1
fi

# Reject multi-line keys (could inject extra authorized_keys entries)
if [[ "$SSH_KEY" == *$'\n'* ]]; then
    echo "install-ssh-key: Decoded key contains multiple lines; refusing"
    exit 1
fi

# Validate it looks like a public key (type + base64 blob, optional comment)
if ! echo "$SSH_KEY" | grep -qE '^(ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp(256|384|521)|sk-ssh-ed25519@openssh\.com|sk-ecdsa-sha2-nistp256@openssh\.com) [A-Za-z0-9+/=]+( [^ ]+)?$'; then
    echo "install-ssh-key: Invalid SSH key format"
    exit 1
fi

# install_key_for_user <user> <ssh_dir>
# Installs SSH key with symlink attack protection and atomic file updates
install_key_for_user() {
    local user="$1"
    local ssh_dir="$2"

    # Validate SSH directory (reject symlinks and non-directories)
    if [ -L "$ssh_dir" ]; then
        echo "install-ssh-key: $ssh_dir is a symlink; refusing to proceed" >&2
        exit 1
    fi

    if [ -e "$ssh_dir" ] && [ ! -d "$ssh_dir" ]; then
        echo "install-ssh-key: $ssh_dir exists and is not a directory; refusing to proceed" >&2
        exit 1
    fi

    if [ ! -d "$ssh_dir" ]; then
        mkdir -m 700 "$ssh_dir"
    fi
    chmod 700 "$ssh_dir"
    chown "$user:$user" "$ssh_dir"

    local auth_keys="$ssh_dir/authorized_keys"

    # Reject symlinks and non-regular files for authorized_keys
    if [ -L "$auth_keys" ]; then
        rm -f "$auth_keys" 2>/dev/null || true
    fi

    if [ -e "$auth_keys" ] && [ ! -f "$auth_keys" ]; then
        echo "install-ssh-key: $auth_keys is not a regular file; refusing to proceed" >&2
        exit 1
    fi

    # Check if key is already present
    if [ -f "$auth_keys" ] && grep -qF "$SSH_KEY" "$auth_keys"; then
        return
    fi

    # Atomic update via temp file + rename
    umask 077
    local tmp_keys
    tmp_keys=$(mktemp -p "$ssh_dir" authorized_keys.tmp.XXXXXX)

    if [ -f "$auth_keys" ]; then
        cat "$auth_keys" > "$tmp_keys"
    fi

    echo "$SSH_KEY" >> "$tmp_keys"
    chmod 600 "$tmp_keys"
    chown "$user:$user" "$tmp_keys"
    mv -f "$tmp_keys" "$auth_keys"
}

install_key_for_user "root" "/root/.ssh"
install_key_for_user "clawdbot" "/home/clawdbot/.ssh"

echo "install-ssh-key: SSH key installed for root and clawdbot"
