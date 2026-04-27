# `ae-upstream-sync` 스킬 설계 제안서

> 본 문서는 업스트림 동기화 작업을 자동화/반자동화하는 Claude Code 스킬의 **설계 제안**이다.
> 실제 스킬 파일 생성은 사용자 승인 후 별도 단계로 진행한다 (파일 배포 위치 결정 필요).

---

## 1. 목표

### 1.1 문제 정의

현재 ae-adk 업스트림 동기화는 100% 수동이다.

| 단계 | 현재 소요 시간 | 에러 가능성 |
|---|---|---|
| 릴리즈 감지 | 수 분 (수동 gh api 호출) | 저 |
| Delta 분석 | 수 시간 (파일별 육안 비교) | **고** (누락 빈번) |
| SPEC 초안 | 수 시간 (매번 템플릿 재작성) | 중 |
| 이식 변환 | 수 일 (파일별 grep + Edit 반복) | **고** (잔존 false negative) |
| 검증 | 수 시간 (grep 명령 반복) | 중 |

업스트림이 주당 1-3회 릴리즈되는 현실에서 수동 추적은 시간당 생산성을 낮추고 누락을 유발한다.

### 1.2 스킬 목표

- **Low-hanging fruit 자동화**: 릴리즈 감지·delta 요약·SPEC 스켈레톤 생성
- **에러 방지**: 파일 경로 실존 체크, 네임스페이스 변환 미완 감지
- **학습 축적**: 반복되는 이식 패턴을 재사용 가능한 치환 규칙으로 등록

### 1.3 비목표

- **완전 자동 SPEC 작성**: 사용자 판단(스코프 분할, 결정 사항)을 대체하지 않는다
- **자동 이식 실행**: 파일 변환은 사용자 확인을 거친 후 수행 (민감한 작업)
- **GitHub Action 배포**: 본 문서는 CLI 스킬에 집중. Action은 후속 논의

---

## 2. 스킬 메타데이터

### 2.1 YAML Frontmatter (초안)

```yaml
---
name: ae-upstream-sync
description: ae-adk 업스트림(modu-ai/moai-adk) 동기화 전용 스킬. 릴리즈 감지, delta 분석, SPEC 초안 생성, 네임스페이스 변환 검증을 반자동화.
---
```

### 2.2 트리거 조건

사용자가 다음 중 하나를 요청할 때 스킬 로드:

- 직접 호출: `/ae upstream-check`, `/ae upstream-analyze`
- 자연어: "moai-adk 업데이트 확인", "업스트림 새 릴리즈 있어?", "업스트림 반영 준비"
- 파일 수정 맥락: `.moai/specs/SPEC-UPDATE-*` 디렉토리 작업 중

---

## 3. 서브커맨드 설계

스킬은 5개 서브커맨드를 제공한다.

### 3.1 `/ae upstream-check`

**목적**: 신규 릴리즈 감지 및 현재 동기화 상태 비교

**입력**: 없음 (`--since <version>` 옵션)

**동작**:
1. `internal/template/templates/.ae/config/sections/system.yaml.tmpl`에서 현재 `template_version` 추출
2. `gh api repos/modu-ai/moai-adk/releases`로 최근 릴리즈 조회
3. 현재 버전 이후 신규 릴리즈 리스트 출력
4. 각 릴리즈의 `body` 첫 섹션(Summary) 요약 제시

**출력 포맷**:
```markdown
## 신규 릴리즈 감지

현재 동기화 버전: v2.13.2
최신 업스트림: v2.14.0

| 버전 | 날짜 | 요약 |
|---|---|---|
| v2.14.0 | 2026-04-24 | Utility Hardening — MX validator correctness, tree-sitter 16-language |
| v2.13.2 | 2026-04-23 | (현재 반영 완료) |

권고: SPEC-UPDATE-006 착수 또는 004에 흡수
```

### 3.2 `/ae upstream-analyze`

**목적**: 특정 버전 범위 delta 분석

**입력**: `<from-version>..<to-version>` (예: `v2.13.2..v2.14.0`)

**동작**:
1. 업스트림 shallow clone (`/tmp/moai-upstream-<version>/`)
2. `git diff --name-only <from>..<to>`로 변경 파일 목록 추출
3. 파일별로 4개 버킷(A/B/C/D) 자동 분류 시도
   - `.claude/skills/moai-*` 신규 → **A. 이식**
   - `.claude/skills/moai-*` 수정 → **B. 갱신**
   - `.claude/skills/moai-*` 삭제 → **C. 삭제**
   - `docs-site/*`, `scripts/*` 등 → **D. 무관**
4. ae-delta.md와 대조하여 고유 자산 영향 경고
5. 변경 요약 Markdown으로 출력 (SPEC spec.md의 "환경 > 업스트림 변경 요약" 섹션에 바로 복사 가능)

**출력 예시**:
```markdown
## Delta 분석: v2.13.2 → v2.14.0

### A. 이식 후보 (0건)
(신규 스킬/에이전트 없음)

### B. 갱신 (12건)
- internal/hook/mx/validator.go (+145 -32) — method receiver detection
- internal/tool/complexity/tree_sitter.go (+289 -0) — new
- ...

### C. 삭제 (0건)

### D. 무관 (4건)
- docs-site/README.md (Hextra docs)
- ...

### 경고 (ae-delta.md 기반)
- 없음. 모든 변경이 ae 고유 자산과 독립.
```

### 3.3 `/ae upstream-draft`

**목적**: SPEC-UPDATE-NNN 스켈레톤 생성

**입력**: `<NNN>` SPEC 번호, `<from>..<to>` 버전 범위

**동작**:
1. `.moai/specs/SPEC-UPDATE-NNN/` 디렉토리 생성
2. 템플릿 기반 3개 파일 생성:
   - `spec.md` — Environment/Assumptions/Requirements/Exclusions/Risks 스켈레톤
   - `plan.md` — Phase 구조 스켈레톤
   - `acceptance.md` — AC 플레이스홀더
3. `/ae upstream-analyze` 결과를 `spec.md`의 "업스트림 변경 요약" 섹션에 삽입
4. 체크리스트 사본 (`docs/upstream-sync/checklist.md` → `progress-checklist.md`) 복사
5. 사용자에게 "이제 REQ 항목을 채우고 plan.md Phase를 상세화하세요" 안내

**중요**: 이 서브커맨드는 **스켈레톤만 생성**한다. REQ 작성은 사용자 판단 영역.

### 3.4 `/ae upstream-verify`

**목적**: SPEC 작성 중 또는 실행 중 파일 경로 + 네임스페이스 무결성 검증

**입력**: `<SPEC-ID>` (예: `SPEC-UPDATE-004`)

**동작**:

1. **파일 경로 실존 검증**
   - spec.md 본문에서 경로 추출
   - 각 경로의 실존 확인 (`os.Stat`)
   - `✓` / `✗ MISSING` 출력

2. **네임스페이스 무결성 검증** (`Skill("ae-tool-ast-grep")` 활용)
   - 변환 규칙표 로드
   - `internal/template/templates/` 전수 스캔
   - 허용 예외를 제외한 moai 참조 검출
   - Go 정규식 기반 (grep false positive 회피)

3. **AC 자동화 커버리지 체크**
   - acceptance.md 본문에서 `Validation:` 섹션 추출
   - 각 Validation이 `go test` / `grep` / 수동 중 어느 것에 해당하는지 분류
   - 수동 비율이 20% 초과 시 경고

**출력**:
```markdown
## SPEC-UPDATE-004 검증

### 파일 경로 (12건 중)
✓ internal/template/templates/.ae/config/sections/llm.yaml
✗ MISSING internal/template/templates/.ae/config/sections/system.yaml
✗ MISSING internal/template/templates/.ae/config/sections/lsp.yaml
...

### 네임스페이스 무결성
- 이식 7스킬: 2건 잔존 (ae-workflow-gan-loop/SKILL.md:45, :78)
- 전체: 총 5건

### AC 자동화
- 자동: 21건
- 부분 자동화: 5건
- 수동: 2건 (AC-09 wizard UI, AC-17 Windows fixture)
- 경고 없음 (수동 비율 7%)
```

### 3.5 `/ae upstream-delta-update`

**목적**: `.moai/upstream/ae-delta.md` 갱신

**입력**: SPEC-UPDATE-NNN 머지 후 자동 또는 `--spec SPEC-UPDATE-NNN`

**동작**:
1. SPEC 완료로 추가된 ae 고유 자산 감지 (스킬 이름, 플랫폼 명령 등)
2. ae-delta.md의 해당 섹션에 append
3. 업데이트 이력 테이블에 항목 추가
4. git diff 출력 후 사용자 확인 요청

---

## 4. 파일 구조 (제안)

### 4.1 배포 위치 옵션

| 옵션 | 경로 | 장점 | 단점 |
|---|---|---|---|
| **A** | 프로젝트 루트 `.claude/skills/ae-upstream-sync/` | 개발자 도구로 즉시 사용 가능 | `.claude/`가 moai-adk 영역이라 정책상 위반 가능 |
| **B** | `internal/template/templates/.claude/skills/ae-upstream-sync/` | ae-adk 사용자 프로젝트에도 배포 | 최종 사용자에게 불필요한 기능 배포 |
| **C** | 별도 리포지토리 (예: `ae-adk-tools`) | 관심사 분리 | 관리 오버헤드 |
| **D** | `docs/upstream-sync/skill-proposal/` 하위 (임시) | 단순 문서로 시작 | 실제 스킬 로드 안됨 |

**권고**: **옵션 A**. 이유:
- 업스트림 동기화는 **ae-adk 개발자 전용 도구**
- 최종 사용자(ae-adk로 프로젝트 하는 사람)에게는 불필요
- `.claude/` 영역 수정이 policy violation 우려가 있으나, **사용자 명시적 승인**(이 제안서 채택)이 있으면 예외 허용

옵션 A 채택 시 디렉토리 구조:

```
.claude/skills/ae-upstream-sync/
├── SKILL.md              # 메인 스킬 정의 + 서브커맨드 라우팅
├── commands/
│   ├── check.md          # /ae upstream-check 상세
│   ├── analyze.md        # /ae upstream-analyze 상세
│   ├── draft.md          # /ae upstream-draft 상세
│   ├── verify.md         # /ae upstream-verify 상세
│   └── delta-update.md   # /ae upstream-delta-update 상세
├── templates/
│   ├── spec-template.md       # SPEC-UPDATE-NNN/spec.md 스켈레톤
│   ├── plan-template.md       # plan.md 스켈레톤
│   └── acceptance-template.md # acceptance.md 스켈레톤
└── rules/
    ├── namespace-conversion.md    # 변환 규칙표 (단일 출처)
    └── file-path-patterns.md     # 경로 패턴 (실존 검증용)
```

### 4.2 SKILL.md 골격

```markdown
---
name: ae-upstream-sync
description: ae-adk 업스트림(modu-ai/moai-adk) 동기화 반자동화. Release 감지, delta 분석, SPEC 초안 생성, 네임스페이스 검증을 제공.
---

# ae-upstream-sync

## 트리거

- 직접: /ae upstream-*
- 자연어: "moai-adk 업데이트", "업스트림 릴리즈", "업스트림 반영"
- 맥락: .moai/specs/SPEC-UPDATE-* 편집 중

## 서브커맨드 라우팅

| 서브커맨드 | 파일 | 요약 |
|---|---|---|
| check | commands/check.md | 신규 릴리즈 감지 |
| analyze | commands/analyze.md | 특정 버전 범위 delta 분석 |
| draft | commands/draft.md | SPEC 스켈레톤 생성 |
| verify | commands/verify.md | 파일 경로 + 네임스페이스 검증 |
| delta-update | commands/delta-update.md | ae-delta.md 갱신 |

## 실행 디렉티브

Step 1: 서브커맨드 판별
Step 2: 해당 commands/*.md 로드
Step 3: Read rules/namespace-conversion.md (변환 작업 시)
Step 4: 지시된 동작 실행
Step 5: 결과를 Markdown으로 표현

## 규칙 참조

- 파일 경로 정확성 게이트: rules/file-path-patterns.md
- 네임스페이스 변환 규칙: rules/namespace-conversion.md
- 전체 프로세스: docs/upstream-sync/process-guide.md
- 체크리스트: docs/upstream-sync/checklist.md
```

### 4.3 rules/namespace-conversion.md 골격

이 파일은 **네임스페이스 변환 규칙의 Single Source of Truth**.
- SPEC-UPDATE-005 spec.md의 변환 규칙표
- process-guide.md 4.4
- `internal/template/namespace_integrity_test.go`

세 곳이 이 파일을 참조하도록 구성. 규칙 변경 시 한 곳만 수정하면 전체에 반영.

---

## 5. 외부 도구 통합

### 5.1 `gh` CLI

`/ae upstream-check`, `/ae upstream-analyze`가 gh API를 호출. 사용자 `gh auth login` 사전 완료 필요.

### 5.2 `git` (shallow clone)

`/ae upstream-analyze`가 `git clone --depth 30 --branch vX.Y.Z`로 업스트림 스냅샷 획득.

### 5.3 `ae-tool-ast-grep`

네임스페이스 검증 시 ast-grep으로 구조화된 검색 수행. Go 정규식 fallback 가능.

### 5.4 `mcp__context7__` (외부 문서)

Anthropic API, Claude Code 신규 기능의 공식 문서 확인 (E-1 ~ E-4 외부 검증 필요 항목).

---

## 6. 점진적 도입 계획

**Phase 1 (MVP, 1-2 세션)**:
- `SKILL.md` 뼈대
- `/ae upstream-check` 서브커맨드 (가장 단순)
- `rules/namespace-conversion.md` 작성

**Phase 2 (주간 반복, 2-3 세션)**:
- `/ae upstream-verify` (파일 경로 + 네임스페이스)
- SPEC-UPDATE-004/005 실전 적용

**Phase 3 (월간)**:
- `/ae upstream-analyze` (delta 분석)
- SPEC 초안 템플릿 정교화

**Phase 4 (반기)**:
- `/ae upstream-draft` (스켈레톤 생성)
- `/ae upstream-delta-update` (ae-delta.md 자동 갱신)
- GitHub Action 연계 (주 1회 릴리즈 알림)

---

## 7. 대안 검토

### 7.1 GitHub Action 단독

**장점**: 완전 자동화, 주기 실행
**단점**: 판단 사항(스코프 분할, 결정 3건)은 여전히 수동. Action이 이슈만 생성하고 대기.

**결론**: **스킬 + Action 조합**이 이상적. 스킬이 Action 이슈를 받아 SPEC 초안 착수.

### 7.2 전용 Go CLI (`ae upstream sync`)

**장점**: 성능, 바이너리 배포, Claude Code 무관 실행
**단점**: Claude Code 협업 역량(자연어 해석, 문서 편집) 상실

**결론**: Go CLI는 `ae upstream check`, `ae upstream analyze` 등의 **저수준 오퍼레이션**에 적합. 스킬이 CLI를 감싸는 아키텍처 권고.

### 7.3 계속 수동

**장점**: 구현 비용 0
**단점**: 업스트림 페이스에 뒤처짐, 누락 발생

**결론**: 장기적으로 지속 불가.

---

## 8. 리스크

| 리스크 | 완화 |
|---|---|
| `gh auth` 미로그인 시 실패 | 스킬 시작 시 `gh auth status` 체크 후 안내 |
| 업스트림 clone 시 임시 디스크 사용 증가 | `/tmp/` 활용 + 종료 시 정리 스크립트 |
| 네임스페이스 변환 규칙 업데이트 누락 | SSOT 파일(`rules/namespace-conversion.md`)로 한 곳만 수정 |
| 스킬 서브커맨드 간 의존성 혼동 | commands/*.md에 전제 조건 명시 |
| 자동 분류(A/B/C/D)의 오분류 | 사용자 확인 단계 유지, 단순 heuristic만 적용 |

---

## 9. 다음 단계

본 제안서를 사용자가 검토 후 아래 3개 질문에 답하면 구현 착수 가능:

1. **배포 위치**: 옵션 A (프로젝트 루트 `.claude/skills/`) 허용?
2. **MVP 범위**: Phase 1(check + verify)부터 시작하는가?
3. **도구 스택**: gh CLI + ast-grep + context7 MCP 조합 OK?

답변에 따라 후속 SPEC 또는 직접 스킬 구현으로 진행.

---

## 10. 관련 참조

- [README.md](./README.md) — 업스트림 동기화 개요
- [process-guide.md](./process-guide.md) — 5단계 수동 프로세스
- [checklist.md](./checklist.md) — 반복 사용 체크리스트
- [SPEC-UPDATE-004-005-review.md](../specs-review/SPEC-UPDATE-004-005-review.md) — 이전 SPEC 리뷰에서 도출된 요건

---

Version: 1.0.0 (proposal)
Last Updated: 2026-04-24
Status: Awaiting user decision on deployment location
