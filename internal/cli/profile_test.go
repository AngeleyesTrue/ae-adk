package cli

import (
	"bytes"
	"strings"
	"testing"
)

// --- TDD: ae profile 커맨드 사양 테스트 ---

// === profileCmd 메타데이터 테스트 (테이블 주도) ===

func TestProfileCmd_Metadata(t *testing.T) {
	tests := []struct {
		name  string
		check func(*testing.T)
	}{
		{"Registered", func(t *testing.T) {
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
		}},
		{"Use", func(t *testing.T) {
			if profileCmd.Use == "" {
				t.Error("profileCmd.Use should not be empty")
			}
			if profileCmd.Use != "profile" {
				t.Errorf("profileCmd.Use = %q, want %q", profileCmd.Use, "profile")
			}
		}},
		{"Short", func(t *testing.T) {
			if profileCmd.Short == "" {
				t.Error("profileCmd.Short should not be empty")
			}
		}},
		{"Long", func(t *testing.T) {
			if profileCmd.Long == "" {
				t.Error("profileCmd.Long should not be empty")
			}
		}},
		{"GroupID", func(t *testing.T) {
			if profileCmd.GroupID != "tools" {
				t.Errorf("profileCmd.GroupID = %q, want %q", profileCmd.GroupID, "tools")
			}
		}},
		{"HasRunE", func(t *testing.T) {
			if profileCmd.RunE == nil {
				t.Error("profileCmd.RunE should not be nil")
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.check)
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

// === 서브커맨드 등록 테스트 (테이블 주도) ===

func TestProfileCmd_Subcommands(t *testing.T) {
	expected := []string{"list", "current", "delete"}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, cmd := range profileCmd.Commands() {
				if cmd.Name() == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("profileCmd should have %q subcommand", name)
			}
		})
	}
	t.Run("Count", func(t *testing.T) {
		count := len(profileCmd.Commands())
		if count < 3 {
			t.Errorf("profileCmd should have at least 3 subcommands, got %d", count)
		}
	})
}

// === 서브커맨드 메타데이터 테스트 (테이블 주도) ===

func TestProfileSubcmds_Metadata(t *testing.T) {
	// list
	t.Run("list/Use", func(t *testing.T) {
		if profileListCmd.Use != "list" {
			t.Errorf("profileListCmd.Use = %q, want %q", profileListCmd.Use, "list")
		}
	})
	t.Run("list/Short", func(t *testing.T) {
		if profileListCmd.Short == "" {
			t.Error("profileListCmd.Short should not be empty")
		}
	})
	t.Run("list/HasRunE", func(t *testing.T) {
		if profileListCmd.RunE == nil {
			t.Error("profileListCmd.RunE should not be nil")
		}
	})
	t.Run("list/Alias_ls", func(t *testing.T) {
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
	})

	// current
	t.Run("current/Use", func(t *testing.T) {
		if profileCurrentCmd.Use != "current" {
			t.Errorf("profileCurrentCmd.Use = %q, want %q", profileCurrentCmd.Use, "current")
		}
	})
	t.Run("current/Short", func(t *testing.T) {
		if profileCurrentCmd.Short == "" {
			t.Error("profileCurrentCmd.Short should not be empty")
		}
	})
	t.Run("current/HasRunE", func(t *testing.T) {
		if profileCurrentCmd.RunE == nil {
			t.Error("profileCurrentCmd.RunE should not be nil")
		}
	})

	// delete
	t.Run("delete/Use", func(t *testing.T) {
		if profileDeleteCmd.Use != "delete [name]" {
			t.Errorf("profileDeleteCmd.Use = %q, want %q", profileDeleteCmd.Use, "delete [name]")
		}
	})
	t.Run("delete/Short", func(t *testing.T) {
		if profileDeleteCmd.Short == "" {
			t.Error("profileDeleteCmd.Short should not be empty")
		}
	})
	t.Run("delete/HasRunE", func(t *testing.T) {
		if profileDeleteCmd.RunE == nil {
			t.Error("profileDeleteCmd.RunE should not be nil")
		}
	})
	t.Run("delete/Alias_rm", func(t *testing.T) {
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
	})
}

// === profileDeleteCmd 인자 검증 테스트 ===

func TestProfileDeleteCmd_RequiresExactlyOneArg(t *testing.T) {
	if profileDeleteCmd.Args == nil {
		t.Fatal("profileDeleteCmd.Args should not be nil (expects ExactArgs(1))")
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"zero_args", []string{}, true},
		{"one_arg", []string{"test"}, false},
		{"two_args", []string{"a", "b"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := profileDeleteCmd.Args(profileDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

// === runProfileList 출력 테스트 ===

func TestRunProfileList_ProducesOutput(t *testing.T) {
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
	buf := new(bytes.Buffer)
	profileCmd.SetOut(buf)
	profileCmd.SetErr(buf)

	if err := profileCmd.Flags().Set("setup", "false"); err != nil {
		t.Fatalf("failed to set setup flag: %v", err)
	}

	err := runProfileCmd(profileCmd, []string{})
	if err != nil {
		t.Errorf("runProfileCmd without --setup should return nil, got error: %v", err)
	}
}

func TestRunProfileCmd_WithoutSetup_ShowsHelp(t *testing.T) {
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
