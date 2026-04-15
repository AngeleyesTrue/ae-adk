---
name: ae-workflow-auto
description: >
  Context-Isolated Auto Pipeline. Executes Run -> Sync-Review Loop -> Final Merge
  using independent Agent Teams per phase. Each team has exactly one teammate that
  inherits project skills (/ae run, /ae sync, /ae review). Context is released
  between phases via TeamDelete to prevent hallucination and context exhaustion.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-15"
  tags: "auto, pipeline, autonomous, context-isolated"

# AE Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 8000

# AE Extension: Triggers
triggers:
  keywords: ["auto", "pipeline", "autonomous"]
  agents: []
  phases: ["run"]
---

# Auto Pipeline Workflow

## Purpose

Execute a Context-Isolated pipeline: Run -> Sync-Review Loop -> Final Merge.
Each phase runs in an independent Agent Team (1 teammate per team). After phase
completion, TeamDelete releases the context entirely (Context Zero Principle).

## Architectural Justification

Agent Teams are used instead of simple Agent() calls because Agent Teams teammates
inherit project skills and can execute skill commands like `/ae run`, `/ae sync`,
`/ae review`. Simple Agent() sub-agents operate in isolated contexts without access
to project skill definitions. This is the key architectural reason for using Agent Teams.

## Input

- $ARGUMENTS: SPEC-ID to execute (e.g., SPEC-AUTH-001)
- Flags: --iterations N, --skip-run, --no-copilot

## Configuration

Load from `.ae/config/sections/auto.yaml`:

| Key | Default | Description |
|-----|---------|-------------|
| context_isolated.enabled | true | Enable context-isolated pipeline |
| context_isolated.sync_review_iterations | 3 | Number of Sync-Review iterations |
| context_isolated.copilot.enabled | true | Enable Copilot review integration |
| context_isolated.copilot.check_iteration | 1 | Iteration to check Copilot |
| context_isolated.copilot.wait_minutes | 10 | Minutes to wait for Copilot review |
| context_isolated.copilot.bot_login | copilot-pull-request-reviewer[bot] | Copilot bot identifier |
| context_isolated.teammate.count | 1 | Teammates per team (always 1) |
| context_isolated.teammate.mode | auto | Teammate permission mode |
| context_isolated.final_merge.strategy | squash | Merge strategy |
| context_isolated.final_merge.delete_branch | true | Delete branch after merge |
| context_isolated.final_merge.require_ci_pass | true | Require CI pass before merge |

## CLI Flag Override

| Flag | Default | Description | Config Override |
|------|---------|-------------|-----------------|
| --iterations N | 3 | Sync-Review iteration count | sync_review_iterations |
| --skip-run | false | Skip Run Phase | N/A |
| --no-copilot | false | Disable Copilot review waiting | copilot.enabled |

## Pre-Execution

1. Parse CLI flags from $ARGUMENTS
2. Load auto.yaml configuration
3. Apply flag overrides to config values
4. Verify SPEC document exists at `.ae/specs/{SPEC-ID}/spec.md`
5. Verify git state is clean (no uncommitted changes)

## Pipeline Sequence

```
/ae auto SPEC-XXX [--iterations N] [--skip-run] [--no-copilot]
  |
  +- Phase 1: Run (Team #1) [skipped if --skip-run]
  |   +- TeamCreate("run-team")
  |   +- Spawn 1 teammate (general-purpose, mode: auto)
  |   +- Teammate executes: /ae run {spec_id}
  |   +- Branch: feature/{spec_id}
  |   +- Await completion
  |   +- TeamDelete("run-team")
  |
  +- Phase 2: Sync-Review Loop (N iterations)
  |   |
  |   +- Iteration 1:
  |   |   +- Sync (Team: sync-team-1)
  |   |   |   +- TeamCreate -> teammate: /ae sync {spec_id} -> TeamDelete
  |   |   +- Orchestrator: gh pr list --head feature/{spec_id} -> PR number
  |   |   +- Copilot Wait (if enabled AND iteration == check_iteration)
  |   |   |   +- Bash: sleep {wait_minutes * 60}
  |   |   |   +- Bash: gh api repos/{owner}/{repo}/pulls/{pr}/reviews
  |   |   |   +- Filter for copilot-pull-request-reviewer[bot]
  |   |   +- Review (Team: review-team-1)
  |   |       +- TeamCreate -> teammate: /ae review {spec_id} + PR#{number} + Copilot comments -> TeamDelete
  |   |
  |   +- Iteration 2..N:
  |       +- Sync (Team: sync-team-{i}) -> TeamDelete
  |       +- Orchestrator: PR number query
  |       +- Review (Team: review-team-{i}) + PR#{number} -> TeamDelete
  |
  +- Phase 3: Final Merge
      +- gh pr list --head feature/{spec_id} --state open
      +- gh pr checks {pr_number}
      +- If all pass: gh pr merge {pr_number} --squash --delete-branch
      +- If fail: Report to user with options
```

## Phase 1: Run Phase

### Condition
Skip if `--skip-run` flag is specified.

### Execution

```
Report: Phase Start
  "Starting Run Phase for {spec_id}"
  "Team: run-team, Task: /ae run {spec_id}"

TeamCreate("run-team")

Spawn teammate:
  - subagent_type: "general-purpose"
  - mode: "auto"
  - prompt: |
      Execute /ae run {spec_id}.
      Create branch feature/{spec_id} if it does not exist.
      Implement all SPEC requirements.
      Self-review iteratively until satisfied.
      Commit all changes.
      Push the feature branch to remote: git push -u origin feature/{spec_id}
      IMPORTANT: The branch MUST be pushed to remote before completion.

Await teammate completion.

TeamDelete("run-team")

Report: Phase Complete
  "Run Phase complete. Branch feature/{spec_id} pushed."
```

### Completion Criteria
- All SPEC requirements implemented
- Tests passing
- Changes committed
- Feature branch pushed to remote

## Phase 2: Sync-Review Loop

### Iteration Logic

```
iterations = --iterations flag OR config.sync_review_iterations (default 3)
copilot_enabled = NOT --no-copilot AND config.copilot.enabled
check_iteration = config.copilot.check_iteration (default 1)

FOR i = 1 TO iterations:

  ## Sync Phase (iteration i)

  Report: Sync Phase Start
    "Sync-Review Loop: iteration {i}/{iterations}"
    "Team: sync-team-{i}, Task: /ae sync {spec_id}"

  TeamCreate("sync-team-{i}")

  Spawn teammate:
    - subagent_type: "general-purpose"
    - mode: "auto"
    - prompt: |
        Execute /ae sync {spec_id}.
        Synchronize documentation with code changes.
        Create or update the pull request from feature/{spec_id} to main.
        Commit and push any documentation changes.

  Await teammate completion.
  TeamDelete("sync-team-{i}")

  ## PR Number Resolution

  Query the open PR for the feature branch:
    Bash: gh pr list --head feature/{spec_id} --state open --json number,url --jq '.[0]'

  Extract pr_number and pr_url from the result.
  If no PR found: Report warning, continue (PR may have been merged or not created).

  ## Copilot Wait (conditional)

  IF copilot_enabled AND i == check_iteration AND pr_number exists:

    Report: Copilot Wait
      "Waiting {wait_minutes} minutes for Copilot review on PR #{pr_number}..."

    Bash: sleep {wait_minutes * 60}

    Query Copilot reviews:
      Bash: gh api repos/{owner}/{repo}/pulls/{pr_number}/reviews --jq '[.[] | select(.user.login == "{bot_login}")]'

    IF copilot_comments exist:
      Store copilot_feedback for Review Phase prompt
    ELSE:
      copilot_feedback = null
      Report: "No Copilot review comments found. Proceeding with standard review."

  ## Review Phase (iteration i)

  Report: Review Phase Start
    "Team: review-team-{i}, Task: /ae review {spec_id}"

  TeamCreate("review-team-{i}")

  Spawn teammate:
    - subagent_type: "general-purpose"
    - mode: "auto"
    - prompt: |
        Execute /ae review {spec_id}.
        Review PR #{pr_number} ({pr_url}) with a critical perspective.
        Focus on: code quality, security, test coverage, architectural consistency.
        If issues are found, fix them, commit, and push.
        Iterate your review-fix loop until no critical issues remain.
        {IF copilot_feedback: "Copilot review feedback to address: {copilot_feedback}"}
        {IF i == iterations: "This is the FINAL review iteration. Be thorough."}

  Await teammate completion.
  TeamDelete("review-team-{i}")

  Report: Iteration Complete
    "Sync-Review iteration {i}/{iterations} complete."

END FOR
```

## Phase 3: Final Merge

### Execution

```
Report: Final Merge Start
  "Attempting final merge for feature/{spec_id}..."

## Query PR
Bash: gh pr list --head feature/{spec_id} --state open --json number --jq '.[0].number'

IF no open PR:
  Report: "No open PR found for feature/{spec_id}. Pipeline complete without merge."
  Return.

## Check CI
Bash: gh pr checks {pr_number}

IF all checks pass:
  Bash: gh pr merge {pr_number} --squash --delete-branch
  Report: Merge Complete
    "PR #{pr_number} merged successfully via squash. Branch deleted."

IF checks fail:
  Report: Merge Blocked
    "PR #{pr_number} has failing CI checks."
  AskUserQuestion:
    Question: "CI checks are failing on PR #{pr_number}. How would you like to proceed?"
    Options:
      - "Wait and retry (Recommended): Wait for CI to complete and retry merge"
      - "Force merge: Merge despite failing checks"
      - "Manual intervention: Stop pipeline for manual review"
      - "Abort: Cancel the merge"

IF merge conflict detected:
  Report: Merge Conflict
    "PR #{pr_number} has merge conflicts with the target branch."
  AskUserQuestion:
    Question: "Merge conflict detected on PR #{pr_number}. Manual rebase is recommended."
    Options:
      - "Rebase attempt (Recommended): Try automatic rebase"
      - "Manual intervention: Stop for manual conflict resolution"
      - "Abort: Cancel the merge"
```

## Error Recovery

### TeamCreate Failure
- Retry up to 2 times with 5-second delay
- After 3 failures: Report to user via AskUserQuestion
  - Options: Retry, Abort pipeline

### Teammate Execution Failure
- TeamDelete to clean up the current team
- Report error details to user via AskUserQuestion
  - Options: Retry current phase, Skip to next phase, Abort pipeline

### Copilot API Failure
- Log warning: "Copilot API call failed. Proceeding without Copilot feedback."
- Continue with standard Review Phase prompt (no Copilot comments)
- Non-blocking: pipeline continues

### PR Query Failure
- If gh pr list returns empty or error: Log warning and continue
- Review Phase proceeds without PR number in prompt
- Non-blocking: pipeline continues

### Partial Completion Recovery
- Use `--skip-run` flag to resume from Sync-Review Loop
- Manually specify `--iterations` to control remaining iterations
- No automated state persistence required

## Progress Reporting

Use AE output style Progress template for all phase reports:

```markdown
👁️ AE ★ Auto Pipeline ─────────────────────
📋 SPEC: {spec_id}
📊 Phase: {phase_name} ({iteration}/{total} if applicable)
⏳ {status_message}
────────────────────────────────────────────
```

### Report Points
- Pipeline start: SPEC ID, configuration summary, total iterations
- Phase start: Team name, task description, current iteration
- Phase completion: Result summary, next phase guidance
- Copilot wait: Duration, PR number
- Final merge: CI status, merge result
- Error: Error details, recovery options

## Team Summary (default 3 iterations)

| # | Phase | Team Name | Teammates | Task |
|---|-------|-----------|-----------|------|
| 1 | Run | run-team | 1 | Implementation + self-review |
| 2 | Sync 1 | sync-team-1 | 1 | Doc sync + PR creation |
| 3 | Review 1 | review-team-1 | 1 | Review + Copilot feedback |
| 4 | Sync 2 | sync-team-2 | 1 | Doc sync + PR update |
| 5 | Review 2 | review-team-2 | 1 | Review |
| 6 | Sync 3 | sync-team-3 | 1 | Doc sync + PR update |
| 7 | Review 3 | review-team-3 | 1 | Final review |

Total: 7 teams created/deleted, 1 teammate each.

## Context Zero Trade-off

Each phase starts with a clean context (preventing hallucination and context
exhaustion), but must re-explore the codebase from scratch. The orchestrator
mitigates this by including structured handoff data in each phase's prompt:

- Branch name: `feature/{spec_id}`
- PR number and URL (queried between phases)
- Key changed files (from previous phase reports)
- Copilot feedback (if available)
- Iteration context (current/total)

---

Version: 1.0.0
Source: SPEC-PIPELINE-001
