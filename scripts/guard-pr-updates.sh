#!/usr/bin/env bash
set -euo pipefail

remote_name="${1:-origin}"
remote_url="${2:-}"

if ! command -v gh >/dev/null 2>&1; then
  {
    echo "nanofuse PR guard: blocked push because GitHub CLI is not available."
    echo "Install and authenticate gh, then retry. PR branches are immutable after creation."
  } >&2
  exit 1
fi

repo="$(gh repo view --json nameWithOwner --jq .nameWithOwner 2>/dev/null || true)"
if [[ -z "$repo" ]]; then
  {
    echo "nanofuse PR guard: blocked push because repository identity could not be verified with gh."
    echo "Remote: ${remote_name} ${remote_url}"
  } >&2
  exit 1
fi

blocked=0

while read -r local_ref local_oid remote_ref remote_oid; do
  [[ -z "${local_ref:-}" ]] && continue

  # Branch deletes do not update a PR diff.
  if [[ "$local_oid" =~ ^0+$ ]]; then
    continue
  fi

  branch=""
  if [[ "$local_ref" == refs/heads/* ]]; then
    branch="${local_ref#refs/heads/}"
  elif [[ "$remote_ref" == refs/heads/* ]]; then
    branch="${remote_ref#refs/heads/}"
  else
    continue
  fi

  errfile="$(mktemp)"
  if ! prs="$(gh pr list \
    --repo "$repo" \
    --head "$branch" \
    --state all \
    --limit 100 \
    --json number,state,title,url \
    --jq '.[] | "#\(.number) \(.state) \(.url) \(.title)"' 2>"$errfile")"; then
    {
      echo "nanofuse PR guard: blocked push to '${branch}' because GitHub PR state could not be verified."
      sed 's/^/gh: /' "$errfile"
      echo "Create a fresh branch and retry after gh can verify PR state."
    } >&2
    rm -f "$errfile"
    blocked=1
    continue
  fi
  rm -f "$errfile"

  if [[ -n "$prs" ]]; then
    {
      echo "nanofuse PR guard: blocked push to '${branch}' on ${remote_name}."
      echo "A GitHub PR already exists for this branch; PR branches are immutable after creation."
      echo "$prs"
      echo "Create a fresh branch and a fresh PR instead."
    } >&2
    blocked=1
  fi
done

exit "$blocked"
