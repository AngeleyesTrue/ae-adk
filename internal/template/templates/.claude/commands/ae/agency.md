---
allowed-tools: Agent, AskUserQuestion, TaskCreate, TaskUpdate, TaskList, TaskGet, Bash, Read, Write, Edit, Glob, Grep
---

# Agency Pipeline

Self-evolving creative production system for websites and web applications.

## Usage

$ARGUMENTS

## Pipeline

Planner -> [Copywriter, Designer] (parallel) -> Builder -> Evaluator -> Learner

## Subcommands

- brief: Create project brief from user request
- build: Execute full pipeline (brief -> build -> evaluate)
- review: Re-evaluate existing build
- learn: Process learnings from evaluation
- evolve: Apply graduated learnings
- resume: Continue interrupted pipeline
- profile: Show/set agency configuration
- sync-upstream: Check for upstream skill updates
- rollback: Revert last evolution
- config: Edit agency configuration

## Default Behavior

Without subcommand, routes to full pipeline execution (brief -> build -> evaluate -> learn).
