#!/usr/bin/env bash
# Run the nanofuse Vagrant/hypervisor validation loop from the host.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$SCRIPT_DIR"

log() {
    printf '[closed-loop] %s\n' "$*"
}

run() {
    log "$*"
    "$@"
}

log "repo: $REPO_ROOT"
run vagrant --version
run vagrant status

if ! vagrant status --machine-readable | grep -q ',state,running$'; then
    log "VM is not running; starting/provisioning it now"
    run vagrant up
else
    log "VM is already running; syncing and provisioning"
    run vagrant rsync
    run vagrant provision
fi

log "guest capability preflight"
run vagrant ssh -c 'set -euo pipefail; uname -a; test -r /dev/kvm && test -w /dev/kvm && echo "/dev/kvm OK"'

log "repo gates inside guest"
run vagrant ssh -c 'set -euo pipefail; cd /nanofuse; sudo mage ci'

log "daemon and Firecracker verification"
run vagrant ssh -c 'set -euo pipefail; sudo systemctl restart nanofused; sleep 2; sudo systemctl status nanofused --no-pager; nanofuse health; sudo /vagrant-scripts/verify.sh'

log "closed-loop validation complete"
