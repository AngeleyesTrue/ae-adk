package cli

import (
	"bytes"
	"strings"
	"testing"
)

// --- TDD: ae profile 커맨드 사양 테스트 ---

// === profileCmd 등록 및 메타데이터 테스트 ===

func TestProfileCmd_Registered(t *testing.T) {
	// profileCmd가 rootCmd의 서브커맨드로 등록되어 있는지 확인
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "profile" {
			found = true
			break
		}
	}
	if !found {
		t.Error("profile should be registered on rootCmd")
	}
}

func TestProfileCmd_Use(t *testing.T) {
	if profileCmd.Use == "" {
		t.Error("profileCmd.Use should not be empty")
	}
	if profileCmd.Use != "profile" {
		t.Errorf("profileCmd.Use = %q, want %q", profileCmd.Use, "profile")
	}
}

func TestProfileCmd_Short(t *testing.T) {
	if profileCmd.Short == "" {
		t.Error("profileCmd.Short should not be empty")
	}
}

func TestProfileCmd_Long(t *testing.T) {
	if profileCmd.Long == "" {
		t.Error("profileCmd.Long should not be empty")
	}
}

func TestProfileCmd_GroupID(t *testing.T) {
	// profileCmd는 "tools" 그룹에 속해야 함
	if profileCmd.GroupID != "tools" {
		t.Errorf("profileCmd.GroupID = %q, want %q", profileCmd.GroupID, "tools")
	}
}

func TestProfileCmd_HasRunE(t *testing.T) {
	if profileCmd.RunE == nil {
		t.Error("profileCmd.RunE should not be nil")
	}
}

// === --setup / -s 플래그 테스트 ===

func TestProfileCmd_HasSetupFlag(t *testing.T) {
	flag := profileCmd.Flags().Lookup("setup")
	if flag == nil {
		t.Fatal("profileCmd should have --setup flag")
	}
	if flag.Shorthand != "s" {
		t.Errorf("--setup shorthand = %q, want %q", flag.Shorthand, "s")
	}
	if flag.DefValue != "false" {
		t.Errorf("--setup default = %q, want %q", flag.DefValue, "false")
	}
}

// === 서브커맨드 등록 테스트 ===

func TestProfileCmd_HasListSubcommand(t *testing.T) {
	found := false
	for _, cmd := range profileCmd.Commands() {
		if cmd.Name() == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("profileCmd should have 'list' subcommand")
	}
}

func TestProfileCmd_HasCurrentSubcommand(t *testing.T) {
	found := false
	for _, cmd := range profileCmd.Commands() {
		if cmd.Name() == "current" {
			found = true
			break
		}
	}
	if !found {
		t.Error("profileCmd should have 'current' subcommand")
	}
}

func TestProfileCmd_HasDeleteSubcommand(t *testing.T) {
	found := false
	for _, cmd := range profileCmd.Commands() {
		if cmd.Name() == "delete" {
			found = true
			break
		}
	}
	if !found {
		t.Error("profileCmd should have 'delete' subcommand")
	}
}

// === profileListCmd 메타데이터 및 별칭 테스트 ===

func TestProfileListCmd_Use(t *testing.T) {
	if profileListCmd.Use != "list" {
		t.Errorf("profileListCmd.Use = %q, want %q", profileListCmd.Use, "list")
	}
}

func TestProfileListCmd_Short(t *testing.T) {
	if profileListCmd.Short == "" {
		t.Error("profileListCmd.Short should not be empty")
	}
}

func TestProfileListCmd_HasLsAlias(t *testing.T) {
	// "ls" 별칭이 존재하는지 확인
	found := false
	for _, alias := range profileListCmd.Aliases {
		if alias == "ls" {
			found = true
			break
		}
	}
	if !found {
		t.Error("profileListCmd should have 'ls' alias")
	}
}

func TestProfileListCmd_HasRunE(t *testing.T) {
	if profileListCmd.RunE == nil {
		t.Error("profileListCmd.RunE should not be nil")
	}
}

// === profileCurrentCmd 메타데이터 테스트 ===

func TestProfileCurrentCmd_Use(t *testing.T) {
	if profileCurrentCmd.Use != "current" {
		t.Errorf("profileCurrentCmd.Use = %q, want %q", profileCurrentCmd.Use, "current")
	}
}

func TestProfileCurrentCmd_Short(t *testing.T) {
	if profileCurrentCmd.Short == "" {
		t.Error("profileCurrentCmd.Short should not be empty")
	}
}

func TestProfileCurrentCmd_HasRunE(t *testing.T) {
	if profileCurrentCmd.RunE == nil {
		t.Error("profileCurrentCmd.RunE should not be nil")
	}
}

// === profileDeleteCmd 메타데이터, 별칭, 인자 검증 테스트 ===

func TestProfileDeleteCmd_Use(t *testing.T) {
	if profileDeleteCmd.Use != "delete [name]" {
		t.Errorf("profileDeleteCmd.Use = %q, want %q", profileDeleteCmd.Use, "delete [name]")
	}
}

func TestProfileDeleteCmd_Short(t *testing.T) {
	if profileDeleteCmd.Short == "" {
		t.Error("profileDeleteCmd.Short should not be empty")
	}
}

func TestProfileDeleteCmd_HasRmAlias(t *testing.T) {
	// "rm" 별칭이 존재하는지 확인
	found := false
	for _, alias := range profileDeleteCmd.Aliases {
		if alias == "rm" {
			found = true
			break
		}
	}
	if !found {
		t.Error("profileDeleteCmd should have 'rm' alias")
	}
}

func TestProfileDeleteCmd_RequiresExactlyOneArg(t *testing.T) {
	if profileDeleteCmd.Args == nil {
		t.Fatal("profileDeleteCmd.Args should not be nil (expects ExactArgs(1))")
	}

	// 인자 없이 호출 시 에러 반환 확인
	err := profileDeleteCmd.Args(profileDeleteCmd, []string{})
	if err == nil {
		t.Error("profileDeleteCmd should reject zero arguments")
	}

	// 인자 1개는 정상
	err = profileDeleteCmd.Args(profileDeleteCmd, []string{"test"})
	if err != nil {
		t.Errorf("profileDeleteCmd should accept exactly 1 argument, got error: %v", err)
	}

	// 인자 2개 이상은 에러
	err = profileDeleteCmd.Args(profileDeleteCmd, []string{"a", "b"})
	if err == nil {
		t.Error("profileDeleteCmd should reject more than 1 argument")
	}
}

func TestProfileDeleteCmd_HasRunE(t *testing.T) {
	if profileDeleteCmd.RunE == nil {
		t.Error("profileDeleteCmd.RunE should not be nil")
	}
}

// === runProfileList 출력 테스트 ===

func TestRunProfileList_ProducesOutput(t *testing.T) {
	// runProfileList는 프로필 목록 또는 "no profiles found" 메시지를 출력해야 함
	buf := new(bytes.Buffer)
	profileListCmd.SetOut(buf)
	profileListCmd.SetErr(buf)

	err := runProfileList(profileListCmd, nil)
	if err != nil {
		t.Fatalf("runProfileList error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("runProfileList should produce output")
	}
}

func TestRunProfileList_ContainsDefaultOrNoProfiles(t *testing.T) {
	// 출력에는 프로필 이름이 포함되거나 "no profiles found" 메시지가 포함되어야 함
	buf := new(bytes.Buffer)
	profileListCmd.SetOut(buf)
	profileListCmd.SetErr(buf)

	err := runProfileList(profileListCmd, nil)
	if err != nil {
		t.Fatalf("runProfileList error: %v", err)
	}

	output := buf.String()
	hasDefault := strings.Contains(output, "default")
	hasNoProfiles := strings.Contains(output, "no profiles found")
	if !hasDefault && !hasNoProfiles {
		t.Errorf("runProfileList should show 'default' profile or 'no profiles found', got %q", output)
	}
}

// === runProfileCurrent 출력 테스트 ===

func TestRunProfileCurrent_ProducesOutput(t *testing.T) {
	// runProfileCurrent는 현재 프로필 이름을 출력해야 함
	buf := new(bytes.Buffer)
	profileCurrentCmd.SetOut(buf)
	profileCurrentCmd.SetErr(buf)

	err := runProfileCurrent(profileCurrentCmd, nil)
	if err != nil {
		t.Fatalf("runProfileCurrent error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Error("runProfileCurrent should produce output")
	}
}

func TestRunProfileCurrent_OutputNotEmpty(t *testing.T) {
	// CLAUDE_CONFIG_DIR 미설정 시 "default" 반환 확인
	buf := new(bytes.Buffer)
	profileCurrentCmd.SetOut(buf)
	profileCurrentCmd.SetErr(buf)

	err := runProfileCurrent(profileCurrentCmd, nil)
	if err != nil {
		t.Fatalf("runProfileCurrent error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if len(output) == 0 {
		t.Error("runProfileCurrent output should not be empty")
	}
}

// === runProfileCmd --setup 없이 호출 시 help 표시 테스트 ===

func TestRunProfileCmd_WithoutSetup_ReturnsNil(t *testing.T) {
	// --setup 없이 호출하면 cmd.Help()를 호출하고 nil을 반환해야 함
	buf := new(bytes.Buffer)
	profileCmd.SetOut(buf)
	profileCmd.SetErr(buf)

	// 플래그 초기화 후 --setup 없이 실행
	if err := profileCmd.Flags().Set("setup", "false"); err != nil {
		t.Fatalf("failed to set setup flag: %v", err)
	}

	err := runProfileCmd(profileCmd, []string{})
	if err != nil {
		t.Errorf("runProfileCmd without --setup should return nil, got error: %v", err)
	}
}

func TestRunProfileCmd_WithoutSetup_ShowsHelp(t *testing.T) {
	// --setup 없이 호출하면 help 출력이 발생해야 함
	buf := new(bytes.Buffer)
	profileCmd.SetOut(buf)
	profileCmd.SetErr(buf)

	if err := profileCmd.Flags().Set("setup", "false"); err != nil {
		t.Fatalf("failed to set setup flag: %v", err)
	}

	_ = runProfileCmd(profileCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "profile") {
		t.Errorf("runProfileCmd without --setup should show help containing 'profile', got %q", output)
	}
}

// === 서브커맨드 수 검증 ===

func TestProfileCmd_SubcommandCount(t *testing.T) {
	// profileCmd는 최소 3개의 서브커맨드(list, current, delete)를 가져야 함
	count := len(profileCmd.Commands())
	if count < 3 {
		t.Errorf("profileCmd should have at least 3 subcommands, got %d", count)
	}
}
