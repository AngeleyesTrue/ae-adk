# ae-adk 업스트림 동기화 가이드

> ae-adk는 `modu-ai/moai-adk` 업스트림 기반 포크 프로젝트다.
> 이 디렉토리는 업스트림 변경을 감지·분석·반영하는 반복 작업의 표준을 정의한다.

---

## 문서 목록

| 문서 | 역할 |
|---|---|
| [README.md](./README.md) | 전체 개요·철학·용어 (본 문서) |
| [process-guide.md](./process-guide.md) | 5단계 동기화 프로세스 상세 |
| [checklist.md](./checklist.md) | 반복 사용 체크리스트 템플릿 |
| [skill-design.md](./skill-design.md) | `ae-upstream-sync` 스킬 설계 제안서 |

---

## 1. 핵심 원칙

### 1.1 소스 관리 경계

```
modu-ai/moai-adk (업스트림)
        │
        │  moai-adk가 ae-adk를 만드는 도구
        ▼
AngeleyesTrue/ae-adk (본 프로젝트)
        │
        ├── .claude/, .moai/, CLAUDE.md   ← moai-adk 도구 영역
        │                                    (직접 수정 금지)
        │
        ├── internal/template/templates/   ← ae-adk 임베디드 템플릿
        │                                    (SPEC-UPDATE 대상)
        │
        └── cmd/, internal/, pkg/          ← ae-adk Go 코드 (개발 대상)
```

두 개의 구분된 영역이 있다.
- **상단 (moai-adk 도구 영역)**: 이 프로젝트에서 Claude Code를 구동시키는 도구. `moai update`로 갱신.
- **하단 (ae-adk 제품 영역)**: 사용자 프로젝트에 배포되는 임베디드 템플릿 + CLI. **SPEC-UPDATE가 수정하는 곳**.

### 1.2 "moai-adk가 ae-adk를 만들고 있다"

이 문장은 모든 업스트림 작업의 나침반이다.

- ae-adk 코드 작업 중 `.claude/`, `.moai/`, `CLAUDE.md` 영역을 수정하면 도구 자체가 변질된다.
- 따라서 업스트림 동기화는 오직 `internal/template/templates/` (임베디드 템플릿)과 `internal/`, `cmd/` (Go 코드)만을 수정 범위로 삼는다.

### 1.3 `ae-delta.md`는 Single Source of Truth

`.moai/upstream/ae-delta.md`가 ae-adk 고유 자산의 권위 있는 목록이다.

- 업스트림에 없는 ae-adk 전용 기능(ae 플랫폼 명령, ae-design-impeccable, bracket-scope 커밋 등)은 전부 등록됨.
- SPEC-UPDATE 착수 **Phase 0**에서 반드시 확인해야 한다.
- 새 고유 기능을 추가할 때마다 같이 갱신한다.

---

## 2. 용어 정의

| 용어 | 정의 |
|---|---|
| **업스트림 (upstream)** | `modu-ai/moai-adk` 원본 리포지토리 |
| **업스트림 릴리즈** | GitHub Releases 페이지의 태그된 버전 (예: v2.14.0) |
| **Delta** | 이전 반영 버전 이후 업스트림의 파일·구조 변경 집합 |
| **네임스페이스 변환** | `moai-xxx` → `ae-xxx`, `.moai/` → `.ae/`, `<moai>` → `<ae>` 등 일괄 치환 |
| **예외 경로** | 네임스페이스 변환을 적용하지 않는 허용 구역 (예: `modu-ai/moai-adk`) |
| **고유 자산 (ae-only asset)** | 업스트림에 대응이 없는 ae-adk 전용 코드/스킬 |
| **보존 대상 (preserve target)** | 업스트림에서 삭제/변경돼도 ae-adk에서 유지해야 하는 항목 |
| **이식 (port)** | 업스트림 파일을 ae 네임스페이스로 변환하여 템플릿에 복사 |
| **흡수 (absorb)** | 기존 ae 자산의 내용을 새 업스트림 자산에 병합 |

---

## 3. 업스트림 반영 SPEC 시리즈

| SPEC | 스코프 | 상태 |
|---|---|---|
| SPEC-UPDATE-001 | v2.7.13 ~ v2.8.3 | 완료 (PR #14) |
| SPEC-UPDATE-002 | v2.8.4 ~ v2.9.1 | 완료 (PR #18) |
| SPEC-UPDATE-003 | v2.9.1 ~ v2.10.1 | 완료 (PR #25) |
| SPEC-UPDATE-004 | v2.10.2 ~ v2.13.2 (품질·성능) | draft |
| SPEC-UPDATE-005 | v2.13.x (Design+DB 재편) | draft |
| SPEC-UPDATE-006 | v2.14.0+ (Utility Hardening) | 미착수 (예정) |

---

## 4. 동기화 주기

업스트림은 주당 1-3회 릴리즈 페이스. 수동 추적은 비효율적이므로 다음 원칙을 따른다.

- **모니터링**: 주 1회 `gh api` 호출로 신규 릴리즈 확인 (자동화는 `skill-design.md` 참조)
- **일괄 반영**: 연관된 3-5개 마이너/패치를 한 SPEC으로 묶어 반영
- **긴급 반영**: 보안 패치는 단독 SPEC-SECURITY-PATCH-NNN으로 즉시 처리
- **대규모 변경**: 구조 변경(예: Agency 재편, DB 신설)은 단독 SPEC으로 분리

---

## 5. 자주 놓치는 포인트

1. **템플릿 vs 프로젝트 루트 혼동**
   - 수정 대상: `internal/template/templates/.claude/*`, `internal/template/templates/.ae/*`
   - **금지**: 프로젝트 루트의 `.claude/*`, `.moai/*` (이들은 `ae update`로 갱신됨)

2. **system.yaml.tmpl vs system.yaml**
   - 템플릿 소스는 `.tmpl` 확장자가 붙음 (Go embed + text/template 렌더링)
   - SPEC 작성 시 파일명을 반드시 실제 레이아웃으로 맞출 것

3. **네임스페이스 예외의 근거**
   - `.claude/skills/moai/workflows/`, `.claude/skills/moai/team/`는 moai 네임스페이스 유지
   - 이유: 워크플로우/팀 스킬은 ae-adk와 moai-adk 간 공통 인프라 레이어로 취급

4. **code_comments 정책**
   - `.ae/config/sections/language.yaml`의 `code_comments: ko`가 기본값
   - 업스트림 이식 시 영어 주석을 한국어로 강제 변환하지 않는다 (사용자 프로젝트 설정 존중)

5. **ae-adk 고유 기능**
   - 업스트림에서 해당 파일이 삭제됐다고 무조건 따라 삭제하면 안 됨
   - `ae-delta.md` 확인 후 보존/삭제 별도 판단

---

## 6. 빠른 시작

신규 릴리즈 확인부터 SPEC 초안까지:

```bash
# 1. 업스트림 최신 릴리즈 확인
gh api repos/modu-ai/moai-adk/releases --jq '.[0:5] | .[] | "\(.tag_name) | \(.published_at)"'

# 2. 현재 동기화 버전 확인
grep template_version internal/template/templates/.ae/config/sections/system.yaml.tmpl

# 3. delta 범위 결정 (현재 동기화 버전 ~ 최신)
# → SPEC-UPDATE-NNN 작성 시작
```

자세한 단계는 [process-guide.md](./process-guide.md), 체크리스트는 [checklist.md](./checklist.md)를 참조한다.

---

Version: 1.0.0
Last Updated: 2026-04-24
Maintainer: Angeleyes
