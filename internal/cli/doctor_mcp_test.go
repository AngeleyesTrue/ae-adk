package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeMCPJSONFile is a shared helper that writes a {"mcpServers": ...} document
// to the given path. Each call uses t.Helper() so failures point at the caller.
func writeMCPJSONFile(t *testing.T, path string, servers map[string]any) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"mcpServers": servers})
	if err != nil {
		t.Fatalf("marshal mcp json: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write mcp json %s: %v", path, err)
	}
}

// TestCheckMCPScopeDuplicates verifies duplicate MCP server key detection (REQ-15).
//
// 각 sub-test가 자기 t.TempDir()을 사용하도록 분리하여 향후 t.Parallel() 도입 시
// 파일 이름 충돌(project.mcp.json 등 고정 이름)이 발생하지 않도록 격리한다.
func TestCheckMCPScopeDuplicates(t *testing.T) {
	t.Parallel()

	t.Run("PositiveDuplicate", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// project and global both have "shared-server"; only global has "global-only"
		projectPath := filepath.Join(dir, "project.mcp.json")
		globalPath := filepath.Join(dir, "global.mcp.json")

		writeMCPJSONFile(t, projectPath, map[string]any{
			"context7":      map[string]any{"command": "npx"},
			"shared-server": map[string]any{"command": "shared"},
		})
		writeMCPJSONFile(t, globalPath, map[string]any{
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

	t.Run("MultipleDuplicatesSorted", func(t *testing.T) {
		// 3개 중복이 결정적(사전순)으로 출력되는지 검증한다 — sort.Strings 회귀 보호.
		t.Parallel()
		dir := t.TempDir()
		projectPath := filepath.Join(dir, "project.mcp.json")
		globalPath := filepath.Join(dir, "global.mcp.json")

		writeMCPJSONFile(t, projectPath, map[string]any{
			"zeta":  map[string]any{"command": "z"},
			"alpha": map[string]any{"command": "a"},
			"mu":    map[string]any{"command": "m"},
		})
		writeMCPJSONFile(t, globalPath, map[string]any{
			"zeta":  map[string]any{"command": "z"},
			"alpha": map[string]any{"command": "a"},
			"mu":    map[string]any{"command": "m"},
		})

		result := checkMCPScopeDuplicatesWithPaths(projectPath, globalPath)

		if result.Status != CheckWarn {
			t.Fatalf("expected CheckWarn, got %v (message: %s)", result.Status, result.Message)
		}
		// 사전순 alpha → mu → zeta 가 동일 message 안에 정확한 순서로 등장해야 한다.
		idxAlpha := strings.Index(result.Message, "alpha")
		idxMu := strings.Index(result.Message, "mu")
		idxZeta := strings.Index(result.Message, "zeta")
		if idxAlpha < 0 || idxMu < 0 || idxZeta < 0 {
			t.Fatalf("expected all three keys in message, got: %s", result.Message)
		}
		if !(idxAlpha < idxMu && idxMu < idxZeta) {
			t.Errorf("expected alphabetical order alpha<mu<zeta in message, got: %s", result.Message)
		}
	})

	t.Run("NegativeNoOverlap", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		projectPath := filepath.Join(dir, "project.mcp.json")
		globalPath := filepath.Join(dir, "global.mcp.json")

		writeMCPJSONFile(t, projectPath, map[string]any{
			"project-only": map[string]any{"command": "proj"},
		})
		writeMCPJSONFile(t, globalPath, map[string]any{
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
		t.Parallel()
		dir := t.TempDir()
		projectPath := filepath.Join(dir, "project.mcp.json")
		globalPath := filepath.Join(dir, "global.mcp.json")

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
		t.Parallel()
		dir := t.TempDir()
		projectPath := filepath.Join(dir, "nonexistent-project.mcp.json")
		globalPath := filepath.Join(dir, "nonexistent-global.mcp.json")

		result := checkMCPScopeDuplicatesWithPaths(projectPath, globalPath)

		if result.Status != CheckOK {
			t.Errorf("missing files should result in CheckOK, got %v", result.Status)
		}
	})
}
