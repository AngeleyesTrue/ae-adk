package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCwdChangedHandler_EventType(t *testing.T) {
	h := NewCwdChangedHandler()
	if h.EventType() != EventCwdChanged {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventCwdChanged)
	}
}

func TestCwdChangedHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "directory changed with old/new cwd",
			input: &HookInput{
				SessionID: "sess-001",
				CWD:       "/Users/user/project/src",
				OldCwd:    "/Users/user/project",
				NewCwd:    "/Users/user/project/src",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "root directory",
			input: &HookInput{
				SessionID: "sess-002",
				CWD:       "/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCwdChangedHandler()
			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Errorf("Handle() error = %v, want nil", err)
			}
			if out == nil {
				t.Error("Handle() returned nil output")
			}
		})
	}
}

func TestCwdChangedHandler_EnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")

	// Create .ae/config directory to trigger AE_PROJECT_DIR export
	aeDir := filepath.Join(tmpDir, ".ae", "config")
	if err := os.MkdirAll(aeDir, 0o755); err != nil {
		t.Fatalf("failed to create ae config dir: %v", err)
	}

	t.Setenv("CLAUDE_ENV_FILE", envFile)

	h := NewCwdChangedHandler()
	_, err := h.Handle(context.Background(), &HookInput{
		SessionID: "sess-env",
		CWD:       tmpDir,
		NewCwd:    tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("failed to read env file: %v", err)
	}
	content := string(data)
	if content == "" {
		t.Error("env file is empty, expected AE_PROJECT_DIR export")
	}
}

func TestCwdChangedHandler_NoEnvFileWithoutAeDir(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")

	t.Setenv("CLAUDE_ENV_FILE", envFile)

	h := NewCwdChangedHandler()
	_, err := h.Handle(context.Background(), &HookInput{
		SessionID: "sess-no-ae",
		CWD:       tmpDir,
		NewCwd:    tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// File should not be created when no .ae/config exists
	if _, err := os.Stat(envFile); err == nil {
		t.Error("env file should not exist when no .ae/config present")
	}
}

func TestCwdChangedHandler_ShellInjectionBlocked(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")

	// Create a directory with shell metacharacters in path (simulated via cwd string)
	maliciousCwd := tmpDir + `"; rm -rf /; #`

	// Create .ae/config in tmpDir to trigger the export
	aeDir := filepath.Join(tmpDir, ".ae", "config")
	if err := os.MkdirAll(aeDir, 0o755); err != nil {
		t.Fatalf("failed to create ae config dir: %v", err)
	}

	t.Setenv("CLAUDE_ENV_FILE", envFile)

	h := NewCwdChangedHandler()
	_, err := h.Handle(context.Background(), &HookInput{
		SessionID: "sess-inject",
		CWD:       maliciousCwd,
		NewCwd:    maliciousCwd,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// Env file should NOT be created because the path contains shell metacharacters
	if _, err := os.Stat(envFile); err == nil {
		t.Error("env file should not be written when cwd contains shell metacharacters")
	}
}

func TestContainsShellMetachars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{"/Users/user/project", false},
		{"/tmp/safe-dir_123", false},
		{"C:/Users/dev/project", false},
		{`C:\Users\dev\project`, false},       // Windows path: backslash is safe
		{"/tmp/a;b", false},                    // semicolons are safe inside double quotes
		{"/tmp/a|b", false},                    // pipes are safe inside double quotes
		{"/tmp/a&b", false},                    // ampersand is safe inside double quotes
		{"/tmp/a'b", false},                    // single quotes are safe inside double quotes
		{`/tmp/"; rm -rf /; #`, true},          // double quote breaks out
		{"/tmp/$(whoami)", true},               // $ triggers command expansion
		{"/tmp/`id`", true},                    // backtick triggers command substitution
		{"/tmp/foo\nbar", true},                // newline can inject commands
		{"/tmp/foo\rbar", true},                // carriage return also dangerous
		{"", false},
	}

	for _, tt := range tests {
		got := containsShellMetachars(tt.path)
		if got != tt.want {
			t.Errorf("containsShellMetachars(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
