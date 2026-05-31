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

check_branch() {
  local branch="$1"
  [[ -z "$branch" ]] && return 0

  local errfile
  errfile="$(mktemp)"
  if ! prs="$(gh pr list \
    --repo "$repo" \
    --head "$branch" \
    --state open \
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
    return 0
  fi
  rm -f "$errfile"

  if [[ -n "$prs" ]]; then
    {
      echo "nanofuse PR guard: blocked push to '${branch}' on ${remote_name}."
      echo "An open GitHub PR exists for this branch; open PR branches are immutable."
      echo "$prs"
      echo "Close the PR before reusing this branch for fixes."
    } >&2
    blocked=1
  fi
}

processed_refs=0

while read -r local_ref local_oid remote_ref remote_oid; do
  [[ -z "${local_ref:-}" ]] && continue
  processed_refs=1

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

  check_branch "$branch"
done

if [[ "$processed_refs" -eq 0 ]]; then
  current_branch="$(git branch --show-current 2>/dev/null || true)"
  check_branch "$current_branch"
fi

exit "$blocked"
