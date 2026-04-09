package version

import (
	"runtime/debug"
	"strings"
	"sync"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Default value test (version is set at build time)
	got := GetVersion()
	if got == "" {
		t.Errorf("GetVersion() = %q, want non-empty string", got)
	}
}

func TestGetCommit(t *testing.T) {
	// Default value test
	got := GetCommit()
	if got != "none" {
		t.Errorf("GetCommit() = %q, want %q (default)", got, "none")
	}
}

func TestGetDate(t *testing.T) {
	// Default value test
	got := GetDate()
	if got != "unknown" {
		t.Errorf("GetDate() = %q, want %q (default)", got, "unknown")
	}
}

func TestGetFullVersion(t *testing.T) {
	got := GetFullVersion()

	// Verify format: "VERSION (commit: COMMIT, built: DATE)"
	if !strings.Contains(got, Version) {
		t.Errorf("GetFullVersion() = %q, should contain Version %q", got, Version)
	}
	if !strings.Contains(got, Commit) {
		t.Errorf("GetFullVersion() = %q, should contain Commit %q", got, Commit)
	}
	if !strings.Contains(got, Date) {
		t.Errorf("GetFullVersion() = %q, should contain Date %q", got, Date)
	}
	if !strings.Contains(got, "commit:") {
		t.Errorf("GetFullVersion() = %q, should contain 'commit:'", got)
	}
	if !strings.Contains(got, "built:") {
		t.Errorf("GetFullVersion() = %q, should contain 'built:'", got)
	}
}

func TestGetFullVersion_Format(t *testing.T) {
	// Test exact format with current values
	expected := Version + " (commit: " + Commit + ", built: " + Date + ")"
	got := GetFullVersion()
	if got != expected {
		t.Errorf("GetFullVersion() = %q, want %q", got, expected)
	}
}

func TestVersionVariables_Defaults(t *testing.T) {
	// Test that version variables are set (not empty)
	tests := []struct {
		name     string
		variable string
		notEmpty bool
	}{
		{"Version", Version, true},
		{"Commit", Commit, true},
		{"Date", Date, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.notEmpty && tt.variable == "" {
				t.Errorf("%s = %q, want non-empty string", tt.name, tt.variable)
			}
		})
	}
}

func TestVersionVariables_Modifiable(t *testing.T) {
	// sync.Once 포함 전체 상태 저장 및 리셋
	restore := saveAndReset(t)
	defer restore()

	// Modify values (simulating ldflags injection)
	Version = "1.0.0"
	Commit = "abc123"
	Date = "2026-01-15"

	// Verify getters return modified values
	if GetVersion() != "1.0.0" {
		t.Errorf("GetVersion() after modification = %q, want %q", GetVersion(), "1.0.0")
	}
	if GetCommit() != "abc123" {
		t.Errorf("GetCommit() after modification = %q, want %q", GetCommit(), "abc123")
	}
	if GetDate() != "2026-01-15" {
		t.Errorf("GetDate() after modification = %q, want %q", GetDate(), "2026-01-15")
	}

	// Verify full version format with modified values
	expected := "1.0.0 (commit: abc123, built: 2026-01-15)"
	if got := GetFullVersion(); got != expected {
		t.Errorf("GetFullVersion() with modified values = %q, want %q", got, expected)
	}
}

// saveAndReset은 패키지 변수를 저장하고 sync.Once를 초기화하는 테스트 헬퍼.
// 반환된 함수를 defer로 호출하면 원래 상태로 복원됨.
func saveAndReset(t *testing.T) func() {
	t.Helper()
	origVersion := Version
	origCommit := Commit
	origDate := Date
	origOnce := once
	origReadBuildInfo := readBuildInfo
	return func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
		once = origOnce
		readBuildInfo = origReadBuildInfo
	}
}

// newOnce는 새 sync.Once 포인터를 생성하는 헬퍼.
func newOnce() *sync.Once {
	return &sync.Once{}
}

// mockBuildInfo는 테스트용 BuildInfo를 생성하는 헬퍼.
func mockBuildInfo(moduleVersion string, settings []debug.BuildSetting) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{
				Version: moduleVersion,
			},
			Settings: settings,
		}, true
	}
}

// AC-004: Version "v1.0.0" + BuildInfo module "v1.2.4" → GetVersion()이 sync.Once 폴백을 트리거하여 "v1.2.4" 반환
func TestResolveFromBuildInfo_FallbackVersion(t *testing.T) {
	restore := saveAndReset(t)
	defer restore()

	Version = defaultVersion // "v1.0.0" (기본값)
	Commit = "none"
	Date = "unknown"
	once = newOnce()
	readBuildInfo = mockBuildInfo("v1.2.4", []debug.BuildSetting{
		{Key: "vcs.revision", Value: "abc123def456"},
		{Key: "vcs.time", Value: "2026-04-01T12:00:00Z"},
	})

	got := GetVersion()
	if got != "v1.2.4" {
		t.Errorf("GetVersion() = %q, want %q (BuildInfo 폴백 기대)", got, "v1.2.4")
	}
	if GetCommit() != "abc123def456" {
		t.Errorf("GetCommit() = %q, want %q (vcs.revision 폴백 기대)", GetCommit(), "abc123def456")
	}
	if GetDate() != "2026-04-01T12:00:00Z" {
		t.Errorf("GetDate() = %q, want %q (vcs.time 폴백 기대)", GetDate(), "2026-04-01T12:00:00Z")
	}
}

// AC-005: ldflags "v2.0.0" (기본값 아님) + BuildInfo "v1.2.4" → GetVersion() = "v2.0.0" (폴백 없음)
func TestResolveFromBuildInfo_NoFallbackWhenLdflagsSet(t *testing.T) {
	restore := saveAndReset(t)
	defer restore()

	Version = "v2.0.0" // ldflags로 주입된 값
	Commit = "release-commit-hash"
	Date = "2026-03-15"
	once = newOnce()
	readBuildInfo = mockBuildInfo("v1.2.4", []debug.BuildSetting{
		{Key: "vcs.revision", Value: "other-commit"},
		{Key: "vcs.time", Value: "2026-01-01T00:00:00Z"},
	})

	got := GetVersion()
	if got != "v2.0.0" {
		t.Errorf("GetVersion() = %q, want %q (ldflags 값 유지 기대)", got, "v2.0.0")
	}
	if GetCommit() != "release-commit-hash" {
		t.Errorf("GetCommit() = %q, want %q (ldflags 값 유지 기대)", GetCommit(), "release-commit-hash")
	}
	if GetDate() != "2026-03-15" {
		t.Errorf("GetDate() = %q, want %q (ldflags 값 유지 기대)", GetDate(), "2026-03-15")
	}
}

// 각 필드(Version/Commit/Date)는 독립적으로 폴백되어야 함
func TestResolveFromBuildInfo_IndependentFieldFallback(t *testing.T) {
	restore := saveAndReset(t)
	defer restore()

	// Version은 ldflags로 설정됨, Commit/Date는 기본값
	Version = "v3.0.0"
	Commit = "none"
	Date = "unknown"
	once = newOnce()
	readBuildInfo = mockBuildInfo("v1.2.4", []debug.BuildSetting{
		{Key: "vcs.revision", Value: "independent-commit"},
		{Key: "vcs.time", Value: "2026-02-20T08:30:00Z"},
	})

	_ = GetVersion() // sync.Once 트리거

	// Version은 ldflags 값 유지 (v3.0.0 != defaultVersion이므로 폴백 안 함)
	if Version != "v3.0.0" {
		t.Errorf("Version = %q, want %q (ldflags 값 유지 기대)", Version, "v3.0.0")
	}
	// Commit은 기본값이므로 vcs.revision으로 폴백
	if Commit != "independent-commit" {
		t.Errorf("Commit = %q, want %q (독립 폴백 기대)", Commit, "independent-commit")
	}
	// Date는 기본값이므로 vcs.time으로 폴백
	if Date != "2026-02-20T08:30:00Z" {
		t.Errorf("Date = %q, want %q (독립 폴백 기대)", Date, "2026-02-20T08:30:00Z")
	}
}

// AC-008: ldflags 미주입 + ReadBuildInfo 실패 → 기본값 유지
func TestResolveFromBuildInfo_ReadBuildInfoFails(t *testing.T) {
	restore := saveAndReset(t)
	defer restore()

	Version = defaultVersion
	Commit = "none"
	Date = "unknown"
	once = newOnce()
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return nil, false // ReadBuildInfo 실패
	}

	got := GetVersion()
	if got != defaultVersion {
		t.Errorf("GetVersion() = %q, want %q (ReadBuildInfo 실패 시 기본값 유지)", got, defaultVersion)
	}
	if GetCommit() != "none" {
		t.Errorf("GetCommit() = %q, want %q (기본값 유지 기대)", GetCommit(), "none")
	}
	if GetDate() != "unknown" {
		t.Errorf("GetDate() = %q, want %q (기본값 유지 기대)", GetDate(), "unknown")
	}
}

// GetVersion()이 sync.Once를 트리거하여 한 번만 resolveFromBuildInfo 실행
func TestGetVersion_TriggersSyncOnce(t *testing.T) {
	restore := saveAndReset(t)
	defer restore()

	callCount := 0
	Version = defaultVersion
	Commit = "none"
	Date = "unknown"
	once = newOnce()
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		callCount++
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v9.9.9"},
		}, true
	}

	// 첫 번째 호출: resolveFromBuildInfo 실행
	_ = GetVersion()
	if callCount != 1 {
		t.Errorf("readBuildInfo 호출 횟수 = %d, want 1 (첫 호출에서 실행 기대)", callCount)
	}

	// 두 번째 호출: sync.Once로 인해 재실행 안 됨
	_ = GetVersion()
	if callCount != 1 {
		t.Errorf("readBuildInfo 호출 횟수 = %d, want 1 (sync.Once로 재실행 방지 기대)", callCount)
	}
}

// BuildInfo의 모듈 버전이 "(devel)"이면 폴백하지 않음
func TestResolveFromBuildInfo_SkipDevelVersion(t *testing.T) {
	restore := saveAndReset(t)
	defer restore()

	Version = defaultVersion
	Commit = "none"
	Date = "unknown"
	once = newOnce()
	readBuildInfo = mockBuildInfo("(devel)", []debug.BuildSetting{
		{Key: "vcs.revision", Value: "dev-commit"},
		{Key: "vcs.time", Value: "2026-04-01T00:00:00Z"},
	})

	got := GetVersion()
	if got != defaultVersion {
		t.Errorf("GetVersion() = %q, want %q ((devel) 버전은 폴백 스킵 기대)", got, defaultVersion)
	}
	// Commit/Date는 기본값이므로 vcs 정보로 독립 폴백
	if GetCommit() != "dev-commit" {
		t.Errorf("GetCommit() = %q, want %q (vcs.revision 폴백 기대)", GetCommit(), "dev-commit")
	}
	if GetDate() != "2026-04-01T00:00:00Z" {
		t.Errorf("GetDate() = %q, want %q (vcs.time 폴백 기대)", GetDate(), "2026-04-01T00:00:00Z")
	}
}

// BuildInfo의 모듈 버전이 빈 문자열이면 폴백하지 않음
func TestResolveFromBuildInfo_SkipEmptyModuleVersion(t *testing.T) {
	restore := saveAndReset(t)
	defer restore()

	Version = defaultVersion
	Commit = "none"
	Date = "unknown"
	once = newOnce()
	readBuildInfo = mockBuildInfo("", nil)

	got := GetVersion()
	if got != defaultVersion {
		t.Errorf("GetVersion() = %q, want %q (빈 모듈 버전은 폴백 스킵 기대)", got, defaultVersion)
	}
}
