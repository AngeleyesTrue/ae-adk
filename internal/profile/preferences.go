package profile

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProfilePreferences holds per-profile user preferences.
// Stored at ~/.ae/claude-profiles/<name>/preferences.yaml.
// These settings apply across all projects when using this profile.
type ProfilePreferences struct {
	// Identity
	UserName string `yaml:"user_name,omitempty"`

	// Language settings
	ConversationLang string `yaml:"conversation_lang,omitempty"` // "en", "ko", "ja", "zh"
	GitCommitLang    string `yaml:"git_commit_lang,omitempty"`
	CodeCommentLang  string `yaml:"code_comment_lang,omitempty"`
	DocLang          string `yaml:"doc_lang,omitempty"`

	// LLM settings
	ModelPolicy string `yaml:"model_policy,omitempty"` // "high", "medium", "low"
	Model       string `yaml:"model,omitempty"`        // e.g. "claude-opus-4-6"

	// Launch settings
	Bypass      bool   `yaml:"bypass,omitempty"`       // --dangerously-skip-permissions
	EffortLevel string `yaml:"effort_level,omitempty"` // "low", "medium", "high", "xhigh", "max"

	// Display settings
	StatuslineMode     string          `yaml:"statusline_mode,omitempty"`     // "minimal", "default", "verbose"
	StatuslinePreset   string          `yaml:"statusline_preset,omitempty"`   // "full", "compact", "minimal", "custom"
	StatuslineSegments map[string]bool `yaml:"statusline_segments,omitempty"` // segment toggles for custom preset
	StatuslineTheme    string          `yaml:"statusline_theme,omitempty"`    // "default", "catppuccin-mocha", "catppuccin-latte"
	TeammateDisplay    string          `yaml:"teammate_display,omitempty"`    // "auto", "in-process", "tmux"

	// Permission mode (REQ-22)
	// 위저드에서 선택한 권한 모드. TeammateDisplay와는 별개의 의미를 가지므로
	// 전용 필드로 분리한다. 가능한 값: "default", "auto", "acceptEdits".
	// 보안 정책: "bypassPermissions"는 settings.json의 disableBypassPermissionsMode
	// 정책과 충돌하므로 절대 허용하지 않는다. ValidatePermissionMode 참조.
	PermissionMode string `yaml:"permission_mode,omitempty"`
}

// ValidatePermissionMode reports whether the given value is an acceptable
// preferences-level permission_mode. Empty string ("" = unset) is also valid.
//
// Rejected values include "bypassPermissions" — that mode is reserved for
// Claude Code CLI's `--dangerously-skip-permissions` flag (modeled by the
// Bypass field above) and must not be settable via the preferences YAML to
// keep the disableBypassPermissionsMode policy meaningful.
func ValidatePermissionMode(mode string) bool {
	switch mode {
	case "", "default", "auto", "acceptEdits":
		return true
	}
	return false
}

const (
	preferencesFile       = "preferences.yaml"
	legacyPreferencesFile = ".preferences.yaml"
)

// GetPreferencesPath returns the path to preferences.yaml for a profile.
func GetPreferencesPath(profileName string) string {
	baseDir := GetBaseDir()
	if profileName == "" || profileName == "default" {
		return filepath.Join(baseDir, preferencesFile)
	}
	return filepath.Join(baseDir, profileName, preferencesFile)
}

// IsSetup checks if a profile has been configured by looking for
// preferences.yaml (or legacy .preferences.yaml).
func IsSetup(profileName string) bool {
	path := GetPreferencesPath(profileName)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	// Check legacy dot-prefixed file
	dir := filepath.Dir(path)
	legacyPath := filepath.Join(dir, legacyPreferencesFile)
	if _, err := os.Stat(legacyPath); err == nil {
		return true
	}
	return false
}

// ReadPreferences reads the preferences for a profile.
// Returns a zero-value ProfilePreferences if the file does not exist.
// Automatically migrates legacy .preferences.yaml to preferences.yaml.
func ReadPreferences(profileName string) (ProfilePreferences, error) {
	path := GetPreferencesPath(profileName)
	migrateOldFile(filepath.Dir(path), legacyPreferencesFile, preferencesFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ProfilePreferences{}, nil
		}
		return ProfilePreferences{}, fmt.Errorf("read preferences: %w", err)
	}
	var prefs ProfilePreferences
	if err := yaml.Unmarshal(data, &prefs); err != nil {
		return ProfilePreferences{}, fmt.Errorf("parse preferences: %w", err)
	}
	// 화이트리스트에 없는 PermissionMode 값은 안전한 기본(빈 문자열)으로 정규화한다.
	// 직접 YAML을 편집하여 "bypassPermissions" 같은 위험 값을 주입하는 경로를 차단한다.
	if !ValidatePermissionMode(prefs.PermissionMode) {
		slog.Warn("invalid permission_mode in preferences; resetting to empty (allowed: default/auto/acceptEdits)",
			"path", path,
			"value", prefs.PermissionMode,
		)
		prefs.PermissionMode = ""
	}
	return prefs, nil
}

// migrateOldFile renames a legacy dot-prefixed file to its new name
// if the old file exists and the new file does not.
func migrateOldFile(dir, oldName, newName string) {
	oldPath := filepath.Join(dir, oldName)
	newPath := filepath.Join(dir, newName)

	// Only migrate if old file exists and new file does not
	if _, err := os.Stat(oldPath); err != nil {
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		return // new file already exists, skip
	}
	_ = os.Rename(oldPath, newPath)
}

// WritePreferences saves the preferences for a profile.
// Creates the profile directory if it does not exist.
//
// 쓰기 시점에도 PermissionMode 화이트리스트를 강제한다. 위저드는
// huh.Select로 이미 옵션을 제한하지만, 다른 호출 경로(테스트 픽스처 등)가
// 우회 값을 저장하는 것을 막는다.
func WritePreferences(profileName string, prefs ProfilePreferences) error {
	if !ValidatePermissionMode(prefs.PermissionMode) {
		return fmt.Errorf("invalid permission_mode %q (allowed: default/auto/acceptEdits)", prefs.PermissionMode)
	}
	path := GetPreferencesPath(profileName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	data, err := yaml.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("marshal preferences: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write preferences: %w", err)
	}
	return nil
}
