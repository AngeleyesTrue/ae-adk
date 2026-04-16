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

// ============================================================================
// SPEC-PIPELINE-002: Structural Workflow Separation Tests
// ============================================================================

// --- Structural Separation Tests (core invariants) ---

// TestRunWorkflowNoPhase4 verifies run.md does NOT contain "### Phase 4"
// header. This is the core structural invariant — if Phase 4 exists,
// cascade is possible. (AC-01)
func TestRunWorkflowNoPhase4(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/run.md")
	if err != nil {
		t.Fatalf("read run.md: %v", err)
	}

	content := string(data)

	if strings.Contains(content, "### Phase 4") {
		t.Error("run.md must NOT contain '### Phase 4' — this creates a cascade trigger in auto pipeline")
	}

	if strings.Contains(content, "Sync Documentation") {
		t.Error("run.md must NOT contain 'Sync Documentation' AskUserQuestion option — cascade trigger removed")
	}
}

// TestRunWorkflowPhase3Preserved verifies run.md still contains "### Phase 3"
// (git operations) to ensure we didn't accidentally remove too much. (AC-01)
func TestRunWorkflowPhase3Preserved(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/run.md")
	if err != nil {
		t.Fatalf("read run.md: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "### Phase 3") {
		t.Error("run.md must still contain '### Phase 3' (git operations)")
	}
}

// TestRunWorkflowCompletionCriteriaNoPhase4 verifies run.md Completion Criteria
// section does NOT reference "Phase 4". Catches partial cleanup where the section
// header is removed but the completion criteria reference remains. (AC-01)
func TestRunWorkflowCompletionCriteriaNoPhase4(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/run.md")
	if err != nil {
		t.Fatalf("read run.md: %v", err)
	}

	content := string(data)

	// Extract Completion Criteria section (up to the next --- separator,
	// not including version footer which may mention "Phase 4" in changelog)
	ccIdx := strings.Index(content, "## Completion Criteria")
	if ccIdx == -1 {
		t.Fatal("run.md must contain '## Completion Criteria' section")
	}
	ccSection := content[ccIdx:]
	if sepIdx := strings.Index(ccSection, "\n---"); sepIdx != -1 {
		ccSection = ccSection[:sepIdx]
	}

	if strings.Contains(ccSection, "Phase 4") {
		t.Error("run.md Completion Criteria must NOT reference 'Phase 4'")
	}
}

// TestAutoSyncWorkflowExists verifies auto-sync.md exists in embedded templates
// and has valid frontmatter (name: ae-workflow-auto-sync). (AC-02)
func TestAutoSyncWorkflowExists(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto-sync.md")
	if err != nil {
		t.Fatalf("auto-sync.md must exist in embedded templates: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "name: ae-workflow-auto-sync") {
		t.Error("auto-sync.md must have 'name: ae-workflow-auto-sync' in frontmatter")
	}
}

// TestAutoSyncNoMergeCapability verifies auto-sync.md does NOT contain
// `gh pr merge` command. This is the second core structural invariant. (AC-02)
func TestAutoSyncNoMergeCapability(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto-sync.md")
	if err != nil {
		t.Fatalf("read auto-sync.md: %v", err)
	}

	content := string(data)

	// Check for gh pr merge on each line, skipping only known safe contexts:
	// - The specific [HARD] pipeline safety constraint explanation line
	// - Lines starting with "Source:" (provenance metadata)
	// Any other occurrence is treated as an executable command reference.
	// NOTE: We match the specific [HARD] safety explanation rather than all
	// [HARD] lines to prevent false negatives if a future [HARD] rule
	// accidentally includes a gh pr merge command.
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip the specific pipeline safety constraint explanation line
		if strings.Contains(line, "[HARD] This workflow has NO merge capability") {
			continue
		}
		// Skip provenance metadata line
		if strings.HasPrefix(trimmed, "Source:") {
			continue
		}
		if strings.Contains(line, "gh pr merge") {
			t.Errorf("auto-sync.md line %d must NOT contain 'gh pr merge' as a command: %s", i+1, trimmed)
		}
	}
}

// TestAutoSyncNoPhase4 verifies auto-sync.md does NOT contain
// "### Phase 4" header. (AC-02)
func TestAutoSyncNoPhase4(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto-sync.md")
	if err != nil {
		t.Fatalf("read auto-sync.md: %v", err)
	}

	content := string(data)

	if strings.Contains(content, "### Phase 4") {
		t.Error("auto-sync.md must NOT contain '### Phase 4'")
	}

	if strings.Contains(content, "Phase 4:") {
		t.Error("auto-sync.md must NOT contain 'Phase 4:' header")
	}
}

// TestAutoSyncNoMergeFlag verifies auto-sync.md Supported Flags section does NOT
// list `--merge` as a supported flag. Section-scoped check to avoid false negatives
// from explanatory text elsewhere in the file. (AC-02)
func TestAutoSyncNoMergeFlag(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto-sync.md")
	if err != nil {
		t.Fatalf("read auto-sync.md: %v", err)
	}

	content := string(data)

	// Extract the Supported Flags section (from "## Supported Flags" to next "##" header)
	flagsStart := strings.Index(content, "## Supported Flags")
	if flagsStart == -1 {
		t.Fatal("auto-sync.md must contain '## Supported Flags' section")
	}

	afterFlags := content[flagsStart+len("## Supported Flags"):]
	nextSection := strings.Index(afterFlags, "\n## ")
	var flagsSection string
	if nextSection != -1 {
		flagsSection = afterFlags[:nextSection]
	} else {
		flagsSection = afterFlags
	}

	if strings.Contains(flagsSection, "--merge") {
		t.Error("auto-sync.md Supported Flags section must NOT list '--merge' as a supported flag")
	}
}

// TestAutoSyncNoAutoMergeOption verifies auto-sync.md does NOT contain
// "Auto-Merge PR" option text. (AC-02)
func TestAutoSyncNoAutoMergeOption(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto-sync.md")
	if err != nil {
		t.Fatalf("read auto-sync.md: %v", err)
	}

	content := string(data)

	if strings.Contains(content, "Auto-Merge PR") {
		t.Error("auto-sync.md must NOT contain 'Auto-Merge PR' option text")
	}
}

// --- Auto Pipeline Prompt Tests ---

// TestAutoPromptUsesAutoSync verifies auto.md Phase 2 Sync prompt contains
// `/ae auto-sync` (not `/ae sync`). (AC-03)
func TestAutoPromptUsesAutoSync(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "/ae auto-sync") {
		t.Error("auto.md must contain '/ae auto-sync' in Phase 2 Sync prompt")
	}
}

// TestAutoFinalMergeUserGate verifies auto.md Phase 3 contains AskUserQuestion
// BEFORE gh pr merge (strings.Index ordering within Phase 3 section). Also
// verifies gh pr merge appears inside a conditional block gated by user
// approval. Scoped to Phase 3 to avoid false positives from diagram text. (AC-04)
func TestAutoFinalMergeUserGate(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md: %v", err)
	}

	content := string(data)

	// Extract Phase 3 section to avoid false positives from diagram text.
	// Use "## Error Recovery" as end marker because Phase 3's code fence
	// contains "## " headers (e.g., "## Query PR") that would break
	// generic section extraction.
	phase3Idx := strings.Index(content, "## Phase 3: Final Merge")
	if phase3Idx == -1 {
		t.Fatal("auto.md must contain '## Phase 3: Final Merge'")
	}

	phase3Content := content[phase3Idx:]

	errorRecoveryIdx := strings.Index(phase3Content, "## Error Recovery")
	if errorRecoveryIdx != -1 {
		phase3Content = phase3Content[:errorRecoveryIdx]
	}

	askIdx := strings.Index(phase3Content, "AskUserQuestion")
	mergeIdx := strings.Index(phase3Content, "gh pr merge")

	if askIdx == -1 {
		t.Fatal("auto.md Phase 3 must contain 'AskUserQuestion'")
	}
	if mergeIdx == -1 {
		t.Fatal("auto.md Phase 3 must contain 'gh pr merge'")
	}
	if askIdx >= mergeIdx {
		t.Error("auto.md Phase 3: AskUserQuestion must appear BEFORE gh pr merge (index ordering)")
	}

	// Verify merge is inside a conditional block
	conditionalIdx := strings.Index(phase3Content, `IF "Merge PR"`)
	if conditionalIdx == -1 {
		t.Fatal("auto.md Phase 3 must contain 'IF \"Merge PR\"' conditional")
	}
	if conditionalIdx >= mergeIdx {
		t.Error("auto.md Phase 3: 'IF \"Merge PR\"' conditional must appear BEFORE gh pr merge")
	}
	if conditionalIdx <= askIdx {
		t.Error("auto.md Phase 3: 'IF \"Merge PR\"' conditional must appear AFTER AskUserQuestion")
	}
}

// TestAutoFinalMergeOptionsText verifies auto.md Phase 3 AskUserQuestion
// contains the three required option keywords: "Merge PR", "Skip merge",
// and "Abort". Scoped to Phase 3 section to avoid false positives from
// diagram or error recovery text elsewhere. (AC-04)
func TestAutoFinalMergeOptionsText(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md: %v", err)
	}

	content := string(data)

	// Scope to Phase 3 section only
	phase3Idx := strings.Index(content, "## Phase 3: Final Merge")
	if phase3Idx == -1 {
		t.Fatal("auto.md must contain '## Phase 3: Final Merge'")
	}
	phase3Content := content[phase3Idx:]
	if errIdx := strings.Index(phase3Content, "## Error Recovery"); errIdx != -1 {
		phase3Content = phase3Content[:errIdx]
	}

	requiredOptions := []string{"Merge PR", "Skip merge", "Abort"}
	for _, opt := range requiredOptions {
		if !strings.Contains(phase3Content, opt) {
			t.Errorf("auto.md Phase 3 AskUserQuestion must contain option keyword %q", opt)
		}
	}
}

// TestAutoSyncHardRoutingConstraint verifies auto.md Phase 2 Sync teammate
// spawn prompt contains [HARD] and /ae auto-sync in the same prompt block. (AC-03)
func TestAutoSyncHardRoutingConstraint(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md: %v", err)
	}

	content := string(data)

	// Find the Phase 2 Sync section (around the spawn teammate prompt)
	syncPhaseIdx := strings.Index(content, "## Phase 2: Sync-Review Loop")
	if syncPhaseIdx == -1 {
		t.Fatal("auto.md must contain '## Phase 2: Sync-Review Loop'")
	}

	phase3Idx := strings.Index(content, "## Phase 3:")
	if phase3Idx == -1 {
		t.Fatal("auto.md must contain '## Phase 3:'")
	}

	phase2Content := content[syncPhaseIdx:phase3Idx]

	if !strings.Contains(phase2Content, "[HARD]") {
		t.Error("auto.md Phase 2 must contain [HARD] routing constraint")
	}

	if !strings.Contains(phase2Content, "/ae auto-sync") {
		t.Error("auto.md Phase 2 must contain '/ae auto-sync' in sync prompt")
	}
}

// TestAutoPhase2NoUnsafeSyncCommand verifies auto.md Phase 2 spawn prompt
// does NOT contain the bare "/ae sync " command (without "auto-" prefix).
// If both "/ae sync" and "/ae auto-sync" coexist, the cascade risk partially
// returns since the teammate may invoke the unsafe variant.
func TestAutoPhase2NoUnsafeSyncCommand(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md: %v", err)
	}

	content := string(data)

	syncPhaseIdx := strings.Index(content, "## Phase 2: Sync-Review Loop")
	if syncPhaseIdx == -1 {
		t.Fatal("auto.md must contain '## Phase 2: Sync-Review Loop'")
	}

	phase3Idx := strings.Index(content, "## Phase 3:")
	if phase3Idx == -1 {
		t.Fatal("auto.md must contain '## Phase 3:'")
	}

	phase2Content := content[syncPhaseIdx:phase3Idx]

	// Check for bare "/ae sync" (without "auto-" prefix).
	// Catches all forms: "/ae sync ", "/ae sync{", "/ae sync\n", etc.
	lines := strings.Split(phase2Content, "\n")
	for i, line := range lines {
		// Skip lines that contain the safe "/ae auto-sync"
		if strings.Contains(line, "/ae auto-sync") {
			continue
		}
		// Skip lines that are part of the [HARD] constraint explanation
		if strings.Contains(line, "[HARD]") {
			continue
		}
		if strings.Contains(line, "/ae sync") {
			t.Errorf("auto.md Phase 2 line %d must NOT contain bare '/ae sync' command (use '/ae auto-sync'): %s",
				i+1, strings.TrimSpace(line))
		}
	}
}

// TestAutoPhase3CIWaitPresent verifies auto.md Phase 3 contains CI pending
// handling with --watch --fail-fast. (REQ-06 / AC-07)
func TestAutoPhase3CIWaitPresent(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md: %v", err)
	}

	content := string(data)

	// Scope to Phase 3 section
	phase3Idx := strings.Index(content, "## Phase 3: Final Merge")
	if phase3Idx == -1 {
		t.Fatal("auto.md must contain '## Phase 3: Final Merge'")
	}

	phase3Content := content[phase3Idx:]
	if errIdx := strings.Index(phase3Content, "## Error Recovery"); errIdx != -1 {
		phase3Content = phase3Content[:errIdx]
	}

	if !strings.Contains(phase3Content, "--watch") {
		t.Error("auto.md Phase 3 must contain '--watch' for CI pending handling (REQ-06)")
	}

	if !strings.Contains(phase3Content, "--fail-fast") {
		t.Error("auto.md Phase 3 must contain '--fail-fast' for CI pending handling (REQ-06)")
	}

	if !strings.Contains(phase3Content, "checks still pending") {
		t.Error("auto.md Phase 3 must contain CI pending detection logic")
	}
}

// --- Backward Compatibility Tests ---

// TestSyncWorkflowUnchanged_Phase4Preserved verifies sync.md still contains
// "### Phase 4" and "Auto-Merge PR" option text — confirming interactive sync
// is not degraded. (AC-05)
func TestSyncWorkflowUnchanged_Phase4Preserved(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/sync.md")
	if err != nil {
		t.Fatalf("read sync.md: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "### Phase 4") {
		t.Error("sync.md must still contain '### Phase 4' (interactive sync not degraded)")
	}

	if !strings.Contains(content, "Auto-Merge PR") {
		t.Error("sync.md must still contain 'Auto-Merge PR' option text")
	}
}

// TestSyncWorkflowUnchanged_MergePreserved verifies sync.md still contains
// "Step 3.4" and "gh pr merge" — confirming interactive merge capability
// intact. (AC-05)
func TestSyncWorkflowUnchanged_MergePreserved(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/sync.md")
	if err != nil {
		t.Fatalf("read sync.md: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "Step 3.4") {
		t.Error("sync.md must still contain 'Step 3.4' (auto-merge section)")
	}

	if !strings.Contains(content, "gh pr merge") {
		t.Error("sync.md must still contain 'gh pr merge' (merge capability)")
	}
}

// --- Registration Tests ---

// TestAutoSyncSkillRegistration verifies SKILL.md contains "auto-sync"
// subcommand. (AC-06)
func TestAutoSyncSkillRegistration(t *testing.T) {
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

	if !strings.Contains(content, "auto-sync") {
		t.Error("SKILL.md must contain 'auto-sync' subcommand")
	}
}

// TestAutoSyncCLAUDERegistration verifies CLAUDE.md Subcommands line
// contains "auto-sync". Scoped to the Subcommands line to prevent false
// positives from incidental mentions elsewhere. (AC-06)
func TestAutoSyncCLAUDERegistration(t *testing.T) {
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

	// Find the Subcommands line specifically
	lines := strings.Split(content, "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "Subcommands:") && strings.Contains(line, "auto-sync") {
			found = true
			break
		}
	}
	if !found {
		t.Error("CLAUDE.md must contain 'auto-sync' in its Subcommands: line")
	}
}

// --- Diagram Consistency Test ---

// TestAutoPipelineDiagramUsesAutoSync verifies auto.md Pipeline Sequence diagram
// references `/ae auto-sync` (not `/ae sync {spec_id}` without auto- prefix). (AC-08)
func TestAutoPipelineDiagramUsesAutoSync(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto.md")
	if err != nil {
		t.Fatalf("read auto.md: %v", err)
	}

	content := string(data)

	// Find the Pipeline Sequence diagram section
	diagramStart := strings.Index(content, "## Pipeline Sequence")
	if diagramStart == -1 {
		t.Fatal("auto.md must contain '## Pipeline Sequence'")
	}

	// Find end of diagram section (next ## section)
	afterDiagram := content[diagramStart+len("## Pipeline Sequence"):]
	nextSection := strings.Index(afterDiagram, "\n## ")
	var diagramSection string
	if nextSection != -1 {
		diagramSection = afterDiagram[:nextSection]
	} else {
		diagramSection = afterDiagram
	}

	if !strings.Contains(diagramSection, "/ae auto-sync") {
		t.Error("auto.md Pipeline Sequence diagram must reference '/ae auto-sync'")
	}
}

// --- Structural Parity Test ---

// TestAutoSyncStructuralParityWithSync verifies auto-sync.md and sync.md share
// the same phase structure (Phase 0, Phase 1, Phase 2, Phase 3 headers present
// in both). Verifies auto-sync.md does NOT contain Phase 4 while sync.md does.
// This detects maintenance drift when sync.md adds or renames phases. (AC-09)
func TestAutoSyncStructuralParityWithSync(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	syncData, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/sync.md")
	if err != nil {
		t.Fatalf("read sync.md: %v", err)
	}

	autoSyncData, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto-sync.md")
	if err != nil {
		t.Fatalf("read auto-sync.md: %v", err)
	}

	syncContent := string(syncData)
	autoSyncContent := string(autoSyncData)

	// Both must share Phase 0, Phase 1, Phase 2, Phase 3
	sharedPhases := []string{
		"### Phase 0:",
		"### Phase 1:",
		"### Phase 2:",
		"### Phase 3:",
	}

	for _, phase := range sharedPhases {
		if !strings.Contains(syncContent, phase) {
			t.Errorf("sync.md must contain %q", phase)
		}
		if !strings.Contains(autoSyncContent, phase) {
			t.Errorf("auto-sync.md must contain %q for structural parity with sync.md", phase)
		}
	}

	// sync.md must have Phase 4, auto-sync.md must NOT
	if !strings.Contains(syncContent, "### Phase 4") {
		t.Error("sync.md must contain '### Phase 4' (structural separation)")
	}
	if strings.Contains(autoSyncContent, "### Phase 4") {
		t.Error("auto-sync.md must NOT contain '### Phase 4' (structural separation)")
	}
}

// TestAutoSyncGracefulExitRetryCommand verifies auto-sync.md Graceful Exit
// section references "/ae auto-sync" (not "/ae sync") as the retry command.
// Prevents regression where users are directed to the interactive sync workflow
// which has merge capability. (AC-02 supplementary)
func TestAutoSyncGracefulExitRetryCommand(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/ae/workflows/auto-sync.md")
	if err != nil {
		t.Fatalf("read auto-sync.md: %v", err)
	}

	content := string(data)

	// Extract Graceful Exit section
	exitIdx := strings.Index(content, "## Graceful Exit")
	if exitIdx == -1 {
		t.Fatal("auto-sync.md must contain '## Graceful Exit' section")
	}

	exitSection := content[exitIdx:]
	if nextIdx := strings.Index(exitSection, "\n---"); nextIdx != -1 {
		exitSection = exitSection[:nextIdx]
	}

	if !strings.Contains(exitSection, "/ae auto-sync") {
		t.Error("auto-sync.md Graceful Exit must reference '/ae auto-sync' as retry command")
	}

	// Check each line for bare "/ae sync" (without "auto-" prefix).
	// Previous logic had a false-negative: it passed when both "/ae sync" and
	// "/ae auto-sync" coexisted. Line-by-line check prevents this.
	exitLines := strings.Split(exitSection, "\n")
	for i, line := range exitLines {
		if strings.Contains(line, "/ae auto-sync") {
			continue
		}
		if strings.Contains(line, "/ae sync") {
			t.Errorf("auto-sync.md Graceful Exit line %d must NOT use bare '/ae sync': %s",
				i+1, strings.TrimSpace(line))
		}
	}
}
