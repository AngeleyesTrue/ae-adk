---
id: SPEC-PLATFORM-002
version: "1.0.0"
status: draft
created: "2026-03-23"
updated: "2026-03-23"
author: Angeleyes
priority: medium
issue_number: 0
depends_on: SPEC-PLATFORM-001
---

# SPEC-PLATFORM-002: E2E 테스트, Git 규칙 커스터마이징, 나노 바나나 스킬

## 1. 환경 (Environment)

- ae-adk, Go 1.26, Cobra CLI
- 선행 SPEC: SPEC-PLATFORM-001
- 분리 배경: 핵심 플랫폼 기능과 독립적 구현 가능한 부가 기능 분리
- TODO #7 (E2E), #8 (Git 규칙), #9 (나노 스킬)
- 제약: moai-adk 파일 수정 금지 (ae-nano-* 신규 추가만 허용)

## 2. 가정 (Assumptions)

- Go 빌드 태그로 E2E 분리 가능
- Claude in Chrome MCP 서버 활용 가능
- git-convention.yaml, git-strategy.yaml 이미 존재
- ae-nano-* 패턴으로 업스트림 충돌 없음

## 3. 요구사항 (Requirements)

### REQ-001: E2E 테스트 프레임워크 (State-Driven)

Claude in Chrome MCP 서버를 활용한 E2E 테스트 프레임워크를 구축한다.
실제 CLI 바이너리를 빌드하고 실행하여 전체 파이프라인을 검증한다.

### REQ-002: E2E 테스트 구조 (Ubiquitous)

- 디렉토리: `tests/e2e/`
- 빌드 태그: `//go:build e2e`
- 일반 `go test`에서 E2E 테스트가 실행되지 않아야 한다
- E2E 실행: `go test -tags e2e ./tests/e2e/...`

### REQ-003: git-convention.yaml 커스터마이징 (State-Driven)

git-convention.yaml 파일을 ae-adk 프로젝트에 맞게 커스터마이징한다.
커밋 메시지 포맷, 허용 타입, 스코프 규칙 등을 정의한다.

### REQ-004: git-strategy.yaml 커스터마이징 (State-Driven)

git-strategy.yaml 파일을 ae-adk 프로젝트에 맞게 커스터마이징한다.
브랜치 네이밍, 머지 전략, 릴리스 플로우 등을 정의한다.

### REQ-005: 나노 스킬 컬렉션 구조 (Ubiquitous)

- 디렉토리: `.claude/skills/ae-nano-*`
- 각 스킬은 500줄 이하로 유지
- YAML frontmatter로 메타데이터 관리
- 업스트림 sync 시 충돌 없는 네이밍 (ae-nano-* 접두사)

### REQ-006: 초기 나노 스킬 후보 (Optional)

다음 나노 스킬 후보를 검토하고 필요에 따라 구현한다:

- **ae-nano-snippet**: 자주 쓰는 코드 조각 빠른 삽입
- **ae-nano-scaffold**: 파일/디렉토리 템플릿 생성
- **ae-nano-platform**: 플랫폼 상태 조회/진단
- **ae-nano-debug**: 디버그 정보 수집/출력

## 4. 사양 (Specifications)

### 4.1 E2E 디렉토리 구조

```
tests/
  e2e/
    e2e_test.go          # //go:build e2e, 공통 헬퍼
    update_test.go       # update 커맨드 E2E
    sync_test.go         # sync 커맨드 E2E
    testdata/            # 테스트 픽스처
```

### 4.2 나노 스킬 디렉토리 구조

```
.claude/
  skills/
    ae-nano-snippet.md
    ae-nano-scaffold.md
    ae-nano-platform.md
    ae-nano-debug.md
```

### 4.3 금지 수정 영역

다음 파일/디렉토리는 moai-adk 원본이므로 수정 금지:

- `.claude/` 내 ae-nano-* 외 기존 파일
- `.moai/` 내 spec 외 기존 파일
- `CLAUDE.md`

## 5. 추적성 (Traceability)

| 요구사항 | TODO | 우선순위 |
|----------|------|----------|
| REQ-001, REQ-002 | TODO #7 (E2E 테스트) | Primary |
| REQ-003, REQ-004 | TODO #8 (Git 규칙) | Secondary |
| REQ-005, REQ-006 | TODO #9 (나노 스킬) | Optional |
