package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultAutoConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultAutoConfig()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"ContextIsolated.Enabled", cfg.ContextIsolated.Enabled, true},
		{"SyncReviewIterations", cfg.ContextIsolated.SyncReviewIterations, DefaultSyncReviewIterations},
		{"Copilot.Enabled", cfg.ContextIsolated.Copilot.Enabled, true},
		{"Copilot.WaitMinutes", cfg.ContextIsolated.Copilot.WaitMinutes, DefaultCopilotWaitMinutes},
		{"Copilot.BotLogin", cfg.ContextIsolated.Copilot.BotLogin, DefaultCopilotBotLogin},
		{"Copilot.CheckIteration", cfg.ContextIsolated.Copilot.CheckIteration, DefaultCopilotCheckIteration},
		{"Teammate.Count", cfg.ContextIsolated.Teammate.Count, DefaultTeammateCount},
		{"Teammate.Mode", cfg.ContextIsolated.Teammate.Mode, DefaultTeammateMode},
		{"FinalMerge.Strategy", cfg.ContextIsolated.FinalMerge.Strategy, DefaultFinalMergeStrategy},
		{"FinalMerge.DeleteBranch", cfg.ContextIsolated.FinalMerge.DeleteBranch, true},
		{"FinalMerge.RequireCIPass", cfg.ContextIsolated.FinalMerge.RequireCIPass, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestAutoConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		modify  func(*AutoConfig)
		wantErr bool
	}{
		{
			name:    "valid default config",
			modify:  func(_ *AutoConfig) {},
			wantErr: false,
		},
		{
			name: "iterations zero returns error",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.SyncReviewIterations = 0
			},
			wantErr: true,
		},
		{
			name: "iterations negative returns error",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.SyncReviewIterations = -1
			},
			wantErr: true,
		},
		{
			name: "wait_minutes negative returns error",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.Copilot.WaitMinutes = -1
			},
			wantErr: true,
		},
		{
			name: "wait_minutes zero is valid",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.Copilot.WaitMinutes = 0
			},
			wantErr: false,
		},
		{
			name: "teammate count zero returns error",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.Teammate.Count = 0
			},
			wantErr: true,
		},
		{
			name: "teammate count negative returns error",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.Teammate.Count = -1
			},
			wantErr: true,
		},
		{
			name: "check_iteration negative returns error",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.Copilot.CheckIteration = -1
			},
			wantErr: true,
		},
		{
			name: "check_iteration zero is valid",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.Copilot.CheckIteration = 0
			},
			wantErr: false,
		},
		{
			name: "iterations exceeds max returns error",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.SyncReviewIterations = MaxSyncReviewIterations + 1
			},
			wantErr: true,
		},
		{
			name: "iterations at max is valid",
			modify: func(c *AutoConfig) {
				c.ContextIsolated.SyncReviewIterations = MaxSyncReviewIterations
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultAutoConfig()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestAutoConfigFromYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	autoYAML := []byte(`auto:
  context_isolated:
    enabled: false
    sync_review_iterations: 5
    copilot:
      enabled: false
      check_iteration: 2
      wait_minutes: 15
      bot_login: "custom-bot[bot]"
    teammate:
      count: 3
      mode: "manual"
      model: "opus"
    final_merge:
      strategy: "merge"
      delete_branch: false
      require_ci_pass: false
`)
	if err := os.WriteFile(filepath.Join(sectionsDir, "auto.yaml"), autoYAML, 0o644); err != nil {
		t.Fatalf("failed to write auto.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".ae"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify loaded values
	if cfg.Auto.ContextIsolated.Enabled != false {
		t.Error("Auto.ContextIsolated.Enabled: expected false")
	}
	if cfg.Auto.ContextIsolated.SyncReviewIterations != 5 {
		t.Errorf("Auto.ContextIsolated.SyncReviewIterations: got %d, want 5",
			cfg.Auto.ContextIsolated.SyncReviewIterations)
	}
	if cfg.Auto.ContextIsolated.Copilot.Enabled != false {
		t.Error("Auto.ContextIsolated.Copilot.Enabled: expected false")
	}
	if cfg.Auto.ContextIsolated.Copilot.CheckIteration != 2 {
		t.Errorf("Auto.ContextIsolated.Copilot.CheckIteration: got %d, want 2",
			cfg.Auto.ContextIsolated.Copilot.CheckIteration)
	}
	if cfg.Auto.ContextIsolated.Copilot.WaitMinutes != 15 {
		t.Errorf("Auto.ContextIsolated.Copilot.WaitMinutes: got %d, want 15",
			cfg.Auto.ContextIsolated.Copilot.WaitMinutes)
	}
	if cfg.Auto.ContextIsolated.Copilot.BotLogin != "custom-bot[bot]" {
		t.Errorf("Auto.ContextIsolated.Copilot.BotLogin: got %q, want %q",
			cfg.Auto.ContextIsolated.Copilot.BotLogin, "custom-bot[bot]")
	}
	if cfg.Auto.ContextIsolated.Teammate.Count != 3 {
		t.Errorf("Auto.ContextIsolated.Teammate.Count: got %d, want 3",
			cfg.Auto.ContextIsolated.Teammate.Count)
	}
	if cfg.Auto.ContextIsolated.Teammate.Mode != "manual" {
		t.Errorf("Auto.ContextIsolated.Teammate.Mode: got %q, want %q",
			cfg.Auto.ContextIsolated.Teammate.Mode, "manual")
	}
	if cfg.Auto.ContextIsolated.Teammate.Model != "opus" {
		t.Errorf("Auto.ContextIsolated.Teammate.Model: got %q, want %q",
			cfg.Auto.ContextIsolated.Teammate.Model, "opus")
	}
	if cfg.Auto.ContextIsolated.FinalMerge.Strategy != "merge" {
		t.Errorf("Auto.ContextIsolated.FinalMerge.Strategy: got %q, want %q",
			cfg.Auto.ContextIsolated.FinalMerge.Strategy, "merge")
	}
	if cfg.Auto.ContextIsolated.FinalMerge.DeleteBranch != false {
		t.Error("Auto.ContextIsolated.FinalMerge.DeleteBranch: expected false")
	}
	if cfg.Auto.ContextIsolated.FinalMerge.RequireCIPass != false {
		t.Error("Auto.ContextIsolated.FinalMerge.RequireCIPass: expected false")
	}

	// Verify section is marked as loaded
	sections := loader.LoadedSections()
	if !sections["auto"] {
		t.Error("expected auto section to be loaded")
	}
}

func TestAutoConfigDefaults(t *testing.T) {
	t.Parallel()

	// Load without auto.yaml - should use defaults
	tempDir := t.TempDir()
	root := setupTestdataDir(t, tempDir, []string{"user.yaml"})

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(root, ".ae"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Auto.ContextIsolated.SyncReviewIterations != DefaultSyncReviewIterations {
		t.Errorf("Auto.ContextIsolated.SyncReviewIterations: got %d, want default %d",
			cfg.Auto.ContextIsolated.SyncReviewIterations, DefaultSyncReviewIterations)
	}
	if !cfg.Auto.ContextIsolated.Copilot.Enabled {
		t.Error("Auto.ContextIsolated.Copilot.Enabled: expected default true")
	}

	sections := loader.LoadedSections()
	if sections["auto"] {
		t.Error("expected auto section to NOT be loaded when file is missing")
	}
}

func TestAutoSectionName(t *testing.T) {
	t.Parallel()

	if !IsValidSectionName("auto") {
		t.Error("IsValidSectionName(\"auto\") should return true")
	}
}

func TestNewDefaultConfigContainsAuto(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	if cfg.Auto.ContextIsolated.SyncReviewIterations != DefaultSyncReviewIterations {
		t.Errorf("Auto.ContextIsolated.SyncReviewIterations: got %d, want %d",
			cfg.Auto.ContextIsolated.SyncReviewIterations, DefaultSyncReviewIterations)
	}
}

func TestAutoConfigInvalidYAMLFallsBackToDefaults(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// Write auto.yaml with invalid values (iterations = 0)
	invalidYAML := []byte(`auto:
  context_isolated:
    sync_review_iterations: 0
`)
	if err := os.WriteFile(filepath.Join(sectionsDir, "auto.yaml"), invalidYAML, 0o644); err != nil {
		t.Fatalf("failed to write auto.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".ae"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Invalid config should fall back to defaults
	if cfg.Auto.ContextIsolated.SyncReviewIterations != DefaultSyncReviewIterations {
		t.Errorf("expected default SyncReviewIterations %d, got %d",
			DefaultSyncReviewIterations, cfg.Auto.ContextIsolated.SyncReviewIterations)
	}

	// Section should NOT be marked as loaded
	if loader.LoadedSections()["auto"] {
		t.Error("auto section should not be marked as loaded when validation fails")
	}
}

func TestAutoConfigPartialYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// Write partial auto.yaml - only override iterations
	partialYAML := []byte(`auto:
  context_isolated:
    sync_review_iterations: 5
`)
	if err := os.WriteFile(filepath.Join(sectionsDir, "auto.yaml"), partialYAML, 0o644); err != nil {
		t.Fatalf("failed to write auto.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".ae"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Overridden field should have new value
	if cfg.Auto.ContextIsolated.SyncReviewIterations != 5 {
		t.Errorf("SyncReviewIterations: got %d, want 5",
			cfg.Auto.ContextIsolated.SyncReviewIterations)
	}

	// Non-specified fields should have default values (from wrapper initialization)
	if cfg.Auto.ContextIsolated.Copilot.WaitMinutes != DefaultCopilotWaitMinutes {
		t.Errorf("Copilot.WaitMinutes: got %d, want default %d",
			cfg.Auto.ContextIsolated.Copilot.WaitMinutes, DefaultCopilotWaitMinutes)
	}
	if cfg.Auto.ContextIsolated.Teammate.Count != DefaultTeammateCount {
		t.Errorf("Teammate.Count: got %d, want default %d",
			cfg.Auto.ContextIsolated.Teammate.Count, DefaultTeammateCount)
	}
}

func TestAutoDefaultConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"DefaultSyncReviewIterations", DefaultSyncReviewIterations, 3},
		{"DefaultCopilotWaitMinutes", DefaultCopilotWaitMinutes, 10},
		{"DefaultCopilotCheckIteration", DefaultCopilotCheckIteration, 1},
		{"DefaultCopilotBotLogin", DefaultCopilotBotLogin, "copilot-pull-request-reviewer[bot]"},
		{"DefaultTeammateCount", DefaultTeammateCount, 1},
		{"DefaultTeammateMode", DefaultTeammateMode, "auto"},
		{"DefaultFinalMergeStrategy", DefaultFinalMergeStrategy, "squash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}
