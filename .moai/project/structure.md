# AE-ADK 패키지 구조

> 마지막 업데이트: 2026-03-24

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
│   │   ├── integration/             # 통합 유틸
│   │   ├── migration/               # 마이그레이션
│   │   ├── project/                 # 프로젝트 메타데이터/검증
│   │   ├── quality/                 # 품질 검증
│   │   ├── pathutil_windows.go      # Windows 8.3 경로 변환
│   │   └── pathutil_nonwindows.go   # Unix/Mac No-op 스텁
│   ├── defs/                        # 공유 상수/타입
│   ├── foundation/                  # 기반 유틸리티
│   │   └── trust/                   # TRUST 5 프레임워크
│   ├── git/                         # git 저수준 연산
│   │   ├── convention/              # 커밋 컨벤션
│   │   └── ops/                     # git 연산
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
│   │   └── hook/                    # LSP hook 연동
│   ├── manifest/                    # 매니페스트 관리
│   ├── merge/                       # merge 워크플로우
│   ├── profile/                     # 프로파일 관리
│   ├── resilience/                  # 재시도/서킷브레이커
│   ├── shell/                       # 셸 감지 및 실행
│   ├── statusline/                  # 상태바 세그먼트
│   ├── template/                    # 템플릿 렌더링 + 임베디드 FS
│   │   └── templates/               # 프로젝트에 배포되는 템플릿 파일
│   ├── tmux/                        # tmux 통합
│   ├── ui/                          # 터미널 UI 컴포넌트
│   ├── update/                      # 바이너리 업데이트 + 릴리즈 캐시
│   └── workflow/                    # SPEC ID 파싱, 워크플로우 조율
├── docs/                            # 프로젝트 문서
├── .moai/                           # MoAI 설정/상태
│   ├── config/sections/             # YAML 설정 파일
│   ├── project/                     # 프로젝트 메타 문서
│   ├── specs/                       # SPEC 문서
│   └── design/                      # 디자인 시스템
└── .claude/                         # Claude Code 에이전트/스킬/규칙
    ├── agents/                      # 에이전트 정의
    ├── rules/                       # 규칙 파일
    ├── skills/                      # 스킬 정의 (moai-* + ae-*)
    └── hooks/                       # Hook 스크립트
```

## 핵심 의존 관계

```
cmd/ae
  └── internal/cli
        ├── deps.go          (의존성 구성 루트, defaultDecisionEngine 인라인)
        ├── root.go          (Cobra 루트 명령어, 커맨드 그룹 등록)
        ├── hook.go          → internal/hook
        ├── init.go          → internal/core/project, internal/config
        ├── update.go        → internal/update, internal/manifest
        └── wizard/          → internal/i18n, internal/defs

internal/hook
  ├── session_end.go         → internal/lsp, internal/manifest
  ├── post_tool.go           → internal/astgrep, internal/hook/mx
  ├── pre_tool.go            → internal/hook/security
  └── quality/               → internal/lsp

internal/template
  ├── settings.go            → BuildSmartPATH() (플랫폼별 PATH 구성)
  └── templates/             → Go embed FS (프로젝트 배포용 템플릿)

internal/update
  ├── checker.go             → ae-adk GitHub Releases 조회
  ├── updater.go             → 바이너리 다운로드/교체/검증
  ├── cache.go               → 릴리즈 캐시
  ├── orchestrator.go        → 업데이트 워크플로우 조율
  └── rollback.go            → 업데이트 실패 시 롤백
```

## 제거 완료 (SPEC-REFACTOR-001)

| 패키지 | 제거 이유 |
|--------|-----------|
| `internal/rank/` | 외부 rank 서비스 - ae-adk 핵심 범위 외 |
| `internal/ralph/` | rank 결정 엔진 - rank 패키지와 함께 삭제 |

## 제거 완료 (SPEC-REFACTOR-002)

| 파일 | 제거 이유 |
|------|-----------|
| `internal/cli/cc.go`, `cg.go`, `glm.go` | moai-adk 런처 - ae-adk에서 미사용 |
| `internal/cli/launcher.go` | 런처 통합 로직 - cc/glm/cg 삭제와 함께 제거 |
| `session_end.go` GLM 관련 함수 | GLM 모드 전용 정리 로직 |

## 신규 예정

| 파일/패키지 | SPEC | 용도 |
|------------|------|------|
| `internal/update/upstream.go` | SPEC-UPDATE-001 | moai-adk 업스트림 릴리즈 추적 |
| `internal/cli/platform_*.go` | SPEC-PLATFORM-001 | ae win / ae mac 명령어 |
| `internal/platform/` | SPEC-PLATFORM-001 | 플랫폼 진단/프로필 비즈니스 로직 |
| `.claude/skills/ae-lang-csharp/` | SPEC-SKILL-001 | C# 전용 스킬 (19파일) |

## 상수 관리 정책

모든 공유 상수는 `internal/defs` 패키지에서 단일 관리한다.
개별 패키지에 중복 정의 금지, `defs.*` 참조 사용.
