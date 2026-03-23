---
id: SPEC-REFACTOR-001
title: AE-ADK 코드베이스 정리 및 레거시 제거
status: done
created: 2026-03-23
updated: 2026-03-23
branch: spec/refactor-001
author: RealFarm
---

# SPEC-REFACTOR-001: AE-ADK 코드베이스 정리

## 개요

AE-ADK 코드베이스에서 불필요한 코드를 제거하고 유지보수성을 향상시킵니다.
총 47개 파일 변경, 24,848줄 삭제.

## 완료 기준

- [x] coverage padding 테스트 파일 제거
- [x] rank 시스템 완전 제거 (internal/rank/, internal/ralph/, CLI, hooks)
- [x] Deprecated 상수 제거
- [x] 레거시 마이그레이션 테스트 정리
- [x] `go mod tidy` 완료
- [x] 빌드 및 `go vet` 통과

## 마일스톤

### Milestone 1: coverage padding 테스트 제거 (internal/cli/)

**제거 파일 (7개):**
- `internal/cli/coverage_fixes_test.go`
- `internal/cli/coverage_improvement_test.go`
- `internal/cli/coverage_test.go`
- `internal/cli/init_coverage_test.go`
- `internal/cli/misc_coverage_test.go`
- `internal/cli/remaining_coverage_test.go`
- `internal/cli/target_coverage_test.go`

### Milestone 2: rank 시스템 완전 제거

**제거 패키지:**
- `internal/rank/` (auth, browser, client, config, device, patterns, pricing, sync_state, transcript)
- `internal/ralph/` (engine.go)

**제거 파일:**
- `internal/cli/rank.go`
- `internal/cli/rank_test.go`
- `internal/cli/rank_nonauth_test.go`
- `internal/hook/rank_session.go`
- `internal/hook/rank_session_test.go`

**변경 파일:**
- `internal/cli/deps.go`: ralph 엔진 → 인라인 `defaultDecisionEngine`
- `internal/hook/session_end.go`: rank 세션 훅 제거

### Milestone 3: Deprecated 상수 제거

**변경 파일:**
- `internal/core/project/validator.go`: `BackupTimestampFormat`, `BackupsDir` → `defs.*` 참조
- `internal/cli/wizard/types.go`: `LangNameMap` 제거

### Milestone 4: coverage_boost 테스트 제거 (전체 패키지)

**제거 파일 (7개):**
- `internal/cli/wizard/coverage_boost_test.go`
- `internal/core/git/worktree_coverage_test.go`
- `internal/core/project/coverage_extra_test.go`
- `internal/hook/coverage_boost_test.go`
- `internal/hook/lifecycle/cleanup_coverage_test.go`
- `internal/merge/confirm_coverage_test.go`
- `internal/merge/coverage_extra_test.go`

### Milestone 5: 레거시 마이그레이션 테스트 정리

**변경 파일:**
- `internal/cli/integration_test.go`: rank 관련 테스트 케이스 제거
- `internal/cli/deps_test.go`: ralph 엔진 의존성 테스트 제거

### Milestone 6: 빌드 정리

- `go mod tidy` 실행
- `go build ./...` 통과
- `go vet ./...` 통과

## 영향 범위

이 SPEC은 순수 코드 정리 작업으로, 기존 기능에는 영향 없음.
ralph 엔진은 동일 로직을 `defaultDecisionEngine`으로 인라인 유지.

## 관련 문서

- `docs/TODO.md` - 작업 항목 #1, #6 완료 처리
- `.moai/project/product.md` - 제거된 기능 섹션 업데이트
- `.moai/project/structure.md` - 패키지 구조 업데이트
