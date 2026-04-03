# Commit Convention Guide

## Active Convention: Bracket-Scope

When writing commit messages, use the bracket-scope format:

```
type: [Scope] description
```

### Format Rules

- Type: lowercase (feat, fix, refactor, docs, test, chore, build, ci, perf, revert, style)
- Scope: PascalCase, enclosed in square brackets, optional
- Multi-scope: slash-separated (e.g., [Web/Auth])
- Breaking change: add `!` after type (e.g., `feat!: [API] change endpoint`)
- Header max length: 100 characters
- Description must not start with `[` or `]`
- Space required after closing bracket

### Examples

Valid:
- `feat: [Web] add restore button`
- `fix: [Auth/Session] resolve token expiration`
- `refactor: [Core/DB] optimize query performance`
- `feat!: [API] change login endpoint signature`
- `chore: update dependencies` (no scope is valid)

Invalid:
- `feat(scope): description` (use square brackets, not parentheses)
- `feat: [Web]description` (missing space after bracket)

### Scope Query

When commit scopes are not configured in git-strategy.yaml:
- Query the user for appropriate scope names via AskUserQuestion
- Suggest default scopes: Tests, Docs, Build, DB, Auth, Solution
- Save the response to `git-strategy.yaml` under `commit_scopes`

### Body Format

```
type: [Scope] short description

- Detail change 1
- Detail change 2
```

Use the language configured in `language.yaml` `git_commit_messages` setting.
