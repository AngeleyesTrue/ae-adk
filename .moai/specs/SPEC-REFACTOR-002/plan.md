---
spec_id: SPEC-REFACTOR-002
type: plan
version: "1.0.0"
---

# SPEC-REFACTOR-002 구현 계획: cc/glm/cg 런치 명령어 삭제

> **의존성**: 이 SPEC은 SPEC-REFACTOR-001 (rank 시스템 제거) 완료 후 실행해야 한다. SPEC-REFACTOR-001이 deps.go에서 rank/ralph 관련 코드를 먼저 제거하므로, 이 SPEC에서는 deps.go의 GLM 전용 코드만 제거하면 된다.

## 1. 개요

ae-adk CLI에서 `ae cc`, `ae glm`, `ae cg` 세 가지 런치 명령어와 관련 코드를 완전히 제거하는 리팩토링 작업이다. 이는 ae-adk가 moai-adk 포크로서 불필요해진 자체 런처 기능을 정리하는 것이 목적이다.

## 2. 마일스톤

### 마일스톤 1: 사전 분석 (Primary Goal)

**목표**: 삭제 전 의존성 그래프를 완전히 파악하여 안전한 삭제 순서를 결정한다.

- [ ] launcher.go 내 범용 함수(`launchClaude`, `parseProfileFlag`, `launchClaudeDefault`) 외부 참조 분석
- [ ] `internal/tmux/` 패키지가 GLM 이외에서 사용되는지 확인
- [ ] `oauth_token_preservation_test.go` 내 GLM 참조 확인
- [ ] `root.go`의 launch 그룹에 cc/glm/cg 외 다른 명령어가 등록되어 있는지 확인
- [ ] `internal/cli/deps.go`에서 GLM 전용 코드 식별
- [ ] session_end.go의 `clearTmuxSessionEnv` 및 `cleanupGLMSettingsLocal`이 GLM 전용인지 최종 확인

**완료 기준**: 삭제 대상 함수/변수/import 목록이 확정되고, launcher.go 전체 삭제 여부가 결정됨.

### 마일스톤 2: 소스 및 테스트 파일 삭제 (Primary Goal)

**목표**: 독립적인 cc/glm/cg 파일을 삭제한다.

- [ ] `internal/cli/cc.go` 삭제
- [ ] `internal/cli/glm.go` 삭제
- [ ] `internal/cli/cg.go` 삭제
- [ ] `internal/cli/cc_test.go` 삭제
- [ ] `internal/cli/glm_test.go` 삭제
- [ ] `internal/cli/glm_compat_test.go` 삭제
- [ ] `internal/cli/glm_model_override_test.go` 삭제
- [ ] `internal/cli/glm_new_test.go` 삭제
- [ ] `internal/cli/glm_team_test.go` 삭제

**완료 기준**: 9개 파일 삭제 완료.

### 마일스톤 3: 관련 파일 수정 (Primary Goal)

**목표**: 삭제된 코드를 참조하는 나머지 파일을 정리한다.

- [ ] `internal/cli/launcher.go` 수정 또는 삭제
  - 마일스톤 1 분석 결과에 따라 전체 삭제 또는 GLM/CC/CG 전용 함수만 제거
- [ ] `internal/cli/launcher_test.go` 수정 또는 삭제
- [ ] `internal/cli/root.go` 수정
  - `rootCmd.Long` 텍스트에서 cc/glm/cg 언급 제거
  - `"launch"` 커맨드 그룹 등록 제거 (다른 launch 명령어가 없는 경우)
- [ ] `internal/hook/session_end.go` 수정
  - `clearTmuxSessionEnv()` 함수 및 호출부 제거
  - `cleanupGLMSettingsLocal()` 함수 및 호출부 제거
  - `glmEnvVarsToClean` 변수 제거
- [ ] `internal/cli/deps.go` 수정
  - GLM 전용 import 및 코드 제거
  - rank/ralph 관련 코드는 SPEC-REFACTOR-001에서 이미 제거됨 (선행 완료 전제)

**완료 기준**: 모든 파일에서 삭제된 함수/변수 참조가 제거됨.

### 마일스톤 4: 빌드 및 테스트 검증 (Secondary Goal)

- [ ] `go build ./...` 성공 확인
- [ ] `go vet ./...` 경고 없음 확인
- [ ] `go test ./...` 전체 통과 확인 (삭제된 테스트 제외)
- [ ] `golangci-lint run ./...` 정적 분석 통과 확인

### 마일스톤 5: 선택적 정리 (Optional Goal)

- [ ] `oauth_token_preservation_test.go`에서 GLM 참조 정리 (해당되는 경우)
- [ ] `go mod tidy`
- [ ] 삭제된 코드와 관련된 `@MX` 태그 업데이트

## 3. 기술 접근법

### 삭제 전략: Bottom-Up 접근

1. **리프 노드 먼저 삭제**: 다른 코드에서 참조되지 않는 파일(cc.go, cg.go, glm.go 및 테스트)을 먼저 삭제
2. **중간 노드 정리**: launcher.go에서 삭제된 함수를 참조하는 코드 제거
3. **루트 노드 정리**: root.go의 등록/텍스트, session_end.go의 GLM 정리 로직, deps.go의 import 정리
4. **검증**: 빌드 및 테스트 통과 확인

### 위험 요소 및 대응

| 위험 | 영향 | 대응 방안 |
|------|------|-----------|
| launcher.go 범용 함수가 외부에서 참조됨 | launcher.go 전체 삭제 불가 | 부분 수정으로 전환, 범용 함수만 보존 |
| session_end.go의 GLM 정리 로직 제거 시 다른 정리 로직 훼손 | 세션 종료 시 리소스 누수 | Handle() 함수 내 호출 순서를 신중하게 수정 |
| 미발견 GLM 참조가 빌드 실패 유발 | 빌드 불가 | `grep -rn "GLM\|glm\|clearTmux\|cleanupGLM"` 전체 검색으로 사전 확인 |
| SPEC-REFACTOR-001 완료 전 실행 | deps.go 상태 불일치 | 반드시 SPEC-REFACTOR-001 완료 후 실행 |

## 4. 범위 밖 (Out of Scope)

- `internal/rank/`, `internal/ralph/` 패키지 삭제 (SPEC-REFACTOR-001에서 처리 완료)
- `internal/tmux/` 패키지 삭제 (다른 사용처 확인 필요)
- `.claude/`, `.moai/`, `CLAUDE.md` 수정 (moai-adk 영역)
- 커버리지 테스트 파일 삭제 (SPEC-REFACTOR-001에서 처리 완료)

## 5. 추적성

| 태그 | 참조 |
|------|------|
| SPEC-REFACTOR-002 | `spec.md` |
| TODO-3 | `docs/TODO.md` 항목 3 |
| REQ-001 ~ REQ-011 | `spec.md` 요구사항 섹션 |
