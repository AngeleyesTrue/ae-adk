package platform

import "strings"

// Validator는 플랫폼별 검증 로직 인터페이스이다.
type Validator interface {
	RunChecks() []PlatformCheck
}

// NewValidator는 대상 플랫폼에 맞는 Validator를 생성한다.
func NewValidator(sys SystemInfo, targetPlatform string) Validator {
	switch targetPlatform {
	case "windows":
		return &WindowsValidator{sys: sys}
	case "darwin":
		return &DarwinValidator{sys: sys}
	default:
		return &GenericValidator{sys: sys, platform: targetPlatform}
	}
}

// commonToolChecks는 모든 플랫폼에서 공통으로 실행하는 도구 버전 확인 목록이다.
func commonToolChecks(sys SystemInfo) []PlatformCheck {
	return []PlatformCheck{
		checkToolVersion(sys, "Go", "go", "version"),
		checkToolVersion(sys, "Node.js", "node", "--version"),
		checkToolVersion(sys, "Git", "git", "--version"),
		checkToolVersion(sys, "AE", "ae", "--version"),
	}
}

// checkToolVersion는 도구 실행 가능 여부를 확인하는 공통 진단 항목이다.
func checkToolVersion(sys SystemInfo, toolName string, cmd string, args ...string) PlatformCheck {
	check := PlatformCheck{Name: toolName}
	out, err := sys.ExecCommand(cmd, args...)
	if err != nil {
		check.Status = StatusWarn
		check.Message = "설치되지 않았거나 PATH에 없습니다"
		return check
	}
	check.Status = StatusOK
	check.Message = out
	return check
}

// GenericValidator는 플랫폼 특화 로직이 없는 범용 검증기이다.
type GenericValidator struct {
	sys      SystemInfo
	platform string
}

func (v *GenericValidator) RunChecks() []PlatformCheck {
	return commonToolChecks(v.sys)
}

// WindowsValidator는 Windows 전용 검증기이다.
type WindowsValidator struct {
	sys SystemInfo
}

// @MX:NOTE: [AUTO] Windows 전용 진단 항목 - UTF-8, MCP, Git Bash, WSL2, LongPath, Hook bash, 도구 버전
func (v *WindowsValidator) RunChecks() []PlatformCheck {
	checks := make([]PlatformCheck, 0, 14)

	// REQ-004: Windows 전용 검증
	checks = append(checks, v.checkUTF8CodePage())
	checks = append(checks, v.checkMCPServerPaths()...)
	checks = append(checks, v.checkGitBash())
	checks = append(checks, v.checkWSL2())
	checks = append(checks, v.checkLongPathsEnabled())
	checks = append(checks, v.checkHookBash())

	// 공통 도구 버전
	checks = append(checks, commonToolChecks(v.sys)...)

	return checks
}

func (v *WindowsValidator) checkUTF8CodePage() PlatformCheck {
	check := PlatformCheck{Name: "UTF-8 CodePage"}
	out, err := v.sys.ExecCommand("cmd.exe", "/c", "chcp")
	if err != nil {
		check.Status = StatusWarn
		check.Message = "코드 페이지 확인 불가"
		return check
	}
	if strings.Contains(out, "65001") {
		check.Status = StatusOK
		check.Message = "UTF-8 (65001) 활성화됨"
	} else {
		check.Status = StatusWarn
		check.Message = "UTF-8이 아닙니다: " + out
		check.Detail = "chcp 65001 실행으로 UTF-8 활성화 가능"
	}
	return check
}

func (v *WindowsValidator) checkMCPServerPaths() []PlatformCheck {
	mcpTools := []struct {
		name string
		path string
	}{
		{"npx", "npx"},
		{"npx.cmd", "npx.cmd"},
		{"pwsh.exe", "pwsh.exe"},
	}

	checks := make([]PlatformCheck, 0, len(mcpTools))
	for _, tool := range mcpTools {
		check := PlatformCheck{Name: "MCP: " + tool.name}
		_, err := v.sys.ExecCommand("where.exe", tool.path)
		if err != nil {
			// which로도 시도 (Git Bash 환경)
			_, err2 := v.sys.ExecCommand("which", tool.path)
			if err2 != nil {
				check.Status = StatusWarn
				check.Message = "PATH에서 찾을 수 없습니다"
			} else {
				check.Status = StatusOK
				check.Message = "Git Bash PATH에서 발견"
			}
		} else {
			check.Status = StatusOK
			check.Message = "시스템 PATH에서 발견"
		}
		checks = append(checks, check)
	}
	return checks
}

func (v *WindowsValidator) checkGitBash() PlatformCheck {
	check := PlatformCheck{Name: "Git Bash"}
	msystem := v.sys.GetEnv("MSYSTEM")
	if msystem != "" {
		check.Status = StatusOK
		check.Message = "MSYSTEM=" + msystem
		return check
	}
	// Git Bash 설치 확인
	gitBashPaths := []string{
		`C:\Program Files\Git\bin\bash.exe`,
		`C:\Program Files (x86)\Git\bin\bash.exe`,
	}
	for _, p := range gitBashPaths {
		if v.sys.FileExists(p) {
			check.Status = StatusOK
			check.Message = "설치됨: " + p
			return check
		}
	}
	check.Status = StatusWarn
	check.Message = "Git Bash를 찾을 수 없습니다"
	return check
}

func (v *WindowsValidator) checkWSL2() PlatformCheck {
	check := PlatformCheck{Name: "WSL2"}
	out, err := v.sys.ExecCommand("wsl.exe", "--status")
	if err != nil {
		check.Status = StatusOK
		check.Message = "WSL 미설치 또는 비활성화"
		return check
	}
	check.Status = StatusOK
	check.Message = "WSL 활성화됨"
	check.Detail = out
	return check
}

// @MX:NOTE: [AUTO] 260자 경로 길이 제한 확인 - LongPathsEnabled 레지스트리 값 확인
func (v *WindowsValidator) checkLongPathsEnabled() PlatformCheck {
	check := PlatformCheck{Name: "LongPaths"}
	out, err := v.sys.ExecCommand("reg", "query",
		`HKLM\SYSTEM\CurrentControlSet\Control\FileSystem`,
		"/v", "LongPathsEnabled")
	if err != nil {
		check.Status = StatusWarn
		check.Message = "레지스트리 확인 불가"
		return check
	}
	if strings.Contains(out, "0x1") {
		check.Status = StatusOK
		check.Message = "260자 제한 해제됨"
	} else {
		check.Status = StatusWarn
		check.Message = "260자 경로 제한이 활성화되어 있습니다"
		check.Detail = "관리자 권한으로 레지스트리에서 LongPathsEnabled=1 설정 권장"
	}
	return check
}

func (v *WindowsValidator) checkHookBash() PlatformCheck {
	check := PlatformCheck{Name: "Hook Bash"}
	bashPaths := []string{
		`C:\Program Files\Git\bin\bash.exe`,
		`C:\Program Files (x86)\Git\bin\bash.exe`,
		"/usr/bin/bash",
	}
	for _, p := range bashPaths {
		if v.sys.FileExists(p) {
			check.Status = StatusOK
			check.Message = "발견: " + p
			return check
		}
	}
	_, err := v.sys.ExecCommand("which", "bash")
	if err == nil {
		check.Status = StatusOK
		check.Message = "PATH에서 bash 발견"
		return check
	}
	check.Status = StatusFail
	check.Message = "bash를 찾을 수 없습니다 - Hook 실행 불가"
	return check
}

// DarwinValidator는 macOS 전용 검증기이다.
type DarwinValidator struct {
	sys SystemInfo
}

// @MX:NOTE: [AUTO] macOS 전용 진단 항목 - Homebrew, symlink, shell, 도구 버전
func (v *DarwinValidator) RunChecks() []PlatformCheck {
	checks := make([]PlatformCheck, 0, 10)

	// REQ-005: macOS 전용 검증
	checks = append(checks, v.checkHomebrew())
	checks = append(checks, v.checkSymlinks()...)
	checks = append(checks, v.checkShellCompat())

	// 공통 도구 버전
	checks = append(checks, commonToolChecks(v.sys)...)

	return checks
}

func (v *DarwinValidator) checkHomebrew() PlatformCheck {
	check := PlatformCheck{Name: "Homebrew"}
	arch := v.sys.GOARCH()

	var expectedPath string
	if arch == "arm64" {
		expectedPath = "/opt/homebrew/bin/brew"
	} else {
		expectedPath = "/usr/local/bin/brew"
	}

	if v.sys.FileExists(expectedPath) {
		out, _ := v.sys.ExecCommand(expectedPath, "--version")
		check.Status = StatusOK
		check.Message = out
		check.Detail = "Path: " + expectedPath
		return check
	}

	// 다른 경로에서 시도
	out, err := v.sys.ExecCommand("brew", "--version")
	if err != nil {
		check.Status = StatusWarn
		check.Message = "Homebrew 미설치"
		return check
	}
	check.Status = StatusOK
	check.Message = out
	return check
}

func (v *DarwinValidator) checkSymlinks() []PlatformCheck {
	tools := []struct {
		name string
		path string
	}{
		{"node", "/usr/local/bin/node"},
		{"python3", "/usr/local/bin/python3"},
	}

	checks := make([]PlatformCheck, 0, len(tools))
	for _, tool := range tools {
		check := PlatformCheck{Name: "Symlink: " + tool.name}
		out, err := v.sys.ExecCommand("readlink", "-f", tool.path)
		if err != nil {
			check.Status = StatusOK
			check.Message = "심볼릭 링크 없음 (직접 설치)"
		} else {
			check.Status = StatusOK
			check.Message = tool.path + " -> " + out
		}
		checks = append(checks, check)
	}
	return checks
}

func (v *DarwinValidator) checkShellCompat() PlatformCheck {
	check := PlatformCheck{Name: "Shell"}
	shell := v.sys.GetEnv("SHELL")
	if shell == "" {
		check.Status = StatusWarn
		check.Message = "$SHELL 미설정"
		return check
	}
	check.Status = StatusOK
	check.Message = shell
	if strings.Contains(shell, "zsh") {
		check.Detail = "기본 셸 (macOS Catalina+)"
	} else if strings.Contains(shell, "bash") {
		check.Detail = "Bash 셸"
	}
	return check
}
