package hook

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// sha256File returns the SHA256 hex digest of a file's content.
// Used to verify file content equality in idempotency tests without depending
// on filesystem-specific modtime precision (Windows NTFS 100ns vs Linux ext4 1ns)
// or OS-cached metadata.
func sha256File(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file for hash: %v", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// TestInjectCLAUDEEnvFile_NoEnvFile verifies REQ-17: when the project root has
// no .env file, injectCLAUDEEnvFile is a no-op (no settings.local.json created).
// This contract holds on Windows (early-return) and non-Windows (build-tag stub).
func TestInjectCLAUDEEnvFile_NoEnvFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No .env file present.
	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("injectCLAUDEEnvFile with no .env should not error: %v", err)
	}
	// settings.local.json should NOT be created.
	settingsPath := filepath.Join(dir, ".claude", "settings.local.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Errorf("settings.local.json should not be created when .env is absent")
	}
}

func TestInjectCLAUDEEnvFile_EmptyProjectRoot(t *testing.T) {
	t.Parallel()

	if err := injectCLAUDEEnvFile(""); err != nil {
		t.Fatalf("empty projectRoot should not error: %v", err)
	}
}

func TestInjectCLAUDEEnvFile_WithEnvFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific injection: skipping on non-Windows (function is no-op)")
	}

	dir := t.TempDir()

	// Create .env file.
	envFilePath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFilePath, []byte("API_KEY=secret"), 0o600); err != nil {
		t.Fatalf("create .env: %v", err)
	}

	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("injectCLAUDEEnvFile: %v", err)
	}

	got := claudeEnvFilePath(dir)
	if got == "" {
		t.Fatal("CLAUDE_ENV_FILE not written to settings.local.json")
	}
	// Path should use forward slashes.
	for _, c := range got {
		if c == '\\' {
			t.Errorf("path should use forward slashes, got: %s", got)
			break
		}
	}
}

func TestInjectCLAUDEEnvFile_Idempotent(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific injection: skipping on non-Windows (function is no-op)")
	}

	dir := t.TempDir()
	envFilePath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFilePath, []byte("X=1"), 0o600); err != nil {
		t.Fatalf("create .env: %v", err)
	}

	// First call writes the file.
	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("first call: %v", err)
	}
	settingsPath := filepath.Join(dir, ".claude", "settings.local.json")
	hash1 := sha256File(t, settingsPath)

	// Second call should be idempotent (same value → no write).
	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("second call: %v", err)
	}
	hash2 := sha256File(t, settingsPath)

	// Compare file content via SHA256 instead of modtime — content equality is the
	// true idempotency invariant and is independent of filesystem timestamp precision.
	if hash1 != hash2 {
		t.Errorf("idempotency violated: settings.local.json content changed on second call\nfirst SHA256:  %s\nsecond SHA256: %s", hash1, hash2)
	}
}

// TestInjectCLAUDEEnvFile_PreservesHeterogeneousEnvTypes verifies that
// existing env entries with non-string values (numbers, objects, booleans)
// are preserved verbatim when CLAUDE_ENV_FILE is injected.
//
// Regression guard for the previous map[string]string parse approach which
// failed with "json: cannot unmarshal number into Go value of type string"
// and silently aborted the entire injection — leaving CLAUDE_ENV_FILE unset.
func TestInjectCLAUDEEnvFile_PreservesHeterogeneousEnvTypes(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific injection: skipping on non-Windows (function is no-op)")
	}

	dir := t.TempDir()

	// Pre-populate .claude/settings.local.json with mixed-type env entries
	// that an external tool (or a future Claude Code release) might add.
	settingsDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("mkdir settings dir: %v", err)
	}
	settingsPath := filepath.Join(settingsDir, "settings.local.json")
	pre := `{
  "env": {
    "STRING_VAR": "hello",
    "NUMBER_VAR": 42,
    "BOOL_VAR": true,
    "OBJECT_VAR": {"nested": "value"}
  },
  "otherTopLevel": "preserved"
}`
	if err := os.WriteFile(settingsPath, []byte(pre), 0o600); err != nil {
		t.Fatalf("write pre-existing settings: %v", err)
	}

	// Create .env so injection actually runs.
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("X=1"), 0o600); err != nil {
		t.Fatalf("create .env: %v", err)
	}

	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("injectCLAUDEEnvFile must succeed when env contains heterogeneous types, got: %v", err)
	}

	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}

	var got struct {
		Env            map[string]json.RawMessage `json:"env"`
		OtherTopLevel  string                     `json:"otherTopLevel"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("settings.local.json must remain valid JSON, got: %v\ncontent: %s", err, string(raw))
	}

	// CLAUDE_ENV_FILE must have been injected.
	if _, ok := got.Env["CLAUDE_ENV_FILE"]; !ok {
		t.Error("CLAUDE_ENV_FILE was not injected (heterogeneous env aborted injection — regression)")
	}

	// All pre-existing entries must be preserved with their original types.
	// json.MarshalIndent re-formats RawMessage payloads with whitespace, so byte
	// comparison fails for object values. Compare structurally by decoding each
	// raw payload back to interface{} and using reflect.DeepEqual.
	mustPreserve := func(key string, want any) {
		t.Helper()
		raw, ok := got.Env[key]
		if !ok {
			t.Errorf("env[%q] missing after injection (must preserve)", key)
			return
		}
		var have any
		if err := json.Unmarshal(raw, &have); err != nil {
			t.Errorf("env[%q] roundtrip parse failed: %v (raw: %s)", key, err, string(raw))
			return
		}
		if !reflect.DeepEqual(have, want) {
			t.Errorf("env[%q] = %v (%T), want %v (%T) — type/value not preserved", key, have, have, want, want)
		}
	}
	mustPreserve("STRING_VAR", "hello")
	mustPreserve("NUMBER_VAR", float64(42)) // json numbers decode to float64
	mustPreserve("BOOL_VAR", true)
	mustPreserve("OBJECT_VAR", map[string]any{"nested": "value"})

	// Top-level fields outside of "env" must also be preserved.
	if got.OtherTopLevel != "preserved" {
		t.Errorf("otherTopLevel = %q, want %q", got.OtherTopLevel, "preserved")
	}
}

func TestInjectCLAUDEEnvFile_NonWindowsNoOp(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("this test verifies the non-Windows no-op path")
	}

	dir := t.TempDir()
	// Create .env file — even with it present, function should be no-op on non-Windows.
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("X=1"), 0o600); err != nil {
		t.Fatalf("create .env: %v", err)
	}

	if err := injectCLAUDEEnvFile(dir); err != nil {
		t.Fatalf("non-Windows injectCLAUDEEnvFile should not error: %v", err)
	}

	// settings.local.json must NOT be created on non-Windows.
	settingsPath := filepath.Join(dir, ".claude", "settings.local.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Errorf("settings.local.json must not be created on non-Windows by injectCLAUDEEnvFile")
	}
}
