# TLS Certificates Layer Dockerfile
# Provides CA certificates and certificate management tools
#
# Build: docker build -t nanofuse-layer-tls-certificates .
# Export: docker export $(docker create nanofuse-layer-tls-certificates) | tar -C rootfs -xf -

FROM ubuntu:24.04

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install certificate tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    openssl \
    curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create certificate directory (0750: owner and group access; matches layer.yaml default)
# Private keys inside use 600, so group access to directory is safe.
RUN mkdir -p /etc/nanofuse/certs && chmod 0750 /etc/nanofuse/certs

# Create certificate manager script
RUN cat > /usr/local/bin/cert-manager << 'EOF'
#!/bin/bash
# NanoFuse Certificate Manager
# Simple certificate management for the gateway

set -euo pipefail

CERT_DIR="${CERT_DIR:-/etc/nanofuse/certs}"
CA_BUNDLE="${CA_BUNDLE:-/etc/ssl/certs/ca-certificates.crt}"
CONFIG_FILE="/etc/nanofuse/cert-manager.conf"
DEFAULT_RENEWAL_DAYS=30

# Validate positive integer with optional upper bound.
# Usage: validate_positive_int VALUE NAME [MAX_VALUE]
# This function signature matches post-install.sh hooks for consistency across the codebase.
# Both are standalone scripts that never share a namespace, so there's no risk of collision.
validate_positive_int() {
    local value="$1"
    local name="$2"
    local max_value="${3:-}"
    if ! [[ "${value}" =~ ^[0-9]+$ ]] || [[ "${value}" -lt 1 ]]; then
        echo "Error: ${name} must be a positive integer, got: '${value}'" >&2
        return 1
    fi
    if [[ -n "${max_value}" ]] && [[ "${value}" -gt "${max_value}" ]]; then
        echo "Error: ${name} must be between 1 and ${max_value}, got: '${value}'" >&2
        return 1
    fi
    return 0
}

# Load renewal days from config file, environment, or use default
get_renewal_days() {
    local value=""
    if [[ -n "${RENEWAL_DAYS:-}" ]]; then
        value="${RENEWAL_DAYS}"
    elif [[ -f "${CONFIG_FILE}" ]]; then
        value=$(grep -E '^RENEWAL_DAYS=' "${CONFIG_FILE}" 2>/dev/null | tail -1 | cut -d'=' -f2-)
    fi
    # Validate: must be positive integer and <= 3650 days (10 years) to prevent overflow
    local max_renewal_days=3650
    if [[ -n "${value}" ]]; then
        if validate_positive_int "${value}" "RENEWAL_DAYS" "${max_renewal_days}"; then
            echo "${value}"
            return
        fi
        echo "Warning: RENEWAL_DAYS '${value}' invalid or exceeds max ${max_renewal_days}, using default ${DEFAULT_RENEWAL_DAYS}" >&2
    fi
    echo "${DEFAULT_RENEWAL_DAYS}"
}

RENEWAL_DAYS=$(get_renewal_days)

usage() {
    cat << USAGE
Usage: cert-manager <command> [arguments]

Commands:
    generate-self-signed <cn> [days]  Generate self-signed certificate
    verify [cert-path]                Verify certificate validity
    info [cert-path]                  Show certificate information
    renew                             Check and renew certificates if needed
    help                              Show this help message

Examples:
    cert-manager generate-self-signed gateway.local 365
    cert-manager verify /etc/nanofuse/certs/server.crt
    cert-manager info /etc/nanofuse/certs/server.crt
    cert-manager renew
USAGE
}

generate_self_signed() {
    local cn="${1:-gateway.local}"
    local days="${2:-365}"

    # Validate days is a positive integer within allowed range.
    # Max 825 days per CA/Browser Forum Baseline Requirements (effective Sep 2020) for public CAs.
    # Self-signed certs aren't bound by this, but we enforce it as a best practice to encourage
    # regular rotation and prevent certificates that outlive their intended purpose.
    local max_days=825
    # Enforce numeric value and range [1, max_days]
    if ! [[ "${days}" =~ ^[0-9]+$ ]] || [[ "${days}" -lt 1 ]] || [[ "${days}" -gt ${max_days} ]]; then
        echo "Error: days must be a positive integer between 1 and ${max_days}, got: '${days}'" >&2
        return 1
    fi

    openssl req -x509 -nodes -days "${days}" -newkey rsa:2048 \
        -keyout "${CERT_DIR}/server.key" \
        -out "${CERT_DIR}/server.crt" \
        -subj "/CN=${cn}" \
        -addext "subjectAltName=DNS:${cn},DNS:localhost,IP:127.0.0.1"

    chmod 600 "${CERT_DIR}/server.key"
    chmod 644 "${CERT_DIR}/server.crt"

    echo "Generated self-signed certificate for ${cn}"
    echo "  Certificate: ${CERT_DIR}/server.crt"
    echo "  Private key: ${CERT_DIR}/server.key"
}

verify_cert() {
    local cert="${1:-${CERT_DIR}/server.crt}"
    # Check against configured renewal window (default 30 days)
    local check_seconds=$((RENEWAL_DAYS * 24 * 60 * 60))

    if openssl x509 -in "${cert}" -noout -checkend "${check_seconds}"; then
        echo "Certificate is valid for at least ${RENEWAL_DAYS} days"
        return 0
    else
        echo "Certificate will expire within ${RENEWAL_DAYS} days or is invalid"
        return 1
    fi
}

renew_cert() {
    local cert="${CERT_DIR}/server.crt"
    local key="${CERT_DIR}/server.key"
    local lockfile="${CERT_DIR}/.cert-renewal.lock"

    # Use flock to prevent concurrent renewal operations
    # The lock is released automatically when the subshell exits
    (
        # flock -n returns immediately if lock cannot be acquired (non-blocking).
        # Exit code 1 means lock held by another process; other non-zero codes indicate errors.
        # We treat both cases as "skip" since we can't proceed either way.
        flock_exit=0
        flock -n 200 || flock_exit=$?
        if [[ ${flock_exit} -ne 0 ]]; then
            if [[ ${flock_exit} -eq 1 ]]; then
                echo "Another certificate renewal is already in progress, skipping"
            else
                echo "Warning: flock failed with exit code ${flock_exit}, skipping renewal" >&2
            fi
            exit 0
        fi

        if [[ ! -f "${cert}" ]]; then
            echo "No certificate found at ${cert}, generating new one..."
            generate_self_signed "gateway.local" 365
            exit 0
        fi

        # Check if renewal is needed
        local check_seconds=$((RENEWAL_DAYS * 24 * 60 * 60))
        if openssl x509 -in "${cert}" -noout -checkend "${check_seconds}" 2>/dev/null; then
            echo "Certificate is still valid, no renewal needed"
            exit 0
        fi

        # Extract CN from certificate subject using RFC2253 format.
        # openssl outputs "subject=CN=value,..." which we parse with sed.
        # RFC2253 escapes commas as \, and backslashes as \\.
        # The sed command uses POSIX BRE (Basic Regular Expressions) syntax (no -E):
        #   - '\(' and '\)' delimit a capturing group, and '\|' denotes alternation.
        #   - The pattern '\([^,\\]\|\\.\)*' is a capturing group containing zero or more
        #     repetitions of either a non-comma/non-backslash character or a backslash
        #     followed by any character.
        #   - Overall, we strip the "subject=" prefix, find 'CN=' and capture its value,
        #     then unescape '\,' to ',' and '\\' to '\'.
        # Example: "subject=CN=Doe\, John,O=Acme" -> "Doe, John"
        # Note: If certificate has multiple CN values (uncommon but valid per X.509),
        # head -1 selects the first match. This is intentional for simplicity; production
        # environments with complex subject DNs should validate certificates manually.
        local cn=""
        cn=$(openssl x509 -in "${cert}" -noout -subject -nameopt RFC2253 2>/dev/null \
            | sed -n 's/^subject=//; s/^.*CN=\(\([^,\\]\|\\.\)*\).*$/\1/p' \
            | head -1 \
            | sed 's/\\,/,/g; s/\\\\/\\/g')
        cn="${cn:-gateway.local}"

        echo "Certificate needs renewal, regenerating..."
        # Note: This directly overwrites existing certificates. For production ACME certificates,
        # atomic replacement (generate to temp, then rename) should be implemented. For self-signed
        # certs managed by this tool, brief unavailability during renewal is acceptable since
        # the renewal timer runs during low-traffic periods (daily with random delay).
        generate_self_signed "${cn}" 365
    ) 200>"${lockfile}"
    # Cleanup the lockfile after releasing the lock to avoid leaving stale files behind.
    rm -f "${lockfile}"
    # Note: File descriptor 200 is intentionally reserved for certificate renewal locking.
    # Callers of cert-manager must ensure fd 200 is not already in use; otherwise, locking
    # behavior may be undefined. The flock command uses fd 200 to hold the lock on ${lockfile}
    # for the duration of the subshell. When the subshell exits, fd 200 is closed and the lock
    # is released. We chose fd 200 to avoid conflicts with standard fds (0-2) and commonly
    # used fds (3-9). Bash 4.1+ supports {varname}>file for auto-assignment, but we use a
    # fixed fd for compatibility with older bash versions that may be in base images.
}

show_info() {
    local cert="${1:-${CERT_DIR}/server.crt}"

    echo "Certificate information for ${cert}:"
    openssl x509 -in "${cert}" -noout -subject -issuer -dates -fingerprint
}

case "${1:-help}" in
    generate-self-signed)
        shift
        cn="${1:-gateway.local}"
        days="${2:-365}"
        generate_self_signed "${cn}" "${days}"
        ;;
    verify)
        shift
        cert="${1:-${CERT_DIR}/server.crt}"
        verify_cert "${cert}"
        ;;
    info)
        shift
        cert="${1:-${CERT_DIR}/server.crt}"
        show_info "${cert}"
        ;;
    renew)
        renew_cert
        ;;
    help|*)
        usage
        ;;
esac
EOF
chmod +x /usr/local/bin/cert-manager
