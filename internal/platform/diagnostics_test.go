package platform

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestCollectToolVersions는 도구 버전 수집이 올바르게 동작하는지 확인한다.
func TestCollectToolVersions(t *testing.T) {
	t.Parallel()

	t.Run("모든 도구가 설치된 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.Commands["go version"] = "go version go1.26 windows/amd64"
		mock.Commands["node --version"] = "v22.0.0"
		mock.Commands["git --version"] = "git version 2.44.0"
		mock.Commands["ae --version"] = "ae v1.0.0"

		versions := collectToolVersions(mock)

		expected := map[string]string{
			"go":   "go version go1.26 windows/amd64",
			"node": "v22.0.0",
			"git":  "git version 2.44.0",
			"ae":   "ae v1.0.0",
		}

		for tool, wantVer := range expected {
			if got, ok := versions[tool]; !ok {
				t.Errorf("도구 %q 가 결과에 없음", tool)
			} else if got != wantVer {
				t.Errorf("versions[%q] = %q, want %q", tool, got, wantVer)
			}
		}
	})

	t.Run("일부 도구가 미설치인 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.Commands["go version"] = "go version go1.26 linux/amd64"
		mock.Commands["git --version"] = "git version 2.44.0"
		// node와 ae는 Commands에 없으므로 ExecCommand가 에러 반환

		versions := collectToolVersions(mock)

		if versions["go"] != "go version go1.26 linux/amd64" {
			t.Errorf("versions[go] = %q", versions["go"])
		}
		if versions["node"] != "not found" {
			t.Errorf("versions[node] = %q, want 'not found'", versions["node"])
		}
		if versions["ae"] != "not found" {
			t.Errorf("versions[ae] = %q, want 'not found'", versions["ae"])
		}
	})

	t.Run("모든 도구가 미설치인 경우", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		// Commands가 비어있음

		versions := collectToolVersions(mock)

		for _, tool := range []string{"go", "node", "git", "ae"} {
			if versions[tool] != "not found" {
				t.Errorf("versions[%q] = %q, want 'not found'", tool, versions[tool])
			}
		}
	})
}

// TestFormatDiagnostics는 진단 결과 텍스트 포맷이 올바른지 확인한다.
func TestFormatDiagnostics(t *testing.T) {
	t.Parallel()

	timestamp := time.Date(2026, 3, 24, 12, 0, 0, 0, time.UTC)

	t.Run("기본 포맷 (verbose=false)", func(t *testing.T) {
		t.Parallel()
		profile := &PlatformProfile{
			Platform:  "windows",
			Timestamp: timestamp,
			Checks: []PlatformCheck{
				{Name: "Go", Status: StatusOK, Message: "go1.26", Detail: "상세 정보"},
				{Name: "Node.js", Status: StatusWarn, Message: "미설치"},
				{Name: "Git", Status: StatusFail, Message: "에러 발생"},
			},
			ToolVersions: map[string]string{"go": "go1.26"},
		}

		output := FormatDiagnostics(profile, false)

		// 필수 요소 포함 확인
		requiredParts := []string{
			"Platform: windows",
			"=== Diagnostics ===",
			"[OK]",
			"[!!]",
			"[XX]",
			"Go",
			"Node.js",
			"Git",
			"1 passed / 1 warnings / 1 failed",
			"=== Tool Versions ===",
		}
		for _, part := range requiredParts {
			if !strings.Contains(output, part) {
				t.Errorf("출력에 %q 가 없음", part)
			}
		}

		// verbose=false이면 Detail이 출력되지 않아야 함
		if strings.Contains(output, "상세 정보") {
			t.Error("verbose=false인데 Detail이 출력됨")
		}
	})

	t.Run("verbose 모드 (Detail 포함)", func(t *testing.T) {
		t.Parallel()
		profile := &PlatformProfile{
			Platform:  "darwin",
			Timestamp: timestamp,
			Checks: []PlatformCheck{
				{Name: "Homebrew", Status: StatusOK, Message: "4.0.0", Detail: "Path: /opt/homebrew"},
			},
			ToolVersions: map[string]string{},
		}

		output := FormatDiagnostics(profile, true)

		if !strings.Contains(output, "Path: /opt/homebrew") {
			t.Error("verbose=true인데 Detail이 출력되지 않음")
		}
	})

	t.Run("빈 체크 리스트", func(t *testing.T) {
		t.Parallel()
		profile := &PlatformProfile{
			Platform:     "linux",
			Timestamp:    timestamp,
			Checks:       []PlatformCheck{},
			ToolVersions: map[string]string{},
		}

		output := FormatDiagnostics(profile, false)

		if !strings.Contains(output, "0 passed / 0 warnings / 0 failed") {
			t.Errorf("빈 체크 리스트에 대한 카운트가 잘못됨: %s", output)
		}
	})

	t.Run("상태별 카운트 정확성", func(t *testing.T) {
		t.Parallel()
		profile := &PlatformProfile{
			Platform:  "windows",
			Timestamp: timestamp,
			Checks: []PlatformCheck{
				{Name: "A", Status: StatusOK, Message: "ok"},
				{Name: "B", Status: StatusOK, Message: "ok"},
				{Name: "C", Status: StatusOK, Message: "ok"},
				{Name: "D", Status: StatusWarn, Message: "warn"},
				{Name: "E", Status: StatusWarn, Message: "warn"},
				{Name: "F", Status: StatusFail, Message: "fail"},
			},
			ToolVersions: map[string]string{},
		}

		output := FormatDiagnostics(profile, false)

		expected := "3 passed / 2 warnings / 1 failed"
		if !strings.Contains(output, expected) {
			t.Errorf("카운트 문자열이 잘못됨, want %q in output:\n%s", expected, output)
		}
	})
}

// TestFormatDiff는 프로필 차이 포맷이 올바른지 확인한다.
func TestFormatDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		diff     *ProfileDiff
		contains []string // 출력에 포함되어야 하는 문자열
		absent   []string // 출력에 포함되지 않아야 하는 문자열
	}{
		{
			name:     "nil diff",
			diff:     nil,
			contains: []string{"이전 프로필과 동일합니다"},
		},
		{
			name: "변경사항 없는 diff",
			diff: &ProfileDiff{
				ChangedTools: make(map[string][2]string),
			},
			contains: []string{"이전 프로필과 동일합니다"},
		},
		{
			name: "PATH 추가만",
			diff: &ProfileDiff{
				AddedPaths:   []string{"/new/bin"},
				ChangedTools: make(map[string][2]string),
			},
			contains: []string{"Changes from Previous Profile", "Added paths", "+ /new/bin"},
			absent:   []string{"Removed paths", "Changed tools", "Status changes"},
		},
		{
			name: "PATH 삭제만",
			diff: &ProfileDiff{
				RemovedPaths: []string{"/old/bin"},
				ChangedTools: make(map[string][2]string),
			},
			contains: []string{"Removed paths", "- /old/bin"},
			absent:   []string{"Added paths"},
		},
		{
			name: "도구 버전 변경",
			diff: &ProfileDiff{
				ChangedTools: map[string][2]string{
					"go": {"1.25", "1.26"},
				},
			},
			contains: []string{"Changed tools", "go:", "1.25", "->", "1.26"},
		},
		{
			name: "상태 변경",
			diff: &ProfileDiff{
				ChangedTools: make(map[string][2]string),
				StatusDiffs: []CheckStatusDiff{
					{Name: "Node", OldStatus: StatusOK, NewStatus: StatusFail},
				},
			},
			contains: []string{"Status changes", "Node:", "ok", "->", "fail"},
		},
		{
			name: "복합 변경사항",
			diff: &ProfileDiff{
				AddedPaths:   []string{"/new"},
				RemovedPaths: []string{"/old"},
				ChangedTools: map[string][2]string{"go": {"1.25", "1.26"}},
				StatusDiffs: []CheckStatusDiff{
					{Name: "AE", OldStatus: StatusWarn, NewStatus: StatusOK},
				},
			},
			contains: []string{
				"Added paths", "+ /new",
				"Removed paths", "- /old",
				"Changed tools", "go:",
				"Status changes", "AE:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := FormatDiff(tt.diff)

			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("출력에 %q 가 없음\n출력:\n%s", s, output)
				}
			}
			for _, s := range tt.absent {
				if strings.Contains(output, s) {
					t.Errorf("출력에 %q 가 있으면 안 됨\n출력:\n%s", s, output)
				}
			}
		})
	}
}

// TestStatusIconText는 상태 아이콘 텍스트 변환이 올바른지 확인한다.
func TestStatusIconText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status CheckStatus
		want   string
	}{
		{"StatusOK", StatusOK, "[OK]"},
		{"StatusWarn", StatusWarn, "[!!]"},
		{"StatusFail", StatusFail, "[XX]"},
		{"알 수 없는 상태", CheckStatus("unknown"), "[??]"},
		{"빈 상태", CheckStatus(""), "[??]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := statusIconText(tt.status)
			if got != tt.want {
				t.Errorf("statusIconText(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

// TestRunDiagnostics는 진단 실행이 올바르게 동작하는지 확인한다.
func TestRunDiagnostics(t *testing.T) {
	t.Parallel()

	t.Run("기본 진단 실행 (generic 플랫폼)", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.GOOSVal = "linux"
		mock.Commands["go version"] = "go version go1.26 linux/amd64"
		mock.Commands["node --version"] = "v22.0.0"
		mock.Commands["git --version"] = "git version 2.44.0"
		mock.Commands["ae --version"] = "ae v1.0.0"


		profile, err := RunDiagnostics(mock, "linux", "/usr/bin:/bin", nil)
		if err != nil {
			t.Fatalf("RunDiagnostics() error = %v", err)
		}

		if profile == nil {
			t.Fatal("RunDiagnostics() returned nil profile")
		}
		if profile.Platform != "linux" {
			t.Errorf("Platform = %q, want %q", profile.Platform, "linux")
		}

		// 도구 버전 확인
		if profile.ToolVersions["go"] != "go version go1.26 linux/amd64" {
			t.Errorf("ToolVersions[go] = %q", profile.ToolVersions["go"])
		}

		// Timestamp가 설정되었는지 확인
		if profile.Timestamp.IsZero() {
			t.Error("Timestamp가 zero value")
		}
	})

	t.Run("Windows 플랫폼 진단", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.GOOSVal = "windows"
		// Windows 전용 체크에 필요한 명령어 mock
		mock.Commands["cmd.exe /c chcp"] = "Active code page: 65001"
		mock.Commands["where.exe npx"] = "C:\\nodejs\\npx"
		mock.Commands["where.exe npx.cmd"] = "C:\\nodejs\\npx.cmd"
		mock.Commands["where.exe pwsh.exe"] = "C:\\Windows\\pwsh.exe"
		mock.Commands["wsl.exe --status"] = "WSL version 2"
		mock.Commands["reg query HKLM\\SYSTEM\\CurrentControlSet\\Control\\FileSystem /v LongPathsEnabled"] = "LongPathsEnabled    REG_DWORD    0x1"
		mock.Files[`C:\Program Files\Git\bin\bash.exe`] = []byte{}
		mock.EnvVars["MSYSTEM"] = "MINGW64"

		// 공통 도구
		mock.Commands["go version"] = "go version go1.26"
		mock.Commands["node --version"] = "v22.0.0"
		mock.Commands["git --version"] = "git version 2.44.0"
		mock.Commands["ae --version"] = "ae v1.0.0"

		profile, err := RunDiagnostics(mock, "windows", `C:\Windows\system32`, nil)
		if err != nil {
			t.Fatalf("RunDiagnostics() error = %v", err)
		}

		if profile.Platform != "windows" {
			t.Errorf("Platform = %q, want 'windows'", profile.Platform)
		}

		// Windows 진단에는 최소 10개 이상의 체크가 있어야 함
		// (UTF-8, MCP x3, Git Bash, WSL2, LongPaths, Hook Bash, Go, Node, Git, AE)
		if len(profile.Checks) < 10 {
			t.Errorf("Windows 체크 수 = %d, want >= 10", len(profile.Checks))
		}
	})

	t.Run("darwin 플랫폼 진단", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.GOOSVal = "darwin"
		mock.GOARCHVal = "arm64"
		mock.Files["/opt/homebrew/bin/brew"] = []byte{}
		mock.Commands["/opt/homebrew/bin/brew --version"] = "Homebrew 4.3.0"
		mock.Commands["readlink -f /usr/local/bin/node"] = "/opt/homebrew/bin/node"
		mock.Commands["readlink -f /usr/local/bin/python3"] = "/opt/homebrew/bin/python3"
		mock.EnvVars["SHELL"] = "/bin/zsh"

		// 공통 도구
		mock.Commands["go version"] = "go version go1.26 darwin/arm64"
		mock.Commands["node --version"] = "v22.0.0"
		mock.Commands["git --version"] = "git version 2.44.0"
		mock.Commands["ae --version"] = "ae v1.0.0"

		profile, err := RunDiagnostics(mock, "darwin", "/opt/homebrew/bin:/usr/bin", nil)
		if err != nil {
			t.Fatalf("RunDiagnostics() error = %v", err)
		}

		if profile.Platform != "darwin" {
			t.Errorf("Platform = %q, want 'darwin'", profile.Platform)
		}

		// darwin 진단에는 최소 7개 이상의 체크가 있어야 함
		// (Homebrew, Symlink x2, Shell, Go, Node, Git, AE)
		if len(profile.Checks) < 7 {
			t.Errorf("darwin 체크 수 = %d, want >= 7", len(profile.Checks))
		}
	})

	t.Run("프로필에 PATH 항목이 포함됨", func(t *testing.T) {
		t.Parallel()
		mock := NewMockSystemInfo()
		mock.GOOSVal = "linux"
		mock.Commands["go version"] = "go1.26"
		mock.Commands["node --version"] = "v22"
		mock.Commands["git --version"] = "git 2.44"
		mock.Commands["ae --version"] = "ae 1.0"

		profile, err := RunDiagnostics(mock, "linux", "/usr/bin:/bin", nil)
		if err != nil {
			t.Fatalf("RunDiagnostics() error = %v", err)
		}

		// PATH 항목이 포함되어 있어야 함 (BuildSmartPATH 출력 기반)
		if profile.PATH == nil {
			t.Error("PATH가 nil")
		}
	})
}

// TestFormatDiagnostics_LabelAlignment는 라벨 정렬이 올바른지 확인한다.
func TestFormatDiagnostics_LabelAlignment(t *testing.T) {
	t.Parallel()

	profile := &PlatformProfile{
		Platform:  "test",
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Checks: []PlatformCheck{
			{Name: "A", Status: StatusOK, Message: "ok"},
			{Name: "LongName", Status: StatusOK, Message: "ok"},
		},
		ToolVersions: map[string]string{},
	}

	output := FormatDiagnostics(profile, false)

	// 두 줄 모두 같은 위치에 메시지가 정렬되어야 함
	lines := strings.Split(output, "\n")
	var diagLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "[OK]") {
			diagLines = append(diagLines, line)
		}
	}

	if len(diagLines) != 2 {
		t.Fatalf("진단 라인 수 = %d, want 2", len(diagLines))
	}

	// "ok" 텍스트의 위치가 같은지 확인 (정렬)
	idx0 := strings.LastIndex(diagLines[0], "ok")
	idx1 := strings.LastIndex(diagLines[1], "ok")
	if idx0 != idx1 {
		t.Errorf("메시지 정렬이 맞지 않음: line1 ok@%d, line2 ok@%d\nline1: %q\nline2: %q",
			idx0, idx1, diagLines[0], diagLines[1])
	}
}

// TestFormatDiagnostics_ToolVersions는 도구 버전 섹션 출력을 확인한다.
func TestFormatDiagnostics_ToolVersions(t *testing.T) {
	t.Parallel()

	profile := &PlatformProfile{
		Platform:  "test",
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Checks:    []PlatformCheck{},
		ToolVersions: map[string]string{
			"go":   "go1.26",
			"node": "v22.0.0",
		},
	}

	output := FormatDiagnostics(profile, false)

	// 각 도구 버전이 출력에 포함되어야 함
	for tool, ver := range profile.ToolVersions {
		searchStr := fmt.Sprintf("%s:", tool)
		if !strings.Contains(output, searchStr) {
			t.Errorf("출력에 도구 %q 가 없음", tool)
		}
		if !strings.Contains(output, ver) {
			t.Errorf("출력에 버전 %q 가 없음", ver)
		}
	}
}
