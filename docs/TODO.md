# AE-ADK TODO

> 마지막 업데이트: 2026-03-24

---

## 완료된 항목

### [DONE] #1 - 코드 정리 (SPEC-REFACTOR-001, 2026-03-23)

coverage padding 테스트 14개 파일 삭제, deprecated 상수 제거, 레거시 마이그레이션 테스트 정리.

### [DONE] #6 - rank 시스템 제거 (SPEC-REFACTOR-001, 2026-03-23)

`internal/rank/`, `internal/ralph/` 패키지 및 CLI/Hook 연동 전체 삭제. 47개 파일 변경, 24,848줄 삭제.

---

### [DONE] #3 - cc/glm/cg 런치 명령어 삭제 (SPEC-REFACTOR-002, 2026-03-24)

- `ae cc`, `ae glm`, `ae cg` 소스 파일 3개 + 테스트 6개 삭제
- `launcher.go` 전체 삭제
- `session_end.go` GLM 정리 로직 제거
- `root.go` help 텍스트 및 launch 그룹 제거
- `refactor_002_test.go` 62개 검증 테스트 추가

---

## 진행 대기 (SPEC 작성 완료)

### #2 - moai-adk 업스트림 변경 분석 → SPEC-UPDATE-001

- v2.7.13~v2.7.20 릴리즈 변경사항 카테고리별 분석
- rank 제거 반영 확인, Claude Code v2.1.80 기능 통합 확인
- 템플릿 비교 분석

### #10 - update 명령어 업스트림 모니터링 기능 → SPEC-UPDATE-001

- `ae update --upstream-check`: moai-adk 새 릴리즈 확인
- `ae update --upstream-diff`: 템플릿 비교
- `ae update --upstream-sync`: 선택적 템플릿 동기화
- UpstreamChecker 구현 (`internal/update/upstream.go`)
- system.yaml에 동기화 버전 추적

### #4 - ae win / ae mac 플랫폼 전환 명령어 → SPEC-PLATFORM-001

- 플랫폼 자동 감지 및 settings.json PATH 재구성
- settings.json 백업 시스템 (타임스탬프, 5개 로테이션)
- Windows 진단: UTF-8, MCP 경로, Git Bash, WSL2, 260자 제한, Hook bash
- macOS 진단: Homebrew (Intel/Apple Silicon), 심링크, 셸 호환성
- 플랫폼 프로필 저장 (`~/.ae/platform-profile.json`)

### #7 - E2E 테스트 프레임워크 → SPEC-PLATFORM-002

- `tests/e2e/` 디렉토리, `//go:build e2e` 빌드 태그
- Claude in Chrome MCP 연동 브라우저 검증

### #8 - 커밋/브랜치 규칙 커스터마이징 → SPEC-PLATFORM-002

- `git-convention.yaml` 최적화 (타입, 스코프, 한글 커밋)
- `git-strategy.yaml` 최적화 (브랜치 네이밍, mode)

### #9 - 나노 바나나 스킬 → SPEC-PLATFORM-002

- `.claude/skills/ae-nano-*` 패턴, 500줄 이하
- 후보: snippet, scaffold, platform, debug

### #5 - C# 스킬 고도화 → SPEC-SKILL-001

- `ae-lang-csharp` 스킬 생성 (16 모듈 + 2 references)
- 금지 패키지 제거: MediatR → Wolverine 3.x, AutoMapper → Mapster 7.x, FluentAssertions → AwesomeAssertions
- spk-dotnet-guide 5개 모듈 마이그레이션 + 11개 신규 모듈

---

## 미지정 (SPEC 없음)

### #11 - 테스트 커버리지 개선

- coverage padding 제거 후 공백 구간 보강
- 패키지별 85% 커버리지 목표

### #12 - CI/CD 파이프라인 안정화

- Go 1.26 환경 기준 Actions 워크플로우 최종 검증

### #13 - Claude Code 신기능 모니터링 체계

- Claude Code 릴리즈 노트 추적 자동화
- moai-adk 업스트림과 독립적으로 신기능 반영 가능한 구조 확보
