# Windows Compatibility: ae-adk 코드 변경사항

ae-adk 소스 코드에서 Windows 호환성을 위해 수정된 부분.
**이 문서는 ae-adk 개발자 및 `ae init`으로 설치하는 사용자에 해당합니다.**

## 변경 요약

| 파일 | 변경 내용 | 영향 시점 |
|------|----------|----------|
| `internal/template/settings.go` | `BuildSmartPATH()` Windows 분기 추가 | `ae init` / `ae update` |
| `internal/template/templates/.mcp.json.tmpl` | Windows에서 `pwsh.exe` 사용 | `ae init` |
| `Makefile` | 크로스 플랫폼 호환성 개선 | `make` 실행 시 |
| `internal/cmd/datestamp/main.go` | `date -u` 대체 유틸리티 | `make` 실행 시 |

## BuildSmartPATH() Windows 지원

`ae init` 실행 시 `settings.json`의 `env.PATH`를 생성하는 함수.

### 추가된 Windows 경로

**시스템 필수:**
- `%SystemRoot%\system32`, `%SystemRoot%`
- `%SystemRoot%\System32\Wbem`, `WindowsPowerShell\v1.0`, `OpenSSH`

**개발 도구:**
- `%ProgramFiles%\nodejs` (npx, npm, node)
- `%ProgramFiles%\Go\bin` (go)
- `%ProgramFiles%\PowerShell\7` (pwsh)
- `%ProgramFiles%\GitHub CLI` (gh)
- `%ProgramFiles%\Git\cmd` (git)
- `%ProgramFiles%\Docker\Docker\resources\bin` (docker)

**사용자 도구:**
- `%LOCALAPPDATA%\Programs\ae` (ae 바이너리 - install.ps1 설치)
- `%APPDATA%\npm` (전역 npm 패키지)
- `%LOCALAPPDATA%\Python\bin` (pip, python)
- `%LOCALAPPDATA%\Microsoft\WindowsApps` (winget)
- `%LOCALAPPDATA%\Programs\Microsoft VS Code\bin` (code CLI)

**Git Bash 호환:**
- `/usr/local/bin`, `/usr/local/sbin`, `/usr/bin`, `/bin`, `/usr/sbin`, `/sbin`

### ae 바이너리 설치 경로

| 설치 방식 | Windows 경로 | macOS/Linux 경로 |
|-----------|-------------|-----------------|
| install.ps1 (릴리스) | `%LOCALAPPDATA%\Programs\ae\ae.exe` | `~/.local/bin/ae` |
| go install (소스) | `%GOPATH%\bin\ae.exe` | `$GOPATH/bin/ae` |

`%GOPATH%\bin`은 `homeDir/go/bin`으로 이미 BuildSmartPATH()에 포함.

## .mcp.json.tmpl 변경

Windows 분기에서 `cmd.exe` → `pwsh.exe`로 변경:

```
Before: "command": "cmd.exe", "args": ["/c", "npx -y ..."]
After:  "command": "pwsh.exe", "args": ["-NoProfile", "-Command", "npx -y ..."]
```

이유: `cmd.exe`는 PATH 해석이 제한적이고, `npx.cmd` 실행에 문제 발생.

## Makefile 변경

| 항목 | Before | After |
|------|--------|-------|
| DATE | `date -u +"%Y-%m-%dT%H:%M:%SZ"` | `go run ./internal/cmd/datestamp` (fallback: date) |
| chmod | `chmod +x ...` | `chmod +x ... 2>/dev/null \|\| true` |
| clean | `rm -rf ...` | `go clean` + `rm -rf ... 2>/dev/null \|\| true` |

## 테스트

```bash
# Windows 전용 테스트
go test ./internal/template/ -run "TestBuildSmartPATH_Windows" -v

# MCP 템플릿 테스트
go test ./internal/template/ -run "TestMCPTemplatePlatformCommands" -v
```

---

Version: 1.0.0
Last Updated: 2026-03-18
