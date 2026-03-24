# AE-ADK 제품 개요

> 마지막 업데이트: 2026-03-24

## 제품 비전

AE-ADK(AngelEyes Agent Development Kit)는 moai-adk를 포크하여 **Angeleyes 1인 개발자에게 완전히 최적화된** Claude Code 개발 하네스이다.

핵심 철학:

1. **개인 최적화**: 범용 도구가 아닌, Angeleyes의 기술 스택(C#/Wolverine/Mapster)과 작업 환경(Windows+macOS)에 정확히 맞춘 도구
2. **업스트림 추적**: moai-adk의 버전업을 지속 추적하고, 유용한 변경만 선별적으로 반영
3. **독립 진화**: moai-adk 업데이트가 중단되더라도 Claude Code 신기능을 독립적으로 모니터링하고 최신화
4. **경량 유지**: 사용하지 않는 기능(rank 시스템, cc/glm/cg 런처 등)을 적극적으로 제거하여 코드베이스를 간결하게 유지

## 사용자

- **Angeleyes (RealFarm)**: 1인 개발자
- **주 환경**: Windows 11 (70%) + macOS (30%), 빈번한 플랫폼 전환
- **기술 스택**: Go (ae-adk 자체), C# .NET 10 (프로덕션 프로젝트), Claude Code

## 핵심 기능

### 현재 동작 중

- **SPEC 기반 DDD/TDD 워크플로우**: /moai plan → run → sync 파이프라인
- **MoAI 오케스트레이터**: 전문 에이전트 위임 기반 작업 처리
- **Hook 시스템**: pre/post tool, subagent-stop, session-end 등 라이프사이클 관리
- **LSP 품질 게이트**: 빌드/린트/테스트 자동 검증
- **템플릿 동기화**: `ae update -t` 로 임베디드 템플릿 배포, 3-way 머지
- **바이너리 자동 업데이트**: GitHub Releases 기반 체크섬 검증 업데이트
- **멀티 언어 지원 (i18n)**: 한국어/영어/일본어/중국어
- **Windows/Unix 크로스 플랫폼**: Git Bash, PowerShell, zsh 지원

- **ae win / ae mac** (SPEC-PLATFORM-001): 플랫폼 전환 시 PATH 자동 재구성 및 환경 진단

### 계획 중 (SPEC 작성 완료)

- **업스트림 모니터링** (SPEC-UPDATE-001): moai-adk 릴리즈 추적, 선택적 템플릿 동기화
- **C# 스킬 고도화** (SPEC-SKILL-001): Wolverine/Mapster/AwesomeAssertions 기반 16모듈 스킬
- **E2E/Git규칙/나노스킬** (SPEC-PLATFORM-002): 테스트 인프라, 커밋 규칙 최적화, 소형 스킬

## 업스트림 관계

### moai-adk (modu-ai/moai-adk)

- **포크 원본**: moai-adk v2.7.20 기준
- **동기화 전략**: 선택적 반영 (자동 머지 아닌 수동 분석 후 적용)
- **동기화 범위**: 주로 템플릿 파일 (.claude/, .moai/ 하위), Hook 이벤트 핸들러, 에이전트 정의
- **독립 영역**: Go 코드는 ae-adk에서 독자적으로 관리 (직접 머지하지 않음)

### Claude Code

- **현재**: moai-adk를 통해 간접적으로 Claude Code 기능 반영
- **목표**: moai-adk 업데이트 중단 시에도 Claude Code 릴리즈 노트를 독립적으로 추적하고, 신기능(Hook 이벤트, StatusLine API, Skill 구조 등)을 ae-adk에 직접 반영할 수 있는 체계 구축

## 제거된 기능

### rank 시스템 (SPEC-REFACTOR-001, 2026-03-23)

- 외부 서비스 의존성으로 ae-adk 핵심 가치와 무관
- `internal/rank/`, `internal/ralph/`, CLI, Hook 연동 전체 삭제
- 47개 파일 변경, 24,848줄 삭제

### coverage padding 테스트 (SPEC-REFACTOR-001, 2026-03-23)

- 실제 검증 가치 없는 14개 패딩 테스트 파일 삭제

### deprecated 상수 (SPEC-REFACTOR-001, 2026-03-23)

- `BackupTimestampFormat`, `BackupsDir` → `defs.*` 참조로 교체
- `LangNameMap` 별칭 제거

### cc/glm/cg 런처 (SPEC-REFACTOR-002, 2026-03-24)

- moai-adk에서 상속받은 Claude/GLM/CG 모드 런처
- ae-adk에서는 `claude` 바이너리를 직접 실행하므로 불필요
- `internal/cli/cc.go`, `cg.go`, `glm.go`, `launcher.go` 및 관련 테스트 전체 삭제
- `session_end.go` GLM 정리 로직 제거, `root.go` launch 그룹 제거
- 62개 검증 테스트(`refactor_002_test.go`) 추가

## 기술 스택

| 영역 | 기술 |
|------|------|
| 언어 | Go 1.26 |
| CLI 프레임워크 | Cobra |
| TUI | charmbracelet/huh, lipgloss |
| 배포 | GitHub Releases (AngeleyesTrue/ae-adk) |
| 템플릿 | Go embed FS |
| 업스트림 | moai-adk v2.7.20 (modu-ai/moai-adk) |
| 대상 플랫폼 | Windows 11, macOS (Intel + Apple Silicon) |
