# AE-ADK TODO

> 마지막 업데이트: 2026-03-23

---

## 완료된 항목

### [DONE] #1 - 코드 정리 (SPEC-REFACTOR-001, 2026-03-23)

**coverage padding 테스트 제거 (Milestone 1, 4)**
- `internal/cli/coverage_fixes_test.go` 삭제
- `internal/cli/coverage_improvement_test.go` 삭제
- `internal/cli/coverage_test.go` 삭제
- `internal/cli/init_coverage_test.go` 삭제
- `internal/cli/misc_coverage_test.go` 삭제
- `internal/cli/remaining_coverage_test.go` 삭제
- `internal/cli/target_coverage_test.go` 삭제
- `internal/cli/wizard/coverage_boost_test.go` 삭제
- `internal/core/git/worktree_coverage_test.go` 삭제
- `internal/core/project/coverage_extra_test.go` 삭제
- `internal/hook/coverage_boost_test.go` 삭제
- `internal/hook/lifecycle/cleanup_coverage_test.go` 삭제
- `internal/merge/confirm_coverage_test.go` 삭제
- `internal/merge/coverage_extra_test.go` 삭제

**레거시 마이그레이션 테스트 정리 (Milestone 5)**
- `internal/cli/integration_test.go`에서 rank 관련 테스트 제거
- `internal/cli/deps_test.go`에서 ralph 엔진 의존성 테스트 제거

**Deprecated 상수 제거 (Milestone 3)**
- `internal/core/project/validator.go`: `BackupTimestampFormat`, `BackupsDir` 상수 → `defs.*` 패키지 참조로 교체
- `internal/cli/wizard/types.go`: `LangNameMap` 제거

### [DONE] #6 - rank 시스템 제거 (SPEC-REFACTOR-001, 2026-03-23)

- `internal/rank/` 패키지 전체 삭제 (auth, browser, client, config, device, patterns, pricing, sync_state, transcript 포함)
- `internal/ralph/` 패키지 전체 삭제 (engine.go)
- `internal/cli/rank.go` CLI 커맨드 삭제
- `internal/cli/rank_test.go`, `rank_nonauth_test.go` 삭제
- `internal/hook/rank_session.go`, `rank_session_test.go` 삭제
- `internal/cli/deps.go`: ralph 엔진을 인라인 `defaultDecisionEngine`으로 교체
- `internal/hook/session_end.go`: rank 세션 훅 제거
- `go mod tidy` 완료, 빌드/vet 통과 (Milestone 6)

---

## 대기 중인 항목

### #2 - 테스트 커버리지 개선

- 실제 기능 테스트 보강 (coverage padding 제거 후 공백 구간)
- 패키지별 85% 커버리지 목표 달성

### #3 - 문서 현행화

- README.md: rank 시스템 관련 언급 제거
- CHANGELOG.md: SPEC-REFACTOR-001 항목 추가

### #4 - Windows 호환성 검증

- `docs/windows-compat-ae.md` 기준 통합 테스트 재검토

### #5 - CI/CD 파이프라인 안정화

- Go 1.26 환경 기준 Actions 워크플로우 최종 검증
