package template

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// humanGateHeaderPattern matches H2 (##) or H3 (###) Markdown headers whose
// title contains the literal phrase "HUMAN GATE" (case-insensitive, with
// optional whitespace, hyphen, or underscore separator). Compiled once.
//
// Examples that MATCH (these are headers):
//   - "## HUMAN GATE: Approval Checkpoint"
//   - "### Human Gate — Plan Approval"
//   - "## human-gate confirmation"
//   - "### HUMAN_GATE Final Confirmation"
//
// Examples that DO NOT MATCH (these are not H2/H3 headers):
//   - "Do not add HUMAN GATE blocks to auto.md" (prose, not a header)
//   - "# HUMAN GATE" (H1, out of scope)
//   - "#### HUMAN GATE" (H4, out of scope)
//   - "the HUMAN GATE pattern" (inline mention)
//
// SPEC-UPDATE-004 REQ-26 / AC-26 — case-insensitive flag and multi-line anchor.
var humanGateHeaderPattern = regexp.MustCompile(`(?im)^#{2,3}\s+.*human[\s\-_]?gate`)

// humanGateFile describes a single workflow file's expected HUMAN GATE state.
type humanGateFile struct {
	// path is the file path within the embedded template FS, rooted at the
	// embedded FS root (no leading slash, no internal/template/templates prefix).
	path string

	// minHeaders is the minimum number of HUMAN GATE H2/H3 headers expected.
	// For interactive workflow files, this is 2. For auto pipeline files, this
	// is 0 and any presence is a critical SPEC-PIPELINE-002 cascade-prevention
	// violation.
	minHeaders int

	// maxHeaders is the maximum number of HUMAN GATE H2/H3 headers allowed.
	// Setting to 0 enforces strict absence (auto pipeline files).
	maxHeaders int

	// description is the human-readable purpose, used in failure messages.
	description string
}

// readWorkflowFile loads a workflow file from the embedded template FS.
func readWorkflowFile(t *testing.T, path string) string {
	t.Helper()
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// findHumanGateHeaderLines returns the 1-indexed line numbers and contents of
// every HUMAN GATE H2/H3 header occurrence in the given file content. Used for
// detailed diagnostic messages on failure.
func findHumanGateHeaderLines(content string) []string {
	var hits []string
	for i, line := range strings.Split(content, "\n") {
		if humanGateHeaderPattern.MatchString(line) {
			hits = append(hits, fmt.Sprintf("L%d: %s", i+1, line))
		}
	}
	return hits
}

// TestHumanGateAbsentAuto asserts that auto.md does NOT contain any HUMAN GATE
// H2/H3 section header. Adding such a header would break SPEC-PIPELINE-002
// cascade prevention by introducing an unbounded user-input wait inside an
// otherwise non-interactive auto pipeline.
//
// SPEC-UPDATE-004 REQ-26 (HARD), AC-26 Scenario 1.
func TestHumanGateAbsentAuto(t *testing.T) {
	t.Parallel()

	const path = ".claude/skills/ae/workflows/auto.md"
	content := readWorkflowFile(t, path)

	hits := findHumanGateHeaderLines(content)
	if len(hits) > 0 {
		t.Errorf(
			"%s contains %d HUMAN GATE H2/H3 section header(s), want 0\n"+
				"Headers found:\n  %s\n"+
				"Adding HUMAN GATE as an H2/H3 section to auto.md violates SPEC-PIPELINE-002 "+
				"cascade prevention (Phase 4 removed, merge capability absent, AskUserQuestion "+
				"merge gate). Plain prose mentions of \"HUMAN GATE\" outside of H2/H3 headers "+
				"are permitted (e.g., a warning telling readers not to add the header).",
			path, len(hits), strings.Join(hits, "\n  "),
		)
	}
}

// TestHumanGateAbsentAutoSync mirrors TestHumanGateAbsentAuto for auto-sync.md.
//
// SPEC-UPDATE-004 REQ-26 (HARD), AC-26 Scenario 2.
func TestHumanGateAbsentAutoSync(t *testing.T) {
	t.Parallel()

	const path = ".claude/skills/ae/workflows/auto-sync.md"
	content := readWorkflowFile(t, path)

	hits := findHumanGateHeaderLines(content)
	if len(hits) > 0 {
		t.Errorf(
			"%s contains %d HUMAN GATE H2/H3 section header(s), want 0\n"+
				"Headers found:\n  %s\n"+
				"Adding HUMAN GATE as an H2/H3 section to auto-sync.md violates "+
				"SPEC-PIPELINE-002 cascade prevention. Plain prose mentions are permitted.",
			path, len(hits), strings.Join(hits, "\n  "),
		)
	}
}

// TestHumanGatePresentInteractive asserts that the three interactive workflow
// skill files (plan.md, run.md, sync.md) each contain at least 2 HUMAN GATE
// H2/H3 section headers. These gates expose explicit user-confirmation points
// that the orchestrator is expected to honor before destructive or
// commitment-grade actions.
//
// SPEC-UPDATE-004 REQ-26, AC-26 Scenarios 3-5.
func TestHumanGatePresentInteractive(t *testing.T) {
	t.Parallel()

	files := []humanGateFile{
		{
			path:        ".claude/skills/ae/workflows/plan.md",
			minHeaders:  2,
			maxHeaders:  -1, // unlimited upper bound
			description: "interactive plan workflow (SPEC creation)",
		},
		{
			path:        ".claude/skills/ae/workflows/run.md",
			minHeaders:  2,
			maxHeaders:  -1,
			description: "interactive run workflow (DDD/TDD implementation)",
		},
		{
			path:        ".claude/skills/ae/workflows/sync.md",
			minHeaders:  2,
			maxHeaders:  -1,
			description: "interactive sync workflow (docs + PR)",
		},
	}

	for _, f := range files {
		f := f
		t.Run(f.path, func(t *testing.T) {
			t.Parallel()

			content := readWorkflowFile(t, f.path)
			hits := findHumanGateHeaderLines(content)

			if len(hits) < f.minHeaders {
				t.Errorf(
					"%s (%s) has %d HUMAN GATE H2/H3 header(s), want >= %d\n"+
						"Headers found:\n  %s\n"+
						"Each interactive workflow MUST expose at least %d explicit gate "+
						"sections so the orchestrator surfaces user-confirmation prompts via "+
						"AskUserQuestion before destructive actions.",
					f.path, f.description, len(hits), f.minHeaders,
					strings.Join(hits, "\n  "),
					f.minHeaders,
				)
			}
		})
	}
}

// TestHumanGateProsePermittedInAuto guards against an over-aggressive
// regression: it confirms that a documentation-style mention of "HUMAN GATE"
// inside auto.md (NOT as an H2/H3 header) is correctly NOT flagged. This is a
// fixture-based positive test for the regex's specificity. AC-26 Scenario 7.
func TestHumanGateProsePermittedInAuto(t *testing.T) {
	t.Parallel()

	// Synthetic fixture: prose mention of HUMAN GATE that should NOT match.
	fixtures := []struct {
		name string
		text string
		want int
	}{
		{
			name: "prose-warning-paragraph",
			text: "Note: do NOT add HUMAN GATE blocks to auto.md. " +
				"SPEC-PIPELINE-002 forbids this.",
			want: 0,
		},
		{
			name: "inline-code-mention",
			text: "The string `HUMAN GATE` is reserved for interactive workflows.",
			want: 0,
		},
		{
			name: "h4-out-of-scope",
			text: "#### HUMAN GATE (this H4 is intentionally permitted)",
			want: 0,
		},
		{
			name: "h2-header-matches",
			text: "## HUMAN GATE: Approval Checkpoint",
			want: 1,
		},
		{
			name: "h3-with-hyphen",
			text: "### HUMAN-GATE Final Confirmation",
			want: 1,
		},
		{
			name: "h2-mixed-case",
			text: "## Human Gate — plan approval",
			want: 1,
		},
		{
			name: "h2-with-underscore",
			text: "## HUMAN_GATE Final Confirmation",
			want: 1,
		},
	}

	for _, fx := range fixtures {
		fx := fx
		t.Run(fx.name, func(t *testing.T) {
			t.Parallel()
			got := len(findHumanGateHeaderLines(fx.text))
			if got != fx.want {
				t.Errorf("regex matched %d header(s) in fixture %q, want %d (input: %q)",
					got, fx.name, fx.want, fx.text)
			}
		})
	}
}
