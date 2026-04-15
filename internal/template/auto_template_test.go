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
