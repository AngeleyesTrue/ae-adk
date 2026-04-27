# Future SPEC Proposals

> SPEC-UPDATE-004 (v2.10.2~v2.13.2 업스트림 반영) 완료 시점에 식별된 후속 SPEC 제안 모음.
> 각 제안은 "왜 필요한가 / 무엇을 달성할 것인가 / 어떤 방향으로 진행할 것인가"를 정리한다.
>
> Status: Proposal (정식 SPEC 전환 전 단계)
> Created: 2026-04-27
> Trigger: SPEC-UPDATE-004 v1.2.0 머지 직후 evaluator-active + expert-security 평가 결과

---

## 0. 개요

SPEC-UPDATE-004 머지 시점에 다음 4건이 별도 SPEC으로 분리되었다. 두 건(`SPEC-UPDATE-005`, `SPEC-UPDATE-006`)은 이미 `.moai/specs/`에 초안이 존재하며, 두 건(`SPEC-LSP-CORE-002`, `SPEC-SECURITY-BYPASS-001`)은 본 문서들이 최초의 제안서다.

| ID | 분류 | 상태 | 트리거 | 우선순위 |
|---|---|---|---|---|
| [SPEC-UPDATE-005](./SPEC-UPDATE-005.md) | 업스트림 동기화 | `.moai/specs/`에 v1.2.0 draft 존재 | SPEC-UPDATE-004 EX-01~11 이월분 (agency 일괄 삭제, /moai db 도입, design 재편) | High |
| [SPEC-UPDATE-006](./SPEC-UPDATE-006.md) | 업스트림 동기화 | `.moai/specs/`에 v1.0.0 draft 존재 | SPEC-UPDATE-004 EX-13 (v2.14.0 Utility Hardening) | High |
| [SPEC-LSP-CORE-002](./SPEC-LSP-CORE-002.md) | 신규 도입 | Proposal only (draft 없음) | SPEC-UPDATE-004 REQ-23 hold (powernap go.mod 부재) | Medium |
| [SPEC-SECURITY-BYPASS-001](./SPEC-SECURITY-BYPASS-001.md) | 보안 강화 | Proposal only (draft 없음) | SPEC-UPDATE-004 expert-security F-1 finding | Medium |

---

## 1. 제안 요약

### SPEC-UPDATE-005: Agency 일괄 삭제 + /moai db 도입 + Design 재편

- **이유**: SPEC-UPDATE-004 스코프에서 분리된 design/db/agency 재편 작업이 미반영 상태. moai-adk v2.13.x 업스트림이 이미 적용한 `/moai db` 명령, agency → moai-domain-* 흡수, `.moai/design/` 디렉토리 신설 등이 ae-adk에 누락되어 있다.
- **목표**: ae-adk를 moai-adk v2.13.2 upstream과 design/db 영역에서 동등 수준으로 정렬한다. agency 6개 에이전트 + 5개 ae-agency-* 스킬을 일괄 삭제하고 moai-domain-* 스킬로 대체.
- **방향**: 이미 `.moai/specs/SPEC-UPDATE-005/` v1.2.0 초안이 존재. 4단계 BLOCKER가 v1.2.0에서 해소됨. 머지 가능 상태.

### SPEC-UPDATE-006: v2.14.0 Utility Hardening (Detection Improvements)

- **이유**: moai-adk v2.14.0 (2026-04-24 릴리즈)는 non-breaking이지만 Detection Improvements로 기존 ae 코드베이스에 신규 위반(MX validator method receiver, paired @MX:REASON 누락)을 노출시킬 수 있다. SPEC-UPDATE-004 EX-13으로 의도적 분리됨.
- **목표**: tree-sitter 16-language complexity scanner, ast-grep 5-language rule seeding, LSP subprocess hygiene, transition_mode grace flag를 ae-adk에 이식하면서 grace 기간 동안 기존 위반을 차단하지 않는다.
- **방향**: 이미 `.moai/specs/SPEC-UPDATE-006/` v1.0.0 초안이 존재. powernap 미사용으로 인한 적용 범위 재정의 + REQ-04 분할(5 full + 11 scaffold)이 차별화 포인트.

### SPEC-LSP-CORE-002: powernap 기반 다중 언어 LSP 클라이언트 도입

- **이유**: SPEC-UPDATE-004 REQ-23이 v1.2.0에서 hold 처리됨. ae-adk는 현재 자체 `internal/lsp/` 패키지로 LSP 처리 중이지만 16개 언어를 일관되게 지원하려면 검증된 LSP 클라이언트 라이브러리가 필요하다. moai-adk v2.13.0이 채택한 `charmbracelet/x/powernap`은 crush(23k+ stars) 프로덕션에서 검증됨.
- **목표**: ae-adk가 16개 언어 모든 LSP 서버를 powernap 기반 단일 추상화로 관리하도록 마이그레이션. 기존 `internal/lsp/` 호환 인터페이스 유지하며 내부 구현만 교체.
- **방향**: 신규 SPEC 작성 필요. Path B(MCP bridge)와 차별화하는 의사결정 포함. moai-adk v2.13.2의 lsp-client.md 패턴을 ae 네임스페이스로 변환.

### SPEC-SECURITY-BYPASS-001: disableBypassPermissionsMode 정책 강제

- **이유**: SPEC-UPDATE-004 REQ-18로 settings.json.tmpl에 `disableBypassPermissionsMode: false` 필드가 추가됐지만, ae-adk Go 코드 내부에서 이 필드를 읽거나 강제하는 로직은 없다. expert-security F-1 finding(HIGH)으로 식별됨. 사용자가 `true`로 설정해도 ae-adk subagent는 여전히 `mode: "bypassPermissions"`로 spawn 가능.
- **목표**: ae-adk가 `disableBypassPermissionsMode`를 직접 honor하여 보안 정책을 코드 차원에서 강제한다. Claude Code v2.1.111+ 하니스에 의존하지 않는 다층 방어를 구축.
- **방향**: 신규 SPEC 작성 필요. `.claude/settings.json` 파서 + Agent spawn 시 mode 검증 + 회귀 테스트. CWE-732 대응.

---

## 2. 우선순위 권고

```
SPEC-UPDATE-005 ──┬─ (이미 draft 존재, 머지 가능) ──→ 다음 sprint
SPEC-UPDATE-006 ──┘

SPEC-SECURITY-BYPASS-001 ──→ SPEC-UPDATE-005 머지 후 착수
                              (settings.json 파서 코드 영향 영역 확정 필요)

SPEC-LSP-CORE-002 ──→ SPEC-UPDATE-006 머지 후 착수
                       (v2.14.0 LSP hygiene 패턴 통합)
```

각 제안 문서에 상세한 배경·목표·방향·스코프·의존성·회피 사유가 포함된다.

---

## 3. 참고 문서

- `.moai/specs/SPEC-UPDATE-004/spec.md` v1.2.0 — 본 후속 SPEC들의 발생 컨텍스트
- `.moai/upstream/ae-delta.md` — ae-adk 고유 자산 체크리스트 (각 SPEC 작업 시 필수 사전 점검)
- `.claude/rules/moai/core/lsp-client.md` — powernap 도입 결정 근거 (moai 영역 참고용)
- `docs/specs-review/SPEC-UPDATE-004-005-review.md` — SPEC-UPDATE-004/005 비판적 리뷰 보고서
