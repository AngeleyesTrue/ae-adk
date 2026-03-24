package platform

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// CheckStatus는 개별 진단 항목의 결과 상태를 나타낸다.
type CheckStatus string

const (
	StatusOK   CheckStatus = "ok"
	StatusWarn CheckStatus = "warn"
	StatusFail CheckStatus = "fail"
)

// PlatformCheck는 단일 플랫폼 진단 항목의 결과를 보관한다.
type PlatformCheck struct {
	Name    string      `json:"name"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
	Detail  string      `json:"detail,omitempty"`
}

// PlatformProfile은 플랫폼 진단 결과를 저장하는 프로필이다.
type PlatformProfile struct {
	Platform     string            `json:"platform"`
	Timestamp    time.Time         `json:"timestamp"`
	Checks       []PlatformCheck   `json:"checks"`
	PATH         []string          `json:"path"`
	ToolVersions map[string]string `json:"tool_versions"`
}

// ProfileDiff는 두 프로필 간의 차이점을 나타낸다.
type ProfileDiff struct {
	AddedPaths   []string          `json:"added_paths,omitempty"`
	RemovedPaths []string          `json:"removed_paths,omitempty"`
	ChangedTools map[string][2]string `json:"changed_tools,omitempty"` // key -> [old, new]
	StatusDiffs  []CheckStatusDiff `json:"status_diffs,omitempty"`
}

// CheckStatusDiff는 진단 항목의 상태 변경을 나타낸다.
type CheckStatusDiff struct {
	Name      string      `json:"name"`
	OldStatus CheckStatus `json:"old_status"`
	NewStatus CheckStatus `json:"new_status"`
}

// PlatformFlags는 ae win/mac 명령어의 공통 플래그를 보관한다.
type PlatformFlags struct {
	Force      bool
	Verbose    bool
	JSON       bool
	Auto       bool
	DryRun     bool
	SkipBackup bool
}

// PathVerifyResult는 PATH 검증 결과를 나타낸다.
type PathVerifyResult struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

// SystemInfo는 테스트 가능한 시스템 정보 인터페이스이다.
// @MX:ANCHOR: [AUTO] 플랫폼 진단의 핵심 추상화 인터페이스
// @MX:REASON: [AUTO] fan_in>=3, validator/diagnostics/cli에서 공통 사용
type SystemInfo interface {
	GOOS() string
	GOARCH() string
	HomeDir() string
	GetEnv(key string) string
	FileExists(path string) bool
	DirExists(path string) bool
	ExecCommand(name string, args ...string) (string, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
}

// DefaultSystemInfo는 실제 OS를 사용하는 SystemInfo 구현체이다.
type DefaultSystemInfo struct{}

func (d *DefaultSystemInfo) GOOS() string   { return runtime.GOOS }
func (d *DefaultSystemInfo) GOARCH() string { return runtime.GOARCH }

func (d *DefaultSystemInfo) HomeDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

func (d *DefaultSystemInfo) GetEnv(key string) string {
	return os.Getenv(key)
}

func (d *DefaultSystemInfo) FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (d *DefaultSystemInfo) DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (d *DefaultSystemInfo) ExecCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func (d *DefaultSystemInfo) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (d *DefaultSystemInfo) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}
