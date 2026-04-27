# SPEC-UPDATE-006 요약: moai-adk v2.14.0 Utility Hardening 반영

> Status: Draft v1.0.0 존재 (`.moai/specs/SPEC-UPDATE-006/`)
> Trigger: SPEC-UPDATE-004 EX-13 (v2.14.0 이월)
> Priority: High
> Document type: Proposal-level summary (정식 SPEC은 `.moai/specs/SPEC-UPDATE-006/spec.md` 참조)

---

## 1. 배경 (Why now?)

### 1.1 발생 경위

moai-adk v2.14.0 (2026-04-24 릴리즈)는 **Detection Improvements** 중심의 utility hardening 릴리즈다. SPEC-UPDATE-004 v1.2.0 검토 시 다음 사실이 확인됐다.

- v2.14.0의 변경은 non-breaking이지만, Detection Improvements 특성상 **기존 ae-adk 코드베이스에 신규 위반을 노출**할 수 있다
- 특히 MX validator 정확성 개선(method receiver detection, paired @MX:REASON 검증)은 기존 ae 코드에 잠재된 위반을 새롭게 가시화함
- 이 변경을 SPEC-UPDATE-004와 함께 한 PR에 묶으면 품질·성능 작업과 detection 개선이 혼재되어 검토 비용이 증가

따라서 v2.14.0은 SPEC-UPDATE-004 EX-13으로 **의도적으로 분리**되어 SPEC-UPDATE-006으로 단일 릴리즈 SPEC을 신설했다.

### 1.2 현재 상태

`.moai/specs/SPEC-UPDATE-006/` v1.0.0 초안이 존재. 주요 차별화 포인트:
- REQ-04 tree-sitter 16-language complexity를 5 full + 11 scaffold로 분할 (REQ-04a, REQ-04b)
- REQ-07/REQ-08 LSP stderr drain, singleflight를 powernap 미사용 사정에 맞게 조건부 처리 (ae-adk 자체 internal/lsp 사정 반영)
- REQ-11 transition_mode 일시 적용 + 위반 수정 별도 SPEC 권고

---

## 2. 필요성 (Rationale)

### 2.1 업스트림 동기화 누적 차단

v2.14.0이 미반영 상태로 시간이 누적되면 v2.15.x, v2.16.x 변경과 함께 한 번에 처리해야 하는데, Detection Improvements 특성상 한 번에 묶을수록 신규 위반 폭증으로 작업 마비 위험이 커진다. 단일 릴리즈 단위로 처리하는 것이 위험을 분산하는 합리적 전략.

### 2.2 MX validator 정확성 결함의 누적 손실

v2.14.0의 method receiver detection 개선이 적용되지 않은 상태에서:
- @MX:ANCHOR fan_in 계산이 부정확 (메서드 리시버 미인식 → 호출 횟수 과소 추정)
- @MX:WARN paired @MX:REASON 검증 부재 → 위험 코드의 이유 누락이 발견되지 않음
- 결과: ae 코드베이스의 코드 컨텍스트 품질이 천천히 저하

### 2.3 LSP subprocess hygiene의 안정성

v2.14.0의 stderr drain + singleflight.Group 도입은 다음을 해결한다:
- LSP 서버 stderr 미배출 시 PIPE 버퍼 가득 차서 프로세스 hang
- 동일 LSP 요청 다중 발생 시 race condition

ae-adk가 자체 internal/lsp를 가지므로 powernap 미사용 사정에 맞게 조건부 적용해야 함.

### 2.4 ast-grep rule seeding의 실용성

v2.14.0이 5개 언어(Go/TypeScript/Python/Rust/Java)에 대해 ast-grep security rule seed를 제공한다. 이 seed는 OWASP Top 10 패턴 매칭에 즉시 사용 가능 — ae-adk가 이를 흡수하면 /ae review와 expert-security 에이전트의 효과가 즉시 향상된다.

---

## 3. 목표 (Goals)

### 3.1 기능 목표

1. **MX validator 정확성**: method receiver detection, @MX:REASON pairing 검증, Windows-native 구현
2. **tree-sitter 16-language 복잡도 측정**: 5 full + 11 scaffold (REQ-04a/04b)
3. **ast-grep 5-language rule seeding**: Go/TypeScript/Python/Rust/Java + suppression policy
4. **LSP subprocess hygiene** (조건부): ae-adk internal/lsp 사정에 맞춰 stderr drain + singleflight 적용
5. **transition_mode grace flag**: v2.14.0 즉시 적용 시 기존 위반 노출 차단 메커니즘
6. **go-tree-sitter 의존성 추가**: ~50 MiB 바이너리 크기 영향 (의도적 수용)

### 3.2 비기능 목표

- **transition_mode 활성 기간 명시**: 일정 기간 후 자동 비활성 (또는 별도 violation-fix SPEC 머지까지)
- **위반 수정 SPEC 분리**: 본 SPEC은 detection 개선만, 실제 위반 fix는 violation-fix SPEC으로 분리

### 3.3 비목표

- 기존 위반의 자동 수정 (검출만, 수정은 별도 SPEC)
- transition_mode를 영구 기본값으로 유지 (임시 grace만)
- 비ae 네임스페이스 코드(예: vendor/, _generated.go)에 대한 적용 (제외 대상 유지)

---

## 4. 방향 (Direction)

### 4.1 단계별 접근 (개략)

**Phase 1**: tree-sitter 16-language complexity scanner (REQ-04a 5 full)
**Phase 2**: tree-sitter scaffold 11언어 (REQ-04b 점진 적용 가능)
**Phase 3**: MX validator 정확성 개선 (method receiver, @MX:REASON pairing, Windows-native)
**Phase 4**: ast-grep rule seeding 5-language + suppression policy
**Phase 5**: LSP hygiene 조건부 적용 (ae internal/lsp 사정)
**Phase 6**: transition_mode 도입 + 회귀 테스트
**Phase 7**: 위반 분포 측정 + violation-fix SPEC 초안 권고서

상세 단계는 `.moai/specs/SPEC-UPDATE-006/plan.md` 참조.

### 4.2 위험 (요약)

| 위험 | 완화 |
|---|---|
| Detection 개선으로 기존 위반 대량 노출 → CI 차단 | transition_mode grace 일시 적용 (REQ-11) |
| go-tree-sitter ~50 MiB 바이너리 영향 | 의도적 수용, build 옵션으로 조건부 포함 검토 |
| ae-adk internal/lsp가 powernap 미사용으로 v2.14.0 hygiene 패턴 호환 어려움 | 조건부 처리 (REQ-07/08 재정의) |
| transition_mode 영구화 위험 | 시간 제한 + violation-fix SPEC 추적 |

---

## 5. 의존성 / 트리거

- **선행 머지**: SPEC-UPDATE-004 완료 (✅ 본 시점 완료됨)
- **병행 가능**: SPEC-UPDATE-005 (파일 영역 비중첩)
- **후속 SPEC**: 본 SPEC 머지 직후 violation-fix SPEC 작성 권고
- **SPEC-LSP-CORE-002와의 관계**: powernap 도입 시 본 SPEC의 LSP hygiene 패턴이 자연스럽게 흡수되므로, 본 SPEC을 SPEC-LSP-CORE-002 선행 머지 권장

---

## 6. 다음 단계

- 정식 SPEC 문서: `.moai/specs/SPEC-UPDATE-006/spec.md` v1.0.0
- 실행 명령: `/ae run SPEC-UPDATE-006` (단, SPEC-UPDATE-005 완료 권고)
- 예상 PR 규모: ~80~120 파일 변경 (검출 도구 + 룰 + 회귀 테스트)
- 후속: violation-fix SPEC을 별도 issue/SPEC으로 추적

---

## 7. 참고 자료

- 정식 SPEC: `.moai/specs/SPEC-UPDATE-006/spec.md` (v1.0.0)
- 계획: `.moai/specs/SPEC-UPDATE-006/plan.md`
- 인수: `.moai/specs/SPEC-UPDATE-006/acceptance.md`
- moai-adk v2.14.0 릴리즈 노트 (업스트림)
- SPEC-UPDATE-004 EX-13 (v2.14.0 이월 명시 및 사유)
- 본 디렉토리 [SPEC-LSP-CORE-002.md](./SPEC-LSP-CORE-002.md) (powernap 도입 시 본 SPEC LSP hygiene 흡수 관계)
