# SPEC-SECURITY-BYPASS-001 제안서: disableBypassPermissionsMode 정책 강제

> Status: Proposal (draft 없음)
> Trigger: SPEC-UPDATE-004 expert-security F-1 finding (HIGH)
> Priority: Medium
> Estimated Scope: ~10~15 REQ (settings 파서, Agent spawn 검증, 회귀 테스트)
> Dependencies: SPEC-UPDATE-005 머지 후 권장 (settings.json 영향 영역 안정 후)

---

## 1. 배경 (Why now?)

### 1.1 발생 경위

SPEC-UPDATE-004 REQ-18은 `internal/template/templates/.claude/settings.json.tmpl`에 `"disableBypassPermissionsMode": false` 필드를 추가했다. 머지 시점의 expert-security review(F-1, HIGH)에서 다음이 발견됐다.

> 필드는 사용자 프로젝트의 `.claude/settings.json`에 렌더되지만, ae-adk Go 코드 어디에서도 이를 **읽거나 강제하지 않는다**. `grep -rn "disableBypassPermissionsMode" internal/`은 템플릿 파일만 반환한다. 파서, 검증자, 정책 게이트가 부재.

이 필드는 Claude Code v2.1.111+ 하니스가 honor한다는 가정으로 설계됐지만 (`worktree-integration.md` 호환성 표 참조), ae-adk는:

1. 사용자가 v2.1.111 미만의 Claude Code를 사용할 수 있고
2. 향후 ae-adk가 자체 Agent spawn 로직을 가질 수 있고
3. 다층 방어(Defense in Depth) 원칙상 단일 컨트롤에 의존하면 안 된다

는 이유로 **ae-adk 자체에서도 이 정책을 honor**해야 한다.

### 1.2 위협 모델

| 시나리오 | 영향 |
|---|---|
| 사용자가 보안 환경에서 `disableBypassPermissionsMode: true`로 설정 | ae-adk subagent가 prompt에 `mode: "bypassPermissions"` 요청 시 **무시되고 통과**됨 (현재) |
| 악성 SPEC 또는 prompt injection이 subagent에 bypassPermissions를 강제 | 상위 정책 무시되어 destructive 작업 실행 가능 |
| Claude Code v2.1.110 이하 사용자 | 하니스 차원 보호 부재, ae-adk가 유일한 방어선 |

### 1.3 OWASP / CWE 매핑

- **CWE-732**: Incorrect Permission Assignment for Critical Resource — 필드는 정의했으나 강제하지 않음
- **OWASP A01:2021** Broken Access Control — 정책이 코드에서 분리됨
- **A05:2021** Security Misconfiguration — 사용자가 보안 설정을 켜도 효과가 없는 footgun

---

## 2. 필요성 (Rationale)

### 2.1 정책 필드의 목적 회복

`disableBypassPermissionsMode`라는 이름은 사용자에게 "이 옵션을 켜면 bypass mode가 비활성화된다"는 명확한 약속을 한다. 현재 구현은 이 약속을 어긴다 — 사용자 기대와 실제 동작 사이의 갭은 보안 결함으로 분류된다.

### 2.2 다층 방어

Claude Code 하니스에만 의존하면:
- 하니스 버전 다운그레이드 시 보호 상실
- ae-adk가 자체 Agent API(예: `internal/agent/spawn`)를 만들면 우회 경로 발생
- 하니스 버그/우회 발견 시 즉각 영향

ae-adk가 자체적으로 강제하면 위 시나리오에서도 안전하다.

### 2.3 SPEC-PIPELINE-002 cascade prevention 정렬

SPEC-PIPELINE-002는 auto pipeline의 cascade를 방지하기 위해 **구조적 안전장치**(Phase 4 제거, merge capability 부재, AskUserQuestion merge gate)를 도입했다. 이 철학과 일관되게, 권한 우회 방지도 구조적으로 강제해야 한다 — 단순 설정 파일 필드만으로는 부족.

---

## 3. 목표 (Goals)

### 3.1 기능 목표

1. ae-adk가 사용자 프로젝트의 `.claude/settings.json`에서 `disableBypassPermissionsMode` 필드를 읽는다
2. 값이 `true`일 때 ae-adk subagent spawn 시 `mode: "bypassPermissions"` 요청을 **거부**(또는 `acceptEdits`로 자동 다운그레이드)한다
3. 거부/다운그레이드 발생 시 audit log에 기록한다 (operational visibility)
4. 회귀 테스트로 정책 강제 동작을 보장한다

### 3.2 비기능 목표

- **명시적 fallback**: 필드 부재(legacy settings.json) 시 기본값을 안전 측으로 결정하되, 기존 사용자에게 무언의 동작 변화를 강제하지 않는다 (마이그레이션 가이드 제공)
- **로깅**: 거부/다운그레이드 이벤트는 `~/.ae/logs/security-events.log`에 timestamp + agent name + 원래 mode + 결정 mode 기록
- **테스트 커버리지**: 정책 분기 100%

### 3.3 비목표

- bypassPermissions mode 자체를 ae-adk에서 제거 (정책 필드가 false인 사용자에게는 영향 없도록)
- Claude Code 하니스 동작 복제 (하니스가 또한 honor하므로 중복 강제는 OK)
- 다른 settings.json 필드(예: allowedTools, statusLine)에 대한 정책 강제 (out of scope)

---

## 4. 방향 (Direction)

### 4.1 구현 단계

**Phase 0: 영향 분석**
- `grep -rn "bypassPermissions\|mode:.*bypass" internal/` — 현재 어떤 spawn 경로에서 이 mode를 사용 중인지 확인
- ae-adk가 직접 Agent를 spawn하는 코드 경로 식별 (현재는 Claude Code가 spawn하므로 ae 코드는 prompt만 작성)
- "강제"의 실제 enforcement point 결정

**Phase 1: settings 파서**
- `internal/config/security.go` 신규 — `LoadSecurityPolicy(projectRoot string) (SecurityPolicy, error)` 구현
- `SecurityPolicy` 구조체에 `DisableBypassPermissionsMode bool` 필드
- 사용자 프로젝트 `.claude/settings.json` 우선, 없으면 `~/.claude/settings.json`, 둘 다 없으면 기본값(false = 기존 호환)

**Phase 2: enforcement point**
- 후보 1: ae가 prompt를 생성할 때 `mode: "bypassPermissions"` 요청 패턴이 있는 prompt를 거부 (정적 분석)
- 후보 2: ae가 직접 Agent를 spawn하는 future API에서 mode 검증
- 후보 3: hook handler (`PreToolUse`?)에서 spawn 요청 가로채기

`Phase 0`의 영향 분석 결과에 따라 후보 1 + 후보 3 조합이 유력 (현재 ae는 직접 spawn 안 하므로)

**Phase 3: 로깅 + audit**
- `internal/foundation/security_audit.go` 신규 — `LogPolicyDecision(event PolicyEvent) error`
- 로그 파일: `~/.ae/logs/security-events.log` (JSON-line 형식)
- 회전 정책: 30일 retention (mx.yaml 패턴 참조)

**Phase 4: 회귀 테스트**
- `internal/config/security_test.go` — 정책 로딩 케이스
- 통합 테스트: 가짜 settings.json + 가짜 spawn 요청 → 정책 강제 검증
- 회귀: 사용자가 false로 설정한 경우 기존 동작과 동일

**Phase 5: 문서**
- `internal/template/templates/.claude/rules/ae/core/security-policy.md` 신규 — 정책 동작 명세
- README의 보안 섹션에 `disableBypassPermissionsMode` 의미·기본값·기록 위치 명시

### 4.2 의사결정 포인트

| 결정 | 옵션 | 권장 |
|---|---|---|
| 기본값 | (A) `false` (기존 호환), (B) `true` (안전 우선) | (A) — 기존 사용자에게 무언의 변경 회피, 마이그레이션 가이드는 별도 |
| 거부 시 동작 | (a) 에러 반환, (b) `acceptEdits`로 자동 다운그레이드 + 경고 | (b) — 워크플로우 중단 회피, 사용자에게 명확한 경고 노출 |
| Enforcement point | (i) prompt 정적 분석, (ii) hook 차단, (iii) 양자 | (iii) — Defense in Depth |
| Audit log 위치 | (X) `~/.ae/logs/`, (Y) 프로젝트 `.ae/logs/` | (X) — 사용자 프로젝트 외부 보존 (조작 회피) |

### 4.3 위험 완화

| 위험 | 완화 |
|---|---|
| **R-1**: 정책 강제로 정상 워크플로우 중단 | 다운그레이드 + 경고 패턴 (에러 대신), 기본값 false 유지 |
| **R-2**: 정책 파서 자체에 보안 결함 | TOCTOU 방지 (Stat 후 Open이 아닌 단일 OpenFile), JSON parse limit, error에 정보 누출 없음 |
| **R-3**: hook 차단 우회 (사용자가 hook 비활성화) | 다층 방어 — prompt 정적 분석은 hook 비활성화에도 작동 |
| **R-4**: audit log 변조 | append-only 모드, 권한 0o600, 파일 무결성은 future SPEC |
| **R-5**: Claude Code 하니스 v2.1.111+에서 이미 honor → 이중 강제로 사용자 혼란 | 문서에 "ae-adk와 Claude Code 하니스 모두 honor"명시, 동일한 결정 출력 보장 |

---

## 5. 영향 파일 (예상)

```
internal/config/security.go                           (신규, SecurityPolicy 파서)
internal/config/security_test.go                      (신규)
internal/foundation/security_audit.go                 (신규, audit 로깅)
internal/foundation/security_audit_test.go            (신규)
internal/agent/spawn.go (or equivalent)               (정책 검증 enforcement point)
internal/agent/spawn_test.go                          (회귀 테스트)
internal/hook/pre_tool_use.go (또는 신규 hook)         (정책 차단)
internal/template/templates/.claude/rules/ae/core/security-policy.md  (신규 룰)
internal/template/templates/.claude/settings.json.tmpl                (필드 설명 주석 강화)
README.md or docs/                                    (보안 섹션 추가)
~/.ae/logs/security-events.log                        (런타임 산출물, gitignore)
```

---

## 6. 의존성 / 트리거

- **선행 머지**: SPEC-UPDATE-005 권장 (settings.json 영향 영역 안정 후) — UPDATE-005가 settings.json을 변경하면 본 SPEC의 파서가 그 변경을 반영해야 함
- **선행 정보**: ae-adk가 향후 자체 Agent spawn API를 가질 계획이 있는지 확인 (있다면 enforcement point 설계 변경)
- **트리거**: 보안 감사 또는 외부 사용자가 본 결함을 보고할 경우 우선순위 상향

---

## 7. 우선순위 산정

| 기준 | 점수 (1-5, 5가 높음) | 사유 |
|---|---|---|
| 위협 발생 가능성 | 2 | 현재는 명백한 악성 사용 사례 보고 없음, 내부 도구 |
| 위협 발생 시 임팩트 | 4 | bypassPermissions로 destructive 작업 실행 가능 |
| 사용자 기대 갭 | 5 | 필드 이름이 명확한 약속을 함, 현재 구현은 어김 |
| 구현 비용 | 2 (낮음) | ~10~15 REQ, 테스트 포함 1~2 sprint |
| 차단성 | 1 | 현재 워크플로우 차단 없음, 미래 결함 회피 성격 |

**가중 평균**: 우선순위 **Medium**. SPEC-UPDATE-005 머지 후 SPEC-LSP-CORE-002와 병행 가능.

---

## 8. 다음 단계 (When to start)

- **착수 전 체크**: ae-adk가 직접 Agent spawn을 하는 코드 경로가 신설될 예정인지 manager-strategy 확인
- **초안 작성**: `/ae plan SPEC-SECURITY-BYPASS-001 "disableBypassPermissionsMode 정책 강제"`
- **작성자**: expert-security (위협 모델, OWASP 매핑) + expert-backend (Go 구현) 협업

---

## 9. 참고 자료

- expert-security F-1 finding (SPEC-UPDATE-004 머지 시 검토 보고서)
- Claude Code v2.1.111 릴리즈 노트 (`disableBypassPermissionsMode` 하니스 동작)
- `.claude/rules/moai/workflow/worktree-integration.md` (Minimum Version Requirements 표)
- OWASP Top 10 2021 — A01: Broken Access Control, A05: Security Misconfiguration
- CWE-732 (Incorrect Permission Assignment for Critical Resource)
- SPEC-PIPELINE-002 cascade prevention 패턴 (구조적 안전장치 철학 참조)
