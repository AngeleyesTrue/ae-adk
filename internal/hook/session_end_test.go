package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionEndHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewSessionEndHandler()

	if got := h.EventType(); got != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", got, EventSessionEnd)
	}
}

func TestSessionEndHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *HookInput
		setupDir bool
	}{
		{
			name: "normal session end",
			input: &HookInput{
				SessionID:     "sess-end-1",
				CWD:           "", // will be set in test
				HookEventName: "SessionEnd",
			},
			setupDir: true,
		},
		{
			name: "session end without project dir",
			input: &HookInput{
				SessionID:     "sess-end-2",
				CWD:           "/tmp",
				HookEventName: "SessionEnd",
			},
			setupDir: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.setupDir {
				tmpDir := t.TempDir()
				stateDir := filepath.Join(tmpDir, ".ae", "state")
				if err := os.MkdirAll(stateDir, 0o755); err != nil {
					t.Fatalf("setup state dir: %v", err)
				}
				tt.input.CWD = tmpDir
				tt.input.ProjectDir = tmpDir
			}

			h := NewSessionEndHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// SessionEnd hooks return empty JSON {} per Claude Code protocol
			// They should NOT have hookSpecificOutput set
			if got.HookSpecificOutput != nil {
				t.Error("SessionEnd hook should not set hookSpecificOutput")
			}
		})
	}
}

func TestCleanupCurrentSessionTeam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sessionID string
		teams     map[string]string // teamName -> leadSessionId
		wantGone  []string          // team dirs that should be removed
		wantKept  []string          // team dirs that should remain
	}{
		{
			name:      "removes matching session team",
			sessionID: "sess-abc-123",
			teams: map[string]string{
				"my-team":    "sess-abc-123",
				"other-team": "sess-xyz-789",
			},
			wantGone: []string{"my-team"},
			wantKept: []string{"other-team"},
		},
		{
			name:      "no match leaves all teams",
			sessionID: "sess-no-match",
			teams: map[string]string{
				"team-a": "sess-111",
				"team-b": "sess-222",
			},
			wantGone: nil,
			wantKept: []string{"team-a", "team-b"},
		},
		{
			name:      "empty teams dir",
			sessionID: "sess-empty",
			teams:     map[string]string{},
			wantGone:  nil,
			wantKept:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			homeDir := t.TempDir()
			teamsDir := filepath.Join(homeDir, ".claude", "teams")
			if err := os.MkdirAll(teamsDir, 0o755); err != nil {
				t.Fatalf("setup teams dir: %v", err)
			}

			// Create team directories with config.json
			for name, leadSessionID := range tt.teams {
				teamDir := filepath.Join(teamsDir, name)
				if err := os.MkdirAll(teamDir, 0o755); err != nil {
					t.Fatalf("create team dir %s: %v", name, err)
				}
				cfg := teamConfig{LeadSessionID: leadSessionID}
				data, err := json.Marshal(cfg)
				if err != nil {
					t.Fatalf("marshal config for %s: %v", name, err)
				}
				if err := os.WriteFile(filepath.Join(teamDir, "config.json"), data, 0o644); err != nil {
					t.Fatalf("write config for %s: %v", name, err)
				}
			}

			cleanupCurrentSessionTeam(tt.sessionID, homeDir)

			for _, name := range tt.wantGone {
				if _, err := os.Stat(filepath.Join(teamsDir, name)); !os.IsNotExist(err) {
					t.Errorf("team dir %q should have been removed", name)
				}
			}
			for _, name := range tt.wantKept {
				if _, err := os.Stat(filepath.Join(teamsDir, name)); os.IsNotExist(err) {
					t.Errorf("team dir %q should still exist", name)
				}
			}
		})
	}
}

func TestCleanupCurrentSessionTeam_MissingTeamsDir(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	// Don't create .claude/teams/ — should not panic or error
	cleanupCurrentSessionTeam("any-session", homeDir)
}

func TestCleanupCurrentSessionTeam_InvalidConfigJSON(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	teamDir := filepath.Join(teamsDir, "bad-config")
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(teamDir, "config.json"), []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	// Should not panic; directory should remain (bad config is not cleaned up)
	cleanupCurrentSessionTeam("any-session", homeDir)

	if _, err := os.Stat(teamDir); os.IsNotExist(err) {
		t.Error("team dir with invalid config should not be removed")
	}
}

func TestGarbageCollectStaleTeams(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("setup teams dir: %v", err)
	}

	// Create a stale team dir (modtime > 24h ago)
	staleDir := filepath.Join(teamsDir, "stale-team")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatalf("create stale dir: %v", err)
	}
	staleTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(staleDir, staleTime, staleTime); err != nil {
		t.Fatalf("set stale time: %v", err)
	}

	// Create a fresh team dir (modtime < 24h)
	freshDir := filepath.Join(teamsDir, "fresh-team")
	if err := os.MkdirAll(freshDir, 0o755); err != nil {
		t.Fatalf("create fresh dir: %v", err)
	}

	garbageCollectStaleTeams(homeDir)

	// Stale should be gone
	if _, err := os.Stat(staleDir); !os.IsNotExist(err) {
		t.Error("stale team dir should have been removed")
	}

	// Fresh should remain
	if _, err := os.Stat(freshDir); os.IsNotExist(err) {
		t.Error("fresh team dir should still exist")
	}
}

func TestGarbageCollectStaleTeams_MissingDir(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	// Don't create .claude/teams/ — should not panic
	garbageCollectStaleTeams(homeDir)
}

func TestCleanupOrphanedTmuxSessions_GracefulWithContext(t *testing.T) {
	t.Parallel()

	// With a cancelled context, the function should return without panic.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Should not panic or hang.
	cleanupOrphanedTmuxSessions(ctx)
}

// TestAETmuxSessionPrefix verifies the naming convention constant used to
// filter tmux sessions during cleanup. Only sessions with this prefix are
// eligible for orphan cleanup — user-created sessions are never touched.
func TestAETmuxSessionPrefix(t *testing.T) {
	t.Parallel()

	if aeTmuxSessionPrefix == "" {
		t.Fatal("aeTmuxSessionPrefix must not be empty")
	}
	if aeTmuxSessionPrefix != "ae-" {
		t.Errorf("aeTmuxSessionPrefix = %q, want %q", aeTmuxSessionPrefix, "ae-")
	}
}

func TestSessionEndHandler_AlwaysReturnsEmptyOutput(t *testing.T) {
	t.Parallel()

	h := NewSessionEndHandler()
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "test-always-empty",
		CWD:           t.TempDir(),
		HookEventName: "SessionEnd",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("output should never be nil")
	}
	if got.Decision != "" {
		t.Errorf("Decision should be empty, got %q", got.Decision)
	}
	if got.ExitCode != 0 {
		t.Errorf("ExitCode should be 0, got %d", got.ExitCode)
	}
}

// TestCleanupCurrentSessionTeam_AlsoRemovesTaskDir verifies that the corresponding
// tasks directory is also removed when a session team directory is deleted.
func TestCleanupCurrentSessionTeam_AlsoRemovesTaskDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sessionID    string
		teams        map[string]string // teamName -> leadSessionId
		wantTeamGone []string          // team directories that should be removed
		wantTeamKept []string          // team directories that should be kept
		wantTaskGone []string          // task directories that should be removed
		wantTaskKept []string          // task directories that should be kept
	}{
		{
			name:      "remove both team/task directories for matching session",
			sessionID: "sess-abc-123",
			teams: map[string]string{
				"my-team":    "sess-abc-123",
				"other-team": "sess-xyz-789",
			},
			wantTeamGone: []string{"my-team"},
			wantTeamKept: []string{"other-team"},
			wantTaskGone: []string{"my-team"},
			wantTaskKept: []string{"other-team"},
		},
		{
			name:      "keep all team/task directories when no match",
			sessionID: "sess-no-match",
			teams: map[string]string{
				"team-a": "sess-111",
				"team-b": "sess-222",
			},
			wantTeamGone: nil,
			wantTeamKept: []string{"team-a", "team-b"},
			wantTaskGone: nil,
			wantTaskKept: []string{"team-a", "team-b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			homeDir := t.TempDir()
			teamsDir := filepath.Join(homeDir, ".claude", "teams")
			tasksDir := filepath.Join(homeDir, ".claude", "tasks")
			if err := os.MkdirAll(teamsDir, 0o755); err != nil {
				t.Fatalf("failed to create teams directory: %v", err)
			}
			if err := os.MkdirAll(tasksDir, 0o755); err != nil {
				t.Fatalf("failed to create tasks directory: %v", err)
			}

			// Create team directories and their corresponding task directories
			for name, leadSessionID := range tt.teams {
				teamDir := filepath.Join(teamsDir, name)
				if err := os.MkdirAll(teamDir, 0o755); err != nil {
					t.Fatalf("failed to create team directory %s: %v", name, err)
				}
				cfg := teamConfig{LeadSessionID: leadSessionID}
				data, err := json.Marshal(cfg)
				if err != nil {
					t.Fatalf("failed to marshal config for %s: %v", name, err)
				}
				if err := os.WriteFile(filepath.Join(teamDir, "config.json"), data, 0o644); err != nil {
					t.Fatalf("failed to write config file for %s: %v", name, err)
				}

				// Also create the corresponding task directory
				taskDir := filepath.Join(tasksDir, name)
				if err := os.MkdirAll(taskDir, 0o755); err != nil {
					t.Fatalf("failed to create task directory %s: %v", name, err)
				}
			}

			cleanupCurrentSessionTeam(tt.sessionID, homeDir)

			// Verify team directory removal
			for _, name := range tt.wantTeamGone {
				if _, err := os.Stat(filepath.Join(teamsDir, name)); !os.IsNotExist(err) {
					t.Errorf("team directory %q should have been removed", name)
				}
			}
			for _, name := range tt.wantTeamKept {
				if _, err := os.Stat(filepath.Join(teamsDir, name)); os.IsNotExist(err) {
					t.Errorf("team directory %q should have been kept", name)
				}
			}

			// Verify task directory removal
			for _, name := range tt.wantTaskGone {
				if _, err := os.Stat(filepath.Join(tasksDir, name)); !os.IsNotExist(err) {
					t.Errorf("task directory %q should have been removed", name)
				}
			}
			for _, name := range tt.wantTaskKept {
				if _, err := os.Stat(filepath.Join(tasksDir, name)); os.IsNotExist(err) {
					t.Errorf("task directory %q should have been kept", name)
				}
			}
		})
	}
}

// TestGarbageCollectStaleTeams_AlsoRemovesTaskDir verifies that the corresponding
// stale tasks directory is also removed during stale team directory GC.
func TestGarbageCollectStaleTeams_AlsoRemovesTaskDir(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	teamsDir := filepath.Join(homeDir, ".claude", "teams")
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	if err := os.MkdirAll(teamsDir, 0o755); err != nil {
		t.Fatalf("failed to create teams directory: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks directory: %v", err)
	}

	// Create stale team/task directories (25 hours ago)
	staleTeamDir := filepath.Join(teamsDir, "stale-team")
	if err := os.MkdirAll(staleTeamDir, 0o755); err != nil {
		t.Fatalf("failed to create stale team directory: %v", err)
	}
	staleTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(staleTeamDir, staleTime, staleTime); err != nil {
		t.Fatalf("failed to set stale time: %v", err)
	}

	staleTaskDir := filepath.Join(tasksDir, "stale-team")
	if err := os.MkdirAll(staleTaskDir, 0o755); err != nil {
		t.Fatalf("failed to create stale task directory: %v", err)
	}

	// Create fresh team/task directories
	freshTeamDir := filepath.Join(teamsDir, "fresh-team")
	if err := os.MkdirAll(freshTeamDir, 0o755); err != nil {
		t.Fatalf("failed to create fresh team directory: %v", err)
	}

	freshTaskDir := filepath.Join(tasksDir, "fresh-team")
	if err := os.MkdirAll(freshTaskDir, 0o755); err != nil {
		t.Fatalf("failed to create fresh task directory: %v", err)
	}

	garbageCollectStaleTeams(homeDir)

	// Verify stale team directory removal
	if _, err := os.Stat(staleTeamDir); !os.IsNotExist(err) {
		t.Error("stale team directory should have been removed")
	}

	// Verify stale task directory removal
	if _, err := os.Stat(staleTaskDir); !os.IsNotExist(err) {
		t.Error("stale task directory should have been removed")
	}

	// Verify fresh team directory is kept
	if _, err := os.Stat(freshTeamDir); os.IsNotExist(err) {
		t.Error("fresh team directory should have been kept")
	}

	// Verify fresh task directory is kept
	if _, err := os.Stat(freshTaskDir); os.IsNotExist(err) {
		t.Error("fresh task directory should have been kept")
	}
}

// TestGarbageCollectOrphanedTasks verifies that orphaned task directories
// left without a corresponding team directory are cleaned up.
func TestGarbageCollectOrphanedTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		teamNames    []string // directories present in ~/.claude/teams/
		taskNames    []string // directories present in ~/.claude/tasks/
		wantTaskGone []string // task directories that should be removed (orphans)
		wantTaskKept []string // task directories that should be kept (have matching team)
	}{
		{
			name:         "remove orphaned task directories with no matching team",
			teamNames:    []string{"team-a"},
			taskNames:    []string{"team-a", "orphan-1", "orphan-2"},
			wantTaskGone: []string{"orphan-1", "orphan-2"},
			wantTaskKept: []string{"team-a"},
		},
		{
			name:         "nothing removed when all tasks have matching teams",
			teamNames:    []string{"team-x", "team-y"},
			taskNames:    []string{"team-x", "team-y"},
			wantTaskGone: nil,
			wantTaskKept: []string{"team-x", "team-y"},
		},
		{
			name:         "remove all tasks when no teams exist",
			teamNames:    []string{},
			taskNames:    []string{"orphan-a", "orphan-b"},
			wantTaskGone: []string{"orphan-a", "orphan-b"},
			wantTaskKept: nil,
		},
		{
			name:         "no action when only teams exist and tasks are empty",
			teamNames:    []string{"team-z"},
			taskNames:    []string{},
			wantTaskGone: nil,
			wantTaskKept: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			homeDir := t.TempDir()
			teamsDir := filepath.Join(homeDir, ".claude", "teams")
			tasksDir := filepath.Join(homeDir, ".claude", "tasks")
			if err := os.MkdirAll(teamsDir, 0o755); err != nil {
				t.Fatalf("failed to create teams directory: %v", err)
			}
			if err := os.MkdirAll(tasksDir, 0o755); err != nil {
				t.Fatalf("failed to create tasks directory: %v", err)
			}

			// Create team directories
			for _, name := range tt.teamNames {
				teamDir := filepath.Join(teamsDir, name)
				if err := os.MkdirAll(teamDir, 0o755); err != nil {
					t.Fatalf("failed to create team directory %s: %v", name, err)
				}
			}

			// Create task directories
			for _, name := range tt.taskNames {
				taskDir := filepath.Join(tasksDir, name)
				if err := os.MkdirAll(taskDir, 0o755); err != nil {
					t.Fatalf("failed to create task directory %s: %v", name, err)
				}
			}

			garbageCollectOrphanedTasks(homeDir)

			// Verify removal
			for _, name := range tt.wantTaskGone {
				if _, err := os.Stat(filepath.Join(tasksDir, name)); !os.IsNotExist(err) {
					t.Errorf("orphaned task directory %q should have been removed", name)
				}
			}

			// Verify preservation
			for _, name := range tt.wantTaskKept {
				if _, err := os.Stat(filepath.Join(tasksDir, name)); os.IsNotExist(err) {
					t.Errorf("task directory %q should have been kept", name)
				}
			}
		})
	}
}

// TestGarbageCollectOrphanedTasks_MissingDir verifies that no panic or error
// occurs when the ~/.claude/tasks/ directory does not exist.
func TestGarbageCollectOrphanedTasks_MissingDir(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	// Do not create ~/.claude/tasks/ directory — must not panic or return error
	garbageCollectOrphanedTasks(homeDir)
}

// TestCleanupBogusRootDir_RemovesDirectory verifies that the bogus "{}"
// directory is removed when it exists in the project root.
func TestCleanupBogusRootDir_RemovesDirectory(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	bogusDir := filepath.Join(projectDir, "{}")
	if err := os.MkdirAll(bogusDir, 0o755); err != nil {
		t.Fatalf("setup: create bogus dir: %v", err)
	}
	// Add a nested file to ensure RemoveAll is used, not Remove.
	nestedFile := filepath.Join(bogusDir, ".claude", "agent-memory", "expert-backend", "memory.md")
	if err := os.MkdirAll(filepath.Dir(nestedFile), 0o755); err != nil {
		t.Fatalf("setup: create nested dirs: %v", err)
	}
	if err := os.WriteFile(nestedFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("setup: create nested file: %v", err)
	}

	cleanupBogusRootDir(projectDir)

	if _, err := os.Stat(bogusDir); !os.IsNotExist(err) {
		t.Error("bogus {} directory should have been removed")
	}
}

// TestCleanupBogusRootDir_NoDirectory verifies that the function is a no-op
// when the "{}" directory does not exist.
func TestCleanupBogusRootDir_NoDirectory(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// No "{}" directory created — must not panic
	cleanupBogusRootDir(projectDir)
}

// TestCleanupBogusRootDir_IgnoresFile verifies that a regular file named "{}"
// (not a directory) is not removed.
func TestCleanupBogusRootDir_IgnoresFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	bogusFile := filepath.Join(projectDir, "{}")
	if err := os.WriteFile(bogusFile, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("setup: create file: %v", err)
	}

	cleanupBogusRootDir(projectDir)

	// The file should remain untouched.
	if _, err := os.Stat(bogusFile); os.IsNotExist(err) {
		t.Error("regular file named {} should not have been removed")
	}
}

// TestCleanupBogusRootDir_IgnoresSymlink verifies that a symlink named "{}"
// is not followed or removed (symlink attack prevention).
func TestCleanupBogusRootDir_IgnoresSymlink(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	realDir := filepath.Join(projectDir, "real-dir")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatalf("setup: create real dir: %v", err)
	}
	symlinkPath := filepath.Join(projectDir, "{}")
	if err := os.Symlink(realDir, symlinkPath); err != nil {
		t.Skipf("symlink creation requires elevated privileges on this platform: %v", err)
	}

	cleanupBogusRootDir(projectDir)

	// The symlink itself must still exist.
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		t.Error("symlink named {} should not have been removed")
	}
	// The symlink target (real directory) must still exist.
	if _, err := os.Stat(realDir); os.IsNotExist(err) {
		t.Error("symlink target should not have been removed")
	}
}
