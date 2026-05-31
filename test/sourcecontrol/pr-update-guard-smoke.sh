#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

cat >"$tmpdir/gh" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail

if [[ "$1 $2" == "repo view" ]]; then
  echo "daax-dev/nanofuse"
  exit 0
fi

if [[ "$1 $2" == "pr list" ]]; then
  head_branch=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --head)
        head_branch="$2"
        shift 2
        ;;
      *)
        shift
        ;;
    esac
  done

  if [[ "$head_branch" == "branch-with-pr" ]]; then
    echo "#123 OPEN https://github.com/daax-dev/nanofuse/pull/123 guarded PR"
  fi
  if [[ "$head_branch" == "pr-list-fails" ]]; then
    echo "simulated gh failure" >&2
    exit 22
  fi
  exit 0
fi

echo "unexpected gh call: $*" >&2
exit 64
STUB
chmod +x "$tmpdir/gh"

guard="$repo_root/scripts/guard-pr-updates.sh"
zero_oid="0000000000000000000000000000000000000000"
local_oid="1111111111111111111111111111111111111111"

PATH="$tmpdir:$PATH" "$guard" origin git@github.com:daax-dev/nanofuse.git <<EOF
refs/heads/new-branch ${local_oid} refs/heads/new-branch ${zero_oid}
EOF

if PATH="$tmpdir:$PATH" "$guard" origin git@github.com:daax-dev/nanofuse.git <<EOF
refs/heads/branch-with-pr ${local_oid} refs/heads/branch-with-pr ${zero_oid}
EOF
then
  echo "expected guard to block branch-with-pr" >&2
  exit 1
fi

if PATH="$tmpdir:$PATH" "$guard" origin git@github.com:daax-dev/nanofuse.git <<EOF
refs/heads/pr-list-fails ${local_oid} refs/heads/pr-list-fails ${zero_oid}
EOF
then
  echo "expected guard to block when gh cannot verify PR state" >&2
  exit 1
fi

PATH="$tmpdir:$PATH" "$guard" origin git@github.com:daax-dev/nanofuse.git <<EOF
refs/tags/v-test ${local_oid} refs/tags/v-test ${zero_oid}
EOF

PATH="$tmpdir:$PATH" "$guard" origin git@github.com:daax-dev/nanofuse.git <<EOF
refs/heads/branch-with-pr ${zero_oid} refs/heads/branch-with-pr ${local_oid}
EOF

mkdir "$tmpdir/noop-work"
git -C "$tmpdir/noop-work" init -q
git -C "$tmpdir/noop-work" checkout -q -b branch-with-pr
if (cd "$tmpdir/noop-work" && PATH="$tmpdir:$PATH" "$guard" origin git@github.com:daax-dev/nanofuse.git </dev/null); then
  echo "expected guard to block no-op push from branch-with-pr" >&2
  exit 1
fi

git -C "$tmpdir/noop-work" checkout -q -b branch-without-pr
(cd "$tmpdir/noop-work" && PATH="$tmpdir:$PATH" "$guard" origin git@github.com:daax-dev/nanofuse.git </dev/null)
