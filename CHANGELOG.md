# Changelog

All notable changes to AE-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-18

### Added

- AE-ADK initial release (AngelEyes Agentic Development Kit)
- Windows compatibility: `BuildSmartPATH()` with `case "windows"` support
- Windows MCP server: `pwsh.exe` based template for `.mcp.json`
- Cross-platform Makefile with `internal/cmd/datestamp` utility
- Platform auto-detection via `runtime.GOOS` fallback in `WithPlatform()`
- Test environment at `tests/ae-adk-test/` for template verification
- Windows compatibility documentation (`docs/windows-compat-ae.md`, `docs/windows-compat-moai.md`)

### Changed

- All template files renamed from moai to ae naming convention
- Template directory `.moai/` → `.ae/` for ae-adk installed projects
- Output style name: `MoAI` → `AE`
- Hook directory: `hooks/moai/` → `hooks/ae/`
- Config env key: `MOAI_CONFIG_SOURCE` → `AE_CONFIG_SOURCE`
- Git branch prefix: `moai/` → `ae/`
- Tmux session prefix: `moai-` → `ae-`
- Install path: `Programs\moai` → `Programs\ae`
- GitHub repository: `modu-ai/moai-adk` → `AngeleyesTrue/ae-adk`
