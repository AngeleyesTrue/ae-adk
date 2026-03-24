---
id: SPEC-UPDATE-001
version: "1.0.0"
status: draft
created: "2026-03-23"
updated: "2026-03-23"
author: Angeleyes
priority: medium
issue_number: 0
---

# SPEC-UPDATE-001: moai-adk 업스트림 업데이트 반영 및 update 명령어 수정

## 환경 (Environment)

### 프로젝트 컨텍스트

- ae-adk는 modu-ai/moai-adk의 포크 프로젝트
- 현재 동기화 버전: v2.7.20 (2026-03-20)
- 분석 대상 릴리즈 범위: v2.7.13 ~ v2.7.20
- ae-adk 바이너리 배포: GitHub Releases (AngeleyesTrue/ae-adk)
- 템플릿 소스: `internal/template/templates/` (Go embed으로 바이너리에 포함)

### 기존 시스템 구조

- **바이너리 업데이트**: `internal/update/` 패키지 (Checker, Updater, Rollback, Orchestrator)
- **템플릿 동기화**: `internal/cli/update.go` (`runTemplateSync`)
- **업데이트 플래그**: `--check`, `--force`, `--yes`, `--templates-only`, `--binary`, `--config`

### 업스트림 변경 요약 (v2.7.13 ~ v2.7.20)

- **v2.7.20**: Claude Code v2.1.80 통합, `moai rank` 완전 제거
- **v2.7.13~v2.7.19**: Hook 시스템 개선, StatusLine 업데이트, Skill 구조 변경

## 가정 (Assumptions)

### 기술 가정

- [A-01] moai-adk의 GitHub Releases API는 공개 접근 가능하다 (인증 불필요)
- [A-02] 업스트림 변경은 주로 템플릿 파일(`internal/template/templates/`)에 집중된다
- [A-03] Go 코드 변경은 ae-adk에서 이미 독립적으로 커스터마이징되어 직접 머지 대신 수동 분석이 필요하다
- [A-04] 현재 3-way 머지 엔진은 YAML, JSON, Markdown 등 주요 포맷을 지원한다
- [A-05] rank 기능 관련 코드는 SPEC-REFACTOR-001에서 처리 완료이므로 이 SPEC에서는 확인만 수행한다

### 비즈니스 가정

- [A-06] 업스트림 모니터링은 자동 적용이 아닌 알림 및 분석 용도이다
- [A-07] 템플릿 선택적 반영은 사용자 확인 후에만 실행된다

## 요구사항 (Requirements)

### REQ-01: 업스트림 릴리즈 변경 분석

**WHEN** 사용자가 업스트림 분석을 요청하면 **THEN** 시스템은 moai-adk의 v2.7.13~v2.7.20 릴리즈 변경 내역을 카테고리별로 분류하여 보고한다.

분류 카테고리:
- 템플릿 변경 (`.claude/`, `.moai/` 하위 파일)
- Go 코드 변경 (Hook, CLI, Core 등)
- 설정 구조 변경 (YAML 스키마, 새 설정 키)
- 제거된 기능 (rank 등)
- 신규 기능 (Claude Code v2.1.80 관련)

### REQ-02: rank 제거 반영 확인

**WHEN** 업스트림 분석이 수행되면 **THEN** 시스템은 ae-adk Go 코드에서 rank 관련 참조를 스캔하여 제거 상태를 보고한다.

### REQ-03: Claude Code v2.1.80 기능 통합 확인

**WHEN** 업스트림 분석이 수행되면 **THEN** 시스템은 Claude Code v2.1.80의 신기능이 ae-adk 템플릿에 반영되었는지 확인한다.

### REQ-04: 템플릿 비교 기능

**WHEN** 사용자가 `ae update --upstream-diff` 를 실행하면 **THEN** 시스템은 현재 임베디드 템플릿과 moai-adk 최신 릴리즈의 템플릿을 비교하여 차이점을 표시한다.

### REQ-05: 업스트림 변경 감지

**WHEN** 사용자가 `ae update --check` 를 실행하면 **THEN** 시스템은 ae-adk 릴리즈 확인과 함께 moai-adk 업스트림의 새 릴리즈도 확인하여 표시한다.

```
Current version:   ae-adk v1.0.0
Latest ae-adk:     v1.0.1 (업데이트 가능)
Upstream moai-adk: v2.7.22 (synced: v2.7.20, 2 releases behind)
```

### REQ-06: 선택적 템플릿 동기화

**WHEN** 업스트림 변경이 감지되고 사용자가 선택적 반영을 요청하면 **THEN** 시스템은 변경된 템플릿 파일 중 사용자가 선택한 파일만 업데이트한다.

### REQ-07: 업스트림 동기화 버전 추적

시스템은 **항상** 현재 동기화된 moai-adk 버전을 `.ae/config/sections/system.yaml`에 기록한다.

```yaml
ae:
  template_version: "1.0.0"
  upstream:
    synced_version: "v2.7.20"
    synced_date: "2026-03-20"
    source: "modu-ai/moai-adk"
```

### REQ-08: 안전한 업스트림 반영

시스템은 업스트림 변경을 사용자 확인 없이 자동 적용**하지 않아야 한다**.

금지 동작:
- 사용자 확인 없는 자동 템플릿 교체
- 3-way 머지 실패 시 업스트림 버전으로 강제 덮어쓰기
- 사용자 커스터마이징 (.ae/config/sections/) 자동 삭제

## 사양 (Specifications)

### S-01: 업스트림 체커 구현

파일: `internal/update/upstream.go`

```go
type UpstreamChecker interface {
    CheckLatest(ctx context.Context) (*UpstreamInfo, error)
    CompareWithCurrent(syncedVersion string) (*UpstreamDiff, error)
}

type UpstreamInfo struct {
    Version     string
    Date        time.Time
    ReleaseURL  string
    ChangeLog   string
}

type UpstreamDiff struct {
    CurrentSynced   string
    LatestAvailable string
    ReleasesBehind  int
    Changes         []ChangeEntry
}
```

### S-02: 업스트림 릴리즈 API 엔드포인트

- moai-adk 릴리즈 조회: `https://api.github.com/repos/modu-ai/moai-adk/releases`
- Rate limit 고려: 캐시 활용 (기존 `internal/update/cache.go` 패턴 재사용)
- 타임아웃: 30초

### S-03: CLI 플래그 확장

| 플래그 | 타입 | 설명 |
|--------|------|------|
| `--upstream-check` | bool | 업스트림 moai-adk 새 릴리즈 확인 |
| `--upstream-diff` | bool | 현재 템플릿과 업스트림 템플릿 비교 |
| `--upstream-sync` | bool | 선택적 업스트림 템플릿 반영 |
| `--upstream-version` | string | 특정 업스트림 버전 지정 (기본: latest) |

### S-04: 영향받는 파일 목록

**신규 파일:**
- `internal/update/upstream.go`
- `internal/update/upstream_test.go`

**수정 파일:**
- `internal/cli/update.go` - CLI 플래그 추가, 업스트림 통합
- `internal/update/types.go` - UpstreamInfo, UpstreamDiff 타입 추가
- `internal/update/cache.go` - 업스트림 캐시 지원

## 추적성 (Traceability)

| 요구사항 | TODO 항목 | 관련 파일 |
|----------|----------|----------|
| REQ-01 | TODO #2 (업스트림 분석) | internal/update/upstream.go |
| REQ-02 | TODO #2 (rank 확인) | internal/rank/, internal/cli/rank.go |
| REQ-03 | TODO #2 (Claude Code v2.1.80) | internal/template/templates/ |
| REQ-04 | TODO #10 (update 수정) | internal/cli/update.go |
| REQ-05 | TODO #10 (변경 감지) | internal/update/checker.go |
| REQ-06 | TODO #10 (선택적 반영) | internal/cli/update.go |
| REQ-07 | TODO #10 (버전 추적) | .ae/config/sections/system.yaml |
| REQ-08 | TODO #10 (안전성) | internal/cli/update.go |
