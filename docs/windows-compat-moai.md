# Windows Compatibility: moai 실행 환경

moai-adk가 설치된 프로젝트에서 Windows 환경 문제와 해결 방법.
**이 문서는 현재 moai 바이너리로 실행 중인 프로젝트에 해당합니다.**

## 환경 개요

| 항목 | macOS/Linux | Windows |
|------|-------------|---------|
| Claude Code Shell | /bin/bash (native) | Git Bash (MINGW64) |
| MCP Server Shell | /bin/bash -l -c | pwsh.exe -NoProfile -Command |
| moai 위치 | ~/.local/bin/moai | %LOCALAPPDATA%\Programs\moai\moai.exe |
| npx 위치 | /usr/local/bin/npx | C:\Program Files\nodejs\npx.cmd |

## 문제 1: MCP Server에서 npx 인식 불가

### 증상

MCP 서버(context7, sequential-thinking)가 시작되지 않음.

### 원인

`.mcp.json`이 `/bin/bash`를 사용하는데, settings.json PATH에 `C:\Program Files\nodejs\`가 없음.

### 해결

`.mcp.json`에서 PowerShell 7 사용:

```json
{
  "context7": {
    "command": "C:\\Program Files\\PowerShell\\7\\pwsh.exe",
    "args": ["-NoProfile", "-Command", "npx -y @upstash/context7-mcp@latest"]
  }
}
```

## 문제 2: settings.json PATH에 Windows 경로 누락

### 증상

moai, gh, npm, go, pip 등 명령어를 찾지 못함. Status line 미표시.

### 원인

settings.json의 `env.PATH`가 시스템 PATH를 덮어쓰는데, Windows 경로가 누락됨.

### 해결

`.claude/settings.json`의 `env.PATH`에 다음 경로 추가:

```
C:\Users\사용자\AppData\Local\Programs\moai   (moai 바이너리)
C:\WINDOWS\system32                            (시스템)
C:\Program Files\nodejs                        (npx, npm, node)
C:\Program Files\Go\bin                        (go)
C:\Program Files\PowerShell\7                  (pwsh)
C:\Program Files\GitHub CLI                    (gh)
C:\Program Files\Git\cmd                       (git)
C:\Users\사용자\AppData\Roaming\npm            (전역 npm)
C:\Users\사용자\AppData\Local\Python\bin       (pip)
```

## 문제 3: Status Line 미표시

### 원인

settings.json PATH에 moai 설치 경로 누락.

### 해결

PATH에 `%LOCALAPPDATA%\Programs\moai` 추가 후 세션 재시작.

### 확인

```bash
which moai
echo '{}' | bash .moai/status_line.sh
```

## 문제 4: Makefile Unix 명령어

### 해결

Git Bash에서 `make` 실행. PowerShell에서는 직접 `go build`/`go test` 사용.

## macOS ↔ Windows 전환 체크리스트

1. `.mcp.json` 수정: `/bin/bash` ↔ `pwsh.exe`
2. `.claude/settings.json` PATH 수정: 플랫폼별 경로
3. `moai` 바이너리 설치 확인: `which moai`
4. Status line 확인: `echo '{}' | bash .moai/status_line.sh`

## 알려진 제한사항

| 항목 | 설명 |
|------|------|
| tmux (CG Mode) | Windows 미지원 |
| 한글 사용자명 | 일부 도구에서 인코딩 문제 가능 |
| 파일 경로 길이 | 260자 제한 주의 |

---

Version: 1.0.0
Last Updated: 2026-03-18
