# SPEC-UPDATE-004 / SPEC-UPDATE-005 비판적 리뷰 보고서

> Reviewer: Claude (Opus 4.7, 1M context) · MoAI Orchestrator
> Date: 2026-04-24
> Mode: ultrathink (multi-perspective critical review)
> Scope: `.moai/specs/SPEC-UPDATE-004/`, `.moai/specs/SPEC-UPDATE-005/`
> Review philosophy: 의심을 기본값으로 두고, 결함과 누락을 적극적으로 찾는 skeptical 평가 (moai evaluator-active 패턴 차용)

---

## 0. 요약 (Executive Summary)

본 리뷰는 SPEC-UPDATE-004(v2.10.2 ~ v2.13.2 품질·성능)와 SPEC-UPDATE-005(v2.13.x Design+DB 재편) 두 문서를 다음 9개 관점에서 검증했다.

1. 요구사항 품질 (EARS 형식, 명확성, 원자성)
2. 내부 일관성 (REQ 간 교차 참조)
3. 리스크 완화 전략 적절성
4. AC 검증 가능성 (자동화 스크립트 실행 가능성)
5. 가정(Assumption) 타당성
6. 제외 항목(Exclusion) 적절성
7. SPEC 간 의존성·호환성
8. 일반화의 오류 체크
9. 외부 검증 필요 항목 (업스트림 API, 모델 ID, 환경 변수)

### 0.1 등급 판정

| 항목 | SPEC-UPDATE-004 | SPEC-UPDATE-005 |
|---|---|---|
| 문서 구조 완전성 | A (spec/plan/acceptance 3종 완비) | A |
| EARS 형식 준수 | B+ (28 REQ 모두 준수, 일부 추상도 과다) | B+ (37 REQ 모두 준수, 일부 업스트림 참조 의존) |
| AC 자동화 커버리지 | B (28 AC 중 약 18건 완전 자동화) | A- (37 AC 중 약 30건 자동화) |
| 리스크 완화 적절성 | B | B- (R-01 완화책이 AC-35와 중복, 다른 리스크 누락) |
| 업스트림 일치 (팩트체크) | **C** (REQ-01 경로 불일치, v2.14.0 스코프 갭) | B (v2.13.x 중심은 정확, v2.14.0 반영 필요) |
| **종합** | **B (머지 전 Critical 2건·High 3건 수정 필요)** | **B+ (Critical 0건·High 4건 권고)** |

### 0.2 머지 차단 요건 (Blocker)

다음 항목은 **실행 전 반드시 해소해야 하는 블로커**다.

| # | SPEC | 이슈 | 심각도 |
|---|---|---|---|
| B-1 | 004 | REQ-01 파일 경로 `.ae/config/sections/system.yaml`이 실존하지 않음. 실제 파일은 `system.yaml.tmpl`. AC-01 검증 스크립트 전부 실패 예상 | Critical |
| B-2 | 004 | `v2.14.0` Utility Hardening (2026-04-24 릴리즈)이 SPEC-004 스코프에서 누락. MX validator method receiver detection, tree-sitter 16-language, ast-grep 5-language rule seeding, LSP hygiene 등 중요 변경 미반영 | High |
| B-3 | 004 | REQ-04 "하드코딩 상수 12건"의 구체적 대상 누락. 업스트림 커밋 해시(`d29771e`) 참조만 있음 → 구현 시 누락 위험 | High |
| B-4 | 005 | `ae-agency-frontend-patterns`의 `ae-ref-frontend-patterns` 리네이밍(REQ-09) — 기존 `ae-ref-*` 카테고리 정의와 의미 충돌 예상. A-06의 "의미 확장 허용" 선언만으로는 카테고리 네임스페이스 오염 방지 불가 | High |
| B-5 | 005 | REQ-08 "ae-agency-evaluation-criteria → ae-workflow-gan-loop 흡수" 전에 **업스트림 moai-workflow-gan-loop 본문 확인 단계 누락**. 기존 개념과 중복·상충 검토 없이 병합 시 의미 손실 위험 | High |

---

## 1. 사전 검증 (팩트체크)

리뷰 시작 전 SPEC 본문이 현재 코드베이스와 일치하는지 기초 팩트를 확인했다.

### 1.1 파일 경로 실존 여부

| SPEC | 명시 경로 | 실존 여부 | 비고 |
|---|---|---|---|
| 004 REQ-01 | `.ae/config/sections/system.yaml` | **부재** | 실제 파일은 `system.yaml.tmpl` (텍스트 템플릿). Go embed 대상 |
| 004 REQ-10 | `.ae/config/sections/llm.yaml` | 존재 | ✓ |
| 004 REQ-19 | `.ae/config/sections/lsp.yaml` | **부재** (sections/ 하위에 lsp.yaml 없음) | 추가 확인 필요 |
| 004 REQ-21 | `.ae/config/sections/workflow.yaml` | 존재 | ✓ |
| 005 REQ-01 (대상) | 6개 agency 에이전트 파일 | 존재 | ✓ 모두 실존 |
| 005 REQ-05~09 (대상) | 5개 `ae-agency-*` 스킬 | 존재 | ✓ 5개 모두 실존 |

**발견**: SPEC-004 REQ-01, REQ-19의 파일 경로가 현행 레이아웃과 불일치. AC-01·AC-19의 grep 검증 구문이 현재 상태로는 모두 실패한다. SPEC-004 실행 전 **경로 현행화 필수**.

### 1.2 moai-adk 릴리즈 현황 (2026-04-24 기준)

```
v2.14.0 | 2026-04-24T02:15:03Z  ← NEW (SPEC-004/005 작성 이후)
v2.13.2 | 2026-04-23T04:26:46Z  ← SPEC-004 스코프 끝점
v2.13.1 | 2026-04-22T23:21:25Z
v2.13.0 | 2026-04-22T16:38:21Z
v2.12.0 | 2026-04-17T05:18:43Z
v2.10.2 | 2026-04-11T05:14:16Z  ← SPEC-004 스코프 시작점
v2.10.1 | 2026-04-09T13:12:51Z  ← SPEC-003 스코프
```

**v2.14.0 주요 변경 (Breaking None, Detection Improvements Yes)**:
- MX validator method receiver detection (`func (r *Receiver) ExportedMethod()` 포함)
- Word-boundary fan-in counting (substring false positive 제거)
- Bounded goroutine pools (`runtime.NumCPU()*2` semaphore)
- tree-sitter 16-language cyclomatic complexity (5 언어 query seed + 11 언어 scaffold)
- ast-grep 5-language rule seeding (Ruby/PHP/Elixir/C#/Kotlin) + suppression policy (`// ast-grep-ignore` + `// @MX:REASON` pairing)
- LSP stderr drain goroutine (128 KiB deadlock 방지)
- LSP Manager `singleflight.Group` (exactly-once factory)
- `transition_mode` flag in `mx.yaml` (v2.14 한정 grace)
- `@MX:REASON` pairing enforcement (ANCHOR: P1, WARN: P2)

**영향**: SPEC-004/005 실행 중 v2.14.0 변경이 별도로 반영되지 않으면, 머지 직후 또 다른 SPEC-UPDATE-006 사이클이 즉시 필요해진다. 최소한 SPEC-004 스코프 재정의(≤v2.14.0 품질 변경 흡수) 또는 SPEC-UPDATE-006 사전 초안화가 요구된다.

### 1.3 CLAUDE.md 자 수

| 파일 | 자 수 | 40,000자 제한 대비 |
|---|---|---|
| `internal/template/templates/CLAUDE.md` | 19,896 bytes | 49.7% (여유 ~20,000 bytes) |

**판정**: SPEC-005 R-10("CLAUDE.md 40,000자 제한 초과")은 과대평가된 리스크. design/db 섹션 추가에 여유 충분. 완화책(Skill 참조로 축약)은 굳이 강제하지 않아도 됨.

### 1.4 grep 패턴 취약점 사전 분석

SPEC-005 AC-35 Scenario 1~6의 grep 패턴을 정적 분석한 결과:

| Scenario | 패턴 | 취약점 |
|---|---|---|
| 1 | `grep -rn "moai" ... \| grep -v "modu-ai/moai-adk"` | `modu-ai/moai-adk.git`, URL, CHANGELOG 인용 등에서 false positive 발생 가능. 또한 **일반적 영어 단어 "moai"**(이스터섬 모아이)가 스킬 예시 본문에 등장할 확률은 낮지만 제로는 아님 |
| 2 | `grep -rn "/moai " --include="*.md" --include="*.tmpl"` | 코드 블록 안의 "정보 제공용" `/moai` 언급 (예: "업스트림에서는 `/moai plan`이라고 불렀지만...")이 걸림 |
| 3 | `grep -rn "\.moai/"` | 점 이스케이프 OK. 단, `..moai/` (오타) 같은 케이스 false negative |
| 4 | `grep -rn "<moai>"` | ✓ (엄격) |
| 5 | `grep -rn 'Skill("moai-' ...' | grep -v "moai/workflows\|moai/team"` | `Skill("moai-` 이외에 `Skill('moai-` (single quote) 변형 누락 |
| 6 | `grep -rn "^name: moai-" --include="SKILL.md"` | `name: "moai-xxx"` (double-quoted YAML) 누락 |

**권고**: `internal/template/namespace_integrity_test.go`에서 grep 대신 **구조화된 Go 정규식**으로 교체. 테스트 함수 내부에서 YAML 파싱으로 name 필드를 추출하고 워드 바운더리까지 검사하는 방식이 robust.

---

## 2. SPEC-UPDATE-004 상세 리뷰

### 2.1 요구사항 품질 (EARS 형식·명확성)

**장점**:
- 28개 REQ 모두 WHEN/THEN 또는 WHILE/IF/THEN 구조 준수
- REQ-07·REQ-24는 7개 에이전트를 구체적으로 나열 → 구현 누락 방지
- REQ-26은 [HARD] 표시로 검증 강제

**결함**:

| # | REQ | 이슈 | 권고 |
|---|---|---|---|
| Q-1 | REQ-02 | "업스트림 v2.10.2 스펙과 정렬" - 구체 필드 구조 없음 | `AstGrepGate` 구조체 필드 3~5개를 spec에 명시 (예: `enabled bool`, `rules []string`, `severity string`) |
| Q-2 | REQ-04 | "하드코딩 상수 12건" - 목록 없음. `d29771e` 커밋 해시만 참조 | 12개 상수명을 spec의 부록에 명시 (rebase로 해시 소실 대비) |
| Q-3 | REQ-15 | "프로젝트 `.mcp.json`과 글로벌 `~/.claude/.mcp.json`의 스코프 충돌" - "충돌" 정의 없음 | 충돌 정의: 동일 `server name` 존재 여부로 한정. 또는 `mcpServers.*.command` 불일치 |
| Q-4 | REQ-17 | "Windows + .env 존재 시 주입" - `.env` 파일의 어떤 값을 어떻게 주입? | `.env` 원본 내용을 파싱하지 않고 **파일 절대 경로만** `settings.local.json` `env.CLAUDE_ENV_FILE` 필드에 기록한다고 명시 (업스트림 실제 동작 확인 필요) |
| Q-5 | REQ-19 | "16개 언어 file_extensions 필드" - 16개 언어 목록과 각 확장자 미제공 | 16개 언어 × 확장자 매핑 표를 spec 부록으로 추가 (Go: `.go`, Python: `.py,.pyi`, TypeScript: `.ts,.tsx` 등) |
| Q-6 | REQ-22 | "4개 언어 번역 리소스" - 번역 품질 게이트 없음 | 최소 번역자 검토 체크박스 또는 "user English fallback on missing translation" 명시 |
| Q-7 | REQ-27 | "41개 스킬" - 구체적 스킬 목록 없음. 카테고리 표현만 | 41개 스킬 리스트를 Appendix로 첨부. Phase 6.1 Tasks의 "N개"에 실수로 계산 오류 내재 가능 |

### 2.2 내부 일관성

**양호한 교차 참조**:
- REQ-07(`agentEffortMap`)와 REQ-24(7개 에이전트 effort 필드 복구): 값 일치 (`manager-spec=xhigh`, `manager-strategy=xhigh` 등) ✓
- REQ-08(`ApplyEffortPolicy` 구현)와 REQ-07(map 정의): 구현→호출 관계 명확 ✓

**일관성 결함**:

| # | 설명 | 권고 |
|---|---|---|
| C-1 | REQ-06는 `ModelIDOpus47 = "claude-opus-4-7"` 상수 정의, REQ-10은 `llm.yaml`의 `claude_models.high = "claude-opus-4-7"` 값 설정 → 중복. **어느 쪽이 Single Source of Truth인가?** | 상수를 SSOT로 하고 템플릿에서는 상수를 참조하도록 플랜에 명시 |
| C-2 | A-08 "template_version 현재 2.9.1 → 실제 v2.10.1 → 본 SPEC 후 v2.13.2" — SPEC-003이 PR #25로 머지되었는데 왜 버전이 2.9.1로 멈춰있는가? 원인 진단 없이 "보정"만 수행 | Phase 1.1 앞에 "원인 조사" 하위 task 추가. 재발 방지 (템플릿 내 system.yaml이 아니라 `system.yaml.tmpl`이었기 때문일 가능성 높음) |
| C-3 | AC-26 Scenario 3~5는 `grep -c "HUMAN GATE" ... ≥ 2` 검증. 그러나 HUMAN GATE 블록이 **섹션 헤더** 형태인지 **단순 언급** 형태인지 구분 없음 | `^##.*HUMAN GATE` (H2 헤더) 또는 `^###.*HUMAN GATE` (H3 헤더) 패턴으로 격상 |
| C-4 | REQ-13 "fixed `thinking.budget_tokens` 지시어 제거" vs AC-13 `grep -i "budget_tokens" → 미매칭 또는 Adaptive 맥락만` — "Adaptive 맥락만"의 자동 판별 불가 | grep으로는 "전면 제거" 또는 "특정 허용 패턴 제외" 방식으로 단순화 권고 |

### 2.3 리스크 완화 평가

| 리스크 | 등급 | 완화 평가 | 보강 권고 |
|---|---|---|---|
| R-01 HUMAN GATE 침투 | High | grep 회귀 테스트 | **섹션 헤더 패턴**으로 엄격화 (위 C-3). CI 실패 메시지에 SPEC-PIPELINE-002 링크 추가 |
| R-02 ApplyEffortPolicy preserve | Medium | manifest hash + preserve | "사용자 커스텀 effort 값"의 정의 모호. `effort: XHIGH` (대문자), `effort: "high"` (문자열) 등 변형 케이스의 정규화 정책 필요 |
| R-03 llm.yaml 기본값 변경 | Medium | 템플릿 기본값만 변경 | 신규 `ae init` 사용자가 Opus 4.7 비용을 인지하지 못할 위험. **비용 경고 메시지** 추가 |
| R-04 lsp.yaml 스키마 변경 | Medium | compliance 테스트 + fallback | **fallback 제거 시점** 미정. v2 → v3 마이그레이션 기한(예: "SPEC-UPDATE-006까지 제거") 명시 필요 |
| R-05 role_profiles 비용 영향 | Medium | 기본값만 변경 | 사용자에게 비용 변동 안내 부족 — **Phase 5.2 Tasks**에 "CHANGELOG에 기본 모델 변경 고지" 추가 |
| R-06 41개 스킬 agency 오염 | Low | `find -prune` | Phase 6.1에서 제외 목록을 **테스트**로 강제 (`TestAgencySkillsNotTouched`) |
| R-07 version jump | Low | DRY-RUN | `ae update`의 비교 로직이 SemVer semantic으로 n-step diff를 어떻게 처리하는지 사전 검증 필요 |

**누락된 리스크**:

| # | 누락 리스크 | 설명 | 등급 |
|---|---|---|---|
| R-NEW-1 | `powernap` v0.1.4 breaking API | REQ-23에서 go.mod/go.sum만 업데이트. v0.1.3 → v0.1.4 changelog 미검증 | Medium |
| R-NEW-2 | Windows 경로 이스케이핑 | REQ-17 `injectCLAUDEEnvFile` — 공백·유니코드 포함 경로 처리 명시 없음 | Medium |
| R-NEW-3 | settings.json merge 원자성 | REQ-17 `settings.local.json` merge — O_EXCL + os.Rename 패턴 미적용 (005 REQ-27은 적용하는데 004는 누락) | Medium |
| R-NEW-4 | v2.14.0 누락 | 팩트체크 1.2 참조. v2.14.0 이 본 SPEC 머지 직후 다시 별도 SPEC 필요 | High |

### 2.4 AC 검증 가능성

**자동화 분석** (28개 AC):

| 상태 | 개수 | 예시 |
|---|---|---|
| 완전 자동화 (grep/go test 단독) | 18 | AC-01, AC-06, AC-07, AC-13, AC-20, AC-26, AC-28 |
| 부분 자동화 (수동 확인 병용) | 7 | AC-05 (통합 테스트), AC-09 (수동 UI), AC-10 Scenario 2 (전후 diff) |
| 수동 전용 | 3 | AC-04 (상수 참조 정합성), AC-11 (5원칙 의미 검증), AC-17 Windows fixture |

**AC별 구체 결함**:

| AC | 이슈 | 권고 |
|---|---|---|
| AC-01 | `internal/template/templates/.ae/config/sections/system.yaml`이 실존하지 않음 (팩트체크 1.1) | `system.yaml.tmpl` 경로로 수정 |
| AC-04 | "grep -rn `<하드코딩 값>`"에서 `<하드코딩 값>`이 실제 값으로 치환돼야 하는데 spec에 값 없음 | Q-2 권고와 동일 (상수 목록 부록) |
| AC-05 Scenario 2 | `syscall.Exec` 이후 환경변수 검증은 "자식 프로세스에서 접근 가능"으로 기술. 실제 테스트 방법 불명확 | 자식 프로세스 모킹 (테스트용 바이너리 spawn) 또는 **환경변수 설정 직전 state 검증**으로 대체 |
| AC-09 | 수동 UI 확인 다수 — CI 그린 판정에 인간 개입 필요 | `go test`로 검증 가능한 부분(번역 리소스 존재, 선택기 옵션 리스트)과 수동 부분을 분리 명시 |
| AC-22 | `go test -coverprofile` 경로가 shell 문법. `go test -cover`만으로 커버리지 100% 강제는 부정확. `-coverprofile` + `go tool cover -func` 조합 필요 | 명령어 정확성 재검토 |
| AC-26 | (C-3과 동일) `grep -c "HUMAN GATE"` - 코드 블록 내 예시 문자열도 카운트됨 | H2/H3 헤더 패턴 엄격화 |
| AC-27 | Scenario 1 bash for-loop — `ae-agency-*`가 제외되어야 하는데 루프 변수 `dir`가 `grep -v` 직후 들어옴. 셸 인용 취약 | Go 테이블 테스트로 대체 (`[]string{"ae-domain-backend", ...}` 명시) |

### 2.5 가정 타당성

| Assumption | 평가 | 근거/리스크 |
|---|---|---|
| A-01 ApplyEffortPolicy preserve | 약함 | "사용자 커스텀 effort 값" 정의 모호 (R-02 중복) |
| A-02 llm.yaml 기본값 변경 무해 | 보통 | "기존 프로파일 보존" 동의, 단 **신규 프로젝트** 영향은 명시 |
| A-03 lsp.yaml 스키마 변경 + 자동 마이그레이션 | 보통 | "자동 마이그레이션 경고와 함께"가 강한 약속. 실제 구현 체크 필요 |
| A-04 Opus 4.7 fixed budget_tokens → HTTP 400 | **외부 검증 필요** | Anthropic API 공식 문서 또는 업스트림 v2.12.0 커밋 메시지 확인 필요. `ae-workflow-thinking` 재설계 전제이므로 중요 |
| A-05 role_profiles 1:1 매핑 가능 | 보통 | ae-adk `workflow.yaml`의 role_profiles가 업스트림과 동일 구조인지 사전 검증 필요 |
| A-06 Windows injectCLAUDEEnvFile 경로 보존 | 강함 | OS 분기 경로 변경 없음은 일반적 패턴 |
| A-07 41개 스킬에서 ae-agency-* 제외 | 약함 | 제외 5개가 정확히 `ae-agency-client-interview`, `ae-agency-copywriting`, `ae-agency-design-system`, `ae-agency-evaluation-criteria`, `ae-agency-frontend-patterns`인데 spec에 **구체명 누락** (plan.md Phase 6.1에는 부분 언급) |
| A-08 version 2.9.1 → 2.13.2 한 번에 갱신 허용 | **조사 필요** | SPEC-003 종료 시 왜 누락됐는지 근본 원인 미규명 (C-2 중복) |
| A-12 `moai glm` 수정 대응 불필요 | 강함 | SPEC-REFACTOR-002에서 glm 제거 ✓. 단 settings.local.json 공통 로직 영향 분리 확인 필요 |

### 2.6 제외 항목 (Exclusions) 검토

| EX | 평가 | 비고 |
|---|---|---|
| EX-01 ~ EX-07 | 타당 | SPEC-005로 이월 명시 |
| EX-08 `moai glm` 오염 | 타당 | A-12와 일관 |
| EX-09 v2.10.1 이하 재반영 | 타당 | SPEC-003 완료 범위 |
| EX-10 ~ EX-11 user-invocable/Agency 섹션 | 타당 | |
| EX-12 Hextra docs-site | 타당 | ae-adk 범위 밖 |
| **누락** | v2.14.0 대응 | EX-13으로 "v2.14.0 Utility Hardening은 SPEC-UPDATE-006으로 이월" 명시 권고. 명시하지 않으면 스코프 혼선 |

### 2.7 Phase 실행 가능성

Plan.md Phase 1~6의 현실성:

- **Phase 1**: Effort 시스템 신규 파일 2개 생성 — 타당
- **Phase 2**: 7개 하위 작업, Go 코드 다수. **2.3 하드코딩 상수 12건 추출**은 "기존 참조 지점 일괄 치환"이 MultiEdit 1~2회로 끝나지 않을 가능성 (Q-2와 연결)
- **Phase 3**: 규칙/스킬 본문 작업. 업스트림 텍스트 fetch 필수인데 "업스트림에서 가져오기" 지시만 있음 → Phase 1에 **업스트림 fetch 전략** (gh api release tarball? git clone shallow? 특정 파일 raw fetch?) 추가 권고
- **Phase 4**: HUMAN GATE 이식 + 7 에이전트 effort — 가장 위험. Phase 6 검증에서 실패 시 롤백 절차 없음
- **Phase 5**: 설정 파일 업데이트 — 5.3 lsp.yaml 16개 언어 file_extensions 맵 없음 (Q-5)
- **Phase 6**: 41개 스킬 + 회귀 테스트 — 대규모 일괄 편집. **실패 시 부분 롤백 경로** 명시 없음

---

## 3. SPEC-UPDATE-005 상세 리뷰

### 3.1 최우선 제약 (네임스페이스 변환 무결성)

**장점**:
- AC-35를 "최우선 HARD"로 격상, 6개 Scenario 명시
- REQ-37로 CI 게이트 자동화 강제
- 허용 예외 5종 명시

**결함**:

| # | 설명 | 권고 |
|---|---|---|
| N-1 | Scenario 1 `grep -v "modu-ai/moai-adk"` - `modu-ai\nmoai-adk` (줄바꿈 삽입된 CHANGELOG 인용) false positive 처리 불가 | grep 대신 Go `regexp` + 다중 패턴 combiner |
| N-2 | `name: "moai-xxx"` (YAML quoted) 형태 false negative | YAML 파싱으로 추출 (팩트체크 1.4) |
| N-3 | Scenario 4: `<moai>` 태그 외에 `[moai]` Markdown link, `moai:` (word prefix) 누락 | "moai" 단어 전체를 검사하는 word boundary 기반 테스트 추가 |
| N-4 | 허용 예외 `builder-agent`, `builder-skill`, `builder-plugin` — 이들은 `.claude/agents/` 하위에 있으므로 스킬 네임스페이스(`moai-*`)와 무관 | 환경 섹션에서 "이 예외는 어디에 적용되는가"를 명시 |
| N-5 | `.agency.archived/` 생성물이 네임스페이스 변환 검증 대상인지 불명확. 이 디렉토리는 프로젝트 루트이므로 템플릿 스캔 대상 아니지만 **user data로 간주**해야 함 | EX-06과 연결하여 "스캔 제외" 명시 |

### 3.2 네임스페이스 변환 규칙 표 검토

**현행 표** (spec.md 79행 ~):
```
| 원본 | 변환 | 예외 |
|---|---|---|
| `.claude/agents/moai/` | `.claude/agents/ae/` | - |
...
```

**빠진 변환 규칙**:

| 원본 | 변환 필요? | 근거 |
|---|---|---|
| `MOAI_XXX` (env vars) | **검토 필요** | `CLAUDE_CODE_EFFORT_LEVEL` (REQ-05)은 업스트림 표준으로 유지, 그러나 ae 고유 env var는 `AE_XXX` 접두사 고려 |
| `moai-adk`, `moai-adk-go` (프로젝트명) | 변환 (`ae-adk`) | **spec에 명시됨** ✓ |
| `github.com/modu-ai/...` (Go import) | 변환 금지 | **spec에 명시됨** ✓ |
| YAML 주석 `# moai-xxx` | 변환 | spec 누락 |
| JSON 키 `"moai"` (claude-in-chrome 식별자?) | 변환 금지 | spec에 맥락 없음 |
| Shell 스크립트 변수 `$MOAI_XXX` | 검토 필요 | `.claude/hooks/ae/*.sh` 내부 |

### 3.3 요구사항 품질

**양호**:
- REQ-10 ~ REQ-16 이식 7개 스킬 각각 명시
- REQ-21 (brand 3-file), REQ-22 (db 7-file) 파일명 구체적
- REQ-35 [HARD] 표시 적절

**결함**:

| # | REQ | 이슈 | 권고 |
|---|---|---|---|
| RQ-1 | REQ-08 흡수 | 업스트림 `moai-workflow-gan-loop` **현재 본문 미분석**. 기존 5개 개념(Rubric Anchoring 등)과 충돌·중복 검토 없이 병합 계획 | Phase 2.6 완료 후 Phase 3.4 전에 **공식 content diff 단계** 삽입 |
| RQ-2 | REQ-23 `.ae/design/` 스캐폴딩 | SHA-256 user-edit 보존 로직이 **REQ-26 `ae migrate agency`와 중복**. 두 곳이 같은 로직을 공유하는지, 독립 구현인지 불명 | 공통 헬퍼 패키지 `internal/template/integrity/` 제안 |
| RQ-3 | REQ-28 16개 언어 DB 탐지 | 16개 언어 매니페스트 매핑이 spec에 없음. plan.md 6.3에 일부 언어만 기재 (PHP composer.json, C++ CMakeLists.txt, R DESCRIPTION, Flutter pubspec.yaml 누락) | 16개 언어 × 매니페스트 파일 × ORM 후보 매트릭스 appendix 추가 |
| RQ-4 | REQ-29 tripartite 구조 | "Brand Context / Design Brief / Relationship" 3개 섹션. 각 섹션의 **역할 분담과 책임 경계**가 spec에 없음 | 업스트림 `.claude/rules/moai/design/constitution.md` v3.3.0의 핵심 정의 3~5줄씩 spec에 인용 |
| RQ-5 | REQ-32 `user-invocable: false` | Claude Code가 이 frontmatter를 실제로 지원하는지 (A-08) **외부 검증 필요** | 공식 문서 URL 링크 또는 버전 번호 명시 |

### 3.4 내부 일관성

**양호**:
- REQ-01~04 (agency 삭제) + REQ-30 (CLAUDE.md 제거) 교차 참조 ✓
- REQ-15 (이식) + REQ-08 (흡수) 순서 의존 명시 ✓
- REQ-10, REQ-11 + REQ-32 (user-invocable) 크로스 링크 ✓

**결함**:

| # | 설명 |
|---|---|
| CN-1 | REQ-15 이식 + REQ-08 흡수 간 **Phase 2.6 → Phase 3.4 순서 의존**이 Plan.md에는 있으나 spec에 명시 없음. REQ 수준에서도 "REQ-08은 REQ-15 이식 완료를 전제로 한다" 명시 필요 |
| CN-2 | REQ-17 (`/ae design` 이중 경로) + REQ-24 (design.yaml path_selection) — default 값이 `path_b`라는 것은 **plan.md 5.4에만 있음**. spec에도 명시 필요 |
| CN-3 | REQ-21 3-file (brand-voice, visual-identity, target-audience) + REQ-10 (ae-domain-brand-design) — 어느 파일을 어느 스킬이 읽는가? Brand Context 로딩 정책이 분산됨 |
| CN-4 | Plan.md Phase 4 (Agency 제거) Phase 3 (스킬 재편)과 **병렬 가능 주장** — 실제로는 agency 에이전트가 스킬을 호출하는 경우 순서 의존. Phase 2 (이식)이 Phase 4 전에 완료돼야 안전 |

### 3.5 리스크 완화 평가

| 리스크 | 등급 | 완화 평가 | 보강 |
|---|---|---|---|
| R-01 네임스페이스 누락 | Critical | AC-35 + REQ-37 | Section 3.1의 grep 취약점 5건 해소 필요 |
| R-02 gan-loop 개념 손실 | High | content diff 기록 | **Phase 3.4 전 공식 diff 단계 추가** (RQ-1와 중복) |
| R-03 migrate 실패 시 원본 손상 | High | 원본 읽기만 + .archived + --resume | `.migration-state.json` 원자성 보장 명시 필요 (O_EXCL+Rename) |
| R-04 리네이밍 참조 깨짐 | Medium | grep 전수 | 리네이밍 구 이름을 `namespace_integrity_test.go`에도 추가 검증 (현재 Scenario 없음) |
| R-05 사용자 혼란 | Medium | CLAUDE.md + migrate 메시지 | **Deprecation warning을 기존 명령 실행 시점**(예: `ae agency` 호출 시)에도 노출 고려 |
| R-06 debounce race | Medium | O_EXCL+Rename | 10초 debounce 기준의 합리성? 사용자가 1초 간격으로 여러 schema 파일 수정 시 효과적인가? |
| R-07 learnings/evolution 마이그레이션 | Medium | 하위 디렉토리 포함 | **데이터 포맷 호환성** — 업스트림 moai가 어떤 구조를 기대하는지 검증 없이 "파일 이동만" 수행하면 새 구조가 읽지 못할 가능성 |
| R-08 이중 경로 혼선 | Medium | yaml roles 분담 | 자연어 yaml 필드는 LLM 해석 의존. 더 강한 **선언적 yaml** 필요 (예: `impeccable_runs_before: ["brand_design"]`) |
| R-09 Sprint Contract 연동 | Medium | gan-loop SKILL.md에 역할 분담 | evaluator-active 기존 구현 분석 단계 누락 |
| R-10 CLAUDE.md 40K 초과 | Low | 요약 + Skill 참조 | **팩트체크 1.3에 따라 과대평가**. 완화 노력 불필요 |
| R-11 ae-ref 카테고리 혼란 | Low | "A-06 허용" | ae-ref-* 기존 카테고리 정의 문서화가 **사전 단계**로 누락. Phase 1에 "ae-ref-* 카테고리 정의 명시" 추가 권고 |

**누락된 리스크**:

| # | 누락 리스크 | 등급 |
|---|---|---|
| R-NEW-1 | v2.14.0 병행 반영 필요성 (SPEC-004와 동일) | High |
| R-NEW-2 | .agency.archived/ git add 시 민감 데이터 누출 (learnings 내 개인정보) | Medium |
| R-NEW-3 | DB schema sync 훅의 대량 변경 대응 (10,000+ 줄 prisma schema) | Low |
| R-NEW-4 | 한국어 code_comments 기본값과 업스트림 영어 주석 이식 충돌 (A-12 관련) | Low |

### 3.6 AC 검증 가능성

**자동화 분석** (37개 AC):

| 상태 | 개수 | 비고 |
|---|---|---|
| 완전 자동화 | 30 | 파일 존재/grep/go test |
| 부분 자동화 | 5 | AC-27 (schema 파일 편집 시뮬레이션), AC-28 (각 언어 fixture) |
| 수동 전용 | 2 | AC-31 R-10 (40K 자동 체크 가능, 수동으로 분류됨) |

**AC별 구체 결함**:

| AC | 이슈 | 권고 |
|---|---|---|
| AC-08 Scenario 2 | `for concept in ...; do grep -l ...; done` — grep -l은 파일명 출력. 개념 실제 포함 여부는 확인되나 **맥락 유효성** 검증 불가 | 의미 검증은 리뷰어 수동 단계로 분리 명시 |
| AC-17 Scenario 2 | design.md "moai 참조 0건" — 업스트림 이식본이므로 변환 누락 가능성 높음. **집중 QA 영역** | Phase 2에 design.md 별도 grep 게이트 추가 |
| AC-26 Scenario 2 | `--resume` 로직 — state 파일 손상 케이스 미검증 | state 파일 파싱 에러 → 재시작 프롬프트 시나리오 추가 |
| AC-27 | PostToolUse 훅 발동 통합 테스트 방법 불명확 — Go 단위 테스트로 훅 이벤트 주입 | Mock `ToolEvent` 구조체 사용 명시 |
| AC-28 Scenario 1 | 16개 언어 각각 fixture 필요. **fixture 16개 생성 작업** plan.md에 누락 | Phase 6.3 Tasks에 "16개 fixture 생성" 하위 작업 추가 |
| AC-35 | Section 3.1의 grep 취약점 5건 | Go 테스트 기반 구조화된 검증 |

### 3.7 Phase 실행 가능성

- **Phase 1 (인프라)**: ✓ 타당. namespace_integrity_test.go skeleton 우선 작성
- **Phase 2 (이식 7스킬)**: 업스트림 fetch 전략 명시 필요 (SPEC-004 Phase 3 결함과 동일)
- **Phase 3 (재편 5스킬)**: 3.4 흡수 전 업스트림 본문 분석 단계 누락 (RQ-1)
- **Phase 4 (삭제)**: ✓ 명확
- **Phase 5 (신설)**: 5.3 design.md.tmpl 메타데이터 구조 미명시
- **Phase 6 (Go 코드)**: 6.3 16개 언어 fixture 작업 과소평가 (실제로 많은 시간 소요)
- **Phase 7 (라우터/CLAUDE.md)**: ✓ 작업량 현실적
- **Phase 8 (user-invocable)**: ✓ (Phase 2.1/2.2에서 선반영 권고했으므로 실질적으로 검증만)
- **Phase 9 (네임스페이스 검증)**: **핵심 Phase**. grep 취약점(3.1) 해소 후 실행
- **Phase 10 (통합)**: ✓

### 3.8 사용자 확정 결정 재검토

| 결정 | 평가 |
|---|---|
| 결정 1: Agency 6 에이전트 삭제 + 재분배 | 타당. **단 재분배 매핑의 구체 프롬프트 위치** 언급 필요. 예: "evaluator-active SKILL.md에 어떤 GAN loop 언급이 추가되는가" |
| 결정 2: 5 스킬 옵션 C | 타당. 단 RQ-4 (`ae-ref-frontend-patterns` 카테고리) 재검토 여지 |
| 결정 3: `/ae db` 옵션 A (전체 이식) | 타당. 단 ae-adk 사용자 중 DB 프로젝트 비중 확인 없음. **DB 없는 프로젝트에서의 `/ae db init` 동작** 명시 필요 |

---

## 4. SPEC 간 의존성·호환성

### 4.1 머지 순서

SPEC-005 A-01: "SPEC-UPDATE-004 선행 머지 필수"

**검증**:
- SPEC-004 Phase 4 (HUMAN GATE 이식) + SPEC-005 Phase 7 (CLAUDE.md agency 제거) — **파일 충돌 없음** ✓
- SPEC-004 REQ-24 (7 에이전트 effort 필드 복구) + SPEC-005 REQ-01 (6 agency 에이전트 삭제) — **대상 겹침 없음** ✓ (삭제 대상 6개는 effort 필드 복구 대상 7개와 별개)

### 4.2 교차 검증 항목 (AC-Scenario F, SPEC-005)

| 검증 | 평가 |
|---|---|
| 7 에이전트 effort 유지 | ✓ |
| HUMAN GATE plan/run/sync 유지 | ✓ |
| template_version >= 2.13.2 | ✓ |
| **누락**: SPEC-004 R-04 fallback 제거 여부 | 추가 검증 필요 |

### 4.3 Agency 완전 제거 후 연쇄 영향

| 영향 | 완화 여부 |
|---|---|
| `/ae agency` 호출 시 404 | CLAUDE.md migration notice (R-05) |
| agency 에이전트를 참조하던 다른 문서/스킬 | grep 전수 (Phase 4.1 Tasks 3) |
| .moai/upstream/ae-delta.md 갱신 | Phase 9.4에서 처리 ✓ |
| **누락**: agency 사용자가 만든 `.agency/learnings/` 데이터 | `ae migrate agency`로 이전 (REQ-26), 단 R-07 완화책 미완 |

---

## 5. 일반화의 오류 체크리스트

사용자 요청에 따라 일반화 오류를 중점 점검.

| # | 항목 | 오류 의심 | 권고 |
|---|---|---|---|
| G-1 | "16개 언어" (SPEC-004 REQ-19·19, SPEC-005 REQ-28) | 실제 언어 목록이 여러 곳에 분산. SPEC-004의 16과 SPEC-005의 16이 동일한지 검증 없음 | **언어 리스트를 `.ae/config/sections/languages.yaml` 같은 공통 SSOT로 통합** 권고 |
| G-2 | "41개 스킬" (SPEC-004 REQ-27) | 카테고리 표현만. `ls`로 실제 수 세면 다를 수 있음 | 실제 대상 리스트 Appendix |
| G-3 | "7 이식 스킬" (SPEC-005 전체) | 7개는 정확. 단 **각 스킬의 하위 파일 수**(modules/, examples.md 등) 언급 없음. 이식 시 이 파일들 처리 방식 누락 | Phase 2 Tasks에 "SKILL.md + 하위 파일 전수 복사 + 네임스페이스 변환" 명시 |
| G-4 | "모든 frontmatter 필드 보존" (SPEC-004 REQ-24, SPEC-005 REQ-07) | "모든"의 범위 불명 - moai 네임스페이스가 frontmatter의 description에 있으면? | description 필드도 변환 대상임을 명시 |
| G-5 | "업스트림 v2.13.x" (SPEC-005 전체) | v2.13.0 / v2.13.1 / v2.13.2 차이 기술됨 ✓ | 우수 |
| G-6 | "5 핵심 개념 보존" (SPEC-005 AC-08) | 이름 기반 존재 확인만. 개념 의미 보존은 grep으로 불가 | 의미 보존은 리뷰어 수동 게이트로 분리 명시 |
| G-7 | "허용 예외 제외" (SPEC-005 AC-35 전반) | 예외가 5종인지 6종인지 spec 내부 불일치. 3.1 환경 섹션(예외 5종) vs 3.2 Scenario 필터 | 예외 목록 단일 출처화 (환경 섹션에만 기록) |

---

## 6. 외부 검증 필요 항목

아래 항목은 내부 코드 분석으로 결정 불가. **Anthropic 공식 문서 또는 업스트림 실제 커밋 fetch** 필요.

| # | 항목 | 검증 방법 |
|---|---|---|
| E-1 | `claude-opus-4-7` 모델 ID 존재 | Anthropic Models API 공식 문서 (현재 이 세션도 Opus 4.7로 동작 중이므로 존재 확률 높음) |
| E-2 | `CLAUDE_CODE_EFFORT_LEVEL` 환경변수 지원 | Claude Code v2.12.0+ 공식 릴리즈 노트 |
| E-3 | `thinking.budget_tokens` HTTP 400 거부 (Opus 4.7) | Anthropic Messages API 공식 문서 (`thinking` 파라미터 섹션) |
| E-4 | `user-invocable: false` frontmatter 지원 | Claude Code Skills 공식 문서 |
| E-5 | `charmbracelet/x/powernap` v0.1.4 changelog | GitHub Releases |
| E-6 | 업스트림 d29771e 커밋 내용 (하드코딩 상수 12건) | `gh api repos/modu-ai/moai-adk/commits/d29771e` |
| E-7 | 업스트림 `moai-workflow-gan-loop` 본문 | git clone 또는 tarball fetch |
| E-8 | v2.14.0 Detection Improvements 실제 영향 | 현재 ae-adk 코드베이스와 diff |

---

## 7. 우선순위별 권고사항

### 7.1 Blocker (머지 전 필수)

| # | SPEC | 권고 | 대응 Phase |
|---|---|---|---|
| P0-1 | 004 | REQ-01 경로를 `system.yaml.tmpl`로 수정. AC-01 grep 경로도 동시 수정 | Phase 1.1 |
| P0-2 | 004 | REQ-19 경로 검증 및 실제 파일 위치 확인 (현행 `sections/` 하위에 lsp.yaml 없음) | Phase 5.3 전 |
| P0-3 | 004 | v2.14.0 대응 결정: (a) SPEC-004 스코프 확장, (b) SPEC-UPDATE-006 초안 작성, (c) EX-13로 명시 이월 중 하나 선택 | Phase 0 |
| P0-4 | 005 | REQ-08 흡수 전에 업스트림 `moai-workflow-gan-loop` 본문 fetch 및 기존 5개 개념과의 overlap 분석 단계 삽입 | Phase 2.6 → 3.4 사이 |
| P0-5 | 005 | AC-35의 6개 Scenario 중 grep 취약점 5건을 Go 정규식 기반 테스트로 교체 | Phase 1.2 |

### 7.2 High (머지 전 강력 권고)

| # | SPEC | 권고 |
|---|---|---|
| P1-1 | 004 | REQ-04 "12개 상수" 목록을 Appendix로 첨부 |
| P1-2 | 004 | REQ-19 "16개 언어 × file_extensions" 매핑 매트릭스 Appendix |
| P1-3 | 004 | REQ-27 "41개 스킬" 실제 대상 리스트 Appendix |
| P1-4 | 004 | Phase 3·Phase 5 시작 전 "업스트림 fetch 전략" 명시 (gh api / git shallow / raw URL 중 택일) |
| P1-5 | 005 | REQ-09 `ae-ref-frontend-patterns` 카테고리 충돌 — `ae-domain-frontend-patterns` 대안 재검토 |
| P1-6 | 005 | SHA-256 보존 로직을 `internal/template/integrity/` 공통 패키지로 추출하여 REQ-23·REQ-26이 공유 |
| P1-7 | 005 | REQ-28 16개 언어 DB 탐지 매트릭스 Appendix (언어 × 매니페스트 × ORM) |
| P1-8 | 공통 | v2.14.0 반영 전략 확정 (P0-3과 연결) |

### 7.3 Medium (품질 개선)

| # | 권고 |
|---|---|
| P2-1 | AC-26 "HUMAN GATE" 검증을 H2/H3 헤더 패턴으로 엄격화 |
| P2-2 | REQ-22 다국어 번역 품질 게이트 명시 (검토자 체크박스 또는 fallback 정책) |
| P2-3 | A-08 template_version 누락 원인 조사 (Phase 1.1 앞에 근본 원인 분석 단계) |
| P2-4 | CLAUDE.md 기본값 변경 시 CHANGELOG에 비용 영향 고지 (R-03, R-05) |
| P2-5 | `.agency.archived/`를 `.gitignore`에 자동 추가 (R-NEW-2) |

### 7.4 Low (선택적 개선)

| # | 권고 |
|---|---|
| P3-1 | Skill("ae-ref-frontend-patterns") 예시 사용법 문서화 |
| P3-2 | 16개 언어 fixture를 `testdata/db-detection-fixtures/` 하위에 organize |
| P3-3 | SPEC-005 R-10 (CLAUDE.md 40K) 제거 또는 "Low, 현행 50% 여유" 표기 |

---

## 8. 테스트 커버리지 갭 분석

사용자 요청: "테스트 포함하여 검증"

### 8.1 SPEC-004 테스트 매트릭스

| 테스트 유형 | 있음 | 없음 (권고) |
|---|---|---|
| 단위 (Go) | Effort 상수, agentEffortMap, normalizeModel, ApplyEffortPolicy | `AstGrepGate` YAML 언마샬링, `__updated_input_marker__` edge cases (공백/유니코드) |
| 통합 (embed) | LSP compliance 3종, HUMAN GATE 회귀 | settings.local.json merge 원자성 (R-NEW-3) |
| E2E | 없음 | `ae init` → effort 주입 → Claude Code 기동 |
| Property-based | 없음 | `normalizeModel` - 임의 모델 ID 입력 → 안정성 |
| Race | `go test -race` 명시 | Phase 4 HUMAN GATE 회귀와 병렬 실행 시 오염 없음 검증 |

### 8.2 SPEC-005 테스트 매트릭스

| 테스트 유형 | 있음 | 없음 (권고) |
|---|---|---|
| 단위 | migrate 시나리오 4종, DB sync 4종, DB detection 2종 | agency 삭제 회귀 (`TestAgencyAgentsRemoved`), 리네이밍 참조 전수 (`TestRenamedSkillReferences`) |
| 통합 | namespace_integrity 6종, design_skill, db_skill | CLAUDE.md 자 수 regression, `user-invocable: false` 파싱 |
| 시나리오 | migrate SIGINT/resume/rollback | `.agency/` 없는 프로젝트에서 `ae migrate agency` 실행 시 graceful skip |
| Property-based | 없음 | 16개 언어 fixture permutation |

### 8.3 권고: 신규 테스트 5종

```go
// SPEC-005 보강
TestRenamedSkillReferences       // ae-agency-client-interview/frontend-patterns 구 이름 전수 0건
TestTemplateNamespaceIntegrity   // YAML 파싱 기반 구조화 검증 (grep 취약점 해소)
TestAgencyRemovalCompleteness    // 6 에이전트 + 5 스킬 + .agency/ + rules/agency/ 전수 부재
TestMigrateStateAtomicity        // .migration-state.json O_EXCL + Rename
TestMigrateAgencyGracefulSkip    // .agency/ 없을 때 에러 없이 no-op 종료
```

---

## 9. 결론

SPEC-UPDATE-004/005는 전반적으로 **성숙한 설계**이며, 특히 SPEC-005의 AC-35(네임스페이스 무결성)는 self-aware하고 방어적인 구조다. 다만 아래 두 가지 **Critical 이슈**는 머지 전 반드시 해소해야 한다.

1. **SPEC-004 REQ-01 파일 경로 오류** — 현행 템플릿에 `system.yaml`이 없고 `system.yaml.tmpl`만 존재. 이 경로를 맞지 않게 실행하면 본 SPEC의 첫 번째 Task부터 실패한다.
2. **v2.14.0 스코프 갭** — 같은 날 릴리즈된 v2.14.0 Utility Hardening이 어디에도 포함되지 않아 머지 직후 즉시 새 SPEC이 필요해진다. 결정이 필요하다: (a) 004 스코프 확장, (b) UPDATE-006 선제 초안화, (c) 명시적 EX-13 이월.

그 외에는 "일반화의 오류" 관점에서 **구체 목록이 외부 참조·카테고리로 흩어진 경우가 다수**(상수 12개, 언어 16개, 스킬 41개) — Appendix로 single source of truth를 확보해야 구현 시 누락을 방지할 수 있다.

SPEC-005의 흡수 작업(REQ-08)은 **업스트림 본문을 보지 않고 병합 계획**을 세운 부분이 유일한 본질적 리스크다. Phase 2.6 완료 후 공식 diff 단계를 추가하면 해소 가능하다.

### 9.1 Go/No-Go 권고

| SPEC | 현재 상태 | Blocker 해소 후 권고 |
|---|---|---|
| SPEC-UPDATE-004 | **조건부 No-Go** | P0-1·P0-2·P0-3 해소 후 Go |
| SPEC-UPDATE-005 | **조건부 Go** | P0-4·P0-5 해소 권장하되 실행 중에도 대응 가능 |

### 9.2 다음 단계 제안

1. 이 리뷰에서 도출된 Blocker 5건을 사용자(Angeleyes)가 검토 후 SPEC에 반영
2. v2.14.0 대응 결정(P0-3) 확정
3. SPEC-004 spec.md·acceptance.md의 경로 수정 (P0-1·P0-2)
4. SPEC-005 Phase 1.2 `namespace_integrity_test.go` skeleton 작성 시 Go 정규식 구조로 설계 (P0-5)
5. 본 리뷰 문서를 `.moai/specs/SPEC-UPDATE-004/review.md`·`.moai/specs/SPEC-UPDATE-005/review.md`로 심볼릭/복사하여 SPEC 실행 중 상시 참조 가능하도록 배치

---

**리뷰 완료.** 본 문서는 evaluator-active의 skeptical 평가 원칙에 따라 작성되었으며, 의도적으로 결함을 찾는 방향으로 편향되어 있음을 밝힌다. 개선 제안이 아닌 것은 모두 "양호" 또는 "타당"으로 명시하였다.
