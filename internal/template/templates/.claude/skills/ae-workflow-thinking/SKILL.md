---
name: ae-workflow-thinking
description: >
  Sequential Thinking MCP and UltraThink mode for deep analysis, complex
  problem decomposition, and structured reasoning workflows.
  Use when performing multi-step analysis, architecture decisions, technology selection
  trade-offs, breaking change assessment, or when --deepthink flag is specified.
  Do NOT use for simple decisions or straightforward implementation tasks.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__sequential-thinking__sequentialthinking
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  modularized: "false"

# AE Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level_1_tokens: 100
  level_2_tokens: 3000

# AE Extension: Triggers
triggers:
  keywords: ["sequential thinking", "deepthink", "deep analysis", "complex problem", "architecture decision", "technology selection", "trade-off", "breaking change"]
  phases:
    - plan
  agents:
    - manager-strategy
    - manager-spec
---

# Sequential Thinking & UltraThink

Structured reasoning system for complex problem analysis and decision-making.

## Activation Triggers

Use Sequential Thinking MCP when:

- Breaking down complex problems into steps
- Planning and design with room for revision
- Architecture decisions affect 3+ files
- Technology selection between multiple options
- Performance vs maintainability trade-offs
- Breaking changes under consideration
- Multiple approaches exist to solve the same problem
- Repetitive errors occur

## Tool Parameters

**Required Parameters:**
- `thought` (string): Current thinking step content
- `nextThoughtNeeded` (boolean): Whether another step is needed
- `thoughtNumber` (integer): Current thought number (starts from 1)
- `totalThoughts` (integer): Estimated total thoughts needed

**Optional Parameters:**
- `isRevision` (boolean): Whether this revises previous thinking
- `revisesThought` (integer): Which thought is being reconsidered
- `branchFromThought` (integer): Branching point for alternatives
- `branchId` (string): Branch identifier
- `needsMoreThoughts` (boolean): If more thoughts needed beyond estimate

## Usage Pattern

**Step 1 - Initial Analysis:**
```
thought: "Analyzing the problem: [describe problem]"
nextThoughtNeeded: true
thoughtNumber: 1
totalThoughts: 5
```

**Step 2 - Decomposition:**
```
thought: "Breaking down: [sub-problems]"
nextThoughtNeeded: true
thoughtNumber: 2
totalThoughts: 5
```

**Step 3 - Revision (if needed):**
```
thought: "Revising thought 2: [correction]"
isRevision: true
revisesThought: 2
thoughtNumber: 3
totalThoughts: 5
nextThoughtNeeded: true
```

**Final Step - Conclusion:**
```
thought: "Conclusion: [final answer]"
thoughtNumber: 5
totalThoughts: 5
nextThoughtNeeded: false
```

## Adaptive Thinking (Claude Opus 4.7+)

Claude Opus 4.7 introduces Adaptive Thinking: the model dynamically allocates reasoning depth based on the session `effort` level rather than a caller-supplied token budget.

**Effort levels:**

| Level | Reasoning depth | Typical use |
|-------|----------------|-------------|
| `low` | Minimal | Trivial routing, completion of well-defined templates |
| `medium` | Moderate | Standard tasks, simple refactors, single-file edits |
| `high` | Substantial | Default for Opus 4.7. Implementation work, multi-file edits |
| `xhigh` | Extended | Reasoning-intensive analysis (manager-spec, manager-strategy, evaluator-active, expert-security, expert-refactoring) |
| `max` | Maximum | Reserved for the deepest analyses on Opus 4.7+ only |

**Rules:**

- Do NOT set fixed `thinking.budget_tokens` on Anthropic Messages API calls when targeting Opus 4.7. The API rejects fixed thinking budgets with **HTTP 400** for this model family.
- Allocate reasoning depth via the `effort` field on agent or skill frontmatter, not via `budget_tokens`.
- `xhigh` and `max` require Opus 4.7 or higher. Older models silently fall back to `high`.
- Sequential Thinking MCP (below) and Adaptive Thinking are complementary: Sequential Thinking structures the reasoning process step-by-step; Adaptive Thinking determines how much reasoning capacity each step receives.

**Anti-pattern (will fail on Opus 4.7):**

```jsonc
// DO NOT DO THIS for claude-opus-4-7
{
  "thinking": { "type": "enabled", "budget_tokens": 8000 }  // -> HTTP 400
}
```

**Correct pattern for Opus 4.7:**

Set effort at the skill or agent level (no API budget_tokens) and let the model self-allocate:

```yaml
---
name: my-deep-analysis-skill
effort: xhigh
---
```

## UltraThink Mode

Enhanced analysis mode activated by `--deepthink` flag.

**Activation:**
```
"Implement authentication system --deepthink"
"Refactor the API layer --deepthink"
```

**Process:**
1. Request Analysis: Identify core task, detect keywords, recognize complexity
2. Sequential Thinking: Begin structured reasoning
3. Execution Planning: Map subtasks to agents, identify parallel opportunities
4. Execution: Launch agents, integrate results

**UltraThink Parameters:**

Initial Analysis:
```
thought: "Analyzing user request: [content]"
nextThoughtNeeded: true
thoughtNumber: 1
totalThoughts: [estimate]
```

Subtask Decomposition:
```
thought: "Breaking down: 1) [task1] 2) [task2] 3) [task3]"
nextThoughtNeeded: true
thoughtNumber: 2
```

Agent Mapping:
```
thought: "Mapping: [task1] → expert-backend, [task2] → expert-frontend"
nextThoughtNeeded: true
thoughtNumber: 3
```

Execution Strategy:
```
thought: "Strategy: [tasks1,2] parallel, [task3] depends on [task1]"
nextThoughtNeeded: true
thoughtNumber: 4
```

Final Plan:
```
thought: "Plan: Launch [agents] in parallel, then [agent]"
nextThoughtNeeded: false
```

## When to Use

**UltraThink is ideal for:**
- Complex multi-domain tasks (backend + frontend + testing)
- Architecture decisions affecting multiple files
- Performance optimization requiring analysis
- Security review needs
- Refactoring with behavior preservation

**Benefits:**
- Structured decomposition of complex problems
- Explicit agent-task mapping with justification
- Identification of parallel execution opportunities
- Context maintenance throughout reasoning
- Revision capability when approaches need adjustment

## Guidelines

1. Start with reasonable totalThoughts estimate
2. Use isRevision when correcting previous thoughts
3. Maintain thoughtNumber sequence
4. Set nextThoughtNeeded to false only when complete
5. Use branching for exploring alternatives

<!-- ae:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "`--deepthink` and `ultrathink` are the same thing" | They are distinct mechanisms. `--deepthink` invokes the Sequential Thinking MCP tool (structured step-by-step reasoning chain). `ultrathink` is a keyword that sets `effort: max` in Claude Code, enabling Adaptive Thinking on Opus 4.7. Confusing them leads to wrong tool selection. |
| "I need to set `budget_tokens` to make Opus 4.7 think harder" | Opus 4.7 rejects fixed `budget_tokens` with HTTP 400. Reasoning depth is controlled by the `effort` field in agent or skill frontmatter, not by API-level token budgets. |
| "Sequential Thinking is too slow, I'll skip it for architecture decisions" | Architecture decisions affect many files and sessions. The overhead of Sequential Thinking is negligible compared to the cost of a wrong architecture choice discovered during implementation. |
| "Low complexity tasks don't need `effort: xhigh`" | Correct — but reasoning-intensive agents (manager-spec, manager-strategy, evaluator-active, expert-security, expert-refactoring) always need `effort: xhigh` regardless of individual task complexity. |
| "I can use Sequential Thinking MCP inside a GLM API session" | Sequential Thinking MCP generates `server_tool_use` content blocks that are not compatible with the GLM API. In CG mode, restrict `--deepthink` to the Claude leader pane. |

<!-- ae:evolvable-end -->

<!-- ae:evolvable-start id="red-flags" -->
## Red Flags

- Agent frontmatter sets `thinking: {type: enabled, budget_tokens: N}` for Opus 4.7 (will cause HTTP 400)
- Sequential Thinking used for single-file edits or trivial routing decisions
- `--deepthink` flag passed to an agent that lacks `mcp__sequential-thinking__sequentialthinking` in allowed-tools
- `effort` field missing from reasoning-intensive agent frontmatter (manager-spec, manager-strategy, evaluator-active, expert-security, expert-refactoring)
- UltraThink (`ultrathink` keyword) confused with `--deepthink` flag in user instructions

<!-- ae:evolvable-end -->

<!-- ae:evolvable-start id="verification" -->
## Verification

- [ ] `effort` field present in frontmatter for all reasoning-intensive agents (xhigh or max)
- [ ] No `budget_tokens` field in API calls targeting Opus 4.7 (grep for `budget_tokens` in agent prompts)
- [ ] `mcp__sequential-thinking__sequentialthinking` listed in allowed-tools when `--deepthink` is used
- [ ] Sequential Thinking invoked only for decisions affecting 3+ files or multi-domain scope
- [ ] `dart` language key used in lsp.yaml.tmpl (not `flutter`) — confirm with grep

<!-- ae:evolvable-end -->
