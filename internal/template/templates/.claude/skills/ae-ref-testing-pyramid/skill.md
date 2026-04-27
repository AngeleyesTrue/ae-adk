---
name: ae-ref-testing-pyramid
description: >
  Test pyramid strategy, coverage targets, test patterns, and quality metrics
  reference. Agent-extending skill that amplifies expert-testing and manager-tdd
  expertise with production-grade testing patterns.
  NOT for: production code implementation, architecture design, DevOps, security audits.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-03-30"
  tags: "testing, pyramid, coverage, tdd, patterns, reference"
  agent: "expert-testing"

# AE Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 900

# AE Extension: Triggers
triggers:
  keywords: ["test pyramid", "test strategy", "coverage target", "tdd workflow", "test patterns"]
  agents: ["expert-testing", "manager-tdd"]
  phases: ["run"]
---

# Testing Pyramid Reference

## Target Agents

- `expert-testing` - Primary: applies patterns during test creation and coverage analysis
- `manager-tdd` - Secondary: applies during RED-GREEN-REFACTOR cycles

## Test Pyramid Ratios

```
       /  E2E  \        10% — Critical user journeys only
      /----------\
     / Integration \    20% — API endpoints, DB queries, service boundaries
    /----------------\
   /    Unit Tests    \  70% — Functions, hooks, utilities, pure logic
  /--------------------\
```

| Level | Speed | Reliability | Maintenance | Coverage Target |
|-------|-------|-------------|-------------|-----------------|
| Unit | Fast (<100ms) | High | Low | 70% of tests |
| Integration | Medium (1-5s) | Medium | Medium | 20% of tests |
| E2E | Slow (10-60s) | Lower | High | 10% of tests |

## Coverage Targets by Context

| Context | Target | Rationale |
|---------|--------|-----------|
| Critical business logic | 95%+ | Revenue/security impact |
| API endpoints | 90%+ | Contract compliance |
| Utility functions | 85%+ | Reuse reliability |
| UI components | 80%+ | Rendering correctness |
| Configuration/glue code | 60%+ | Low complexity |
| Generated code | 0% | Don't test generated code |

## Test Pattern: AAA (Arrange-Act-Assert)

```
// Arrange: Set up test data and preconditions
input := CreateTestUser("test@example.com")

// Act: Execute the function under test
result, err := service.CreateUser(ctx, input)

// Assert: Verify the outcome
assert.NoError(t, err)
assert.Equal(t, "test@example.com", result.Email)
```

## Unit Test Patterns

| Pattern | When | Example |
|---------|------|---------|
| Table-Driven | Multiple input/output combinations | Go: `tests := []struct{...}` |
| Mock/Stub | External dependencies (DB, API) | Interface injection, mock frameworks |
| Snapshot | Complex output comparison | Jest snapshots, golden files |
| Property-Based | Mathematical properties | quickcheck, hypothesis |
| Boundary Value | Edge cases | 0, -1, MAX_INT, empty string, nil |

## Integration Test Patterns

| Pattern | When | Example |
|---------|------|---------|
| Testcontainers | Real DB needed | Docker-based PostgreSQL for tests |
| HTTP Test Server | API endpoint testing | httptest.NewServer (Go), supertest (Node) |
| In-Memory DB | Fast DB tests | SQLite for development |
| Fixture Loading | Consistent test data | Factory functions, seed files |

## What to Test vs What NOT to Test

### ALWAYS Test
- Business logic and calculations
- Input validation and error handling
- Authentication and authorization flows
- Data transformations and mappings
- Edge cases and boundary conditions
- Race conditions (with -race flag in Go)

### NEVER Test
- Framework internals (React rendering, Express routing)
- Third-party library behavior
- Simple getters/setters with no logic
- Private methods directly (test via public API)
- Generated code (protobuf, swagger)
- CSS styling and layout (use visual regression tools instead)

## Test Quality Metrics

| Metric | Target | Tool |
|--------|--------|------|
| Line Coverage | 85%+ | go test -cover, istanbul, coverage.py |
| Branch Coverage | 75%+ | go test -covermode=count |
| Mutation Score | 70%+ | go-mutesting, Stryker |
| Test Execution Time | <2 min (unit), <10 min (all) | CI timer |
| Flaky Test Rate | <1% | CI history analysis |

## Test File Conventions

| Language | Test File | Location |
|----------|-----------|----------|
| Go | `*_test.go` | Same package |
| TypeScript | `*.test.ts` / `*.spec.ts` | `__tests__/` or co-located |
| Python | `test_*.py` | `tests/` directory |
| Java | `*Test.java` | `src/test/` mirror |
| Rust | `#[cfg(test)] mod tests` | Same file or `tests/` |

## TDD RED-GREEN-REFACTOR Quick Reference

```
RED:     Write a failing test that defines expected behavior
GREEN:   Write minimal code to make the test pass
REFACTOR: Clean up while keeping tests green
```

Rules:
- Never write production code without a failing test
- Write the smallest test that fails
- Write the simplest code that passes
- Refactor only when all tests are green
- One assertion per test (when practical)

<!-- ae:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "E2E tests cover everything, unit tests are redundant" | E2E is slow and flaky. The pyramid (many unit, some integration, few E2E) is shaped by feedback speed and stability. |
| "100% coverage means the code works" | Coverage measures execution, not assertion quality. Tests that execute without asserting anything still hit 100%. |
| "Mocking everything makes tests fast" | Over-mocking makes tests reflect the mocks, not the system. Mock at boundaries (HTTP, DB), test real logic. |

<!-- ae:evolvable-end -->

<!-- ae:evolvable-start id="red-flags" -->
## Red Flags

- Test pyramid inverted (many E2E, few unit) — slow and flaky CI
- Test files without explicit assertions
- Mocks at every layer, including pure functions
- Coverage > 90% but bugs ship to production routinely
- No characterization tests added when modifying legacy code

<!-- ae:evolvable-end -->

<!-- ae:evolvable-start id="verification" -->
## Verification

- [ ] Test pyramid shape verified: many unit, some integration, few E2E
- [ ] Coverage >= 85% for new code, with assertion quality reviewed
- [ ] Mocks scoped to true boundaries (network, filesystem, time)
- [ ] Legacy code modifications gated by characterization tests
- [ ] Flaky tests quarantined with an issue link, not silently retried

<!-- ae:evolvable-end -->
