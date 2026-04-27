---
name: ae-examples
description: >
  Concrete usage examples for AE workflows organized by language and scenario.
  Provides copy-pasteable invocation patterns for /ae plan, /ae run, /ae sync,
  /ae fix, /ae loop, and /ae project across the 16 supported languages.
  Use when looking for a worked example for a specific language or workflow.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "foundation"
  status: "active"
  updated: "2026-04-27"
  tags: "examples, samples, workflow, patterns, languages"

# AE Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# AE Extension: Triggers
triggers:
  keywords: ["example", "sample", "how-to", "demo", "language"]
  agents: ["manager-spec", "manager-ddd", "manager-tdd", "manager-docs", "manager-quality"]
  phases: ["plan", "run", "sync"]
---

# AE Skill Examples

Worked examples for AE workflows. Use this reference when looking for a copy-pasteable invocation pattern for a specific language or workflow phase.

For execution patterns and flag reference, see: [reference.md](reference.md)
For MX tag examples, see: [mx-tag.md](mx-tag.md)

---

## Language Neutrality

ae-adk supports 16 programming languages on equal footing. The examples below cover each language with consistent toolchain assumptions: an LSP server for diagnostics, a test framework for verification, and a lint tool for style enforcement.

| Language | File Extensions | LSP Server | Test Framework | Lint Tool |
|----------|----------------|------------|----------------|-----------|
| go | `.go` | `gopls` | `go test` | `golangci-lint` |
| python | `.py`, `.pyi` | `pyright` or `pylsp` | `pytest` | `ruff` |
| typescript | `.ts`, `.tsx`, `.mts`, `.cts` | `typescript-language-server` | `jest` or `vitest` | `eslint` |
| javascript | `.js`, `.jsx`, `.mjs`, `.cjs` | `typescript-language-server` | `jest` or `vitest` | `eslint` |
| rust | `.rs` | `rust-analyzer` | `cargo test` | `cargo clippy` |
| java | `.java` | `jdtls` | `junit` (Maven/Gradle) | `checkstyle` or `spotbugs` |
| kotlin | `.kt`, `.kts` | `kotlin-language-server` | `junit` (Gradle) | `ktlint` or `detekt` |
| swift | `.swift` | `sourcekit-lsp` | `swift test` | `swiftlint` |
| ruby | `.rb`, `.rake`, `.gemspec` | `solargraph` or `ruby-lsp` | `rspec` or `minitest` | `rubocop` |
| php | `.php`, `.phtml` | `intelephense` | `phpunit` | `phpstan` or `php-cs-fixer` |
| cpp | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.h` | `clangd` | `gtest` (CTest) | `clang-tidy` |
| csharp | `.cs` | `omnisharp` | `dotnet test` (xUnit/NUnit) | `dotnet format` / `roslynator` |
| scala | `.scala`, `.sc` | `metals` | `scalatest` (sbt) | `scalafmt` / `scalafix` |
| elixir | `.ex`, `.exs` | `elixir-ls` | `mix test` (ExUnit) | `credo` |
| r | `.r`, `.R` | `languageserver` | `testthat` | `lintr` |
| dart | `.dart` | `dart language-server` (Dart SDK) | `dart test` / `flutter test` | `dart analyze` |

Language detection priority: Inspect indicator files in the project root (`go.mod`, `pyproject.toml`, `package.json`, `Cargo.toml`, etc.). When multiple are present, prefer the one with the most associated source files. If detection fails, the orchestrator prompts the user to specify the language explicitly.

---

## Workflow Examples

### Example 1: Create a SPEC for a new feature

```
/ae plan "Add JWT-based authentication to the API"
```

The manager-spec subagent produces `.ae/specs/SPEC-AUTH-001/spec.md` with EARS-format requirements and acceptance criteria.

### Example 2: Run an existing SPEC with TDD

```
/ae run SPEC-AUTH-001
```

The manager-tdd subagent (per `quality.yaml development_mode`) executes RED-GREEN-REFACTOR until acceptance criteria pass.

### Example 3: Sync documentation and create a PR

```
/ae sync SPEC-AUTH-001
```

The manager-docs subagent updates README, CHANGELOG, and creates the pull request.

### Example 4: Fix all LSP errors in the project

```
/ae fix
```

Detects the project language, runs the appropriate LSP, and applies safe auto-fixes.

### Example 5: Iterative quality improvement

```
/ae loop --max 5 --auto-fix
```

Runs LSP -> AST-grep -> Tests -> Coverage in a loop until all pass or max iterations reached.

### Example 6: Generate project documentation

```
/ae project
```

Produces `structure.md`, `tech.md`, and `codemaps/` based on detected language and architecture.

---

## Per-Language SPEC Examples

Each example assumes the project marker is in the repository root.

### go

```
# Project marker: go.mod
/ae plan "Add a /healthz endpoint that returns 200 OK with JSON {status: ok}"
/ae run SPEC-HTTP-001
# Verifies: go test ./..., golangci-lint run, gopls diagnostics
```

### python

```
# Project marker: pyproject.toml
/ae plan "Add a CLI command 'mycli stats' that prints repo line counts"
/ae run SPEC-CLI-001
# Verifies: pytest --tb=short, ruff check, pyright
```

### typescript

```
# Project marker: tsconfig.json + package.json
/ae plan "Add a useDebounce hook with TypeScript types"
/ae run SPEC-HOOK-001
# Verifies: npm test, eslint, tsc --noEmit
```

### javascript

```
# Project marker: package.json
/ae plan "Add an Express middleware for request logging"
/ae run SPEC-MW-001
# Verifies: npm test, eslint
```

### rust

```
# Project marker: Cargo.toml
/ae plan "Add a Result-returning parse function for ISO-8601 dates"
/ae run SPEC-PARSE-001
# Verifies: cargo test, cargo clippy -- -D warnings
```

### java

```
# Project marker: pom.xml or build.gradle
/ae plan "Add a UserService.findByEmail method with JPA"
/ae run SPEC-SVC-001
# Verifies: mvn test or gradle test, checkstyle
```

### kotlin

```
# Project marker: build.gradle.kts
/ae plan "Add a ktor route /users/{id} returning a User DTO"
/ae run SPEC-KTOR-001
# Verifies: gradle test, ktlint
```

### swift

```
# Project marker: Package.swift
/ae plan "Add a Codable User struct with default coding keys"
/ae run SPEC-CODABLE-001
# Verifies: swift test, swiftlint
```

### ruby

```
# Project marker: Gemfile
/ae plan "Add an ApplicationRecord scope :recent for the last 7 days"
/ae run SPEC-SCOPE-001
# Verifies: bundle exec rspec, rubocop
```

### php

```
# Project marker: composer.json
/ae plan "Add a UserRepository::findByEmail method using PDO"
/ae run SPEC-REPO-001
# Verifies: vendor/bin/phpunit, phpstan analyse
```

### cpp

```
# Project marker: CMakeLists.txt
/ae plan "Add a thread-safe LruCache class with unit tests"
/ae run SPEC-CACHE-001
# Verifies: ctest --test-dir build, clang-tidy
```

### csharp

```
# Project marker: *.csproj or *.sln
/ae plan "Add a WeatherForecastController returning a 7-day forecast"
/ae run SPEC-API-001
# Verifies: dotnet test, dotnet format
```

### scala

```
# Project marker: build.sbt
/ae plan "Add a HttpClient trait with a Future[Response] retrieval method"
/ae run SPEC-HTTP-002
# Verifies: sbt test, scalafmt
```

### elixir

```
# Project marker: mix.exs
/ae plan "Add a Phoenix LiveView Counter with increment/decrement"
/ae run SPEC-LV-001
# Verifies: mix test, mix credo
```

### r

```
# Project marker: DESCRIPTION
/ae plan "Add a function summarize_groups returning a tibble of group means"
/ae run SPEC-RFN-001
# Verifies: Rscript -e 'testthat::test_package(".")', lintr
```

### dart

```
# Project marker: pubspec.yaml
/ae plan "Add a CounterCubit with increment/decrement actions and stream tests"
/ae run SPEC-CUBIT-001
# Verifies: dart test or flutter test, dart analyze
```

---

## Resources

For detailed flag reference and execution patterns, see [reference.md](reference.md).
For MX tag annotation examples, see [mx-tag.md](mx-tag.md).
For workflow concepts and methodology selection, see `.claude/rules/ae/workflow/`.

---

Version: 1.0.0
Last Updated: 2026-04-27
