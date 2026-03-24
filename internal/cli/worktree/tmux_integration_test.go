package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetActiveMode_Parse tests parsing llm.yaml for team_mode.
func TestGetActiveMode_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantMode string
	}{
		{
			name: "empty team_mode returns cc",
			content: `
llm:
    team_mode: ""
`,
			wantMode: "cc",
		},
		{
			name: "cc mode explicit",
			content: `
llm:
    team_mode: "cc"
`,
			wantMode: "cc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			// Create the full path structure: .ae/config/sections/llm.yaml
			configDir := filepath.Join(tempDir, ".ae", "config", "sections")
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				t.Fatal(err)
			}
			llmPath := filepath.Join(configDir, "llm.yaml")
			if err := os.WriteFile(llmPath, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			got, err := GetActiveMode(tempDir)
			if err != nil {
				t.Fatalf("GetActiveMode() error = %v", err)
			}
			if got != tt.wantMode {
				t.Errorf("GetActiveMode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

// TestGetActiveMode_Default tests that missing file returns cc.
func TestGetActiveMode_Default(t *testing.T) {
	tempDir := t.TempDir()
	// No llm.yaml file

	got, err := GetActiveMode(tempDir)
	if err != nil {
		t.Fatalf("GetActiveMode() error = %v", err)
	}
	if got != "cc" {
		t.Errorf("GetActiveMode() = %v, want cc", got)
	}
}

// TestBuildTmuxSessionConfig_CCMode tests that CC mode builds config correctly.
func TestBuildTmuxSessionConfig_CCMode(t *testing.T) {
	tempDir := t.TempDir()

	// Create llm.yaml with empty team_mode (cc)
	llmContent := `
llm:
    team_mode: ""
`
	configDir := filepath.Join(tempDir, ".ae", "config", "sections")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	llmPath := filepath.Join(configDir, "llm.yaml")
	if err := os.WriteFile(llmPath, []byte(llmContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := BuildTmuxSessionConfig("test-project", "SPEC-TEST-001", "/worktree", tempDir)
	if err != nil {
		t.Fatalf("BuildTmuxSessionConfig() error = %v", err)
	}

	if cfg.ActiveMode != "cc" {
		t.Errorf("ActiveMode = %v, want cc", cfg.ActiveMode)
	}

	if cfg.ProjectName != "test-project" {
		t.Errorf("ProjectName = %v, want test-project", cfg.ProjectName)
	}

	if cfg.SpecID != "SPEC-TEST-001" {
		t.Errorf("SpecID = %v, want SPEC-TEST-001", cfg.SpecID)
	}

	if cfg.WorktreePath != "/worktree" {
		t.Errorf("WorktreePath = %v, want /worktree", cfg.WorktreePath)
	}
}
