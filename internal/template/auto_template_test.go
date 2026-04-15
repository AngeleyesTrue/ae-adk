package template

import (
	"io/fs"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAutoTemplateConfigRendering(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".ae/config/sections/auto.yaml")
	if err != nil {
		t.Fatalf("read auto.yaml template: %v", err)
	}

	// Verify it parses as valid YAML
	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("auto.yaml is not valid YAML: %v", err)
	}

	// Verify top-level key is "auto"
	if _, ok := parsed["auto"]; !ok {
		t.Error("auto.yaml should have 'auto' as top-level key")
	}

	// Verify expected default values are present in the content
	content := string(data)
	expectedValues := []string{
		"context_isolated",
		"sync_review_iterations",
		"copilot",
		"teammate",
		"final_merge",
		"strategy",
		"squash",
	}
	for _, val := range expectedValues {
		if !strings.Contains(content, val) {
			t.Errorf("auto.yaml should contain %q", val)
		}
	}
}

func TestAutoTemplateSkillRegistration(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/SKILL.md")
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}

	content := string(data)

	// Verify "auto" appears as a subcommand in Priority 1
	if !strings.Contains(content, "**auto**") {
		t.Error("SKILL.md should contain '**auto**' as a subcommand in Priority 1")
	}

	// Verify alias "pipeline" is listed
	if !strings.Contains(content, "pipeline") {
		t.Error("SKILL.md should contain 'pipeline' as an alias for auto")
	}
}

func TestAutoTemplateCLAUDERegistration(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, "CLAUDE.md")
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}

	content := string(data)

	// Verify "auto" appears in the Section 3 Subcommands line (not just anywhere in the file)
	if !strings.Contains(content, "Subcommands: plan, run, sync, auto,") {
		t.Error("CLAUDE.md Section 3 should list 'auto' in the Subcommands line")
	}
}

func TestSkillAllowedToolsAgent(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/SKILL.md")
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}

	content := string(data)

	// Regression: allowed-tools must use "Agent", not "Task"
	// Bug: "Task" prevented AE skill from delegating to sub-agents
	if strings.Contains(content, "allowed-tools: Task,") {
		t.Error("SKILL.md allowed-tools must use 'Agent', not 'Task' — Agent tool is required for sub-agent delegation")
	}
	if !strings.Contains(content, "allowed-tools: Agent,") {
		t.Error("SKILL.md should have 'Agent' as the first entry in allowed-tools")
	}
}

func TestProjectWorkflowPhaseRouting(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/project.md")
	if err != nil {
		t.Fatalf("read project.md: %v", err)
	}

	content := string(data)

	// Regression: "New Project" must route to Phase 0.3 (Deep Interview), not Phase 0.5
	// Bug: Phase 0.5 routing skipped the interview entirely
	if strings.Contains(content, "New Project selected: Proceed to Phase 0.5") {
		t.Error("project.md must route New Project to Phase 0.3, not Phase 0.5 — Phase 0.5 skips the interview")
	}
	if !strings.Contains(content, "New Project selected: Proceed to Phase 0.3") {
		t.Error("project.md should route New Project selection to Phase 0.3 (Deep Interview)")
	}

	// Verify Phase 0.3 section actually exists
	if !strings.Contains(content, "## Phase 0.3:") {
		t.Error("project.md must define Phase 0.3 section")
	}
}

func TestAutoTemplateCommandRegistration(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Verify auto command template exists
	data, err := fs.ReadFile(fsys, ".claude/commands/ae/auto.md.tmpl")
	if err != nil {
		t.Fatalf("read auto.md.tmpl command: %v", err)
	}

	content := string(data)

	// Verify it delegates to ae skill with auto subcommand
	if !strings.Contains(content, `Skill("ae")`) {
		t.Error("auto.md.tmpl should delegate to Skill(\"ae\")")
	}
	if !strings.Contains(content, "auto $ARGUMENTS") {
		t.Error("auto.md.tmpl should pass 'auto $ARGUMENTS'")
	}

	// Verify frontmatter has description and argument-hint
	if !strings.Contains(content, "description:") {
		t.Error("auto.md.tmpl should have description in frontmatter")
	}
	if !strings.Contains(content, "argument-hint:") {
		t.Error("auto.md.tmpl should have argument-hint in frontmatter")
	}
}

func TestAutoTemplateWorkflowSkeleton(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md workflow: %v", err)
	}

	content := string(data)

	// Verify frontmatter exists
	if !strings.HasPrefix(content, "---") {
		t.Error("auto.md should start with YAML frontmatter")
	}

	// Verify name is present
	if !strings.Contains(content, "name:") {
		t.Error("auto.md should contain 'name:' in frontmatter")
	}

	// Verify it mentions auto in the name or description
	if !strings.Contains(content, "auto") {
		t.Error("auto.md should reference 'auto'")
	}
}
