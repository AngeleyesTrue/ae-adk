---
spec_id: SPEC-PLATFORM-002
type: acceptance
version: "1.0.0"
---

# SPEC-PLATFORM-002 인수 조건

## 인수 조건 (Acceptance Criteria)

### AC-1: E2E 디렉토리 구조

- `tests/e2e/` 디렉토리가 존재한다
- 모든 E2E 테스트 파일에 `//go:build e2e` 빌드 태그가 포함된다
- `tests/e2e/testdata/` 디렉토리에 테스트 픽스처가 관리된다

**검증**: `ls tests/e2e/` 및 파일 헤더 확인

### AC-2: E2E/일반 테스트 분리

- `go test ./...` 실행 시 E2E 테스트가 포함되지 않는다
- `go test -tags e2e ./tests/e2e/...` 실행 시 E2E 테스트만 실행된다
- 일반 유닛 테스트와 E2E 테스트의 실행 경로가 완전히 분리된다

**검증**: 태그 없이 `go test ./...` → E2E 미실행, 태그 포함 시 E2E 실행 확인

### AC-3: git-convention 규칙 적용

- git-convention.yaml에 ae-adk 프로젝트용 커밋 타입이 정의된다
- 허용 스코프 목록이 프로젝트 구조를 반영한다
- 커밋 메시지 포맷 규칙이 명시된다

**검증**: yaml 파싱 및 최근 커밋과의 일관성 확인

### AC-4: git-strategy 브랜치 네이밍

- git-strategy.yaml에 브랜치 네이밍 규칙이 정의된다
- 머지 전략 (squash/rebase/merge) 이 명시된다
- 릴리스 플로우가 문서화된다

**검증**: yaml 파싱 및 현재 브랜치 네이밍과의 정합성 확인

### AC-5: 나노 스킬 구조

- `.claude/skills/ae-nano-*.md` 패턴으로 스킬 파일이 존재한다
- 각 스킬 파일이 500줄 이하이다
- 각 스킬 파일에 YAML frontmatter가 포함된다 (name, version, description 필수)
- frontmatter의 name 필드가 파일명과 일치한다

**검증**: `wc -l` 500줄 이하, frontmatter 파싱 성공

### AC-6: 나노 스킬 업스트림 독립성

- ae-nano-* 파일은 moai-adk 원본에 존재하지 않는다
- `moai update` 실행 시 ae-nano-* 파일에 충돌이 발생하지 않는다
- 기존 `.claude/skills/` 내 moai-adk 파일을 수정하지 않는다

**검증**: upstream remote에서 ae-nano-* 파일 부재 확인, sync 시뮬레이션

## 품질 게이트 (Quality Gates)

| 게이트 | 기준 | 적용 대상 |
|--------|------|-----------|
| 빌드 태그 분리 | E2E 테스트가 일반 빌드에 미포함 | AC-1, AC-2 |
| YAML 유효성 | git-convention, git-strategy yaml 파싱 성공 | AC-3, AC-4 |
| 파일 크기 제한 | 나노 스킬 500줄 이하 | AC-5 |
| 네이밍 규칙 | ae-nano-* 접두사 준수 | AC-5, AC-6 |
| 업스트림 안전성 | moai-adk 파일 무수정 | AC-6 |

## 완료 정의 (Definition of Done)

- [ ] AC-1 ~ AC-6 전체 통과
- [ ] 모든 품질 게이트 충족
- [ ] `go test ./...` 기존 테스트 전체 통과 (회귀 없음)
- [ ] `go test -tags e2e ./tests/e2e/...` E2E 테스트 통과
- [ ] moai-adk 원본 파일 무수정 확인
- [ ] SPEC-PLATFORM-001 완료 상태 확인 (선행 조건)
