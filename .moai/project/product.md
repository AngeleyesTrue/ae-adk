# AE-ADK 제품 개요

> 마지막 업데이트: 2026-03-23

## 제품 정의

AE-ADK(AngelEyes Agent Development Kit)는 Claude Code 기반 AI 에이전트 개발 및 실행을 위한 CLI 도구입니다.

## 핵심 기능

- SPEC 기반 DDD/TDD 워크플로우
- MoAI 오케스트레이터 통합
- hook 시스템 (pre/post/subagent-stop 등)
- LSP 품질 게이트
- 멀티 언어 지원 (i18n)
- Windows/Unix 크로스 플랫폼 지원

## 패키지 구조

```
ae-adk/
├── cmd/ae/          # CLI 진입점
├── internal/
│   ├── cli/         # 명령어 핸들러 (init, run, sync, plan, ...)
│   ├── core/        # 핵심 도메인 (config, project, git, ...)
│   ├── hook/        # hook 라이프사이클 관리
│   ├── defs/        # 공유 상수/타입 정의
│   ├── i18n/        # 다국어 지원
│   ├── lsp/         # LSP 품질 게이트
│   ├── merge/       # merge 워크플로우
│   ├── loop/        # loop 실행 관리
│   └── ...
└── pkg/             # 공개 패키지
```

## 제거된 기능

### rank 시스템 (SPEC-REFACTOR-001, 2026-03-23 제거)

- **제거 이유**: AE-ADK 핵심 가치와 무관한 외부 서비스 의존성, 코드베이스 복잡도 증가
- **제거 범위**:
  - `internal/rank/` - rank 클라이언트 패키지 (인증, 브라우저, 전사록 등)
  - `internal/ralph/` - ralph 결정 엔진
  - `internal/cli/rank.go` - rank CLI 커맨드
  - `internal/hook/rank_session.go` - rank 세션 훅
- **대체**: `internal/cli/deps.go`에 인라인 `defaultDecisionEngine` 구현

### coverage padding 테스트 (SPEC-REFACTOR-001, 2026-03-23 제거)

- **제거 이유**: 실제 기능 검증 없이 커버리지 수치만 높이는 무의미한 테스트
- **제거 범위**: 14개 `*_coverage_*.go`, `*_boost_*.go` 파일
- **후속 작업**: 의미 있는 기능 테스트로 대체 예정 (#2)

### Deprecated 상수 (SPEC-REFACTOR-001, 2026-03-23 제거)

- `BackupTimestampFormat`, `BackupsDir`: `validator.go` → `defs` 패키지 참조
- `LangNameMap`: `wizard/types.go`에서 제거
