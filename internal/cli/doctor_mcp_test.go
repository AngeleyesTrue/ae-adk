package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckMCPScopeDuplicates verifies duplicate MCP server key detection (REQ-15).
func TestCheckMCPScopeDuplicates(t *testing.T) {
	dir := t.TempDir()

	writeMCPJSON := func(path string, servers map[string]any) {
		t.Helper()
		raw, err := json.Marshal(map[string]any{"mcpServers": servers})
		if err != nil {
			t.Fatalf("marshal mcp json: %v", err)
		}
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			t.Fatalf("write mcp json %s: %v", path, err)
		}
	}

	t.Run("PositiveDuplicate", func(t *testing.T) {
		// project and global both have "shared-server"; only global has "global-only"
		projectPath := filepath.Join(dir, "project.mcp.json")
		globalPath := filepath.Join(dir, "global.mcp.json")

		writeMCPJSON(projectPath, map[string]any{
			"context7":      map[string]any{"command": "npx"},
			"shared-server": map[string]any{"command": "shared"},
		})
		writeMCPJSON(globalPath, map[string]any{
			"shared-server": map[string]any{"command": "shared"},
			"global-only":   map[string]any{"command": "global"},
		})

		result := checkMCPScopeDuplicatesWithPaths(projectPath, globalPath)

		if result.Status != CheckWarn {
			t.Errorf("expected CheckWarn, got %v (message: %s)", result.Status, result.Message)
		}
		if !strings.Contains(result.Message, "shared-server") {
			t.Errorf("expected 'shared-server' in message, got: %s", result.Message)
		}
		if !strings.Contains(result.Message, "1 duplicate") {
			t.Errorf("expected '1 duplicate' in message, got: %s", result.Message)
		}
	})

	t.Run("NegativeNoOverlap", func(t *testing.T) {
		projectPath := filepath.Join(dir, "project2.mcp.json")
		globalPath := filepath.Join(dir, "global2.mcp.json")

		writeMCPJSON(projectPath, map[string]any{
			"project-only": map[string]any{"command": "proj"},
		})
		writeMCPJSON(globalPath, map[string]any{
			"global-only": map[string]any{"command": "global"},
		})

		result := checkMCPScopeDuplicatesWithPaths(projectPath, globalPath)

		if result.Status != CheckOK {
			t.Errorf("expected CheckOK, got %v (message: %s)", result.Status, result.Message)
		}
		if !strings.Contains(result.Message, "no duplicate") {
			t.Errorf("expected 'no duplicate' in message, got: %s", result.Message)
		}
	})

	t.Run("MalformedJSONSkipsGracefully", func(t *testing.T) {
		projectPath := filepath.Join(dir, "project3.mcp.json")
		globalPath := filepath.Join(dir, "global3.mcp.json")

		// Write malformed JSON to both files; loadMCPServerKeys must not panic.
		if err := os.WriteFile(projectPath, []byte("not-valid-json{{{"), 0o644); err != nil {
			t.Fatalf("write malformed project: %v", err)
		}
		if err := os.WriteFile(globalPath, []byte("{broken"), 0o644); err != nil {
			t.Fatalf("write malformed global: %v", err)
		}

		// Must not panic; both files parse as empty maps so no duplicates.
		result := checkMCPScopeDuplicatesWithPaths(projectPath, globalPath)

		if result.Status != CheckOK {
			t.Errorf("malformed JSON should result in CheckOK (no duplicates), got %v", result.Status)
		}
	})

	t.Run("MissingFilesSkipGracefully", func(t *testing.T) {
		projectPath := filepath.Join(dir, "nonexistent-project.mcp.json")
		globalPath := filepath.Join(dir, "nonexistent-global.mcp.json")

		result := checkMCPScopeDuplicatesWithPaths(projectPath, globalPath)

		if result.Status != CheckOK {
			t.Errorf("missing files should result in CheckOK, got %v", result.Status)
		}
	})
}
