package template

import (
	"io/fs"
	"strings"
	"testing"
)

// =============================================================================
// SPEC-UPDATE-002 Verification Tests (TDD RED Phase)
// All tests should FAIL before implementation and PASS after.
// =============================================================================

// --- AC-01: CLAUDE.md v14.0.0 Synchronization ---

func TestUpdate002_CLAUDEmd_AgencySection(t *testing.T) {
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

	// Section 3: /agency command reference
	if !strings.Contains(content, "/agency") {
		t.Error("CLAUDE.md should contain /agency command reference (Section 3)")
	}

	// Section 4: Agency Agents
	for _, agent := range []string{"planner", "copywriter", "designer", "learner"} {
		if !strings.Contains(content, agent) {
			t.Errorf("CLAUDE.md should mention Agency agent: %s", agent)
		}
	}

	// Section 6: Harness-Based Quality Routing
	if !strings.Contains(content, "Harness") {
		t.Error("CLAUDE.md should contain Harness-Based Quality Routing section")
	}

	// Section 9: Agency Configuration
	if !strings.Contains(content, "Agency Configuration") || !strings.Contains(content, ".agency/") {
		t.Error("CLAUDE.md should contain Agency Configuration reference")
	}
}

func TestUpdate002_CLAUDEmd_EvaluatorActive(t *testing.T) {
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

	if !strings.Contains(content, "evaluator-active") {
		t.Error("CLAUDE.md should mention evaluator-active in Evaluator Agents section")
	}
}

func TestUpdate002_CLAUDEmd_AeNaming(t *testing.T) {
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

	// Check completion markers use <ae> not <moai>
	if strings.Contains(content, "<moai>") {
		t.Error("CLAUDE.md should use <ae> completion markers, not <moai>")
	}

	// Paths should use .ae/ not .moai/
	if strings.Contains(content, ".moai/") {
		t.Error("CLAUDE.md should use .ae/ paths, not .moai/")
	}
}

func TestUpdate002_CLAUDEmd_NoCGMode(t *testing.T) {
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

	// CG Mode section should not exist
	if strings.Contains(content, "CG Mode") && strings.Contains(content, "GLM") {
		// Only flag if there's a dedicated CG Mode section (not just a brief exclusion mention)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "### CG Mode") || strings.Contains(line, "## CG Mode") {
				t.Error("CLAUDE.md should not contain a CG Mode section (ae-adk does not use GLM)")
			}
		}
	}
}

func TestUpdate002_CLAUDEmd_VersionUpdated(t *testing.T) {
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

	if !strings.Contains(content, "14.0.0") {
		t.Error("CLAUDE.md version should be updated to 14.0.0")
	}
}

// --- AC-02: Harness Design Files ---

func TestUpdate002_HarnessYaml(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".ae/config/sections/harness.yaml")
	if err != nil {
		t.Fatalf("harness.yaml should exist: %v", err)
	}
	content := string(data)

	for _, level := range []string{"minimal", "standard", "thorough"} {
		if !strings.Contains(content, level) {
			t.Errorf("harness.yaml should define %q level", level)
		}
	}
}

func TestUpdate002_ConstitutionYaml(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	_, err = fs.ReadFile(fsys, ".ae/config/sections/constitution.yaml")
	if err != nil {
		t.Fatalf("constitution.yaml should exist: %v", err)
	}
}

func TestUpdate002_EvaluatorProfiles(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	profiles := []string{"default.yaml", "strict.yaml", "lenient.yaml", "frontend.yaml"}
	for _, profile := range profiles {
		path := ".ae/config/evaluator-profiles/" + profile
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			t.Errorf("evaluator profile %s should exist: %v", profile, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("evaluator profile %s should not be empty", profile)
		}
	}
}

// --- AC-03: evaluator-active Agent ---

func TestUpdate002_EvaluatorActiveAgent(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/agents/ae/evaluator-active.md")
	if err != nil {
		t.Fatalf("evaluator-active.md should exist: %v", err)
	}
	content := string(data)

	// 4-dimension scoring
	dimensions := []string{"Functionality", "Security", "Craft", "Consistency"}
	for _, dim := range dimensions {
		if !strings.Contains(content, dim) {
			t.Errorf("evaluator-active.md should mention scoring dimension: %s", dim)
		}
	}

	// Should use ae naming
	if strings.Contains(content, "moai") && !strings.Contains(content, "moai-adk") {
		t.Error("evaluator-active.md should use ae naming convention")
	}
}

func TestUpdate002_AgentCount(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	entries, err := fs.ReadDir(fsys, ".claude/agents/ae")
	if err != nil {
		t.Fatalf("ReadDir agents/ae: %v", err)
	}

	var mdCount int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mdCount++
		}
	}

	// 19 existing + 1 evaluator-active + 6 agency agents = 26
	if mdCount < 26 {
		t.Errorf("expected at least 26 agent .md files (19 existing + 1 evaluator-active + 6 agency), got %d", mdCount)
	}
}

// --- AC-04: Hook Events Extension ---

func TestUpdate002_HookEventCount(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/settings.json.tmpl")
	if err != nil {
		t.Fatalf("read settings.json.tmpl: %v", err)
	}
	content := string(data)

	// Count unique hook event names (keys under "hooks")
	expectedEvents := []string{
		"SessionStart", "PreCompact", "SessionEnd",
		"PreToolUse", "PostToolUse", "Stop",
		"SubagentStop", "PostToolUseFailure", "Notification",
		"SubagentStart", "UserPromptSubmit", "PermissionRequest",
		"TeammateIdle", "TaskCompleted", "WorktreeCreate", "WorktreeRemove",
		// New events (8)
		"StopFailure", "PostCompact", "InstructionsLoaded", "CwdChanged",
		"FileChanged", "Elicitation", "ElicitationResult", "PermissionDenied",
	}

	var foundCount int
	for _, event := range expectedEvents {
		// Look for the event as a JSON key in the hooks section
		if strings.Contains(content, "\""+event+"\"") {
			foundCount++
		}
	}

	if foundCount < 24 {
		t.Errorf("expected at least 24 hook events registered, found %d", foundCount)
	}
}

func TestUpdate002_PermissionDeniedHandler(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	_, err = fs.ReadFile(fsys, ".claude/hooks/ae/handle-permission-denied.sh")
	if err != nil {
		// Try .tmpl extension
		_, err = fs.ReadFile(fsys, ".claude/hooks/ae/handle-permission-denied.sh.tmpl")
		if err != nil {
			t.Error("handle-permission-denied.sh(.tmpl) should exist in .claude/hooks/ae/")
		}
	}
}

// --- AC-05: Workflow Improvements ---

func TestUpdate002_DriftGuard(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/rules/ae/workflow/workflow-modes.md")
	if err != nil {
		t.Fatalf("read workflow-modes.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "Drift Guard") {
		t.Error("workflow-modes.md should contain Drift Guard section")
	}
	if !strings.Contains(content, "30%") {
		t.Error("workflow-modes.md Drift Guard should specify 30% drift threshold")
	}
}

func TestUpdate002_ManagerSpecWhatWhy(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/agents/ae/manager-spec.md")
	if err != nil {
		t.Fatalf("read manager-spec.md: %v", err)
	}
	content := string(data)

	// What/Why boundary validation
	if !strings.Contains(content, "What") || !strings.Contains(content, "Why") {
		t.Error("manager-spec.md should contain What/Why boundary validation")
	}
}

// --- AC-06: manager-quality Model Change ---

func TestUpdate002_ManagerQualityModel(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/agents/ae/manager-quality.md")
	if err != nil {
		t.Fatalf("read manager-quality.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "model: sonnet") {
		t.Error("manager-quality.md model should be 'sonnet' (was 'haiku')")
	}
	if strings.Contains(content, "model: haiku") {
		t.Error("manager-quality.md should NOT have model: haiku (should be sonnet)")
	}
}

// --- AC-08: Agency v3.2 Integration ---

func TestUpdate002_AgencyAgents(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	agencyAgents := []string{
		"agency-planner.md",
		"agency-copywriter.md",
		"agency-designer.md",
		"agency-builder.md",
		"agency-evaluator.md",
		"agency-learner.md",
	}

	for _, agent := range agencyAgents {
		path := ".claude/agents/ae/" + agent
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			t.Errorf("Agency agent %s should exist: %v", agent, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("Agency agent %s should not be empty", agent)
		}
		// Verify ae naming
		content := string(data)
		if strings.Contains(content, "moai-adk") {
			// moai-adk as upstream reference is OK
		} else if strings.Contains(content, "moai") && !strings.Contains(content, "moai-adk") {
			t.Errorf("Agency agent %s should use ae naming, found 'moai' reference", agent)
		}
	}
}

func TestUpdate002_AgencySkills(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	skillDirs := []string{
		"ae-agency-copywriting",
		"ae-agency-design-system",
		"ae-agency-evaluation-criteria",
		"ae-agency-frontend-patterns",
		"ae-agency-client-interview",
	}

	for _, dir := range skillDirs {
		path := ".claude/skills/" + dir + "/SKILL.md"
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			t.Errorf("Agency skill %s should exist: %v", path, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("Agency skill %s should not be empty", dir)
		}
	}
}

func TestUpdate002_AgencyCommand(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Agency command template
	_, err = fs.ReadFile(fsys, ".claude/commands/ae/agency.md.tmpl")
	if err != nil {
		// Try without .tmpl
		_, err = fs.ReadFile(fsys, ".claude/commands/ae/agency.md")
		if err != nil {
			t.Error("Agency command (agency.md or agency.md.tmpl) should exist in .claude/commands/ae/")
		}
	}
}

func TestUpdate002_AgencyConstitution(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/rules/ae/agency/constitution.md")
	if err != nil {
		t.Fatalf("Agency constitution should exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Agency") {
		t.Error("Agency constitution should mention 'Agency'")
	}
	if !strings.Contains(content, "GAN Loop") || !strings.Contains(content, "Builder") {
		t.Error("Agency constitution should describe GAN Loop and Builder-Evaluator pattern")
	}
}

func TestUpdate002_AgencyDefaultConfig(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	configFiles := []string{
		".agency/config.yaml",
		".agency/fork-manifest.yaml",
	}

	for _, path := range configFiles {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			t.Errorf("Agency config %s should exist: %v", path, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("Agency config %s should not be empty", path)
		}
	}

	// Context files
	contextFiles := []string{
		".agency/context/brand-voice.md",
		".agency/context/visual-identity.md",
		".agency/context/target-audience.md",
		".agency/context/tech-preferences.md",
		".agency/context/quality-standards.md",
	}

	for _, path := range contextFiles {
		_, err := fs.ReadFile(fsys, path)
		if err != nil {
			t.Errorf("Agency context %s should exist: %v", path, err)
		}
	}

	// Brief template
	_, err = fs.ReadFile(fsys, ".agency/templates/brief-template.md")
	if err != nil {
		t.Error("Agency brief template should exist at .agency/templates/brief-template.md")
	}
}

func TestUpdate002_AgencyNaming(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Walk all agency-related files and check naming
	agencyPaths := []string{
		".claude/agents/ae/agency-planner.md",
		".claude/agents/ae/agency-copywriter.md",
		".claude/agents/ae/agency-designer.md",
		".claude/agents/ae/agency-builder.md",
		".claude/agents/ae/agency-evaluator.md",
		".claude/agents/ae/agency-learner.md",
	}

	for _, path := range agencyPaths {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			continue // Existence already tested above
		}
		content := string(data)

		// Should NOT contain .moai/ paths (should be .ae/)
		if strings.Contains(content, ".moai/") {
			t.Errorf("Agency file %s should not contain .moai/ paths", path)
		}
	}
}

// --- AC-09: Exclusion Verification ---

func TestUpdate002_NoGLMReferences(t *testing.T) {
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

	// No GLM compatibility automation
	if strings.Contains(content, "GLM_") || strings.Contains(content, "CLAUDE_GLM") {
		t.Error("CLAUDE.md should not contain GLM environment variable references")
	}
}

func TestUpdate002_NoRemovedCommands(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Check no ae cc, ae glm, ae cg commands exist
	removedCommands := []string{
		".claude/commands/ae/cc.md",
		".claude/commands/ae/glm.md",
		".claude/commands/ae/cg.md",
		".claude/commands/ae/cc.md.tmpl",
		".claude/commands/ae/glm.md.tmpl",
		".claude/commands/ae/cg.md.tmpl",
	}

	for _, cmd := range removedCommands {
		if _, err := fs.ReadFile(fsys, cmd); err == nil {
			t.Errorf("removed command %s should not exist", cmd)
		}
	}
}

func TestUpdate002_NoMoaiCompletionMarkers(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var violations []string
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		// Only check text files
		if !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			return nil
		}
		content := string(data)
		if strings.Contains(content, "<moai>") {
			violations = append(violations, path)
		}
		return nil
	})

	if len(violations) > 0 {
		t.Errorf("found <moai> completion markers in %d files (should be <ae>): %v",
			len(violations), violations)
	}
}

// --- AC-10: Template Integrity ---

func TestUpdate002_TemplateFileCount(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var totalFiles int
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path != "." && !d.IsDir() {
			totalFiles++
		}
		return nil
	})

	// 494 original + ~29 new = ~523 minimum
	if totalFiles < 520 {
		t.Errorf("expected at least 520 embedded files (494 original + ~29 new), got %d", totalFiles)
	}
	t.Logf("total embedded files: %d", totalFiles)
}

func TestUpdate002_NoUnintendedMoaiReferences(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Allowlist: files that legitimately reference "moai" (e.g., upstream tracking)
	allowlist := map[string]bool{
		".agency/fork-manifest.yaml": true, // upstream references to moai-adk
	}

	var violations []string
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if allowlist[path] {
			return nil
		}
		// Check CLAUDE.md and agent definitions for unintended moai references
		if path == "CLAUDE.md" || strings.HasPrefix(path, ".claude/agents/ae/") {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				return nil
			}
			content := string(data)
			// moai-adk as upstream reference is allowed; bare "moai" command references are not
			// Check for patterns like "moai plan", "moai run", "/moai" etc
			if strings.Contains(content, "/moai ") || strings.Contains(content, "/moai\n") {
				violations = append(violations, path+" (contains /moai command reference)")
			}
			if strings.Contains(content, "moai plan") || strings.Contains(content, "moai run") || strings.Contains(content, "moai sync") {
				violations = append(violations, path+" (contains 'moai plan/run/sync' reference)")
			}
		}
		return nil
	})

	if len(violations) > 0 {
		t.Errorf("found unintended 'moai' references in %d files: %v",
			len(violations), violations)
	}
}

func TestUpdate002_AeLangCsharpPreserved(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// ae-lang-csharp skill must still exist (SPEC-SKILL-001 preservation)
	_, err = fs.ReadFile(fsys, ".claude/skills/ae-lang-csharp/SKILL.md")
	if err != nil {
		t.Error("ae-lang-csharp skill should be preserved unchanged")
	}
}

func TestUpdate002_DynamicTeamGenerationPreserved(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// No static team-* agent files should exist
	entries, err := fs.ReadDir(fsys, ".claude/agents/ae")
	if err != nil {
		t.Fatalf("ReadDir agents/ae: %v", err)
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "team-") {
			t.Errorf("static team agent file %s should not exist (dynamic generation preserved)", e.Name())
		}
	}
}

func TestUpdate002_BracketScopeConventionPreserved(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/rules/ae/workflow/commit-convention.md")
	if err != nil {
		t.Fatalf("read commit-convention.md: %v", err)
	}

	if !strings.Contains(string(data), "Bracket-Scope") {
		t.Error("bracket-scope commit convention should be preserved")
	}
}

// --- Spec-Workflow Phase 2.0 ---

func TestUpdate002_SprintContract(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/rules/ae/workflow/spec-workflow.md")
	if err != nil {
		t.Fatalf("read spec-workflow.md: %v", err)
	}
	content := string(data)

	// Phase 2.0 sprint contract should be referenced
	if !strings.Contains(content, "Sprint") || !strings.Contains(content, "contract") {
		t.Error("spec-workflow.md should reference sprint contract (Phase 2.0)")
	}
}
