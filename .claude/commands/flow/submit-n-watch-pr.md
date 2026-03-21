---
description: Submit PR and watch until all CI checks pass and Copilot review has zero comments.
---

# /flow:submit-n-watch-pr - Automated PR Submission with CI and Copilot Review Loop

Submit a pull request and iterate until all CI checks pass and GitHub Copilot review returns zero actionable comments.

## User Input

```text
$ARGUMENTS
```

**Expected Input**: Optional PR number or branch name. If not provided, uses current branch.

## Branch Naming Convention (MANDATORY)

Branch names MUST follow this format:
```
{host}/{task#}/{relevant-slug}
```

Example: `galway/task-24/layer-types-interfaces`

**If current branch does not match this format:**
1. Determine correct branch name from context (hostname, task ID, feature description)
2. Rename the branch: `git branch -m old-name new-name`
3. Update remote: `git push origin :old-name new-name`
4. Set upstream: `git push -u origin new-name`

## Pre-Submission Checklist

Before ANY PR submission, verify:

```bash
# 1. Get current hostname for branch naming
hostname

# 2. Ensure rebased on main with zero conflicts
git fetch origin main
git rebase origin/main
# If conflicts: STOP, resolve, then continue

# 3. Run full test suite
go test ./... -count=1

# 4. Run linter
golangci-lint run ./...

# 5. Verify no secrets or sensitive data
git diff origin/main --stat
```

## Workflow Phases

### Phase 1: Branch Validation and PR Creation

```bash
# Check branch name format
BRANCH=$(git branch --show-current)
HOST=$(hostname)

# Validate format: {host}/{task#}/{slug}
if ! echo "$BRANCH" | grep -qE "^[a-z]+/task-[0-9]+/"; then
  echo "ERROR: Branch name must be {host}/{task#}/{slug}"
  echo "Current: $BRANCH"
  echo "Expected format: $HOST/task-XX/description-slug"
  # Prompt user for correct task number and create proper branch name
fi

# Create PR if not exists
gh pr create --title "..." --body "..."
```

### Phase 2: Wait for CI Checks

```bash
# Poll CI status every 30 seconds until all checks complete
while true; do
  STATUS=$(gh pr checks $PR_NUMBER --json state --jq '.[] | .state' | sort -u)

  if echo "$STATUS" | grep -q "FAILURE"; then
    echo "CI FAILED - fetching details..."
    gh pr checks $PR_NUMBER
    break
  fi

  if echo "$STATUS" | grep -q "PENDING"; then
    echo "Waiting for CI... (sleeping 30s)"
    sleep 30
    continue
  fi

  if [ "$STATUS" = "SUCCESS" ]; then
    echo "All CI checks passed!"
    break
  fi
done
```

### Phase 3: Fetch and Parse Failures

For each failure or Copilot comment, create a JSONL checklist:

```jsonl
{"id": 1, "type": "ci_failure", "check": "Lint and Format Check", "file": null, "line": null, "issue": "gofmt formatting issues", "status": "pending"}
{"id": 2, "type": "copilot", "check": null, "file": "internal/layerbuild/cache.go", "line": 91, "issue": "Auto-touch goroutine lacks error handling", "status": "pending"}
{"id": 3, "type": "copilot", "check": null, "file": "internal/layerbuild/fetcher.go", "line": 436, "issue": "Directory traversal validation should happen before path construction", "status": "pending"}
```

### Phase 4: Fix All Issues (Ultrathink)

For each issue in the checklist:

1. **Read the relevant code context**
2. **Ultrathink the root cause and best fix**
3. **Apply the fix**
4. **Mark item as fixed in checklist**
5. **Add to "lessons learned" for future prevention**

```markdown
## Lessons Learned (append to memory)

- Always close files explicitly in loops, not with defer
- Validate tar header names BEFORE constructing paths
- Never use fmt.Sprintf for SQL - always use parameterized queries
- Log errors in goroutines, don't silently discard
```

### Phase 5: Commit, Push, Close Old PR, Create New PR

```bash
# Stage all fixes
git add -A

# Commit with descriptive message
git commit -m "fix: address CI failures and Copilot review feedback

- Fixed: [list each issue]

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

# Force push (we're iterating on same branch)
git push --force-with-lease

# Close old PR
gh pr close $OLD_PR_NUMBER --comment "Superseded by new PR with fixes"

# Create new PR
gh pr create --title "..." --body "..."
```

### Phase 6: Repeat Until Zero Issues

Loop back to Phase 2 until:
- All CI checks pass (SUCCESS)
- Copilot review has zero actionable comments
- Or comments are documented as "not resolvable" with justification

## Commands to Fetch PR Data

```bash
# Get PR details
gh pr view $PR_NUMBER --json title,body,state,statusCheckRollup,headRefName

# Get CI check status
gh pr checks $PR_NUMBER

# Get Copilot review comments
gh api repos/{owner}/{repo}/pulls/{pr}/comments --jq '.[] | {path: .path, line: .line, body: .body, user: .user.login}'

# Get CI run logs for failures
gh run view $RUN_ID --log-failed
```

## MUST DO Checklist Template

Create this checklist at start of each iteration:

```markdown
## PR #{number} - Iteration {n} MUST DO Checklist

### CI Failures
- [ ] {check_name}: {description}

### Copilot Comments
- [ ] {file}:{line} - {summary of issue}

### Lessons Learned This Iteration
- {lesson 1}
- {lesson 2}

### Status
- Iteration: {n}
- CI: {PASS/FAIL}
- Copilot Comments: {count}
- Target: 0 comments, all CI green
```

## Exit Criteria

The workflow completes when:
1. All CI checks show SUCCESS
2. Copilot review returns zero new comments
3. PR is ready for human review/merge

## Error Recovery

If stuck in loop:
1. After 3 iterations with same issue, mark as "needs human review"
2. Document why the issue cannot be auto-resolved
3. Create a backlog task for follow-up
4. Proceed with PR noting the limitation

## Integration with Backlog

```bash
# If fixing creates new work, track it
backlog task create "Follow-up: {description}" \
  --label "tech-debt" \
  --priority low
```

## Post-Completion

After successful PR merge:
1. Update any related backlog tasks to Done
2. Delete the feature branch
3. Log lessons learned to `.claude/lessons-learned.md` (create if not exists)
