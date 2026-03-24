# Changelog

All notable changes to AE-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `ae win` / `ae mac` 플랫폼 전환 및 진단 명령어 (SPEC-PLATFORM-001)
  - settings.json PATH 자동 재구성 (BuildSmartPATH 활용)
  - Windows 진단: UTF-8, MCP 서버 경로, Git Bash, WSL2, LongPaths, Hook bash
  - macOS 진단: Homebrew, 심볼릭 링크, 셸 호환성
  - 공통: 도구 버전 확인 (ae, go, node, git)
  - settings.json 백업 (타임스탬프 기반, 최근 5개 유지)
  - 플랫폼 프로필 저장/비교 (~/.ae/platform-profile.json)
  - 플래그: --force, --verbose, --json, --auto, --dry-run, --skip-backup
- `internal/platform` 패키지 신규 추가 (97.8% 테스트 커버리지)

### Removed

- cc/glm/cg 런치 명령어 완전 삭제 (SPEC-REFACTOR-002)

### Fixed

- .mcp.json: cmd.exe 경유 → npx 직접 호출로 변경 (MCP 서버 연결 실패 해결)
- settings.json: PATH에 nodejs/npm/bun 경로 추가

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
