---
spec_id: SPEC-PLATFORM-001
type: acceptance
version: "1.0.0"
---

# SPEC-PLATFORM-001 수락 기준

## 1. 핵심 수락 기준: ae win / ae mac

### AC-001: 플랫폼 자동 감지

**시나리오 1: 잘못된 플랫폼에서 실행**

Given macOS에서 `ae win`을 실행할 때
When 플랫폼 감지가 수행되면
Then "현재 플랫폼은 darwin입니다. --force로 강제 실행할 수 있습니다" 경고가 출력되고 실행이 중단된다

**시나리오 2: --force로 강제 실행**

Given macOS에서 `ae win --force`를 실행할 때
When 플랫폼 감지가 수행되면
Then 경고 없이 Windows 진단이 실행된다

**시나리오 3: 올바른 플랫폼에서 실행**

Given Windows에서 `ae win`을 실행할 때
When 플랫폼 감지가 수행되면
Then 경고 없이 정상 진행된다

### AC-002: settings.json 백업

**시나리오 1: 백업 생성**

Given settings.json이 존재할 때
When `ae win`을 실행하면
Then `settings.json.YYYYMMDD-HHMMSS.bak` 형식의 백업 파일이 생성된다

**시나리오 2: 백업 로테이션**

Given 백업 파일이 6개 이상 존재할 때
When 새 백업이 생성되면
Then 가장 오래된 백업이 삭제되어 최대 5개만 유지된다

**시나리오 3: --skip-backup**

Given settings.json이 존재할 때
When `ae win --skip-backup`을 실행하면
Then 백업 파일이 생성되지 않는다

### AC-003: PATH 재구성

**시나리오 1: Windows PATH 생성**

Given Windows에서 `ae win`을 실행할 때
When PATH가 재구성되면
Then settings.json의 PATH에 Windows 형식 경로(`C:\...`)가 설정된다

**시나리오 2: macOS PATH 생성**

Given macOS에서 `ae mac`을 실행할 때
When PATH가 재구성되면
Then settings.json의 PATH에 Unix 형식 경로(`/usr/local/...`)가 설정된다

**시나리오 3: --dry-run**

Given `ae win --dry-run`을 실행할 때
When PATH 재구성이 수행되면
Then 변경 예정 내용이 출력되지만 settings.json은 수정되지 않는다

### AC-004: Windows UTF-8 검증

**시나리오 1: UTF-8 활성화됨**

Given Windows 코드페이지가 65001일 때
When UTF-8 검증이 수행되면
Then CheckOK "UTF-8 코드페이지 활성화됨"이 출력된다

**시나리오 2: UTF-8 비활성화**

Given Windows 코드페이지가 65001이 아닐 때
When UTF-8 검증이 수행되면
Then CheckWarn "UTF-8 코드페이지가 비활성화됨. chcp 65001 실행 권장"이 출력된다

### AC-005: Windows MCP 경로 확인

**시나리오 1: npx.cmd 존재**

Given `npx.cmd`가 PATH에 존재할 때
When MCP 경로 검증이 수행되면
Then CheckOK "npx.cmd 발견: C:\...\npx.cmd"이 출력된다

**시나리오 2: pwsh.exe 미존재**

Given `pwsh.exe`가 PATH에 없을 때
When MCP 경로 검증이 수행되면
Then CheckWarn "pwsh.exe를 찾을 수 없음. PowerShell 7 설치 권장"이 출력된다

### AC-006: Windows Git Bash / WSL2 감지

**시나리오 1: Git Bash 환경**

Given MINGW64 환경에서 실행 중일 때
When Git Bash 감지가 수행되면
Then CheckOK "Git Bash (MINGW64) 감지됨"이 출력된다

**시나리오 2: WSL2 설치됨**

Given WSL2가 설치되어 있을 때
When WSL2 검증이 수행되면
Then CheckOK "WSL2 사용 가능"이 출력된다

### AC-007: Windows 260자 경로 제한

**시나리오 1: LongPaths 활성화**

Given 레지스트리 LongPathsEnabled가 1일 때
When 경로 제한 검증이 수행되면
Then CheckOK "긴 경로 지원 활성화됨"이 출력된다

**시나리오 2: LongPaths 비활성화**

Given 레지스트리 LongPathsEnabled가 0이거나 없을 때
When 경로 제한 검증이 수행되면
Then CheckWarn "260자 경로 제한이 활성화됨. 레지스트리 설정 변경 권장"이 출력된다

### AC-008: Windows Hook bash 확인

**시나리오 1: bash 존재**

Given Git Bash의 bash.exe가 PATH에 존재할 때
When Hook bash 검증이 수행되면
Then CheckOK "bash 발견: C:\Program Files\Git\bin\bash.exe"가 출력된다

**시나리오 2: bash 미존재**

Given bash.exe가 PATH에 없을 때
When Hook bash 검증이 수행되면
Then CheckFail "bash를 찾을 수 없음. Git Bash 설치 필요"가 출력된다

### AC-009: macOS Homebrew 검증

**시나리오 1: Apple Silicon Homebrew**

Given `/opt/homebrew/bin/brew`가 존재할 때
When Homebrew 검증이 수행되면
Then CheckOK "Homebrew (Apple Silicon): /opt/homebrew"이 출력된다

**시나리오 2: Intel Homebrew**

Given `/usr/local/bin/brew`가 존재할 때
When Homebrew 검증이 수행되면
Then CheckOK "Homebrew (Intel): /usr/local"이 출력된다

**시나리오 3: Homebrew 미설치**

Given brew가 존재하지 않을 때
When Homebrew 검증이 수행되면
Then CheckWarn "Homebrew가 설치되지 않음"이 출력된다

### AC-010: macOS 심볼릭 링크 해석

**시나리오 1: 심볼릭 링크 정상**

Given `node`가 `/opt/homebrew/bin/node` → 실제 경로로 연결될 때
When 심볼릭 링크 검증이 수행되면
Then CheckOK "node: /opt/homebrew/bin/node → /opt/homebrew/Cellar/node/..."이 출력된다

**시나리오 2: 깨진 심볼릭 링크**

Given `node` 심볼릭 링크가 존재하지 않는 경로를 가리킬 때
When 심볼릭 링크 검증이 수행되면
Then CheckWarn "node: 심볼릭 링크가 깨져 있음"이 출력된다

### AC-011: macOS 셸 호환성

**시나리오 1: zsh 사용 중**

Given 기본 셸이 zsh일 때
When 셸 호환성 검증이 수행되면
Then CheckOK "기본 셸: zsh"이 출력된다

**시나리오 2: .zshrc 없음**

Given `.zshrc` 파일이 존재하지 않을 때
When 셸 호환성 검증이 수행되면
Then CheckWarn ".zshrc 파일이 없음. 셸 설정 확인 필요"가 출력된다

### AC-012: 공통 도구 버전 확인

**시나리오 1: 도구 설치됨**

Given go, node, git이 모두 설치되어 있을 때
When 도구 버전 확인이 수행되면
Then 각 도구에 대해 CheckOK "go: 1.26.x", CheckOK "node: v22.x.x", CheckOK "git: 2.x.x"이 출력된다

**시나리오 2: 도구 미설치**

Given node가 설치되지 않았을 때
When 도구 버전 확인이 수행되면
Then CheckWarn "node: 설치되지 않음"이 출력된다

### AC-013: 진단 결과 출력 형식

**시나리오 1: 기본 출력**

Given `ae win`을 플래그 없이 실행할 때
When 진단이 완료되면
Then CheckOK/CheckWarn/CheckFail 형식으로 요약 결과가 출력된다

**시나리오 2: --verbose**

Given `ae win --verbose`를 실행할 때
When 진단이 완료되면
Then 각 항목별 상세 정보(경로, 버전, 설정값)가 추가로 출력된다

**시나리오 3: --json**

Given `ae win --json`을 실행할 때
When 진단이 완료되면
Then 전체 결과가 JSON 형식으로 출력된다

### AC-014: 프로필 저장

**시나리오 1: 프로필 저장**

Given `ae win`이 성공적으로 완료되면
When 프로필 저장이 수행되면
Then `~/.ae/platform-profile.json`에 PlatformProfile이 저장된다

**시나리오 2: 이전 프로필과 비교**

Given 이전 프로필이 `~/.ae/platform-profile.json`에 존재할 때
When 새 진단이 완료되면
Then 이전 프로필과의 변경사항(새로운 경고, 해결된 문제)이 출력된다

### AC-015: PATH 검증

**시나리오 1: 존재하지 않는 경로 경고**

Given BuildSmartPATH가 `/usr/local/missing`를 포함할 때
When PATH 검증이 수행되면
Then CheckWarn "/usr/local/missing 경로가 존재하지 않음"이 출력된다

**시나리오 2: --auto로 자동 제외**

Given BuildSmartPATH가 존재하지 않는 경로를 포함하고 `--auto` 플래그가 설정될 때
When PATH 검증이 수행되면
Then 존재하지 않는 경로가 최종 PATH에서 자동 제외된다

> **참고**: AC-016~AC-021은 **SPEC-PLATFORM-002**로 분리됨

## 2. 품질 게이트 (Definition of Done)

### 필수 조건

- [ ] 테스트 커버리지 85% 이상 (`go test -cover`)
- [ ] 레이스 컨디션 없음 (`go test -race`)
- [ ] 린트 통과 (`golangci-lint run`)
- [ ] 정적 분석 통과 (`go vet ./...`)
- [ ] 기존 테스트 전체 통과 (`go test ./...`)
- [ ] settings.json 백업/복원 정상 동작
- [ ] moai-adk 영역(.claude/, .moai/, CLAUDE.md) 미변경
- [ ] Windows, macOS 양쪽 플랫폼에서 수동 검증 완료

### 권장 조건

- [ ] `ae doctor` 출력과 일관된 형식 유지

## 3. 검증 방법

### 자동 검증

| 검증 | 명령어 |
|------|--------|
| 단위/통합 테스트 | `go test ./...` |
| 레이스 감지 | `go test -race ./...` |
| 린트 | `golangci-lint run` |
| 빌드 확인 | `go build ./...` |
| 크로스 컴파일 (Windows) | `GOOS=windows GOARCH=amd64 go build ./...` |
| 크로스 컴파일 (macOS) | `GOOS=darwin GOARCH=arm64 go build ./...` |

### 수동 검증

| 검증 | 방법 |
|------|------|
| Windows 상세 진단 | `ae win --verbose` 실행 후 모든 CheckOK/CheckWarn/CheckFail 확인 |
| macOS 상세 진단 | `ae mac --verbose` 실행 후 모든 CheckOK/CheckWarn/CheckFail 확인 |
| 플랫폼 불일치 경고 | macOS에서 `ae win` 실행 시 경고 확인, `--force`로 강제 실행 확인 |
| 백업/복원 | `ae win` 실행 후 백업 파일 생성 확인, 5개 초과 시 로테이션 확인 |
| 드라이런 | `ae win --dry-run` 실행 후 settings.json 미변경 확인 |
| JSON 출력 | `ae win --json` 실행 후 유효한 JSON 출력 확인 |
