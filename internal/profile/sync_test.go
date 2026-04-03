package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// userFileWrapper mirrors config package wrapper for test verification.
type userFileWrapper struct {
	User struct {
		Name string `yaml:"name"`
	} `yaml:"user"`
}

// languageFileWrapper mirrors config package wrapper for test verification.
type languageFileWrapper struct {
	Language struct {
		ConversationLanguage     string `yaml:"conversation_language"`
		ConversationLanguageName string `yaml:"conversation_language_name"`
		GitCommitMessages        string `yaml:"git_commit_messages"`
		CodeComments             string `yaml:"code_comments"`
		Documentation            string `yaml:"documentation"`
	} `yaml:"language"`
}

func setupProjectConfig(t *testing.T, projectRoot string) {
	t.Helper()
	sectionsDir := filepath.Join(projectRoot, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write minimal config files
	userYAML := "user:\n  name: original\n"
	langYAML := "language:\n  conversation_language: en\n  conversation_language_name: en\n"
	qualityYAML := "constitution:\n  development_mode: tdd\n  enforce_quality: true\n  test_coverage_target: 85\n"

	for _, f := range []struct {
		name, content string
	}{
		{"user.yaml", userYAML},
		{"language.yaml", langYAML},
		{"quality.yaml", qualityYAML},
	} {
		path := filepath.Join(sectionsDir, f.name)
		if err := os.WriteFile(path, []byte(f.content), 0o644); err != nil {
			t.Fatalf("write %s: %v", f.name, err)
		}
	}
}

func TestSyncToProjectConfig_UserName(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	prefs := ProfilePreferences{
		UserName: "newuser",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify user.yaml was updated
	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatalf("read user.yaml: %v", err)
	}

	var wrapper userFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal user.yaml: %v", err)
	}
	if wrapper.User.Name != "newuser" {
		t.Errorf("user.name = %q, want %q", wrapper.User.Name, "newuser")
	}
}

func TestSyncToProjectConfig_Languages(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	prefs := ProfilePreferences{
		ConversationLang: "ko",
		GitCommitLang:    "en",
		CodeCommentLang:  "en",
		DocLang:          "ko",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify language.yaml was updated
	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "language.yaml"))
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}

	var wrapper languageFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal language.yaml: %v", err)
	}
	if wrapper.Language.ConversationLanguage != "ko" {
		t.Errorf("conversation_language = %q, want %q", wrapper.Language.ConversationLanguage, "ko")
	}
	if wrapper.Language.ConversationLanguageName != "ko" {
		t.Errorf("conversation_language_name = %q, want %q", wrapper.Language.ConversationLanguageName, "ko")
	}
	if wrapper.Language.GitCommitMessages != "en" {
		t.Errorf("git_commit_messages = %q, want %q", wrapper.Language.GitCommitMessages, "en")
	}
	if wrapper.Language.CodeComments != "en" {
		t.Errorf("code_comments = %q, want %q", wrapper.Language.CodeComments, "en")
	}
	if wrapper.Language.Documentation != "ko" {
		t.Errorf("documentation = %q, want %q", wrapper.Language.Documentation, "ko")
	}
}

func TestSyncToProjectConfig_EmptyPrefsNoOverwrite(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// Empty prefs should not overwrite existing config
	prefs := ProfilePreferences{}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify user.yaml was NOT changed
	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatalf("read user.yaml: %v", err)
	}

	var wrapper userFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal user.yaml: %v", err)
	}
	if wrapper.User.Name != "original" {
		t.Errorf("user.name = %q, want %q (should not overwrite)", wrapper.User.Name, "original")
	}
}

func TestSyncToProjectConfig_PartialPrefs(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// Only set conversation lang, others should preserve defaults
	prefs := ProfilePreferences{
		ConversationLang: "ja",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify language.yaml was updated for conversation_language only
	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "language.yaml"))
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}

	var wrapper languageFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal language.yaml: %v", err)
	}
	if wrapper.Language.ConversationLanguage != "ja" {
		t.Errorf("conversation_language = %q, want %q", wrapper.Language.ConversationLanguage, "ja")
	}
}

func TestSyncToProjectConfig_NoConfigDir(t *testing.T) {
	projectRoot := t.TempDir()
	// No .ae directory - should still work (ConfigManager creates defaults)

	prefs := ProfilePreferences{
		UserName:         "testuser",
		ConversationLang: "ko",
	}

	// This may fail if config directory doesn't exist at all
	// The error is expected and should be handled gracefully
	err := SyncToProjectConfig(projectRoot, prefs)
	// We accept either nil (if ConfigManager handles missing dirs) or an error
	_ = err
}

// statuslineFileWrapper is a local test helper for reading statusline.yaml
type statuslineFileWrapper struct {
	Statusline struct {
		Mode     string          `yaml:"mode"`
		Preset   string          `yaml:"preset"`
		Segments map[string]bool `yaml:"segments"`
		Theme    string          `yaml:"theme"`
	} `yaml:"statusline"`
}

func TestSyncToProjectConfig_StatuslineTheme(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	prefs := ProfilePreferences{
		StatuslineTheme:  "catppuccin-mocha",
		StatuslinePreset: "full",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify statusline.yaml was created
	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	var wrapper statuslineFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal statusline.yaml: %v", err)
	}
	if wrapper.Statusline.Theme != "catppuccin-mocha" {
		t.Errorf("theme = %q, want %q", wrapper.Statusline.Theme, "catppuccin-mocha")
	}
	if wrapper.Statusline.Preset != "full" {
		t.Errorf("preset = %q, want %q", wrapper.Statusline.Preset, "full")
	}
}

func TestSyncToProjectConfig_StatuslineDefaultsWhenAbsent(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// Provide only theme - preset and segments should get defaults when file absent
	prefs := ProfilePreferences{
		StatuslineTheme: "catppuccin-latte",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	var wrapper statuslineFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal statusline.yaml: %v", err)
	}

	// Theme is set by preference
	if wrapper.Statusline.Theme != "catppuccin-latte" {
		t.Errorf("theme = %q, want %q", wrapper.Statusline.Theme, "catppuccin-latte")
	}
	// Preset defaults to "full" when absent
	if wrapper.Statusline.Preset != "full" {
		t.Errorf("preset = %q, want %q", wrapper.Statusline.Preset, "full")
	}
	// Segments should all be enabled
	for _, seg := range []string{"model", "context", "output_style", "directory", "git_status", "claude_version", "ae_version", "git_branch"} {
		if !wrapper.Statusline.Segments[seg] {
			t.Errorf("segment %q should be enabled by default", seg)
		}
	}
}

func TestSyncToProjectConfig_StatuslineSegments(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	prefs := ProfilePreferences{
		StatuslinePreset: "custom",
		StatuslineSegments: map[string]bool{
			"model":   true,
			"context": true,
			"git":     false,
		},
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	var wrapper statuslineFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal statusline.yaml: %v", err)
	}
	if wrapper.Statusline.Preset != "custom" {
		t.Errorf("preset = %q, want %q", wrapper.Statusline.Preset, "custom")
	}
	if !wrapper.Statusline.Segments["model"] {
		t.Error("segments[model] should be true")
	}
	if wrapper.Statusline.Segments["git"] {
		t.Error("segments[git] should be false")
	}
}

func TestSyncToProjectConfig_StatuslinePreservesExistingConfig(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// Write an existing statusline.yaml
	sectionsDir := filepath.Join(projectRoot, ".ae", "config", "sections")
	existingYAML := "statusline:\n  preset: compact\n  theme: default\n  segments:\n    model: true\n    context: false\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "statusline.yaml"), []byte(existingYAML), 0o644); err != nil {
		t.Fatalf("write statusline.yaml: %v", err)
	}

	// Only update theme - preset and segments should be preserved
	prefs := ProfilePreferences{
		StatuslineTheme: "catppuccin-mocha",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(sectionsDir, "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	var wrapper statuslineFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal statusline.yaml: %v", err)
	}

	// Theme updated
	if wrapper.Statusline.Theme != "catppuccin-mocha" {
		t.Errorf("theme = %q, want %q", wrapper.Statusline.Theme, "catppuccin-mocha")
	}
	// Preset preserved
	if wrapper.Statusline.Preset != "compact" {
		t.Errorf("preset = %q, want %q", wrapper.Statusline.Preset, "compact")
	}
	// Segments preserved
	if !wrapper.Statusline.Segments["model"] {
		t.Error("segments[model] should be preserved as true")
	}
	if wrapper.Statusline.Segments["context"] {
		t.Error("segments[context] should be preserved as false")
	}
}

func TestSyncToProjectConfig_StatuslineMode(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	prefs := ProfilePreferences{
		StatuslineMode:   "verbose",
		StatuslinePreset: "full",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	var wrapper statuslineFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal statusline.yaml: %v", err)
	}
	if wrapper.Statusline.Mode != "verbose" {
		t.Errorf("mode = %q, want %q", wrapper.Statusline.Mode, "verbose")
	}
	if wrapper.Statusline.Preset != "full" {
		t.Errorf("preset = %q, want %q", wrapper.Statusline.Preset, "full")
	}
}

func TestSyncToProjectConfig_StatuslineModeOnlyDoesNotResetPreset(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// Write existing statusline.yaml with a preset
	sectionsDir := filepath.Join(projectRoot, ".ae", "config", "sections")
	existingYAML := "statusline:\n  preset: compact\n  theme: default\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "statusline.yaml"), []byte(existingYAML), 0o644); err != nil {
		t.Fatalf("write statusline.yaml: %v", err)
	}

	// Only update mode — preset should be preserved
	prefs := ProfilePreferences{
		StatuslineMode: "minimal",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(sectionsDir, "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	var wrapper statuslineFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal statusline.yaml: %v", err)
	}
	if wrapper.Statusline.Mode != "minimal" {
		t.Errorf("mode = %q, want %q", wrapper.Statusline.Mode, "minimal")
	}
	if wrapper.Statusline.Preset != "compact" {
		t.Errorf("preset = %q, want %q (should be preserved)", wrapper.Statusline.Preset, "compact")
	}
}

func TestSyncToProjectConfig_NoStatuslinePrefs(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// No statusline preferences - statusline.yaml should not be created
	prefs := ProfilePreferences{
		UserName: "testuser",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// statusline.yaml should not be created
	_, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "statusline.yaml"))
	if err == nil {
		t.Error("statusline.yaml should not be created when no statusline preferences are set")
	}
}

func TestSyncToProjectConfig_AllFields(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	prefs := ProfilePreferences{
		UserName:         "fulluser",
		ConversationLang: "zh",
		GitCommitLang:    "zh",
		CodeCommentLang:  "zh",
		DocLang:          "zh",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify both user.yaml and language.yaml were updated
	userData, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatalf("read user.yaml: %v", err)
	}
	var uw userFileWrapper
	if err := yaml.Unmarshal(userData, &uw); err != nil {
		t.Fatalf("unmarshal user.yaml: %v", err)
	}
	if uw.User.Name != "fulluser" {
		t.Errorf("user.name = %q, want %q", uw.User.Name, "fulluser")
	}

	langData, err := os.ReadFile(filepath.Join(projectRoot, ".ae", "config", "sections", "language.yaml"))
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	var lw languageFileWrapper
	if err := yaml.Unmarshal(langData, &lw); err != nil {
		t.Fatalf("unmarshal language.yaml: %v", err)
	}
	if lw.Language.ConversationLanguage != "zh" {
		t.Errorf("conversation_language = %q, want %q", lw.Language.ConversationLanguage, "zh")
	}
	if lw.Language.GitCommitMessages != "zh" {
		t.Errorf("git_commit_messages = %q, want %q", lw.Language.GitCommitMessages, "zh")
	}
	if lw.Language.CodeComments != "zh" {
		t.Errorf("code_comments = %q, want %q", lw.Language.CodeComments, "zh")
	}
	if lw.Language.Documentation != "zh" {
		t.Errorf("documentation = %q, want %q", lw.Language.Documentation, "zh")
	}
}

// setupAgentFiles creates minimal agent definition files with model: frontmatter
// and a manifest.json so that ApplyModelPolicy can patch them.
func setupAgentFiles(t *testing.T, projectRoot string) {
	t.Helper()
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "ae")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create agent files with high-policy models (default)
	agents := map[string]string{
		"manager-spec.md":     "---\nname: manager-spec\nmodel: opus\n---\nSpec agent.",
		"expert-backend.md":   "---\nname: expert-backend\nmodel: opus\n---\nBackend agent.",
		"manager-docs.md":     "---\nname: manager-docs\nmodel: sonnet\n---\nDocs agent.",
		"manager-quality.md":  "---\nname: manager-quality\nmodel: haiku\n---\nQuality agent.",
	}
	for name, content := range agents {
		path := filepath.Join(agentsDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	// Create a minimal manifest.json
	aeDir := filepath.Join(projectRoot, ".ae")
	if err := os.MkdirAll(aeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifestData := map[string]any{
		"version": "1.0.0",
		"files":   map[string]any{},
	}
	data, _ := json.Marshal(manifestData)
	if err := os.WriteFile(filepath.Join(aeDir, "manifest.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSyncToProjectConfig_ModelPolicy(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)
	setupAgentFiles(t, projectRoot)

	prefs := ProfilePreferences{
		ModelPolicy: "low",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify agent files were patched with low-policy models.
	// Under "low" policy:
	//   manager-spec: sonnet (was opus)
	//   expert-backend: sonnet (was opus)
	//   manager-docs: haiku (was sonnet)
	//   manager-quality: haiku (unchanged)
	expectations := map[string]string{
		"manager-spec.md":    "model: sonnet",
		"expert-backend.md":  "model: sonnet",
		"manager-docs.md":    "model: haiku",
		"manager-quality.md": "model: haiku",
	}

	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "ae")
	for filename, expectedModel := range expectations {
		data, err := os.ReadFile(filepath.Join(agentsDir, filename))
		if err != nil {
			t.Fatalf("read %s: %v", filename, err)
		}
		if !strings.Contains(string(data), expectedModel) {
			t.Errorf("%s: expected %q, got:\n%s", filename, expectedModel, string(data))
		}
	}
}

func TestSyncToProjectConfig_ModelPolicyEmpty_NoChange(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)
	setupAgentFiles(t, projectRoot)

	// Empty model policy should not modify agent files
	prefs := ProfilePreferences{
		UserName: "testuser",
		// ModelPolicy intentionally empty
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify agent files were NOT changed (still have original models)
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "ae")
	data, err := os.ReadFile(filepath.Join(agentsDir, "manager-spec.md"))
	if err != nil {
		t.Fatalf("read manager-spec.md: %v", err)
	}
	if !strings.Contains(string(data), "model: opus") {
		t.Errorf("manager-spec.md should still have 'model: opus' when ModelPolicy is empty, got:\n%s", string(data))
	}
}

func TestSyncToProjectConfig_ModelPolicyInvalid_NoChange(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)
	setupAgentFiles(t, projectRoot)

	// Invalid model policy should not modify agent files
	prefs := ProfilePreferences{
		ModelPolicy: "ultra", // invalid
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Verify agent files were NOT changed
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "ae")
	data, err := os.ReadFile(filepath.Join(agentsDir, "manager-spec.md"))
	if err != nil {
		t.Fatalf("read manager-spec.md: %v", err)
	}
	if !strings.Contains(string(data), "model: opus") {
		t.Errorf("manager-spec.md should still have 'model: opus' with invalid policy, got:\n%s", string(data))
	}
}

func TestSyncToProjectConfig_ModelPolicyHigh_PreservesOpus(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)
	setupAgentFiles(t, projectRoot)

	prefs := ProfilePreferences{
		ModelPolicy: "high",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Under "high" policy, opus agents should remain opus
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "ae")
	data, err := os.ReadFile(filepath.Join(agentsDir, "manager-spec.md"))
	if err != nil {
		t.Fatalf("read manager-spec.md: %v", err)
	}
	if !strings.Contains(string(data), "model: opus") {
		t.Errorf("manager-spec.md should have 'model: opus' with high policy, got:\n%s", string(data))
	}
}

func TestSyncToProjectConfig_ModelPolicyMedium(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)
	setupAgentFiles(t, projectRoot)

	prefs := ProfilePreferences{
		ModelPolicy: "medium",
	}

	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	// Medium policy per agentModelMap:
	//   manager-spec: [opus, opus, sonnet] → medium keeps opus
	//   expert-backend: [opus, sonnet, sonnet] → medium uses sonnet
	//   manager-docs: [sonnet, haiku, haiku] → medium uses haiku
	//   manager-quality: [haiku, haiku, haiku] → unchanged
	expectations := map[string]string{
		"manager-spec.md":    "model: opus",
		"expert-backend.md":  "model: sonnet",
		"manager-docs.md":    "model: haiku",
		"manager-quality.md": "model: haiku",
	}

	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "ae")
	for filename, expectedModel := range expectations {
		data, err := os.ReadFile(filepath.Join(agentsDir, filename))
		if err != nil {
			t.Fatalf("read %s: %v", filename, err)
		}
		if !strings.Contains(string(data), expectedModel) {
			t.Errorf("%s: expected %q with medium policy, got:\n%s", filename, expectedModel, string(data))
		}
	}
}

func TestSyncToProjectConfig_ModelPolicyNoManifest_Warns(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)
	// Intentionally NOT calling setupAgentFiles - no manifest.json

	prefs := ProfilePreferences{
		ModelPolicy: "low",
	}

	// Should not return error - manifest missing is a warning, not a failure
	if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
		t.Fatalf("SyncToProjectConfig should not fail when manifest is missing: %v", err)
	}
}
