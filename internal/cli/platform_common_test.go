package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/AngeleyesTrue/ae-adk/internal/platform"
)

// --- mockSystemInfo는 platform.SystemInfo를 테스트용으로 구현한다 ---

type mockSystemInfo struct {
	goos     string
	goarch   string
	homeDir  string
	envVars  map[string]string
	files    map[string][]byte // 파일 내용 저장소
	dirs     map[string]bool   // 디렉터리 존재 여부
	execFunc func(name string, args ...string) (string, error)
}

func newMockSystemInfo() *mockSystemInfo {
	return &mockSystemInfo{
		goos:    runtime.GOOS,
		goarch:  runtime.GOARCH,
		homeDir: "/mock/home",
		envVars: make(map[string]string),
		files:   make(map[string][]byte),
		dirs:    make(map[string]bool),
	}
}

func (m *mockSystemInfo) GOOS() string   { return m.goos }
func (m *mockSystemInfo) GOARCH() string { return m.goarch }
func (m *mockSystemInfo) HomeDir() string { return m.homeDir }

func (m *mockSystemInfo) GetEnv(key string) string {
	return m.envVars[key]
}

func (m *mockSystemInfo) FileExists(path string) bool {
	_, ok := m.files[path]
	return ok
}

func (m *mockSystemInfo) DirExists(path string) bool {
	return m.dirs[path]
}

func (m *mockSystemInfo) ExecCommand(name string, args ...string) (string, error) {
	if m.execFunc != nil {
		return m.execFunc(name, args...)
	}
	return "", nil
}

func (m *mockSystemInfo) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (m *mockSystemInfo) WriteFile(path string, data []byte, _ os.FileMode) error {
	m.files[path] = data
	return nil
}

// --- winCmd / macCmd 등록 테스트 ---

func TestWinCmd_Exists(t *testing.T) {
	if winCmd == nil {
		t.Fatal("winCmd는 nil이 아니어야 한다")
	}
}

func TestWinCmd_Use(t *testing.T) {
	if winCmd.Use != "win" {
		t.Errorf("winCmd.Use = %q, want %q", winCmd.Use, "win")
	}
}

func TestWinCmd_Short(t *testing.T) {
	if winCmd.Short == "" {
		t.Error("winCmd.Short는 빈 문자열이 아니어야 한다")
	}
}

func TestWinCmd_Long(t *testing.T) {
	if winCmd.Long == "" {
		t.Error("winCmd.Long은 빈 문자열이 아니어야 한다")
	}
}

func TestWinCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "win" {
			found = true
			break
		}
	}
	if !found {
		t.Error("win 명령어가 rootCmd의 하위 명령어로 등록되어야 한다")
	}
}

func TestMacCmd_Exists(t *testing.T) {
	if macCmd == nil {
		t.Fatal("macCmd는 nil이 아니어야 한다")
	}
}

func TestMacCmd_Use(t *testing.T) {
	if macCmd.Use != "mac" {
		t.Errorf("macCmd.Use = %q, want %q", macCmd.Use, "mac")
	}
}

func TestMacCmd_Short(t *testing.T) {
	if macCmd.Short == "" {
		t.Error("macCmd.Short는 빈 문자열이 아니어야 한다")
	}
}

func TestMacCmd_Long(t *testing.T) {
	if macCmd.Long == "" {
		t.Error("macCmd.Long은 빈 문자열이 아니어야 한다")
	}
}

func TestMacCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "mac" {
			found = true
			break
		}
	}
	if !found {
		t.Error("mac 명령어가 rootCmd의 하위 명령어로 등록되어야 한다")
	}
}

// --- 플래그 등록 테스트 ---

func TestWinCmd_HasPlatformFlags(t *testing.T) {
	expectedFlags := []string{"force", "verbose", "json", "auto", "dry-run", "skip-backup"}
	for _, name := range expectedFlags {
		if winCmd.Flags().Lookup(name) == nil {
			t.Errorf("win 명령어에 --%s 플래그가 있어야 한다", name)
		}
	}
}

func TestMacCmd_HasPlatformFlags(t *testing.T) {
	expectedFlags := []string{"force", "verbose", "json", "auto", "dry-run", "skip-backup"}
	for _, name := range expectedFlags {
		if macCmd.Flags().Lookup(name) == nil {
			t.Errorf("mac 명령어에 --%s 플래그가 있어야 한다", name)
		}
	}
}

func TestWinCmd_VerboseShortFlag(t *testing.T) {
	f := winCmd.Flags().ShorthandLookup("v")
	if f == nil {
		t.Error("win 명령어에 -v 단축 플래그가 있어야 한다")
	}
}

func TestMacCmd_VerboseShortFlag(t *testing.T) {
	f := macCmd.Flags().ShorthandLookup("v")
	if f == nil {
		t.Error("mac 명령어에 -v 단축 플래그가 있어야 한다")
	}
}

// --- parsePlatformFlags 테스트 ---

// newTestCmdWithFlags는 플랫폼 플래그가 등록된 테스트용 cobra.Command를 생성한다.
func newTestCmdWithFlags() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	addPlatformFlags(cmd)
	return cmd
}

func TestParsePlatformFlags_DefaultValues(t *testing.T) {
	cmd := newTestCmdWithFlags()
	// 기본값으로 파싱 (플래그 미설정)
	flags := parsePlatformFlags(cmd)

	if flags.Force {
		t.Error("Force 기본값은 false여야 한다")
	}
	if flags.Verbose {
		t.Error("Verbose 기본값은 false여야 한다")
	}
	if flags.JSON {
		t.Error("JSON 기본값은 false여야 한다")
	}
	if flags.Auto {
		t.Error("Auto 기본값은 false여야 한다")
	}
	if flags.DryRun {
		t.Error("DryRun 기본값은 false여야 한다")
	}
	if flags.SkipBackup {
		t.Error("SkipBackup 기본값은 false여야 한다")
	}
}

func TestParsePlatformFlags_AllTrue(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.SetArgs([]string{
		"--force", "--verbose", "--json", "--auto", "--dry-run", "--skip-backup",
	})
	// Execute를 통해 플래그를 파싱
	_ = cmd.Execute()

	flags := parsePlatformFlags(cmd)

	if !flags.Force {
		t.Error("--force 설정 후 Force는 true여야 한다")
	}
	if !flags.Verbose {
		t.Error("--verbose 설정 후 Verbose는 true여야 한다")
	}
	if !flags.JSON {
		t.Error("--json 설정 후 JSON은 true여야 한다")
	}
	if !flags.Auto {
		t.Error("--auto 설정 후 Auto는 true여야 한다")
	}
	if !flags.DryRun {
		t.Error("--dry-run 설정 후 DryRun은 true여야 한다")
	}
	if !flags.SkipBackup {
		t.Error("--skip-backup 설정 후 SkipBackup은 true여야 한다")
	}
}

func TestParsePlatformFlags_PartialFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantFlag string // 검증할 필드명
	}{
		{
			name:     "force만 설정",
			args:     []string{"--force"},
			wantFlag: "Force",
		},
		{
			name:     "verbose 단축키 사용",
			args:     []string{"-v"},
			wantFlag: "Verbose",
		},
		{
			name:     "dry-run만 설정",
			args:     []string{"--dry-run"},
			wantFlag: "DryRun",
		},
		{
			name:     "skip-backup만 설정",
			args:     []string{"--skip-backup"},
			wantFlag: "SkipBackup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestCmdWithFlags()
			cmd.SetArgs(tt.args)
			_ = cmd.Execute()

			flags := parsePlatformFlags(cmd)

			switch tt.wantFlag {
			case "Force":
				if !flags.Force {
					t.Errorf("Force가 true여야 한다")
				}
			case "Verbose":
				if !flags.Verbose {
					t.Errorf("Verbose가 true여야 한다")
				}
			case "DryRun":
				if !flags.DryRun {
					t.Errorf("DryRun이 true여야 한다")
				}
			case "SkipBackup":
				if !flags.SkipBackup {
					t.Errorf("SkipBackup이 true여야 한다")
				}
			}
		})
	}
}

// --- findSettingsJSON 테스트 ---

func TestFindSettingsJSON_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"env":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	result := findSettingsJSON()
	if result == "" {
		t.Error("settings.json이 존재할 때 경로를 반환해야 한다")
	}
	if !strings.HasSuffix(result, filepath.Join(".claude", "settings.json")) {
		t.Errorf("반환 경로가 .claude/settings.json으로 끝나야 한다, got %q", result)
	}
}

func TestFindSettingsJSON_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	result := findSettingsJSON()
	if result != "" {
		t.Errorf("settings.json이 없을 때 빈 문자열을 반환해야 한다, got %q", result)
	}
}

func TestFindSettingsJSON_DirExistsButNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// .claude 디렉터리는 있지만 settings.json은 없음
	t.Chdir(tmpDir)

	result := findSettingsJSON()
	if result != "" {
		t.Errorf(".claude/는 있지만 settings.json이 없으면 빈 문자열이어야 한다, got %q", result)
	}
}

// --- updateSettingsPATH 테스트 ---

func TestUpdateSettingsPATH_CreatesEnvIfMissing(t *testing.T) {
	mock := newMockSystemInfo()
	settingsPath := "/mock/settings.json"
	// env 키가 없는 JSON
	mock.files[settingsPath] = []byte(`{"other": "value"}`)

	newPATH := "/usr/local/bin:/usr/bin"
	if err := updateSettingsPATH(mock, settingsPath, newPATH); err != nil {
		t.Fatalf("updateSettingsPATH 실패: %v", err)
	}

	// 업데이트된 파일 확인
	data, ok := mock.files[settingsPath]
	if !ok {
		t.Fatal("settings.json이 작성되어야 한다")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	env, ok := result["env"].(map[string]interface{})
	if !ok {
		t.Fatal("env 객체가 생성되어야 한다")
	}

	pathVal, ok := env["PATH"].(string)
	if !ok {
		t.Fatal("env.PATH가 문자열이어야 한다")
	}

	if runtime.GOOS == "windows" {
		// Windows에서는 슬래시가 백슬래시로 변환됨
		expected := strings.ReplaceAll(newPATH, "/", "\\")
		if pathVal != expected {
			t.Errorf("env.PATH = %q, want %q", pathVal, expected)
		}
	} else {
		if pathVal != newPATH {
			t.Errorf("env.PATH = %q, want %q", pathVal, newPATH)
		}
	}
}

func TestUpdateSettingsPATH_UpdatesExistingEnv(t *testing.T) {
	mock := newMockSystemInfo()
	settingsPath := "/mock/settings.json"
	// 기존 env.PATH가 있는 JSON
	mock.files[settingsPath] = []byte(`{
  "env": {
    "PATH": "/old/path",
    "OTHER": "keep"
  },
  "name": "test"
}`)

	newPATH := "/new/path:/another/path"
	if err := updateSettingsPATH(mock, settingsPath, newPATH); err != nil {
		t.Fatalf("updateSettingsPATH 실패: %v", err)
	}

	data := mock.files[settingsPath]
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	// 기존 필드가 보존되는지 확인
	if result["name"] != "test" {
		t.Error("기존 name 필드가 보존되어야 한다")
	}

	env := result["env"].(map[string]interface{})

	// 다른 env 변수가 보존되는지 확인
	if env["OTHER"] != "keep" {
		t.Error("기존 OTHER 환경변수가 보존되어야 한다")
	}

	pathVal := env["PATH"].(string)
	if runtime.GOOS == "windows" {
		expected := strings.ReplaceAll(newPATH, "/", "\\")
		if pathVal != expected {
			t.Errorf("env.PATH = %q, want %q", pathVal, expected)
		}
	} else {
		if pathVal != newPATH {
			t.Errorf("env.PATH = %q, want %q", pathVal, newPATH)
		}
	}
}

func TestUpdateSettingsPATH_ReadError(t *testing.T) {
	mock := newMockSystemInfo()
	// 존재하지 않는 파일
	err := updateSettingsPATH(mock, "/nonexistent/settings.json", "/new/path")
	if err == nil {
		t.Error("존재하지 않는 파일을 읽을 때 오류를 반환해야 한다")
	}
	if !strings.Contains(err.Error(), "read settings") {
		t.Errorf("오류 메시지에 'read settings'가 포함되어야 한다, got %q", err.Error())
	}
}

func TestUpdateSettingsPATH_InvalidJSON(t *testing.T) {
	mock := newMockSystemInfo()
	settingsPath := "/mock/settings.json"
	mock.files[settingsPath] = []byte(`{invalid json}`)

	err := updateSettingsPATH(mock, settingsPath, "/new/path")
	if err == nil {
		t.Error("잘못된 JSON일 때 오류를 반환해야 한다")
	}
	if !strings.Contains(err.Error(), "parse settings") {
		t.Errorf("오류 메시지에 'parse settings'가 포함되어야 한다, got %q", err.Error())
	}
}

func TestUpdateSettingsPATH_PreservesIndentation(t *testing.T) {
	mock := newMockSystemInfo()
	settingsPath := "/mock/settings.json"
	mock.files[settingsPath] = []byte(`{"env":{}}`)

	if err := updateSettingsPATH(mock, settingsPath, "/usr/bin"); err != nil {
		t.Fatalf("updateSettingsPATH 실패: %v", err)
	}

	data := mock.files[settingsPath]
	// json.MarshalIndent 결과는 들여쓰기가 있어야 함
	if !strings.Contains(string(data), "\n") {
		t.Error("출력 JSON에 줄바꿈이 포함되어야 한다 (MarshalIndent)")
	}
}

// --- runPlatformCommand 플랫폼 불일치 테스트 ---

func TestRunPlatformCommand_PlatformMismatch(t *testing.T) {
	// 현재 플랫폼과 다른 타겟으로 실행 (--force 없이)
	var targetPlatform string
	if runtime.GOOS == "windows" {
		targetPlatform = "darwin"
	} else {
		targetPlatform = "windows"
	}

	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{}) // --force 없음

	err := runPlatformCommand(cmd, targetPlatform)
	if err != nil {
		t.Fatalf("플랫폼 불일치 시 오류 대신 경고를 출력해야 한다: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--force") {
		t.Errorf("출력에 '--force' 안내가 포함되어야 한다, got %q", output)
	}
	if !strings.Contains(output, runtime.GOOS) {
		t.Errorf("출력에 현재 플랫폼(%s)이 포함되어야 한다, got %q", runtime.GOOS, output)
	}
	if !strings.Contains(output, targetPlatform) {
		t.Errorf("출력에 대상 플랫폼(%s)이 포함되어야 한다, got %q", targetPlatform, output)
	}
}

func TestRunPlatformCommand_PlatformMismatchReturnsNil(t *testing.T) {
	// 플랫폼 불일치 시 error가 아닌 nil을 반환하는지 확인
	var targetPlatform string
	if runtime.GOOS == "windows" {
		targetPlatform = "linux"
	} else {
		targetPlatform = "windows"
	}

	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := runPlatformCommand(cmd, targetPlatform)
	if err != nil {
		t.Errorf("플랫폼 불일치 시 nil을 반환해야 한다, got error: %v", err)
	}
}

// --- dry-run 모드 테스트 ---

func TestRunPlatformCommand_DryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dry-run", "--force"})
	_ = cmd.Execute()

	// --force로 플랫폼 체크 건너뛰고 dry-run 모드 진입
	err := runPlatformCommand(cmd, runtime.GOOS)
	if err != nil {
		t.Fatalf("dry-run 모드 실행 실패: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Dry Run") {
		t.Errorf("dry-run 출력에 'Dry Run'이 포함되어야 한다, got %q", output)
	}
	if !strings.Contains(output, "PATH Preview") {
		t.Errorf("dry-run 출력에 'PATH Preview'가 포함되어야 한다, got %q", output)
	}
}

func TestRunPlatformCommand_DryRunDoesNotModifyFiles(t *testing.T) {
	// dry-run 모드에서는 settings.json이 수정되지 않아야 함
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsContent := `{"env": {"PATH": "/original/path"}}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dry-run", "--force"})
	_ = cmd.Execute()

	if err := runPlatformCommand(cmd, runtime.GOOS); err != nil {
		t.Fatalf("dry-run 실행 실패: %v", err)
	}

	// settings.json이 변경되지 않았는지 확인
	data, readErr := os.ReadFile(settingsPath)
	if readErr != nil {
		t.Fatalf("settings.json 읽기 실패: %v", readErr)
	}
	if string(data) != settingsContent {
		t.Error("dry-run 모드에서 settings.json이 변경되지 않아야 한다")
	}
}

// --- --json 플래그 테스트 ---

func TestRunPlatformCommand_JSONOutput(t *testing.T) {
	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--json", "--force", "--skip-backup"})
	_ = cmd.Execute()

	err := runPlatformCommand(cmd, runtime.GOOS)
	if err != nil {
		t.Fatalf("--json 모드 실행 실패: %v", err)
	}

	output := buf.String()

	// JSON 출력에서 PlatformProfile 구조체 파싱 시도
	// 출력에 다른 텍스트가 섞일 수 있으므로 JSON 시작점을 찾는다
	jsonStart := strings.Index(output, "{")
	if jsonStart < 0 {
		t.Fatalf("JSON 출력이 없다: %q", output)
	}

	// JSON 종료 위치 찾기 (마지막 })
	jsonEnd := strings.LastIndex(output, "}")
	if jsonEnd < 0 {
		t.Fatalf("JSON 종료 지점이 없다: %q", output)
	}

	jsonStr := output[jsonStart : jsonEnd+1]
	var profile platform.PlatformProfile
	if err := json.Unmarshal([]byte(jsonStr), &profile); err != nil {
		t.Fatalf("JSON 파싱 실패: %v\njson: %s", err, jsonStr)
	}

	if profile.Platform != runtime.GOOS {
		t.Errorf("profile.Platform = %q, want %q", profile.Platform, runtime.GOOS)
	}
	if len(profile.Checks) == 0 {
		t.Error("profile.Checks가 비어있지 않아야 한다")
	}
	if profile.ToolVersions == nil {
		t.Error("profile.ToolVersions가 nil이 아니어야 한다")
	}
}

// --- --skip-backup 플래그 테스트 ---

func TestRunPlatformCommand_SkipBackup(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"env":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--force", "--skip-backup"})
	_ = cmd.Execute()

	if err := runPlatformCommand(cmd, runtime.GOOS); err != nil {
		t.Fatalf("--skip-backup 실행 실패: %v", err)
	}

	// .bak 파일이 생성되지 않았는지 확인
	entries, readErr := os.ReadDir(claudeDir)
	if readErr != nil {
		t.Fatalf("디렉터리 읽기 실패: %v", readErr)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bak") {
			t.Errorf("--skip-backup 시 백업 파일이 생성되지 않아야 한다, found: %s", e.Name())
		}
	}
}

func TestRunPlatformCommand_BackupCreated(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"env":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	// --skip-backup 없이 실행 → 백업 생성됨
	cmd.SetArgs([]string{"--force"})
	_ = cmd.Execute()

	if err := runPlatformCommand(cmd, runtime.GOOS); err != nil {
		t.Fatalf("백업 포함 실행 실패: %v", err)
	}

	// .bak 파일이 생성되었는지 확인
	entries, readErr := os.ReadDir(claudeDir)
	if readErr != nil {
		t.Fatalf("디렉터리 읽기 실패: %v", readErr)
	}
	foundBackup := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bak") {
			foundBackup = true
			break
		}
	}
	if !foundBackup {
		t.Error("--skip-backup이 없을 때 .bak 백업 파일이 생성되어야 한다")
	}
}

// --- addPlatformFlags 테스트 ---

func TestAddPlatformFlags_AllFlagsRegistered(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	addPlatformFlags(cmd)

	expectedFlags := map[string]string{
		"force":       "플랫폼 불일치 경고 무시",
		"verbose":     "상세 진단 출력",
		"json":        "JSON 형식 출력",
		"auto":        "존재하지 않는 경로 자동 제외",
		"dry-run":     "실제 변경 없이 미리보기",
		"skip-backup": "settings.json 백업 건너뛰기",
	}

	for name, expectedUsage := range expectedFlags {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("플래그 %q가 등록되어야 한다", name)
			continue
		}
		if f.DefValue != "false" {
			t.Errorf("플래그 %q의 기본값이 false여야 한다, got %q", name, f.DefValue)
		}
		if !strings.Contains(f.Usage, expectedUsage) {
			t.Errorf("플래그 %q 사용법이 %q를 포함해야 한다, got %q", name, expectedUsage, f.Usage)
		}
	}
}

// --- 테이블 기반 parsePlatformFlags 조합 테스트 ---

func TestParsePlatformFlags_Combinations(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantForce  bool
		wantJSON   bool
		wantDryRun bool
		wantAuto   bool
	}{
		{
			name:       "force와 json 조합",
			args:       []string{"--force", "--json"},
			wantForce:  true,
			wantJSON:   true,
			wantDryRun: false,
			wantAuto:   false,
		},
		{
			name:       "dry-run과 auto 조합",
			args:       []string{"--dry-run", "--auto"},
			wantForce:  false,
			wantJSON:   false,
			wantDryRun: true,
			wantAuto:   true,
		},
		{
			name:       "플래그 없음",
			args:       []string{},
			wantForce:  false,
			wantJSON:   false,
			wantDryRun: false,
			wantAuto:   false,
		},
		{
			name:       "전체 플래그",
			args:       []string{"--force", "--json", "--dry-run", "--auto", "--verbose", "--skip-backup"},
			wantForce:  true,
			wantJSON:   true,
			wantDryRun: true,
			wantAuto:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestCmdWithFlags()
			cmd.SetArgs(tt.args)
			_ = cmd.Execute()

			flags := parsePlatformFlags(cmd)

			if flags.Force != tt.wantForce {
				t.Errorf("Force = %v, want %v", flags.Force, tt.wantForce)
			}
			if flags.JSON != tt.wantJSON {
				t.Errorf("JSON = %v, want %v", flags.JSON, tt.wantJSON)
			}
			if flags.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", flags.DryRun, tt.wantDryRun)
			}
			if flags.Auto != tt.wantAuto {
				t.Errorf("Auto = %v, want %v", flags.Auto, tt.wantAuto)
			}
		})
	}
}

// --- winCmd/macCmd의 RunE 함수가 올바른 타겟 플랫폼을 전달하는지 테스트 ---

func TestWinCmd_TargetsPlatformWindows(t *testing.T) {
	// winCmd.Long에서 Windows 관련 키워드가 포함되어 있는지 확인
	if !strings.Contains(winCmd.Long, "Windows") {
		t.Error("winCmd.Long에 'Windows'가 포함되어야 한다")
	}
	// winCmd.GroupID가 tools인지 확인
	if winCmd.GroupID != "tools" {
		t.Errorf("winCmd.GroupID = %q, want %q", winCmd.GroupID, "tools")
	}
}

func TestMacCmd_TargetsPlatformDarwin(t *testing.T) {
	// macCmd.Long에서 macOS 관련 키워드가 포함되어 있는지 확인
	if !strings.Contains(macCmd.Long, "macOS") {
		t.Error("macCmd.Long에 'macOS'가 포함되어야 한다")
	}
	// macCmd.GroupID가 tools인지 확인
	if macCmd.GroupID != "tools" {
		t.Errorf("macCmd.GroupID = %q, want %q", macCmd.GroupID, "tools")
	}
}

// --- winCmd/macCmd의 Long 설명 내용 검증 ---

func TestWinCmd_LongDescription(t *testing.T) {
	keywords := []string{"UTF-8", "MCP", "Git Bash", "WSL2", "LongPaths"}
	for _, kw := range keywords {
		if !strings.Contains(winCmd.Long, kw) {
			t.Errorf("winCmd.Long에 %q가 포함되어야 한다", kw)
		}
	}
}

func TestMacCmd_LongDescription(t *testing.T) {
	keywords := []string{"Homebrew", "심볼릭 링크", "셸 호환성"}
	for _, kw := range keywords {
		if !strings.Contains(macCmd.Long, kw) {
			t.Errorf("macCmd.Long에 %q가 포함되어야 한다", kw)
		}
	}
}

// --- runPlatformCommand --force를 통한 전체 실행 흐름 테스트 ---

func TestRunPlatformCommand_ForceExecution(t *testing.T) {
	// --force와 --skip-backup으로 플랫폼 불일치를 강제 무시하고 실행
	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--force", "--skip-backup"})
	_ = cmd.Execute()

	err := runPlatformCommand(cmd, runtime.GOOS)
	if err != nil {
		t.Fatalf("--force 실행 실패: %v", err)
	}

	output := buf.String()
	// 진단 결과 출력이 있어야 함 (FormatDiagnostics 결과)
	if output == "" {
		t.Error("--force 실행 시 출력이 비어있지 않아야 한다")
	}
}

// --- runPlatformCommand가 현재 플랫폼과 동일할 때 경고 없이 실행되는지 테스트 ---

func TestRunPlatformCommand_SamePlatformNoWarning(t *testing.T) {
	cmd := newTestCmdWithFlags()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--skip-backup"})
	_ = cmd.Execute()

	err := runPlatformCommand(cmd, runtime.GOOS)
	if err != nil {
		t.Fatalf("동일 플랫폼 실행 실패: %v", err)
	}

	output := buf.String()
	// 플랫폼 불일치 경고가 없어야 함
	if strings.Contains(output, "--force 플래그로 강제 실행") {
		t.Error("동일 플랫폼일 때 불일치 경고가 나타나지 않아야 한다")
	}
}

// --- updateSettingsPATH Windows 경로 변환 테스트 ---

func TestUpdateSettingsPATH_WindowsPathConversion(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows에서만 실행되는 테스트")
	}

	mock := newMockSystemInfo()
	settingsPath := "/mock/settings.json"
	mock.files[settingsPath] = []byte(`{"env":{}}`)

	// 슬래시 경로 입력
	newPATH := "C:/Users/test/bin;C:/Program Files/Go/bin"
	if err := updateSettingsPATH(mock, settingsPath, newPATH); err != nil {
		t.Fatalf("updateSettingsPATH 실패: %v", err)
	}

	data := mock.files[settingsPath]
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	env := result["env"].(map[string]interface{})
	pathVal := env["PATH"].(string)

	// Windows에서 슬래시가 백슬래시로 변환되어야 함
	if strings.Contains(pathVal, "/") {
		t.Errorf("Windows에서 PATH에 슬래시가 남아있으면 안 된다: %q", pathVal)
	}
	if !strings.Contains(pathVal, `\`) {
		t.Errorf("Windows에서 PATH에 백슬래시가 포함되어야 한다: %q", pathVal)
	}
}

func TestUpdateSettingsPATH_NonWindowsKeepsSlash(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("비Windows 플랫폼에서만 실행되는 테스트")
	}

	mock := newMockSystemInfo()
	settingsPath := "/mock/settings.json"
	mock.files[settingsPath] = []byte(`{"env":{}}`)

	newPATH := "/usr/local/bin:/usr/bin:/opt/homebrew/bin"
	if err := updateSettingsPATH(mock, settingsPath, newPATH); err != nil {
		t.Fatalf("updateSettingsPATH 실패: %v", err)
	}

	data := mock.files[settingsPath]
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	env := result["env"].(map[string]interface{})
	pathVal := env["PATH"].(string)

	if pathVal != newPATH {
		t.Errorf("비Windows에서 PATH가 변경되지 않아야 한다: got %q, want %q", pathVal, newPATH)
	}
}

// --- winCmd/macCmd 사용법(usage) 출력 테스트 ---

func TestWinCmd_UsageContainsFlags(t *testing.T) {
	usage := winCmd.UsageString()
	expectedInUsage := []string{"--force", "--verbose", "--json", "--auto", "--dry-run", "--skip-backup"}
	for _, flag := range expectedInUsage {
		if !strings.Contains(usage, flag) {
			t.Errorf("win usage에 %q가 포함되어야 한다", flag)
		}
	}
}

func TestMacCmd_UsageContainsFlags(t *testing.T) {
	usage := macCmd.UsageString()
	expectedInUsage := []string{"--force", "--verbose", "--json", "--auto", "--dry-run", "--skip-backup"}
	for _, flag := range expectedInUsage {
		if !strings.Contains(usage, flag) {
			t.Errorf("mac usage에 %q가 포함되어야 한다", flag)
		}
	}
}

// --- 빈 settings.json 처리 테스트 ---

func TestUpdateSettingsPATH_EmptyEnvObject(t *testing.T) {
	mock := newMockSystemInfo()
	settingsPath := "/mock/settings.json"
	mock.files[settingsPath] = []byte(`{"env": {}}`)

	if err := updateSettingsPATH(mock, settingsPath, "/new/bin"); err != nil {
		t.Fatalf("빈 env 객체 업데이트 실패: %v", err)
	}

	data := mock.files[settingsPath]
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	env, ok := result["env"].(map[string]interface{})
	if !ok {
		t.Fatal("env가 객체여야 한다")
	}
	if _, ok := env["PATH"]; !ok {
		t.Error("env.PATH가 설정되어야 한다")
	}
}
