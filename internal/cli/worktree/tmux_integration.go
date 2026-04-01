package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AngeleyesTrue/ae-adk/internal/tmux"
)

// TmuxIntegration implements SPEC-WORKTREE-002 R5: Tmux Integration requirements.
// After worktree creation, it automatically creates a tmux session and injects the required environment variables.
//
// @MX:NOTE: SPEC-WORKTREE-002 R5 implementation - automatic tmux session creation and env var injection
// @MX:SPEC: SPEC-WORKTREE-002

// TmuxSessionConfig holds the configuration for creating a tmux session.
type TmuxSessionConfig struct {
	// ProjectName is the project name (e.g., "ae-adk-go")
	ProjectName string

	// SpecID is the SPEC identifier (e.g., "SPEC-WORKTREE-002")
	SpecID string

	// WorktreePath is the absolute path to the worktree
	WorktreePath string

	// ActiveMode is the current LLM mode (cc)
	ActiveMode string
}

// CreateTmuxSession creates a tmux session for the worktree.
//
// R5.1: Session name pattern: ae-{ProjectName}-{SPEC-ID}
// R5.4: After session creation, cd to worktree and execute /ae run command
//
// @MX:ANCHOR: Core entry point for the worktree-based development workflow
// @MX:REASON: tmux session automation is a core feature of SPEC-WORKTREE-002, called by multiple clients
// @MX:SPEC: SPEC-WORKTREE-002
func CreateTmuxSession(ctx context.Context, cfg *TmuxSessionConfig, tmuxMgr tmux.SessionManager) error {
	if cfg == nil {
		return fmt.Errorf("tmux session config is required")
	}

	if tmuxMgr == nil {
		return fmt.Errorf("tmux manager is required")
	}

	// R5.1: Generate session name
	sessionName := GenerateTmuxSessionName(cfg.ProjectName, cfg.SpecID)

	// R5.4: Create tmux session (detached mode)
	sessionCfg := &tmux.SessionConfig{
		Name:       sessionName,
		MaxVisible: 1, // Use a single pane
		Panes: []tmux.PaneConfig{
			{
				SpecID:  cfg.SpecID,
				Command: buildTmuxInitialCommand(cfg),
			},
		},
	}

	result, err := tmuxMgr.Create(ctx, sessionCfg)
	if err != nil {
		return fmt.Errorf("create tmux session: %w", err)
	}

	// Print log output
	fmt.Printf("Tmux session created: %s\n", result.SessionName)
	fmt.Printf("Panes created: %d\n", result.PaneCount)
	fmt.Printf("Attached: %v\n", result.Attached)
	fmt.Printf("Worktree path: %s\n", cfg.WorktreePath)
	fmt.Printf("To attach: tmux attach-session -t %s\n", sessionName)

	return nil
}

// buildTmuxInitialCommand builds the initial command to run in the tmux pane.
// R5.4: cd to worktree + execute /ae run
func buildTmuxInitialCommand(cfg *TmuxSessionConfig) string {
	// cd to the worktree path
	cdCmd := fmt.Sprintf("cd %s", cfg.WorktreePath)

	// Execute the /ae run command
	aeCmd := fmt.Sprintf("/ae run %s", cfg.SpecID)

	// Chain the two commands (separated by ;)
	return fmt.Sprintf("%s ; %s", cdCmd, aeCmd)
}

// IsTmuxAvailable checks whether tmux is available in the current environment.
// R1: Used for tmux availability detection in the Execution Mode Selection Gate.
//
// @MX:NOTE: SPEC-WORKTREE-002 R1 implementation - tmux availability detection
// @MX:SPEC: SPEC-WORKTREE-002
func IsTmuxAvailable() bool {
	// Check the $TMUX environment variable
	return os.Getenv("TMUX") != ""
}

// GetActiveMode reads the current active mode from .ae/config/sections/llm.yaml.
// R1.1: active mode detection.
//
// Returns: "cc" (default/empty)
//
// @MX:NOTE: SPEC-WORKTREE-002 R1.1 implementation - LLM mode detection
// @MX:SPEC: SPEC-WORKTREE-002
func GetActiveMode(projectRoot string) (string, error) {
	llmConfigPath := filepath.Join(projectRoot, ".ae", "config", "sections", "llm.yaml")

	data, err := os.ReadFile(llmConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "cc", nil
		}
		return "cc", fmt.Errorf("read llm.yaml: %w", err)
	}

	// Simple YAML parsing for llm.team_mode field
	// Look for any line containing "team_mode:" regardless of indentation
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "team_mode:") {
			// Extract value after "team_mode:"
			parts := strings.SplitN(trimmed, "team_mode:", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				// Remove quotes if present
				value = strings.Trim(value, "\"")
				value = strings.Trim(value, "'")
				if value == "" || value == "cc" {
					return "cc", nil
				}
				return value, nil
			}
		}
	}

	return "cc", nil
}

// BuildTmuxSessionConfig builds a tmux session configuration from worktree information.
//
// @MX:NOTE: SPEC-WORKTREE-002 integration function - tmux config builder
// @MX:SPEC: SPEC-WORKTREE-002
func BuildTmuxSessionConfig(projectName, specID, worktreePath, projectRoot string) (*TmuxSessionConfig, error) {
	activeMode, err := GetActiveMode(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("get active mode: %w", err)
	}

	cfg := &TmuxSessionConfig{
		ProjectName:  projectName,
		SpecID:       specID,
		WorktreePath: worktreePath,
		ActiveMode:   activeMode,
	}

	return cfg, nil
}
