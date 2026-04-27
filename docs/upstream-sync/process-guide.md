# ae-adk 업스트림 동기화 프로세스

> [README.md](./README.md)를 먼저 읽고 이 문서로 진행한다.
> 이 문서는 **신규 업스트림 릴리즈 → SPEC-UPDATE 착수 → 머지**까지의 5단계를 설명한다.

---

## 개요

```
Phase 0 (Pre-check) → Phase 1 (Detect) → Phase 2 (Analyze)
                             ↓
        Phase 3 (SPEC Draft) → Phase 4 (Execute) → Phase 5 (Verify & Merge)
```

각 Phase는 독립적으로 재실행 가능하며 중단 시 체크리스트([checklist.md](./checklist.md))로 복구한다.

---

## Phase 0: 사전 점검 (Pre-check)

SPEC 작업 착수 **전에** 반드시 수행.

### 0.1 ae-delta.md 읽기

```bash
# 매번 읽을 것. 고유 자산 목록은 계속 증가한다.
cat .moai/upstream/ae-delta.md
```

확인 항목:
- 최근 추가된 고유 자산
- 업스트림에서 삭제됐지만 ae에서 보존해야 할 항목
- 제거된 기능 목록 (재도입 방지)
- 네임스페이스 변환 예외 목록

### 0.2 메모리 체크

```bash
# 관련 메모리 항목 (auto-memory)
ls ~/.claude/projects/D--Sources-ae-adk/memory/ | grep -E "upstream|moai|update"
```

주요 메모리:
- `project_moai_update_todo.md` — 이전 SPEC 진행 상태
- `feedback_upstream_sync_checklist.md` — SPEC-UPDATE 작업 규칙
- `project_ae_vs_moai.md` — ae-adk vs moai-adk 수정 경계

### 0.3 현재 동기화 버전 확인

```bash
grep -E "template_version|^  version" \
  internal/template/templates/.ae/config/sections/system.yaml.tmpl
```

**주의**: 파일명이 `.tmpl` 확장자를 갖는다. SPEC 작성 시 경로를 `.yaml.tmpl`로 기록한다.

### 0.4 워킹 트리 청소

```bash
git status
# Untracked / Modified 파일 없을 때 착수. 있으면 commit 또는 stash.
```

---

## Phase 1: 업스트림 변경 감지 (Detect)

### 1.1 최신 릴리즈 조회

```bash
# 최근 10개 릴리즈 요약
gh api repos/modu-ai/moai-adk/releases \
  --jq '.[0:10] | .[] | "\(.tag_name) | \(.published_at)"'

# 특정 버전 릴리즈 노트 (예: v2.14.0)
gh api repos/modu-ai/moai-adk/releases/tags/v2.14.0 --jq '.body'
```

### 1.2 Delta 범위 결정

| 상황 | 스코프 |
|---|---|
| 최근 릴리즈가 패치만 (예: v2.13.2 → v2.13.3) | 단일 패치 적용 또는 누적 반영까지 대기 |
| 여러 마이너 릴리즈 누적 (v2.10.2 ~ v2.13.2) | 주제별 그룹핑 (품질·성능 / 구조 변경) |
| 구조적 재편 포함 (Agency 삭제 등) | 별도 SPEC으로 분리 |
| 보안 패치 포함 | 단독 긴급 SPEC (SPEC-SECURITY-PATCH-NNN) |

### 1.3 파일 변경 덤프

업스트림 원본 파일을 확보하는 방법 (우선순위 순):

**옵션 A: git shallow clone (권장)**
```bash
mkdir -p /tmp/moai-upstream && cd /tmp/moai-upstream
git clone --depth 30 --branch main https://github.com/modu-ai/moai-adk.git .
# 특정 태그 체크아웃
git checkout v2.14.0
```

**옵션 B: gh api raw fetch (단일 파일)**
```bash
gh api repos/modu-ai/moai-adk/contents/.claude/skills/moai-domain-brand-design/SKILL.md?ref=v2.13.0 \
  --jq '.content' | base64 -d > /tmp/moai-brand-design.md
```

**옵션 C: tarball fetch (전체 스냅샷)**
```bash
gh api repos/modu-ai/moai-adk/tarball/v2.14.0 > /tmp/moai-v2.14.0.tar.gz
tar -xzf /tmp/moai-v2.14.0.tar.gz -C /tmp/
```

### 1.4 diff 생성

```bash
# 이전 반영 버전 대비 변경 파일 목록
cd /tmp/moai-upstream
git diff --name-only v2.10.1..v2.14.0 \
  -- '.claude/' '.moai/' 'CLAUDE.md' '*.go' \
  > /tmp/changed-files.txt
wc -l /tmp/changed-files.txt
```

---

## Phase 2: 변경 분석 (Analyze)

### 2.1 변경 분류

모든 변경을 다음 4개 버킷 중 하나로 분류한다.

| 버킷 | 의미 | 처리 |
|---|---|---|
| **A. 이식 (Port)** | 업스트림 신규 자산 (스킬·에이전트·룰 등) | ae 네임스페이스 변환 후 `internal/template/templates/`에 복사 |
| **B. 갱신 (Update)** | 기존 자산 버그픽스·개선 | ae 해당 파일의 대응 부분 재이식 |
| **C. 삭제 (Remove)** | 업스트림 제거 자산 | **ae-delta.md 대조 후** 보존/삭제 결정 |
| **D. 무관 (Skip)** | ae와 무관한 변경 (glm, docs-site 등) | EX-XX 항목으로 spec에 명시 |

### 2.2 분류 체크리스트

각 변경에 대해 확인:

- [ ] `ae-delta.md`의 고유 자산 목록과 겹치는가?
- [ ] 네임스페이스 변환 대상인가? (moai-xxx / .moai/ / <moai> 등)
- [ ] 기존 ae-adk 파일과 이름 충돌이 있는가?
- [ ] 의존성 있는 Go 코드가 있는가? (import, 구조체 필드)
- [ ] 테스트 fixture 필요한가?

### 2.3 Go 코드 변경 분석

```bash
# 업스트림 Go 패키지 구조 확인
cd /tmp/moai-upstream
find . -name "*.go" -newer <이전-sync-시점-파일> | grep -v _test.go
```

분석 포인트:
- 새 패키지(`internal/foundation/effort.go` 등) — ae도 동일 경로로 추가
- 기존 패키지 API 변경 — breaking / non-breaking 구분
- hook 핸들러 변경 — ae hook 체계와 매핑
- 의존성 업데이트 (`go.mod`) — breaking change 점검

### 2.4 리스크 맵 작성

고위험 변경은 전용 리스크 항목으로 분리.

리스크 심각도 기준:
- **Critical**: cascade 유발·롤백 불가·데이터 손실 (예: .agency/ 삭제)
- **High**: 빌드 실패·주요 기능 미동작 (예: 스키마 변경)
- **Medium**: 부가 기능 품질 저하 (예: UI 변경)
- **Low**: 텍스트 변경·주석·문서

---

## Phase 3: SPEC 초안 작성 (Draft)

### 3.1 SPEC 파일 구조

```
.moai/specs/SPEC-UPDATE-NNN/
├── spec.md          # 요구사항 (EARS 형식)
├── plan.md          # Phase별 구현 계획
├── acceptance.md    # AC 시나리오
└── research.md      # (옵션) 업스트림 변경 심층 분석
```

### 3.2 spec.md 필수 섹션

1. **Environment** — 프로젝트 컨텍스트·기존 구조·업스트림 변경 요약·ae 고유 자산 보존
2. **Assumptions** — 기술·비즈니스 가정
3. **Requirements** — EARS 형식 REQ 항목들
4. **Exclusions** — 명시적 제외 (다른 SPEC 이월 또는 대응 불필요)
5. **Affected Files** — 파일별 대상 리스트
6. **Risks** — R-NN 형식 리스크와 완화책
7. **Acceptance Criteria Summary** — AC 요약 (상세는 acceptance.md)

### 3.3 EARS 형식 규칙

```
# 일반 이벤트 트리거
WHEN [조건] **THEN** 시스템은 [결과]한다.

# 상태 기반 트리거
WHILE [상태] **THEN** 시스템은 [결과]한다.

# 조건부 로직
IF [조건] **THEN** 시스템은 [결과]한다.

# 복합
WHILE [상태], **IF** [조건] **THEN** 시스템은 [결과]한다.
```

### 3.4 파일 경로 정확성 게이트

SPEC 작성 후 **반드시** 실제 파일 존재 여부 검증:

```bash
# spec.md 본문에서 경로 추출 후 실존 확인
grep -oE '(internal/template/templates|\.claude|\.ae|\.moai)/[^\s`]+' \
  .moai/specs/SPEC-UPDATE-NNN/spec.md \
  | sort -u \
  | while read f; do
      [ -e "$f" ] && echo "✓ $f" || echo "✗ MISSING $f"
    done
```

**Critical**: `✗ MISSING` 항목이 나오면 SPEC 실행 전 경로 수정 필수.
(SPEC-UPDATE-004의 `system.yaml` vs `system.yaml.tmpl` 사례 참조)

### 3.5 구체 목록 Appendix

다음 항목은 카테고리·숫자 대신 **구체 목록**으로 명시:

- "N개 상수" → 상수명 전수 리스트
- "16개 언어" → 언어명 × 확장자 × ORM 매트릭스
- "41개 스킬" → 디렉토리명 리스트 (`ls` 결과)

Appendix가 있으면 일반화 오류 예방 + 이후 리뷰자 이해도 상승.

---

## Phase 4: 실행 (Execute)

### 4.1 브랜치 전략

```bash
git checkout -b feature/SPEC-UPDATE-NNN
```

### 4.2 Phase 단위 커밋

- Phase 1 완료 시점에 첫 커밋
- 각 Phase 경계에서 커밋 + 간단한 diff stat 기록
- `feat(spec-update-NNN): [Phase X] description` (bracket-scope)

### 4.3 네임스페이스 변환 원칙

이식(Port) 작업 시 **매 파일**에 대해 3단계:

1. **(a) 식별**: 원본 파일에서 moai 참조 전수 grep
   ```bash
   grep -rn "moai" <원본-파일> | tee /tmp/before.txt
   ```

2. **(b) 매핑**: 변환 규칙표 적용 (아래 4.4 참조)

3. **(c) 재검증**: 변환 후 grep 재확인
   ```bash
   grep -rn "moai" <변환-후-파일> \
     | grep -v "<허용-예외>" \
     | tee /tmp/after.txt
   # after.txt가 비어있어야 통과
   ```

### 4.4 네임스페이스 변환 규칙

| 원본 (moai-adk) | 변환 (ae-adk) | 예외 |
|---|---|---|
| `.claude/agents/moai/` | `.claude/agents/ae/` | - |
| `.claude/rules/moai/` | `.claude/rules/ae/` | - |
| `.claude/skills/moai-*` | `.claude/skills/ae-*` | `.claude/skills/moai/workflows/`, `.claude/skills/moai/team/` |
| `.claude/hooks/moai/` | `.claude/hooks/ae/` | - |
| `.claude/commands/moai/` | `.claude/commands/ae/` | - |
| `.moai/` | `.ae/` | - |
| `<moai>DONE</moai>` | `<ae>DONE</ae>` | - |
| `Skill("moai-xxx")` | `Skill("ae-xxx")` | workflows/team 카테고리 |
| `/moai design` 등 | `/ae design` | - |
| `moai migrate`, `moai update` | `ae migrate`, `ae update` | - |
| `moai-adk` (프로젝트명) | `ae-adk` | `modu-ai/moai-adk` 보존 |
| `MoAI` (대문자) | `AE` | - |

**예외 경로의 근거**:
- `modu-ai/moai-adk` — 업스트림 GitHub 레포 식별자
- `github.com/modu-ai/...` — Go 패키지 import path
- CHANGELOG/README 인용 — 업스트림 노트 인용 맥락
- `.claude/skills/moai/workflows/`, `.claude/skills/moai/team/` — 공통 인프라 레이어
- 일반 `manager-*`, `builder-agent/skill/plugin` — 프리픽스 없는 에이전트 이름

### 4.5 변환 도구 권장사항

grep 단독으로는 취약 (YAML quoted, 줄바꿈 삽입 등). **Go 정규식 기반 테스트**로 자동화:

```go
// internal/template/namespace_integrity_test.go
func TestNamespaceIntegrity_ImportedSkills(t *testing.T) {
    allowed := regexp.MustCompile(`modu-ai/moai-adk|moai/workflows|moai/team`)
    moaiPattern := regexp.MustCompile(`\bmoai\b`)

    // 대상 디렉토리 walk
    filepath.WalkDir("internal/template/templates/.claude/skills/ae-domain-brand-design",
        func(path string, d fs.DirEntry, err error) error {
            if d.IsDir() || !strings.HasSuffix(path, ".md") {
                return nil
            }
            content, _ := os.ReadFile(path)
            for _, line := range strings.Split(string(content), "\n") {
                if allowed.MatchString(line) {
                    continue
                }
                if moaiPattern.MatchString(line) {
                    t.Errorf("%s: moai reference: %s", path, line)
                }
            }
            return nil
        })
}
```

---

## Phase 5: 검증 및 머지 (Verify & Merge)

### 5.1 Go 빌드·테스트 게이트

```bash
go build ./...                   # 에러 0건
go vet ./...                     # 경고 0건
go test ./...                    # 전체 통과
go test -race ./...              # 레이스 0건
go test -cover ./internal/...    # 커버리지 목표 확인
```

### 5.2 템플릿 통합 테스트

```bash
go test ./internal/template/...

# 구체적 테스트
go test -v -run "TestNamespaceIntegrity_" ./internal/template/...
go test -v -run "TestAgencyRemoval" ./internal/template/...
```

### 5.3 grep 기반 수동 검증

```bash
# 네임스페이스 잔존 최종 확인
grep -rn "moai" internal/template/templates/.claude/rules/ae/ \
  | grep -v "modu-ai/moai-adk\|upstream"
# 결과 없거나 주석/CHANGELOG 맥락만 있어야 통과
```

### 5.4 `ae-delta.md` 갱신

SPEC-UPDATE의 마지막 Task로 `.moai/upstream/ae-delta.md`에 변경 이력 추가:

```markdown
| YYYY-MM-DD | SPEC-UPDATE-NNN | 변경 요약 |
```

추가한 고유 자산이 있으면 체크리스트 섹션에도 등록.

### 5.5 PR 생성

```bash
gh pr create \
  --title "[spec-update-NNN] moai-adk vX.Y.Z ~ vX.Y.Z 반영" \
  --body-file /tmp/pr-body.md
```

PR 본문 구조:
- 요약 (변경 버킷별 카운트)
- 영향 파일 리스트
- 테스트 결과
- 체크리스트 완료 여부

### 5.6 머지 후 후속 작업

- `project_moai_update_todo.md` 메모리 갱신
- `template_version` 반영 확인
- 사용자 프로젝트 동기화 테스트 (`ae update` 실행)

---

## 실패 시 복구

### 경로 검증 실패 (Phase 3.4)
→ spec.md 수정 후 SPEC 재승인

### 네임스페이스 잔존 (Phase 4.3 (c) 또는 Phase 5.2)
→ Go 테스트 실패 파일만 선별 Edit, 재실행

### Go 빌드 실패 (Phase 5.1)
→ expert-debug 서브에이전트 또는 /ae fix 실행

### 흡수 작업 의미 손실 발견 (Phase 4 중)
→ 흡수 전 원본 파일 임시 보존 (`git stash` 또는 사본 저장) 후 재병합

---

## 참고 자료

- [README.md](./README.md) — 개요와 철학
- [checklist.md](./checklist.md) — 반복 사용 체크리스트
- [skill-design.md](./skill-design.md) — 자동화 스킬 설계
- [SPEC-UPDATE-004-005-review.md](../specs-review/SPEC-UPDATE-004-005-review.md) — 이전 SPEC 리뷰 학습

---

Version: 1.0.0
Last Updated: 2026-04-24
