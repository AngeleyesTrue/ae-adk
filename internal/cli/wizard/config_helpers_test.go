package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSaveScopesToConfig(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 기존 git-strategy.yaml 생성
	initial := `git_strategy:
  mode: "manual"
  provider: "github"
`
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// scope 저장
	err := SaveScopesToConfig(dir, []string{"Web", "Auth", "DB"})
	if err != nil {
		t.Fatalf("SaveScopesToConfig: %v", err)
	}

	// 결과 검증
	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	gs, ok := parsed["git_strategy"].(map[string]any)
	if !ok {
		t.Fatal("git_strategy section not found")
	}

	scopesRaw, ok := gs["commit_scopes"]
	if !ok {
		t.Fatal("commit_scopes not found in git_strategy")
	}

	scopes, ok := scopesRaw.([]any)
	if !ok {
		t.Fatalf("commit_scopes type = %T, want []any", scopesRaw)
	}

	if len(scopes) != 3 {
		t.Fatalf("commit_scopes length = %d, want 3", len(scopes))
	}

	expected := []string{"Web", "Auth", "DB"}
	for i, s := range scopes {
		if s.(string) != expected[i] {
			t.Errorf("scopes[%d] = %q, want %q", i, s, expected[i])
		}
	}
}

func TestSaveScopesToConfig_EmptyScopes(t *testing.T) {
	// 빈 scope 목록이면 아무 작업도 하지 않음
	err := SaveScopesToConfig(t.TempDir(), nil)
	if err != nil {
		t.Errorf("SaveScopesToConfig(empty) should return nil, got %v", err)
	}
}

func TestSaveScopesToConfig_PreservesExistingFields(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 기존 필드가 있는 git-strategy.yaml 생성
	initial := `git_strategy:
  mode: "team"
  provider: "gitlab"
  auto_merge: true
`
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// scope 저장
	err := SaveScopesToConfig(dir, []string{"API", "Core"})
	if err != nil {
		t.Fatalf("SaveScopesToConfig: %v", err)
	}

	// 기존 필드가 보존되었는지 검증
	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	gs, ok := parsed["git_strategy"].(map[string]any)
	if !ok {
		t.Fatal("git_strategy section not found")
	}

	// 기존 필드 확인
	if gs["mode"] != "team" {
		t.Errorf("mode = %v, want 'team'", gs["mode"])
	}
	if gs["provider"] != "gitlab" {
		t.Errorf("provider = %v, want 'gitlab'", gs["provider"])
	}
	if gs["auto_merge"] != true {
		t.Errorf("auto_merge = %v, want true", gs["auto_merge"])
	}

	// 새 필드 확인
	scopesRaw, ok := gs["commit_scopes"]
	if !ok {
		t.Fatal("commit_scopes not found")
	}
	scopes := scopesRaw.([]any)
	if len(scopes) != 2 {
		t.Fatalf("commit_scopes length = %d, want 2", len(scopes))
	}
}

func TestSaveScopesToConfig_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 주석이 포함된 git-strategy.yaml 생성
	initial := `# Git Strategy Configuration
git_strategy:
  mode: "manual" # 운영 모드
  provider: "github"
  # 브랜치 설정
  auto_merge: true
`
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := SaveScopesToConfig(dir, []string{"Web", "Auth"})
	if err != nil {
		t.Fatalf("SaveScopesToConfig: %v", err)
	}

	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// 주석이 보존되었는지 확인
	if !strings.Contains(content, "# Git Strategy Configuration") {
		t.Error("top-level comment should be preserved")
	}
	if !strings.Contains(content, "# 운영 모드") {
		t.Error("inline comment should be preserved")
	}
	if !strings.Contains(content, "# 브랜치 설정") {
		t.Error("section comment should be preserved")
	}

	// scope도 정상 저장되었는지 확인
	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	gs := parsed["git_strategy"].(map[string]any)
	scopesRaw, ok := gs["commit_scopes"]
	if !ok {
		t.Fatal("commit_scopes not found")
	}
	scopes := scopesRaw.([]any)
	if len(scopes) != 2 {
		t.Fatalf("commit_scopes length = %d, want 2", len(scopes))
	}
}

func TestSaveScopesToConfig_NoGitStrategyKey(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// git_strategy 키가 없는 YAML
	initial := `other_section:
  key: value
`
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := SaveScopesToConfig(dir, []string{"Web"})
	if err == nil {
		t.Fatal("expected error when git_strategy key is missing")
	}
	if !strings.Contains(err.Error(), "git_strategy section not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveScopesToConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	err := SaveScopesToConfig(dir, []string{"Web"})
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestSaveScopesToConfig_FlowStyleOutput(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	initial := `git_strategy:
  mode: "manual"
`
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := SaveScopesToConfig(dir, []string{"Web", "Auth"})
	if err != nil {
		t.Fatalf("SaveScopesToConfig: %v", err)
	}

	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// FlowStyle이면 인라인 형식이어야 함: [Web, Auth]
	if !strings.Contains(content, "[Web, Auth]") {
		t.Errorf("expected flow-style sequence [Web, Auth], got:\n%s", content)
	}
}

func TestSaveScopesToConfig_SyncsToModeCommitStyle(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".ae", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 모드별 commit_style.scopes가 있는 git-strategy.yaml 생성
	initial := `git_strategy:
  mode: "personal"
  manual:
    commit_style:
      format: bracket-scope
      scopes: []
  personal:
    commit_style:
      format: bracket-scope
      scopes: []
  team:
    commit_style:
      format: bracket-scope
      scopes: []
`
	gitStrategyPath := filepath.Join(sectionsDir, "git-strategy.yaml")
	if err := os.WriteFile(gitStrategyPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := SaveScopesToConfig(dir, []string{"Web", "Auth"})
	if err != nil {
		t.Fatalf("SaveScopesToConfig: %v", err)
	}

	data, err := os.ReadFile(gitStrategyPath)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	gs := parsed["git_strategy"].(map[string]any)

	// commit_scopes 최상위 확인
	topScopes := gs["commit_scopes"].([]any)
	if len(topScopes) != 2 {
		t.Fatalf("top-level commit_scopes length = %d, want 2", len(topScopes))
	}

	// 각 모드의 commit_style.scopes 동기화 확인
	for _, mode := range []string{"manual", "personal", "team"} {
		modeSection := gs[mode].(map[string]any)
		commitStyle := modeSection["commit_style"].(map[string]any)
		modeScopes := commitStyle["scopes"].([]any)
		if len(modeScopes) != 2 {
			t.Errorf("%s.commit_style.scopes length = %d, want 2", mode, len(modeScopes))
		}
		if modeScopes[0].(string) != "Web" {
			t.Errorf("%s.commit_style.scopes[0] = %q, want %q", mode, modeScopes[0], "Web")
		}
		if modeScopes[1].(string) != "Auth" {
			t.Errorf("%s.commit_style.scopes[1] = %q, want %q", mode, modeScopes[1], "Auth")
		}
	}
}
