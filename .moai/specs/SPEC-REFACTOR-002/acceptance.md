---
spec_id: SPEC-REFACTOR-002
type: acceptance
version: "1.0.0"
---

# SPEC-REFACTOR-002 인수 기준: cc/glm/cg 런치 명령어 삭제

## 1. 인수 기준 시나리오

### AC-001: 소스 파일 완전 삭제 (REQ-001)

```gherkin
Scenario: cc/glm/cg 소스 파일이 완전히 삭제됨
  Given ae-adk 프로젝트가 존재한다
  When 리팩토링이 완료된다
  Then internal/cli/cc.go 파일이 존재하지 않는다
  And internal/cli/glm.go 파일이 존재하지 않는다
  And internal/cli/cg.go 파일이 존재하지 않는다
```

### AC-002: 테스트 파일 완전 삭제 (REQ-002)

```gherkin
Scenario: cc/glm/cg 관련 테스트 파일이 완전히 삭제됨
  Given ae-adk 프로젝트가 존재한다
  When 리팩토링이 완료된다
  Then internal/cli/cc_test.go 파일이 존재하지 않는다
  And internal/cli/glm_test.go 파일이 존재하지 않는다
  And internal/cli/glm_compat_test.go 파일이 존재하지 않는다
  And internal/cli/glm_model_override_test.go 파일이 존재하지 않는다
  And internal/cli/glm_new_test.go 파일이 존재하지 않는다
  And internal/cli/glm_team_test.go 파일이 존재하지 않는다
```

### AC-003: launcher.go 정리 (REQ-003)

```gherkin
Scenario: launcher.go에서 모드별 런치 함수가 제거됨
  Given ae-adk 프로젝트가 존재한다
  When 리팩토링이 완료된다
  Then launcher.go에 applyCCMode 함수가 존재하지 않는다
  And launcher.go에 applyGLMMode 함수가 존재하지 않는다
  And launcher.go에 applyCGMode 함수가 존재하지 않는다
  And launcher.go에 unifiedLaunch 함수가 존재하지 않는다
  And launcher.go에 unifiedLaunchDefault 함수가 존재하지 않는다
  And launcher.go에 resolveMode 함수가 존재하지 않는다
```

```gherkin
Scenario: launcher.go 전체 삭제 (분석 결과에 따라)
  Given launcher.go 내 범용 함수가 외부에서 참조되지 않음이 확인되었다
  When 리팩토링이 완료된다
  Then internal/cli/launcher.go 파일이 존재하지 않는다
  And internal/cli/launcher_test.go 파일이 존재하지 않는다
```

### AC-003-B: launcher_test.go 정리 (REQ-004)

```gherkin
Scenario: launcher_test.go가 launcher.go와 동일하게 정리됨
  Given ae-adk 프로젝트가 존재한다
  When 리팩토링이 완료된다
  Then launcher_test.go에 삭제된 함수의 테스트가 존재하지 않는다
  And launcher.go가 전체 삭제된 경우 launcher_test.go도 삭제되어야 한다
  And `go test ./internal/cli/...`가 통과해야 한다
```

### AC-004: root.go 명령어 등록 제거 (REQ-005)

```gherkin
Scenario: root.go에서 cc/glm/cg 명령어 참조가 제거됨
  Given ae-adk 프로젝트가 존재한다
  When 리팩토링이 완료된다
  Then root.go의 rootCmd.Long 텍스트에 "ae cc" 문자열이 포함되지 않는다
  And root.go의 rootCmd.Long 텍스트에 "ae glm" 문자열이 포함되지 않는다
  And root.go의 rootCmd.Long 텍스트에 "ae cg" 문자열이 포함되지 않는다
```

```gherkin
Scenario: launch 커맨드 그룹 제거 (해당되는 경우)
  Given "launch" 그룹에 cc/glm/cg 외 다른 명령어가 등록되어 있지 않다
  When 리팩토링이 완료된다
  Then root.go에 "launch" 그룹 등록이 존재하지 않는다
```

### AC-005: session_end.go GLM 정리 로직 제거 (REQ-006)

```gherkin
Scenario: session_end.go에서 GLM 환경 정리 코드가 제거됨
  Given ae-adk 프로젝트가 존재한다
  When 리팩토링이 완료된다
  Then session_end.go에 clearTmuxSessionEnv 함수가 존재하지 않는다
  And session_end.go에 cleanupGLMSettingsLocal 함수가 존재하지 않는다
  And session_end.go에 glmEnvVarsToClean 변수가 존재하지 않는다
  And session_end.go의 Handle 함수에서 GLM 관련 호출이 존재하지 않는다
```

### AC-006: deps.go GLM 의존성 제거 (REQ-007)

```gherkin
Scenario: deps.go에서 GLM 전용 의존성이 제거됨
  Given ae-adk 프로젝트가 존재한다
  When 리팩토링이 완료된다
  Then deps.go에 GLM 전용 import가 존재하지 않는다
  And deps.go에 GLM 전용 코드가 존재하지 않는다
  And rank/ralph 관련 코드는 SPEC-REFACTOR-001에서 이미 제거된 상태이다
```

### AC-007: 빌드 무결성 (REQ-008)

```gherkin
Scenario: 프로젝트가 정상 빌드된다
  Given 리팩토링이 완료되었다
  When "go build ./..." 명령을 실행한다
  Then 빌드가 성공한다

Scenario: 정적 분석 통과
  Given 리팩토링이 완료되었다
  When "go vet ./..." 명령을 실행한다
  Then 경고 메시지가 없다
```

### AC-008: 기존 테스트 통과 (REQ-009)

```gherkin
Scenario: 기존 테스트가 전부 통과한다
  Given 리팩토링이 완료되었다
  When "go test ./..." 명령을 실행한다
  Then 남아 있는 모든 테스트가 통과한다
```

### AC-009: moai-adk 영역 불변 (REQ-010)

```gherkin
Scenario: moai-adk 영역 파일이 수정되지 않음
  Given 리팩토링이 완료되었다
  When git diff HEAD --name-only로 변경 파일 목록을 확인한다
  Then .claude/ 디렉토리 내 파일이 변경 목록에 포함되지 않는다
  And .moai/ 디렉토리 내 설정 파일이 변경 목록에 포함되지 않는다
  And CLAUDE.md 파일이 변경 목록에 포함되지 않는다
```

### AC-010: 기존 명령어 정상 동작 (암시적 요구사항)

```gherkin
Scenario: ae 기본 명령어가 정상 동작한다
  Given 리팩토링이 완료되었다
  When "ae --help" 명령을 실행한다
  Then 도움말이 정상 출력된다
  And cc, glm, cg 명령어가 도움말에 표시되지 않는다
```

## 2. 품질 게이트

### 필수 통과 항목

| 항목 | 검증 명령어 | 기대 결과 |
|------|-------------|-----------|
| Go 빌드 | `go build ./...` | 성공 (exit 0) |
| Go vet | `go vet ./...` | 경고 없음 |
| 전체 테스트 | `go test ./...` | 전체 PASS |
| 파일 삭제 확인 | `ls internal/cli/cc.go internal/cli/glm.go internal/cli/cg.go 2>&1` | "No such file" 출력 |
| GLM 참조 잔존 확인 | `grep -rn "applyCCMode\|applyGLMMode\|applyCGMode\|unifiedLaunch" internal/` | 결과 없음 |
| moai 영역 불변 | `git diff HEAD --name-only \| grep -E "^(\.claude/\|\.moai/config\|CLAUDE\.md)"` | 결과 없음 |

## 3. 완료의 정의 (Definition of Done)

- [ ] 9개 파일 삭제 완료 (소스 3 + 테스트 6)
- [ ] 5개 파일 수정 완료 (launcher.go, launcher_test.go, root.go, session_end.go, deps.go)
- [ ] `go build ./...` 성공
- [ ] `go vet ./...` 경고 없음
- [ ] `go test ./...` 전체 PASS
- [ ] 삭제된 함수/변수 참조가 프로젝트 내 잔존하지 않음
- [ ] moai-adk 영역 파일 미수정
- [ ] `ae --help`에서 cc/glm/cg 미표시

## 4. 추적성

| 수락 기준 | 요구사항 | 마일스톤 |
|-----------|----------|---------|
| AC-001 | REQ-001 | Milestone 2 |
| AC-002 | REQ-002 | Milestone 2 |
| AC-003 | REQ-003 | Milestone 3 |
| AC-003-B | REQ-004 | Milestone 3 |
| AC-004 | REQ-005 | Milestone 3 |
| AC-005 | REQ-006 | Milestone 3 |
| AC-006 | REQ-007 | Milestone 3 |
| AC-007 | REQ-008 | Milestone 4 |
| AC-008 | REQ-009 | Milestone 4 |
| AC-009 | REQ-010 | Milestone 4 |
| AC-010 | - | Milestone 4 |
