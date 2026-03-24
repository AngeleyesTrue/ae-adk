---
id: SPEC-REFACTOR-002
version: "1.0.0"
status: draft
created: "2026-03-23"
updated: "2026-03-23"
author: Angeleyes
priority: high
issue_number: 0
depends_on: SPEC-REFACTOR-001
---

# SPEC-REFACTOR-002: cc/glm/cg 런치 명령어 삭제

## 1. 환경 (Environment)

### 프로젝트 컨텍스트

- **프로젝트**: ae-adk (Go CLI 도구, moai-adk 포크)
- **언어/런타임**: Go 1.26+, Cobra CLI 프레임워크
- **대상 명령어**: `ae cc`, `ae glm`, `ae cg`
- **연관 TODO**: `docs/TODO.md` 항목 3
- **선행 SPEC**: SPEC-REFACTOR-001 (rank 시스템 제거)

### 현재 코드 구조

| 파일 | 줄 수 | 역할 |
|------|--------|------|
| `internal/cli/cc.go` | 60 | Claude 모드 런처 (`ae cc`) |
| `internal/cli/glm.go` | 823 | GLM 모드 런처 (`ae glm`) + setup/status 서브커맨드 |
| `internal/cli/cg.go` | 55 | Claude+GLM 하이브리드 모드 (`ae cg`) |
| `internal/cli/launcher.go` | 577 | 통합 런치 로직 (모드별 함수 포함) |
| `internal/hook/session_end.go` | 651 | 세션 종료 시 GLM 환경 정리 |
| `internal/cli/deps.go` | 355 | 의존성 구성 루트 |

### 불변 제약

- `.claude/`, `.moai/`, `CLAUDE.md` 파일은 moai-adk 영역이므로 **절대 수정 금지**
- `ae` 바이너리의 다른 명령어(`init`, `update`, `hook`, `worktree`, `doctor` 등)는 영향 없어야 함

## 2. 가정 (Assumptions)

### 검증된 가정

| 가정 | 신뢰도 | 근거 |
|------|---------|------|
| cc/glm/cg 명령어는 ae-adk 전용이며 moai-adk에는 없음 | 높음 | moai-adk는 Python 기반이며 Go CLI 런처가 없음 |
| launcher.go의 함수들은 cc/glm/cg 에서만 호출됨 | 높음 | `@MX:ANCHOR` 주석에서 fan_in=3 (runCC, runCG, runGLM) 확인 |
| session_end.go의 GLM 정리 로직은 GLM 모드 전용임 | 높음 | `clearTmuxSessionEnv`, `cleanupGLMSettingsLocal` 함수명과 주석으로 확인 |
| rank/ralph 코드는 SPEC-REFACTOR-001에서 이미 제거됨 | 높음 | SPEC-REFACTOR-001 완료가 선행 조건 |

### 주의 필요 가정

| 가정 | 신뢰도 | 검증 방법 |
|------|---------|-----------|
| launcher.go 파일 전체 삭제 가능 여부 | 중간 | `launchClaude()`, `parseProfileFlag()` 등 범용 함수가 다른 곳에서 사용되는지 확인 필요 |
| `oauth_token_preservation_test.go`에 GLM 참조 존재 여부 | 중간 | 파일 내용 검사 필요 |
| root.go help 텍스트에 cc/glm/cg 언급이 있음 | 높음 | `Use 'ae cc', 'ae cg', or 'ae glm' to launch Claude Code.` 확인됨 |

## 3. 요구사항 (Requirements)

### REQ-001: 소스 파일 삭제 (Ubiquitous)

시스템은 **항상** 다음 소스 파일을 완전히 삭제해야 한다:

- `internal/cli/cc.go`
- `internal/cli/glm.go`
- `internal/cli/cg.go`

### REQ-002: 테스트 파일 삭제 (Ubiquitous)

시스템은 **항상** 다음 테스트 파일을 완전히 삭제해야 한다:

- `internal/cli/cc_test.go`
- `internal/cli/glm_test.go`
- `internal/cli/glm_compat_test.go`
- `internal/cli/glm_model_override_test.go`
- `internal/cli/glm_new_test.go`
- `internal/cli/glm_team_test.go`

### REQ-003: launcher.go 정리 (Event-Driven)

**WHEN** cc/glm/cg 소스 파일이 삭제되면 **THEN** `internal/cli/launcher.go`에서:

- `applyCCMode()`, `applyGLMMode()`, `applyCGMode()` 함수 제거
- `unifiedLaunch()`, `unifiedLaunchDefault()`, `unifiedLaunchFunc` 변수 제거
- `resolveMode()` 함수 제거
- GLM 관련 헬퍼 함수 제거: `removeGLMEnv()`, `setGLMEnv()`, `loadGLMConfig()`, `getGLMAPIKey()`, `injectGLMEnvForTeam()`, `injectTmuxSessionEnv()`
- `resetTeamModeForCC()`, `cleanupAeWorktrees()` 제거 여부 검토
- `launchClaude()`, `launchClaudeDefault()`, `parseProfileFlag()` 등 범용 함수는 다른 곳에서 참조되는지 확인 후 결정
- **파일 전체 삭제가 가능하면 파일 삭제 우선**, 일부 범용 함수가 남으면 해당 함수만 보존

### REQ-004: launcher_test.go 정리 (Event-Driven)

**WHEN** launcher.go가 수정 또는 삭제되면 **THEN** `internal/cli/launcher_test.go`도 동일하게 정리한다.

### REQ-005: root.go 명령어 등록 제거 (Event-Driven)

**WHEN** cc/glm/cg 커맨드가 삭제되면 **THEN** `internal/cli/root.go`에서:

- `rootCmd.Long` 설명 텍스트에서 `Use 'ae cc', 'ae cg', or 'ae glm' to launch Claude Code.` 문구 제거 또는 수정
- `"launch"` 커맨드 그룹 제거 검토 (다른 launch 명령어가 없다면 그룹도 제거)
- **참고**: `ccCmd`, `glmCmd`, `cgCmd`는 각각의 `.go` 파일 `init()`에서 `rootCmd.AddCommand()`를 호출하므로, 파일 삭제 시 자동으로 등록이 해제됨

### REQ-006: session_end.go GLM 정리 로직 제거 (Event-Driven)

**WHEN** GLM 모드가 삭제되면 **THEN** `internal/hook/session_end.go`에서:

- `clearTmuxSessionEnv()` 함수 및 호출부 제거
- `cleanupGLMSettingsLocal()` 함수 및 호출부 제거
- `glmEnvVarsToClean` 변수 제거
- GLM 관련 주석 정리

### REQ-007: deps.go GLM 의존성 제거 (Event-Driven)

**WHEN** GLM 모드가 삭제되면 **THEN** `internal/cli/deps.go`에서:

- GLM 전용 import 및 사용처 제거
- **주의**: rank/ralph 관련 의존성은 SPEC-REFACTOR-001에서 이미 제거됨

### REQ-008: 빌드 무결성 보장 (Unwanted)

시스템은 리팩토링 후 빌드 실패가 **발생하지 않아야 한다**:

- `go build ./...` 성공
- `go vet ./...` 경고 없음
- 미사용 import 없음
- 미사용 변수/함수 없음

### REQ-009: 기존 테스트 통과 보장 (Unwanted)

시스템은 리팩토링 후 기존 테스트 실패가 **발생하지 않아야 한다**:

- `go test ./...` 전체 통과 (삭제된 테스트 제외)
- cc/glm/cg와 무관한 테스트는 결과 변동 없음

### REQ-010: moai-adk 영역 불변 (Unwanted)

시스템은 다음 경로의 파일을 **수정하지 않아야 한다**:

- `.claude/` 디렉토리 내 모든 파일
- `.moai/` 디렉토리 내 설정/스킬 파일 (specs 제외)
- `CLAUDE.md`

### REQ-011: oauth_token_preservation_test.go 검토 (Optional)

**가능하면** `internal/cli/oauth_token_preservation_test.go`에서 GLM 참조를 확인하고 정리를 제공한다.

## 4. 사양 (Specifications)

### 삭제 대상 파일 목록 (9개)

| 구분 | 파일 경로 |
|------|-----------|
| 소스 | `internal/cli/cc.go` |
| 소스 | `internal/cli/glm.go` |
| 소스 | `internal/cli/cg.go` |
| 테스트 | `internal/cli/cc_test.go` |
| 테스트 | `internal/cli/glm_test.go` |
| 테스트 | `internal/cli/glm_compat_test.go` |
| 테스트 | `internal/cli/glm_model_override_test.go` |
| 테스트 | `internal/cli/glm_new_test.go` |
| 테스트 | `internal/cli/glm_team_test.go` |

### 수정 대상 파일 목록 (5개)

| 파일 | 수정 내용 |
|------|-----------|
| `internal/cli/launcher.go` | cc/glm/cg 모드 함수 제거 또는 파일 전체 삭제 |
| `internal/cli/launcher_test.go` | 런처 테스트 정리 또는 파일 삭제 |
| `internal/cli/root.go` | help 텍스트에서 cc/glm/cg 언급 제거, launch 그룹 제거 검토 |
| `internal/hook/session_end.go` | GLM 정리 함수 및 변수 제거 |
| `internal/cli/deps.go` | GLM 전용 의존성 제거 |

### 검토 대상 파일 (1개)

| 파일 | 검토 내용 |
|------|-----------|
| `internal/cli/oauth_token_preservation_test.go` | GLM 참조 유무 확인 후 정리 |

### 기술적 의사결정 포인트

1. **launcher.go 전체 삭제 vs 부분 수정**: `launchClaude()`, `parseProfileFlag()` 등 범용 함수가 cc/glm/cg 외부에서 사용되는지 구현 시 확인 필요. 외부 참조가 없으면 파일 전체 삭제 가능.

2. **launch 그룹 존속 여부**: cc/glm/cg가 유일한 launch 그룹 명령어라면 `root.go`에서 `launch` 그룹 자체를 제거.

3. **tmux 관련 코드 범위**: `clearTmuxSessionEnv()`는 session_end.go에만 존재하므로 GLM과 함께 삭제. `internal/tmux/` 패키지 자체의 존속은 다른 사용처 확인 후 결정 (이 SPEC 범위 밖).

## 5. 추적성 (Traceability)

| 태그 | 참조 |
|------|------|
| SPEC-REFACTOR-002 | 본 문서 |
| TODO-3 | `docs/TODO.md` 항목 3 "cc/glm/cg 명령어 삭제" |
| REQ-001 ~ REQ-011 | 본 문서 요구사항 섹션 |
