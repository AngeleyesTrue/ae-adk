---
name: ae-workflow-research
description: >
  Self-research workflow for optimizing ae-adk components through
  binary eval experimentation loops. Adapted from autoresearch pattern
  with 5-layer safety architecture.
user-invocable: false
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "experimental"
  updated: "2026-04-09"
  tags: "research, eval, experiment, self-improvement, autoresearch"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["research", "eval", "experiment", "optimize skill", "improve agent"]
  agents: ["researcher"]
  phases: ["research"]
---

# Research Workflow

## Purpose

Optimize ae-adk components (skills, agents, rules, config) through iterative binary eval experimentation.

## Data Locations

| Data | Location |
|------|----------|
| Eval suites | `.ae/research/evals/{type}/{name}.eval.yaml` |
| Baselines | `.ae/research/baselines/{name}.baseline.json` |
| Experiments | `.ae/research/experiments/{name}/exp-NNN.json` |
| Changelogs | `.ae/research/experiments/{name}/changelog.md` |
| Observations | `.ae/research/observations/` |
| Dashboard | `ae research status` (CLI) |

## Eval Suite Schema

```yaml
target:
  path: .claude/skills/ae-lang-go/SKILL.md
  type: skill
test_inputs:
  - name: scenario-name
    prompt: "Test prompt for evaluation"
evals:
  - name: criterion-name
    question: "Binary yes/no question?"
    pass: "What yes looks like"
    fail: "What triggers no"
    weight: must_pass  # or nice_to_have
settings:
  runs_per_experiment: 3
  max_experiments: 20
  pass_threshold: 0.80
  target_score: 0.95
```

## Safety Layers

1. **FrozenGuard**: Constitution files cannot be modified
2. **Worktree Sandbox**: All experiments in isolated worktrees
3. **Canary Regression**: Proposed changes tested against baselines
4. **Rate Limiter**: Max experiments per session/week
5. **Human Approval**: Required before merging to main
