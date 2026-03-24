---
spec_id: SPEC-PLATFORM-001
type: plan
version: "1.0.0"
---

# SPEC-PLATFORM-001 구현 계획

## 마일스톤 개요

| 목표 | 범위 | 상태 |
|------|------|------|
| Primary Goal | 핵심 명령어: ae win / ae mac, PATH 재구성, 백업 | 미착수 |
| Secondary Goal | 플랫폼별 상세 진단 (Windows/macOS 고유 검증) | 미착수 |

> **참고**: Tertiary Goal(E2E 테스트)과 Optional Goal(Git 규칙, 나노 스킬)은 **SPEC-PLATFORM-002**로 분리되었다.

## Primary Goal: 핵심 명령어

### Task 1: Cobra 명령어 등록

- `cmd/platform_win.go`에 `winCmd` 정의
- `cmd/platform_mac.go`에 `macCmd` 정의
- `cmd/platform_common.go`에 공통 플래그 및 실행 흐름 정의
- `cmd/root.go`의 `"project"` 그룹에 winCmd, macCmd 등록
- 공통 플래그: `--force`, `--verbose`, `--json`, `--auto`, `--dry-run`, `--skip-backup`

### Task 2: settings.json 백업 시스템

- 타임스탬프 기반 백업 파일명 생성 (`settings.json.20260323-153000.bak`)
- 최근 5개 백업만 유지하는 로테이션 로직
- `--skip-backup` 플래그 처리
- 백업 디렉터리: settings.json과 동일 경로

### Task 3: PATH 재구성

- `BuildSmartPATH()` 호출하여 플랫폼별 PATH 생성
- `settings.json`의 PATH 필드만 업데이트 (다른 필드 보존)
- `--dry-run` 플래그로 변경 미리보기 출력
- PATH 검증: 존재하지 않는 디렉터리 경고
- `--auto` 플래그로 미존재 경로 자동 제외

### Task 4: 도구 버전 확인

- 공통 도구: ae, go, node, git 버전 확인
- 설치되지 않은 도구에 대해 CheckWarn 출력
- 버전 정보를 PlatformProfile에 저장

## Secondary Goal: 플랫폼별 상세 진단

### Task 5: Windows UTF-8 검증

- `chcp` 명령어로 현재 코드페이지 확인
- 65001(UTF-8)이 아니면 CheckWarn 출력 및 설정 방법 안내

### Task 6: Windows MCP 경로 확인

- `npx.cmd` 존재 및 경로 확인
- `pwsh.exe` 존재 및 경로 확인
- PATH에 포함되어 있는지 검증

### Task 7: Windows Git Bash / WSL2 감지

- MINGW64 환경 변수로 Git Bash 감지
- `wsl --status`로 WSL2 설치 및 상태 확인
- Git Bash 경로 변환 규칙 안내

### Task 8: Windows 260자 경로 제한

- 레지스트리 `LongPathsEnabled` 값 확인
- 비활성화 시 CheckWarn 출력 및 활성화 방법 안내

### Task 9: Windows Hook bash 확인

- `bash.exe` 존재 여부 확인
- Git Bash의 bash인지 WSL의 bash인지 구분
- Hook 실행에 필요한 bash 경로 안내

### Task 10: macOS Homebrew 검증

- Intel(`/usr/local/bin/brew`) vs Apple Silicon(`/opt/homebrew/bin/brew`) 경로 확인
- Homebrew 설치 여부 및 경로가 PATH에 포함되어 있는지 검증

### Task 11: macOS 심볼릭 링크 해석

- 주요 도구(node, go 등)의 심볼릭 링크를 실제 경로로 해석
- PATH에 심볼릭 링크 경로와 실제 경로 모두 포함되어 있는지 확인

### Task 12: macOS 셸 호환성

- 현재 셸(zsh, bash) 확인
- 셸 설정 파일(`.zshrc`, `.bash_profile`) 존재 여부 확인
- ae-adk 관련 PATH 설정이 셸 설정에 반영되어 있는지 확인

### Task 13: 프로필 저장/로드

- `~/.ae/platform-profile.json`에 PlatformProfile 저장
- 이전 프로필과 비교하여 변경사항 표시
- 프로필 포맷: JSON, 타임스탬프/플랫폼/체크 결과/PATH/도구 버전 포함

## 아키텍처 설계

```
┌─────────────────────────────────────────┐
│             CLI Layer (cmd/)            │
│  platform_win.go  platform_mac.go      │
│         platform_common.go              │
│  - Cobra 명령어 정의                     │
│  - 플래그 파싱                           │
│  - 출력 형식 제어 (text/verbose/json)    │
└─────────────────┬───────────────────────┘
                  │ 호출
┌─────────────────▼───────────────────────┐
│      Business Logic Layer (platform/)   │
│  profile.go      - 프로필 저장/로드/비교  │
│  diagnostics.go  - 진단 항목 실행/수집    │
│  validator.go    - 플랫폼별 검증 로직     │
│  - 순수 Go 로직                          │
│  - 테스트 용이                           │
└─────────────────────────────────────────┘
```

**설계 원칙:**

- CLI 레이어는 사용자 인터페이스만 담당한다 (Cobra 명령어, 플래그, 출력 형식)
- `platform/` 패키지는 비즈니스 로직만 담당한다 (진단, 검증, 프로필)
- CLI → platform 단방향 의존성만 허용한다
- `platform/` 패키지는 테스트 용이성을 위해 인터페이스 기반으로 설계한다

## 리스크

| # | 리스크 | 완화 방안 |
|---|--------|----------|
| 1 | settings.json 구조가 예상과 다름 | 실제 파일을 파싱하여 구조를 먼저 확인, 최소한의 필드만 수정 |
| 2 | BuildSmartPATH 경로 미존재 | REQ-008로 존재 여부 검증, --auto로 자동 제외 |
| 3 | 크로스 플랫폼 프로필 혼동 | 프로필에 platform 필드 명시, 다른 플랫폼 프로필은 덮어쓰지 않음 |
| 4 | Windows 레지스트리 접근 권한 부족 | graceful degradation: 접근 실패 시 CheckWarn으로 안내만 출력 |
| 5 | PATH 환경변수 Git Bash/PowerShell 충돌 | 셸 타입 감지 후 적절한 PATH 형식(Unix/Windows) 적용 |

## 테스트 전략

### 단위 테스트 (Unit Tests)

- `platform/` 패키지의 모든 공개 함수에 대한 단위 테스트
- 모의 파일 시스템을 사용한 settings.json 백업/복원 테스트
- PATH 생성 로직 테스트 (플랫폼별 경로 형식)

### 통합 테스트 (Integration Tests)

- Cobra 명령어 실행 흐름 테스트 (플래그 조합)
- settings.json 읽기/쓰기/백업 통합 테스트
- 프로필 저장/로드/비교 통합 테스트

### 플랫폼 특화 테스트 (Build Tags)

- `//go:build windows` 태그로 Windows 전용 테스트
- `//go:build darwin` 태그로 macOS 전용 테스트
- CI에서 양쪽 플랫폼 모두 테스트 실행

## 구현 순서

```
Task 1 (명령어 등록)
  └─→ Task 2 (백업 시스템)
       └─→ Task 3 (PATH 재구성)
            └─→ Task 4 (도구 버전)
                 ├─→ Task 5~9 (Windows 진단)
                 ├─→ Task 10~12 (macOS 진단)
                 └─→ Task 13 (프로필 저장)
```

## 전문가 협의

| 전문가 | 협의 내용 |
|--------|----------|
| expert-backend | Go 패키지 구조, 인터페이스 설계, 에러 핸들링 패턴 |
| expert-devops | 크로스 플랫폼 빌드, CI 파이프라인, build tags 전략 |
| expert-testing | 테스트 커버리지 목표, 모의 객체 설계, 플랫폼별 테스트 분리 |
