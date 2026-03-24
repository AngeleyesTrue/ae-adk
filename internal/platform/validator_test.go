package platform

import (
	"fmt"
	"testing"
)

// TestNewValidator는 플랫폼에 맞는 Validator를 올바르게 생성하는지 확인한다.
func TestNewValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		targetPlatform string
		wantType       string
	}{
		{
			name:           "windows는 WindowsValidator",
			targetPlatform: "windows",
			wantType:       "*platform.WindowsValidator",
		},
		{
			name:           "darwin은 DarwinValidator",
			targetPlatform: "darwin",
			wantType:       "*platform.DarwinValidator",
		},
		{
			name:           "linux는 GenericValidator",
			targetPlatform: "linux",
			wantType:       "*platform.GenericValidator",
		},
		{
			name:           "알 수 없는 플랫폼은 GenericValidator",
			targetPlatform: "freebsd",
			wantType:       "*platform.GenericValidator",
		},
		{
			name:           "빈 문자열은 GenericValidator",
			targetPlatform: "",
			wantType:       "*platform.GenericValidator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			v := NewValidator(mock, tt.targetPlatform)

			gotType := fmt.Sprintf("%T", v)
			if gotType != tt.wantType {
				t.Errorf("NewValidator(%q) type = %q, want %q", tt.targetPlatform, gotType, tt.wantType)
			}
		})
	}
}

// TestCheckToolVersion은 도구 실행 가능 여부 확인 로직을 테스트한다.
func TestCheckToolVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		toolName   string
		cmd        string
		args       []string
		mockOutput string
		mockError  error
		wantStatus CheckStatus
		wantMsg    string // 포함되어야 하는 메시지
	}{
		{
			name:       "성공: 도구가 설치됨",
			toolName:   "Go",
			cmd:        "go",
			args:       []string{"version"},
			mockOutput: "go version go1.26 windows/amd64",
			wantStatus: StatusOK,
			wantMsg:    "go version go1.26 windows/amd64",
		},
		{
			name:       "실패: 도구 미설치",
			toolName:   "Node.js",
			cmd:        "node",
			args:       []string{"--version"},
			mockError:  fmt.Errorf("not found"),
			wantStatus: StatusWarn,
			wantMsg:    "설치되지 않았거나 PATH에 없습니다",
		},
		{
			name:       "성공: 버전 출력에 공백 포함",
			toolName:   "Git",
			cmd:        "git",
			args:       []string{"--version"},
			mockOutput: "git version 2.44.0.windows.1",
			wantStatus: StatusOK,
			wantMsg:    "git version 2.44.0.windows.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()

			// 명령어 키 생성
			key := tt.cmd
			for _, a := range tt.args {
				key += " " + a
			}

			if tt.mockError != nil {
				mock.CommandErrors[key] = tt.mockError
			} else {
				mock.Commands[key] = tt.mockOutput
			}

			check := checkToolVersion(mock, tt.toolName, tt.cmd, tt.args...)

			if check.Name != tt.toolName {
				t.Errorf("Name = %q, want %q", check.Name, tt.toolName)
			}
			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", check.Status, tt.wantStatus)
			}
			if check.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", check.Message, tt.wantMsg)
			}
		})
	}
}

// TestGenericValidator_RunChecks는 범용 검증기의 체크 항목을 확인한다.
func TestGenericValidator_RunChecks(t *testing.T) {
	t.Parallel()

	t.Run("모든 도구가 설치된 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.Commands["go version"] = "go1.26"
		mock.Commands["node --version"] = "v22.0.0"
		mock.Commands["git --version"] = "git 2.44.0"
		mock.Commands["ae --version"] = "ae 1.0.0"

		v := &GenericValidator{sys: mock, platform: "linux"}
		checks := v.RunChecks()

		if len(checks) != 4 {
			t.Fatalf("RunChecks() returned %d checks, want 4", len(checks))
		}

		// 모든 체크가 OK인지 확인
		for _, c := range checks {
			if c.Status != StatusOK {
				t.Errorf("check %q Status = %q, want %q", c.Name, c.Status, StatusOK)
			}
		}

		// 이름 확인
		expectedNames := []string{"Go", "Node.js", "Git", "AE"}
		for i, name := range expectedNames {
			if checks[i].Name != name {
				t.Errorf("checks[%d].Name = %q, want %q", i, checks[i].Name, name)
			}
		}
	})

	t.Run("모든 도구가 미설치인 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		// Commands가 비어있음

		v := &GenericValidator{sys: mock, platform: "linux"}
		checks := v.RunChecks()

		for _, c := range checks {
			if c.Status != StatusWarn {
				t.Errorf("check %q Status = %q, want %q", c.Name, c.Status, StatusWarn)
			}
		}
	})
}

// TestWindowsValidator_RunChecks는 Windows 검증기의 전체 체크를 확인한다.
func TestWindowsValidator_RunChecks(t *testing.T) {
	t.Parallel()

	t.Run("모든 체크 통과", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.GOOSVal = "windows"

		// UTF-8 CodePage
		mock.Commands["cmd.exe /c chcp"] = "Active code page: 65001"

		// MCP Server Paths
		mock.Commands["where.exe npx"] = "C:\\nodejs\\npx"
		mock.Commands["where.exe npx.cmd"] = "C:\\nodejs\\npx.cmd"
		mock.Commands["where.exe pwsh.exe"] = "C:\\Windows\\pwsh.exe"

		// Git Bash (MSYSTEM 환경 변수)
		mock.EnvVars["MSYSTEM"] = "MINGW64"

		// WSL2
		mock.Commands["wsl.exe --status"] = "Default Version: 2"

		// LongPaths
		mock.Commands["reg query HKLM\\SYSTEM\\CurrentControlSet\\Control\\FileSystem /v LongPathsEnabled"] = "LongPathsEnabled    REG_DWORD    0x1"

		// Hook Bash
		mock.Files[`C:\Program Files\Git\bin\bash.exe`] = []byte{}

		// 공통 도구
		mock.Commands["go version"] = "go1.26"
		mock.Commands["node --version"] = "v22"
		mock.Commands["git --version"] = "git 2.44"
		mock.Commands["ae --version"] = "ae 1.0"

		v := &WindowsValidator{sys: mock}
		checks := v.RunChecks()

		// 최소 10개 이상의 체크가 있어야 함
		if len(checks) < 10 {
			t.Fatalf("RunChecks() returned %d checks, want >= 10", len(checks))
		}

		// 중요한 체크 이름 확인
		checkNames := make(map[string]bool)
		for _, c := range checks {
			checkNames[c.Name] = true
		}

		requiredNames := []string{
			"UTF-8 CodePage",
			"Git Bash",
			"WSL2",
			"LongPaths",
			"Hook Bash",
			"Go",
			"Node.js",
			"Git",
			"AE",
		}
		for _, name := range requiredNames {
			if !checkNames[name] {
				t.Errorf("필수 체크 %q 가 결과에 없음", name)
			}
		}
	})
}

// TestWindowsValidator_checkUTF8CodePage는 UTF-8 코드 페이지 확인을 테스트한다.
func TestWindowsValidator_checkUTF8CodePage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		output     string
		err        error
		wantStatus CheckStatus
	}{
		{
			name:       "UTF-8 활성화됨",
			output:     "Active code page: 65001",
			wantStatus: StatusOK,
		},
		{
			name:       "UTF-8 아닌 코드 페이지",
			output:     "Active code page: 949",
			wantStatus: StatusWarn,
		},
		{
			name:       "명령어 실행 실패",
			err:        fmt.Errorf("cmd not found"),
			wantStatus: StatusWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			if tt.err != nil {
				mock.CommandErrors["cmd.exe /c chcp"] = tt.err
			} else {
				mock.Commands["cmd.exe /c chcp"] = tt.output
			}

			v := &WindowsValidator{sys: mock}
			check := v.checkUTF8CodePage()

			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q (message: %s)", check.Status, tt.wantStatus, check.Message)
			}
			if check.Name != "UTF-8 CodePage" {
				t.Errorf("Name = %q, want 'UTF-8 CodePage'", check.Name)
			}
		})
	}
}

// TestWindowsValidator_checkMCPServerPaths는 MCP 서버 경로 확인을 테스트한다.
func TestWindowsValidator_checkMCPServerPaths(t *testing.T) {
	t.Parallel()

	t.Run("모든 도구를 where.exe로 찾은 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.Commands["where.exe npx"] = "C:\\npx"
		mock.Commands["where.exe npx.cmd"] = "C:\\npx.cmd"
		mock.Commands["where.exe pwsh.exe"] = "C:\\pwsh.exe"

		v := &WindowsValidator{sys: mock}
		checks := v.checkMCPServerPaths()

		if len(checks) != 3 {
			t.Fatalf("checkMCPServerPaths() returned %d checks, want 3", len(checks))
		}
		for _, c := range checks {
			if c.Status != StatusOK {
				t.Errorf("check %q Status = %q, want OK", c.Name, c.Status)
			}
		}
	})

	t.Run("which 폴백으로 찾은 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		// where.exe는 실패하지만 which는 성공
		mock.CommandErrors["where.exe npx"] = fmt.Errorf("not found")
		mock.Commands["which npx"] = "/usr/bin/npx"
		mock.CommandErrors["where.exe npx.cmd"] = fmt.Errorf("not found")
		mock.Commands["which npx.cmd"] = "/usr/bin/npx.cmd"
		mock.CommandErrors["where.exe pwsh.exe"] = fmt.Errorf("not found")
		mock.Commands["which pwsh.exe"] = "/usr/bin/pwsh.exe"

		v := &WindowsValidator{sys: mock}
		checks := v.checkMCPServerPaths()

		for _, c := range checks {
			if c.Status != StatusOK {
				t.Errorf("check %q Status = %q, want OK (Git Bash fallback)", c.Name, c.Status)
			}
		}
	})

	t.Run("모두 찾지 못한 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		// 모든 명령어가 실패
		mock.CommandErrors["where.exe npx"] = fmt.Errorf("not found")
		mock.CommandErrors["which npx"] = fmt.Errorf("not found")
		mock.CommandErrors["where.exe npx.cmd"] = fmt.Errorf("not found")
		mock.CommandErrors["which npx.cmd"] = fmt.Errorf("not found")
		mock.CommandErrors["where.exe pwsh.exe"] = fmt.Errorf("not found")
		mock.CommandErrors["which pwsh.exe"] = fmt.Errorf("not found")

		v := &WindowsValidator{sys: mock}
		checks := v.checkMCPServerPaths()

		for _, c := range checks {
			if c.Status != StatusWarn {
				t.Errorf("check %q Status = %q, want Warn", c.Name, c.Status)
			}
		}
	})
}

// TestWindowsValidator_checkGitBash는 Git Bash 확인을 테스트한다.
func TestWindowsValidator_checkGitBash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		envVars    map[string]string
		files      map[string][]byte
		wantStatus CheckStatus
	}{
		{
			name:       "MSYSTEM 환경 변수 설정됨",
			envVars:    map[string]string{"MSYSTEM": "MINGW64"},
			wantStatus: StatusOK,
		},
		{
			name:       "Program Files에서 bash.exe 발견",
			envVars:    map[string]string{},
			files:      map[string][]byte{`C:\Program Files\Git\bin\bash.exe`: {}},
			wantStatus: StatusOK,
		},
		{
			name:       "Program Files (x86)에서 bash.exe 발견",
			envVars:    map[string]string{},
			files:      map[string][]byte{`C:\Program Files (x86)\Git\bin\bash.exe`: {}},
			wantStatus: StatusOK,
		},
		{
			name:       "Git Bash를 찾을 수 없음",
			envVars:    map[string]string{},
			files:      map[string][]byte{},
			wantStatus: StatusWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			mock.EnvVars = tt.envVars
			if tt.files != nil {
				mock.Files = tt.files
			}

			v := &WindowsValidator{sys: mock}
			check := v.checkGitBash()

			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q (message: %s)", check.Status, tt.wantStatus, check.Message)
			}
		})
	}
}

// TestWindowsValidator_checkWSL2는 WSL2 확인을 테스트한다.
func TestWindowsValidator_checkWSL2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		output     string
		err        error
		wantStatus CheckStatus
	}{
		{
			name:       "WSL 활성화됨",
			output:     "Default Version: 2",
			wantStatus: StatusOK,
		},
		{
			name:       "WSL 미설치",
			err:        fmt.Errorf("wsl not found"),
			wantStatus: StatusOK, // WSL 미설치도 OK로 처리
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			if tt.err != nil {
				mock.CommandErrors["wsl.exe --status"] = tt.err
			} else {
				mock.Commands["wsl.exe --status"] = tt.output
			}

			v := &WindowsValidator{sys: mock}
			check := v.checkWSL2()

			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", check.Status, tt.wantStatus)
			}
		})
	}
}

// TestWindowsValidator_checkLongPathsEnabled는 LongPaths 레지스트리 확인을 테스트한다.
func TestWindowsValidator_checkLongPathsEnabled(t *testing.T) {
	t.Parallel()

	regKey := `reg query HKLM\SYSTEM\CurrentControlSet\Control\FileSystem /v LongPathsEnabled`

	tests := []struct {
		name       string
		output     string
		err        error
		wantStatus CheckStatus
	}{
		{
			name:       "LongPaths 활성화됨",
			output:     "LongPathsEnabled    REG_DWORD    0x1",
			wantStatus: StatusOK,
		},
		{
			name:       "LongPaths 비활성화됨",
			output:     "LongPathsEnabled    REG_DWORD    0x0",
			wantStatus: StatusWarn,
		},
		{
			name:       "레지스트리 접근 실패",
			err:        fmt.Errorf("access denied"),
			wantStatus: StatusWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			if tt.err != nil {
				mock.CommandErrors[regKey] = tt.err
			} else {
				mock.Commands[regKey] = tt.output
			}

			v := &WindowsValidator{sys: mock}
			check := v.checkLongPathsEnabled()

			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q (message: %s)", check.Status, tt.wantStatus, check.Message)
			}
		})
	}
}

// TestWindowsValidator_checkHookBash는 Hook Bash 확인을 테스트한다.
func TestWindowsValidator_checkHookBash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		files      map[string][]byte
		commands   map[string]string
		wantStatus CheckStatus
	}{
		{
			name:       "파일 경로로 bash 발견",
			files:      map[string][]byte{`C:\Program Files\Git\bin\bash.exe`: {}},
			wantStatus: StatusOK,
		},
		{
			name:       "PATH에서 bash 발견",
			files:      map[string][]byte{},
			commands:   map[string]string{"which bash": "/usr/bin/bash"},
			wantStatus: StatusOK,
		},
		{
			name:       "bash를 찾을 수 없음",
			files:      map[string][]byte{},
			commands:   map[string]string{},
			wantStatus: StatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			mock.Files = tt.files
			if tt.commands != nil {
				mock.Commands = tt.commands
			}
			// which bash가 실패하도록 기본 에러 설정 (찾을 수 없는 경우)
			if _, ok := tt.commands["which bash"]; !ok {
				mock.CommandErrors["which bash"] = fmt.Errorf("not found")
			}

			v := &WindowsValidator{sys: mock}
			check := v.checkHookBash()

			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q (message: %s)", check.Status, tt.wantStatus, check.Message)
			}
		})
	}
}

// TestDarwinValidator_RunChecks는 macOS 검증기의 전체 체크를 확인한다.
func TestDarwinValidator_RunChecks(t *testing.T) {
	t.Parallel()

	t.Run("모든 체크 포함 확인", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.GOOSVal = "darwin"
		mock.GOARCHVal = "arm64"
		mock.Files["/opt/homebrew/bin/brew"] = []byte{}
		mock.Commands["/opt/homebrew/bin/brew --version"] = "Homebrew 4.3.0"
		mock.EnvVars["SHELL"] = "/bin/zsh"

		// symlink 확인
		mock.Commands["readlink -f /usr/local/bin/node"] = "/opt/homebrew/bin/node"
		mock.Commands["readlink -f /usr/local/bin/python3"] = "/opt/homebrew/bin/python3"

		// 공통 도구
		mock.Commands["go version"] = "go1.26"
		mock.Commands["node --version"] = "v22"
		mock.Commands["git --version"] = "git 2.44"
		mock.Commands["ae --version"] = "ae 1.0"

		v := &DarwinValidator{sys: mock}
		checks := v.RunChecks()

		// 최소 7개 체크: Homebrew, Symlink(node), Symlink(python3), Shell, Go, Node, Git, AE
		if len(checks) < 7 {
			t.Fatalf("RunChecks() returned %d checks, want >= 7", len(checks))
		}

		// 체크 이름 확인
		checkNames := make(map[string]bool)
		for _, c := range checks {
			checkNames[c.Name] = true
		}

		requiredNames := []string{"Homebrew", "Shell", "Go", "Node.js", "Git", "AE"}
		for _, name := range requiredNames {
			if !checkNames[name] {
				t.Errorf("필수 체크 %q 가 결과에 없음", name)
			}
		}
	})
}

// TestDarwinValidator_checkHomebrew는 Homebrew 확인을 테스트한다.
func TestDarwinValidator_checkHomebrew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		arch       string
		files      map[string][]byte
		commands   map[string]string
		cmdErrors  map[string]error
		wantStatus CheckStatus
	}{
		{
			name:       "ARM64: /opt/homebrew에서 발견",
			arch:       "arm64",
			files:      map[string][]byte{"/opt/homebrew/bin/brew": {}},
			commands:   map[string]string{"/opt/homebrew/bin/brew --version": "Homebrew 4.3.0"},
			wantStatus: StatusOK,
		},
		{
			name:       "AMD64: /usr/local에서 발견",
			arch:       "amd64",
			files:      map[string][]byte{"/usr/local/bin/brew": {}},
			commands:   map[string]string{"/usr/local/bin/brew --version": "Homebrew 4.3.0"},
			wantStatus: StatusOK,
		},
		{
			name:       "PATH에서 brew 발견 (폴백)",
			arch:       "arm64",
			files:      map[string][]byte{},
			commands:   map[string]string{"brew --version": "Homebrew 4.3.0"},
			wantStatus: StatusOK,
		},
		{
			name:       "Homebrew 미설치",
			arch:       "arm64",
			files:      map[string][]byte{},
			commands:   map[string]string{},
			cmdErrors:  map[string]error{"brew --version": fmt.Errorf("not found")},
			wantStatus: StatusWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			mock.GOARCHVal = tt.arch
			mock.Files = tt.files
			mock.Commands = tt.commands
			if tt.cmdErrors != nil {
				mock.CommandErrors = tt.cmdErrors
			}

			v := &DarwinValidator{sys: mock}
			check := v.checkHomebrew()

			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q (message: %s)", check.Status, tt.wantStatus, check.Message)
			}
		})
	}
}

// TestDarwinValidator_checkSymlinks는 심볼릭 링크 확인을 테스트한다.
func TestDarwinValidator_checkSymlinks(t *testing.T) {
	t.Parallel()

	t.Run("심볼릭 링크 존재", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.Commands["readlink -f /usr/local/bin/node"] = "/opt/homebrew/bin/node"
		mock.Commands["readlink -f /usr/local/bin/python3"] = "/opt/homebrew/bin/python3"

		v := &DarwinValidator{sys: mock}
		checks := v.checkSymlinks()

		if len(checks) != 2 {
			t.Fatalf("checkSymlinks() returned %d checks, want 2", len(checks))
		}

		for _, c := range checks {
			if c.Status != StatusOK {
				t.Errorf("check %q Status = %q, want OK", c.Name, c.Status)
			}
			if c.Message == "" {
				t.Errorf("check %q Message is empty", c.Name)
			}
		}
	})

	t.Run("심볼릭 링크 없음 (직접 설치)", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.CommandErrors["readlink -f /usr/local/bin/node"] = fmt.Errorf("no such file")
		mock.CommandErrors["readlink -f /usr/local/bin/python3"] = fmt.Errorf("no such file")

		v := &DarwinValidator{sys: mock}
		checks := v.checkSymlinks()

		for _, c := range checks {
			if c.Status != StatusOK {
				t.Errorf("check %q Status = %q, want OK (직접 설치도 OK)", c.Name, c.Status)
			}
		}
	})
}

// TestDarwinValidator_checkShellCompat는 셸 호환성 확인을 테스트한다.
func TestDarwinValidator_checkShellCompat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		shell      string
		wantStatus CheckStatus
		wantDetail string // 포함되어야 하는 Detail
	}{
		{
			name:       "zsh (기본 셸)",
			shell:      "/bin/zsh",
			wantStatus: StatusOK,
			wantDetail: "기본 셸 (macOS Catalina+)",
		},
		{
			name:       "bash",
			shell:      "/bin/bash",
			wantStatus: StatusOK,
			wantDetail: "Bash 셸",
		},
		{
			name:       "fish 셸",
			shell:      "/usr/local/bin/fish",
			wantStatus: StatusOK,
			wantDetail: "", // fish에 대한 특별한 Detail 없음
		},
		{
			name:       "SHELL 미설정",
			shell:      "",
			wantStatus: StatusWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := NewMockSystemInfo()
			if tt.shell != "" {
				mock.EnvVars["SHELL"] = tt.shell
			}

			v := &DarwinValidator{sys: mock}
			check := v.checkShellCompat()

			if check.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", check.Status, tt.wantStatus)
			}
			if tt.wantDetail != "" && check.Detail != tt.wantDetail {
				t.Errorf("Detail = %q, want %q", check.Detail, tt.wantDetail)
			}
		})
	}
}

// TestValidatorInterface는 모든 Validator 구현체가 인터페이스를 충족하는지 확인한다.
func TestValidatorInterface(t *testing.T) {
	t.Parallel()

	mock := NewMockSystemInfo()
	var _ Validator = &GenericValidator{sys: mock, platform: "linux"}
	var _ Validator = &WindowsValidator{sys: mock}
	var _ Validator = &DarwinValidator{sys: mock}
}
