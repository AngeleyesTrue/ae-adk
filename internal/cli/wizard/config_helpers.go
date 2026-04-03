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
// yaml.Node 기반으로 파싱하여 기존 주석과 포맷을 보존한다.
// git_strategy 최상위의 commit_scopes와 각 모드(manual/personal/team)의
// commit_style.scopes 필드를 모두 업데이트한다.
func SaveScopesToConfig(projectRoot string, scopes []string) error {
	if len(scopes) == 0 {
		return nil
	}

	gitStrategyPath := filepath.Join(aeSectionsDir(projectRoot), "git-strategy.yaml")
	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		return fmt.Errorf("read git-strategy.yaml: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse git-strategy.yaml: %w", err)
	}

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return fmt.Errorf("invalid yaml document structure")
	}

	gsNode := findYAMLMapValue(doc.Content[0], "git_strategy")
	if gsNode == nil {
		return fmt.Errorf("git_strategy section not found")
	}

	// commit_scopes를 git_strategy 최상위에 설정
	setYAMLMapValue(gsNode, "commit_scopes", buildYAMLStringSequence(scopes))

	// 각 모드의 commit_style.scopes도 동기화
	for _, mode := range []string{"manual", "personal", "team"} {
		modeNode := findYAMLMapValue(gsNode, mode)
		if modeNode == nil {
			continue
		}
		csNode := findYAMLMapValue(modeNode, "commit_style")
		if csNode == nil {
			continue
		}
		setYAMLMapValue(csNode, "scopes", buildYAMLStringSequence(scopes))
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("marshal git-strategy.yaml: %w", err)
	}

	return os.WriteFile(gitStrategyPath, out, 0644)
}

// findYAMLMapValue는 mapping 노드에서 key에 해당하는 value 노드를 찾는다.
func findYAMLMapValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// setYAMLMapValue는 mapping 노드에서 key의 value를 설정하거나 새로 추가한다.
// mapping이 nil이거나 MappingNode가 아니면 아무 작업도 하지 않는다.
func setYAMLMapValue(mapping *yaml.Node, key string, value *yaml.Node) {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	// 키가 없으면 새로 추가
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
		value,
	)
}

// buildYAMLStringSequence는 문자열 슬라이스로부터 YAML 시퀀스 노드를 생성한다.
func buildYAMLStringSequence(items []string) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
	for _, item := range items {
		seq.Content = append(seq.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: item,
			Tag:   "!!str",
		})
	}
	return seq
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
