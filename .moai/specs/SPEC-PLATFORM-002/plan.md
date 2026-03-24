---
spec_id: SPEC-PLATFORM-002
type: plan
version: "1.0.0"
---

# SPEC-PLATFORM-002 구현 계획

## 마일스톤 개요

| # | 마일스톤 | 우선순위 | TODO | 예상 규모 |
|---|----------|----------|------|-----------|
| M1 | E2E 테스트 프레임워크 | Primary | #7 | Medium |
| M2 | Git 규칙 커스터마이징 | Secondary | #8 | Small |
| M3 | 나노 바나나 스킬 | Optional | #9 | Small |

## M1: E2E 테스트 프레임워크 (Primary)

### 태스크

1. **T1-1**: `tests/e2e/` 디렉토리 및 `//go:build e2e` 기반 구조 생성
2. **T1-2**: E2E 공통 헬퍼 작성 (바이너리 빌드, 실행, 출력 캡처)
3. **T1-3**: Claude in Chrome MCP 연동 테스트 러너 검토
4. **T1-4**: update 커맨드 E2E 테스트 작성
5. **T1-5**: sync 커맨드 E2E 테스트 작성
6. **T1-6**: CI 파이프라인에 E2E 스텝 추가 (`go test -tags e2e`)

### 리스크

- **R1-1**: Claude in Chrome MCP 서버의 CI 환경 가용성 → 로컬 전용 fallback 필요
- **R1-2**: E2E 테스트 실행 시간 증가 → 빌드 태그로 분리하여 선택적 실행

## M2: Git 규칙 커스터마이징 (Secondary)

### 태스크

1. **T2-1**: git-convention.yaml 커스터마이징 (커밋 타입, 스코프, 포맷)
2. **T2-2**: git-strategy.yaml 커스터마이징 (브랜치 네이밍, 머지 전략)
3. **T2-3**: 기존 커밋 이력과의 일관성 검증

### 리스크

- **R2-1**: 기존 커밋과 새 규칙의 불일치 → 신규 커밋부터 적용, 소급 미적용

## M3: 나노 바나나 스킬 (Optional)

### 태스크

1. **T3-1**: 나노 스킬 구조 표준 정의 (frontmatter 스키마, 500줄 제한)
2. **T3-2**: ae-nano-snippet 스킬 작성
3. **T3-3**: ae-nano-scaffold 스킬 작성
4. **T3-4**: ae-nano-platform 스킬 작성 (선택)
5. **T3-5**: ae-nano-debug 스킬 작성 (선택)

### 리스크

- **R3-1**: 업스트림 sync 시 `.claude/skills/` 충돌 → ae-nano-* 접두사로 방지
- **R3-2**: 스킬 비대화 → 500줄 제한 AC로 강제

## 구현 순서

```
M1 (E2E) ──→ M2 (Git 규칙) ──→ M3 (나노 스킬)
  Primary       Secondary         Optional
```

- M1은 다른 기능의 검증 인프라이므로 최우선 구현
- M2는 M1과 독립적이나, E2E 인프라 확보 후 진행이 효율적
- M3는 완전 독립적이며 필요 시 병렬 진행 가능

## 의존성

- SPEC-PLATFORM-001 완료 후 착수 (플랫폼 코어 기능 전제)
- M1~M3 사이에는 강한 의존성 없음 (순서는 권장사항)
