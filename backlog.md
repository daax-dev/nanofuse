
Backlog - CI failures (PR #15)
This file was created by the GitHub Copilot Chat Assistant to collect one task per failing CI job from PR #15 (https://github.com/daax-dev/nanofuse/pull/15). Each task includes a clear description, suggested fix, and acceptance criteria required for the task to be considered done.

Context: workflow run for PR #15 failed. Known failing job log extracted from job ID 56498005151. The commit ref used for file links is f9c66d54406e3be4b7ec06ab71f5cd83c714ad4c.

Task 1 — golangci-lint config error (job 56498005151)
Title: CI: golangci-lint config error — add version: 1 to .golangci.yml (job 56498005151)

Summary:
The golangci-lint step fails to load the configuration with: "unsupported version of the configuration: """. This is caused by the repository-level .golangci.yml missing the top-level "version" field expected by the golangci-lint binary used in CI.

Failure details:

Job ID: 56498005151
Repo: daax-dev/nanofuse
Ref (commit): f9c66d54406e3be4b7ec06ab71f5cd83c714ad4c
Workflow / PR: https://github.com/daax-dev/nanofuse/pull/15
Log excerpt:
Error: can't load config: unsupported version of the configuration: "" See https://golangci-lint.run/docs/product/migration-guide for migration instructions

File to change (link uses the workflow ref):

.golangci.yml at ref f9c66d54406e3be4b7ec06ab71f5cd83c714ad4c:
https://github.com/daax-dev/nanofuse/blob/f9c66d54406e3be4b7ec06ab71f5cd83c714ad4c/.golangci.yml
Suggested fix (exact change):
Prepend the following line to the top of .golangci.yml:

YAML
version: 1
Acceptance criteria (when this task is Done):

 .golangci.yml contains version: 1 at the top of the file.
 CI is re-run and the golangci-lint job no longer fails with the "unsupported version" error (may still report linter findings).
 The change is merged into main (via branch/PR).
Suggested labels: ci, bug

Task 2 — failing job: capture logs and identify root cause
Title: CI: capture logs and identify root cause for failing job <job-id> (PR #15)

Summary:
Another job in the PR's workflow failed. Logs for that job are not present in the extracted context. Collect the job id and full logs, determine the root cause, and propose a targeted fix.

Required inputs / reproduction steps:

Job ID (from the workflow run) or link to the job logs in the GitHub Actions UI
Full job log (paste or attach) or permission to fetch logs from the workflow run
Indicate whether this failure is reproducible locally (provide reproduction steps if available)
Suggested remediation steps:

Inspect the failing step and the exact error message
If configuration-related, propose a minimal config change
If script/command-related, propose a small code/CI change with patch
Acceptance criteria (when this task is Done):

 The job ID and full logs are attached to this task.
 A clear root-cause statement is written (1–2 sentences) with evidence from the log.
 A proposed fix is provided as either a one-line patch or specific CI change with expected outcome.
 The fix is committed to a branch and re-run shows the job no longer failing (or the CI team confirms the root cause and defers the fix with justification).
Suggested labels: ci, triage

Task 3 — failing job: apply fix or collect missing artifacts
Title: CI: failing job <job-id> — apply patch or collect additional artifacts (PR #15)

Summary:
A third job failed. Implement the identified fix (if simple) or collect additional artifacts required to diagnose the failure.

Required inputs / reproduction steps:

Job ID and logs (same as Task 2)
Any relevant files referenced by the job (e.g., CI script, workflow yml path: .github/workflows/ci.yaml)
Acceptance criteria (when this task is Done):

 Job ID and logs/files are documented.
 If the fix is straightforward (configuration or one-line change), a branch exists with the patch and CI rerun passes.
 Otherwise, missing info is listed and the next action/owner is assigned.
Suggested labels: ci, triage

Notes for maintainers:

If you prefer GitHub Issues rather than backlog.md entries, use the Task headings above as issue titles and paste the corresponding sections into each issue body.
To create a PR from this branch locally, run the commands below (after pushing):
