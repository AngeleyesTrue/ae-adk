# SPEC-UPDATE-005 요약: Agency 일괄 삭제 + /moai db 도입 + Design 재편

> Status: Draft v1.2.0 존재 (`.moai/specs/SPEC-UPDATE-005/`)
> Trigger: SPEC-UPDATE-004 EX-01 ~ EX-11 이월
> Priority: High
> Document type: Proposal-level summary (정식 SPEC은 `.moai/specs/SPEC-UPDATE-005/spec.md` 참조)

---

## 1. 배경 (Why now?)

### 1.1 발생 경위

SPEC-UPDATE-004는 moai-adk v2.10.2~v2.13.2의 **품질·성능 변경**만 반영하기로 의도적으로 스코프를 분할했다. 동일 기간 업스트림에서 발생한 다음 변경은 SPEC-UPDATE-005로 명시적으로 이월됐다.

| 이월 항목 | 출처 | SPEC-UPDATE-004 제외 사유 |
|---|---|---|
| /moai design → /ae design 통합 | EX-01 | 별도 design 재편 SPEC으로 처리 |
| /moai db → /ae db 도입 | EX-02 | DB 메타데이터 관리 별도 통합 작업 |
| `/agency` 명령 + 6개 에이전트 삭제 | EX-03 | 일괄 삭제 마이그레이션 별도 작업 |
| `moai migrate agency` 명령 이식 | EX-04 | 마이그레이션 도구 별도 작업 |
| moai-domain-brand-design, moai-domain-copywriting 흡수 | EX-05 | agency 흡수 별도 작업 |
| moai-workflow-design-import 등 4개 스킬 | EX-06 | design 재편 별도 작업 |
| v2.13.1 slash command 등록 수정 | EX-07 | db/design 관련이므로 본 SPEC에서 처리 |
| `user-invocable: false` 적용 | EX-10 | brand-design/copywriting 대상이므로 본 SPEC에서 처리 |
| `/agency (DEPRECATED)` 섹션 제거 | EX-11 | agency 자체 유지되므로 본 SPEC에서 처리 |

### 1.2 현재 상태

`.moai/specs/SPEC-UPDATE-005/` v1.2.0 초안이 존재. v1.2.0 BLOCKER 4건이 해소됨:
- Phase 2.6.5 매핑 산출물 경로 변경 (휘발성 `/tmp` → 영구 `.moai/specs/SPEC-UPDATE-005/artifacts/`)
- `namespace_integrity_test.go` 자기 제외 로직 명시 (자기참조 패러독스 해소)
- REQ-27 hook script `.sh` → `.sh.tmpl` 패턴 정렬
- REQ-29 ae-adk 자체 워크스페이스 `.claude/rules/moai/design/constitution.md` 동시 처리 명시

---

## 2. 필요성 (Rationale)

### 2.1 업스트림 정렬 비용 누적

ae-adk가 design/db/agency 영역에서 moai-adk와 분기된 채 유지되면, 매번의 업스트림 동기화(SPEC-UPDATE-NNN)마다 이 영역을 따로 확인·수정해야 한다. 누적 비용이 빠르게 증가하므로 한 번에 정렬하는 것이 합리적.

### 2.2 사용자 혼란 제거

ae-adk 사용자는 현재 다음을 마주한다:
- `/ae` 명령어 패밀리 + 별도 `/agency` 명령어 (이중 명령 체계)
- `.ae/` 설정 + `.agency/` 설정 (이중 설정 디렉토리)
- 6개의 ae-agency-* 에이전트가 ae-* 에이전트와 별도 카탈로그에 존재

업스트림 moai-adk는 이미 agency를 moai-domain-* 스킬로 흡수하여 단일 명령 체계로 통합했다. ae-adk도 동일 구조로 전환해야 사용자 학습 비용이 낮아진다.

### 2.3 /moai db 부재로 인한 실질적 결함

DB 스키마 메타데이터 관리(`.moai/project/db/schema.md`, `erd.mmd`, `migrations.md`)는 다음과 같이 활용된다:
- SPEC 작성 시 도메인 모델 참조
- @MX:ANCHOR 태그가 DB 테이블·컬럼 관계를 표현
- /ae review에서 DB 관련 변경의 일관성 검증

이 도구가 없는 ae-adk는 DB가 있는 프로젝트(예: codedaum, beee)에서 사용성이 떨어진다.

### 2.4 Design 재편의 hybrid path

업스트림 v2.13.x는 `/moai design`을 hybrid 경로(Path A: Claude Design import, Path B: code-based)로 재정의했다. ae-adk가 이를 미반영하면 design 워크플로우가 구식 패턴(.agency/ 디렉토리 의존)에 머물게 된다.

---

## 3. 목표 (Goals)

### 3.1 기능 목표

1. **Agency 일괄 삭제**: 6개 agency 에이전트(planner/builder/evaluator/learner/copywriter/designer) + 5개 ae-agency-* 스킬 + `/agency` 명령 제거
2. **moai-domain-* 스킬 흡수**: brand-design, copywriting을 ae-domain-*로 이식
3. **/ae db 도입**: DB 메타데이터 관리 명령 + `.ae/project/db/` 템플릿 + db.yaml + ae-domain-db-docs 스킬 + DB 스키마 sync 훅 + /ae project Phase 4.1a DB Detection
4. **/ae design hybrid 재정의**: Path A(Claude Design import) + Path B(code-based) 동시 지원
5. **`/agency (DEPRECATED)` 섹션 정리**: 잔존 deprecated 마커 제거
6. **마이그레이션 도구**: `ae migrate agency` 명령 — 기존 사용자가 `.agency/` 자산을 `.ae/project/brand/` 등으로 이전

### 3.2 비기능 목표

- **무중단 마이그레이션**: 기존 ae-adk 사용자의 `.agency/` 자산 자동 백업 후 변환
- **회귀 테스트 ≥85%**: namespace_integrity 검증 (moai → ae 누락 0건)
- **롤백 가이드**: 각 Phase별 롤백 지침 명시 (대규모 삭제 작업이므로)

### 3.3 비목표

- agency 기능 동등 신규 구현 (단순 흡수만)
- Path A/B 외 추가 design 경로 도입
- DB 자체 ORM 또는 마이그레이션 도구 신설 (메타데이터 관리만)

---

## 4. 방향 (Direction)

### 4.1 단계별 접근 (개략)

**Phase 1**: namespace integrity 테스트 신설 (자기 제외 로직 포함)
**Phase 2**: agency 자산 식별 + 백업 도구 작성
**Phase 3**: agency 일괄 삭제 + moai-domain-* 흡수
**Phase 4**: /ae db 명령 + 템플릿 + 훅
**Phase 5**: /ae design hybrid 재편
**Phase 6**: ae migrate agency 도구
**Phase 7**: 회귀 테스트 + 문서 정리

상세 단계는 `.moai/specs/SPEC-UPDATE-005/plan.md` 참조.

### 4.2 위험 (요약)

| 위험 | 완화 |
|---|---|
| 대규모 삭제로 인한 사용자 자산 손실 | ae migrate agency 자동 백업, 롤백 가이드 |
| moai → ae 네임스페이스 누락 | namespace_integrity_test.go (자기 제외 로직 포함) |
| /ae db 신규 도입의 기존 사용자 영향 | opt-in 방식, /ae project가 DB 자동 감지 시에만 활성화 |
| design 재편으로 기존 design 워크플로우 중단 | Path A/B 병행 지원 + 마이그레이션 가이드 |

---

## 5. 의존성 / 트리거

- **선행 머지**: SPEC-UPDATE-004 완료 (✅ 본 시점 완료됨)
- **차단**: SPEC-UPDATE-006 v2.14.0 작업과 동시 진행 가능 (파일 영역 비중첩)
- **후속 SPEC 영향**: SPEC-LSP-CORE-002, SPEC-SECURITY-BYPASS-001 모두 본 SPEC 완료 후 착수 권장

---

## 6. 다음 단계

- 정식 SPEC 문서: `.moai/specs/SPEC-UPDATE-005/spec.md` v1.2.0 (BLOCKER 해소 완료)
- 실행 명령: `/ae run SPEC-UPDATE-005`
- 예상 PR 규모: ~150~200 파일 변경 (대부분 삭제 + 일부 신규)

---

## 7. 참고 자료

- 정식 SPEC: `.moai/specs/SPEC-UPDATE-005/spec.md` (v1.2.0)
- 계획: `.moai/specs/SPEC-UPDATE-005/plan.md`
- 인수: `.moai/specs/SPEC-UPDATE-005/acceptance.md`
- 비판적 리뷰: `docs/specs-review/SPEC-UPDATE-004-005-review.md`
- 업스트림 변경 이력: `.moai/upstream/ae-delta.md`
