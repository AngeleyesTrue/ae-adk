# Changelog

All notable changes to AE-ADK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **SPEC-UPDATE-004**: moai-adk v2.10.2~v2.13.2 업스트림 반영 (품질·성능)
  - Effort 시스템: `EffortLevel{Low,Medium,High,XHigh,Max}` 5단계 + `ModelIDOpus47` 상수, `agentEffortMap` 6개 매핑(manager-spec/strategy=xhigh, evaluator-active/expert-security/expert-refactoring/builder-agent=high), `ApplyEffortPolicy` 보존+idempotency
  - LSP 템플릿: `.ae/config/sections/lsp.yaml.tmpl` 신규 (16개 언어 + default, `command:` 키 통일, `file_extensions` 필드)
  - LSP compliance 회귀 테스트 3종: `TestLSPSchemaKeyCommand`, `TestLSPFileExtensionsComplete`, `TestLSPStructTagConsistency`
  - HUMAN GATE 회귀 테스트: `internal/template/human_gate_regression_test.go` (auto.md/auto-sync.md에 H2/H3 헤더 부재 강제, plan.md/run.md/sync.md에 2개 이상 존재 강제)
  - Profile setup wizard 강화: `claude-opus-4-7` 모델 옵션, 5단계 effort 선택기, `normalizeModel` 헬퍼 (17 시나리오 단위 테스트), 4개 언어(ko/en/ja/zh) statusline 마이그레이션 배너, `auto` permission mode 옵션
  - Doctor 강화: `checkMCPScopeDuplicates` 함수 — 프로젝트 `.mcp.json`과 글로벌 `~/.claude/.mcp.json` 스코프 충돌 감지(warning, exit 0)
  - PermissionRequest 훅: `__updated_input_marker__` 센티넬 거부 로직
  - SessionStart 훅: Windows 전용 `injectCLAUDEEnvFile` (settings.local.json에 `CLAUDE_ENV_FILE` 주입, macOS/Linux 영향 없음)
  - 규칙 문서 갱신: `ae-constitution.md`에 Opus 4.7 Prompt Philosophy 5원칙 섹션, `agent-authoring.md`에 Bash Tool Timeout Ceiling(600,000ms) 섹션, `skill-authoring.md`에 effort 5단계 설명, `ae-workflow-thinking/SKILL.md` Adaptive Thinking 재정의
  - 템플릿 기본값 변경: `llm.yaml`의 `claude_models.high = claude-opus-4-7`, `workflow.yaml` role_profiles(team_lead=opus[1m], reviewer=sonnet)
  - 6개 에이전트 effort 필드 복구: manager-spec, manager-strategy, evaluator-active, expert-security, expert-refactoring, builder-agent
  - 40개 스킬 evolvable blocks 3종(Common Rationalizations, Red Flags, Verification) 복구
  - 16-language 중립성 테이블 복구: `loop.md`, `references/examples.md`, `references/reference.md`
  - `settings.json.tmpl`: `disableBypassPermissionsMode: false` 필드 추가
  - 하드코딩 상수 12건 집중화: `internal/foundation/{constants,envvars}.go`
  - `GateConfig.AstGrepGate` 필드 추가 (Self-Learning Quality Guard)
  - ESLint 품질 게이트 Python-only 오차단 수정
- 후속 SPEC 4건 제안서: `docs/future-specs/` (SPEC-UPDATE-005, SPEC-UPDATE-006, SPEC-LSP-CORE-002, SPEC-SECURITY-BYPASS-001)
- `ae win` / `ae mac` 플랫폼 전환 및 진단 명령어 (SPEC-PLATFORM-001)
  - settings.json PATH 자동 재구성 (BuildSmartPATH 활용)
  - Windows 진단: UTF-8, MCP 서버 경로, Git Bash, WSL2, LongPaths, Hook bash
  - macOS 진단: Homebrew, 심볼릭 링크, 셸 호환성
  - 공통: 도구 버전 확인 (ae, go, node, git)
  - settings.json 백업 (타임스탬프 기반, 최근 5개 유지)
  - 플랫폼 프로필 저장/비교 (~/.ae/platform-profile.json)
  - 플래그: --force, --verbose, --json, --auto, --dry-run, --skip-backup
- `internal/platform` 패키지 신규 추가 (97.8% 테스트 커버리지)

### Changed

- **SPEC-UPDATE-004**: REQ-23 (powernap 도입) hold 처리 — go.mod 부재 확인, SPEC-LSP-CORE-002로 분리
- **SPEC-UPDATE-004**: REQ-24 7개 → 6개 매핑 축소 — plan-auditor 템플릿 부재로 별도 SPEC 분리
- **SPEC-UPDATE-004**: 품질·성능 영역만 반영, agency/db/design 재편은 SPEC-UPDATE-005 이월(EX-01~11), v2.14.0 Utility Hardening은 SPEC-UPDATE-006 이월(EX-13)

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
