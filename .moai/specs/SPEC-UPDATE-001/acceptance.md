---
spec_id: SPEC-UPDATE-001
type: acceptance
version: "1.0.0"
---

# SPEC-UPDATE-001 인수 기준

## 테스트 시나리오

### AC-01: 업스트림 릴리즈 분석 (REQ-01)

```gherkin
Given moai-adk의 v2.7.13~v2.7.20 릴리즈가 GitHub에 공개되어 있고
When 업스트림 분석을 수행하면
Then 릴리즈별 변경 내역이 카테고리(템플릿/Go코드/설정/기능제거/신규기능)로 분류되어야 한다
And 분석 보고서가 .moai/reports/ 디렉토리에 저장되어야 한다
```

### AC-02: rank 제거 상태 확인 (REQ-02)

```gherkin
Given ae-adk 코드베이스가 존재하고
When rank 제거 상태 확인을 수행하면
Then internal/rank/, internal/ralph/, internal/cli/rank.go 등의 존재 여부가 보고되어야 한다
And SPEC-REFACTOR-001과의 의존관계가 명시되어야 한다
```

### AC-03: Claude Code v2.1.80 통합 확인 (REQ-03)

```gherkin
Given moai-adk v2.7.20이 Claude Code v2.1.80 기능을 통합했고
When Claude Code 기능 통합 확인을 수행하면
Then Hook/StatusLine/Skill/에이전트 영역의 반영 상태가 보고되어야 한다
And 미반영 항목 각각에 대해 구체적 액션 아이템(적용/무시/보류)과 사유가 1줄 이상 제시되어야 한다
```

### AC-04: 업스트림 템플릿 비교 (REQ-04)

```gherkin
Given ae-adk 임베디드 템플릿과 moai-adk 최신 릴리즈 템플릿이 접근 가능하면
When `ae update --upstream-diff`를 실행하면
Then 추가/삭제/변경 파일 목록과 Risk level(High/Medium/Low)이 표시되어야 한다
And ae-adk 전용 파일(ae-* 프리픽스)은 비교에서 제외되어야 한다
```

### AC-05: 업스트림 변경 감지 (REQ-05)

```gherkin
Given system.yaml에 upstream.synced_version이 설정되어 있고
When `ae update --check`를 실행하면
Then ae-adk 버전과 함께 업스트림 moai-adk 버전 정보가 표시되어야 한다
```

```gherkin
Given GitHub API에 접근할 수 없는 환경이면
When `ae update --check`를 실행하면
Then ae-adk 릴리즈 확인은 정상 수행되고 업스트림 확인 실패는 경고로만 표시되어야 한다
```

### AC-06: 선택적 템플릿 동기화 (REQ-06)

```gherkin
Given 업스트림에서 파일이 변경되었고
When `ae update --upstream-sync`를 실행하면
Then 변경 파일이 카테고리별로 그룹화되어 표시되어야 한다
And 사용자가 선택한 파일만 3-way 머지가 적용되어야 한다
And 동기화 전 백업이 생성되어야 한다
```

### AC-07: 업스트림 동기화 버전 추적 (REQ-07)

```gherkin
Given 업스트림 동기화가 성공적으로 완료되면
When system.yaml을 읽으면
Then upstream.synced_version과 synced_date가 업데이트되어야 한다
```

### AC-08: 안전한 업스트림 반영 (REQ-08)

```gherkin
Given 업스트림 변경이 감지되었고
When ae update를 --yes 플래그 없이 실행하면
Then 업스트림 변경은 자동 적용되지 않고 사용자 확인 프롬프트가 표시되어야 한다
```

```gherkin
Given 3-way 머지 충돌이 발생하면
When 선택적 동기화를 실행하면
Then 사용자 버전이 보존되고 충돌 파일에 대한 경고가 표시되어야 한다
```

## 품질 게이트

- [ ] `go vet ./internal/update/...` 통과
- [ ] `golangci-lint` 경고 0건
- [ ] `go test -race ./internal/update/...` 통과
- [ ] UpstreamChecker 테스트 커버리지 85% 이상

## 완료 기준 (Definition of Done)

### 마일스톤 1
- [ ] v2.7.13~v2.7.20 변경 분석 보고서 작성 완료
- [ ] rank 코드 존재 확인 및 의존관계 문서화

### 마일스톤 2
- [ ] `internal/update/upstream.go` 구현 완료
- [ ] `ae update --check`에서 업스트림 버전 표시
- [ ] `system.yaml`에 upstream 섹션 추가

### 마일스톤 3
- [ ] `ae update --upstream-diff` 및 `--upstream-sync` 구현 완료

### 마일스톤 4
- [ ] REQ-08 안전성 검증 테스트 작성 완료
- [ ] `ae update --help` 문서 업데이트
