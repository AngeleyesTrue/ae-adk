# AE-ADK 패키지 구조

> 마지막 업데이트: 2026-03-23 (SPEC-REFACTOR-001 반영)

## 디렉토리 트리

```
ae-adk/
├── cmd/ae/                          # CLI 진입점 (main.go)
├── internal/
│   ├── astgrep/                     # AST 패턴 검색
│   ├── cli/                         # 명령어 핸들러
│   │   ├── wizard/                  # 프로파일 설정 마법사
│   │   └── worktree/                # worktree 관리
│   ├── cmd/                         # 내부 커맨드 유틸
│   ├── config/                      # 설정 관리
│   ├── core/
│   │   ├── git/                     # git 연산
│   │   └── project/                 # 프로젝트 메타데이터/검증
│   ├── defs/                        # 공유 상수/타입 (BackupTimestampFormat, BackupsDir 등)
│   ├── foundation/                  # 기반 유틸리티
│   ├── git/                         # git 저수준 연산
│   ├── github/                      # GitHub API 연동
│   ├── hook/                        # hook 라이프사이클
│   │   ├── agents/                  # 에이전트 hook
│   │   ├── lifecycle/               # pre/post 라이프사이클
│   │   ├── mx/                      # MX 태그 hook
│   │   ├── quality/                 # 품질 게이트 hook
│   │   └── security/                # 보안 hook
│   ├── i18n/                        # 다국어 지원
│   ├── loop/                        # loop 실행
│   ├── lsp/                         # LSP 품질 게이트
│   ├── manifest/                    # 매니페스트 관리
│   ├── merge/                       # merge 워크플로우
│   ├── profile/                     # 프로파일 관리
│   ├── resilience/                  # 재시도/서킷브레이커
│   ├── shell/                       # 셸 실행 유틸
│   ├── statusline/                  # 상태바
│   ├── template/                    # 템플릿 렌더링
│   ├── tmux/                        # tmux 통합
│   ├── ui/                          # 터미널 UI
│   ├── update/                      # 업데이트 관리
│   └── workflow/                    # 워크플로우 조율
└── pkg/                             # 공개 패키지
```

## 핵심 의존 관계

```
cmd/ae
  └── internal/cli
        ├── deps.go          (defaultDecisionEngine 인라인 구현)
        ├── hook.go          → internal/hook
        ├── init.go          → internal/core/project, internal/config
        ├── update.go        → internal/update, internal/manifest
        └── wizard/          → internal/i18n, internal/defs

internal/hook
  ├── session_end.go         → internal/lsp, internal/manifest
  ├── post_tool.go           → internal/astgrep, internal/hook/mx
  ├── pre_tool.go            → internal/hook/security
  └── quality/               → internal/lsp

internal/core/project
  └── validator.go           → internal/defs  (BackupTimestampFormat, BackupsDir)
```

## 제거된 패키지 (SPEC-REFACTOR-001)

다음 패키지는 2026-03-23 SPEC-REFACTOR-001을 통해 제거되었습니다.

| 패키지 | 제거 이유 |
|--------|-----------|
| `internal/rank/` | 외부 rank 서비스 클라이언트 - AE-ADK 핵심 범위 외 |
| `internal/ralph/` | rank 결정 엔진 - rank 패키지와 함께 삭제 |

### 이전 의존 관계 (삭제됨)

```
# 삭제 전 구조 (참고용)
internal/cli/deps.go       → internal/ralph  (결정 엔진)
internal/cli/rank.go       → internal/rank   (rank CLI 커맨드)
internal/hook/rank_session.go → internal/rank (rank 세션 추적)
```

### 대체 구현

- `internal/ralph/engine.go`의 결정 로직 → `internal/cli/deps.go`의 `defaultDecisionEngine` 함수로 인라인 통합

## 상수 관리 정책

모든 공유 상수는 `internal/defs` 패키지에서 단일 관리합니다.

- `validator.go` 등 개별 패키지에 중복 정의 금지
- `BackupTimestampFormat`, `BackupsDir` 등은 `defs.*` 참조 사용
