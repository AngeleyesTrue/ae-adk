---
id: SPEC-PLATFORM-001
version: "1.0.0"
status: implemented
created: "2026-03-23"
updated: "2026-03-24"
author: Angeleyes
priority: high
issue_number: 0
tags: [platform, windows, macos, cli]
---

# SPEC-PLATFORM-001: ae win / ae mac 플랫폼 전환 명령어 추가

## 1. 환경 (Environment)

| 항목 | 내용 |
|------|------|
| 제품 | ae-adk |
| 언어/프레임워크 | Go 1.26, Cobra CLI |
| 대상 플랫폼 | Windows 11 (70%), macOS (30%) |
| 사용자 | Angeleyes (개인용) |
| 현재 상태 | 플랫폼 전환 시 PATH 불일치로 MCP 서버/Hook/StatusLine 동작 안됨 |
| 기존 코드 | `template/settings.go` BuildSmartPATH(), `shell/detect.go`, `console_windows.go`/`console_others.go`, `core/pathutil_windows.go`/`pathutil_nonwindows.go` |
| 관련 TODO | TODO #4 |
| 제약 | moai-adk 파일(.claude/, .moai/, CLAUDE.md) 수정 금지, ae init/doctor 동작 변경 금지 |

## 2. 가정 (Assumptions)

| ID | 가정 | 영향 |
|----|------|------|
| A-01 | `runtime.GOOS`가 현재 OS를 정확히 반환한다 | 플랫폼 자동 감지의 기반 |
| A-02 | `BuildSmartPATH()`는 재사용 가능한 형태로 리팩터링할 수 있다 | PATH 재구성 로직 공유 |
| A-03 | `settings.json`의 PATH 필드가 MCP 서버 실행에 직접 영향을 준다 | PATH 재구성의 핵심 근거 |
| A-04 | Windows에서 Git Bash는 MINGW64 환경으로 동작한다 | Git Bash 감지 및 경로 변환 |
| A-05 | macOS Homebrew는 Intel(`/usr/local`)과 Apple Silicon(`/opt/homebrew`) 경로가 다르다 | macOS PATH 구성 |
| A-06 | 사용자는 70% Windows, 30% macOS 환경을 사용한다 | 우선순위 결정 |
| A-07 | Windows에서 PowerShell이 기본 셸이지만, ae-adk 작업은 Git Bash에서 수행한다 | 셸 감지 및 PATH 형식 |
| A-08 | macOS에서 기본 셸은 zsh이다 | macOS 셸 호환성 검증 |

## 3. 요구사항 (Requirements)

### REQ-001: 플랫폼 자동 감지 (Ubiquitous)

`ae win` 또는 `ae mac` 실행 시 `runtime.GOOS`를 통해 현재 플랫폼을 자동 감지한다.

- 현재 플랫폼과 명령어가 불일치하면 경고 메시지를 출력한다
- `--force` 플래그로 경고를 무시하고 강제 실행할 수 있다
- 일치하면 경고 없이 정상 진행한다

### REQ-002: settings.json 백업 (Event-Driven)

PATH 재구성 전에 현재 `settings.json`을 백업한다.

- 타임스탬프 기반 백업 파일명 (예: `settings.json.20260323-153000.bak`)
- 최근 5개까지만 유지하고 오래된 백업은 자동 삭제한다
- `--skip-backup` 플래그로 백업을 건너뛸 수 있다

### REQ-003: PATH 재구성 (Event-Driven)

`BuildSmartPATH()`를 활용하여 현재 플랫폼에 맞는 PATH를 재구성한다.

- 플랫폼별 올바른 경로 형식으로 `settings.json`을 업데이트한다
- `--dry-run` 플래그로 실제 변경 없이 미리보기를 지원한다

### REQ-004: Windows 전용 검증 (Event-Driven)

Windows 환경에서 다음 항목을 검증한다:

- UTF-8 코드페이지 활성화 여부 (chcp 65001)
- MCP 서버 경로 확인 (`npx.cmd`, `pwsh.exe` 등)
- Git Bash 환경 감지 및 MINGW64 경로 변환
- WSL2 설치 및 상태 확인
- 260자 경로 길이 제한 (LongPathsEnabled) 확인
- Hook 실행을 위한 bash 존재 확인
- 도구 버전 확인 (ae, go, node, git)

### REQ-005: macOS 전용 검증 (Event-Driven)

macOS 환경에서 다음 항목을 검증한다:

- Homebrew 설치 경로 (Intel `/usr/local` vs Apple Silicon `/opt/homebrew`)
- 심볼릭 링크 해석 (예: `node` → 실제 경로)
- 셸 호환성 확인 (zsh, bash)
- 도구 버전 확인 (ae, go, node, git)

### REQ-006: 진단 결과 출력 (Ubiquitous)

모든 검증 결과를 일관된 형식으로 출력한다.

- 기본: CheckOK/CheckWarn/CheckFail 형식
- `--verbose`: 상세 진단 정보 포함
- `--json`: JSON 형식으로 출력

### REQ-007: 플랫폼 프로필 저장 (Event-Driven)

진단 결과를 프로필로 저장하여 이후 비교에 활용한다.

- 저장 위치: `~/.ae/platform-profile.json`
- 이전 프로필과 비교하여 변경사항을 표시한다

### REQ-008: BuildSmartPATH 출력 검증 (Event-Driven)

`BuildSmartPATH()`가 생성한 경로들의 실제 존재 여부를 검증한다.

- 존재하지 않는 디렉터리에 대해 경고를 출력한다
- `--auto` 플래그로 존재하지 않는 경로를 자동 제외한다

> **분리됨**: E2E 테스트(TODO #7), Git 규칙(TODO #8), 나노 스킬(TODO #9)은 **SPEC-PLATFORM-002**로 분리됨

## 4. 사양 (Specifications)

### 4.1 CLI 인터페이스

```
ae win [flags]    # Windows 플랫폼 전환 및 진단
ae mac [flags]    # macOS 플랫폼 전환 및 진단
```

**공통 플래그:**

| 플래그 | 타입 | 기본값 | 설명 |
|--------|------|--------|------|
| `--force` | bool | false | 플랫폼 불일치 경고 무시 |
| `--verbose` | bool | false | 상세 진단 출력 |
| `--json` | bool | false | JSON 형식 출력 |
| `--auto` | bool | false | 존재하지 않는 경로 자동 제외 |
| `--dry-run` | bool | false | 실제 변경 없이 미리보기 |
| `--skip-backup` | bool | false | settings.json 백업 건너뛰기 |

### 4.2 파일 구조

```
cmd/
  platform_win.go       # ae win 명령어 (Cobra Command)
  platform_mac.go       # ae mac 명령어 (Cobra Command)
  platform_common.go    # 공통 플래그, 실행 흐름

platform/
  profile.go            # PlatformProfile 저장/로드/비교
  diagnostics.go        # 진단 항목 실행 및 결과 수집
  validator.go          # 플랫폼별 검증 로직
```

### 4.3 데이터 구조

**PlatformProfile:**

```go
type PlatformProfile struct {
    Platform    string            `json:"platform"`     // "windows" | "darwin"
    Timestamp   time.Time         `json:"timestamp"`
    Checks      []PlatformCheck   `json:"checks"`
    PATH        []string          `json:"path"`
    ToolVersions map[string]string `json:"tool_versions"`
}
```

**PlatformCheck:**

```go
type PlatformCheck struct {
    Name    string `json:"name"`
    Status  string `json:"status"`   // "ok" | "warn" | "fail"
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
}
```

### 4.4 실행 흐름

```
ae win / ae mac
  │
  ├─ 1. 플랫폼 감지 (REQ-001)
  │     └─ 불일치 시 경고 → --force로 계속 / 중단
  │
  ├─ 2. settings.json 백업 (REQ-002)
  │     └─ --skip-backup 시 건너뜀
  │
  ├─ 3. PATH 재구성 (REQ-003)
  │     ├─ BuildSmartPATH() 호출
  │     ├─ PATH 검증 (REQ-008)
  │     │     └─ --auto 시 미존재 경로 제외
  │     └─ --dry-run 시 미리보기만 출력
  │
  ├─ 4. 플랫폼별 검증 (REQ-004 / REQ-005)
  │     ├─ Windows: UTF-8, MCP, Git Bash, WSL2, 260자, Hook bash, 도구 버전
  │     └─ macOS: Homebrew, symlink, shell, 도구 버전
  │
  ├─ 5. 진단 결과 출력 (REQ-006)
  │     └─ --verbose / --json 형식 선택
  │
  └─ 6. 프로필 저장 (REQ-007)
        └─ ~/.ae/platform-profile.json
```

### 4.5 moai-adk 파일 보호

ae-adk는 moai-adk 위에 구축되어 있으므로 다음 규칙을 준수한다:

- `.claude/`, `.moai/`, `CLAUDE.md` 파일은 절대 수정하지 않는다
- `settings.json`의 **PATH 필드만** 유일한 예외로 수정을 허용한다
- 그 외 `settings.json` 필드(mcpServers, hooks 등)는 변경하지 않는다

## 5. 추적성 (Traceability)

| 요구사항 | TODO | 설명 |
|----------|------|------|
| REQ-001 | #4 | 플랫폼 자동 감지 |
| REQ-002 | #4 | settings.json 백업 |
| REQ-003 | #4 | PATH 재구성 |
| REQ-004 | #4 | Windows 전용 검증 |
| REQ-005 | #4 | macOS 전용 검증 |
| REQ-006 | #4 | 진단 결과 출력 |
| REQ-007 | #4 | 플랫폼 프로필 저장 |
| REQ-008 | #4 | BuildSmartPATH 출력 검증 |

> **참고**: E2E 테스트(TODO #7), Git 규칙(TODO #8), 나노 스킬(TODO #9)은 **SPEC-PLATFORM-002**에서 다룬다.
