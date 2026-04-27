package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPreferencesPath_Default(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetPreferencesPath("default")
	expected := filepath.Join(tmpDir, "preferences.yaml")
	if path != expected {
		t.Errorf("GetPreferencesPath(default) = %q, want %q", path, expected)
	}
}

func TestGetPreferencesPath_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetPreferencesPath("")
	expected := filepath.Join(tmpDir, "preferences.yaml")
	if path != expected {
		t.Errorf("GetPreferencesPath('') = %q, want %q", path, expected)
	}
}

func TestGetPreferencesPath_Named(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetPreferencesPath("work")
	expected := filepath.Join(tmpDir, "work", "preferences.yaml")
	if path != expected {
		t.Errorf("GetPreferencesPath(work) = %q, want %q", path, expected)
	}
}

func TestReadPreferences_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs, err := ReadPreferences("nonexistent")
	if err != nil {
		t.Fatalf("ReadPreferences(nonexistent) unexpected error: %v", err)
	}
	if prefs.UserName != "" || prefs.ConversationLang != "" {
		t.Errorf("ReadPreferences(nonexistent) = %+v, want zero-value", prefs)
	}
}

func TestWriteAndReadPreferences(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{
		UserName:         "testuser",
		ConversationLang: "ko",
		GitCommitLang:    "en",
		CodeCommentLang:  "en",
		DocLang:          "ko",
		ModelPolicy:      "high",
		Model:            "claude-opus-4-6",
		StatuslinePreset: "compact",
		TeammateDisplay:  "tmux",
	}

	if err := WritePreferences("myprofile", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	got, err := ReadPreferences("myprofile")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}

	if got.UserName != "testuser" {
		t.Errorf("UserName = %q, want %q", got.UserName, "testuser")
	}
	if got.ConversationLang != "ko" {
		t.Errorf("ConversationLang = %q, want %q", got.ConversationLang, "ko")
	}
	if got.GitCommitLang != "en" {
		t.Errorf("GitCommitLang = %q, want %q", got.GitCommitLang, "en")
	}
	if got.ModelPolicy != "high" {
		t.Errorf("ModelPolicy = %q, want %q", got.ModelPolicy, "high")
	}
	if got.Model != "claude-opus-4-6" {
		t.Errorf("Model = %q, want %q", got.Model, "claude-opus-4-6")
	}
	if got.StatuslinePreset != "compact" {
		t.Errorf("StatuslinePreset = %q, want %q", got.StatuslinePreset, "compact")
	}
	if got.TeammateDisplay != "tmux" {
		t.Errorf("TeammateDisplay = %q, want %q", got.TeammateDisplay, "tmux")
	}
}

func TestWritePreferences_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{UserName: "testuser"}

	if err := WritePreferences("newprofile", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	// Verify directory was created
	profileDir := filepath.Join(tmpDir, "newprofile")
	info, err := os.Stat(profileDir)
	if err != nil {
		t.Fatalf("profile directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestReadPreferences_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write invalid YAML
	prefsPath := filepath.Join(tmpDir, "preferences.yaml")
	if err := os.WriteFile(prefsPath, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPreferences("default")
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestWritePreferences_DefaultProfile(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{UserName: "defaultuser"}

	if err := WritePreferences("default", prefs); err != nil {
		t.Fatalf("WritePreferences(default): %v", err)
	}

	// Should write to base dir, not a "default" subdirectory
	expectedPath := filepath.Join(tmpDir, "preferences.yaml")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("preferences file not at expected path %q: %v", expectedPath, err)
	}
}

func TestPreferences_StatuslineSegments(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{
		StatuslinePreset: "custom",
		StatuslineSegments: map[string]bool{
			"model":   true,
			"context": true,
			"git":     false,
		},
	}

	if err := WritePreferences("default", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	got, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}

	if len(got.StatuslineSegments) != 3 {
		t.Errorf("StatuslineSegments length = %d, want 3", len(got.StatuslineSegments))
	}
	if !got.StatuslineSegments["model"] {
		t.Error("StatuslineSegments[model] should be true")
	}
	if got.StatuslineSegments["git"] {
		t.Error("StatuslineSegments[git] should be false")
	}
}

func TestReadPreferences_MigratesLegacyFile(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write a legacy dot-prefixed file
	legacyPath := filepath.Join(tmpDir, ".preferences.yaml")
	content := []byte("user_name: legacyuser\nconversation_lang: ko\n")
	if err := os.WriteFile(legacyPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	// ReadPreferences should migrate and return data
	prefs, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}
	if prefs.UserName != "legacyuser" {
		t.Errorf("UserName = %q, want %q", prefs.UserName, "legacyuser")
	}
	if prefs.ConversationLang != "ko" {
		t.Errorf("ConversationLang = %q, want %q", prefs.ConversationLang, "ko")
	}

	// Verify legacy file was renamed
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Error("legacy file should have been renamed")
	}
	newPath := filepath.Join(tmpDir, "preferences.yaml")
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("new file should exist: %v", err)
	}
}

func TestReadPreferences_MigrationSkippedWhenNewExists(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write both old and new files
	legacyPath := filepath.Join(tmpDir, ".preferences.yaml")
	newPath := filepath.Join(tmpDir, "preferences.yaml")
	if err := os.WriteFile(legacyPath, []byte("user_name: old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("user_name: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should read the new file, not migrate
	prefs, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}
	if prefs.UserName != "new" {
		t.Errorf("UserName = %q, want %q (should read new file)", prefs.UserName, "new")
	}

	// Legacy file should still exist (not removed when new file already exists)
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		t.Error("legacy file should still exist when new file already exists")
	}
}

func TestIsSetup_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write preferences
	if err := WritePreferences("default", ProfilePreferences{UserName: "test"}); err != nil {
		t.Fatal(err)
	}

	if !IsSetup("default") {
		t.Error("IsSetup(default) = false, want true")
	}
}

func TestIsSetup_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	if IsSetup("nonexistent") {
		t.Error("IsSetup(nonexistent) = true, want false")
	}
}

func TestIsSetup_LegacyFileOnly(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write only the legacy file
	legacyPath := filepath.Join(tmpDir, ".preferences.yaml")
	if err := os.WriteFile(legacyPath, []byte("user_name: legacy\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !IsSetup("default") {
		t.Error("IsSetup(default) = false, want true (legacy file exists)")
	}
}

func TestIsSetup_NamedProfile(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write preferences for named profile
	if err := WritePreferences("work", ProfilePreferences{UserName: "worker"}); err != nil {
		t.Fatal(err)
	}

	if !IsSetup("work") {
		t.Error("IsSetup(work) = false, want true")
	}
	if IsSetup("personal") {
		t.Error("IsSetup(personal) = true, want false")
	}
}

func TestPreferences_StatuslineTheme(t *testing.T) {
	tests := []struct {
		name  string
		theme string
	}{
		{"default theme", "default"},
		{"catppuccin-mocha", "catppuccin-mocha"},
		{"catppuccin-latte", "catppuccin-latte"},
		{"empty theme", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			orig := BaseDirOverride
			defer func() { BaseDirOverride = orig }()
			BaseDirOverride = tmpDir

			prefs := ProfilePreferences{
				StatuslineTheme: tt.theme,
			}

			if err := WritePreferences("default", prefs); err != nil {
				t.Fatalf("WritePreferences: %v", err)
			}

			got, err := ReadPreferences("default")
			if err != nil {
				t.Fatalf("ReadPreferences: %v", err)
			}

			if got.StatuslineTheme != tt.theme {
				t.Errorf("StatuslineTheme = %q, want %q", got.StatuslineTheme, tt.theme)
			}
		})
	}
}

func TestPreferences_StatuslineThemePersistsWithOtherFields(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{
		UserName:         "testuser",
		StatuslinePreset: "compact",
		StatuslineTheme:  "catppuccin-mocha",
	}

	if err := WritePreferences("default", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	got, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}

	if got.UserName != "testuser" {
		t.Errorf("UserName = %q, want %q", got.UserName, "testuser")
	}
	if got.StatuslinePreset != "compact" {
		t.Errorf("StatuslinePreset = %q, want %q", got.StatuslinePreset, "compact")
	}
	if got.StatuslineTheme != "catppuccin-mocha" {
		t.Errorf("StatuslineTheme = %q, want %q", got.StatuslineTheme, "catppuccin-mocha")
	}
}

// TestValidatePermissionMode verifies the whitelist guard for PermissionMode (REQ-22 보강).
// The whitelist must accept the empty string and the three legitimate values, and must
// reject anything else — most importantly the privilege-escalating "bypassPermissions"
// which is reserved for the --dangerously-skip-permissions CLI flag (the Bypass field).
func TestValidatePermissionMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		// Allowed
		{"empty (unset)", "", true},
		{"default", "default", true},
		{"auto", "auto", true},
		{"acceptEdits", "acceptEdits", true},
		// Rejected — case-sensitive whitelist
		{"bypassPermissions explicitly rejected", "bypassPermissions", false},
		{"random unknown value", "yolo", false},
		{"uppercase variant of allowed", "DEFAULT", false},
		{"camel-case mismatch", "AcceptEdits", false},
		{"whitespace not stripped", " auto ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ValidatePermissionMode(tt.in); got != tt.want {
				t.Errorf("ValidatePermissionMode(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestReadPreferences_NormalizesInvalidPermissionMode verifies that reading a
// preferences.yaml with a value outside the whitelist (e.g. user-edited
// "bypassPermissions") logs a warning and resets the field to the safe empty
// default — never propagating a privilege-escalating value to downstream
// consumers like SyncToProjectConfig.
func TestReadPreferences_NormalizesInvalidPermissionMode(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefsPath := filepath.Join(tmpDir, "preferences.yaml")
	content := []byte("user_name: testuser\npermission_mode: bypassPermissions\n")
	if err := os.WriteFile(prefsPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences should not error on invalid permission_mode: %v", err)
	}
	if got.PermissionMode != "" {
		t.Errorf("invalid permission_mode should be reset to empty, got %q", got.PermissionMode)
	}
	if got.UserName != "testuser" {
		t.Errorf("other fields should be preserved; UserName = %q, want %q", got.UserName, "testuser")
	}
}

// TestWritePreferences_RejectsInvalidPermissionMode verifies that the write
// path is also guarded — even if a future caller bypasses the wizard's huh.Select
// and constructs the struct directly with a bad value, the persistence layer
// must refuse to commit it to disk AND must NOT have written any file before
// the rejection (regression guard for "validate after write" mistakes).
func TestWritePreferences_RejectsInvalidPermissionMode(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefsPath := GetPreferencesPath("default")

	prefs := ProfilePreferences{
		UserName:       "x",
		PermissionMode: "bypassPermissions", // forbidden
	}
	err := WritePreferences("default", prefs)
	if err == nil {
		t.Fatal("WritePreferences must reject bypassPermissions, got nil error")
	}

	// File must NOT exist on disk after rejection. If a future change moves the
	// validation to AFTER WriteFile, this assertion catches it.
	if _, statErr := os.Stat(prefsPath); !os.IsNotExist(statErr) {
		t.Errorf("preferences file must not exist after rejection (statErr=%v)", statErr)
	}

	// Confirm that the well-known acceptable values do round-trip and produce
	// a file on disk.
	for _, ok := range []string{"", "default", "auto", "acceptEdits"} {
		prefs.PermissionMode = ok
		if err := WritePreferences("default", prefs); err != nil {
			t.Errorf("WritePreferences must accept %q, got error: %v", ok, err)
		}
	}
	if _, statErr := os.Stat(prefsPath); statErr != nil {
		t.Errorf("preferences file must exist after a successful write, got: %v", statErr)
	}
}
