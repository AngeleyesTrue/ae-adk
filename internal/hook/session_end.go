package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AngeleyesTrue/ae-adk/internal/hook/mx"
)

// teamConfig is the minimal structure read from ~/.claude/teams/*/config.json.
type teamConfig struct {
	LeadSessionID string `json:"leadSessionId"`
}

// sessionEndHandler processes SessionEnd events.
// It persists session metrics, cleans up temporary resources, and optionally
// cleans up temporary resources (REQ-HOOK-034). Always returns "allow".
type sessionEndHandler struct{}

// NewSessionEndHandler creates a new SessionEnd event handler.
func NewSessionEndHandler() Handler {
	return &sessionEndHandler{}
}

// EventType returns EventSessionEnd.
func (h *sessionEndHandler) EventType() EventType {
	return EventSessionEnd
}

// Handle processes a SessionEnd event. It logs the session completion,
// performs best-effort team directory cleanup, garbage-collects stale teams,
// clears tmux session env vars, and kills orphaned tmux sessions.
// SessionEnd hooks should not use hookSpecificOutput per Claude Code protocol.
// All cleanup is best-effort: errors are logged with slog.Warn, never returned.
func (h *sessionEndHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session ending",
		"session_id", input.SessionID,
		"project_dir", input.ProjectDir,
	)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("session_end: could not determine home directory",
			"error", err,
		)
		return &HookOutput{}, nil
	}

	cleanupCurrentSessionTeam(input.SessionID, homeDir)
	garbageCollectStaleTeams(homeDir)
	garbageCollectOrphanedTasks(homeDir)
	cleanupOrphanedTmuxSessions(ctx)

	projectDir := input.CWD
	if projectDir == "" {
		projectDir = input.ProjectDir // Fallback for legacy
	}
	if projectDir != "" {
		cleanupBogusRootDir(projectDir)
	}

	// Validate MX tags for files modified during this session.
	// Best-effort: errors are logged, never block session end (AC-SESSION-002).
	if projectDir != "" {
		mxCtx, mxCancel := context.WithTimeout(ctx, 4*time.Second)
		defer mxCancel()
		modifiedFiles := getModifiedGoFiles(mxCtx, projectDir)
		if len(modifiedFiles) > 0 {
			validateMxTags(mxCtx, modifiedFiles, projectDir)
		}
	}

	slog.Info("session_end: cleanup complete",
		"session_id", input.SessionID,
	)

	// SessionEnd hooks return empty JSON {} per Claude Code protocol
	// Do NOT use hookSpecificOutput for SessionEnd events
	return &HookOutput{}, nil
}

// getModifiedGoFiles returns the list of .go files modified in the current session.
// Uses `git diff --name-only HEAD` to find modified files.
// Returns an empty slice if git is unavailable or no Go files were modified.
func getModifiedGoFiles(ctx context.Context, projectDir string) []string {
	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", "HEAD")
	cmd.Dir = projectDir
	out, err := cmd.Output()
	if err != nil {
		// git diff may fail in non-git environments; this is expected
		return nil
	}

	var goFiles []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ".go") {
			absPath := filepath.Join(projectDir, line)
			goFiles = append(goFiles, absPath)
		}
	}
	return goFiles
}

// validateMxTags validates @MX tag presence for the given Go files.
// Observation-only: errors are logged with slog.Warn, never returned.
// AC-SESSION-001: results are logged with slog.Info.
// AC-SESSION-003: respects 4s timeout via context (partial results on timeout).
func validateMxTags(ctx context.Context, filePaths []string, projectRoot string) {
	if len(filePaths) == 0 {
		return
	}

	// Filter to Go files only
	var goFiles []string
	for _, f := range filePaths {
		if strings.HasSuffix(f, ".go") {
			goFiles = append(goFiles, f)
		}
	}
	if len(goFiles) == 0 {
		return
	}

	validator := mx.NewValidator(nil, projectRoot)
	report, err := validator.ValidateFiles(ctx, goFiles)
	if err != nil {
		slog.Warn("session_end: mx validation error",
			"error", err,
			"files", len(goFiles),
		)
		return
	}

	if report == nil {
		return
	}

	// AC-SESSION-003: log timed out files
	if len(report.TimedOutFiles) > 0 {
		slog.Warn("session_end: mx validation timed out for some files",
			"timed_out", len(report.TimedOutFiles),
		)
	}

	// AC-SESSION-001: log validation results
	if report.TotalViolations() > 0 || len(report.TimedOutFiles) > 0 {
		slog.Info("session_end: mx validation complete",
			"files_validated", len(report.FileReports),
			"files_timed_out", len(report.TimedOutFiles),
			"p1_violations", report.P1Count(),
			"p2_violations", report.P2Count(),
			"p3_violations", report.P3Count(),
			"p4_violations", report.P4Count(),
			"duration_ms", report.Duration.Milliseconds(),
		)

		if report.HasBlockingViolations() {
			slog.Warn("session_end: blocking MX violations detected - run /ae run to add missing tags",
				"p1", report.P1Count(),
				"p2", report.P2Count(),
			)
		}
	} else {
		slog.Info("session_end: mx validation passed",
			"files_validated", len(report.FileReports),
			"duration_ms", report.Duration.Milliseconds(),
		)
	}
}

// cleanupCurrentSessionTeam removes the team directory whose leadSessionId
// matches the given sessionID. Errors are logged and never returned.
func cleanupCurrentSessionTeam(sessionID, homeDir string) {
	teamsDir := filepath.Join(homeDir, ".claude", "teams")

	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read teams directory",
				"path", teamsDir,
				"error", err,
			)
		}
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		teamDir := filepath.Join(teamsDir, entry.Name())
		configPath := filepath.Join(teamDir, "config.json")

		data, err := os.ReadFile(configPath)
		if err != nil {
			// Missing config.json is normal; skip silently.
			continue
		}

		var cfg teamConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			slog.Warn("session_end: could not parse team config",
				"path", configPath,
				"error", err,
			)
			continue
		}

		if cfg.LeadSessionID == sessionID {
			if err := os.RemoveAll(teamDir); err != nil {
				slog.Warn("session_end: could not remove team directory",
					"path", teamDir,
					"error", err,
				)
			} else {
				slog.Info("session_end: removed team directory for session",
					"team_dir", teamDir,
					"session_id", sessionID,
				)
				// Also remove the corresponding task directory when the team directory is successfully deleted
				tasksDir := filepath.Join(homeDir, ".claude", "tasks", entry.Name())
				if err := os.RemoveAll(tasksDir); err != nil {
					slog.Warn("session_end: could not remove task directory for session",
						"path", tasksDir,
						"error", err,
					)
				} else {
					slog.Info("session_end: removed task directory for session",
						"task_dir", tasksDir,
						"session_id", sessionID,
					)
				}
			}
		}
	}
}

// garbageCollectStaleTeams removes team directories that have not been
// modified in more than 24 hours. This catches teams left behind by
// interrupted sessions. Errors are logged and never returned.
func garbageCollectStaleTeams(homeDir string) {
	const staleDuration = 24 * time.Hour

	teamsDir := filepath.Join(homeDir, ".claude", "teams")

	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read teams directory for GC",
				"path", teamsDir,
				"error", err,
			)
		}
		return
	}

	cutoff := time.Now().Add(-staleDuration)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			slog.Warn("session_end: could not stat team directory",
				"name", entry.Name(),
				"error", err,
			)
			continue
		}

		if info.ModTime().Before(cutoff) {
			teamDir := filepath.Join(teamsDir, entry.Name())
			if err := os.RemoveAll(teamDir); err != nil {
				slog.Warn("session_end: could not remove stale team directory",
					"path", teamDir,
					"error", err,
				)
			} else {
				slog.Info("session_end: removed stale team directory",
					"path", teamDir,
					"age", time.Since(info.ModTime()).Round(time.Minute),
				)
				// Also remove the corresponding task directory when a stale team directory is successfully deleted
				taskDir := filepath.Join(homeDir, ".claude", "tasks", entry.Name())
				if err := os.RemoveAll(taskDir); err != nil {
					slog.Warn("session_end: could not remove stale task directory",
						"path", taskDir,
						"error", err,
					)
				} else {
					slog.Info("session_end: removed stale task directory",
						"path", taskDir,
					)
				}
			}
		}
	}
}

// garbageCollectOrphanedTasks cleans up orphaned task directories under ~/.claude/tasks/
// that have no corresponding team directory. Collects task directories left behind by
// interrupted sessions or incomplete cleanup. Errors are logged and never returned.
func garbageCollectOrphanedTasks(homeDir string) {
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	teamsDir := filepath.Join(homeDir, ".claude", "teams")

	taskEntries, err := os.ReadDir(tasksDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read tasks directory for orphan GC",
				"path", tasksDir,
				"error", err,
			)
		}
		return
	}

	for _, entry := range taskEntries {
		if !entry.IsDir() {
			continue
		}

		// Check whether the corresponding team directory exists
		teamDir := filepath.Join(teamsDir, entry.Name())
		if _, err := os.Stat(teamDir); err == nil {
			// Team directory exists, so this is not an orphan — keep it
			continue
		}

		// No team directory, so remove the orphaned task directory
		taskDir := filepath.Join(tasksDir, entry.Name())
		if err := os.RemoveAll(taskDir); err != nil {
			slog.Warn("session_end: could not remove orphaned task directory",
				"path", taskDir,
				"error", err,
			)
		} else {
			slog.Info("session_end: removed orphaned task directory",
				"path", taskDir,
			)
		}
	}
}

// getCurrentTmuxSession returns the name of the current tmux session.
// Returns empty string if not in tmux or if detection fails.
func getCurrentTmuxSession(ctx context.Context) string {
	// Check if we're in tmux
	if os.Getenv("TMUX") == "" {
		return ""
	}

	// Use tmux display-message to get current session name.
	cmd := exec.CommandContext(ctx, "tmux", "display-message", "-p", "#S")
	out, err := cmd.Output()
	if err != nil {
		slog.Warn("session_end: could not get current tmux session",
			"error", err,
		)
		return ""
	}

	return strings.TrimSpace(string(out))
}

// aeTmuxSessionPrefix is the naming convention for tmux sessions created by
// AE Agent Teams. Only sessions matching this prefix are eligible for cleanup.
const aeTmuxSessionPrefix = "ae-"

// cleanupOrphanedTmuxSessions kills detached tmux sessions created by AE
// Agent Teams (prefix "ae-"). User-created sessions are never touched.
// The cleanup is capped at 4 seconds to stay within the SessionEnd hook
// timeout budget. If tmux is not installed or no sessions exist, the function
// returns silently.
func cleanupOrphanedTmuxSessions(ctx context.Context) {
	// Reserve 4 seconds for tmux cleanup, leaving 1 second buffer.
	cleanupCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// Get current tmux session name to protect it from being killed.
	currentSession := getCurrentTmuxSession(cleanupCtx)

	// List all tmux sessions.
	listCmd := exec.CommandContext(cleanupCtx, "tmux", "list-sessions")
	out, err := listCmd.Output()
	if err != nil {
		if cleanupCtx.Err() != nil {
			slog.Warn("session_end: tmux cleanup timed out",
				"timeout", 4*time.Second,
			)
		}
		return
	}

	lines := strings.SplitSeq(strings.TrimSpace(string(out)), "\n")
	for line := range lines {
		if line == "" {
			continue
		}
		// Skip the current tmux session - never kill the user's actual session.
		name, _, found := strings.Cut(line, ":")
		if !found || name == "" {
			continue
		}
		if name == currentSession {
			continue
		}

		// Only kill sessions created by AE (prefixed with "ae-").
		// Never kill user-created tmux sessions.
		if !strings.HasPrefix(name, aeTmuxSessionPrefix) {
			continue
		}

		// Sessions currently attached contain "(attached)".
		if strings.Contains(line, "(attached)") {
			continue
		}

		killCmd := exec.CommandContext(cleanupCtx, "tmux", "kill-session", "-t", name)
		if err := killCmd.Run(); err != nil {
			slog.Warn("session_end: could not kill orphaned tmux session",
				"session", name,
				"error", err,
			)
		} else {
			slog.Info("session_end: killed orphaned tmux session",
				"session", name,
			)
		}
	}
}

// cleanupBogusRootDir removes a literal "{}" directory from the project root
// if it exists. This directory is a side-effect of a Claude Code bug where the
// {project_root} template variable used for agent memory paths (memory: project)
// is not substituted when spawning agents inside git worktrees, resulting in a
// directory named "{}" at the worktree root.
//
// The cleanup is best-effort: errors are logged with slog.Warn and never returned.
func cleanupBogusRootDir(projectDir string) {
	bogusDir := filepath.Join(projectDir, "{}")
	info, err := os.Lstat(bogusDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not stat bogus {} directory",
				"path", bogusDir,
				"error", err,
			)
		}
		return
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return
	}
	if !info.IsDir() {
		return
	}
	if err := os.RemoveAll(bogusDir); err != nil {
		slog.Warn("session_end: could not remove bogus {} directory",
			"path", bogusDir,
			"error", err,
		)
		return
	}
	slog.Info("session_end: removed bogus {} directory caused by unresolved agent memory path",
		"path", bogusDir,
	)
}
