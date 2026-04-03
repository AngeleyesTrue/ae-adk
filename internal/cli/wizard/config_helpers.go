package wizard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// aeSectionsDir returns the path to the sections directory relative to the project root.
func aeSectionsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".ae", "config", "sections")
}

// ReadLocaleFromProject reads conversation_language from language.yaml.
// Returns an empty string if the file is missing or parsing fails.
func ReadLocaleFromProject(projectRoot string) string {
	langPath := filepath.Join(aeSectionsDir(projectRoot), "language.yaml")
	data, err := os.ReadFile(langPath)
	if err != nil {
		return ""
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return ""
	}

	langSection, ok := parsed["language"].(map[string]any)
	if !ok {
		return ""
	}

	locale, _ := langSection["conversation_language"].(string)
	return locale
}

// ReadGitHubUsernameFromConfig reads github_username from user.yaml.
// Returns an empty string if the file is missing or the value is absent.
func ReadGitHubUsernameFromConfig(projectRoot string) string {
	return readUserField(projectRoot, "github_username")
}

// ReadGitLabUsernameFromConfig reads gitlab_username from user.yaml.
// Returns an empty string if the file is missing or the value is absent.
func ReadGitLabUsernameFromConfig(projectRoot string) string {
	return readUserField(projectRoot, "gitlab_username")
}

// readUserField reads a specific field from the user section of user.yaml.
func readUserField(projectRoot, field string) string {
	userPath := filepath.Join(aeSectionsDir(projectRoot), "user.yaml")
	data, err := os.ReadFile(userPath)
	if err != nil {
		return ""
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return ""
	}

	userSection, ok := parsed["user"].(map[string]any)
	if !ok {
		return ""
	}

	value, _ := userSection[field].(string)
	return value
}

// SaveScopesToConfig는 commit scope 목록을 git-strategy.yaml에 저장한다.
// 기존 파일 내용은 보존하고 commit_style.scopes 필드만 추가/업데이트한다.
func SaveScopesToConfig(projectRoot string, scopes []string) error {
	if len(scopes) == 0 {
		return nil
	}

	gitStrategyPath := filepath.Join(aeSectionsDir(projectRoot), "git-strategy.yaml")
	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		return fmt.Errorf("read git-strategy.yaml: %w", err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("parse git-strategy.yaml: %w", err)
	}

	gsSection, ok := parsed["git_strategy"].(map[string]any)
	if !ok {
		return fmt.Errorf("git_strategy section not found")
	}

	// commit_scopes를 git_strategy 최상위에 저장
	gsSection["commit_scopes"] = scopes
	parsed["git_strategy"] = gsSection

	out, err := yaml.Marshal(parsed)
	if err != nil {
		return fmt.Errorf("marshal git-strategy.yaml: %w", err)
	}

	return os.WriteFile(gitStrategyPath, out, 0644)
}

// IsGhAuthenticated checks whether the gh CLI is authenticated.
// Returns false if gh is not installed or not authenticated.
func IsGhAuthenticated() bool {
	// Check if the gh CLI is present
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}

	// Run gh auth status — exits 0 when authenticated
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}
