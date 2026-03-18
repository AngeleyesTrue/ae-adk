package version

import "fmt"

// Build-time variables injected via -ldflags.
// Default version for RC/test builds (overridden by -ldflags in production)
var (
	Version = "v2.7.15"
	Commit  = "none"
	Date    = "unknown"
)

// @MX:ANCHOR: [AUTO] 13개 파일에서 참조하는 버전 조회 핵심 함수
// @MX:REASON: 변경 시 CLI, hook, statusline 등 전체 버전 표시에 영향
// GetVersion returns the current version string.
func GetVersion() string {
	return Version
}

// GetCommit returns the build commit hash.
func GetCommit() string {
	return Commit
}

// GetDate returns the build date.
func GetDate() string {
	return Date
}

// GetFullVersion returns a formatted full version string.
func GetFullVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}
