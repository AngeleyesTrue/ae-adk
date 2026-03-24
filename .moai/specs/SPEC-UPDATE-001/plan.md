---
spec_id: SPEC-UPDATE-001
type: plan
version: "1.0.0"
---

# SPEC-UPDATE-001 구현 계획

## 개요

moai-adk 업스트림(v2.7.13~v2.7.20) 변경사항을 분석하고, `ae update` 명령어에 업스트림 모니터링 및 선택적 템플릿 동기화 기능을 추가한다.

## 마일스톤

### 마일스톤 1: 업스트림 분석 (Primary Goal)

**대상**: REQ-01, REQ-02, REQ-03

**태스크:**

- [T-01] moai-adk v2.7.13~v2.7.20 릴리즈 노트 수집 및 카테고리별 분류
- [T-02] rank 제거 상태 확인
- [T-03] Claude Code v2.1.80 기능 통합 확인
- [T-04] 템플릿 비교 분석

### 마일스톤 2: 업스트림 체커 구현 (Secondary Goal)

**대상**: REQ-05, REQ-07

**태스크:**

- [T-05] UpstreamChecker 인터페이스 및 구현
- [T-06] CLI 통합 (--upstream-check)
- [T-07] system.yaml 스키마 확장
- [T-08] UpstreamChecker 테스트

### 마일스톤 3: 선택적 템플릿 동기화 (Tertiary Goal)

**대상**: REQ-04, REQ-06

**태스크:**

- [T-09] 업스트림 템플릿 다운로드 및 비교
- [T-10] 선택적 반영 TUI (`charmbracelet/huh` 기반)
- [T-11] `--upstream-diff` 플래그 구현
- [T-12] `--upstream-sync` 플래그 구현

### 마일스톤 4: 안전성 검증 및 문서화 (Optional Goal)

**태스크:**

- [T-12.5] 안전한 업스트림 반영 검증 (REQ-08)
  - `--yes` 플래그 없이 실행 시 자동 적용이 차단되는지 단위 테스트 작성
  - 3-way 머지 충돌 시 사용자 버전 보존 검증 테스트 작성
  - 사용자 설정 (.ae/config/sections/) 미변경 검증 테스트 작성

- [T-13] 업스트림 분석 보고서 작성
- [T-14] ae update --help 문서 업데이트

## 기술 접근 방식

### 아키텍처

```
internal/update/
├── types.go          # 기존 + UpstreamInfo, UpstreamDiff (신규)
├── checker.go        # ae-adk 릴리즈 체크 (기존, 변경 없음)
├── upstream.go       # moai-adk 업스트림 체크 (신규)
├── upstream_test.go  # 업스트림 체커 테스트 (신규)
├── cache.go          # 릴리즈 캐시 (기존, 업스트림 캐시 추가)
├── orchestrator.go   # 업데이트 오케스트레이터 (기존)
├── updater.go        # 바이너리 다운로드/교체 (기존)
└── rollback.go       # 롤백 (기존)
```

### 리스크 및 대응

| 리스크 | 영향 | 대응 |
|--------|------|------|
| GitHub API rate limit 초과 | 업스트림 확인 실패 | 캐시 적용, graceful degradation |
| moai-adk 릴리즈 형식 변경 | 파싱 실패 | JSON 유연한 파싱, 에러 로깅 |
| 3-way 머지 충돌 | 사용자 파일 손실 | 백업 필수, 충돌 시 사용자 버전 보존 |

### 의존관계

```
SPEC-REFACTOR-001 (rank 제거) ← 확인만 수행, 직접 구현하지 않음
TODO #2 (업스트림 동기화) ← 마일스톤 1이 핵심
TODO #10 (update 수정) ← 마일스톤 2-3이 핵심
```

## 추적성

| 요구사항 | 마일스톤 | 태스크 |
|----------|---------|--------|
| REQ-01 | M1 | T-01 |
| REQ-02 | M1 | T-02 |
| REQ-03 | M1 | T-03 |
| REQ-04 | M3 | T-11 |
| REQ-05 | M2 | T-06 |
| REQ-06 | M3 | T-12 |
| REQ-07 | M2 | T-07 |
| REQ-08 | M4 | T-12.5 |
